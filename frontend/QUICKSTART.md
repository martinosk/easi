# Quick Start Guide

## Prerequisites

Ensure the backend is running at http://localhost:8080

```bash
# In the backend directory
cd /home/devuser/repos/easi/backend
go run main.go
```

## Start the Frontend

```bash
# In the frontend directory
cd /home/devuser/repos/easi/frontend

# Install dependencies (if not already done)
npm install

# Start the development server
npm run dev
```

The application will be available at: **http://localhost:5173**

## First Steps

### 1. Create Your First Component

1. Click the **"+ Add Component"** button in the toolbar
2. Enter a name (e.g., "User Service")
3. Optionally add a description
4. Click **"Create Component"**
5. The component appears on the canvas

### 2. Create More Components

Create a few more components:
- "Authentication API"
- "Database"
- "Frontend App"
- "Email Service"

### 3. Create Relations

**Option A: Drag to Connect**
1. Click and drag from one component to another
2. A dialog opens with source and target pre-filled
3. Select relation type (Triggers or Serves)
4. Optionally add name and description
5. Click "Create Relation"

**Option B: Manual (future feature)**
Currently only drag-to-connect is supported.

### 4. Working with Capabilities

The sidebar includes a **Capabilities** section showing your business capability hierarchy.

**View Capabilities:**
1. Look for the **Capabilities** section in the left sidebar
2. Click the expand/collapse arrows to navigate the hierarchy
3. Capabilities are organized by level (L1, L2, L3, L4)
4. Color indicators show maturity level

**Create a New Capability:**
1. Click the **"+"** button next to "Capabilities" in the sidebar
2. A dialog opens with the following fields:
   - **Name** (required): 1-200 characters
   - **Description** (optional): Up to 1000 characters
   - **Status**: Active (default), Planned, Deprecated, or Retired
   - **Maturity Level**: Genesis (default), Custom Build, Product, or Commodity
3. Click **"Create"** to add the capability
4. New capabilities start at L1 (root level)
5. The tree automatically refreshes to show your new capability

**Maturity Levels (Wardley Map Evolution):**
- **Genesis**: Novel, uncertain, rapidly changing
- **Custom Build**: Understood but requires custom solutions
- **Product**: Increasingly standardized, product/rental options
- **Commodity**: Highly standardized, utility services

### 5. Arrange Components

1. Click and drag components to reposition them
2. Positions are automatically saved
3. Reload the page - positions persist!

### 6. View Details

**Component Details:**
1. Click on any component
2. The right panel shows full details
3. Click the ArchiMate Documentation link to learn more
4. Click X to close

**Relation Details:**
1. Click on any edge/arrow
2. The right panel shows relation details
3. See source, target, and type
4. Click the ArchiMate Documentation link
5. Click X to close

## Canvas Controls

### Zoom
- **Mouse wheel**: Zoom in/out
- **Control buttons**: Bottom-left corner
- **Fit View button**: Toolbar (fits all components)

### Pan
- **Click and drag**: On empty canvas space
- **Mini-map**: Bottom-right corner for navigation

### Selection
- **Click component**: Select and show details
- **Click edge**: Select and show details
- **Click empty space**: Clear selection

## Keyboard Shortcuts

- **ESC**: Close open dialog
- **Mouse wheel**: Zoom in/out

## Color Coding

### Components
- **Blue/Purple gradient**: Application components
- **Purple border**: Selected component

### Relations
- **Orange edges**: Triggers relations
- **Blue edges**: Serves relations
- **Thicker line**: Selected relation
- **Animated**: Selected relation

## Example Workflow

Create a simple microservices architecture:

1. **Create Components:**
   - Frontend App
   - API Gateway
   - Auth Service
   - User Service
   - Database

2. **Create Relations:**
   - Frontend App → (Triggers) → API Gateway
   - API Gateway → (Triggers) → Auth Service
   - API Gateway → (Triggers) → User Service
   - Auth Service → (Triggers) → Database
   - User Service → (Triggers) → Database

3. **Arrange:**
   - Place Frontend App at top
   - API Gateway below it
   - Services in the middle row
   - Database at the bottom

4. **Review:**
   - Click each component to see details
   - Click each relation to see type
   - Check ArchiMate documentation links

## Troubleshooting

### Backend not responding
```bash
# Check if backend is running
curl http://localhost:8080/api/v1/components

# If not, start it
cd /home/devuser/repos/easi/backend
go run main.go
```

### Port 5173 already in use
```bash
# Kill the process using the port
lsof -ti:5173 | xargs kill -9

# Or use a different port
npm run dev -- --port 3000
```

### Build errors
```bash
# Clear cache and rebuild
rm -rf node_modules/.vite dist
npm run dev
```

### Data not loading
1. Open browser DevTools (F12)
2. Check Console for errors
3. Check Network tab for failed requests
4. Verify backend URL in src/api/client.ts

## Development Tips

### Hot Module Replacement
- Changes to React components automatically refresh
- Changes to CSS automatically update
- Changes to TypeScript types require refresh

### Console Logs
The app logs helpful information:
- API requests/responses
- State updates
- Component lifecycle

### Browser DevTools
- React DevTools: Inspect component tree
- Redux DevTools: Not needed (using Zustand)
- Network tab: Monitor API calls

## Building for Production

```bash
# Create optimized build
npm run build

# Preview production build
npm run preview

# Output is in dist/ directory
ls -la dist/
```

## Next Steps

1. Explore the codebase in src/
2. Read the full README.md for details
3. Check IMPLEMENTATION.md for architecture
4. Add more components and relations
5. Experiment with the canvas
6. Check ArchiMate documentation links

## Getting Help

- **Code issues**: Check TypeScript errors in terminal
- **Runtime issues**: Check browser console (F12)
- **API issues**: Check Network tab in DevTools
- **Backend logs**: Check backend terminal output

## Happy Modeling!

You now have a fully functional component modeler. Create complex architectures, explore relations, and leverage ArchiMate principles for your designs.
