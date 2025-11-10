import { useState } from 'react';
import { Link } from 'react-router-dom';
import { Plus, Play, Square, Trash2, Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useServers } from '@/hooks/queries/useServers';
import { useStartServer } from '@/hooks/mutations/useStartServer';
import { useStopServer } from '@/hooks/mutations/useStopServer';
import { useDeleteServer } from '@/hooks/mutations/useDeleteServer';
import { CreateServerDialog } from '@/components/create-server-dialog';
import type { MinecraftServer } from '@/types/api';

function getStatusColor(status: MinecraftServer['status']) {
  switch (status) {
    case 'running':
      return 'bg-green-500';
    case 'stopped':
      return 'bg-gray-500';
    case 'creating':
      return 'bg-blue-500';
    case 'error':
      return 'bg-red-500';
    default:
      return 'bg-gray-500';
  }
}

export function ServersPage() {
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const { data: servers, isLoading, error } = useServers();
  const startServer = useStartServer();
  const stopServer = useStopServer();
  const deleteServer = useDeleteServer();

  const handleStartStop = (server: MinecraftServer) => {
    if (server.status === 'running') {
      stopServer.mutate(server.id);
    } else if (server.status === 'stopped') {
      startServer.mutate(server.id);
    }
  };

  const handleDelete = (id: string) => {
    if (confirm('Are you sure you want to delete this server? This action cannot be undone.')) {
      deleteServer.mutate(id);
    }
  };

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold">Minecraft Servers</h1>
        <Button onClick={() => setCreateDialogOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Server
        </Button>
      </div>

      {isLoading && (
        <div className="flex items-center justify-center h-64">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      )}

      {error && (
        <div className="text-destructive">
          Failed to load servers: {error.message}
        </div>
      )}

      {servers && servers.length === 0 && (
        <Card>
          <CardContent className="flex flex-col items-center justify-center h-64">
            <p className="text-muted-foreground mb-4">No servers yet. Create your first server to get started!</p>
            <Button onClick={() => setCreateDialogOpen(true)}>
              <Plus className="mr-2 h-4 w-4" />
              Create Server
            </Button>
          </CardContent>
        </Card>
      )}

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {servers?.map((server) => (
          <Card key={server.id}>
            <CardHeader>
              <div className="flex items-start justify-between">
                <div>
                  <CardTitle>{server.name}</CardTitle>
                  <CardDescription className="mt-1">{server.motd}</CardDescription>
                </div>
                <Badge className={getStatusColor(server.status)}>
                  {server.status}
                </Badge>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-2 text-sm">
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
                  <span className="text-muted-foreground">Container:</span>
                  <span className="font-mono text-xs">{server.container_id.substring(0, 12)}</span>
                </div>
              </div>
            </CardContent>
            <CardFooter className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => handleStartStop(server)}
                disabled={server.status === 'creating' || server.status === 'error' || startServer.isPending || stopServer.isPending}
                className="flex-1"
              >
                {server.status === 'running' ? (
                  <>
                    <Square className="mr-2 h-4 w-4" />
                    Stop
                  </>
                ) : (
                  <>
                    <Play className="mr-2 h-4 w-4" />
                    Start
                  </>
                )}
              </Button>
              <Button
                variant="outline"
                size="sm"
                asChild
                className="flex-1"
              >
                <Link to={`/servers/${server.id}`}>
                  View Logs
                </Link>
              </Button>
              <Button
                variant="destructive"
                size="sm"
                onClick={() => handleDelete(server.id)}
                disabled={deleteServer.isPending}
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </CardFooter>
          </Card>
        ))}
      </div>

      <CreateServerDialog
        open={createDialogOpen}
        onOpenChange={setCreateDialogOpen}
      />
    </div>
  );
}
