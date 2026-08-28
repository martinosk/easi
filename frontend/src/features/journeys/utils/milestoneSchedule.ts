import type { JourneyMilestone, TargetPeriod } from '../types';
import { comparePeriods } from './period';

export function milestoneScheduleConflicts(milestones: readonly JourneyMilestone[]): Map<string, TargetPeriod> {
  const conflicts = new Map<string, TargetPeriod>();
  let latestAbove: TargetPeriod | null = null;
  for (const milestone of milestones) {
    const period = milestone.targetPeriod;
    if (!period) continue;
    if (latestAbove && comparePeriods(latestAbove, period) > 0) conflicts.set(milestone.id, latestAbove);
    if (!latestAbove || comparePeriods(period, latestAbove) > 0) latestAbove = period;
  }
  return conflicts;
}
