import { createBrowserRouter } from 'react-router-dom';
import { RootLayout } from './routes/layout';
import { ServersPage } from './routes/servers';
import { ServerDetailPage } from './routes/server-detail';
import { ProxyPage } from './routes/proxy';

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
