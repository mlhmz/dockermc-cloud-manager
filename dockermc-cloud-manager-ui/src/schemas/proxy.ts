import { z } from 'zod';
import { containerStatusSchema } from './api';

/**
 * Proxy Server entity schema - represents a running Velocity proxy instance
 */
export const proxyServerSchema = z.object({
  id: z.string(),
  name: z.string(),
  container_id: z.string(),
  volume_id: z.string(),
  default_server_id: z.uuid().or(z.string()).optional(),
  status: containerStatusSchema,
  port: z.number().int().positive(),
  created_at: z.string(),
  updated_at: z.string(),
});
export type ProxyServer = z.infer<typeof proxyServerSchema>;

/**
 * Update Proxy Request schema - validates input for updating proxy configuration
 */
export const updateProxyRequestSchema = z.object({
  default_server_id: z.uuid().optional(),
});
export type UpdateProxyRequest = z.infer<typeof updateProxyRequestSchema>;
