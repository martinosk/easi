# Frontend Architecture

React 19 + TypeScript + Vite application using Zustand for state management and React Flow for canvas visualization.

## Core Principles

**API-First Design**: All data operations go through REST API calls to the backend. The frontend is a pure presentation layer with no business logic.

**Feature-Based Organization**: Code is organized by feature/domain, not by technical type. Each feature is self-contained with its own API functions, components, and hooks.

**State Management**:
- Global state (Zustand) for API data, selection state, and app-wide UI state
- Local state for component-specific UI (form inputs, dialog visibility)
- Store is split into domain slices that are composed together

## Directory Structure

```
src/
├── api/                    # API client singleton and type definitions
├── components/            # Shared UI components
│   ├── canvas/           # Canvas nodes and edges
│   ├── layout/           # Layout components
│   └── shared/           # Reusable UI elements
├── contexts/             # Cross-cutting concerns (e.g., release notes)
├── features/             # Feature modules (business domains)
│   ├── canvas/          # Canvas visualization and interactions
│   ├── capabilities/    # Capability management
│   ├── components/      # Component (system) management
│   ├── navigation/      # Navigation tree
│   ├── relations/       # Relation management
│   └── views/           # View management
├── hooks/                # Shared custom React hooks
├── store/                # Global Zustand store
│   ├── slices/          # Store slices by domain
│   └── appStore.ts      # Combined store
└── test/                 # Test utilities and setup
```

## Key Patterns

**Feature Modules**: Each feature has optional `api/`, `components/`, and `hooks/` subdirectories plus an `index.ts` for public exports.

**Zustand Store Slices**: Each domain has its own slice (state + actions). Slices are composed in `appStore.ts`. Use fine-grained selectors to minimize re-renders.

**Custom Hooks**: Extract complex logic from components into hooks. Common patterns: `useXxxManagement` (orchestration), `useXxxOperations` (business operations), `useXxxState` (state management).

**API Client**: All API calls go through the `apiClient` singleton in `api/client.ts` for centralized error handling and consistent interceptors.

**TypeScript**: Types are colocated with their usage (`api/types.ts` for API types, `store/types/` for store types, component files for component-specific types).

## Development Guidelines

**Components**:
- Keep small and focused
- Extract complex logic into custom hooks
- Use TypeScript for all props and state
- Use `React.memo` only when proven necessary

**Adding a Feature**:
1. Create directory under `src/features/`
2. Add `components/` subdirectory
3. Create store slice in `src/store/slices/` if needed
4. Add slice to `appStore.ts`
5. Add API methods to `apiClient` if needed

**Adding API Endpoints**:
1. Add method to `ApiClient` class in `api/client.ts`
2. Add types to `api/types.ts`
3. Use in store actions or components

**File Naming**:
- Components: `PascalCase.tsx`
- Hooks: `camelCase.ts` with `use` prefix
- Test files: `*.test.ts` or `*.test.tsx`

## Testing

```bash
npm test -- --run                          # All unit tests (use 3 min timeout)
npm test -- --run src/path/to/file.test.ts # Specific test
npm run test:e2e                           # E2E tests
```

Always use `--run` flag with Vitest to avoid watch mode.

## Development

```bash
npm install                # Install dependencies
npm run dev                # Start dev server (http://localhost:5173)
npm run build              # Build for production
```

Backend must be running (default: `http://localhost:8080`). Configure with `VITE_API_URL` environment variable.
