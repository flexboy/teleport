/**
 * Copyright 2023 Gravitational, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React, {
  createContext,
  FC,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react';
import {
  Attempt,
  makeSuccessAttempt,
  useAsync,
  makeEmptyAttempt,
} from 'shared/hooks/useAsync';

import { RootClusterUri } from 'teleterm/ui/uri';
import { useAppContext } from 'teleterm/ui/appContextProvider';
import { Server, TshAbortSignal } from 'teleterm/services/tshd/types';
import createAbortController from 'teleterm/services/tshd/createAbortController';
import { isAccessDeniedError } from 'teleterm/services/tshd/errors';

import { assertUnreachable, retryWithRelogin } from '../utils';

import { canUseConnectMyComputer } from './permissions';

import type {
  AgentProcessState,
  MainProcessClient,
} from 'teleterm/mainProcess/types';

export type CurrentAction =
  | {
      kind: 'download';
      attempt: Attempt<void>;
    }
  | {
      kind: 'start';
      attempt: Attempt<Server>;
      agentProcessState: AgentProcessState;
      /**
       * timeoutLogs contains the last 10 lines of logs captured at the time of the start attempt
       * timing out. Present only if the attempt timed out.
       */
      timeoutLogs: string;
    }
  | {
      kind: 'observe-process';
      agentProcessState: AgentProcessState;
    }
  | {
      kind: 'kill';
      attempt: Attempt<void>;
    }
  | {
      kind: 'remove';
      attempt: Attempt<void>;
    };

export interface ConnectMyComputerContext {
  canUse: boolean;
  currentAction: CurrentAction;
  agentProcessState: AgentProcessState;
  agentNode: Server | undefined;
  startAgent(): Promise<[Server, Error]>;
  downloadAgent(): Promise<[void, Error]>;
  downloadAgentAttempt: Attempt<void>;
  setDownloadAgentAttempt(attempt: Attempt<void>): void;
  downloadAndStartAgent(): Promise<void>;
  killAgent(): Promise<[void, Error]>;
  removeAgent(): Promise<[void, Error]>;
  isAgentConfiguredAttempt: Attempt<boolean>;
  markAgentAsConfigured(): void;
  /**
   * logsFromStartTimeout contains the last 10 lines of logs captured at the time of the start
   * attempt timing out. Present only if the attempt timed out.
   *
   * Prefer to use timeoutLogs from currentAction if possible.
   */
  logsFromStartTimeout: string;
}

const ConnectMyComputerContext = createContext<ConnectMyComputerContext>(null);

export const ConnectMyComputerContextProvider: FC<{
  rootClusterUri: RootClusterUri;
}> = ({ rootClusterUri, children }) => {
  const ctx = useAppContext();
  const {
    mainProcessClient,
    connectMyComputerService,
    clustersService,
    configService,
    workspacesService,
  } = ctx;
  clustersService.useState();

  const [
    isAgentConfiguredAttempt,
    checkIfAgentIsConfigured,
    setAgentConfiguredAttempt,
  ] = useAsync(
    useCallback(
      () => connectMyComputerService.isAgentConfigFileCreated(rootClusterUri),
      [connectMyComputerService, rootClusterUri]
    )
  );

  const rootCluster = clustersService.findCluster(rootClusterUri);
  const canUse = useMemo(
    () =>
      canUseConnectMyComputer(
        rootCluster,
        configService,
        mainProcessClient.getRuntimeSettings()
      ),
    [configService, mainProcessClient, rootCluster]
  );

  const markAgentAsConfigured = useCallback(() => {
    setAgentConfiguredAttempt(makeSuccessAttempt(true));
  }, [setAgentConfiguredAttempt]);
  const markAgentAsNotConfigured = useCallback(() => {
    setDownloadAgentAttempt(makeEmptyAttempt());
    setAgentConfiguredAttempt(makeSuccessAttempt(false));
  }, [setAgentConfiguredAttempt]);

  const [currentActionKind, setCurrentActionKind] =
    useState<CurrentAction['kind']>('observe-process');
  const [logsFromStartTimeout, setLogsFromStartTimeout] = useState('');

  const [agentProcessState, setAgentProcessState] = useState<AgentProcessState>(
    () =>
      mainProcessClient.getAgentState({
        rootClusterUri,
      }) || {
        status: 'not-started',
      }
  );

  const [downloadAgentAttempt, downloadAgent, setDownloadAgentAttempt] =
    useAsync(
      useCallback(async () => {
        setCurrentActionKind('download');
        await connectMyComputerService.downloadAgent();
      }, [connectMyComputerService])
    );

  const [startAgentAttempt, startAgent] = useAsync(
    useCallback(async () => {
      setCurrentActionKind('start');
      setLogsFromStartTimeout('');

      await connectMyComputerService.runAgent(rootClusterUri);

      const abortController = createAbortController();
      try {
        const server = await Promise.race([
          connectMyComputerService.waitForNodeToJoin(
            rootClusterUri,
            abortController.signal
          ),
          throwOnAgentProcessErrors(
            mainProcessClient,
            rootClusterUri,
            abortController.signal
          ),
          wait(20_000, abortController.signal).then(() => {
            setLogsFromStartTimeout(
              mainProcessClient.getAgentLogs({ rootClusterUri })
            );
            throw new NodeWaitJoinTimeout();
          }),
        ]);
        setCurrentActionKind('observe-process');
        workspacesService.setConnectMyComputerAutoStart(rootClusterUri, true);
        return server;
      } catch (error) {
        // in case of any error kill the agent
        await connectMyComputerService.killAgent(rootClusterUri);
        throw error;
      } finally {
        abortController.abort();
      }
    }, [
      connectMyComputerService,
      mainProcessClient,
      rootClusterUri,
      workspacesService,
    ])
  );

  const downloadAndStartAgent = useCallback(async () => {
    const [, error] = await downloadAgent();
    if (error) {
      return;
    }
    await startAgent();
  }, [downloadAgent, startAgent]);

  const [killAgentAttempt, killAgent] = useAsync(
    useCallback(async () => {
      setCurrentActionKind('kill');
      await connectMyComputerService.killAgent(rootClusterUri);
      setCurrentActionKind('observe-process');
      workspacesService.setConnectMyComputerAutoStart(rootClusterUri, false);
    }, [connectMyComputerService, rootClusterUri, workspacesService])
  );

  const [removeAgentAttempt, removeAgent] = useAsync(
    useCallback(async () => {
      const [, error] = await killAgent();
      if (error) {
        throw error;
      }
      setCurrentActionKind('remove');

      try {
        await retryWithRelogin(ctx, rootClusterUri, () =>
          ctx.connectMyComputerService.removeConnectMyComputerNode(
            rootClusterUri
          )
        );
      } catch (e) {
        if (isAccessDeniedError(e)) {
          ctx.notificationsService.notifyInfo({
            title: 'The node may be visible for a few more minutes.',
            description:
              'You do not have permissions to remove nodes, but it will be removed automatically.',
          });
        } else {
          throw e;
        }
      }

      await ctx.connectMyComputerService.removeConnections(rootClusterUri);
      ctx.workspacesService.removeConnectMyComputerState(rootClusterUri);
      await ctx.connectMyComputerService.removeAgentDirectory(rootClusterUri);

      markAgentAsNotConfigured();
    }, [ctx, killAgent, markAgentAsNotConfigured, rootClusterUri])
  );

  useEffect(() => {
    const { cleanup } = mainProcessClient.subscribeToAgentUpdate(
      rootClusterUri,
      setAgentProcessState
    );
    return cleanup;
  }, [mainProcessClient, rootClusterUri]);

  let currentAction: CurrentAction;
  const kind = currentActionKind;

  switch (kind) {
    case 'download': {
      currentAction = { kind, attempt: downloadAgentAttempt };
      break;
    }
    case 'start': {
      currentAction = {
        kind,
        attempt: startAgentAttempt,
        agentProcessState,
        timeoutLogs: logsFromStartTimeout,
      };
      break;
    }
    case 'observe-process': {
      currentAction = { kind, agentProcessState };
      break;
    }
    case 'kill': {
      currentAction = { kind, attempt: killAgentAttempt };
      break;
    }
    case 'remove': {
      currentAction = { kind, attempt: removeAgentAttempt };
      break;
    }
    default: {
      assertUnreachable(kind);
    }
  }

  useEffect(() => {
    // This call checks if the agent is configured even if the user does not have access to Connect My Computer.
    // Unfortunately, we cannot call it only if `canUse === true`, because to resolve `canUse` value
    // we need to fetch some data from the auth server which takes time.
    // This doesn't work for us, because the information if the agent is configured is needed immediately -
    // based on this we replace the setup document with the status document.
    // If we had waited for `canUse` to become true, the user might have seen a setup document
    // which would have been replaced by the other document after 1-2 seconds.
    if (isAgentConfiguredAttempt.status === '') {
      checkIfAgentIsConfigured();
    }
  }, [checkIfAgentIsConfigured, isAgentConfiguredAttempt.status]);

  const isAgentConfigured =
    isAgentConfiguredAttempt.status === 'success' &&
    isAgentConfiguredAttempt.data;
  const agentIsNotStarted =
    currentAction.kind === 'observe-process' &&
    currentAction.agentProcessState.status === 'not-started';

  useEffect(() => {
    const shouldAutoStartAgent =
      isAgentConfigured &&
      canUse &&
      workspacesService.getConnectMyComputerAutoStart(rootClusterUri) &&
      agentIsNotStarted;
    if (shouldAutoStartAgent) {
      downloadAndStartAgent();
    }
  }, [
    canUse,
    downloadAndStartAgent,
    agentIsNotStarted,
    isAgentConfigured,
    rootClusterUri,
    workspacesService,
  ]);

  return (
    <ConnectMyComputerContext.Provider
      value={{
        canUse,
        currentAction,
        agentProcessState,
        agentNode: startAgentAttempt.data,
        killAgent,
        startAgent,
        downloadAgent,
        downloadAgentAttempt,
        setDownloadAgentAttempt,
        downloadAndStartAgent,
        markAgentAsConfigured,
        isAgentConfiguredAttempt,
        logsFromStartTimeout,
        removeAgent,
      }}
      children={children}
    />
  );
};

export const useConnectMyComputerContext = () => {
  const context = useContext(ConnectMyComputerContext);

  if (!context) {
    throw new Error(
      'ConnectMyComputerContext requires ConnectMyComputerContextProvider context.'
    );
  }

  return context;
};

/**
 * Waits for `error` and `exit` events from the agent process and throws when they occur.
 */
function throwOnAgentProcessErrors(
  mainProcessClient: MainProcessClient,
  rootClusterUri: RootClusterUri,
  abortSignal: TshAbortSignal
): Promise<never> {
  return new Promise((_, reject) => {
    const rejectOnError = (agentProcessState: AgentProcessState) => {
      if (
        agentProcessState.status === 'exited' ||
        // TODO(ravicious): 'error' should not be considered a separate process state. See the
        // comment above the 'error' status definition.
        agentProcessState.status === 'error'
      ) {
        reject(new AgentProcessError());
        cleanup();
      }
    };

    const { cleanup } = mainProcessClient.subscribeToAgentUpdate(
      rootClusterUri,
      rejectOnError
    );
    abortSignal.addEventListener(() => {
      cleanup();
      reject(
        new DOMException('throwOnAgentProcessErrors was aborted', 'AbortError')
      );
    });

    // the state may have changed before we started listening, we have to check the current state
    rejectOnError(
      mainProcessClient.getAgentState({
        rootClusterUri,
      })
    );
  });
}

export class AgentProcessError extends Error {
  constructor() {
    super('AgentProcessError');
    this.name = 'AgentProcessError';
  }
}

export class NodeWaitJoinTimeout extends Error {
  constructor() {
    super('NodeWaitJoinTimeout');
    this.name = 'NodeWaitJoinTimeout';
  }
}

/**
 * wait is like wait from the shared package, but it works with TshAbortSignal.
 * TODO(ravicious): Refactor TshAbortSignal so that its interface is the same as AbortSignal.
 * See the comment in createAbortController for more details.
 */
function wait(ms: number, abortSignal: TshAbortSignal): Promise<void> {
  return new Promise((resolve, reject) => {
    const timeout = setTimeout(resolve, ms);
    if (abortSignal) {
      abortSignal.addEventListener(() => {
        clearTimeout(timeout);
        reject(new DOMException('Wait was aborted.', 'AbortError'));
      });
    }
  });
}
