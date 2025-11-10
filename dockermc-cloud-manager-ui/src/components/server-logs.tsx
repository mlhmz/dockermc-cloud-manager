import { useEffect, useRef, useState } from 'react';
import { Loader2, Terminal } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { apiClient } from '@/api/client';
import { Input } from './ui/input';
import { Button } from './ui/button';

interface ServerLogsProps {
  serverId: string;
}

interface CommandMessage {
  type: string
  command: string
}

interface ResponseMessage {
  type: string
  content: string
}

export function ServerLogs({ serverId }: ServerLogsProps) {
  const [logs, setLogs] = useState<ResponseMessage[]>([]);
  const [command, setCommand] = useState<string>("");
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const logsEndRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    let ws: WebSocket;

    try {
      ws = apiClient.createLogWebSocket(serverId, true, '100');
      wsRef.current = ws;

      ws.onopen = () => {
        setIsConnected(true);
        setError(null);
      };

      ws.onmessage = (event) => {
        const logLine = JSON.parse(event.data) as ResponseMessage;
        setLogs((prev) => [...prev, logLine]);
      };

      ws.onerror = () => {
        setError('WebSocket connection error');
        setIsConnected(false);
      };

      ws.onclose = () => {
        setIsConnected(false);
      };
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to connect');
    }

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [serverId]);

  useEffect(() => {
    // Auto-scroll to bottom when new logs arrive
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [logs]);

  const submit = () => {
    const payload = {
      type: "command",
      command: command
    } satisfies CommandMessage
    wsRef.current?.send(JSON.stringify(payload))
  }

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Terminal className="h-5 w-5" />
              Server Logs
            </CardTitle>
            <CardDescription>
              Real-time logs from the Minecraft server
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            {isConnected ? (
              <div className="flex items-center gap-2 text-sm text-green-600">
                <div className="h-2 w-2 rounded-full bg-green-600 animate-pulse" />
                Connected
              </div>
            ) : error ? (
              <div className="text-sm text-destructive">{error}</div>
            ) : (
              <div className="flex items-center gap-2 text-sm text-muted-foreground">
                <Loader2 className="h-4 w-4 animate-spin" />
                Connecting...
              </div>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent className='flex flex-col gap-2'>
        <div className="rounded-md bg-black p-4 font-mono text-sm h-[500px] overflow-y-auto">
          {logs.length === 0 ? (
            <div className="text-gray-500 flex items-center justify-center h-full">
              {isConnected ? 'Waiting for logs...' : 'No logs available'}
            </div>
          ) : (
            <div className="space-y-1">
              {logs.map((log, index) => (
                <div key={index} className="text-green-400 whitespace-pre-wrap break-words">
                  {log.content}
                </div>
              ))}
              <div ref={logsEndRef} />
            </div>
          )}
        </div>
        <div className='flex gap-2'>
          <Input value={command} onChange={e => setCommand(e.currentTarget.value)} />
          <Button onClick={() => submit()}>Submit</Button>
        </div>
      </CardContent>
    </Card>
  );
}
