import { z } from 'zod';
import type {
  MinecraftServer,
  CreateServerRequest,
} from '../schemas/minecraft-server';
import type {
  ProxyServer,
  UpdateProxyRequest,
} from '../schemas/proxy';
import {
  minecraftServerSchema,
  proxyServerSchema,
} from '../schemas';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

class ApiClient {
  /**
   * Fetches JSON data and validates it against a Zod schema using safeParse
   * @param path - API endpoint path
   * @param schema - Zod schema for validation (optional for queries, required for mutations)
   * @param options - Fetch options
   * @returns Parsed and validated data
   * @throws Error if validation fails or HTTP error occurs
   */
  private async fetchJson<T>(
    path: string,
    schema?: z.ZodSchema<T>,
    options?: RequestInit
  ): Promise<T> {
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

    const data = await response.json();

    // Validate with Zod schema if provided
    if (schema) {
      const result = schema.safeParse(data);
      if (!result.success) {
        console.error('API response validation failed:', result.error);
        throw new Error(`Invalid API response: ${result.error.message}`);
      }
      return result.data;
    }

    return data;
  }

  // Servers
  async listServers(): Promise<MinecraftServer[]> {
    return this.fetchJson<MinecraftServer[]>(
      '/api/v1/servers',
      z.array(minecraftServerSchema)
    );
  }

  async getServer(id: string): Promise<MinecraftServer> {
    return this.fetchJson<MinecraftServer>(
      `/api/v1/servers/${id}`,
      minecraftServerSchema
    );
  }

  async createServer(request: CreateServerRequest): Promise<MinecraftServer> {
    return this.fetchJson<MinecraftServer>(
      '/api/v1/servers',
      minecraftServerSchema,
      {
        method: 'POST',
        body: JSON.stringify(request),
      }
    );
  }

  async deleteServer(id: string): Promise<void> {
    return this.fetchJson<void>(`/api/v1/servers/${id}`, undefined, {
      method: 'DELETE',
    });
  }

  async startServer(id: string): Promise<{ status: string }> {
    return this.fetchJson<{ status: string }>(
      `/api/v1/servers/${id}/start`,
      z.object({ status: z.string() }),
      {
        method: 'POST',
      }
    );
  }

  async stopServer(id: string): Promise<{ status: string }> {
    return this.fetchJson<{ status: string }>(
      `/api/v1/servers/${id}/stop`,
      z.object({ status: z.string() }),
      {
        method: 'POST',
      }
    );
  }

  // Proxy
  async getProxy(): Promise<ProxyServer> {
    return this.fetchJson<ProxyServer>(
      '/api/v1/proxy',
      proxyServerSchema
    );
  }

  async updateProxy(request: UpdateProxyRequest): Promise<ProxyServer> {
    return this.fetchJson<ProxyServer>(
      '/api/v1/proxy',
      proxyServerSchema,
      {
        method: 'PATCH',
        body: JSON.stringify(request),
      }
    );
  }

  async startProxy(): Promise<ProxyServer> {
    return this.fetchJson<ProxyServer>(
      '/api/v1/proxy/start',
      proxyServerSchema,
      {
        method: 'POST',
      }
    );
  }

  async stopProxy(): Promise<{ message: string }> {
    return this.fetchJson<{ message: string }>(
      '/api/v1/proxy/stop',
      z.object({ message: z.string() }),
      {
        method: 'POST',
      }
    );
  }

  async regenerateProxyConfig(): Promise<{ message: string }> {
    return this.fetchJson<{ message: string }>(
      '/api/v1/proxy/regenerate-config',
      z.object({ message: z.string() }),
      {
        method: 'POST',
      }
    );
  }

  // WebSocket for logs
  createLogWebSocket(serverId: string, follow = true, tail = '100'): WebSocket {
    const wsUrl = API_BASE_URL.replace(/^http/, 'ws');
    return new WebSocket(`${wsUrl}/api/v1/servers/${serverId}/logs?follow=${follow}&tail=${tail}`);
  }
}

export const apiClient = new ApiClient();
