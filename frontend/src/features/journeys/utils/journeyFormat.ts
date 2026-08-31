import type { CapabilityJourney, JourneyKind, JourneyMaturity, JourneyMilestone, TargetPeriod } from '../types';

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
  if (kind === 'move') return 'capability move';
  return kind === 'maturity' ? 'maturity uplift' : kind;
}

export function journeyDestinationLabel(journey: CapabilityJourney): string {
  if (journey.maturity) return `maturity ${journey.maturity.targetMaturity}`;
  return journey.toApplication.componentName;
}

export function maturityGapLabel(maturity: JourneyMaturity): string {
  if (maturity.maturityGap <= 0) return 'reached';
  return `${maturity.maturityGap} to go`;
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
