---
name: easi-documentation
description: MUST load when creating, updating, or reviewing any documentation in EASI — docs/, architecture canvases, READMEs, INDEX files, or pattern guides. Load when adding a new bounded context canvas, writing a pattern guide, restructuring docs, or reviewing existing documentation for bloat.
compatibility: opencode
---

# EASI Documentation

## Iron Law

**Every line earns its place. If removing it wouldn't confuse a future developer, remove it.**

Code is the source of truth. Documentation explains what code cannot: why, constraints, decisions, and cross-cutting concerns. Never document what the code already says.

## Document Types

EASI has six documentation types. Each has a purpose. Nothing else belongs in `docs/`.

| Type | Location | Purpose | Target length |
|------|----------|---------|---------------|
| Entry point | `AGENTS.md` | Tech stack, commands, conventions — one screen of context for an agent or new dev | 50–80 lines |
| Index | `docs/INDEX.md` | Task-oriented routing table — "I'm working on X, read Y" | 30–50 lines |
| Context map | `docs/architecture/README.md` | BC summary table, context map diagrams, relationship table, architectural principles | 100–160 lines |
| BC canvas | `docs/architecture/<Context>.md` | Single bounded context: purpose, language, communication, rules, architecture | 80–120 lines |
| Pattern guide | `docs/backend/*.md`, `docs/frontend/*.md` | How to do X correctly — rules, code examples, anti-patterns | 60–150 lines |
| User guide | `docs/user/*.md` | End-user-facing configuration or usage instructions | As needed |

Specs (`/specs/`) and skills (`.opencode/skills/`) have their own conventions — see `easi-spec-driven-development` and the skill template respectively. This skill does not govern those.

## READMEs Are Routing Tables

A README answers two questions:
1. What is this directory?
2. Where do I find X?

Content goes in dedicated files, not in the README. A README that exceeds 50 lines (excluding the architecture README, which holds the context map) is a sign that content should be extracted.

**Format**: one-sentence intro, then a table mapping topics to files, then optionally a quick-reference section with the 3–5 most-used commands or patterns.

**Model**: `docs/backend/README.md` — 31 lines, table + quick ref. This is the target.

## Pattern Guides

Pattern guides teach developers how to do something correctly in EASI. They are reference docs, not tutorials.

### Structure

1. **Title** — what this file covers
2. **Rules table** — each rule as a row with rationale (see `database.md`)
3. **Code examples** — WRONG/CORRECT pairs when showing anti-patterns (see `antipatterns.md`), standalone examples when showing patterns
4. **Reference tables** — anything enumerable goes in a table, not prose
5. **Checklist** — optional, for multi-step procedures (see `cross-context-events.md` "Adding a New Cross-Context Event")

### What does NOT belong in a pattern guide

- Philosophy or motivation essays ("Why we chose event sourcing")
- History ("This was introduced in v1.1")
- Comparisons to other frameworks or approaches
- Stakeholder lists or value propositions

### Length gate

If a pattern guide exceeds 150 lines, it probably covers two topics. Split it.

## Bounded Context Canvases

A canvas is a **strategic design artifact** — it records why a bounded context boundary exists, what assumptions shape it, and what forces constrain it. It is not just a developer reference; it is the document that helps architects and developers make good decisions about what belongs inside (and outside) this boundary.

### Required Sections

| Section | Content | Format |
|---------|---------|--------|
| **Purpose** | What this context does, why it exists, and the 2–3 key stakeholder roles whose language it speaks | 2–3 sentences + stakeholder bullets (roles only, not names or departments) |
| **Strategic Classification** | Domain Importance (Core/Supporting/Generic) with 1–2 sentences explaining *why* | Short paragraph. This reasoning drives architectural investment (CQRS/ES vs CRUD, team allocation). |
| **Ubiquitous Language** | Domain terms and their precise meanings | Table: Term / Meaning |
| **Inbound Communication** | Commands received (from UI), events consumed (from other BCs), queries served | Bullet lists grouped by source |
| **Outbound Communication** | Events published, commands issued to other BCs, queries made | Bullet lists grouped by target |
| **Business Rules** | Numbered, testable invariants the domain enforces | Numbered list |
| **Design Constraints** | Load-bearing assumptions that shaped architecture (scale limits, read/write ratios, update frequency) | Numbered list. Only constraints that explain a design choice. |
| **Open Questions** | Unresolved questions about the context **boundary, aggregate structure, or cross-context relationships** | Max 3–5. Feature ideas go to the backlog, not here. Remove questions answered by implementation. |
| **Boundary Health** | 2–3 measurable indicators for whether this context's boundary is healthy | Short list (e.g., cross-context event coupling ratio, orphaned reference count) |
| **Architecture Notes** | Code location, key packages, technical patterns, API style | Compact subsections |

### Sections That Do NOT Belong in a Canvas

| Cut this | Why |
|----------|-----|
| Value Proposition | Redundant with Purpose — the Purpose statement already says why the context exists |
| Business Model / Evolution Stage | Marginal value once the system is built. Domain Importance with reasoning is sufficient. |
| Domain Roles | Merge any genuinely non-obvious role (e.g., "this context does analysis, not CRUD") into the Purpose paragraph. A separate section is redundant. |
| Context Effectiveness / Business Value Metrics | Unmeasurable in practice. Belongs in product management tooling, not a technical canvas. |
| Implementation Priority / Phases | Stale the moment it's written. Tracked in specs and project management. |
| Collaboration Patterns (with code blocks) | Duplicates `cross-context-events.md` and drifts. The Inbound/Outbound sections already capture the intent of each relationship. |

### Canvas length gate

A canvas that exceeds 150 lines is carrying content that belongs elsewhere — usually in `cross-context-events.md` (event details), a spec (design decisions), or the backlog (feature-level open questions).

## Architecture README

The architecture README is the one document that justifies being longer than a routing table because it holds the context map — the system-level view that no single canvas provides.

### Required Sections

1. **Bounded Contexts table** — one row per BC: name, classification, purpose (one sentence), location, canvas link
2. **Context Map** — ASCII or Mermaid diagram showing event flows between contexts
3. **Relationship Types table** — upstream, downstream, relationship type, integration pattern
4. **Cross-Context Integration** — brief note on the published language pattern with a link to `cross-context-events.md`
5. **Context Autonomy** — architectural principles (own event store, no shared DBs, local caches)

### What does NOT belong in the architecture README

- Per-BC detail (purpose paragraphs, key responsibilities lists, published language catalogues) — that's what canvases are for
- Event subscription registries — belongs in `cross-context-events.md`
- Anti-corruption layer code examples — belongs in `cross-context-events.md`
- Domain classification lists that duplicate the summary table

## docs/INDEX.md

The index is a task-oriented routing table. Developers arrive with a task ("I'm writing an API handler") and leave with a link.

### Structure

1. **By Role/Task** — table: "Working on… / Read this"
2. **By Bounded Context** — table: context name, canvas link, classification, status
3. **Core Rules** — link to `CLAUDE.md`
4. **API Documentation** — link to Swagger

Keep it under 50 lines. If a new doc doesn't fit an existing row, add a row — don't add a paragraph.

## Writing Rules

These apply to all EASI documentation.

| Rule | Rationale |
|------|-----------|
| **Tables over prose** for anything enumerable | Scannable, diffable, forces precision |
| **Code examples over descriptions** | Show, don't tell — a WRONG/CORRECT pair teaches faster than a paragraph |
| **No future tense** | Docs describe what IS. Future plans go in specs or issues. "(future)" annotations are a smell. |
| **No comments about what was removed or changed** | Docs are not changelogs. Git history exists. |
| **No duplicate content across files** | One canonical location per fact. Link, don't copy. |
| **No template sections with no content** | If a section heading has nothing to say, delete the heading |
| **Absolute paths from repo root** for code references | `/backend/internal/capabilitymapping/`, not `the Capability Mapping module` |
| **Active voice, present tense** | "Events flow from X to Y", not "Events are flowed from X to Y" |
| **No "Note:", "Important:", "NB:" prefixes** | If it's important, it belongs in the main text. If it's not, cut it. |

## Procedure: Writing or Reviewing a Doc

### Writing a new doc

1. Identify which document type it is (see table above).
2. Check if the content already exists somewhere — if so, update that file instead.
3. Use the structure for that type. Do not invent new sections.
4. Write the content. Apply the writing rules.
5. Check the length gate. If over, split or cut.
6. Add a routing entry in the appropriate README or `INDEX.md`.

### Reviewing an existing doc

Walk the file with these questions:

1. **Does every section have content that helps a developer?** Delete empty or filler sections.
2. **Is any content duplicated from another file?** Keep the canonical location, delete the copy, add a link.
3. **Is any content speculative (assumptions, open questions, future plans)?** Move to a spec or issue, delete from the doc.
4. **Is any content stale (refers to removed code, old decisions, deprecated approaches)?** Delete. Do not mark as deprecated — just remove.
5. **Is any prose a table in disguise?** Convert.
6. **Does the file exceed the length gate for its type?** Split or cut.
7. **Is the file reachable from `INDEX.md` or a README?** If not, either add a route or question whether the file should exist.

## Hard Gates

- No doc exceeds its length gate without justification (architecture README is the sole exception, capped at 160 lines).
- No speculative content: "future", "might", "could", "should we", "open question", "TBD", "assumption" are red flags.
- No orphan docs: every file in `docs/` is reachable from `INDEX.md` or a README.
- No duplicate facts: `git grep` the key phrase before writing it. If it exists elsewhere, link.
- `INDEX.md` updated when any doc is added, renamed, or removed.
