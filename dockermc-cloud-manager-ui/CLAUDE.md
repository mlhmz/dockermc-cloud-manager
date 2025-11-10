# DockerMC Cloud Manager UI - Frontend Documentation

## Project Overview

This is the frontend application for DockerMC Cloud Manager, a web-based management interface for Minecraft servers and Velocity proxy instances running in Docker containers. Built with React, TypeScript, and modern web technologies.

**Tech Stack:**
- React 18 with TypeScript
- Vite (build tool)
- TanStack Query (React Query) for data fetching
- React Router for navigation
- React Hook Form + Zod for form validation
- Tailwind CSS + shadcn/ui components
- WebSocket for real-time server logs

## Architecture Overview

### Directory Structure

```
src/
├── api/
│   └── client.ts              # API client with Zod validation
├── components/
│   ├── ui/                    # shadcn/ui base components
│   ├── create-server-dialog.tsx
│   ├── server-logs.tsx
│   └── status-badge.tsx
├── hooks/
│   ├── queries/               # React Query hooks for data fetching
│   │   ├── useServers.ts
│   │   ├── useServer.ts
│   │   └── useProxy.ts
│   └── mutations/             # React Query hooks for mutations
│       ├── useCreateServer.ts
│       ├── useDeleteServer.ts
│       ├── useStartServer.ts
│       ├── useStopServer.ts
│       ├── useUpdateProxy.ts
│       ├── useStartProxy.ts
│       ├── useStopProxy.ts
│       └── useRegenerateProxyConfig.ts
├── routes/                    # Route components/pages
│   ├── proxy.tsx
│   ├── server-detail.tsx
│   └── servers.tsx
├── schemas/                   # Zod schemas (source of truth for types)
│   ├── api.ts
│   ├── minecraft-server.ts
│   ├── proxy.ts
│   └── index.ts
├── lib/
│   └── utils.ts
└── router.tsx                 # React Router configuration
```

## Core Patterns and Conventions

### 1. Type System: Zod Schemas as Source of Truth

**IMPORTANT:** All types are defined as Zod schemas and TypeScript types are inferred from them.

**Location:** `src/schemas/`

**Pattern:**
```typescript
// Define schema
export const minecraftServerSchema = z.object({
  id: z.uuid(),
  name: z.string(),
  status: containerStatusSchema,
  // ...
});

// Infer TypeScript type
export type MinecraftServer = z.infer<typeof minecraftServerSchema>;
```

**Why:**
- Single source of truth for types
- Runtime validation with `safeParse()`
- Form validation with same schemas
- Type safety + runtime safety

**Importing:**
```typescript
// Always import from @/schemas
import { minecraftServerSchema, type MinecraftServer } from '@/schemas';
```

### 2. API Client: Validated Data Fetching

**Location:** `src/api/client.ts`

**Key Feature:** All API responses are validated with Zod `safeParse()`

**Pattern:**
```typescript
async listServers(): Promise<MinecraftServer[]> {
  return this.fetchJson<MinecraftServer[]>(
    '/api/v1/servers',
    z.array(minecraftServerSchema)  // Schema for validation
  );
}
```

**How it works:**
1. Fetch data from API
2. Parse JSON response
3. Validate with `schema.safeParse(data)`
4. Throw error if validation fails
5. Return validated, type-safe data

**Error handling:** Validation errors are logged to console and thrown with descriptive messages.

### 3. Data Fetching: React Query Hooks

**Location:** `src/hooks/queries/`

**Pattern:**
```typescript
// useServers.ts
export const serversQueryKey = ['servers'];

export function useServers() {
  return useQuery({
    queryKey: serversQueryKey,
    queryFn: () => apiClient.listServers(),
    refetchInterval: 5000,  // Auto-refresh every 5s
  });
}
```

**Usage in components:**
```typescript
const { data: servers, isLoading, error } = useServers();
```

**Conventions:**
- Export query keys for invalidation
- Use descriptive hook names (useServers, useServer, useProxy)
- Enable auto-refetch for real-time data (refetchInterval: 5000ms)

### 4. Mutations: React Query Mutation Hooks

**Location:** `src/hooks/mutations/`

**Pattern:**
```typescript
export function useCreateServer() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (request: CreateServerRequest) =>
      apiClient.createServer(request),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: serversQueryKey });
    },
  });
}
```

**Key Features:**
- Automatic query invalidation on success
- Optimistic updates where appropriate
- Error handling via mutation state

**Usage:**
```typescript
const createServer = useCreateServer();

await createServer.mutateAsync(data);

// Or with callbacks
createServer.mutate(data, {
  onSuccess: () => { /* ... */ },
  onError: (error) => { /* ... */ }
});
```

### 5. Forms: React Hook Form + Zod

**Pattern:**
```typescript
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { createServerRequestSchema, type CreateServerRequest } from '@/schemas';

const form = useForm<CreateServerRequest>({
  resolver: zodResolver(createServerRequestSchema),
  defaultValues: {
    name: '',
    max_players: 20,
    // ...
  },
});

const onSubmit = async (data: CreateServerRequest) => {
  await createServer.mutateAsync(data);
};
```

**Important:**
- Use centralized schemas from `@/schemas`
- Never define schemas inline (except for one-off validation)
- Use `zodResolver` for validation
- Type form values with inferred types

### 6. Real-time Updates

**Auto-refresh:** Most queries use `refetchInterval: 5000` for 5-second polling

**WebSocket for logs:**
```typescript
// API client provides WebSocket factory
const ws = apiClient.createLogWebSocket(serverId, true, '100');

ws.onmessage = (event) => {
  const log = JSON.parse(event.data);
  // Handle log message
};
```

### 7. UI Components: shadcn/ui

**Location:** `src/components/ui/`

**Conventions:**
- Use shadcn/ui components as base
- Tailwind CSS for styling
- Dark mode support via Tailwind
- Responsive design (mobile-first)

**Common patterns:**
```typescript
import { Button } from '@/components/ui/button';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
```

### 8. Status Badge Component

**Usage:**
```typescript
import { StatusBadge } from '@/components/status-badge';

<StatusBadge status={server.status} />
```

**Status values:** `'creating' | 'running' | 'stopped' | 'error'`

**Colors:**
- `creating` - Blue
- `running` - Green
- `stopped` - Gray
- `error` - Red

## Data Models

### Container Status
```typescript
type ContainerStatus = 'creating' | 'running' | 'stopped' | 'error';
```

### Minecraft Server
```typescript
interface MinecraftServer {
  id: string;              // UUID
  name: string;
  container_id: string;
  volume_id: string;
  status: ContainerStatus;
  port?: number;           // Optional, assigned when running
  max_players: number;
  motd: string;
  created_at: string;      // ISO 8601 datetime
  updated_at: string;
}
```

### Proxy Server
```typescript
interface ProxyServer {
  id: string;                    // UUID
  name: string;
  container_id: string;
  volume_id: string;
  default_server_id?: string;    // UUID, optional
  status: ContainerStatus;
  port: number;
  created_at: string;
  updated_at: string;
}
```

### Request Types
```typescript
// Create Server
interface CreateServerRequest {
  name: string;           // 1-100 chars, alphanumeric + hyphens/underscores
  max_players?: number;   // 1-1000, default: 20
  motd?: string;          // Max 255 chars
  version?: string;       // e.g., "1.20.1" or "LATEST"
}

// Update Server (not yet used in UI)
interface UpdateServerRequest {
  max_players?: number;   // 1-1000
  motd?: string;          // Max 255 chars
}

// Update Proxy
interface UpdateProxyRequest {
  default_server_id?: string;  // UUID
}
```

## API Endpoints

**Base URL:** `http://localhost:8080` (configurable via `VITE_API_URL`)

**Servers:**
- `GET /api/v1/servers` - List all servers
- `GET /api/v1/servers/:id` - Get server by ID
- `POST /api/v1/servers` - Create server
- `DELETE /api/v1/servers/:id` - Delete server
- `POST /api/v1/servers/:id/start` - Start server
- `POST /api/v1/servers/:id/stop` - Stop server
- `WS /api/v1/servers/:id/logs` - Stream logs (WebSocket)

**Proxy:**
- `GET /api/v1/proxy` - Get proxy info
- `PATCH /api/v1/proxy` - Update proxy config
- `POST /api/v1/proxy/start` - Start proxy
- `POST /api/v1/proxy/stop` - Stop proxy
- `POST /api/v1/proxy/regenerate-config` - Regenerate Velocity config

## Routing

**Routes:**
- `/` - Servers list page
- `/servers/:id` - Server detail page (logs)
- `/proxy` - Proxy management page

**Navigation:**
```typescript
import { Link } from 'react-router-dom';

<Link to={`/servers/${server.id}`}>View Logs</Link>
```

## Development Guidelines

### Adding a New Feature

1. **Define Schema** (if new data type):
   ```typescript
   // src/schemas/your-feature.ts
   export const yourFeatureSchema = z.object({ /* ... */ });
   export type YourFeature = z.infer<typeof yourFeatureSchema>;
   ```

2. **Add API Method**:
   ```typescript
   // src/api/client.ts
   async getYourFeature(): Promise<YourFeature> {
     return this.fetchJson<YourFeature>(
       '/api/v1/your-feature',
       yourFeatureSchema
     );
   }
   ```

3. **Create Query Hook**:
   ```typescript
   // src/hooks/queries/useYourFeature.ts
   export const yourFeatureQueryKey = ['your-feature'];

   export function useYourFeature() {
     return useQuery({
       queryKey: yourFeatureQueryKey,
       queryFn: () => apiClient.getYourFeature(),
     });
   }
   ```

4. **Use in Component**:
   ```typescript
   const { data, isLoading, error } = useYourFeature();
   ```

### Adding a Form

1. **Ensure schema exists** in `src/schemas/`
2. **Create mutation hook** in `src/hooks/mutations/`
3. **Build form component** with React Hook Form:
   ```typescript
   const form = useForm<YourRequestType>({
     resolver: zodResolver(yourRequestSchema),
     defaultValues: { /* ... */ },
   });
   ```

### Validation Rules

All validation is defined in Zod schemas and matches the backend OpenAPI spec:

- Server names: 1-100 chars, `^[a-zA-Z0-9-_]+$`
- Max players: 1-1000 (integer)
- MOTD: Max 255 chars
- UUIDs: Validated with `z.uuid()`
- Datetimes: ISO 8601 strings

## Common Tasks

### Adding a New Query
1. Add API method to `client.ts` with schema
2. Create hook in `hooks/queries/`
3. Export query key for invalidation

### Adding a New Mutation
1. Add API method to `client.ts` with schema
2. Create hook in `hooks/mutations/`
3. Invalidate related queries in `onSuccess`

### Updating a Schema
1. Modify schema in `src/schemas/`
2. Types auto-update via inference
3. API validation auto-updates
4. Forms auto-update validation

### Adding UI Components
1. Use shadcn/ui CLI: `npx shadcn-ui@latest add <component>`
2. Components added to `src/components/ui/`
3. Customize with Tailwind classes

## Environment Variables

```env
VITE_API_URL=http://localhost:8080
```

## Build and Development

```bash
# Install dependencies
npm install

# Development server
npm run dev

# Type check
npm run build  # Runs tsc -b && vite build

# Preview production build
npm run preview
```

## Testing Strategy

Currently no automated tests. Suggested additions:
- Unit tests for schemas (Zod validation)
- Integration tests for API client
- Component tests with React Testing Library
- E2E tests with Cypress

## Known Limitations

1. **No UpdateServerRequest UI** - Schema exists but no form implemented yet
2. **WebSocket reconnection** - No automatic reconnection logic for logs
3. **Error boundaries** - No React error boundaries implemented
4. **Optimistic updates** - Not implemented for mutations
5. **Pagination** - No pagination for server lists (assumes small scale)

## Future Enhancements

- [ ] Add UpdateServerRequest form (edit server settings)
- [ ] Implement optimistic updates for better UX
- [ ] Add React error boundaries
- [ ] WebSocket auto-reconnect with exponential backoff
- [ ] Server metrics and monitoring
- [ ] Bulk operations (start/stop multiple servers)
- [ ] Search and filter for server lists
- [ ] Dark mode toggle (currently system preference only)

## Backend Integration

This frontend expects a Go backend at `/Users/malek/Code/dockermc-cloud-manager/` with:
- OpenAPI spec at `api/openapi.yaml`
- Models defined in `internal/models/`
- Handlers in `internal/handlers/`

The frontend schemas are designed to match the OpenAPI spec exactly.

## Getting Help

- **API Docs:** See `../api/openapi.yaml`
- **Backend Code:** See `../internal/`
- **shadcn/ui Docs:** https://ui.shadcn.com/
- **React Query Docs:** https://tanstack.com/query/latest
- **Zod Docs:** https://zod.dev/

## Key Files to Reference

- `src/schemas/index.ts` - All type definitions
- `src/api/client.ts` - API integration
- `src/router.tsx` - Route configuration
- `tailwind.config.js` - Tailwind customization
- `vite.config.ts` - Build configuration
