# Component Modeler - Frontend

A modern React + TypeScript + Vite application for graphical component modeling using ArchiMate principles.

## Features

- Visual component modeling with React Flow
- Create and manage application components
- Define relationships between components (Triggers, Serves)
- Drag and drop interface with persistent positioning
- Real-time API integration with backend
- HATEOAS navigation with ArchiMate documentation links
- Professional UI with responsive design

## Tech Stack

- **React 18** - UI library
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server
- **React Flow** - Interactive canvas and node editor
- **Zustand** - State management
- **Axios** - HTTP client
- **React Hot Toast** - Toast notifications

## Prerequisites

- Node.js 18+ installed
- Backend server running at `http://localhost:8080`

## Installation

```bash
# Install dependencies
npm install
```

## Development

```bash
# Start the development server
npm run dev

# The app will be available at http://localhost:5173
```

The development server includes:
- Hot Module Replacement (HMR)
- Fast refresh for React components
- TypeScript type checking
- ESLint integration

## Building for Production

```bash
# Build the application
npm run build

# Preview the production build
npm run preview
```

## Project Structure

```
src/
├── api/
│   ├── client.ts          # API client with axios
│   └── types.ts           # TypeScript interfaces for API
├── components/
│   ├── ComponentCanvas.tsx     # Main React Flow canvas
│   ├── CreateComponentDialog.tsx
│   ├── CreateRelationDialog.tsx
│   ├── ComponentDetails.tsx
│   ├── RelationDetails.tsx
│   └── Toolbar.tsx
├── store/
│   └── appStore.ts        # Zustand state management
├── App.tsx                # Main app component
├── main.tsx               # Entry point
└── index.css              # Global styles
```

## Key Features

### Component Canvas

The main canvas uses React Flow to provide:
- Interactive node-based interface
- Drag and drop positioning
- Automatic edge routing
- Zoom and pan controls
- Mini-map navigation
- Custom styled nodes and edges

### Component Management

- Create new components with name and description
- Automatically positioned on canvas
- View detailed component information
- Access ArchiMate documentation links

### Relation Management

- Create relations between components
- Two relation types: Triggers (orange) and Serves (blue)
- Visual connection by dragging between nodes
- Optional relation names and descriptions
- Color-coded edges based on type

### Position Persistence

- Component positions automatically saved to backend
- Positions restored on page reload
- Real-time updates via API calls
- View-based positioning system

## API Integration

The frontend communicates with the backend REST API:

### Endpoints Used

- `GET /api/v1/components` - List all components
- `POST /api/v1/components` - Create component
- `GET /api/v1/relations` - List all relations
- `POST /api/v1/relations` - Create relation
- `GET /api/v1/views` - List all views
- `POST /api/v1/views` - Create view
- `GET /api/v1/views/{id}` - Get view with positions
- `POST /api/v1/views/{id}/components` - Add component to view
- `PATCH /api/v1/views/{id}/components/{componentId}/position` - Update position

### Error Handling

The application includes comprehensive error handling:
- Network errors displayed as toast notifications
- Loading states for async operations
- Graceful fallbacks for missing data
- Retry mechanisms for failed requests

## State Management

Uses Zustand for simple, performant state management:

```typescript
const {
  components,      // All components
  relations,       // All relations
  currentView,     // Active view with positions
  selectedNodeId,  // Currently selected component
  selectedEdgeId,  // Currently selected relation
  loadData,        // Load all data from API
  createComponent, // Create new component
  createRelation,  // Create new relation
  updatePosition,  // Update component position
} = useAppStore();
```

## Styling

The application uses a modern, professional design system:
- CSS custom properties for theming
- Blue/purple gradient for components
- Color-coded relations (Triggers: orange, Serves: blue)
- Responsive layout with flexbox/grid
- Smooth animations and transitions
- Mobile-friendly design

## Configuration

### API Base URL

To change the backend URL, modify the API client:

```typescript
// src/api/client.ts
const apiClient = new ApiClient('http://your-backend-url');
```

### React Flow Settings

Customize canvas behavior in `ComponentCanvas.tsx`:

```typescript
<ReactFlow
  minZoom={0.1}      // Minimum zoom level
  maxZoom={2}        // Maximum zoom level
  fitView            // Fit all nodes on load
  // ... other options
/>
```

## TypeScript

The project uses strict TypeScript configuration:
- No `any` types allowed
- Full type coverage for API responses
- Interface-based type definitions
- Compile-time error checking

## Browser Support

- Chrome/Edge (last 2 versions)
- Firefox (last 2 versions)
- Safari (last 2 versions)

## Troubleshooting

### Backend Connection Issues

If the app can't connect to the backend:

1. Verify backend is running: `curl http://localhost:8080/api/v1/components`
2. Check CORS is enabled on backend
3. Verify API base URL in `src/api/client.ts`

### Build Errors

```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install

# Clear Vite cache
rm -rf node_modules/.vite
npm run dev
```

## License

MIT
