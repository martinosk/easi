-- Migration: Add Release 0.28.0
-- Description: Adds release notes for version 0.28.0

INSERT INTO releases.releases (version, release_date, notes, created_at) VALUES
('0.28.0', '2026-04-20', '## What''s New in v0.28.0

### Minor
- The AI assistant can now query capability realisations by application, giving richer answers when exploring how capabilities are implemented.
- Anthropic provider connections now use the correct full API URL, resolving connectivity failures for Anthropic-backed assistants.

### Bugs
- Fixed incorrect URL construction for the Anthropic LLM client that caused connection failures.
- Fixed SSRF vulnerability in the Test Connection endpoint — endpoint URLs are now validated before any outbound request is made.
- Fixed misleading error messages when connection parameter resolution fails.', CURRENT_TIMESTAMP)
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  notes = EXCLUDED.notes;
