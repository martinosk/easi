# Authentication Feature

This feature implements the OIDC login flow for EASI.

## Files

- `types.ts` - TypeScript type definitions for authentication
- `api/authApi.ts` - API client for authentication endpoints
- `pages/LoginPage.tsx` - Login page component

## Usage

### Basic Login Flow

```typescript
import { LoginPage } from './features/auth';

// In your app router or main component:
function App() {
  const isAuthenticated = false; // Check authentication state

  if (!isAuthenticated) {
    return <LoginPage />;
  }

  return <YourMainApp />;
}
```

### How it Works

1. User enters their email (e.g., `john@acme.com`)
2. App calls `POST /auth/sessions` with the email
3. Backend extracts domain (`acme.com`) and looks up tenant
4. Backend generates OIDC authorization URL with PKCE
5. App redirects user to their company's IdP
6. User authenticates with their company credentials
7. IdP redirects to `/auth/callback` (handled by backend)
8. Backend validates, creates session, and redirects to `/`

### API Client

```typescript
import { authApi } from './features/auth';

// Initiate login
const response = await authApi.initiateLogin('john@acme.com');
// Returns: { _links: { self: string, authorize: string } }

// Redirect to IdP (using HATEOAS link)
window.location.href = response._links.authorize;
```

### Error Handling

The login page handles these error scenarios:

- **400 Bad Request**: Invalid email format or domain not registered
- **503 Service Unavailable**: IdP temporarily unavailable

Errors are displayed inline in the form.

## Styling

The login page uses the existing design system defined in `index.css`:

- Gradient background matching app header
- Centered card layout
- Responsive design (mobile-friendly)
- Loading states with spinner
- Error message display

## Integration with Backend

The frontend calls these backend endpoints:

### POST /auth/sessions
```json
// Request
{
  "email": "john@acme.com"
}

// Response (200 OK)
{
  "_links": {
    "self": "/auth/sessions",
    "authorize": "https://login.microsoftonline.com/..."
  }
}

// Error (400 Bad Request)
{
  "error": "Bad Request",
  "message": "Unable to process login request"
}
```

### GET /auth/callback
This endpoint is called by the IdP after authentication. It's not directly called by the frontend.
The backend handles token exchange and redirects to `/` on success.

## Security

- Email is validated client-side for UX only
- All authentication logic happens server-side
- PKCE protects against code interception
- Session cookie is httpOnly and Secure
- Tokens never reach browser JavaScript
