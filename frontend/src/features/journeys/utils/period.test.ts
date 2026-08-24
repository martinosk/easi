import { describe, expect, it } from 'vitest';
import { comparePeriods, currentTargetPeriod, formatTargetPeriodCompact, periodRank } from './period';

describe('periodRank', () => {
  it('orders periods chronologically', () => {
    expect(periodRank({ year: 2026, quarter: 1 })).toBeLessThan(periodRank({ year: 2026, quarter: 2 }));
    expect(periodRank({ year: 2026, quarter: 4 })).toBeLessThan(periodRank({ year: 2027, quarter: 1 }));
  });

  it('ranks undated periods after every dated period', () => {
    expect(periodRank(null)).toBe(Number.POSITIVE_INFINITY);
  });
});

describe('comparePeriods', () => {
  it('is negative when the first period is earlier', () => {
    expect(comparePeriods({ year: 2026, quarter: 1 }, { year: 2026, quarter: 3 })).toBeLessThan(0);
  });

  it('is zero for the same quarter', () => {
    expect(comparePeriods({ year: 2026, quarter: 3 }, { year: 2026, quarter: 3 })).toBe(0);
  });

  it('is positive when the first period is later', () => {
    expect(comparePeriods({ year: 2027, quarter: 1 }, { year: 2026, quarter: 4 })).toBeGreaterThan(0);
  });
});

describe('currentTargetPeriod', () => {
  it.each([
    [0, 1],
    [2, 1],
    [3, 2],
    [5, 2],
    [6, 3],
    [8, 3],
    [9, 4],
    [11, 4],
  ])('maps month index %i to quarter %i', (monthIndex, quarter) => {
    expect(currentTargetPeriod(new Date(2026, monthIndex, 15))).toEqual({ year: 2026, quarter });
  });
});

describe('formatTargetPeriodCompact', () => {
  it("formats a period as Q<q>'<yy>", () => {
    expect(formatTargetPeriodCompact({ year: 2026, quarter: 3 })).toBe("Q3'26");
  });

  it('keeps the century out of the label', () => {
    expect(formatTargetPeriodCompact({ year: 2108, quarter: 1 })).toBe("Q1'08");
  });
});
