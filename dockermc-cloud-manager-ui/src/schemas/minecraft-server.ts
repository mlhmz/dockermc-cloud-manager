import { z } from 'zod';
import { containerStatusSchema } from './api';

/**
 * Minecraft Server entity schema - represents a running Minecraft server instance
 */
export const minecraftServerSchema = z.object({
  id: z.uuid(),
  name: z.string(),
  container_id: z.string(),
  volume_id: z.string(),
  status: containerStatusSchema,
  port: z.number(),
  max_players: z.number().int(),
  motd: z.string(),
  created_at: z.string(),
  updated_at: z.string(),
});
export type MinecraftServer = z.infer<typeof minecraftServerSchema>;

/**
 * Create Server Request schema - validates input for creating a new Minecraft server
 * Matches OpenAPI spec validation rules
 */
export const createServerRequestSchema = z.object({
  name: z
    .string()
    .min(1, 'Name is required')
    .max(100, 'Name must be less than 100 characters')
    .regex(/^[a-zA-Z0-9-_]+$/, 'Name can only contain letters, numbers, hyphens, and underscores'),
  max_players: z
    .number()
    .int()
    .min(1, 'Minimum 1 player')
    .max(1000, 'Maximum 1000 players')
    .optional(),
  motd: z
    .string()
    .max(255, 'MOTD must be less than 255 characters')
    .optional(),
  version: z
    .string()
    .optional(),
});
export type CreateServerRequest = z.infer<typeof createServerRequestSchema>;

/**
 * Update Server Request schema - validates input for updating an existing Minecraft server
 */
export const updateServerRequestSchema = z.object({
  max_players: z
    .number()
    .int()
    .min(1, 'Minimum 1 player')
    .max(1000, 'Maximum 1000 players')
    .optional(),
  motd: z
    .string()
    .max(255, 'MOTD must be less than 255 characters')
    .optional(),
});
export type UpdateServerRequest = z.infer<typeof updateServerRequestSchema>;
