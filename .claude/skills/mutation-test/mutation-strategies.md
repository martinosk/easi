# Mutation Strategies Reference

Strategies for generating semantically meaningful mutations. Each category includes examples showing weak tests that would miss the mutation vs. strong tests that catch it.

## 1. Conditional Boundary Mutations

| Original | Mutation | Bug Simulated |
|----------|----------|---------------|
| `x > 0` | `x >= 0` | Off-by-one at zero boundary |
| `x <= limit` | `x < limit` | Excludes boundary value |
| `err != nil` | `err == nil` | Inverted error handling |

```
// age >= 18
// Mutant: age > 18

WEAK:  isAdult(25) → true both ways (25 >= 18 and 25 > 18)
STRONG: isAdult(18) → true vs false (18 >= 18 but NOT 18 > 18)
```

## 2. Logical Operator Mutations

| Original | Mutation | Bug Simulated |
|----------|----------|---------------|
| `a && b` | `a \|\| b` | Wrong logical operator |
| `a \|\| b` | `a && b` | Wrong logical operator |
| `!condition` | `condition` | Removed negation |

```
// isAdmin || isOwner
// Mutant: isAdmin && isOwner

WEAK:  canAccess(true, true) → true both ways
STRONG: canAccess(true, false) → true vs false
```

## 3. Arithmetic Mutations

| Original | Mutation | Bug Simulated |
|----------|----------|---------------|
| `a + b` | `a - b` | Wrong operator |
| `a * b` | `a / b` | Inverted calculation |
| `count + 1` | `count` | Missing increment |
| `index - 1` | `index` | Off-by-one |

```
// price * quantity
// Mutant: price / quantity

WEAK:  calculate(10, 1) → 10 both ways (10*1 = 10/1)
STRONG: calculate(10, 3) → 30 vs 3.33
```

## 4. Return Value Mutations

| Original | Mutation | Bug Simulated |
|----------|----------|---------------|
| `return result` | `return nil` / `return ""` | Missing return value |
| `return result, nil` | `return result, errors.New(...)` | Fabricated error |
| `return true` | `return false` | Wrong boolean |
| `return items` | `return nil` | Empty result |

## 5. Guard / Early Return Removal

Remove entire guard clauses.

| Original | Mutation | Bug Simulated |
|----------|----------|---------------|
| `if err != nil { return err }` | *(remove block)* | Silently swallowed error |
| `if !authorized { return 403 }` | *(remove block)* | Missing authorization |
| `if input == "" { return default }` | *(remove block)* | Missing validation |

```
// processOrder: validateOrder(order); saveOrder(order); sendConfirmation(order);
// Mutant: empty function body

WEAK:  expect(() => processOrder(order)).not.toThrow()  // empty fn also doesn't throw
STRONG: processOrder(order); expect(db.save).toHaveBeenCalledWith(order)
```

## 6. Collection / Loop Mutations

| Original | Mutation | Bug Simulated |
|----------|----------|---------------|
| `for i := 0; ...` | `for i := 1; ...` | Skips first element |
| `append(slice, item)` | `slice` | Lost data |
| `items[len(items)-1]` | `items[0]` | Wrong element |
| `sort.Slice(...)` | *(remove)* | Missing ordering |

## 7. Method Expression Mutations (TypeScript)

| Original | Mutation | Bug Simulated |
|----------|----------|---------------|
| `startsWith()` | `endsWith()` | Wrong string position |
| `toUpperCase()` | `toLowerCase()` | Wrong case |
| `some()` | `every()` | Partial vs full match |
| `filter()` | *(remove)* | Missing filtering |
| `sort()` | *(remove)* | Missing ordering |
| `Math.min()` | `Math.max()` | Wrong extremum |
| `trim()` | `trimStart()` | Incomplete trim |

## 8. Optional Chaining / Nullish Mutations

| Original (TypeScript) | Mutation | Bug Simulated |
|------------------------|----------|---------------|
| `foo?.bar` | `foo.bar` | Missing null guard |
| `foo?.[i]` | `foo[i]` | Missing null guard |
| `x ?? fallback` | `x` | Missing fallback |

| Original (Go) | Mutation | Bug Simulated |
|----------------|----------|---------------|
| `if x != nil` | `if x == nil` | Nil dereference |
| `val, ok := m[key]; if ok` | *(ignore ok)* | Missing key handling |

## 9. Unary Operator Mutations

| Original | Mutation | Bug Simulated |
|----------|----------|---------------|
| `+a` | `-a` | Wrong sign |
| `-a` | `+a` | Wrong sign |
| `++a` / `a++` | `--a` / `a--` | Increment vs decrement |

## 10. String / Constant Mutations

| Original | Mutation | Bug Simulated |
|----------|----------|---------------|
| `"text"` | `""` | Empty string |
| `""` | `"mutated"` | Non-empty where empty expected |
| `http.StatusOK` | `http.StatusNotFound` | Wrong status code |

## 11. State Transition & Event Sourcing Mutations

| Original | Mutation | Bug Simulated |
|----------|----------|---------------|
| `status = Active` | `status = Inactive` | Wrong state |
| `aggregate.Apply(event)` | *(remove)* | Event not applied |
| `events = append(events, e)` | *(remove)* | Lost domain event |
| Field assignment in `When()` | *(remove)* | State not updated from event |
| Version increment | *(remove)* | Broken concurrency check |

---

## Equivalent Mutants

Equivalent mutants produce identical behavior to the original — they cannot be killed and should be skipped.

**Common patterns:**
- Operations with identity elements: `+= 0`, `-= 0`, `*= 1`, `/= 1`
- Boundary conditions where both sides have the same outcome
- Mutations in dead code paths that are never reached
- Duplicate conditions where another check already covers the case

When you identify an equivalent mutant, skip it and note it in the report.

## Identity Values to Avoid in Tests

Tests using these values cannot distinguish between mutated operators:

| Avoid | Why | Use Instead |
|-------|-----|-------------|
| `0` for `+` / `-` | `x + 0 = x - 0` | Non-zero: `3`, `7` |
| `1` for `*` / `/` | `x * 1 = x / 1` | Values > 1: `3`, `5` |
| `""` for string ops | Empty string edge cases | `"abc"`, `"test"` |
| `true, true` for `&&`/`\|\|` | `T && T = T \|\| T` | `true, false` |
| `false, false` for `&&`/`\|\|` | `F && F = F \|\| F` | `true, false` |
| `[]` for array ops | Empty array edge cases | `[1, 2, 3]` |
| All-matching arrays for `some`/`every` | `some(allTrue) = every(allTrue)` | Partially matching arrays |
