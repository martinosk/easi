# 185 — One-Pager Number Field Bounds

> **Status:** pending
> **Depends on:** 175 — OnePagerConfiguration (Number custom fields must exist), 176 — OnePagerFacts (NumberValue recording must exist)
> **Design:** extends [Configurable One-Pagers](../docs/specs/configurable-one-pagers.md); conforms to decisions D1–D10, in particular D2 (config changes never invalidate recorded facts)

---

## Problem Statement

Spec 175 lets a tenant admin define a Number custom field on a One-Pager, and spec 176
lets an architect record any finite number against it. Nothing today stops an architect
recording a negative headcount, an annual cost of zero, or a maturity score of 200 on a
0–5 scale — a Number field validates only that the value is a finite decimal, not that it
makes sense for the field it belongs to.

Tenant admins know the sensible range for a Number field the moment they define it — a
percentage is 0–100, a rating is 1–5, a cost has no natural upper bound but must not be
negative. This slice lets an admin declare that range as part of the field's definition
and has the facts-capture handler enforce it on every new write, so a Number field can
guarantee its values are not just numeric but plausible.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Tenant Administrator** | Constrain a Number field to a sensible range when defining or later editing it, without disturbing the field's identity or any facts already recorded against it. |
| **Enterprise Architect** | Get an immediate, clear rejection when recording a number outside the field's declared range, instead of silently corrupting a stakeholder-facing fact sheet. |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Bounds on Number custom fields

  Scenario Outline: Defining a Number field with a bound
    Given the One-Pager Configuration for subject type "Capability"
    When the administrator defines a Number field "<name>" with <bounds>
    Then "<name>" is listed as an active Number field with <bounds>

    Examples:
      | name              | bounds                        |
      | Maturity score    | minimum 0 and maximum 5        |
      | Headcount         | minimum 0, no maximum          |
      | Discount cap      | no minimum, maximum 100        |
      | Free-form measure | no minimum and no maximum      |

  Scenario: Minimum greater than maximum is rejected at definition time
    Given the One-Pager Configuration for subject type "Capability"
    When the administrator attempts to define a Number field with minimum 10 and maximum 5
    Then the request is rejected with an error explaining the minimum must not exceed
      the maximum
    And no field is defined

  Scenario: Changing bounds later leaves existing facts untouched
    Given an active Number field "Maturity score" with minimum 0 and maximum 5
    And a Capability has a recorded value of 3 for that field
    When the administrator tightens the maximum to 3
    Then the field definition now shows maximum 3
    And the Capability's recorded value of 3 is unchanged
    And no facts event is appended for that Capability

  Scenario: Clearing a bound
    Given an active Number field "Headcount" with minimum 0 and maximum 500
    When the administrator removes the maximum, keeping the minimum
    Then the field shows minimum 0 and no maximum

  Scenario: Recording a value inside the bounds succeeds
    Given an active Number field "Maturity score" with minimum 0 and maximum 5
    When an architect records 4 for that field on a Capability
    Then the value is saved and shown in the Capability's One-Pager section

  Scenario Outline: Recording a value outside the bounds is rejected
    Given an active Number field "Maturity score" with minimum 0 and maximum 5
    When an architect records <value> for that field on a Capability
    Then the write is rejected with a validation error naming the field's bounds
    And no Field Value is stored

    Examples:
      | value |
      | -1    |
      | 5.1   |

  Scenario: A value recorded before a bound was tightened still renders, flagged
    Given an active Number field "Maturity score" with minimum 0 and maximum 5
    And a Capability has a recorded value of 5 for that field
    When the administrator tightens the maximum to 3
    Then the Capability's One-Pager section still shows 5 for that field
    And the value is flagged as outside the field's current bounds

  Scenario: Bounds do not apply to other field types
    Given an active Text field "Business summary"
    Then no bounds configuration is offered for that field

  Scenario: Only administrators can change bounds
    Given a signed-in architect without the tenant-configuration write permission
    Then the configuration response carries no HATEOAS affordance for changing bounds
    And any request to change bounds they submit is rejected as forbidden
```

---

## Business Rules & Invariants

1. **Bounds are Number-only definition metadata** — minimum and maximum are optional
   attributes of a Number custom field only; no other Field Type carries them. Setting
   bounds on a non-Number field is rejected.
2. **Either, both, or neither bound may be set** — a Number field is valid with no
   bounds, a minimum only, a maximum only, or both.
3. **Minimum must not exceed maximum** — when both bounds are set, minimum ≤ maximum;
   violating requests are rejected and no change is applied.
4. **Bounds never change identity** — setting or changing bounds does not affect the
   field's FieldID, type, name, help text, required flag, options, or display position
   (same class of change as renaming, per spec 175 rule 4).
5. **Config changes never invalidate recorded facts** — tightening, loosening, or
   clearing bounds never mutates, blocks re-display of, or appends any facts event for
   already-recorded NumberValues (design doc D2, spec 176 rule 9 precedent).
6. **Bounds gate only new writes** — the facts command handler validates a NumberValue
   against the field's *current* bounds, read from the configuration read model, exactly
   as it validates field existence, retirement, and type match today (spec 176 rule 6).
   A value already stored is never re-validated.
7. **Out-of-bounds existing values render flagged** — when a stored NumberValue falls
   outside the field's current bounds, the One-Pager view marks it as outside range,
   mirroring the retired-Selection-option flag (spec 176 rule 9, spec 177 field
   assembly). The value itself is never altered or hidden.
8. **Every bounds change is a past-tense domain event** — `NumberFieldBoundsChanged`,
   carrying the field's full new bounds state (both optional values, independent of
   whether the caller is setting, tightening, loosening, or clearing). Replay
   reconstructs the current bounds from the latest such event for the field.
9. **Writes are admin-only** — changing bounds is gated by `PermMetaModelWrite`; reading
   bounds is gated by `PermMetaModelRead`, the same matrix as every other configuration
   change (spec 175 rule 12). HATEOAS advertises the change-bounds affordance only to
   authorized callers.
10. **Bounds are a soft hint at the input, a hard gate at the handler** — the frontend
    Number input may use the bounds to constrain the control, but rejection of an
    out-of-range value is always decided server-side against the current configuration
    read model, never trusted from the client.

---

## Acceptance Criteria

- [x] An admin can define a Number field with a minimum only, a maximum only, both, or
      neither
- [x] Attempting to define or update a Number field with minimum > maximum is rejected
      and no change is applied
- [x] An admin can change or clear a Number field's bounds without altering its FieldID,
      type, name, help text, required flag, or position
- [x] Recording a NumberValue within the field's current bounds succeeds
- [x] Recording a NumberValue outside the field's current bounds is rejected at the
      handler with a validation error naming the violated bound, and nothing is stored
- [x] Changing a field's bounds appends no facts events and does not alter any
      previously recorded NumberValue
- [x] A NumberValue recorded before a bound was tightened continues to render in the
      One-Pager view, flagged as outside the field's current bounds
- [x] Attempting to set bounds on a non-Number field is rejected
- [x] Non-admin users receive no bounds-change HATEOAS affordance and their change
      requests return 403
- [x] Every bounds change is persisted as `NumberFieldBoundsChanged`; replay
      reconstructs the current bounds
- [x] Every BDD scenario above has at least one corresponding test
- [ ] Every modified file scores 10.0 in CodeScene per `easi-codehealth` — one file
      (`domain/aggregates/one_pager_configuration.go`) remains at 9.38 after two rounds
      of genuine refactoring; see implementation report for details

---

## Architecture

### Ownership

Entirely within the existing `onepagers` bounded context. No new aggregate, no new
bounded-context boundary, no cross-context integration. `OnePagerConfiguration` gains one
new event and one new command; the facts command handler (`onepagers` application layer)
gains one additional validation step reading the same configuration read model it already
reads.

### Domain Model

`CustomField` (value object within `OnePagerConfiguration`, spec 175) gains two optional
attributes, minimum and maximum, meaningful only when the field's type is Number — same
placement as the existing `required` flag and `helpText`. A constructor-level rule
enforces minimum ≤ maximum when both are present and rejects bounds on any non-Number
field, mirroring the existing `validateOptions` (Selection-only options) rule in the same
value object.

`OnePagerConfiguration.SetNumberFieldBounds(fieldID, min, max)` is a new command method
on the aggregate, following the same shape as `ChangeCustomFieldRequirement`: it loads the
active field, validates the Number-type and min-≤-max invariants, and raises
`NumberFieldBoundsChanged`. The event carries the field's full new bounds state (not a
delta), so replay always derives the current bounds from the most recent such event for
that field — no event ever needs to be interpreted in the context of a prior one.

`NumberValue` (spec 176 facts value object) is unchanged; it still only guarantees
"finite decimal". Bounds checking is definition validation, not value-object validation —
the same split spec 176 already draws between type/retirement checks (handler) and
"is this a real number" (value object constructor).

### API Surface

One new command endpoint on the existing configuration resource, `PermMetaModelWrite`-gated,
following the established one-endpoint-per-configuration-change pattern (spec 175): set a
Number field's bounds, taking the field ID, the optional minimum and maximum, and the
expected aggregate version for optimistic concurrency (409 on conflict, matching every
other configuration write). The configuration read (`PermMetaModelRead`) already returns
the full field definition; it gains the two optional bound values on Number fields.

The facts write endpoint (spec 176) is unchanged at the contract level: it still accepts
a NumberValue envelope and now additionally rejects out-of-bounds values with the same
validation-error shape used for retired-field and type-mismatch rejections today.

The One-Pager view read (spec 177) marks an out-of-bounds recorded value the same way it
already marks a retired-Selection-option value — an additional boolean alongside the
existing per-field render data.

### Persistence

No new tables and no migration beyond the existing event store. Bounds live inside the
same `configuration` JSONB document column in `onepagers.one_pager_configurations`
(spec 175) as part of each Number field's entry; the read model deserializes them like
every other field attribute. No change to `onepagers.one_pager_facts` (spec 176) — the
stored NumberValue and its envelope are unaffected by bounds.

### Frontend

Settings: the add/edit custom field form shows a minimum and a maximum numeric input only
when the selected field type is Number, following the existing pattern where the options
editor appears only for Selection — same conditional-section shape, same React Hook Form +
Zod validation (min ≤ max) mirrored client-side as a fast-fail UX hint ahead of the
authoritative server check.

Facts capture: the Number field's input control passes the field's bounds to the
underlying numeric control as soft min/max hints (so an obviously out-of-range value is
discouraged before submit) and surfaces the handler's rejection message when the server
still declines a value.

One-Pager view: a recorded Number value outside the field's current bounds renders with a
visible flag, next to the value, using the same visual treatment already established for
values referencing a retired Selection option.

### Cross-Context Integration

None. This slice is internal to `onepagers`; no other bounded context is touched, and no
new inbound or outbound event flows are introduced.

---

## Design Decisions

1. **One event, `NumberFieldBoundsChanged`, carrying full new state** — setting,
   tightening, loosening, and clearing bounds are all the same kind of definition edit
   (compare to `CustomFieldRequirementChanged`, a single boolean flag change), not a
   growing collection of independently-identified entries. Alternative: separate
   `NumberFieldBoundsSet` / `NumberFieldBoundsCleared` events, or the
   `SelectionOptionAdded`/`SelectionOptionRetired` two-event shape — rejected: Selection
   options are individually identified, retire-only history items; a Number field has at
   most one current bounds state, so a single "full replacement" event (like
   `CustomFieldRenamed` covering name and help text together) keeps the closed event list
   simpler without losing any information replay needs.
2. **Bounds live on the `CustomField` value object, not a separate entity** — they are
   two more optional attributes alongside `required` and `helpText`, not something with
   its own identity or lifecycle. Alternative: a `NumberFieldBounds` sub-entity modeled
   like `SelectionOption` — rejected, there is nothing to individually identify or retire;
   a field has exactly one current bounds state.
3. **Bounds are handler-enforced against the read model, exactly like existing facts
   validation** — reuses the same validation seam spec 176 rule 6 already established
   (field existence, retirement, type match); adding "within bounds" is one more check in
   the same place, not a new validation layer. Alternative: enforce bounds inside the
   `NumberValue` constructor — rejected, the value object cannot know a field's bounds
   without being handed configuration state at construction time, which breaks the
   existing clean separation between aggregate-owned invariants and cross-aggregate
   (config-vs-facts) validation (spec 176 rule 6/7 split).
4. **Existing out-of-bounds values render flagged, never hidden or mutated** — required
   by D2 (config changes never invalidate facts) and directly precedented by spec 176's
   retired-Selection-option flag, which already solves the identical "definition tightened
   after the fact" problem for a different Field Type. Alternative: silently keep showing
   the value with no indicator — rejected, an architect reviewing a one-pager has no way
   to know the number no longer fits the field's declared range, defeating the purpose of
   adding bounds at all. Alternative: block the one-pager from rendering until fixed —
   rejected, violates D2 outright.
5. **Minimum ≤ maximum enforced at the value-object constructor, not only the command
   handler** — the same invariant-placement rule spec 175's `validateOptions` already
   follows for Selection fields (aggregate-owned data invariants live in the aggregate,
   not only the handler). Alternative: handler-only validation — rejected, would let an
   invalid CustomField state exist inside the aggregate if any other code path ever
   constructed one directly.
6. **No bounds field on any non-Number Field Type** — bounds are meaningless for Text,
   Date, Link, Selection, and Contact Person; the constructor rejects them there the same
   way it already rejects Selection options on a non-Selection field. Alternative: a
   generic "constraints" bag usable by any field type — rejected as speculative, no other
   Field Type has a bounds-shaped requirement today.

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Single full-replacement bounds event | Event payload always carries both values even when only one changes | Matches `CustomFieldRenamed`'s existing precedent; payload is two small optional numbers |
| Handler-side bounds enforcement | Bounds and facts validation live in two aggregates, eventually consistent under a tighten-vs-record race | Same accepted trade-off already made for retired fields and retired Selection options (spec 176 rule 7) |
| Flagging instead of blocking out-of-bounds facts | A one-pager can display a number the current configuration disallows | Explicit visible flag makes the discrepancy obvious instead of silently wrong |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant
- [x] API documentation updated
- [ ] User sign-off
