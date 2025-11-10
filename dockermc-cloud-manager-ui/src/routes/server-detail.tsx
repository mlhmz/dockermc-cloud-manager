import { useParams, Link } from 'react-router-dom';
import { ArrowLeft, Loader2, Play, Square } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useServer } from '@/hooks/queries/useServer';
import { useStartServer } from '@/hooks/mutations/useStartServer';
import { useStopServer } from '@/hooks/mutations/useStopServer';
import { ServerLogs } from '@/components/server-logs';
import { StatusBadge } from '@/components/status-badge';

export function ServerDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { data: server, isLoading, error } = useServer(id!);
  const startServer = useStartServer();
  const stopServer = useStopServer();

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (error || !server) {
    return (
      <div>
        <Link to="/servers">
          <Button variant="ghost" className="mb-4">
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Servers
          </Button>
        </Link>
        <div className="text-destructive">
          Failed to load server: {error?.message || 'Server not found'}
        </div>
      </div>
    );
  }

  const handleStartStop = () => {
    if (server.status === 'running') {
      stopServer.mutate(server.id);
    } else if (server.status === 'stopped') {
      startServer.mutate(server.id);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link to="/servers">
            <Button variant="ghost" size="icon">
              <ArrowLeft className="h-4 w-4" />
            </Button>
          </Link>
          <div>
            <h1 className="text-3xl font-bold">{server.name}</h1>
            <p className="text-muted-foreground">{server.motd}</p>
          </div>
        </div>
        <StatusBadge status={server.status} />
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Server Information</CardTitle>
            <CardDescription>Details about this Minecraft server</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Status:</span>
                <span className="font-medium">{server.status}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Max Players:</span>
                <span className="font-medium">{server.max_players}</span>
              </div>
              {server.port && (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Port:</span>
                  <span className="font-medium">{server.port}</span>
                </div>
              )}
              <div className="flex justify-between">
                <span className="text-muted-foreground">Container ID:</span>
                <span className="font-mono text-xs overflow-x-scroll max-w-64">{server.container_id}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Volume ID:</span>
                <span className="font-mono text-xs overflow-x-scroll max-w-64 text-nowrap">{server.volume_id}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Created:</span>
                <span className="text-sm">{new Date(server.created_at).toLocaleString()}</span>
              </div>
            </div>

            <Button
              onClick={handleStartStop}
              disabled={server.status === 'creating' || server.status === 'error' || startServer.isPending || stopServer.isPending}
              className="w-full"
            >
              {server.status === 'running' ? (
                <>
                  <Square className="mr-2 h-4 w-4" />
                  Stop Server
                </>
              ) : (
                <>
                  <Play className="mr-2 h-4 w-4" />
                  Start Server
                </>
              )}
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Quick Actions</CardTitle>
            <CardDescription>Manage this server</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            <p className="text-sm text-muted-foreground">
              Server actions and configurations will be available here.
            </p>
          </CardContent>
        </Card>
      </div>

      <ServerLogs serverId={server.id} />
    </div>
  );
}
