# 199 — Assistant For Stakeholders

> **Status:** done
> **Depends on:** 153_Architecture_Assistant_Backend_done.md, 154_Architecture_Assistant_Chat_UI_done.md

---

## Problem Statement

The AI assistant is currently available only to admins and architects (`assistant:use` permission). Stakeholders — the read-only audience the assistant's grounded answers are most valuable to — cannot open the chat at all.

Stakeholders must get read-only assistant access. Because they hold no write permissions, the `YOLO (allow changes)` checkbox is meaningless for them and must not be shown. Per spec 154 and EASI convention, the frontend must not decide this from the role — visibility is driven by HATEOAS link presence in the current-session response.

---

## User Personas

| Persona | Needs |
|---------|-------|
| **Stakeholder** | Ask the assistant questions about the architecture, grounded in real data, without write affordances |
| **Architect / Admin** | Unchanged assistant experience, including the YOLO toggle |

---

## User-Facing Behavior (BDD Scenarios)

```gherkin
Feature: Assistant access for stakeholders

  Scenario: Stakeholder sees the chat button
    Given a stakeholder user and a configured AI assistant
    When the current session is loaded
    Then the session links contain "x-assistant"
    And the chat button is visible

  Scenario: Stakeholder gets no YOLO checkbox
    Given a stakeholder user with the chat panel open
    When the chat input is rendered
    Then no "YOLO (allow changes)" checkbox is shown
    And every message is sent with allowWriteOperations=false

  Scenario: Architect keeps the YOLO checkbox
    Given an architect user and a configured AI assistant
    When the current session is loaded
    Then the session links contain "x-assistant-write"
    And the chat input shows the YOLO checkbox

  Scenario: Backend ignores a forged write flag
    Given a stakeholder user
    When they POST a message with allowWriteOperations=true
    Then the assistant runs with write operations disabled
```

---

## Business Rules & Invariants

1. **Stakeholder may use the assistant** — the stakeholder role includes the `assistant:use` permission.
2. **`x-assistant` link unchanged** — present when the role has `assistant:use` and AI is configured.
3. **`x-assistant-write` link** — present only when the `x-assistant` conditions hold and the role has at least one subject write permission. Stakeholders never receive it.
4. **HATEOAS-gated YOLO** — the frontend renders the YOLO checkbox only when the session links contain `x-assistant-write`; it never inspects the role.
5. **Server-side clamp** — the send-message handler forces `allowWriteOperations=false` when the actor holds no write permission inside the agent permission ceiling, keeping the system prompt truthful regardless of client input.

---

## Acceptance Criteria

- [x] `RoleStakeholder.Permissions()` contains `PermAssistantUse`
- [x] Session links for a stakeholder (AI configured) contain `x-assistant` but not `x-assistant-write`
- [x] Session links for admin and architect (AI configured) contain both `x-assistant` and `x-assistant-write`
- [x] Chat input hides the YOLO checkbox when `x-assistant-write` is absent and always sends `allowWriteOperations=false`
- [x] Chat input shows the YOLO checkbox when `x-assistant-write` is present (existing behavior preserved)
- [x] Backend clamps `allowWriteOperations` to false for actors without any ceiling write permission

---

## Architecture

### Ownership

Auth context owns the role permission change and session link generation. Arch-assistant context owns the write-flag clamp. Frontend chat feature consumes the new link.

### API Surface

New session link relation `x-assistant-write` (href identical to `x-assistant`). No new endpoints; `SendMessageRequest` contract unchanged.

### Frontend

`SessionLinks` gains `x-assistant-write`. `App` derives write availability from the link and passes it to `ChatPanel` → `ChatInput`, which conditionally renders the YOLO block.

### Cross-Context Integration

None beyond the existing auth → assistant status checker. The clamp reuses the agent permission ceiling as the single definition of assistant-writable domains.

---

## Design Decisions

1. **Signal write capability via a session link, not conversation links** — the YOLO checkbox renders before any conversation exists, so a per-conversation link cannot gate it. Follows the `x-assistant` precedent from spec 154. Alternative: `_links` on the conversation resource (rejected — checkbox is visible pre-conversation).
2. **Role-based predicate lives in the auth context** — `buildSessionLinks` already gates `x-one-pager-quality` with a role helper; a `canWriteAnySubject` helper mirrors it. Alternative: importing the assistant ceiling into auth (rejected — cross-context import for a simple predicate).
3. **Clamp in the assistant handler using the ceiling** — the ceiling already ANDs actor permissions for tool execution, so a forged flag could not mutate data; the clamp additionally keeps the system prompt's write-mode statement truthful. Alternative: reject the request with 403 (rejected — degrades gracefully instead of breaking stakeholder chats that set the flag).

---

## Trade-offs

| Decision | Trade-off | Mitigation |
|----------|-----------|------------|
| Two write-capability predicates (auth link, assistant clamp) | Could drift | Both are permission-set driven and covered by unit tests on both sides |

---

## Checklist

- [x] Specification ready
- [x] Implementation done
- [x] Unit tests implemented and passing
- [x] Integration tests implemented if relevant (not relevant: no new endpoints, routes, or persistence — link generation and clamping are unit-covered)
- [x] API documentation updated (no Swagger annotation changes: no new endpoints, request/response contracts unchanged)
- [x] User sign-off
