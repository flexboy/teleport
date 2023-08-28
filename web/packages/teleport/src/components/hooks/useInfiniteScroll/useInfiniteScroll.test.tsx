/**
 * Copyright 2023 Gravitational, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import React from 'react';

import { renderHook, act } from '@testing-library/react-hooks';
import { render, screen, waitFor } from 'design/utils/testing';
import { mockIntersectionObserver } from 'jsdom-testing-mocks';

import { useInfiniteScroll, Props } from './useInfiniteScroll';
import { TestResource, newFetchFunc, resourceNames } from './testUtils';

const mio = mockIntersectionObserver();

function hookProps(overrides: Partial<Props<TestResource>> = {}) {
  return {
    fetchFunc: newFetchFunc(7),
    trigger: null,
    clusterId: 'test-cluster',
    filter: {},
    initialFetchSize: 2,
    fetchMoreSize: 3,
    ...overrides,
  };
}

test('fetches data whenever an element is in view', async () => {
  render(<div data-testid="trigger" />);
  const trigger = screen.getByTestId('trigger');
  const { result, waitForNextUpdate } = renderHook(() =>
    useInfiniteScroll(hookProps({ trigger }))
  );
  expect(resourceNames(result)).toEqual([]);

  act(() => mio.enterNode(trigger));
  await waitForNextUpdate();
  expect(resourceNames(result)).toEqual(['r0', 'r1']);

  act(() => mio.leaveNode(trigger));
  expect(resourceNames(result)).toEqual(['r0', 'r1']);

  act(() => mio.enterNode(trigger));
  await waitForNextUpdate();
  expect(resourceNames(result)).toEqual(['r0', 'r1', 'r2', 'r3', 'r4']);
});

test('supports changing nodes', async () => {
  render(
    <>
      <div data-testid="trigger1" />
      <div data-testid="trigger2" />
    </>
  );
  const trigger1 = screen.getByTestId('trigger1');
  const trigger2 = screen.getByTestId('trigger2');
  let props = hookProps({ trigger: trigger1 });
  const { result, rerender, waitForNextUpdate } = renderHook(
    useInfiniteScroll,
    {
      initialProps: props,
    }
  );

  act(() => mio.enterNode(trigger1));
  await waitForNextUpdate();
  expect(resourceNames(result)).toEqual(['r0', 'r1']);

  props = { ...props, trigger: trigger2 };
  rerender(props);

  // Should only register entering trigger2, reading resources r2 through r4.
  act(() => mio.leaveNode(trigger1));
  act(() => mio.enterNode(trigger1));
  act(() => mio.enterNode(trigger2));
  await waitForNextUpdate();
  expect(resourceNames(result)).toEqual(['r0', 'r1', 'r2', 'r3', 'r4']);
});
