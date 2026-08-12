import { z } from 'zod';

export const JOURNEY_KINDS = ['migration', 'consolidation', 'carve-out', 'move'] as const;
export type JourneyKindValue = (typeof JOURNEY_KINDS)[number];

const MAX_NOTE_LENGTH = 2000;

export function minSourceCount(kind: JourneyKindValue): number {
  switch (kind) {
    case 'migration':
      return 1;
    case 'consolidation':
      return 1;
    case 'carve-out':
      return 1;
    case 'move':
      return 0;
  }
}

export function maxSourceCount(kind: JourneyKindValue): number | null {
  return kind === 'carve-out' ? 1 : null;
}

export function sourceCardinalityMessage(kind: JourneyKindValue): string {
  switch (kind) {
    case 'migration':
      return 'A migration needs at least one source application.';
    case 'consolidation':
      return 'A consolidation needs at least one source application.';
    case 'carve-out':
      return 'A carve-out needs exactly one source application.';
    case 'move':
      return 'A move may list any number of source applications.';
  }
}

function addIssue(ctx: z.RefinementCtx, message: string, path: (string | number)[]): void {
  ctx.addIssue({ code: z.ZodIssueCode.custom, message, path });
}

export function violatesSourceCardinality(kind: JourneyKindValue, count: number): boolean {
  if (count < minSourceCount(kind)) return true;
  const max = maxSourceCount(kind);
  return max !== null && count > max;
}

function checkSourceCardinality(data: { kind: JourneyKindValue; fromComponentIds: string[] }, ctx: z.RefinementCtx) {
  if (violatesSourceCardinality(data.kind, data.fromComponentIds.length)) {
    addIssue(ctx, sourceCardinalityMessage(data.kind), ['fromComponentIds']);
  }
}

function checkTargetApplication(data: { toComponentId: string; fromComponentIds: string[] }, ctx: z.RefinementCtx) {
  if (!data.toComponentId.trim()) {
    addIssue(ctx, 'Select a target application', ['toComponentId']);
    return;
  }
  if (data.fromComponentIds.includes(data.toComponentId)) {
    addIssue(ctx, 'The target application cannot also be a source application', ['toComponentId']);
  }
}

function checkMoveDestination(
  data: { kind: JourneyKindValue; targetDomainId: string; resultingName: string },
  ctx: z.RefinementCtx,
) {
  if (data.kind !== 'move') return;
  if (!data.targetDomainId.trim()) addIssue(ctx, 'Select a target business domain', ['targetDomainId']);
  if (!data.resultingName.trim()) addIssue(ctx, 'Enter a resulting name for the capability', ['resultingName']);
}

function checkTargetPeriodPairing(data: { targetYear?: number; targetQuarter?: number }, ctx: z.RefinementCtx) {
  const hasYear = data.targetYear !== undefined;
  const hasQuarter = data.targetQuarter !== undefined;
  if (hasYear !== hasQuarter) {
    addIssue(ctx, 'Set both a target year and quarter, or neither', ['targetQuarter']);
  }
}

function checkNoteLength(data: { note: string }, ctx: z.RefinementCtx) {
  if (data.note.length > MAX_NOTE_LENGTH) {
    addIssue(ctx, `Note must be ${MAX_NOTE_LENGTH} characters or less`, ['note']);
  }
}

export const captureJourneySchema = z
  .object({
    kind: z.enum(JOURNEY_KINDS),
    fromComponentIds: z.array(z.string()),
    toComponentId: z.string(),
    note: z.string(),
    targetYear: z.number().optional(),
    targetQuarter: z.number().optional(),
    targetDomainId: z.string(),
    targetParentId: z.string(),
    resultingName: z.string(),
  })
  .superRefine((data, ctx) => {
    checkSourceCardinality(data, ctx);
    checkTargetApplication(data, ctx);
    checkMoveDestination(data, ctx);
    checkTargetPeriodPairing(data, ctx);
    checkNoteLength(data, ctx);
  });

export type CaptureJourneyFormData = z.infer<typeof captureJourneySchema>;
