# Quick Start Guide - EASI Graphical Component Modeler

## What You Have

A complete graphical architecture modeling tool with:
- **Backend**: Go-based DDD/CQRS/Event Sourcing API
- **Frontend**: React + TypeScript + React Flow canvas

## Quick Start (5 minutes)

### 1. Start the Backend

```bash
# SFrom the easi directory: Start PostgreSQL and backend (if not already running)
docker-compose up -d
```

You should see:
```
Starting server on :8080
```

### 2. Start the Frontend

```bash
# Terminal 2 - From the easi directory
cd frontend

# Start the dev server
npm run dev
```

You should see:
```
  VITE v7.1.7  ready in XXX ms

  ➜  Local:   http://localhost:5173/
```

### 3. Open the Application

Open your browser to: **http://localhost:5173**

## Using the Application

### Create Your First Component

1. Click the **"Add Component"** button in the toolbar
2. Enter a name (e.g., "User Service")
3. Optionally add a description
4. Click **"Create"**
5. The component appears on the canvas!

### Move Components Around

1. Click and drag any component
2. Position is automatically saved to the backend
3. Refresh the page - your positions are preserved!

### Create a Relation

1. Hover over a component - you'll see connection points (small circles)
2. Click and drag from one component to another
3. A dialog opens - select the relation type:
   - **Triggers** (orange arrow) - One component triggers another
   - **Serves** (blue arrow) - One component serves another
4. Optionally add a name and description
5. Click **"Create"**
6. The relation arrow appears connecting your components!

### View Details

1. **Click on a component** - Side panel shows:
   - Name and description
   - Created date
   - **ArchiMate Documentation link** (click to learn more)

2. **Click on a relation arrow** - Side panel shows:
   - Relation type and name
   - Source and target components
   - **ArchiMate Documentation link** for the relation type

### Canvas Controls

- **Mouse wheel** - Zoom in/out
- **Click and drag background** - Pan the canvas
- **Minimap** (bottom right) - Overview of your diagram
- **Controls** (bottom left) - Fit view, zoom controls

### API Features

The backend provides two main bounded contexts:

#### Architecture Modeling
- **Components** - `POST/GET/PUT/DELETE /api/components`
- **Relations** - `POST/GET/DELETE /api/relations`
- **Views** - `POST/GET/DELETE /api/views`

#### Capability Mapping
- **Capabilities** - `POST/GET/PUT /api/capabilities`
  - Hierarchical business capabilities (L1-L4)
  - Metadata (strategy pillars, maturity, ownership)
  - Experts and tags
- **Dependencies** - `POST/GET/DELETE /api/capability-dependencies`
  - Model capability relationships (Requires, Enables, Supports)

### Swagger Documentation
- Open: http://localhost:8080/swagger/

