import type { ContainerStatus } from "./api";

export interface MinecraftServer {
  id: string;
  name: string;
  container_id: string;
  volume_id: string;
  status: ContainerStatus;
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