import type { CapabilityJourney, TargetPeriod } from '../types';

export function periodRank(period: TargetPeriod | null): number {
  return period ? period.year * 4 + period.quarter : Number.POSITIVE_INFINITY;
}

export function comparePeriods(first: TargetPeriod, second: TargetPeriod): number {
  return periodRank(first) - periodRank(second);
}

export function currentTargetPeriod(now: Date): TargetPeriod {
  return { year: now.getFullYear(), quarter: Math.floor(now.getMonth() / 3) + 1 };
}

export function formatTargetPeriodCompact(period: TargetPeriod): string {
  return `Q${period.quarter}'${String(period.year % 100).padStart(2, '0')}`;
}

export function byTargetPeriodThenName(first: CapabilityJourney, second: CapabilityJourney): number {
  const firstRank = periodRank(first.targetPeriod);
  const secondRank = periodRank(second.targetPeriod);
  if (firstRank !== secondRank) return firstRank - secondRank;
  return first.capabilityName.localeCompare(second.capabilityName);
}
