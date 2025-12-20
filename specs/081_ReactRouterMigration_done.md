# React Router Migration

## Description
Replace the current hash-based navigation with proper React Router declarative routing. This enables standard URL patterns, route guards, and integration with code splitting.

## Current State
- Navigation uses `window.location.hash` with manual event listeners
- `App.tsx` manually parses hash values to determine current view
- No support for URL parameters, nested routes, or route guards
- Hash-based URLs are non-standard for modern SPAs

## Target State
- Standard path-based URLs without hash fragments
- Declarative route configuration
- Protected routes for authenticated sections
- Integration with lazy loading (Suspense)

## Requirements

### Phase 1: Route Configuration
- [x] Define route structure in a central configuration (routes/routes.tsx with ROUTES constant)
- [x] Configure routes for existing views (canvas, business-domains, invitations)
- [x] Set up route for login page

### Phase 2: Protected Routes
- [x] Create authenticated route wrapper that checks session state (ProtectedRoute component)
- [x] Redirect unauthenticated users to login
- [x] Preserve intended destination for post-login redirect (via location state)

### Phase 3: App Component Migration
- [x] Remove hash-based navigation logic from App.tsx
- [x] Replace manual view switching with Route components (main.tsx Routes)
- [x] Remove window.location.hash listeners

### Phase 4: Navigation Updates
- [x] Update all navigation actions to use React Router navigation
- [x] Replace any remaining hash-based links (DomainDetailPage fixed)
- [x] Update any external links that reference hash routes

### Phase 5: Backend Fallback
- [x] Ensure backend serves index.html for all frontend routes (backend already configured)
- [x] Verify deep linking works correctly

## Dependencies
- Spec 080 (Code Splitting) integrates well with this spec via Suspense

## Incremental Delivery
1. First: Route configuration alongside existing hash navigation (both work)
2. Second: Protected route wrapper
3. Third: Migrate App.tsx to use routes
4. Fourth: Remove hash navigation code
5. Fifth: Backend configuration for SPA routing

## Checklist
- [x] Specification ready
- [x] Route configuration created (routes/routes.tsx)
- [x] Protected routes implemented (ProtectedRoute component)
- [x] App.tsx migrated (receives view prop from routes)
- [x] Hash navigation removed (all window.location.hash references removed)
- [x] Backend fallback configured (backend already serves SPA routes)
- [x] All navigation paths tested (569 tests passing)
- [x] User sign-off
