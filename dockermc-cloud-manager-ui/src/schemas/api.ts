import { z } from 'zod';

/**
 * Container status enum - represents the current state of a Docker container
 */
export const containerStatusSchema = z.enum(['creating', 'running', 'stopped', 'error']);
export type ContainerStatus = z.infer<typeof containerStatusSchema>;

/**
 * API error response schema
 */
export const apiErrorSchema = z.object({
  error: z.string(),
});
export type ApiError = z.infer<typeof apiErrorSchema>;

/**
 * Health check response schema
 */
export const healthResponseSchema = z.object({
  status: z.string(),
});
export type HealthResponse = z.infer<typeof healthResponseSchema>;
