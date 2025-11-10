import { Outlet, Link, useLocation } from 'react-router-dom';
import { cn } from '@/lib/utils';

export function RootLayout() {
  const location = useLocation();

  const isActive = (path: string) => {
    if (path === '/') return location.pathname === '/';
    return location.pathname.startsWith(path);
  };

  return (
    <div className="min-h-screen bg-background">
      <nav className="border-b">
        <div className="container mx-auto px-4">
          <div className="flex h-16 items-center space-x-8">
            <Link to="/" className="text-xl font-bold">
              DockerMC Cloud Manager
            </Link>
            <div className="flex space-x-4">
              <Link
                to="/servers"
                className={cn(
                  'px-3 py-2 rounded-md text-sm font-medium transition-colors',
                  isActive('/servers')
                    ? 'bg-primary text-primary-foreground'
                    : 'text-foreground hover:bg-accent'
                )}
              >
                Servers
              </Link>
              <Link
                to="/proxy"
                className={cn(
                  'px-3 py-2 rounded-md text-sm font-medium transition-colors',
                  isActive('/proxy')
                    ? 'bg-primary text-primary-foreground'
                    : 'text-foreground hover:bg-accent'
                )}
              >
                Proxy
              </Link>
            </div>
          </div>
        </div>
      </nav>
      <main className="container mx-auto px-4 py-8">
        <Outlet />
      </main>
    </div>
  );
}
