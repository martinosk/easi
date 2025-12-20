# Error Boundaries

## Description
Add React error boundaries to prevent component crashes from taking down the entire application. Critical for features using complex third-party libraries like React Flow and dnd-kit.

## Rationale
- React Flow canvas errors currently crash the entire application
- Users lose all unsaved work when an error occurs
- No graceful degradation or recovery options exist
- Error details are not captured for debugging

## Requirements

### Shared Error Boundary Component
- [ ] Create reusable error boundary component in shared components
- [ ] Display user-friendly error message with option to retry
- [ ] Capture error details for potential logging
- [ ] Provide reset functionality to attempt recovery

### Feature-Level Boundaries
- [ ] Wrap Canvas feature with error boundary
- [ ] Wrap Business Domains feature with error boundary
- [ ] Wrap Navigation tree with error boundary

### Root-Level Fallback
- [ ] Add application-level error boundary as final fallback
- [ ] Display minimal UI allowing user to refresh or navigate away
- [ ] Preserve ability to access other working parts of the application

## Incremental Delivery
1. First: Shared error boundary component
2. Second: Canvas feature boundary (highest risk feature)
3. Third: Business Domains boundary
4. Fourth: Root-level fallback

## Checklist
- [x] Specification ready
- [x] Shared error boundary component created (ErrorBoundary.tsx with DefaultErrorFallback and FeatureErrorFallback)
- [x] Feature boundaries implemented (Canvas, Business Domains, Invitations wrapped)
- [x] Root fallback implemented (RootErrorFallback in main.tsx)
- [x] Manual testing of error scenarios
- [x] User sign-off
