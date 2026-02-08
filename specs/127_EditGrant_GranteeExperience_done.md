# 127 - Edit Grant: Grantee Experience

**Status:** done
**Depends on:** [126_EditGrants_AccessDelegation](126_EditGrants_AccessDelegation_done.md), [068_InvitationSystem](068_InvitationSystem_done.md)

## User Need

When an architect grants a stakeholder edit access to an artifact, the stakeholder currently has no way to know about it, find it, or navigate to it. Meanwhile, if the architect invites a non-user, the grant is silently created for an email that cannot log in. The grantee experience is broken end-to-end.

## Success Criteria

- A grantee can discover their active edit grants without assistance from the grantor
- Each grant provides a direct link to the artifact so the grantee can navigate there in one click
- When a grantor invites a non-user email, a platform invitation is auto-created and the grantor is informed
- The grantor sees clear feedback distinguishing "grant created" from "grant created + invitation sent"

## Vertical Slices

### Slice 1: Artifact Name and Deep Link on Grant Responses

The grant API responses currently show only `artifactType` and `artifactId`. The grantee needs the artifact's display name and a navigable link.

- [x] `GET /api/v1/edit-grants` (grantee's grants) includes `artifactName` for each grant
- [x] Each grant response includes an `artifact` HATEOAS link pointing to the artifact's deep link URL
- [x] Deep link generators exist for capability, component, and view artifact types
- [x] `MyEditAccessPage` renders the artifact name (not raw ID) and a clickable `<Link>` that navigates to the artifact
- [x] The `?capability=<id>` query param is registered in the deep link system and handled by the Business Domains page (selects the capability and opens the details sidebar)
- [x] When an artifact has been deleted (name lookup fails), the grant shows "Deleted artifact" gracefully

### Slice 2: Surface MyEditGrants in the Application

The `MyEditGrants` component exists but is not rendered anywhere. Stakeholders need a place to see their grants. The Settings page is gated behind `metamodel:write`, so stakeholders cannot reach it. Instead, "My Edit Access" is a standalone page accessible from the UserMenu dropdown.

- [x] "My Edit Access" is a standalone page at `/my-edit-access` (not under Settings), following the UsersPage/InvitationsPage table layout pattern
- [x] The UserMenu dropdown shows a "My Edit Access" link with a blue count badge when the user has active grants
- [x] The link only appears when the user has at least one active edit grant (avoid empty UI noise)
- [x] The page renders a table with columns: Artifact (name linked via deep link), Granted by, Reason, Expires
- [x] Loading state shows a spinner; empty state shows an icon with "You have no active edit access grants"
- [x] Expired/revoked grants are not shown (filtered to active only)
- [x] The Settings page is restored to its original `metamodel:write`-only guard (no edit grants logic)

### Slice 3: Non-User Auto-Invitation

When a grantor enters an email for someone who is not an EASI user, the system should automatically create a platform invitation so the person can join and immediately have edit access.

- [x] `POST /api/v1/edit-grants` checks whether the grantee email belongs to an existing user
- [x] If the email is not a registered user AND no pending invitation exists, the backend publishes an `EditGrantForNonUserCreated` event
- [x] A projector in the `auth` context subscribes to `EditGrantForNonUserCreated` and creates a platform invitation (role: stakeholder) using the existing invitation aggregate
- [x] If a pending invitation already exists for that email, no duplicate invitation is created
- [x] The edit grant is created regardless (it is keyed by email, not user ID, so it will resolve when the user joins)
- [x] The create response includes `invitationCreated: true/false` to indicate whether an invitation was also sent
- [x] Domain validation for the invitation email uses the existing tenant domain checker

### Slice 4: Grantor Feedback Enhancement

The grantor needs to understand what happened after creating a grant, especially when a platform invitation was triggered.

- [x] When `invitationCreated` is true in the response, the frontend toast shows: "Edit access granted. An invitation to join EASI was also sent to {email}."
- [x] When `invitationCreated` is false (existing user), the toast shows the current message: "Edit access granted to {email}"
- [x] The `InviteToEditDialog` closes on success as it does today (no change to dialog flow)

## Out of Scope (for now)

- **Email notifications**: No email is sent to the grantee. Discovery is via the in-app "My Edit Access" page (UserMenu). Email notification is a separate concern.
- **Grant acceptance step**: Grants remain immediately active (no pending/accept flow for the grantee).
- **Grantee revoking their own grant**: Grantees cannot opt out of a grant.
- **Admin view of all grants across users**: The existing `EditGrantsList` admin component is sufficient for now.
- **Push notifications or real-time updates**: The grantee discovers grants on their next page load.

## Cross-Context Event Flow

```
accessdelegation                          auth
      |                                     |
  CreateEditGrant                           |
      |                                     |
  [user lookup: email not found]            |
      |                                     |
  EditGrantForNonUserCreated  ------>  InvitationAutoCreateProjector
      |                                     |
  EditGrantActivated                   InvitationCreated
```

The `accessdelegation` context publishes `EditGrantForNonUserCreated` as a published language event. The `auth` context subscribes and handles invitation creation independently. No direct dependency between contexts.

## API Changes

### POST /api/v1/edit-grants (updated response)

The response gains two new fields:

```json
{
  "id": "grant-uuid",
  "granteeEmail": "stakeholder@company.com",
  "artifactType": "capability",
  "artifactId": "artifact-uuid",
  "artifactName": "Customer Onboarding",
  "invitationCreated": false,
  "_links": {
    "self": { "href": "/api/v1/edit-grants/grant-uuid", "method": "GET" },
    "revoke": { "href": "/api/v1/edit-grants/grant-uuid", "method": "DELETE" },
    "artifact": { "href": "/business-domains?capability=artifact-uuid", "method": "GET" }
  }
}
```

### GET /api/v1/edit-grants (grantee view, updated response)

Each grant in the array gains `artifactName` and the `artifact` link:

```json
{
  "id": "grant-uuid",
  "artifactName": "Customer Onboarding",
  "_links": {
    "self": { "href": "/api/v1/edit-grants/grant-uuid", "method": "GET" },
    "artifact": { "href": "/business-domains?capability=artifact-uuid", "method": "GET" }
  }
}
```

## Checklist

- [x] Backend: artifact name resolution for grant responses (cross-context read)
- [x] Backend: deep link generation per artifact type
- [x] Backend: user existence check in create handler
- [x] Backend: `EditGrantForNonUserCreated` published language event
- [x] Backend: `InvitationAutoCreateProjector` in auth context
- [x] Backend: `invitationCreated` field on create response
- [x] Frontend: deep link generators for capability/component artifact types
- [x] Frontend: `?capability=` deep link param registered and handled in Business Domains page
- [x] Frontend: `MyEditAccessPage` standalone page at `/my-edit-access` with table layout
- [x] Frontend: UserMenu dropdown link with count badge (replaces Settings tab)
- [x] Frontend: grantor toast differentiation based on `invitationCreated`
- [x] Unit tests: auto-invitation projector
- [x] Unit tests: artifact name resolution
- [x] Frontend tests: MyEditGrants rendering with links
- [x] Frontend tests: toast message variants
