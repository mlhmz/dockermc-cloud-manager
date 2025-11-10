import { Loader2, Play, Square, RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useProxy } from '@/hooks/queries/useProxy';
import { useServers } from '@/hooks/queries/useServers';
import { useStartProxy } from '@/hooks/mutations/useStartProxy';
import { useStopProxy } from '@/hooks/mutations/useStopProxy';
import { useRegenerateProxyConfig } from '@/hooks/mutations/useRegenerateProxyConfig';
import { useUpdateProxy } from '@/hooks/mutations/useUpdateProxy';
import { useState } from 'react';
import { StatusBadge } from '@/components/status-badge';

export function ProxyPage() {
  const { data: proxy, isLoading: proxyLoading, error: proxyError } = useProxy();
  const { data: servers } = useServers();
  const startProxy = useStartProxy();
  const stopProxy = useStopProxy();
  const regenerateConfig = useRegenerateProxyConfig();
  const updateProxy = useUpdateProxy();
  const [selectedServerId, setSelectedServerId] = useState<string>('');

  const handleStartStop = () => {
    if (proxy?.status === 'running') {
      stopProxy.mutate();
    } else {
      startProxy.mutate();
    }
  };

  const handleUpdateDefaultServer = () => {
    updateProxy.mutate({ default_server_id: selectedServerId || undefined });
  };

  if (proxyLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (proxyError) {
    return (
      <div>
        <h1 className="text-3xl font-bold mb-6">Velocity Proxy</h1>
        <Card>
          <CardContent className="flex flex-col items-center justify-center h-64">
            <p className="text-muted-foreground mb-4">Proxy not found. Start the proxy to get started.</p>
            <Button onClick={() => startProxy.mutate()} disabled={startProxy.isPending}>
              {startProxy.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              <Play className="mr-2 h-4 w-4" />
              Start Proxy
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!proxy) {
    return null;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Velocity Proxy</h1>
        <StatusBadge status={proxy.status}/>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Proxy Information</CardTitle>
            <CardDescription>Details about the Velocity proxy server</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Status:</span>
                <span className="font-medium">{proxy.status}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Port:</span>
                <span className="font-medium">{proxy.port}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Container ID:</span>
                <span className="font-mono text-xs overflow-x-scroll max-w-64">{proxy.container_id}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Volume ID:</span>
                <span className="font-mono text-xs">{proxy.volume_id}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Created:</span>
                <span className="text-sm">{new Date(proxy.created_at).toLocaleString()}</span>
              </div>
            </div>

            <div className="space-y-2">
              <Button
                onClick={handleStartStop}
                disabled={proxy.status === 'creating' || proxy.status === 'error' || startProxy.isPending || stopProxy.isPending}
                className="w-full"
              >
                {proxy.status === 'running' ? (
                  <>
                    <Square className="mr-2 h-4 w-4" />
                    Stop Proxy
                  </>
                ) : (
                  <>
                    <Play className="mr-2 h-4 w-4" />
                    Start Proxy
                  </>
                )}
              </Button>

              <Button
                variant="outline"
                onClick={() => regenerateConfig.mutate()}
                disabled={regenerateConfig.isPending}
                className="w-full"
              >
                {regenerateConfig.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                <RefreshCw className="mr-2 h-4 w-4" />
                Regenerate Config
              </Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Default Lobby Server</CardTitle>
            <CardDescription>
              Select which server players will join when they first connect
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              {proxy.default_server_id && (
                <div className="mb-4 p-3 bg-muted rounded-md">
                  <span className="text-sm text-muted-foreground">Current default server:</span>
                  <p className="font-medium">
                    {servers?.find(s => s.id === proxy.default_server_id)?.name || 'Unknown'}
                  </p>
                </div>
              )}

              <Label htmlFor="default-server">Select Default Server</Label>
              <Select
                value={selectedServerId}
                onValueChange={setSelectedServerId}
              >
                <SelectTrigger id="default-server">
                  <SelectValue placeholder="Select a server..." />
                </SelectTrigger>
                <SelectContent>
                  {servers?.filter(s => s.status === 'running').map((server) => (
                    <SelectItem key={server.id} value={server.id}>
                      {server.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              <Button
                onClick={handleUpdateDefaultServer}
                disabled={updateProxy.isPending}
                className="w-full"
              >
                {updateProxy.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Update Default Server
              </Button>
            </div>

            <div className="pt-4 border-t">
              <p className="text-sm text-muted-foreground">
                {servers?.filter(s => s.status === 'running').length || 0} server(s) currently running
              </p>
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>About the Proxy</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm text-muted-foreground">
          <p>
            The Velocity proxy is the central hub for all Minecraft servers. Players connect to the proxy
            on port {proxy.port}, and it routes them to backend servers.
          </p>
          <p>
            All servers are automatically connected to the proxy when created. The proxy configuration
            is dynamically updated with auto-generated forwarding secrets for security.
          </p>
        </CardContent>
      </Card>
    </div>
  );
}
