import { createBrowserRouter } from 'react-router-dom';
import { RootLayout } from './layout';
import { ServersPage } from './servers';
import { ServerDetailPage } from './server-detail';
import { ProxyPage } from './proxy';

export const router = createBrowserRouter([
  {
    path: '/',
    element: <RootLayout />,
    children: [
      {
        index: true,
        element: <ServersPage />,
      },
      {
        path: 'servers',
        element: <ServersPage />,
      },
      {
        path: 'servers/:id',
        element: <ServerDetailPage />,
      },
      {
        path: 'proxy',
        element: <ProxyPage />,
      },
    ],
  },
]);
