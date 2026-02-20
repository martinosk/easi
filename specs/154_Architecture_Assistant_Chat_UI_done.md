# Architecture Assistant Chat UI

**Status**: pending

**Series**: Chat with your Architecture (3 of 3)
- Spec 152: AI Assistant Configuration
- Spec 153: Architecture Assistant Backend
- **Spec 154**: Architecture Assistant Chat UI (this spec)

**Depends on:** [153_Architecture_Assistant_Backend](153_Architecture_Assistant_Backend_pending.md)

## User Value

> "As an architect, I want a chat panel I can open while looking at the architecture canvas, so I can ask questions without leaving my workspace."

> "As a stakeholder, I want to see which tools the assistant is using, so I can trust the response is grounded in real data."

## Description

Frontend chat panel that consumes the SSE streaming backend from spec 153. Delivered in phases: Phase 1 integrates with a no-tools POC endpoint, Phase 2 adds conversation history and tool indicators. Visibility gated on the `x-assistant` HATEOAS link.

## Design Decisions

- **Slide-out panel (not a page)**: Slides in from the right, overlaying current content. Users keep the canvas visible while chatting. Global overlay controlled by state, rendered as sibling to `DialogManager`. No route change.
- **HATEOAS-gated visibility**: Chat button appears only when `x-assistant` link is present in current-session response. This encodes both `assistant:use` permission AND `configured` status server-side. No hardcoded permission checks. Requires extending session `_links` support.
- **SSE via fetch + ReadableStream**: Not `EventSource` (which only supports GET). Custom `useChat` hook wraps `fetch` POST with streaming body parsing.
- **Single endpoint from day one**: UI always uses `POST /assistant/conversations/{id}/messages`. Creates a conversation on panel open, sends messages to it. Server persists messages and manages context.
- **Optimistic UI**: User message appears immediately. Assistant response streams token-by-token. Tool calls appear as collapsible indicators inline.
- **YOLO toggle with safe default**: Input area includes `YOLO (allow changes)` checkbox, default off. This is display text only. Backend receives `allowWriteOperations=false` when off and `allowWriteOperations=true` when on; actual writes still depend on user permissions.
- **Write operations visible**: In Phase 2, mutation tool calls show pencil/trash icons instead of search icon. Confirmation happens in natural language (system prompt enforced), no special confirmation UI.
- **Conversation list in panel**: Header area with conversation dropdown, new/switch/close controls.

---

## Panel Behavior

- 400px wide on desktop (>768px), full-width on mobile. Fixed position overlay, doesn't push content.
- Opens/closes via navigation chat button. Closes on Escape.
- On open: fetches conversation list, resumes last conversation or shows empty state.
- In Phase 1: no conversation list; panel creates a new conversation on open and sends messages to it. History persisted server-side.
- Empty state shows prompt suggestions: "What applications are in the Finance domain?", "Show me a portfolio summary", "Create a new application called 'Payment Gateway'"

---

## Message Display

**User messages:** Right-aligned, primary color background.

**Assistant messages:** Left-aligned, neutral background. Supports markdown via `react-markdown` + `remark-gfm`. No `rehype-raw` (XSS prevention). Streaming shows blinking cursor, content updates incrementally.

**Tool call indicators (Phase 2):** Shown inline between user and assistant messages.
- Running: pulsing dot + friendly label + "Looking up data..."
- Completed: check icon + label + result preview (collapsed, expandable)
- Error: warning icon + error message
- Visual distinction: search icon for queries, pencil for create/update, trash for delete

**Tool display names:** Map internal names to user-friendly labels (e.g. `list_applications` → "Searching applications", `create_application` → "Creating application", `delete_capability` → "Deleting capability").

---

## Chat Input

Multiline textarea, auto-grows up to 4 lines. Send on Enter (Shift+Enter for newline). Disabled during streaming. Max 2000 characters. Placeholder: "Ask about your architecture..."

Below input: checkbox label `YOLO (allow changes)`. Helper text: `When off, assistant can read only. When on, assistant may apply changes you are already permitted to make.`

---

## SSE Event Handling

| Event | UI Action |
|---|---|
| `token` | Append to current assistant message |
| `tool_call_start` | (Phase 2) Add ToolCallIndicator in "running" state |
| `tool_call_result` | (Phase 2) Update indicator to "completed" with preview |
| `thinking` | Show status text between tool calls |
| `ping` | No-op (keepalive) |
| `done` | Mark complete, invalidate conversation list cache |
| `error` | Display inline as system message |

Error display: "Not configured" links to Settings (if user has `metamodel:write`). "Permission denied" suggests contacting admin. Network errors show "Connection lost. Click to retry."

---

## State Management

Panel open/close and active conversation in Zustand store. Message state in `useChat` hook (React Query + local streaming buffer). Conversation list via React Query.

`yoloEnabled` can be used as panel-local UI state (default `false`), but Phase 2 send-message payload uses `allowWriteOperations`.

Availability check reads `x-assistant` from session `_links` loaded at app init — no extra API call.

---

## Cache Invalidation

Self-contained — assistant mutations only affect assistant query cache. No cross-feature invalidation (conversations don't modify architecture data).

Invalidate conversation list on `done` SSE event (title and `lastMessageAt` change). Create/delete conversation invalidate the list via standard mutation effects.

---

## Checklist

- [ ] Specification approved
- [x] Dependencies: `react-markdown`, `remark-gfm`
- [x] Session `_links` extended for `x-assistant` HATEOAS discovery
- [x] Phase 1: Chat using `POST /assistant/conversations/{id}/messages` (server-side history)
- [x] Chat panel (slide-out, responsive, transitions)
- [x] Message display (user/assistant variants, markdown, streaming cursor)
- [x] YOLO checkbox (default off)
- [x] Tool call indicators (Phase 2: running/completed/error, read/write icons, expandable)
- [x] Chat input (multiline, Enter to send, disabled during streaming)
- [x] Chat button in navigation (HATEOAS-gated)
- [x] SSE streaming hook with event parsing
- [x] Phase 2: Conversation management (list, create, switch, delete)
- [x] Cache invalidation (conversation list on done event — Phase 2, no conversation list yet)
- [x] MSW test handlers
- [x] Tests and build passing
