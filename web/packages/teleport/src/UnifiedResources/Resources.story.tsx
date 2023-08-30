import React from 'react';
import { MemoryRouter } from 'react-router';
import { Resources } from './Resources';
import { createTeleportContext } from 'teleport/mocks/contexts';
import { ContextProvider } from 'teleport';
import cfg from 'teleport/config';
import { apps } from 'teleport/Apps/fixtures';
import { databases } from 'teleport/Databases/fixtures';
import { kubes } from 'teleport/Kubes/fixtures';
import { desktops } from 'teleport/Desktops/fixtures';
import { nodes } from 'teleport/Nodes/fixtures';
import { initialize, mswLoader } from 'msw-storybook-addon';
import { rest, ResponseResolver, MockedRequest, RestContext } from 'msw';

initialize();

export default {
  title: 'Teleport/UnifiedResources',
  loaders: [mswLoader],
};

const allResources = [
  ...apps,
  ...databases,
  ...kubes,
  ...desktops,
  ...nodes,
  ...apps,
  ...databases,
  ...kubes,
  ...desktops,
  ...nodes,
];

const Provider = props => {
  const ctx = createTeleportContext();

  return (
    <MemoryRouter>
      <ContextProvider ctx={ctx}>{props.children}</ContextProvider>
    </MemoryRouter>
  );
};

const story = (resolver: ResponseResolver<MockedRequest, RestContext>) => {
  const ctx = createTeleportContext();

  const s = () => (
    <MemoryRouter>
      <ContextProvider ctx={ctx}>
        <Resources />
      </ContextProvider>
    </MemoryRouter>
  );

  s.parameters = {
    msw: {
      handlers: [
        rest.get(cfg.getUnifiedResourcesUrl('localhost', {}), resolver),
      ],
    },
  };
  return s;
};

export const Empty = story((_, res, ctx) => res(ctx.json({ items: [] })));

export const List = story((_, res, ctx) =>
  res(ctx.json({ items: allResources }))
);

export const Loading = story((_, res, ctx) => res(ctx.delay('infinite')));

export const LoadingAfterScrolling = story((req, res, ctx) => {
  if (req.url.searchParams.get('startKey') === 'next-key') {
    return res(ctx.delay('infinite'));
  }
  return res(ctx.json({ items: allResources, startKey: 'next-key' }));
});

export const Error = story((_, res, ctx) => res(ctx.status(500)));

export const ErrorAfterScrolling = story((req, res, ctx) => {
  if (req.url.searchParams.get('startKey') === 'next-key') {
    return res(ctx.status(500));
  }
  return res(ctx.json({ items: allResources, startKey: 'next-key' }));
});
