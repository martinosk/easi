# Code Splitting and Lazy Loading

## Description
Implement code splitting to reduce initial bundle size and improve application load time. Heavy features like Canvas (React Flow, Dagre) and Business Domains should load on demand.

## Rationale
- Current bundle loads all features upfront regardless of usage
- React Flow and Dagre are substantial libraries loaded even when user only needs Business Domains view
- Initial page load is slower than necessary
- Users on slower connections experience unnecessary delays

## Requirements

### Lazy Loading Infrastructure
- [ ] Add React Suspense wrapper with loading fallback component
- [ ] Create shared loading indicator component for consistency

### Feature Lazy Loading
- [ ] Lazy load Canvas feature (React Flow, Dagre dependencies)
- [ ] Lazy load Business Domains feature
- [ ] Lazy load Invitations feature
- [ ] Lazy load Import wizard feature

### Loading States
- [ ] Display appropriate loading indicator during chunk loading
- [ ] Handle chunk loading failures gracefully
- [ ] Ensure smooth transition when chunk loads

## Dependencies
- Spec 079 (Error Boundaries) should be completed first to handle chunk loading failures

## Incremental Delivery
1. First: Suspense wrapper and loading components
2. Second: Canvas feature lazy loading (largest bundle impact)
3. Third: Business Domains lazy loading
4. Fourth: Remaining features

## Checklist
- [ ] Specification ready
- [ ] Suspense infrastructure in place
- [ ] Canvas lazy loaded
- [ ] Business Domains lazy loaded
- [ ] Other features lazy loaded
- [ ] Bundle size reduction verified
- [ ] User sign-off
