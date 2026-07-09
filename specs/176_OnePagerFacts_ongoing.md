# 176 — One-Pager Facts Capture

> **Status:** pending
> **Depends on:** 175 — OnePagerConfiguration (configuration aggregate and custom field definitions must exist)

---

## Problem Statement

Spec 175 lets a tenant admin define which custom fields the One-Pager for each subject
type carries — but no value can be recorded against them yet. Architects need to capture
stakeholder-facing facts that are not part of the base domain model (a contact person, a
contract link, a review date) directly on the entity they are describing.

Data quality is the point of the feature: a Field Value is a typed, constructor-validated
domain value, never a free-text string. A contact person is a (name, email, company)
triple with a validated email; a link is a validated absolute URL with a label. This
slice delivers the `OnePagerFacts` aggregate, the six typed Field Value kinds, and a
configuration-driven "One-Pager" edit section on all six subject detail panels.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Enterprise Architect** | Record and correct typed facts on a subject as information arrives, incrementally, without being blocked by fields they cannot fill yet |
| **Stakeholder** | Trust that recorded facts are well-formed (valid emails, working links, real dates) |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Capture typed One-Pager facts on a subject

  Scenario Outline: Recording a valid value of each Field Type
    Given the <subject type> One-Pager Configuration defines an active custom field of type <field type>
    When I record <valid value> for that field on a subject
    Then the Field Value is saved and shown in the subject's One-Pager section

    Examples:
      | subject type   | field type     | valid value                                        |
      | Application    | Text           | "Runs on shared Kubernetes cluster"                |
      | Capability     | Number         | 42.5                                               |
      | Vendor         | Date           | 2026-03-01                                         |
      | Application    | Link           | label "MSA" with URL https://contracts.example.com |
      | Internal Team  | Selection      | an active option of the field                      |
      | Acquired Entity| Contact Person | name "A. Larsen", email "al@ext.example", company "Ext ApS" |

  Scenario Outline: Rejecting an invalid value of each Field Type
    Given an active custom field of type <field type> on the subject's configuration
    When I record <invalid value> for that field
    Then the write is rejected with a validation error and no Field Value is stored

    Examples:
      | field type     | invalid value                                  |
      | Text           | a whitespace-only string                       |
      | Text           | a string exceeding the length cap              |
      | Number         | a non-finite value (NaN, Infinity)             |
      | Date           | "March 1st" (not an ISO date)                  |
      | Link           | URL "ftp://x" or a relative path               |
      | Selection      | an option ID not defined on the field          |
      | Contact Person | empty name, or email "not-an-email"            |

  Scenario: Clearing a recorded value
    Given a subject has a Field Value recorded for a custom field
    When I clear that field
    Then the field shows no value in the One-Pager section

  Scenario: Recording against a retired field is rejected
    Given a custom field on the subject type's configuration has been retired
    When a record-value request arrives for that field
    Then the write is rejected
    And the retired field does not appear in the edit form

  Scenario: Value type mismatch is rejected
    Given an active custom field of type Date
    When I record a Text value for that field
    Then the write is rejected with a type-mismatch error

  Scenario: Retired Selection option is rejected on write but existing values render
    Given a Selection field where option "Tier 1" was recorded on a subject and later retired
    When I record option "Tier 1" on another subject
    Then the write is rejected
    But the existing "Tier 1" value on the first subject still renders, flagged as a retired option

  Scenario: Subject deletion archives its facts
    Given a subject with recorded Field Values
    When the subject is deleted in its owning context
    Then its One-Pager facts no longer appear anywhere
    And any subsequent record-value request for that subject is rejected

  Scenario: Required fields are soft at write time
    Given the configuration marks custom field "Contract link" as required
    When I save the One-Pager section leaving "Contract link" empty
    Then the save succeeds
    And the form visually highlights "Contract link" as a missing required field

  Scenario: One facts aggregate per subject
    Given no facts exist yet for a subject
    When I record the first Field Value for it
    Then a single OnePagerFacts aggregate is created for that subject
    And recording further values reuses the same aggregate
```

---

## Business Rules & Invariants

1. **One OnePagerFacts per Subject** — at most one facts aggregate exists per
   (tenant, subject type, subject ID); enforced at the command handler via read-model
   lookup with a DB unique constraint as backstop.
2. **Intrinsic identity** — the aggregate has its own UUID, never derived from the
   subject's ID; the subject is referenced via a `SubjectRef` value object
   (Subject Type + subject ID), scoped to the `onepagers` context.
3. **Implicit creation, verified subject** — the aggregate is created on the first value
   recorded; the creation handler verifies subject existence through the
   onepagers-defined subject port before creating.
4. **Typed values only** — every Field Value is one of the closed set of
   constructor-validated value objects: `TextValue` (trimmed, non-empty, length cap),
   `NumberValue` (finite decimal), `DateValue` (ISO date), `LinkValue{Label, URL}`
   (shared `URL` VO, absolute http(s)), `SelectionValue{OptionID}` (option exists on the
   field definition), `ContactPerson{Name, Email, Company}` (non-empty name, validated
   email, company optional).
5. **Value envelope** — every value is persisted, in event payloads and in the read
   model, as a discriminated `{type, version, value}` envelope (D9).
6. **Definition validation at the handler** — the command handler loads the current
   configuration read model and rejects writes when the field does not exist, is
   retired, the value type does not match the Field Type, or a Selection option is not
   active on the definition.
7. **Aggregate invariants** — the aggregate itself rejects a value without a field ID,
   holds at most one value per field (recording replaces), and rejects all writes once
   archived.
7a. **No-op writes are suppressed** — recording a value equal to the field's current
   value appends no event; the edit form submits only dirty fields on section save, so
   one edit never fans out into events for untouched fields.
8. **Archived on subject deletion** — a policy reacts to subject deletion events by
   appending `OnePagerFactsArchived` to the facts stream; the projector removes the
   read-model rows; archived facts can never be resurrected.
9. **Retired values persist** — retiring a field or a Selection option never removes
   recorded values; existing retired-option values render flagged.
10. **Soft required at write time** — missing required fields are highlighted in the
    edit form but never block saving (D2).
11. **Permission parity with the subject** — recording or clearing a Field Value
    requires the same write permission as editing the subject itself; reading facts
    requires the subject's read permission.

---

## Acceptance Criteria

- [ ] Recording a valid value of each of the six Field Types succeeds and the value
      appears in the subject's One-Pager section.
- [ ] Each invalid-value case in the rejection scenario outline fails with a validation
      error and stores nothing.
- [ ] Clearing a value removes it from the One-Pager section.
- [ ] Writes against retired fields, mismatched types, and retired Selection options
      are rejected; existing retired-option values still render, flagged.
- [ ] Deleting any of the six subject types archives its facts, removes its rows, and
      blocks further writes — verified for each subject type's real deletion event.
- [ ] Saving with an empty required field succeeds; the field is visually highlighted.
- [ ] Only one facts aggregate exists per subject, including under concurrent first
      writes (DB unique backstop).
- [ ] The One-Pager edit section appears on all six detail panels, driven by the
      subject type's configuration, hiding retired fields.
- [ ] The shared `URL` value object exists in the shared kernel with unit tests.
- [ ] Every event and read-model value is stored as a `{type, version, value}` envelope.
- [ ] Endpoints are permission-gated per rule 11 and documented via swagger.

---

## Architecture

### Ownership

The `onepagers` bounded context (created in spec 175) owns the facts aggregate, the
Field Value types, the read model, and the deletion policy. Supplier contexts
(`capabilitymapping`, `enterprisearchitecture`, `architecturemodeling`) are unchanged;
`onepagers` consumes their published-language deletion events and reads subject
existence through its own port.

### Domain Model

- **Aggregate `OnePagerFacts`** — one per Subject, own intrinsic UUID, `SubjectRef` VO,
  a fieldID → Field Value map, and an archived flag. Events produced:
  `FieldValueRecorded(fieldID, value envelope)`, `FieldValueCleared(fieldID)`,
  `OnePagerFactsArchived(reason: subject deleted)`.
- **Field Value VOs** per rule 4, following the existing constructor-validated style
  (`Expert`/`ContactInfo` precedent in `architecturemodeling/domain/valueobjects`).
- **Shared `URL` VO** added to `internal/shared/eventsourcing/valueobjects` (D5).
- **Subject port** — introduced here in minimal form: subject existence check only,
  implemented by composition-root adapters per the `direction_composition_adapters.go`
  precedent (D8). Full built-in field reading arrives with the One-Pager view (177).
- **Cross-aggregate validation split** — handler validates against the configuration
  read model (rule 6); aggregate guards its own invariants (rule 7). The
  retire-field-vs-record-value race between the two aggregates is accepted eventual
  consistency: retired fields keep their values by design, so no invariant is violated.

### API Surface

Facts for a subject as a sub-resource of the one-pager resource: an operation to get all
Field Values for a subject, an operation to record a value for one field (idempotent
replace), and an operation to clear one field. Requests and responses carry the value
envelope. Gating follows the existing static-permission middleware pattern
(`auth.RequirePermission` as in `architecturemodeling/infrastructure/api/routes.go`),
resolved from the subject type: `components:write` for Application, Acquired Entity,
Vendor, and Internal Team; `capabilities:write` for Capability; `enterprise-arch:write`
for Enterprise Capability — matching `:read` permissions for reads. The frontend gates
the edit section via HATEOAS links, never by role inspection.

### Persistence

- Events in the existing event store; value payloads as envelopes (D9).
- Read model `onepagers.one_pager_facts` — one row per (subject, field), PK
  `(tenant_id, subject_type, subject_id, field_id)`, envelope JSONB column plus
  extracted columns for the common render path, RLS tenant-isolation policy per the
  existing table shape.
- The PK doubles as the one-facts-per-subject uniqueness backstop at the
  (tenant_id, subject_type, subject_id) level for aggregate creation.

### Frontend

- A configuration-driven "One-Pager" section on all six detail panels (Capability,
  Enterprise Capability, Application, Acquired Entity, Vendor, Internal Team).
- The section's Zod schema is generated dynamically from the field definitions;
  RHF + `zodResolver` per existing form conventions in `frontend/src/lib/schemas/`.
- Per-Field-Type inputs: text input, number input, date picker, label + URL pair,
  select, contact-person triple (name, email, company).
- Retired fields are not shown; missing required fields are highlighted, not blocking.

### Cross-Context Integration

`onepagers` subscribes to the published-language deletion events of all six subject
types and archives the corresponding facts stream:

| Supplier context | Event |
|---|---|
| capabilitymapping | `CapabilityDeleted` |
| enterprisearchitecture | `EnterpriseCapabilityDeleted` |
| architecturemodeling | `ApplicationComponentDeleted` |
| architecturemodeling | `AcquiredEntityDeleted` |
| architecturemodeling | `VendorDeleted` |
| architecturemodeling | `InternalTeamDeleted` |

No events flow from `onepagers` to other contexts.

---

## Design Decisions

1. **Soft required at write time (D2, write-time aspect)** — the edit form highlights
   missing required fields but saving partial facts is allowed; an EA tool must accept
   incremental data entry. Alternative: hard gate on required fields (rejected — flipping
   optional → required would freeze edits on every incomplete pre-existing subject).
   Completeness computation and indicators are spec 178.
2. **Contact Person is distinct from Expert (D3)** — Contact Person is a custom Field
   Type for stakeholder-facing contacts including external companies; Experts remain a
   built-in catalog concern. Alternative: reuse the Expert VO (rejected — different
   meaning and shape, no shared kernel between the contexts).
3. **Shared `URL` value object (D5)** — added to
   `internal/shared/eventsourcing/valueobjects` since no URL VO exists and Link fields
   need one. Alternative: context-local URL VO (rejected — a URL is domain-agnostic and
   the shared kernel is the designated home per the design doc).
4. **Value envelope from day one (D9)** — `{type, version, value}` in events and read
   model, so new Field Types or evolved value shapes are an upcaster exercise.
   Alternative: bare JSONB values (rejected — schema evolution becomes JSONB
   archaeology).
5. **Handler-side definition validation** — the configuration and facts aggregates stay
   independent; the handler composes them via the read model. Alternative: facts
   aggregate loads the configuration aggregate (rejected — cross-aggregate transaction,
   larger consistency boundary than the invariants require).
6. **Archival via the facts stream, not just the read model** — subject deletion closes
   the aggregate in the domain, making resurrection impossible. Alternative: projector
   row deletion only (rejected — a later command could recreate facts for a dead
   subject).
7. **No-op suppression caps stream growth** — with rule 7a, facts streams grow only with
   real edits at human pace. Facts streams are the designated first snapshot candidate
   if write-path replay ever profiles hot — decided by profiling, not preemptively,
   mirroring the design doc's D6 posture on the completeness cache.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Soft required at write time | Incomplete facts can be saved indefinitely | Visual highlight now; completeness indicators in spec 178 |
| Handler-side definition validation | Retire-vs-record race is eventually consistent | Retired fields keep values by design; no invariant violated |
| Value envelope everywhere | Slightly heavier payloads and mapping code | Uniform (de)serialization; upcasters instead of data migrations |
| Minimal subject port (existence only) | Port grows again in 177 | Interface is consumer-defined; extension is additive |

---

## Checklist

- [x] Specification ready
- [ ] Implementation done
- [ ] Unit tests implemented and passing
- [ ] Integration tests implemented if relevant
- [ ] API documentation updated
- [ ] User sign-off
