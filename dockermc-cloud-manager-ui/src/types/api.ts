// Generated from OpenAPI spec

export type ContainerStatus = 'creating' | 'running' | 'stopped' | 'error';

export interface ApiError {
  error: string;
}

export interface HealthResponse {
  status: string;
}
