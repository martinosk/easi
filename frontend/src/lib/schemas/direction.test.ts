import { describe, expect, it } from 'vitest';
import { captureDirectionSchema } from './direction';

function makeBase(overrides: Partial<Record<string, unknown>> = {}) {
  return {
    type: 'consolidate' as const,
    sourceCapabilityIds: ['cap-1', 'cap-2'],
    placements: [{ targetBusinessDomainId: 'dom-1', resultingName: 'X' }],
    horizon: 'next' as const,
    narrative: 'reason',
    ...overrides,
  };
}

describe('captureDirectionSchema', () => {
  it('accepts a valid consolidate direction with 2+ sources and 1 placement', () => {
    expect(captureDirectionSchema.safeParse(makeBase()).success).toBe(true);
  });

  it('rejects consolidate with fewer than 2 sources', () => {
    const result = captureDirectionSchema.safeParse(makeBase({ sourceCapabilityIds: ['cap-1'] }));
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues.some((i) => i.path.includes('sourceCapabilityIds'))).toBe(true);
    }
  });

  it('rejects decompose without exactly one source', () => {
    const result = captureDirectionSchema.safeParse(
      makeBase({ type: 'decompose', sourceCapabilityIds: ['a', 'b'] }),
    );
    expect(result.success).toBe(false);
  });

  it('rejects stay with placements', () => {
    const result = captureDirectionSchema.safeParse(
      makeBase({ type: 'stay', sourceCapabilityIds: ['cap-1'], placements: [{ targetBusinessDomainId: 'd' }] }),
    );
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues.some((i) => i.path.includes('placements'))).toBe(true);
    }
  });

  it('accepts stay with exactly one source and no placements', () => {
    const result = captureDirectionSchema.safeParse(
      makeBase({ type: 'stay', sourceCapabilityIds: ['cap-1'], placements: [] }),
    );
    expect(result.success).toBe(true);
  });

  it('rejects consolidate with 0 or 2+ placements', () => {
    const a = captureDirectionSchema.safeParse(makeBase({ placements: [] }));
    const b = captureDirectionSchema.safeParse(
      makeBase({
        placements: [
          { targetBusinessDomainId: 'd1', resultingName: 'A' },
          { targetBusinessDomainId: 'd2', resultingName: 'B' },
        ],
      }),
    );
    expect(a.success).toBe(false);
    expect(b.success).toBe(false);
  });

  it('rejects decompose with no placements', () => {
    const result = captureDirectionSchema.safeParse(
      makeBase({ type: 'decompose', sourceCapabilityIds: ['cap-1'], placements: [] }),
    );
    expect(result.success).toBe(false);
  });

  it('rejects when any placement has an empty target business domain', () => {
    const result = captureDirectionSchema.safeParse(
      makeBase({ placements: [{ targetBusinessDomainId: '', resultingName: 'X' }] }),
    );
    expect(result.success).toBe(false);
  });

  it('trims narrative whitespace and leaves it as an empty string when blank', () => {
    const result = captureDirectionSchema.safeParse(makeBase({ narrative: '   ' }));
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.narrative).toBe('');
    }
  });
});
