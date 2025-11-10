// Generated from OpenAPI spec

export type ServerStatus = 'creating' | 'running' | 'stopped' | 'error';

export interface MinecraftServer {
  id: string;
  name: string;
  container_id: string;
  volume_id: string;
  status: ServerStatus;
  port?: number;
  max_players: number;
  motd: string;
  created_at: string;
  updated_at: string;
}

export interface CreateServerRequest {
  name: string;
  max_players?: number;
  motd?: string;
  version?: string;
}

export interface UpdateServerRequest {
  max_players?: number;
  motd?: string;
}

export interface ProxyServer {
  id: string;
  name: string;
  container_id: string;
  volume_id: string;
  default_server_id?: string;
  status: ServerStatus;
  port: number;
  created_at: string;
  updated_at: string;
}

export interface UpdateProxyRequest {
  default_server_id?: string;
}

export interface ApiError {
  error: string;
}

export interface HealthResponse {
  status: string;
}
