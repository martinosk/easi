import { describe, expect, it } from 'vitest';
import { captureDirectionSchema } from './direction';

function makeBase(overrides: Partial<Record<string, unknown>> = {}) {
  return {
    type: 'consolidate' as const,
    sourceCapabilityIds: ['cap-1', 'cap-2'],
    horizon: 'next' as const,
    narrative: 'reason',
    ...overrides,
  };
}

describe('captureDirectionSchema', () => {
  it('accepts a valid consolidate direction with 2+ sources', () => {
    expect(captureDirectionSchema.safeParse(makeBase()).success).toBe(true);
  });

  it('accepts a consolidate draft with a single source (R8)', () => {
    expect(captureDirectionSchema.safeParse(makeBase({ sourceCapabilityIds: ['cap-1'] })).success).toBe(true);
  });

  it('accepts a decompose draft with a single source', () => {
    const result = captureDirectionSchema.safeParse(makeBase({ type: 'decompose', sourceCapabilityIds: ['cap-1'] }));
    expect(result.success).toBe(true);
  });

  it('accepts a stay draft with a single source', () => {
    const result = captureDirectionSchema.safeParse(makeBase({ type: 'stay', sourceCapabilityIds: ['cap-1'] }));
    expect(result.success).toBe(true);
  });

  it('rejects a draft with no sources', () => {
    const result = captureDirectionSchema.safeParse(makeBase({ sourceCapabilityIds: [] }));
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues.some((i) => i.path.includes('sourceCapabilityIds'))).toBe(true);
    }
  });

  it('rejects an unknown direction type', () => {
    expect(captureDirectionSchema.safeParse(makeBase({ type: 'merge' })).success).toBe(false);
  });

  it('rejects an unknown horizon', () => {
    expect(captureDirectionSchema.safeParse(makeBase({ horizon: 'someday' })).success).toBe(false);
  });

  it('trims narrative whitespace and leaves it as an empty string when blank', () => {
    const result = captureDirectionSchema.safeParse(makeBase({ narrative: '   ' }));
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.narrative).toBe('');
    }
  });

  it('does not carry a placements field (placements are no longer captured)', () => {
    const result = captureDirectionSchema.safeParse(makeBase());
    expect(result.success).toBe(true);
    if (result.success) {
      expect('placements' in result.data).toBe(false);
    }
  });
});
