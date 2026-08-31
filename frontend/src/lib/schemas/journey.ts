import { z } from 'zod';

export const JOURNEY_KINDS = ['migration', 'consolidation', 'carve-out', 'move', 'maturity'] as const;
export type JourneyKindValue = (typeof JOURNEY_KINDS)[number];

const MAX_NOTE_LENGTH = 2000;
export const MIN_TARGET_MATURITY = 0;
export const MAX_TARGET_MATURITY = 99;

export function carriesApplications(kind: JourneyKindValue): boolean {
  return kind !== 'maturity';
}

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
    case 'maturity':
      return 0;
  }
}

export function maxSourceCount(kind: JourneyKindValue): number | null {
  if (kind === 'carve-out') return 1;
  return kind === 'maturity' ? 0 : null;
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
    case 'maturity':
      return 'A maturity journey carries no source applications.';
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

function checkTargetApplication(
  data: { kind: JourneyKindValue; toComponentId: string; fromComponentIds: string[] },
  ctx: z.RefinementCtx,
) {
  if (!carriesApplications(data.kind)) return;
  if (!data.toComponentId.trim()) {
    addIssue(ctx, 'Select a target application', ['toComponentId']);
    return;
  }
  if (data.fromComponentIds.includes(data.toComponentId)) {
    addIssue(ctx, 'The target application cannot also be a source application', ['toComponentId']);
  }
}

function checkMaturityTarget(data: { kind: JourneyKindValue; targetMaturity?: number }, ctx: z.RefinementCtx) {
  if (data.kind !== 'maturity') return;
  if (data.targetMaturity === undefined) {
    addIssue(ctx, 'Set the maturity level this journey will reach', ['targetMaturity']);
    return;
  }
  if (data.targetMaturity < MIN_TARGET_MATURITY || data.targetMaturity > MAX_TARGET_MATURITY) {
    addIssue(ctx, `Target maturity must be between ${MIN_TARGET_MATURITY} and ${MAX_TARGET_MATURITY}`, ['targetMaturity']);
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
    targetMaturity: z.number().optional(),
  })
  .superRefine((data, ctx) => {
    checkSourceCardinality(data, ctx);
    checkTargetApplication(data, ctx);
    checkMoveDestination(data, ctx);
    checkMaturityTarget(data, ctx);
    checkTargetPeriodPairing(data, ctx);
    checkNoteLength(data, ctx);
  });

export type CaptureJourneyFormData = z.infer<typeof captureJourneySchema>;
