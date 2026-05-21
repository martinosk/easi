---
name: easi-frontend-styling
description: MUST load when writing or reviewing any UI component, dialog, form, button, layout, or styling in the EASI frontend. Load when picking a UI primitive, deciding how to lay out a component, when you would otherwise reach for `style={{}}` or a bare HTML element, or when adding any visual chrome. The single UI vocabulary in EASI is Mantine v8 — this skill enforces it.
compatibility: opencode
---

# EASI Frontend Styling

## Overview

EASI uses **one UI vocabulary: Mantine v8** (`@mantine/core`). Never write plain HTML for dialogs, buttons, inputs, or form controls. Never reach for a `.btn`, `.dialog`, `.form-input` class. Never use `style={{}}` for static layout — Mantine layout primitives are how layout is expressed.

This skill is exclusively about UI primitives, dialogs, layout, and styling. For HATEOAS link gating, TanStack Query, cache invalidation, and mutation hooks, load `easi-frontend-data` instead.

Spec 168 captures the underlying refactor (three competing vocabularies collapsed to Mantine). The rules below are the post-168 steady state.

## UI Component Framework (Mantine v8)

### Component Mapping

| UI Surface | Mantine Component |
|---|---|
| Dialogs / modals | `Modal` |
| Confirmation prompts | `components/shared/ConfirmationDialog` (a thin Mantine wrapper) |
| Text inputs | `TextInput` |
| Multi-line text inputs | `Textarea` |
| Number inputs | `NumberInput` |
| Single-select dropdowns | `Select` |
| Multi-select dropdowns | `MultiSelect` |
| Checkboxes | `Checkbox` |
| Checkbox groups | `Checkbox.Group` |
| Radio groups | `Radio.Group` |
| Sliders | `Slider` |
| Buttons (primary / default) | `Button` (with `variant` prop — `filled` / `default` / `subtle` / `outline`) |
| Icon-only buttons | `ActionIcon` |
| Vertical spacing | `Stack` (with `gap`) |
| Horizontal grouping | `Group` (with `gap`, `justify`, `align`) |
| Body text | `Text` (with `size`, `c`, `fw`) |
| Headings | `Title` (with `order`) |
| Status badges | `Badge` |
| Alert / inline error | `Alert` |
| Loading spinner | `Loader` |
| Container with background/border | `Paper` or `Card` |

### Forms — react-hook-form + zod

All non-trivial forms use **`react-hook-form`** with **`@hookform/resolvers/zod`** and a zod schema. The schema lives in `src/lib/schemas/{domain}.ts` and is exported via `src/lib/schemas/index.ts`.

```typescript
import { zodResolver } from '@hookform/resolvers/zod';
import { Controller, useForm } from 'react-hook-form';

const { register, control, handleSubmit, formState: { errors, isValid } } = useForm({
  resolver: zodResolver(mySchema),
  defaultValues: DEFAULT_VALUES,
  mode: 'onChange',
});
```

Cross-field validation uses `z.object({...}).superRefine((data, ctx) => ctx.addIssue({ ... }))`. Reference: `src/lib/schemas/direction.ts`.

Mantine inputs that aren't direct `<input>`-like (Select, Checkbox.Group, Slider, etc.) need `<Controller>`. Plain `<TextInput>` and `<Textarea>` work with `{...register('field')}`.

### Test wrapper

Tests that render Mantine components must render through `renderWithProviders` from `src/test/helpers/`. It wraps in `MantineProvider`, `QueryClientProvider`, and (optionally) `MemoryRouter`.

```typescript
import { renderWithProviders } from '../../test/helpers';

renderWithProviders(<MyComponent />, { withRouter: false });
```

The older `MantineTestWrapper` import path still exists but is deprecated — prefer `renderWithProviders` for new tests.

## Layout — no inline `style={{}}` for static layout

**Never use `style={{}}` for static layout.** This is a hard rule from spec 168 invariant #2. Use Mantine layout primitives and system props instead.

### What to do instead

| Want to express | Don't do | Do this |
|---|---|---|
| `flex: 1` on a form control | `<Select style={{ flex: 1 }} />` | `<Box flex={1}><Select /></Box>` |
| `alignSelf: 'flex-start'` on one child | `<Button style={{ alignSelf: 'flex-start' }}>` | Wrap in `<Group justify="flex-start"><Button /></Group>` |
| `marginTop` / `paddingLeft` etc. | `style={{ marginTop: 16 }}` | Mantine spacing props: `mt="md"`, `pl="sm"`, etc. |
| Text colour / size / weight | `style={{ color: '#6b7280', fontSize: 12 }}` | `<Text c="dimmed" size="xs">` |
| Background / border / shadow | `style={{ background: '#fff', boxShadow: '...' }}` | `<Paper>` or `<Card>` with Mantine props |
| Stack of items | `<div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>` | `<Stack gap="sm">` |
| Horizontal row | `<div style={{ display: 'flex', gap: 8 }}>` | `<Group gap="sm">` |
| Grid of cards | `<div style={{ display: 'grid', gridTemplateColumns: '...' }}>` | `<SimpleGrid cols={...}>` |

### Legitimate exceptions

Inline `style={{}}` is allowed only when:

1. **ReactFlow node positioning** — `src/components/canvas/**` and `src/features/canvas/**` use `style` for x/y positioning that ReactFlow demands.
2. **Genuinely runtime-dynamic values** — a colour computed from data, a pixel offset computed from a DOM measurement. If the value is static, it's not an exception.

Static positioning values (`flex: 1`, `alignSelf: 'flex-start'`, fixed `width: 300px`, hard-coded colours) are **not exceptions** — express them in Mantine.

### No hard-coded design tokens in `.tsx`

Hex colours, raw pixel values, and `rem` values do not belong in `.tsx`. The single source of design tokens is `src/index.css` `:root` (consumed by `src/theme/mantine.ts` via CSS variables). Use Mantine tokens via component props (`c="dimmed"`, `fz="sm"`, `radius="md"`, `shadow="sm"`) or `var(--…)` in a `.module.css` file.

### No bare interactive HTML

Do not write `<button>`, `<input>`, `<select>`, `<textarea>`, `<form>`, `<dialog>`, `<fieldset>`, `<legend>` in source `.tsx` files (outside `src/components/canvas/**` and tests). Use the Mantine component that maps to that role. `<form>` is acceptable as the outer wrapper around a react-hook-form `handleSubmit`, but every input inside it must be a Mantine component.

### No `className="btn*"`, `className="dialog*"`, etc.

These class systems are being retired. Do not introduce new usages. Do not invent CSS class names for UI elements — the EASI stylesheet does not back them.

## Reference Implementations

| Pattern | Reference |
|---|---|
| Create dialog with form | `src/features/capabilities/components/CreateCapabilityDialog.tsx` |
| Confirmation dialog | `src/components/shared/ConfirmationDialog.tsx` |
| Form with cross-field validation | `src/features/architecture-direction/components/CaptureDirectionForm.tsx` (uses `src/lib/schemas/direction.ts` with `superRefine`) |
| Detail panel with Mantine | `src/features/architecture-direction/components/DirectionPanel.tsx` |
| Status badge | `src/features/architecture-direction/components/DirectionStatusBadge.tsx` |

## Guidelines

1. **Use Mantine for all UI components.** No plain HTML for dialogs, inputs, buttons, or checkboxes.
2. **Never invent CSS class names** for UI elements — they will not exist in the EASI stylesheet.
3. **No `style={{}}` for static layout.** Use `<Box flex={1}>`, `<Group justify="flex-start">`, Mantine spacing props (`mt`, `gap`, `p`).
4. **No hard-coded hex/rem/px in `.tsx`.** Mantine tokens (`c="dimmed"`, `radius="md"`) or `var(--…)` in `.module.css`.
5. **Forms use `react-hook-form` + `zod`.** Schema in `src/lib/schemas/{domain}.ts`. Cross-field rules via `.superRefine()`.
6. **Render tests through `renderWithProviders`** from `src/test/helpers/`.
7. **Gate ReactFlow node-wrapper HOCs on the feature flag** — do not wrap unconditionally; ReactFlow expects a stable root element.
