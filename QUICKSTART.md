# Quick Start Guide - EASI Graphical Component Modeler

## What You Have

A complete graphical architecture modeling tool with:
- **Backend**: Go-based DDD/CQRS/Event Sourcing API
- **Frontend**: React + TypeScript + React Flow canvas

## Quick Start (5 minutes)

### 1. Start the Backend

```bash
# Terminal 1 - From the easi directory
cd /home/devuser/repos/easi

# Start PostgreSQL (if not already running)
docker-compose up -d

# Run the backend
cd backend
go run cmd/api/main.go
```

You should see:
```
Starting server on :8080
```

### 2. Start the Frontend

```bash
# Terminal 2 - From the easi directory
cd /home/devuser/repos/easi/frontend

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

## Example Scenario

Let's create a simple microservices architecture:

1. **Create Components**:
   - "API Gateway"
   - "User Service"
   - "Order Service"
   - "Database"

2. **Arrange them** visually (drag to position)

3. **Create Relations**:
   - API Gateway **Triggers** → User Service
   - API Gateway **Triggers** → Order Service
   - User Service **Triggers** → Database
   - Order Service **Triggers** → Database

4. **View the result**: A visual architecture diagram with:
   - Orange arrows showing trigger relationships
   - Blue arrows showing serving relationships
   - Clean, professional styling
   - Positions saved automatically

## API Endpoints

The backend exposes these REST APIs:

### Components
- `GET /api/v1/components` - List all components
- `POST /api/v1/components` - Create a component
- `GET /api/v1/components/{id}` - Get component details

### Relations
- `GET /api/v1/relations` - List all relations
- `POST /api/v1/relations` - Create a relation
- `GET /api/v1/relations/{id}` - Get relation details

### Views (Position Persistence)
- `GET /api/v1/views` - List all views
- `POST /api/v1/views` - Create a view
- `GET /api/v1/views/{id}` - Get view with component positions
- `POST /api/v1/views/{id}/components` - Add component to view
- `PATCH /api/v1/views/{id}/components/{id}/position` - Update position

### Swagger Documentation
- Open: http://localhost:8080/swagger/
- Interactive API documentation

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      Browser                            │
│  ┌─────────────────────────────────────────────────┐   │
│  │         React Frontend (Port 5173)              │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐     │   │
│  │  │ Canvas   │  │ Dialogs  │  │ Details  │     │   │
│  │  │ (React   │  │ (Create) │  │ (View)   │     │   │
│  │  │  Flow)   │  └──────────┘  └──────────┘     │   │
│  │  └────┬─────┘                                  │   │
│  │       │ API Client (Axios)                     │   │
│  └───────┼────────────────────────────────────────┘   │
│          │                                             │
└──────────┼─────────────────────────────────────────────┘
           │ HTTP/JSON
           │
┌──────────▼─────────────────────────────────────────────┐
│              Go Backend (Port 8080)                    │
│  ┌─────────────────────────────────────────────────┐  │
│  │            RESTful API Layer                    │  │
│  │  (Chi Router, CORS, Middleware)                 │  │
│  └────────┬────────────────────────────────────────┘  │
│           │                                            │
│  ┌────────▼────────────────────────────────────────┐  │
│  │         CQRS Command/Query Buses                │  │
│  └────────┬────────────────────────────────────────┘  │
│           │                                            │
│  ┌────────▼────────────────────────────────────────┐  │
│  │        Bounded Contexts (DDD)                   │  │
│  │  ┌──────────────┐   ┌──────────────┐           │  │
│  │  │Architecture  │   │Architecture  │           │  │
│  │  │Modeling      │   │Views         │           │  │
│  │  │(Components,  │   │(Positions)   │           │  │
│  │  │ Relations)   │   │              │           │  │
│  │  └──────┬───────┘   └──────┬───────┘           │  │
│  └─────────┼──────────────────┼───────────────────┘  │
│            │                  │                       │
│  ┌─────────▼──────────────────▼───────────────────┐  │
│  │         Event Store (PostgreSQL)               │  │
│  │  - All events (audit trail)                    │  │
│  │  - Event sourcing                              │  │
│  └────────────────────────────────────────────────┘  │
│                                                       │
│  ┌────────────────────────────────────────────────┐  │
│  │         Read Models (PostgreSQL)               │  │
│  │  - Components                                  │  │
│  │  - Relations                                   │  │
│  │  - Views + Positions                           │  │
│  └────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────┘
```

## Tech Stack

### Backend
- **Language**: Go 1.21+
- **Router**: Chi v5
- **Database**: PostgreSQL 13+
- **Patterns**: DDD, CQRS, Event Sourcing
- **API Style**: REST Level 3 (HATEOAS)
- **Documentation**: Swagger/OpenAPI

### Frontend
- **Framework**: React 19
- **Language**: TypeScript 5.9
- **Build Tool**: Vite 7
- **Canvas**: React Flow 12
- **State**: Zustand 5
- **HTTP**: Axios
- **Notifications**: react-hot-toast
- **Testing**: Vitest

## Troubleshooting

### Backend won't start
- Check PostgreSQL is running: `docker-compose ps`
- Check port 8080 is free: `lsof -i :8080`
- Check database connection in backend logs

### Frontend won't start
- Check Node.js version: `node --version` (need 18+)
- Install dependencies: `npm install`
- Check port 5173 is free

### Components not appearing
- Check browser console for errors (F12)
- Verify backend is running at http://localhost:8080
- Check network tab for API responses

### Can't create relations
- Ensure you have at least 2 components
- Make sure to drag from one component to another
- Check the dialog appears after connecting

### Positions not saving
- Check browser console for PATCH request errors
- Verify backend views API is working: http://localhost:8080/api/v1/views
- Check PostgreSQL is running

## Next Steps

1. **Explore the code**:
   - Backend: `/home/devuser/repos/easi/backend/internal/`
   - Frontend: `/home/devuser/repos/easi/frontend/src/`

2. **Read the documentation**:
   - Frontend README: `/home/devuser/repos/easi/frontend/README.md`
   - Implementation summary: `/home/devuser/repos/easi/SPEC_005_IMPLEMENTATION.md`

3. **Run tests**:
   - Backend: `cd backend && go test ./...`
   - Frontend: `cd frontend && npm test`

4. **Build for production**:
   - Backend: `cd backend && go build -o bin/api cmd/api/main.go`
   - Frontend: `cd frontend && npm run build`

## Learn More

- **ArchiMate**: https://pubs.opengroup.org/architecture/archimate3-doc/
- **React Flow**: https://reactflow.dev/
- **DDD**: Domain-Driven Design by Eric Evans
- **CQRS**: https://martinfowler.com/bliki/CQRS.html
- **Event Sourcing**: https://martinfowler.com/eaaDev/EventSourcing.html

## Support

For issues or questions:
1. Check the documentation in `/home/devuser/repos/easi/`
2. Review the implementation summary: `SPEC_005_IMPLEMENTATION.md`
3. Inspect browser console and backend logs
4. Check the spec: `specs/005_GraphicalComponentModeler_ongoing.md`

---

**Enjoy modeling your architecture! 🎨🏗️**
