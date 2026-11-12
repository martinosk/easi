# Design: Guided Data Quality

> Status: proposed 2026-11-12 — Phase-1 design document; slices in specs 235–241
> Author: agent + maosk
> Date: 2026-11-12
> Roadmap: extends move H2-6 (stewardship read side, decision SD8) and uses SD6 (application record). Builds on [`personal-home.md`](personal-home.md) and specs 226–231.

## Problem Statement

A review of a production tenant on 2026-11-12 (305 applications, 1,219 capabilities, 31 views) found that most data-quality gaps are created at the moment of entry and then stay invisible to the person who created them:

| Finding | Count | Why it happens |
|---------|-------|----------------|
| Applications with hosting unknown | 185 of 305 | Creation asks for a name and a description only; hosting is set later, somewhere else, or never |
| Applications with no owner | 293 of 305 | Same |
| Applications realising no capability | 49 | Same; nothing reports an application that is linked to nothing |
| L1 capabilities in no business domain | 20 | Creating an L1 does not ask for a domain |
| Duplicate applications and capabilities | 2 application pairs merged by hand, 3 sibling capability pairs, 3 same-name L1 pairs, 3 groups of near-duplicates | Creation does not show what already exists; nothing detects duplicates afterwards; there is no merge |
| Applications with "(deprecated)" in the name, placeholders, test entries | about 15 | No guidance at entry |

Home (spec 228) and attention items (spec 230) are designed to surface problems, but they do not cover these: hosting, unlinked applications, unassigned L1s and duplicates have no check; a user who creates 100 applications and owns none of them has no personal scope, so nothing is shown to them as theirs; and a finding can only be fixed by leaving the list, opening the subject and finding the right control, one subject at a time.

The goal of this family is that gaps **do not appear in the first place**, and that those which do are **shown to the person who can close them, with the fix one step away**.

## Research Summary

- **Creation is two fields.** The application dialog asks for name and description; the capability dialog for name, description and status. Hosting, ownership, realisations and domain assignment are separate commands reached from the details panel or the domain board.
- **No uniqueness or similarity rule exists** for applications, capabilities or origin entities. Only business domain names are unique. Capability names legitimately repeat under different parents.
- **The creator is known** from the actor of each creation event; Audit serves it today to the navigation tree's "created by" filter. Stewardship does not cache it and it is not an anchor.
- **The stewardship application cache holds** id, name and ownership. It does not hold hosting. The capability cache holds parent, level, status and EA owner; domain assignments of L1s are cached.
- **Spec 230 (pending) fixes twelve checks in code** and states that items are never dismissed. Any new check must therefore be one whose only resolution is fixing the data.
- **Every single-field fix already has a command and a HATEOAS affordance** on the subject (`x-classify-hosting`, `x-nominate-owner`, `x-assign-owner`, domain assignment, EA owner). No bulk or inline surface uses them.
- **There is no merge.** Merging two applications by hand in the reviewed tenant meant re-linking realisations and then deleting, and lost the absorbed application's view placements.
- **Spec 218 (pending) sets the precedent for guidance at entry:** advice at the edge, never a domain block, override recorded.

## Proposed Approach

Three layers, each a set of independent slices.

### Prevent: guidance at the moment of entry

1. **Similar-entity suggestions (spec 237).** While a name is typed in a create dialog, the owning context returns existing entities with a similar name. The user opens the existing one or continues. Advisory, like spec 218.
2. **Guided application creation (spec 238).** Creating an application continues into a short "complete the record" step for hosting, owner and the first realised capability. Each step is an existing command, each is skippable, and what is skipped is said plainly.
3. **Guided capability creation (spec 239).** Creating an L1 asks for its business domain and EA owner in the same flow.

### Detect: findings shown to the person who caused or can close them

4. **Creator as an anchor (spec 235).** What a user created, and nobody yet answers for, is in their scope and in My Work until an owner exists. This is the "applications I created that are orphaned" counter.
5. **Model hygiene checks (spec 236).** Five new checks in the spec 230 engine: hosting unclassified, application realises nothing, L1 outside every domain, duplicate application name, duplicate capability name among siblings.

### Correct: the fix is where the finding is

6. **Fix from the attention list (spec 240).** Findings whose fix is one field are fixed inline and in bulk from the list, using the subject's own affordances.
7. **Merge applications (spec 241).** One command that moves everything the absorbed application had to the surviving one, across contexts, by a published event.

## Key Decisions

| ID | Decision | Rationale / alternative rejected |
|----|----------|----------------------------------|
| Q1 | **Guidance at entry is advisory; the domain gains no new uniqueness invariant** | Same-name capabilities under different parents are legitimate, existing tenants hold duplicates that an invariant would make uneditable, and imports and agents would need an override path. Alternative: reject exact duplicate names with 409 (rejected for those reasons; exact duplicates are caught by a check instead). |
| Q2 | **Similarity is served by the context that owns the entity** | It is a query over that context's own read model. Alternative: a search context (rejected: a second copy of every name for one query). |
| Q3 | **Guided creation creates first, then completes** | Model state is immediate in EASI; the record exists as soon as it is named, and each completion step is an ordinary command. Alternative: one composite "create complete application" command (rejected: a second write path that drifts from the single-field commands, and a cross-context transaction for the realisation). |
| Q4 | **The creator is an anchor only until someone answers for the subject** | Creating is not accountability, but an unowned record has nobody else. The anchor ends when an owner exists, which rewards assigning one. Alternative: creator as a permanent anchor (rejected: an importer would carry hundreds of cards forever). |
| Q5 | **New checks join spec 230's fixed list; none can be dismissed** | Keeps one engine and one rule. It limits duplicate detection to exact normalised names, whose resolution is always rename or merge. Near-duplicates are handled at entry only. Alternative: fuzzy duplicate findings with a "not a duplicate" mark (rejected: it is dismissal under another name and needs a new stored fact). |
| Q6 | **Inline and bulk fixes use the subject's affordances** | The UI stays driven by HATEOAS; stewardship never decides who may change an application. Alternative: stewardship computes fix links (rejected: duplicates the owning context's permission rules). |
| Q7 | **Merge is an Architecture Modeling command with a published `ApplicationMerged` event** | Realisations, assessments, facts, grants and placements live in six contexts; each moves its own data in reaction, inside the publishing transaction. Alternative: a client-side sequence of re-link calls followed by delete (rejected: what was done by hand, not atomic, and loses whatever the client forgets). |
| Q8 | **Capabilities get no merge in this family** | A capability carries a subtree, domain assignments, dependencies, journeys and importance ratings; duplicate capabilities found so far were resolved by deleting an empty copy. The duplicate-name check makes them visible. |

## Slice Map

| Slice | Spec | Content | Depends on |
|-------|------|---------|------------|
| A | 235 | Creator anchor: creator cached, anchors unowned created subjects, "Created by you" on My Work | 228 |
| B | 236 | Model hygiene checks: five checks, hosting in the application cache | 230, 237 |
| C | 237 | Similar-entity suggestions at creation for applications, capabilities and origin entities | — |
| D | 238 | Guided application creation: hosting, owner, first realisation | 237 |
| E | 239 | Guided capability creation: domain and EA owner for a new L1 | 237 |
| F | 240 | Fix from the attention list: inline and bulk single-field fixes | 230, 236, 241 |
| G | 241 | Merge applications | — |

C, D, E and G do not depend on the pending Home slices and can ship first.
