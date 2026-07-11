import { afterEach, describe, expect, it, vi } from 'vitest';
import { formatIsoDate } from './date';

describe('formatIsoDate', () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it('renders the same local calendar date regardless of a negative UTC offset', () => {
    vi.stubEnv('TZ', 'America/New_York');

    expect(formatIsoDate('2023-05-01')).toBe(new Date(2023, 4, 1).toLocaleDateString());
  });

  it('returns the raw string for invalid input', () => {
    expect(formatIsoDate('not-a-date')).toBe('not-a-date');
  });
});
