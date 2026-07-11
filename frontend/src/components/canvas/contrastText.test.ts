import { describe, expect, it } from 'vitest';
import { getContrastTextColor } from './contrastText';

describe('getContrastTextColor', () => {
  it('returns ink text for a light background', () => {
    expect(getContrastTextColor('#ffffff')).toBe('var(--ink)');
  });

  it('returns white text for a dark background', () => {
    expect(getContrastTextColor('#000000')).toBe('#ffffff');
  });

  it('returns ink text for a light pastel fill', () => {
    expect(getContrastTextColor('#E2E7EB')).toBe('var(--ink)');
  });

  it('returns white text for a saturated dark fill', () => {
    expect(getContrastTextColor('#5F4FC7')).toBe('#ffffff');
  });

  it('returns ink text for a mid-tone fill where ink has the higher contrast ratio', () => {
    expect(getContrastTextColor('#E88C7D')).toBe('var(--ink)');
  });

  it('falls back to ink text for an unparsable value', () => {
    expect(getContrastTextColor('not-a-color')).toBe('var(--ink)');
  });

  it('handles hex values without a leading hash', () => {
    expect(getContrastTextColor('000000')).toBe('#ffffff');
  });
});
