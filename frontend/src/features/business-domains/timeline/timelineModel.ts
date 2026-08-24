import type { CapabilityJourney, JourneyMilestone, TargetPeriod } from '../../journeys/types';
import { byTargetPeriodThenName, comparePeriods, periodRank } from '../../journeys/utils/period';

export interface TimelineMilestoneRow {
  milestone: JourneyMilestone;
  overdue: boolean;
  column: number | null;
}

export interface TimelineJourneyRow {
  journey: CapabilityJourney;
  overdue: boolean;
  targetColumn: number | null;
  milestones: TimelineMilestoneRow[];
}

export interface TimelineDomainGroup {
  domainName: string;
  journeys: TimelineJourneyRow[];
}

export interface TimelineSummary {
  planned: number;
  inFlight: number;
  overdueJourneys: number;
  overdueMilestones: number;
}

export interface TimelineModel {
  quarters: TargetPeriod[];
  currentColumn: number;
  groups: TimelineDomainGroup[];
  summary: TimelineSummary;
}

export interface BuildTimelineParams {
  journeys: CapabilityJourney[];
  domainNameFor: (capabilityId: string) => string | undefined;
  current: TargetPeriod;
}

function isPastDue(period: TargetPeriod | null, status: string, current: TargetPeriod): boolean {
  return period !== null && status !== 'done' && comparePeriods(period, current) < 0;
}

export function isMilestoneOverdue(milestone: JourneyMilestone, current: TargetPeriod): boolean {
  return isPastDue(milestone.targetPeriod, milestone.status, current);
}

export function isJourneyOverdue(journey: CapabilityJourney, current: TargetPeriod): boolean {
  return isPastDue(journey.targetPeriod, journey.status, current);
}

function isActive(journey: CapabilityJourney): boolean {
  return journey.status === 'planned' || journey.status === 'in-flight';
}

function shownPeriods(journeys: CapabilityJourney[]): TargetPeriod[] {
  return journeys.flatMap((journey) => [
    ...(journey.targetPeriod ? [journey.targetPeriod] : []),
    ...journey.milestones.flatMap((milestone) => (milestone.targetPeriod ? [milestone.targetPeriod] : [])),
  ]);
}

function quarterRange(periods: TargetPeriod[], current: TargetPeriod): TargetPeriod[] {
  const ranks = [...periods.map(periodRank), periodRank(current)];
  const quarters: TargetPeriod[] = [];
  for (let rank = Math.min(...ranks); rank <= Math.max(...ranks); rank++) {
    const quarter = ((rank - 1) % 4) + 1;
    quarters.push({ year: (rank - quarter) / 4, quarter });
  }
  return quarters;
}

function orderedMilestones(milestones: JourneyMilestone[]): JourneyMilestone[] {
  return [...milestones.filter((m) => m.targetPeriod !== null), ...milestones.filter((m) => m.targetPeriod === null)];
}

function toJourneyRow(
  journey: CapabilityJourney,
  current: TargetPeriod,
  columnOf: (period: TargetPeriod | null) => number | null,
): TimelineJourneyRow {
  return {
    journey,
    overdue: isJourneyOverdue(journey, current),
    targetColumn: columnOf(journey.targetPeriod),
    milestones: orderedMilestones(journey.milestones).map((milestone) => ({
      milestone,
      overdue: isMilestoneOverdue(milestone, current),
      column: columnOf(milestone.targetPeriod),
    })),
  };
}

function groupByDomain(
  rows: TimelineJourneyRow[],
  domainNameFor: (capabilityId: string) => string | undefined,
): TimelineDomainGroup[] {
  const groups = new Map<string, TimelineJourneyRow[]>();
  for (const row of rows) {
    const domainName = domainNameFor(row.journey.capabilityId);
    if (domainName === undefined) continue;
    groups.set(domainName, [...(groups.get(domainName) ?? []), row]);
  }
  return [...groups.entries()]
    .sort(([first], [second]) => first.localeCompare(second))
    .map(([domainName, journeys]) => ({ domainName, journeys }));
}

function summarize(groups: TimelineDomainGroup[]): TimelineSummary {
  const rows = groups.flatMap((group) => group.journeys);
  return {
    planned: rows.filter((row) => row.journey.status === 'planned').length,
    inFlight: rows.filter((row) => row.journey.status === 'in-flight').length,
    overdueJourneys: rows.filter((row) => row.overdue).length,
    overdueMilestones: rows.flatMap((row) => row.milestones).filter((row) => row.overdue).length,
  };
}

export function buildTimeline({ journeys, domainNameFor, current }: BuildTimelineParams): TimelineModel {
  const active = journeys
    .filter(isActive)
    .filter((journey) => domainNameFor(journey.capabilityId) !== undefined)
    .sort(byTargetPeriodThenName);

  const quarters = quarterRange(shownPeriods(active), current);
  const firstRank = periodRank(quarters[0]);
  const columnOf = (period: TargetPeriod | null) => (period ? periodRank(period) - firstRank : null);

  const groups = groupByDomain(
    active.map((journey) => toJourneyRow(journey, current, columnOf)),
    domainNameFor,
  );

  return { quarters, currentColumn: periodRank(current) - firstRank, groups, summary: summarize(groups) };
}
