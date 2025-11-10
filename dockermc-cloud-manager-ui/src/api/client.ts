import type {
  MinecraftServer,
  CreateServerRequest,
} from '../types/minecraft-server';
import type {
  ProxyServer,
  UpdateProxyRequest,
} from '../types/proxy';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

class ApiClient {
  private async fetchJson<T>(path: string, options?: RequestInit): Promise<T> {
    const response = await fetch(path, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(error.error || `HTTP ${response.status}`);
    }

    if (response.status === 204) {
      return undefined as T;
    }

    return response.json();
  }

  // Servers
  async listServers(): Promise<MinecraftServer[]> {
    return this.fetchJson<MinecraftServer[]>('/api/v1/servers');
  }

  async getServer(id: string): Promise<MinecraftServer> {
    return this.fetchJson<MinecraftServer>(`/api/v1/servers/${id}`);
  }

  async createServer(request: CreateServerRequest): Promise<MinecraftServer> {
    return this.fetchJson<MinecraftServer>('/api/v1/servers', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  async deleteServer(id: string): Promise<void> {
    return this.fetchJson<void>(`/api/v1/servers/${id}`, {
      method: 'DELETE',
    });
  }

  async startServer(id: string): Promise<{ status: string }> {
    return this.fetchJson<{ status: string }>(`/api/v1/servers/${id}/start`, {
      method: 'POST',
    });
  }

  async stopServer(id: string): Promise<{ status: string }> {
    return this.fetchJson<{ status: string }>(`/api/v1/servers/${id}/stop`, {
      method: 'POST',
    });
  }

  // Proxy
  async getProxy(): Promise<ProxyServer> {
    return this.fetchJson<ProxyServer>('/api/v1/proxy');
  }

  async updateProxy(request: UpdateProxyRequest): Promise<ProxyServer> {
    return this.fetchJson<ProxyServer>('/api/v1/proxy', {
      method: 'PATCH',
      body: JSON.stringify(request),
    });
  }

  async startProxy(): Promise<ProxyServer> {
    return this.fetchJson<ProxyServer>('/api/v1/proxy/start', {
      method: 'POST',
    });
  }

  async stopProxy(): Promise<{ message: string }> {
    return this.fetchJson<{ message: string }>('/api/v1/proxy/stop', {
      method: 'POST',
    });
  }

  async regenerateProxyConfig(): Promise<{ message: string }> {
    return this.fetchJson<{ message: string }>('/api/v1/proxy/regenerate-config', {
      method: 'POST',
    });
  }

  // WebSocket for logs
  createLogWebSocket(serverId: string, follow = true, tail = '100'): WebSocket {
    const wsUrl = API_BASE_URL.replace(/^http/, 'ws');
    return new WebSocket(`${wsUrl}/api/v1/servers/${serverId}/logs?follow=${follow}&tail=${tail}`);
  }
}

export const apiClient = new ApiClient();
