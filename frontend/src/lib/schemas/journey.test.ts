import { describe, expect, it } from 'vitest';
import { captureJourneySchema } from './journey';

function makeBase(overrides: Partial<Record<string, unknown>> = {}) {
  return {
    kind: 'migration' as const,
    fromComponentIds: ['comp-seabook'],
    toComponentId: 'comp-phoenix',
    note: '',
    targetYear: undefined,
    targetQuarter: undefined,
    targetDomainId: '',
    targetParentId: '',
    resultingName: '',
    ...overrides,
  };
}

describe('captureJourneySchema — kind cardinality (spec 182 rule 3)', () => {
  it('accepts a migration with one from-app', () => {
    expect(captureJourneySchema.safeParse(makeBase()).success).toBe(true);
  });

  it('rejects a migration with zero from-apps', () => {
    const result = captureJourneySchema.safeParse(makeBase({ fromComponentIds: [] }));
    expect(result.success).toBe(false);
  });

  it('accepts a consolidation with two or more from-apps', () => {
    const result = captureJourneySchema.safeParse(
      makeBase({ kind: 'consolidation', fromComponentIds: ['comp-a', 'comp-b'] }),
    );
    expect(result.success).toBe(true);
  });

  it('accepts a consolidation with one from-app (spec 194 rule 3)', () => {
    const result = captureJourneySchema.safeParse(makeBase({ kind: 'consolidation', fromComponentIds: ['comp-a'] }));
    expect(result.success).toBe(true);
  });

  it('rejects a consolidation with zero from-apps', () => {
    const result = captureJourneySchema.safeParse(makeBase({ kind: 'consolidation', fromComponentIds: [] }));
    expect(result.success).toBe(false);
  });

  it('accepts a carve-out with exactly one from-app', () => {
    const result = captureJourneySchema.safeParse(makeBase({ kind: 'carve-out', fromComponentIds: ['comp-a'] }));
    expect(result.success).toBe(true);
  });

  it('rejects a carve-out with zero from-apps', () => {
    const result = captureJourneySchema.safeParse(makeBase({ kind: 'carve-out', fromComponentIds: [] }));
    expect(result.success).toBe(false);
  });

  it('rejects a carve-out with more than one from-app', () => {
    const result = captureJourneySchema.safeParse(
      makeBase({ kind: 'carve-out', fromComponentIds: ['comp-a', 'comp-b'] }),
    );
    expect(result.success).toBe(false);
  });

  it('accepts a move with zero from-apps (implicit sources)', () => {
    const result = captureJourneySchema.safeParse(
      makeBase({ kind: 'move', fromComponentIds: [], targetDomainId: 'domain-1', resultingName: 'Freight invoicing' }),
    );
    expect(result.success).toBe(true);
  });

  it('accepts a move with several from-apps too', () => {
    const result = captureJourneySchema.safeParse(
      makeBase({
        kind: 'move',
        fromComponentIds: ['comp-a', 'comp-b'],
        targetDomainId: 'domain-1',
        resultingName: 'Freight invoicing',
      }),
    );
    expect(result.success).toBe(true);
  });
});

describe('captureJourneySchema — target application (rule 4)', () => {
  it('rejects a missing target application', () => {
    const result = captureJourneySchema.safeParse(makeBase({ toComponentId: '' }));
    expect(result.success).toBe(false);
  });

  it('rejects a target that is also among the from-apps', () => {
    const result = captureJourneySchema.safeParse(makeBase({ toComponentId: 'comp-seabook' }));
    expect(result.success).toBe(false);
  });
});

describe('captureJourneySchema — move destination (rule 5)', () => {
  it('rejects a move without a target domain', () => {
    const result = captureJourneySchema.safeParse(makeBase({ kind: 'move', targetDomainId: '', resultingName: 'X' }));
    expect(result.success).toBe(false);
  });

  it('rejects a move without a resulting name', () => {
    const result = captureJourneySchema.safeParse(
      makeBase({ kind: 'move', targetDomainId: 'domain-1', resultingName: '' }),
    );
    expect(result.success).toBe(false);
  });

  it('accepts a move with domain and resulting name but no parent', () => {
    const result = captureJourneySchema.safeParse(
      makeBase({ kind: 'move', targetDomainId: 'domain-1', resultingName: 'Freight invoicing', targetParentId: '' }),
    );
    expect(result.success).toBe(true);
  });
});

describe('captureJourneySchema — target period must be paired (rule 9)', () => {
  it('accepts neither year nor quarter set', () => {
    const result = captureJourneySchema.safeParse(makeBase({ targetYear: undefined, targetQuarter: undefined }));
    expect(result.success).toBe(true);
  });

  it('accepts both year and quarter set', () => {
    const result = captureJourneySchema.safeParse(makeBase({ targetYear: 2027, targetQuarter: 2 }));
    expect(result.success).toBe(true);
  });

  it('rejects year set without quarter', () => {
    const result = captureJourneySchema.safeParse(makeBase({ targetYear: 2027, targetQuarter: undefined }));
    expect(result.success).toBe(false);
  });

  it('rejects quarter set without year', () => {
    const result = captureJourneySchema.safeParse(makeBase({ targetYear: undefined, targetQuarter: 2 }));
    expect(result.success).toBe(false);
  });
});

describe('captureJourneySchema — note length', () => {
  it('rejects a note over 2000 characters', () => {
    const result = captureJourneySchema.safeParse(makeBase({ note: 'a'.repeat(2001) }));
    expect(result.success).toBe(false);
  });

  it('accepts a note at exactly 2000 characters', () => {
    const result = captureJourneySchema.safeParse(makeBase({ note: 'a'.repeat(2000) }));
    expect(result.success).toBe(true);
  });
});

describe('captureJourneySchema — maturity journeys (spec 211 rules 2 and 4)', () => {
  function maturityBase(overrides: Partial<Record<string, unknown>> = {}) {
    return makeBase({
      kind: 'maturity',
      fromComponentIds: [],
      toComponentId: '',
      targetMaturity: 65,
      ...overrides,
    });
  }

  it('accepts a maturity journey with a target and no applications', () => {
    expect(captureJourneySchema.safeParse(maturityBase()).success).toBe(true);
  });

  it('rejects a maturity journey without a target maturity', () => {
    expect(captureJourneySchema.safeParse(maturityBase({ targetMaturity: undefined })).success).toBe(false);
  });

  it('rejects a target maturity outside 0-99', () => {
    expect(captureJourneySchema.safeParse(maturityBase({ targetMaturity: 100 })).success).toBe(false);
    expect(captureJourneySchema.safeParse(maturityBase({ targetMaturity: -1 })).success).toBe(false);
  });

  it('rejects a maturity journey carrying source applications', () => {
    expect(captureJourneySchema.safeParse(maturityBase({ fromComponentIds: ['comp-a'] })).success).toBe(false);
  });

  it('does not require a target application for a maturity journey', () => {
    expect(captureJourneySchema.safeParse(maturityBase({ toComponentId: '' })).success).toBe(true);
  });
});
