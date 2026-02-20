-- Migration: Add Release 0.26.0
-- Description: Adds release notes for version 0.26.0

INSERT INTO releases (version, release_date, notes, created_at) VALUES
('0.26.0', '2026-02-20', '## What''s New in v0.26.0

### Major

- **Chat with your Architecture**: Introduced an AI-powered architecture assistant that lets architects ask natural-language questions about their application landscape, business capabilities, and value streams — and get answers grounded in real portfolio data. The assistant supports both OpenAI and Anthropic LLM providers, configured per-tenant with encrypted API key storage.
- **Architecture Assistant Tools**: The assistant can query applications, capabilities, business domains, and value streams, as well as create, update, and delete architecture entities via natural language. A "YOLO mode" toggle controls whether the assistant can make changes (read-only by default).
- **Architecture Assistant Chat UI**: A slide-out chat panel accessible from the navigation bar lets users interact with the assistant while keeping the canvas visible. Features streaming responses, markdown rendering, tool call indicators, and conversation history management.
- **AI Configuration Settings**: New Settings tab for tenant admins to configure which LLM provider (OpenAI or Anthropic), model, and parameters the assistant uses, with a connection test button and data residency notice.

### Bugs

- Fixed inherited realisations not being removed when deleting a child capability
- Fixed capability reparenting failing for L4 leaf nodes when updating descendant levels
- Fixed permission inconsistency on capability mapping routes

### API

- New endpoints: `GET/PUT /assistant-config`, `POST /assistant-config/connection-tests` — AI assistant configuration management
- New endpoints: `GET/POST /assistant/conversations`, `GET/DELETE /assistant/conversations/{id}`, `POST /assistant/conversations/{id}/messages` (SSE) — conversation and chat streaming
- New HATEOAS link `x-assistant` on `GET /auth/sessions/current` — controls chat button visibility
- New permission: `assistant:use` (admin and architect roles)', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
