import { describe, expect, it } from 'vitest';
import { normalizeTimeGrade } from './timeGrade';

describe('normalizeTimeGrade', () => {
  it('normalizes an all-uppercase grade to the capitalised TimeGrade', () => {
    expect(normalizeTimeGrade('INVEST')).toBe('Invest');
    expect(normalizeTimeGrade('TOLERATE')).toBe('Tolerate');
    expect(normalizeTimeGrade('MIGRATE')).toBe('Migrate');
    expect(normalizeTimeGrade('ELIMINATE')).toBe('Eliminate');
  });

  it('passes through an already-capitalised grade unchanged', () => {
    expect(normalizeTimeGrade('Invest')).toBe('Invest');
  });

  it('returns null for null, undefined, or unrecognised input', () => {
    expect(normalizeTimeGrade(null)).toBeNull();
    expect(normalizeTimeGrade(undefined)).toBeNull();
    expect(normalizeTimeGrade('not-a-grade')).toBeNull();
  });
});
