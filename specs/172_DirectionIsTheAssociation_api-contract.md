# Spec 172 — Direction Is the Association: Backend API Contract

> **Status:** Implemented (frontend and backend)
> **Frontend:** Complete (MSW stub is the authoritative contract source)
> **Stub files (source of truth):** `frontend/src/test/mocks/spec172/`
> **Spec prose:** `specs/172_DirectionIsTheAssociation_ongoing.md`

## Revision Log

- **2026-08-29 — Amended by spec 207.** Composition read view, source-candidates search and maturity analysis are served by `architecturedirection` (same URLs and shapes); `includedCapabilityCount`/`domainCount` are removed from the EC DTO and served by `GET /enterprise-capability-compositions`. O5 still holds: composition is computed per request from Architecture Direction's own caches.

- **2026-06-11 — Backend implemented; open items §7 resolved.**
  - **O1:** `pagination.cursor` stays an empty-string placeholder; `hasMore` is computed by limit lookahead. The cursor format remains undecided until the client needs paging.
  - **O2:** the acting architect is recorded as an `actor` field on the `DirectionSourceCapabilitiesChanged` payload only; `BaseEvent` is unchanged. Historic events deserialize with an empty actor.
  - **O3:** propose-time cardinality follows the established direction-type rules from spec 167, which §2.10 said to adopt if the domain rules differed: `consolidate` ≥ 2 sources, `decompose` exactly 1, `stay` exactly 1. The 400 messages are "A 'consolidate' direction requires at least 2 sources to be proposed." and "A '<type>' direction requires exactly 1 source to be proposed." Additionally, spec 167's rule 10 still applies: a narrative is required to propose (400 "A narrative is required before advancing a direction to proposed."). The FE renders these messages verbatim; its only hard-coded hint (consolidate ≥ 2) matches.
  - **O4:** `direction.updatedAt` is absent until the first mutation (the read model sets `updated_at` only on updates).
  - **O5:** composition is computed on the fly per request from the active direction sources + capability metadata; no projected composition table. Maturity analysis derives its per-EC membership from the same computation.
  - **O6:** stale sources keep the existing event-driven flag on `direction_source_capabilities` (set when `CapabilityDeleted` arrives).
  - **O7:** transitions enforce the aggregate state machine; an invalid-from-status transition returns **404**, matching this contract.
  - `PUT .../direction` also no longer accepts `placements` (sub-feature deferred; capture does not populate them either).
  - `x-direction` on the EC DTO is now unconditional (previously gated on direction read permission), matching §3.1.

---

## 1. Overview & Conventions

### Base path

All routes below are relative to `@BasePath /api/v1`. The `@Router` annotation in every handler godoc must omit the `/api/v1` prefix.

### Bounded-context ownership

| Concern | Bounded context | Handler package |
|---|---|---|
| EC detail GET (including `includedCapabilityCount`, `domainCount`) | `enterprisearchitecture` | `internal/enterprisearchitecture/...` |
| EC composition read view (`/composition`) | `enterprisearchitecture` | same |
| Source-candidates search (`/capabilities/source-candidates`) | `enterprisearchitecture` | same (eligibility is owned here) |
| Direction writes (capture, update, transitions, source mutations) | `architecturedirection` | `internal/architecturedirection/...` |
| Direction read (`/direction` GET) | `architecturedirection` | same |
| Composition-preview POST | `architecturedirection` | same (calls into eligibility service cross-context) |

### `_links` shape

Every link object is: `{ "href": string, "method": string }`.

Go type alias already present in the codebase: `types.Link{Href string, Method string}` and `types.Links map[string]types.Link`.

Every DTO and every response envelope carries `_links types.Links \`json:"_links,omitempty"\``.

### Error envelope

All error responses use the following shape. Additional fields appear only in 409 bodies (see per-endpoint sections):

```json
{
  "error": "<short class string>",
  "message": "<human-readable explanation>",
  "details": { },
  "_links": { }
}
```

The `details` and `_links` fields on errors are `omitempty`.

### Auth & permissions

All endpoints require cookie-based authentication (`@Security CookieAuth`). Annotate `@Failure 401` and `@Failure 403` on every handler. Permission-gating within `_links` (i.e., whether `edit`/`delete`/`x-propose` appear) must be driven by the actor's role, not by status alone; the stub does not model roles but the backend must. The stub's link presence rules show the minimum set — the backend gates them additionally on actor permissions.

---

## 2. Per-Endpoint Reference

### 2.1 GET `/enterprise-capabilities/{id}`

**Changed fields.** This is an existing endpoint. Spec 172 changes two fields in the response:

- `linkCount` is **removed** and replaced by `includedCapabilityCount` (integer, always present, ≥0).
- `domainCount` is **added** (integer, always present, ≥0).

Both values are computed from the composition algorithm (R2) at read time; see section 4 for how they are calculated.

**Full response schema (changed fields only shown with context):**

```json
{
  "id": "ec-abc123",
  "name": "Customer Identity",
  "description": "",
  "category": "",
  "active": true,
  "targetMaturity": 3,
  "includedCapabilityCount": 4,
  "domainCount": 2,
  "createdAt": "2024-01-15T10:00:00Z",
  "updatedAt": null,
  "_links": {
    "self":          { "href": "/api/v1/enterprise-capabilities/ec-abc123", "method": "GET" },
    "edit":          { "href": "/api/v1/enterprise-capabilities/ec-abc123", "method": "PUT" },
    "delete":        { "href": "/api/v1/enterprise-capabilities/ec-abc123", "method": "DELETE" },
    "x-direction":   { "href": "/api/v1/enterprise-capabilities/ec-abc123/direction", "method": "GET" },
    "x-composition": { "href": "/api/v1/enterprise-capabilities/ec-abc123/composition", "method": "GET" }
  }
}
```

`edit` and `delete` are gated on actor permissions. `x-direction` and `x-composition` are always present (unconditional navigation links).

**Status codes:** 200, 404, 401, 403, 500 (unchanged from prior contract).

---

### 2.2 GET `/enterprise-capabilities/{id}/composition`

**Description:** Returns the full composition of an EC: every domain capability included via the active direction's sources and their subtrees, grouped by business domain, with per-item role classification and carve-out attribution.

**Path params:** `id` (string, required) — enterprise capability ID.

**Request body:** none.

**Success — 200:**

```json
{
  "data": [
    {
      "businessDomainId": "dom-001",
      "businessDomainName": "Customer",
      "items": [
        {
          "capabilityId": "cap-001",
          "name": "Customer Account Creation",
          "level": "L2",
          "businessDomainId": "dom-001",
          "businessDomainName": "Customer",
          "role": "source",
          "carvedOutBy": null,
          "_links": {
            "self":      { "href": "/api/v1/capabilities/cap-001", "method": "GET" },
            "x-exclude": { "href": "/api/v1/enterprise-capabilities/ec-abc123/direction/sources/cap-001", "method": "DELETE" }
          }
        },
        {
          "capabilityId": "cap-002",
          "name": "Customer Fraud Prevention",
          "level": "L3",
          "businessDomainId": "dom-001",
          "businessDomainName": "Customer",
          "role": "carved-out",
          "carvedOutBy": {
            "enterpriseCapabilityId": "ec-take-payment",
            "enterpriseCapabilityName": "Take Payment"
          },
          "_links": {
            "self":        { "href": "/api/v1/capabilities/cap-002", "method": "GET" },
            "x-owning-ec": { "href": "/api/v1/enterprise-capabilities/ec-take-payment", "method": "GET" }
          }
        },
        {
          "capabilityId": "cap-003",
          "name": "Customer Consent",
          "level": "L3",
          "businessDomainId": "dom-001",
          "businessDomainName": "Customer",
          "role": "implicit",
          "carvedOutBy": null,
          "_links": {
            "self": { "href": "/api/v1/capabilities/cap-003", "method": "GET" }
          }
        }
      ]
    },
    {
      "businessDomainId": null,
      "businessDomainName": null,
      "items": [ ]
    }
  ],
  "meta": {
    "sourceCount": 1,
    "includedCount": 2,
    "carvedOutCount": 1,
    "domainCount": 1
  },
  "_links": {
    "self":       { "href": "/api/v1/enterprise-capabilities/ec-abc123/composition", "method": "GET" },
    "up":         { "href": "/api/v1/enterprise-capabilities/ec-abc123", "method": "GET" },
    "x-direction": { "href": "/api/v1/enterprise-capabilities/ec-abc123/direction", "method": "GET" }
  }
}
```

When **no active direction exists**, the response additionally contains `"x-capture-direction"` in `_links` and the `data` array is empty; `meta` counts are all `0`.

When an **active direction exists**, `"x-capture-direction"` is absent from `_links`.

**Conditional `_links` on `data[].items[]`:**

| Condition | Link present |
|---|---|
| Item `role == "source"` AND direction is NOT `agreed` | `x-exclude` |
| Item `role == "source"` AND direction IS `agreed` | `x-exclude` absent |
| Item `role == "carved-out"` AND `carvedOutBy != null` | `x-owning-ec` |
| All items | `self` |

**Status codes:** 200, 404 (EC not found), 401, 403, 500.

---

### 2.3 GET `/capabilities/source-candidates`

**Route note for chi.** This is a **static path** (`/capabilities/source-candidates`), not `/capabilities/{id}`. If the backend registers both a `GET /capabilities/{id}` route and this one on the same router, `source-candidates` must be registered **before** `{id}` so chi matches the literal segment first.

**Query params:**

| Name | Type | Required | Description |
|---|---|---|---|
| `q` | string | yes | Search term (case-insensitive substring match on capability name) |
| `ecId` | string | yes | The EC for which sources are being searched (used for R1 conflict detection) |
| `domainId` | string | no | Filter to capabilities whose `businessDomainId` matches |
| `limit` | integer | no | Max results to return; default 20 |
| `cursor` | string | no | Opaque pagination cursor (see open items §7) |

**Request body:** none.

**400 when `q` or `ecId` is missing:**

```json
{ "error": "BadRequest", "message": "q and ecId are required" }
```

**Success — 200:**

```json
{
  "data": [
    {
      "capabilityId": "cap-001",
      "name": "Customer Account Creation",
      "level": "L2",
      "parentId": "cap-parent",
      "businessDomainId": "dom-001",
      "businessDomainName": "Customer",
      "eligible": true,
      "ineligibilityReason": null,
      "conflictingEnterpriseCapability": null,
      "_links": {
        "self": { "href": "/api/v1/capabilities/cap-001", "method": "GET" }
      }
    },
    {
      "capabilityId": "cap-ineligible",
      "name": "Customer Fraud Prevention",
      "level": "L3",
      "parentId": "cap-parent",
      "businessDomainId": "dom-001",
      "businessDomainName": "Customer",
      "eligible": false,
      "ineligibilityReason": "Already an explicit source of an active direction on 'Take Payment'",
      "conflictingEnterpriseCapability": {
        "id": "ec-take-payment",
        "name": "Take Payment"
      },
      "_links": {
        "self": { "href": "/api/v1/capabilities/cap-ineligible", "method": "GET" },
        "x-conflicting-ec": { "href": "/api/v1/enterprise-capabilities/ec-take-payment", "method": "GET" }
      }
    }
  ],
  "pagination": {
    "hasMore": false,
    "limit": 20,
    "cursor": ""
  },
  "_links": {
    "self": { "href": "/api/v1/capabilities/source-candidates?q=customer&ecId=ec-abc123", "method": "GET" }
  }
}
```

**`_links` on each `SourceCandidate`:**

| Condition | Link present |
|---|---|
| Always | `self` pointing to `/api/v1/capabilities/{capabilityId}` |
| `eligible == false` AND `conflictingEnterpriseCapability != null` | `x-conflicting-ec` pointing to the conflicting EC's detail URL |

**Field nullability:**

- `parentId`: `string | null` (null when root capability)
- `businessDomainId`: `string | null`
- `businessDomainName`: `string | null`
- `ineligibilityReason`: `string | null`
- `conflictingEnterpriseCapability`: `object | null`

**Status codes:** 200, 400 (missing q/ecId), 401, 403, 500.

---

### 2.4 POST `/enterprise-capabilities/{id}/direction/composition-preview`

**Description:** Stateless preview — resolves what the composition would be for a given proposed source set without persisting anything. Used in the capture modal for R2 preview and R1 eligibility pre-flight.

**Path params:** `id` (string, required) — enterprise capability ID.

**Request body:**

```json
{ "sourceCapabilityIds": ["cap-001", "cap-002"] }
```

`sourceCapabilityIds` is required; an empty array is valid.

**Success — 200:**

```json
{
  "includedCapabilities": [
    {
      "capabilityId": "cap-001",
      "name": "Customer Account Creation",
      "level": "L2",
      "businessDomainId": "dom-001",
      "businessDomainName": "Customer",
      "role": "source",
      "carvedOutBy": null
    },
    {
      "capabilityId": "cap-003",
      "name": "Customer Consent",
      "level": "L3",
      "businessDomainId": "dom-001",
      "businessDomainName": "Customer",
      "role": "implicit",
      "carvedOutBy": null
    }
  ],
  "sourceEligibility": [
    {
      "capabilityId": "cap-001",
      "eligible": true,
      "ineligibilityReason": null,
      "conflictingEnterpriseCapability": null
    }
  ],
  "meta": {
    "sourceCount": 1,
    "includedCount": 2,
    "carvedOutCount": 0
  },
  "_links": {
    "self": { "href": "/api/v1/enterprise-capabilities/ec-abc123/direction/composition-preview", "method": "POST" }
  }
}
```

The `meta` here has only three fields (`sourceCount`, `includedCount`, `carvedOutCount`) — there is no `domainCount` in the preview response.

`includedCapabilities` contains **all** resolved items including carved-out ones. `includedCount = count where role != "carved-out"`. `carvedOutCount = count where role == "carved-out"`. `sourceCount = sourceCapabilityIds.length` from the request.

**Status codes:** 200, 404 (EC not found), 401, 403, 500.

---

### 2.5 GET `/enterprise-capabilities/{id}/direction`

**Description:** Returns the active direction (draft/proposed/agreed) for the EC, or null if none exists.

**Path params:** `id` (string, required).

**Request body:** none.

**Success — 200:**

```json
{
  "direction": {
    "id": "dir-abc123",
    "enterpriseCapabilityId": "ec-abc123",
    "type": "consolidate",
    "status": "draft",
    "horizon": "now",
    "narrative": "Consolidate all customer identity capabilities",
    "sourceCapabilities": [
      {
        "id": "cap-001",
        "stale": false,
        "name": "Customer Account Creation",
        "businessDomainId": "dom-001",
        "businessDomainName": "Customer"
      }
    ],
    "placements": [],
    "createdAt": "2024-01-15T10:00:00Z",
    "updatedAt": null,
    "_links": {
      "self":        { "href": "/api/v1/enterprise-capabilities/ec-abc123/direction", "method": "GET" },
      "up":          { "href": "/api/v1/enterprise-capabilities/ec-abc123", "method": "GET" },
      "edit":        { "href": "/api/v1/enterprise-capabilities/ec-abc123/direction", "method": "PUT" },
      "x-add-source": { "href": "/api/v1/enterprise-capabilities/ec-abc123/direction/sources", "method": "POST" },
      "x-propose":   { "href": "/api/v1/enterprise-capabilities/ec-abc123/direction/propose", "method": "POST" },
      "x-reject":    { "href": "/api/v1/enterprise-capabilities/ec-abc123/direction/reject", "method": "POST" }
    }
  },
  "_links": {
    "self":         { "href": "/api/v1/enterprise-capabilities/ec-abc123/direction", "method": "GET" },
    "up":           { "href": "/api/v1/enterprise-capabilities/ec-abc123", "method": "GET" },
    "x-composition": { "href": "/api/v1/enterprise-capabilities/ec-abc123/composition", "method": "GET" }
  }
}
```

When **no active direction:** `"direction": null` and `_links` on the envelope gains `"x-capture-direction": { "href": "...", "method": "POST" }`.

**Field details for `direction.sourceCapabilities[]`:**

- `id`: string (the domain capability ID)
- `stale`: boolean — true when the capability no longer exists in the read model (deleted capability still referenced by the direction)
- `name`: `string | null` — null when stale
- `businessDomainId`: `string | undefined` — may be absent
- `businessDomainName`: `string | null` — null or absent when unassigned

`direction.placements` is always an empty array for this spec; the placements sub-feature is deferred. Backend should persist and return it but capture does not populate it.

**Status codes:** 200, 404 (EC not found), 401, 403, 500.

---

### 2.6 POST `/enterprise-capabilities/{id}/direction`

**Description:** Captures a new direction in `draft` status (R4 — EC must be active; R1 — each source exclusive; R8 — single source allowed in draft).

**Path params:** `id` (string, required).

**Request body:**

```json
{
  "type": "consolidate",
  "sourceCapabilityIds": ["cap-001"],
  "horizon": "now",
  "narrative": "Optional free-text rationale"
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `type` | `"consolidate" \| "decompose" \| "stay"` | yes | |
| `sourceCapabilityIds` | `string[]` | yes | May be a single element in draft (R8); empty array is valid at draft |
| `horizon` | `"now" \| "next" \| "later"` | yes | |
| `narrative` | string | no | Free text |

**Success — 201:** Returns a `Direction` DTO (same shape as §2.5 `direction` object). Set `Location` header to `/api/v1/enterprise-capabilities/{id}/direction`.

**409 — EC inactive (R4):**

```json
{
  "error": "Conflict",
  "message": "Directions can only be captured on active enterprise capabilities."
}
```

**409 — Source already claimed (R1):**

```json
{
  "error": "Conflict",
  "message": "Capability 'Customer Account Creation' is already an explicit source of an active direction on 'Customer Identity'. A domain capability may be the explicit source of at most one active direction.",
  "details": {
    "capabilityId": "cap-001",
    "capabilityName": "Customer Account Creation",
    "conflictingEnterpriseCapabilityId": "ec-customer-identity",
    "conflictingEnterpriseCapabilityName": "Customer Identity"
  },
  "_links": {
    "x-conflicting-ec": { "href": "/api/v1/enterprise-capabilities/ec-customer-identity", "method": "GET" }
  }
}
```

R1 is checked per-source. The backend stops at the first violation and returns 409. The frontend re-fires after correcting each conflict.

**Status codes:** 201, 404 (EC not found), 409, 401, 403, 500.

---

### 2.7 PUT `/enterprise-capabilities/{id}/direction`

**Description:** Updates horizon and/or narrative of the active direction. `sourceCapabilityIds` is NOT in this request body (source mutations use the granular sub-resources in §2.8–2.9).

**Path params:** `id` (string, required).

**Request body:**

```json
{
  "horizon": "next",
  "narrative": "Updated rationale"
}
```

Both fields are optional; send only those being changed.

**409 — Agreed immutable (R5):**

```json
{
  "error": "Conflict",
  "message": "This direction is agreed and its source set is frozen. To recompose, reject the direction and capture a new one.",
  "details": { "directionStatus": "agreed" },
  "_links": {
    "x-reject": { "href": "/api/v1/enterprise-capabilities/ec-abc123/direction/reject", "method": "POST" }
  }
}
```

**Success — 200:** Returns updated `Direction` DTO.

**Status codes:** 200, 404 (EC or active direction not found), 409, 401, 403, 500.

---

### 2.8 POST `/enterprise-capabilities/{id}/direction/sources`

**Description:** Adds a domain capability to the active direction's source set (R1 checked; R5 — agreed direction rejects).

**Path params:** `id` (string, required).

**Request body:**

```json
{ "capabilityId": "cap-002" }
```

If `capabilityId` is already in the source set, the backend is idempotent and returns the current direction unchanged with 200.

**Success — 200:** Returns current `Direction` DTO (not 201 — the direction resource already exists; only the source set is mutated).

**409 — Agreed immutable (R5):** Same body as §2.7.

**409 — Source already claimed (R1):** Same body as §2.6.

**Status codes:** 200, 404 (active direction not found), 409, 401, 403, 500.

---

### 2.9 DELETE `/enterprise-capabilities/{id}/direction/sources/{capabilityId}`

**Description:** Removes a domain capability from the active direction's source set. This is the "exclude" action in the UI.

**Path params:**
- `id` (string, required) — enterprise capability ID
- `capabilityId` (string, required) — domain capability ID to remove

**Request body:** none.

**409 — Agreed immutable (R5):** Same body as §2.7.

**404 when:** no active direction exists, or `capabilityId` is not in the source set.

**Success — 204:** No response body.

**Status codes:** 204, 404, 409, 401, 403, 500.

---

### 2.10 POST `/enterprise-capabilities/{id}/direction/propose`

**Description:** Advances the active direction from `draft` to `proposed`. Backend enforces R8 cardinality at this transition.

**Path params:** `id` (string, required).

**Request body:** none.

**Success — 200:** Returns updated `Direction` DTO with `status: "proposed"`.

**400 — R8 cardinality not met:** When `type == "consolidate"` or `type == "decompose"` and `sourceCapabilityIds.length < 2`:

```json
{
  "error": "BadRequest",
  "message": "A 'consolidate' direction requires at least 2 sources to be proposed."
}
```

Adjust message for `decompose` type. The `stay` type has no minimum source count.

**404 when:** no active direction, or active direction is not in `draft` status.

**Status codes:** 200, 400, 404, 401, 403, 500.

---

### 2.11 POST `/enterprise-capabilities/{id}/direction/agree`

**Description:** Advances the active direction from `proposed` to `agreed`.

**Path params:** `id` (string, required).

**Request body:** none.

**Success — 200:** Returns updated `Direction` DTO with `status: "agreed"`.

**404 when:** no active direction, or active direction is not in `proposed` status.

**Status codes:** 200, 404, 401, 403, 500.

---

### 2.12 POST `/enterprise-capabilities/{id}/direction/reject`

**Description:** Rejects the active direction (any non-rejected status → rejected). Makes the EC available for a new capture.

**Path params:** `id` (string, required).

**Request body:** none.

**Success — 200:** Returns updated `Direction` DTO with `status: "rejected"`.

**404 when:** no active direction (direction already rejected or never existed).

**Status codes:** 200, 404, 401, 403, 500.

---

## 3. HATEOAS Link Matrix

### 3.1 EC DTO (`EnterpriseCapability`) `_links`

| Relation | Always | Condition |
|---|---|---|
| `self` | yes | — |
| `edit` | no | Actor has write permission |
| `delete` | no | Actor has write permission |
| `x-direction` | yes | — (navigation link, unconditional) |
| `x-composition` | yes | — (navigation link, unconditional) |

Note: `x-strategic-importance` and `x-strategic-fit` are pre-existing unconditional links not changed by this spec.

### 3.2 `ECDirectionResponse` envelope `_links`

| Relation | Condition |
|---|---|
| `self` | Always |
| `up` | Always |
| `x-composition` | Always |
| `x-capture-direction` | Only when `direction == null` (no active direction) |

### 3.3 `Direction` DTO `_links` — by status

| Relation | `draft` | `proposed` | `agreed` | `rejected` |
|---|---|---|---|---|
| `self` | yes | yes | yes | yes |
| `up` | yes | yes | yes | yes |
| `edit` | yes (actor-gated) | yes (actor-gated) | no | no |
| `x-add-source` | yes (actor-gated) | yes (actor-gated) | no | no |
| `x-propose` | yes | no | no | no |
| `x-agree` | no | yes | no | no |
| `x-reject` | yes | yes | yes | no |

`edit` and `x-add-source` appear together for `editable` states (draft or proposed), both gated on actor permission. `x-propose` is draft-only. `x-agree` is proposed-only. `x-reject` appears on all non-rejected states.

### 3.4 `CompositionResponse` envelope `_links`

| Relation | Condition |
|---|---|
| `self` | Always |
| `up` | Always |
| `x-direction` | Always |
| `x-capture-direction` | Only when no active direction exists |

### 3.5 `IncludedCapabilityItem` `_links` (per item in composition)

| Relation | Condition |
|---|---|
| `self` | Always (points to `/api/v1/capabilities/{capabilityId}`) |
| `x-exclude` | `role == "source"` AND active direction is NOT `agreed` |
| `x-owning-ec` | `role == "carved-out"` AND `carvedOutBy != null` |

### 3.6 `SourceCandidate` `_links`

| Relation | Condition |
|---|---|
| `self` | Always (points to `/api/v1/capabilities/{capabilityId}`) |
| `x-conflicting-ec` | `eligible == false` AND `conflictingEnterpriseCapability != null` |

### 3.7 409 error body `_links`

| Error type | Relation |
|---|---|
| R1 conflict (capture or add-source) | `x-conflicting-ec` pointing to the EC that already claims the source |
| R5 immutable (add-source, exclude, update) | `x-reject` pointing to the reject transition endpoint |

---

## 4. Domain Algorithm Specification

### R1 — Same-node exclusivity

**Input:** A domain capability ID, a target EC ID, and the set of all currently active directions across all ECs (status `draft`, `proposed`, or `agreed`; EC not deleted).

**Rule:** Search the active directions for any direction whose EC ID differs from the target EC and whose `sourceCapabilityIds` contains the exact capability ID.

**Outcome:** If found, the capability is ineligible; the conflicting EC ID and name are returned. If not found, eligible.

**Implementation note:** Only explicit sources (listed in `sourceCapabilityIds`) trigger R1. A capability that is only implicitly included (descendant of a source) is not itself an explicit source and does not block other ECs from sourcing it directly. Only sourcing the exact same node is rejected.

**Active direction definition:** A direction is active when `status IN ('draft', 'proposed', 'agreed')`. A draft direction **reserves** its sources immediately upon capture. This means R1 is enforced at draft capture time, not only at propose time.

### R2 — Subtree composition with most-specific-wins carve-outs

This is the core algorithm, implemented in `composition.ts:resolveComposition`. Go port must replicate it exactly.

**Inputs:**
- `targetEcId`: the EC being composed
- `allActiveDirections`: slice of `{ecId, sourceCapabilityIds[]}` for all active directions across all ECs
- `allCapabilities`: flat list of all domain capabilities with `{id, parentId}` relationships

**Step 1 — Build indexes:**
1. Build a `childrenByParent` map: for every capability with a `parentId`, add it to the parent's children list.
2. Build a `capById` map for O(1) lookup.

**Step 2 — Build ownership map from other ECs:**
- For every active direction whose `ecId != targetEcId`, for every `sourceId` in that direction's `sourceCapabilityIds`, record `sourceId → { enterpriseCapabilityId, enterpriseCapabilityName }` as "owned by another EC". Call this `ownershipByOtherEc`.
- If the same capability is sourced by multiple other ECs (which R1 prevents but belt-and-suspenders), last-write wins; this case should not occur in a valid system.

**Step 3 — Get target direction:**
- Find the active direction where `ecId == targetEcId`. If none, return empty result.
- The target direction's `sourceCapabilityIds` is the `targetSources` set.

**Step 4 — Depth-first traversal from each source:**
- For each `sourceId` in `targetSources`, call `visitNode(sourceId, ctx)` where `ctx` holds all the indexes plus a `resolved` map.
- `visitNode(capId, ctx)`:
  1. If `resolved` already contains `capId`, return (already processed; prevents duplicate inclusion).
  2. Look up `cap = capById[capId]`. If not found, skip (stale source).
  3. Check `ownershipByOtherEc[capId]`. If present AND `capId` is NOT in `targetSources`:
     - Record `capId` as `role: "carved-out"` with `carvedOutBy = ownershipByOtherEc[capId]`.
     - **Do not recurse into children.** The entire subtree is owned by the carving EC.
     - Return.
  4. Otherwise: record `capId` as `role: "source"` if `capId ∈ targetSources`, else `role: "implicit"`. `carvedOutBy = null`.
  5. Recurse: for each child of `capId`, call `visitNode(child.id, ctx)`.

**Most-specific-wins semantics:** A node that is both an explicit source of the target EC and claimed by another EC is kept by the target EC (step 4.3 condition: carved-out only when NOT in `targetSources`). This is how "sourcing a descendant of another EC's ancestor source" is handled — if the target sources the more-specific node, it wins.

**Output:** The `resolved` map values, in insertion order.

**Role definitions:**
- `"source"`: explicitly listed in the active direction's `sourceCapabilityIds`.
- `"implicit"`: included via subtree descent from a source; not itself an explicit source.
- `"carved-out"`: would be included via a source's subtree, but is itself an explicit source of a different active EC's direction. Its entire subtree carries the same carved-out status (because traversal stops at the carved-out node).

**Count definitions:**

| Field | Definition |
|---|---|
| `sourceCount` | `len(direction.sourceCapabilityIds)` — count of explicit sources, regardless of composition result |
| `includedCount` | Count of resolved capabilities where `role != "carved-out"` |
| `carvedOutCount` | Count of resolved capabilities where `role == "carved-out"` |
| `domainCount` | Distinct `businessDomainId` values among included (non-carved-out) capabilities, excluding null |

For `EnterpriseCapability.includedCapabilityCount`, use the same `includedCount` definition. For `EnterpriseCapability.domainCount`, use the same `domainCount` definition.

### R4 — EC must be active

On `POST /enterprise-capabilities/{id}/direction` (capture), check `ec.active`. If false, return 409 with message `"Directions can only be captured on active enterprise capabilities."` No `details` or `_links` are included in this 409 body.

### R5 — Agreed-direction immutability

On `POST .../direction/sources`, `DELETE .../direction/sources/{capabilityId}`, and `PUT .../direction`: if the active direction has `status == "agreed"`, return 409 with the immutable body. The immutable 409 includes `"details": { "directionStatus": "agreed" }` and `"_links": { "x-reject": {...} }`.

### R8 — Draft cardinality relaxation

At capture time (POST, creates draft) and at source mutation time (add-source, exclude): **no minimum source count is enforced**. An empty source set and a single source are both valid for a draft.

At `POST .../direction/propose`: enforce type cardinality before transitioning:
- `type == "consolidate"`: requires `len(sourceCapabilityIds) >= 2`
- `type == "decompose"`: requires `len(sourceCapabilityIds) >= 2`
- `type == "stay"`: no minimum

Return 400 (not 409) when cardinality fails at propose.

### R9 — Actor on `DirectionSourceCapabilitiesChanged`

The `DirectionSourceCapabilitiesChanged` event must record the acting architect (user ID / subject from the session). This requires adding an actor field to the event payload. The `BaseEvent` struct does not currently carry an actor field — this is a backend-internal decision (see §7 open items). The frontend is not involved in this.

---

## 5. Suggested Go DTOs

These structs belong in the read-model packages of their respective bounded contexts. Use `omitempty` on optional/nullable pointer fields and on `_links`. `types.Links` is the existing `map[string]types.Link` alias.

```go
// --- enterprisearchitecture read model ---

type EnterpriseCapabilityDTO struct {
    ID                     string      `json:"id"`
    Name                   string      `json:"name"`
    Description            string      `json:"description"`
    Category               string      `json:"category"`
    Active                 bool        `json:"active"`
    TargetMaturity         *int        `json:"targetMaturity,omitempty"`
    IncludedCapabilityCount int        `json:"includedCapabilityCount"`
    DomainCount            int         `json:"domainCount"`
    CreatedAt              time.Time   `json:"createdAt"`
    UpdatedAt              *time.Time  `json:"updatedAt,omitempty"`
    Links                  types.Links `json:"_links,omitempty"`
}

type IncludedCapabilityItemDTO struct {
    CapabilityID       string      `json:"capabilityId"`
    Name               string      `json:"name"`
    Level              string      `json:"level"`
    BusinessDomainID   *string     `json:"businessDomainId"`
    BusinessDomainName *string     `json:"businessDomainName"`
    Role               string      `json:"role"`
    CarvedOutBy        *CarvedOutByDTO `json:"carvedOutBy"`
    Links              types.Links `json:"_links,omitempty"`
}

type CarvedOutByDTO struct {
    EnterpriseCapabilityID   string `json:"enterpriseCapabilityId"`
    EnterpriseCapabilityName string `json:"enterpriseCapabilityName"`
}

type CompositionDomainGroupDTO struct {
    BusinessDomainID   *string                     `json:"businessDomainId"`
    BusinessDomainName *string                     `json:"businessDomainName"`
    Items              []IncludedCapabilityItemDTO  `json:"items"`
}

type CompositionMetaDTO struct {
    SourceCount    int `json:"sourceCount"`
    IncludedCount  int `json:"includedCount"`
    CarvedOutCount int `json:"carvedOutCount"`
    DomainCount    int `json:"domainCount"`
}

type CompositionResponseDTO struct {
    Data  []CompositionDomainGroupDTO `json:"data"`
    Meta  CompositionMetaDTO          `json:"meta"`
    Links types.Links                 `json:"_links,omitempty"`
}

type SourceCandidateDTO struct {
    CapabilityID                    string                          `json:"capabilityId"`
    Name                            string                          `json:"name"`
    Level                           string                          `json:"level"`
    ParentID                        *string                         `json:"parentId"`
    BusinessDomainID                *string                         `json:"businessDomainId"`
    BusinessDomainName              *string                         `json:"businessDomainName"`
    Eligible                        bool                            `json:"eligible"`
    IneligibilityReason             *string                         `json:"ineligibilityReason"`
    ConflictingEnterpriseCapability *ConflictingECDTO               `json:"conflictingEnterpriseCapability"`
    Links                           types.Links                     `json:"_links,omitempty"`
}

type ConflictingECDTO struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type PaginationDTO struct {
    HasMore bool   `json:"hasMore"`
    Limit   int    `json:"limit"`
    Cursor  string `json:"cursor"`
}

type SourceCandidatesResponseDTO struct {
    Data       []SourceCandidateDTO `json:"data"`
    Pagination PaginationDTO        `json:"pagination"`
    Links      types.Links          `json:"_links,omitempty"`
}

// --- architecturedirection read model ---

type DirectionSourceCapabilityDTO struct {
    ID                 string  `json:"id"`
    Stale              bool    `json:"stale"`
    Name               *string `json:"name"`
    BusinessDomainID   *string `json:"businessDomainId,omitempty"`
    BusinessDomainName *string `json:"businessDomainName,omitempty"`
}

type DirectionDTO struct {
    ID                     string                         `json:"id"`
    EnterpriseCapabilityID string                         `json:"enterpriseCapabilityId"`
    Type                   string                         `json:"type"`
    Status                 string                         `json:"status"`
    Horizon                string                         `json:"horizon"`
    Narrative              *string                        `json:"narrative,omitempty"`
    SourceCapabilities     []DirectionSourceCapabilityDTO `json:"sourceCapabilities"`
    Placements             []DirectionPlacementDTO        `json:"placements"`
    CreatedAt              time.Time                      `json:"createdAt"`
    UpdatedAt              *time.Time                     `json:"updatedAt,omitempty"`
    Links                  types.Links                    `json:"_links,omitempty"`
}

type DirectionPlacementDTO struct {
    TargetBusinessDomainID string  `json:"targetBusinessDomainId"`
    ResultingName          *string `json:"resultingName,omitempty"`
}

type ECDirectionResponseDTO struct {
    Direction *DirectionDTO `json:"direction"`
    Links     types.Links   `json:"_links,omitempty"`
}

// --- preview (architecturedirection handler) ---

type PreviewIncludedCapabilityDTO struct {
    CapabilityID       string          `json:"capabilityId"`
    Name               string          `json:"name"`
    Level              string          `json:"level"`
    BusinessDomainID   *string         `json:"businessDomainId"`
    BusinessDomainName *string         `json:"businessDomainName"`
    Role               string          `json:"role"`
    CarvedOutBy        *CarvedOutByDTO `json:"carvedOutBy"`
}

type SourceEligibilityDTO struct {
    CapabilityID                    string            `json:"capabilityId"`
    Eligible                        bool              `json:"eligible"`
    IneligibilityReason             *string           `json:"ineligibilityReason"`
    ConflictingEnterpriseCapability *ConflictingECDTO `json:"conflictingEnterpriseCapability"`
}

type CompositionPreviewMetaDTO struct {
    SourceCount    int `json:"sourceCount"`
    IncludedCount  int `json:"includedCount"`
    CarvedOutCount int `json:"carvedOutCount"`
}

type CompositionPreviewResponseDTO struct {
    IncludedCapabilities []PreviewIncludedCapabilityDTO `json:"includedCapabilities"`
    SourceEligibility    []SourceEligibilityDTO         `json:"sourceEligibility"`
    Meta                 CompositionPreviewMetaDTO      `json:"meta"`
    Links                types.Links                    `json:"_links,omitempty"`
}
```

**Swaggo annotation note.** Every handler must carry the full godoc block per EASI API Standards. Tag composition and source-candidates handlers with `@Tags enterprisearchitecture`; direction handlers with `@Tags architecturedirection`. Because `GET /capabilities/source-candidates` does not sit under an `{id}` path, its `@Router` is `/capabilities/source-candidates [get]`.

**chi route ordering caveat.** In the chi router setup, register `GET /capabilities/source-candidates` before `GET /capabilities/{id}`. chi matches routes in registration order for the same path depth, and `source-candidates` is a literal segment that would otherwise be swallowed by the `{id}` wildcard.

---

## 6. Removed Surface

The following must be hard-deleted, not deprecated:

**Endpoints to remove:**

| Method | Path | Reason |
|---|---|---|
| `GET` | `/enterprise-capabilities/{id}/links` | Linking superseded by direction sources |
| `POST` | `/enterprise-capabilities/{id}/links` | Linking superseded by direction sources |
| `DELETE` | `/enterprise-capabilities/{id}/links/{linkId}` | Linking superseded by direction sources |

**Response fields to remove:**

| DTO | Field | Replacement |
|---|---|---|
| `EnterpriseCapabilityDTO` | `linkCount` | `includedCapabilityCount` |
| `EnterpriseCapabilityDTO`'s `_links` | `x-links` | removed |
| `EnterpriseCapabilityDTO`'s `_links` | `x-create-link` | removed |

**Backend artifacts to delete:**

- `EnterpriseCapabilityLink` aggregate and its event handlers.
- `EnterpriseCapabilityLinked` and `EnterpriseCapabilityUnlinked` event types and their deserializers/upcasters.
- Link read model projector(s) and the database table(s) backing them.
- Migration: hard-delete all rows from the link table(s) — no data migration into directions.
- Any `linkCount` population in the EC read model projector.

**Spec note vs. stub divergence.** The stub's `buildEnterpriseCapabilityDto` never emits `x-links` or `x-create-link` on the EC DTO (they are already absent in the stub). This confirms the FE ships without them. The backend must not include them.

---

## 7. Open Items for Backend

These are decisions the stub could not pin down. Each is a **decision needed** before implementation is complete.

**O1 — Pagination cursor format for `source-candidates`.** The stub always returns `cursor: ""` and `hasMore: true/false` based on slice truncation, but never provides a cursor the client can use to fetch the next page. The FE currently does not implement infinite-scroll or next-page fetching on this endpoint. Backend must decide: keyset cursor (e.g., base64-encoded last-seen `capabilityId`) or offset. If the FE does not use the cursor today, a placeholder empty string is acceptable short-term, but the format must be documented before the client is extended.

**O2 — R9 actor field placement.** The spec says `DirectionSourceCapabilitiesChanged` must carry the acting architect, and that `BaseEvent` does not currently carry an actor. Backend must decide whether to add `actor` to `BaseEvent` (affects all events) or add it only to the `DirectionSourceCapabilitiesChanged` payload. The FE is not involved.

**O3 — R8 cardinality rules for `decompose` and `stay`.** The spec says consolidate requires ≥2 sources at propose. The stub does not model R8 enforcement at the propose transition (it just transitions). Backend must confirm: does `decompose` also require ≥2? Does `stay` have any source count constraint? The contract above assumes `consolidate` and `decompose` both require ≥2; adjust if the domain rules differ.

**O4 — `direction.updatedAt` on first create.** The stub sets `updatedAt: undefined` on a freshly captured direction and only sets it on mutation. Backend must confirm this is the intended behavior (not setting `updatedAt = createdAt` on create). The FE types `updatedAt` as `string | undefined` so both are valid — the stub convention is `undefined/absent` until first mutation.

**O5 — Composition read-model materialization strategy.** The composition algorithm (R2) requires the full capability tree and all active directions. Backend must decide whether this is computed on-the-fly per request (simpler, correct by construction, acceptable for current data volumes) or maintained as a projected read model (more complex, required at scale). The contract is identical either way; this is an internal implementation concern.

**O6 — `stale` source detection.** `DirectionSourceCapabilityDTO.stale` is `true` when the capability ID in the direction's source set no longer exists in the capability read model. Backend must decide how to detect staleness: join against the capability table at query time, or maintain a flag in the direction read model when capability-deleted events arrive. The FE renders stale sources with a warning badge.

**O7 — Transition guards for `propose`, `agree`, `reject`.** The stub's transition handler accepts any action on any non-rejected active direction (it does not enforce `propose` is only valid from `draft`, or `agree` only from `proposed`). The backend aggregate must enforce these state machine guards and return 404 (or 409 — decide) when the transition is not valid from the current status. The contract above documents 404 for invalid-status transitions, matching the pattern of "no active direction in the right state" = not found for that operation.

---

## Stub vs. Spec Divergence Notes

Where the stub and the spec prose differ, the stub governs (the FE ships against it).

| Item | Stub behavior | Spec prose | Resolution |
|---|---|---|---|
| R5 immutable 409 on `PUT /direction` | Stub does not model `PUT /direction` at all (no handler). The FE's `directionApi.update` sends `PUT` but the stub has no handler for it. | Spec mentions direction update is in scope. | Backend must implement `PUT /direction` per §2.7. The 409 shape for immutability is extrapolated from the add-source handler (which does model R5). |
| `placements` in `Direction` DTO | Stub always returns `placements: []` | Spec defers placements sub-feature | Backend should persist and return `placements` as an array; capture does not populate it. |
| Transition action routing | Stub uses a single parametric handler `POST .../direction/:action` matching `propose`, `agree`, `reject` | — | Backend should implement three separate named routes in chi: `POST .../direction/propose`, `POST .../direction/agree`, `POST .../direction/reject`. The FE calls each by its explicit path. |
| `cursor` in pagination | Stub always returns `""` | — | Treat as opaque string; see O1. |
