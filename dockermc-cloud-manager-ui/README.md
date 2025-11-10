# DockerMC Cloud Manager UI

Modern web interface for managing Minecraft servers running in Docker containers.

## Features

- **Server Management**: Create, start, stop, and delete Minecraft servers
- **Real-time Logs**: WebSocket-based log streaming for each server
- **Proxy Management**: Control the Velocity proxy and configure default lobby servers
- **Auto-refresh**: Server status updates automatically every 5 seconds
- **Responsive Design**: Built with Tailwind CSS and shadcn/ui components

## Tech Stack

- **React 19** - UI library with React Compiler support
- **React Router v7** - Client-side routing
- **TanStack Query** - Data fetching and caching
- **shadcn/ui** - Beautiful, accessible UI components
- **Tailwind CSS v4** - Utility-first CSS framework
- **Zod** - Schema validation
- **React Hook Form** - Form state management
- **TypeScript** - Type safety

## Getting Started

### Prerequisites

- Node.js 18+ and npm
- Backend API running (default: `http://localhost:8080`)

### Installation

```bash
# Install dependencies
npm install

# Copy environment variables
cp .env.example .env

# Update the API URL in .env if needed
# VITE_API_URL=http://localhost:8080
```

### Development

```bash
# Start the development server
npm run dev
```

The app will be available at [http://localhost:5173](http://localhost:5173)

### Build

```bash
# Build for production
npm run build

# Preview the production build
npm run preview
```

## Project Structure

```
src/
├── api/              # API client
├── components/       # React components
│   ├── ui/          # shadcn/ui components
│   └── ...          # Custom components
├── hooks/           # Custom React hooks
│   ├── queries/     # TanStack Query hooks (one per query)
│   └── mutations/   # TanStack Query mutations (one per mutation)
├── lib/             # Utility functions
├── routes/          # Page components
└── types/           # TypeScript types (generated from OpenAPI spec)
```

## API Integration

The UI communicates with the backend API defined in `/api/openapi.yaml`:

- **REST API**: For server/proxy management operations
- **WebSocket**: For real-time log streaming

All API calls are handled through:
- `src/api/client.ts` - API client implementation
- `src/hooks/queries/` - React Query hooks for data fetching
- `src/hooks/mutations/` - React Query hooks for mutations

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `VITE_API_URL` | Backend API URL | `http://localhost:8080` |

## Features Overview

### Server Management
- View all servers with status badges (running, stopped, creating, error)
- Create new servers with custom configuration
- Start/stop servers with one click
- Delete servers with confirmation
- View detailed server information
- Real-time status updates

### Log Viewer
- WebSocket-based log streaming
- Auto-scroll to latest logs
- Connection status indicator
- 100 most recent log lines on connection

### Proxy Management
- View proxy status and configuration
- Start/stop the Velocity proxy
- Configure default lobby server
- Regenerate proxy configuration
- See connected servers count

## Development Notes

- Uses Vite with Rolldown for faster builds
- React Compiler (babel-plugin-react-compiler) enabled for automatic optimizations
- TypeScript strict mode enabled
- Path aliases configured (`@/*` → `src/*`)
