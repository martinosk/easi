import type { CapabilityJourney, JourneyKind, JourneyMilestone, TargetPeriod } from '../types';

export function journeyStatusLabel(journey: CapabilityJourney): string {
  switch (journey.status) {
    case 'planned':
      return 'planned';
    case 'in-flight':
      return journey.progress !== null ? `in flight (${journey.progress}%)` : 'in flight';
    case 'done':
      return 'done (100%)';
    case 'abandoned':
      return 'abandoned';
  }
}

export function journeyKindLabel(kind: JourneyKind): string {
  return kind === 'move' ? 'capability move' : kind;
}

export function formatTargetPeriod(period: TargetPeriod | null): string {
  return period ? `Q${period.quarter} ${period.year}` : '—';
}

export function milestoneWhenLabel(milestone: JourneyMilestone): string {
  const period = milestone.targetPeriod ? `Q${milestone.targetPeriod.quarter} ${milestone.targetPeriod.year}` : '';
  if (milestone.status === 'done') return period ? `Done · ${period}` : 'Done';
  return period;
}

export function scheduleConflictLabel(milestone: JourneyMilestone, latestAbove: TargetPeriod): string {
  return `Targeted for ${formatTargetPeriod(milestone.targetPeriod)} but listed after a milestone targeted for ${formatTargetPeriod(latestAbove)}`;
}
