import type { ContainerStatus } from "./api";

export interface ProxyServer {
  id: string;
  name: string;
  container_id: string;
  volume_id: string;
  default_server_id?: string;
  status: ContainerStatus;
  port: number;
  created_at: string;
  updated_at: string;
}

export interface UpdateProxyRequest {
  default_server_id?: string;
}
