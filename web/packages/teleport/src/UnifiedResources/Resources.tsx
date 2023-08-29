/*
Copyright 2019-2022 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import React, { useRef } from 'react';
import { Box, Indicator, Flex, ButtonLink, ButtonSecondary } from 'design';

import styled from 'styled-components';

import { Danger } from 'design/Alert';

import {
  FeatureBox,
  FeatureHeader,
  FeatureHeaderTitle,
} from 'teleport/components/Layout';
import useTeleport from 'teleport/useTeleport';
import cfg from 'teleport/config';
import history from 'teleport/services/history/history';
import localStorage from 'teleport/services/localStorage';
import useStickyClusterId from 'teleport/useStickyClusterId';
import AgentButtonAdd from 'teleport/components/AgentButtonAdd';
import { SearchResource } from 'teleport/Discover/SelectResource';
import { useUrlFiltering, useInfiniteScroll } from 'teleport/components/hooks';

import { ResourceCard } from './ResourceCard';
import SearchPanel from './SearchPanel';
import { FilterPanel } from './FilterPanel';

export function Resources() {
  const { isLeafCluster } = useStickyClusterId();
  const enabled = localStorage.areUnifiedResourcesEnabled();
  const teleCtx = useTeleport();
  const canCreate = teleCtx.storeUser.getTokenAccess().create;
  const { clusterId } = useStickyClusterId();

  const filtering = useUrlFiltering({
    fieldName: 'name',
    dir: 'ASC',
  });
  const { params, setParams, replaceHistory, pathname, setSort, onLabelClick } =
    filtering;

  const scrollDetector = useRef(null);
  const {
    setTrigger: setScrollDetector,
    forceFetch,
    resources,
    attempt,
  } = useInfiniteScroll({
    fetchFunc: teleCtx.resourceService.fetchUnifiedResources,
    // trigger: scrollDetector.current,
    clusterId,
    filter: params,
  });

  if (!enabled) {
    history.replace(cfg.getNodesRoute(clusterId));
  }

  const onRetryClicked = () => {
    forceFetch();
  };

  return (
    <FeatureBox>
      {attempt.status === 'failed' && (
        <ErrorBox>
          <ErrorBoxInternal>
            <Danger>
              {attempt.statusText}
              <ButtonLink onClick={onRetryClicked}>Retry</ButtonLink>
            </Danger>
          </ErrorBoxInternal>
        </ErrorBox>
      )}
      <FeatureHeader alignItems="center" justifyContent="space-between">
        <FeatureHeaderTitle>Resources</FeatureHeaderTitle>
        <Flex alignItems="center">
          <AgentButtonAdd
            agent={SearchResource.UNIFIED_RESOURCE}
            beginsWithVowel={false}
            isLeafCluster={isLeafCluster}
            canCreate={canCreate}
          />
        </Flex>
      </FeatureHeader>
      <SearchPanel
        params={params}
        setParams={setParams}
        pathname={pathname}
        replaceHistory={replaceHistory}
      />
      <FilterPanel
        params={params}
        setParams={setParams}
        setSort={setSort}
        pathname={pathname}
        replaceHistory={replaceHistory}
      />
      <ResourcesContainer gap={2}>
        {resources.map((res, i) => (
          <ResourceCard key={i} onLabelClick={onLabelClick} resource={res} />
        ))}
      </ResourcesContainer>
      <div ref={setScrollDetector} />
      <ListFooter>
        <IndicatorContainer status={attempt.status}>
          <Indicator size={INDICATOR_SIZE} />
        </IndicatorContainer>
        {attempt.status === 'failed' && resources.length > 0 && (
          <ButtonSecondary onClick={onRetryClicked}>Load more</ButtonSecondary>
        )}
      </ListFooter>
    </FeatureBox>
  );
}

const INDICATOR_SIZE = '48px';

const ResourcesContainer = styled(Flex)`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
`;

const ErrorBox = styled(Box)`
  position: sticky;
  top: 0;
  z-index: 1;
`;

const ErrorBoxInternal = styled(Box)`
  position: absolute;
  left: 0;
  right: 0;
  margin: ${props => props.theme.space[1]}px 10% 0 10%;
`;

// It's important to make the footer at least as big as the loading indicator,
// since in the typical case, we want to avoid UI "jumping" when loading the
// final fragment finishes, and the final fragment is just one element in the
// final row (i.e. the number of rows doesn't change). It's then important to
// keep the same amount of whitespace below the resource list.
const ListFooter = styled.div`
  margin-top: ${props => props.theme.space[2]}px;
  min-height: ${INDICATOR_SIZE};
  text-align: center;
`;

// Line height is set to 0 to prevent the layout engine from adding extra pixels
// to the element's height.
const IndicatorContainer = styled(Box)`
  display: ${props => (props.status === 'processing' ? 'block' : 'none')};
  line-height: 0;
`;
