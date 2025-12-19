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
- [ ] Define route structure in a central configuration
- [ ] Configure routes for existing views (canvas, business-domains, invitations)
- [ ] Set up route for login page

### Phase 2: Protected Routes
- [ ] Create authenticated route wrapper that checks session state
- [ ] Redirect unauthenticated users to login
- [ ] Preserve intended destination for post-login redirect

### Phase 3: App Component Migration
- [ ] Remove hash-based navigation logic from App.tsx
- [ ] Replace manual view switching with Route components
- [ ] Remove window.location.hash listeners

### Phase 4: Navigation Updates
- [ ] Update all navigation actions to use React Router navigation
- [ ] Replace any remaining hash-based links
- [ ] Update any external links that reference hash routes

### Phase 5: Backend Fallback
- [ ] Ensure backend serves index.html for all frontend routes
- [ ] Verify deep linking works correctly

## Dependencies
- Spec 080 (Code Splitting) integrates well with this spec via Suspense

## Incremental Delivery
1. First: Route configuration alongside existing hash navigation (both work)
2. Second: Protected route wrapper
3. Third: Migrate App.tsx to use routes
4. Fourth: Remove hash navigation code
5. Fifth: Backend configuration for SPA routing

## Checklist
- [ ] Specification ready
- [ ] Route configuration created
- [ ] Protected routes implemented
- [ ] App.tsx migrated
- [ ] Hash navigation removed
- [ ] Backend fallback configured
- [ ] All navigation paths tested
- [ ] User sign-off
