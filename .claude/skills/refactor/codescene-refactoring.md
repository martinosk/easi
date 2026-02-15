# CodeScene Refactoring Reference

How to use CodeScene MCP tools for code health improvements, plus proven patterns from this codebase.

## Tool Usage

### Analysis Flow

1. **Score**: `mcp__codescene__code_health_score` — get the current score (10.0 = optimal, <4.0 = red)
2. **Review**: `mcp__codescene__code_health_review` — get detailed findings with specific code smells
3. **Auto-refactor**: `mcp__codescene__code_health_auto_refactor` — automated fix for a single function

### Auto-Refactor Constraints

`code_health_auto_refactor` supports:
- **Languages**: JavaScript/TypeScript, Java, C#, C++
- **Smells**: Complex Conditional, Bumpy Road Ahead, Complex Method, Deep Nested Complexity, Large Method
- **Limit**: Functions under 300 lines only
- Insert new extracted functions close to the refactored function

For Go files or unsupported smells, apply fixes manually using the patterns below.

### Iterative Loop

```
check score → review → fix → re-check → repeat until 10.0
```

**Goal: Always reach 10.0 Code Health.** Iterate as many times as needed. Only stop iterating if:
- Score is 10.0 (success)
- Remaining findings are inherent to domain/architectural patterns (see "Domain Model Limits" below)

## What Works — Proven Patterns

### Go Backend

| Pattern | When to Use | Example |
|---------|-------------|---------|
| **Extract named helpers** | Repeated logic in handlers/projectors | `getContainerOrFail`, `respondCreatedOrOK` |
| **Parameter objects** | Function has 5+ args that form a cohesive concept | `TagInfo{}`, `CapabilityUpdate{}` |
| **Method on struct** | Standalone function takes struct + many args | `testCtx.createUser()` vs `createTestUser(t, db, ...)` |
| **Handler maps** | Large switch on string/type | `map[string]func` instead of switch |
| **Local variable extraction** | Complex multi-condition if-statements | Named booleans: `isActive := ...` |
| **Table-driven tests** | Duplicate test cases with minor variations | `tests := []struct{...}` |

### React/TypeScript Frontend

| Pattern | When to Use | Example |
|---------|-------------|---------|
| **Extract custom hooks** | Component has complex data-fetching + state logic | `useFilteredTreeData` from large component |
| **Extract sub-components** | Render function exceeds ~80 lines | `OriginEntitySections` from parent |

Both React patterns reliably bring scores from ~9.3 to 10.0.

## What Doesn't Work — Avoid These

| Anti-Pattern | Why It Fails |
|--------------|-------------|
| **Splitting switch-case into sub-methods** when cases are structurally similar | Introduces MORE Code Duplication (score dropped 8.81 → 8.54) |
| **Helper functions with 5+ args** | Creates new Excess Function Arguments smell, offsetting gains |
| **Forced abstractions between different types** (e.g., Create vs Update handlers) | Code isn't actually duplicated, just structurally similar |
| **Over-abstracting domain events** | Constructors naturally have many args; parameter structs add indirection |
| **Bundling args into a struct just to reduce count** | Doesn't fix the root cause; creates a bag of unrelated fields |

## Domain Model Limits — Don't Fight These

These patterns score below 10.0 but are inherent to the architecture and should NOT be refactored:

- **Event-sourced aggregate `apply()` with type switch** (cc=15 for 14 event types): Splitting worsens duplication. Leave as-is.
- **Domain event constructors with 5-6 args**: Explicit event types require explicit fields. Parameter structs just add indirection.
- **Projector methods handling different DTO types**: Can't generalize without Go generics complexity that isn't worth it.

When these are the only remaining findings, classify as **Skip**, document the reason, and stop iterating.

**For everything else: iterate until 10.0.** No exceptions.

## Fixing "Excess Number of Function Arguments"

Do NOT fix by bundling arguments into a struct. Instead investigate the root cause:

1. **Low cohesion / too many responsibilities**: The function does too much. Split responsibilities into separate types.
2. **Missing domain abstraction**: A coherent concept hides behind the arguments. Only introduce a type if it genuinely encapsulates something meaningful — a domain value object, a result type, a configuration.

## Score Thresholds

| Language | Cyclomatic Complexity | Max Args | Max Function Length |
|----------|----------------------|----------|---------------------|
| Go | 9 | 4 | 80 lines |
| TypeScript | 9 | 4 | — |

- Code Duplication indication=2: moderate
- Code Duplication indication=3: significant — worth addressing
