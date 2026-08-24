import { Group, Progress, Stack, Text, UnstyledButton } from '@mantine/core';
import { type CSSProperties, Fragment, useMemo } from 'react';
import type { JourneyMilestone } from '../../journeys/types';
import { journeyKindLabel, journeyStatusLabel } from '../../journeys/utils/journeyFormat';
import { currentTargetPeriod, formatTargetPeriodCompact } from '../../journeys/utils/period';
import type { useBusinessDomainsPage } from '../hooks/useBusinessDomainsPage';
import type { BoardViewMode } from '../hooks/useMapViewState';
import {
  buildTimeline,
  type TimelineJourneyRow,
  type TimelineMilestoneRow,
  type TimelineModel,
  type TimelineSummary,
} from '../timeline/timelineModel';
import boardClasses from './DomainBoard.module.css';
import classes from './TimelineView.module.css';
import { ViewModeToggle } from './ViewModeToggle';

type BusinessDomainsHookReturn = ReturnType<typeof useBusinessDomainsPage>;

export interface TimelineViewProps {
  hookData: BusinessDomainsHookReturn;
  viewMode: BoardViewMode;
  onViewModeChange: (mode: BoardViewMode) => void;
}

function overdueAttr(overdue: boolean): 'true' | undefined {
  return overdue ? 'true' : undefined;
}

function SummaryStat({ count, label, danger }: { count: number; label: string; danger?: boolean }) {
  return (
    <span className={classes.stat} data-danger={danger && count > 0 ? 'true' : undefined}>
      <b>{count}</b>
      <span>{label}</span>
    </span>
  );
}

function TimelineSummaryStats({ summary }: { summary: TimelineSummary }) {
  return (
    <Group gap="sm" className={classes.summary} data-testid="timeline-summary">
      <SummaryStat count={summary.inFlight} label="in flight" />
      <SummaryStat count={summary.planned} label="planned" />
      <SummaryStat count={summary.overdueJourneys} label="journeys overdue" danger />
      <SummaryStat count={summary.overdueMilestones} label="milestones overdue" danger />
    </Group>
  );
}

function TimelineLegend() {
  return (
    <Group gap="md" className={classes.legend}>
      <span className={classes.legendKey}>
        <span className={classes.dot} data-status="done" /> milestone done
      </span>
      <span className={classes.legendKey}>
        <span className={classes.dot} data-status="in-flight" /> in flight
      </span>
      <span className={classes.legendKey}>
        <span className={classes.dot} data-status="planned" /> planned
      </span>
      <span className={classes.legendKey}>
        <span className={classes.dot} data-status="planned" data-overdue="true" /> overdue
      </span>
      <span className={classes.legendKey}>
        <span className={classes.flag}>⚑</span> journey target period
      </span>
      <span className={classes.legendKey}>
        <span className={classes.todayKey} /> current quarter
      </span>
    </Group>
  );
}

function milestonePeriodLabel(milestone: JourneyMilestone): string {
  if (!milestone.targetPeriod) return '';
  const period = formatTargetPeriodCompact(milestone.targetPeriod);
  return milestone.status === 'done' ? `done · ${period}` : period;
}

function MilestoneLane({ row }: { row: TimelineMilestoneRow }) {
  const { milestone, overdue, column } = row;
  return (
    <div className={classes.gridRow} data-testid={`timeline-milestone-${milestone.id}`}>
      <span className={classes.milestone} style={{ gridColumn: `${(column ?? 0) + 2} / -1` }}>
        <span className={classes.dot} data-status={milestone.status} data-overdue={overdueAttr(overdue)} />
        <span className={classes.milestoneLabel}>{milestone.label}</span>
        {milestone.targetPeriod && (
          <span className={classes.milestoneWhen} data-status={milestone.status} data-overdue={overdueAttr(overdue)}>
            {milestonePeriodLabel(milestone)}
          </span>
        )}
        {overdue && <span className={classes.overdueChip}>overdue</span>}
        {!milestone.targetPeriod && <span className={classes.noDateChip}>no date</span>}
      </span>
    </div>
  );
}

function JourneyRows({ row, onActivate }: { row: TimelineJourneyRow; onActivate: (capabilityId: string) => void }) {
  const { journey, overdue, targetColumn, milestones } = row;
  return (
    <>
      <UnstyledButton
        className={[classes.gridRow, classes.journeyRow].join(' ')}
        data-testid={`timeline-journey-${journey.id}`}
        data-overdue={overdueAttr(overdue)}
        onClick={() => onActivate(journey.capabilityId)}
      >
        <span className={classes.journeyLabel}>
          <span className={classes.journeyTop}>
            <span className={classes.journeyName}>{journey.capabilityName}</span>
            <span className={classes.pill} data-status={journey.status} data-overdue={overdueAttr(overdue)}>
              {journeyStatusLabel(journey)}
            </span>
          </span>
          <span className={classes.journeyRoute}>
            {journeyKindLabel(journey.kind)} → {journey.toApplication.componentName}
          </span>
          {journey.status === 'in-flight' && journey.progress !== null && (
            <Progress value={journey.progress} size="xs" className={classes.progress} />
          )}
        </span>
        {targetColumn !== null && journey.targetPeriod && (
          <span
            className={classes.target}
            data-overdue={overdueAttr(overdue)}
            style={{ gridColumn: targetColumn + 2 }}
            data-testid={`timeline-target-${journey.id}`}
          >
            ⚑ {formatTargetPeriodCompact(journey.targetPeriod)}
          </span>
        )}
      </UnstyledButton>
      {milestones.map((milestone) => (
        <MilestoneLane key={milestone.milestone.id} row={milestone} />
      ))}
    </>
  );
}

function TimelineGrid({ model, onActivate }: { model: TimelineModel; onActivate: (capabilityId: string) => void }) {
  const gridStyle = {
    '--tl-columns': `var(--tl-label-w) repeat(${model.quarters.length}, var(--tl-q-w))`,
    '--tl-today-column': model.currentColumn,
  } as CSSProperties;

  return (
    <div className={classes.scroll}>
      <div className={classes.grid} style={gridStyle}>
        <div className={classes.chartBackground} aria-hidden="true" />
        <div className={classes.todayStripe} aria-hidden="true" />

        <div className={[classes.gridRow, classes.axisRow].join(' ')} data-testid="timeline-axis">
          <div className={classes.axisLabel}>Journey</div>
          {model.quarters.map((quarter, column) => (
            <div
              key={formatTargetPeriodCompact(quarter)}
              className={classes.axisCell}
              data-today={column === model.currentColumn ? 'true' : undefined}
            >
              {formatTargetPeriodCompact(quarter)}
            </div>
          ))}
        </div>

        {model.groups.map((group) => (
          <Fragment key={group.domainName}>
            <div className={[classes.gridRow, classes.domainRow].join(' ')}>
              <span className={classes.domainName}>{group.domainName}</span>
            </div>
            {group.journeys.map((row) => (
              <JourneyRows key={row.journey.id} row={row} onActivate={onActivate} />
            ))}
          </Fragment>
        ))}
      </div>
    </div>
  );
}

export function TimelineView({ hookData, viewMode, onViewModeChange }: TimelineViewProps) {
  const { journeys, journeyIndex, openCapabilityById } = hookData;

  const model = useMemo(
    () =>
      buildTimeline({
        journeys,
        domainNameFor: journeyIndex.sourceDomainName,
        current: currentTargetPeriod(new Date()),
      }),
    [journeys, journeyIndex],
  );

  const activateJourney = (capabilityId: string) => {
    onViewModeChange('board');
    openCapabilityById(capabilityId);
  };

  return (
    <Stack gap="md" className={boardClasses.page} data-testid="business-domains-timeline-view">
      <div className={boardClasses.toolbarRow}>
        <Stack gap="sm">
          <Group justify="space-between" wrap="wrap" gap="md">
            <ViewModeToggle value={viewMode} onChange={onViewModeChange} />
            <Text size="sm" c="dimmed">
              Active journeys on a quarter axis — every milestone at its target quarter, ⚑ at the journey's target
              period.
            </Text>
          </Group>
          <TimelineSummaryStats summary={model.summary} />
          <TimelineLegend />
        </Stack>
      </div>

      <div className={boardClasses.boardRow}>
        <div className={boardClasses.boardScroll}>
          <div className={classes.wrap}>
            {model.groups.length === 0 ? (
              <Text c="dimmed" data-testid="timeline-empty">
                No journeys are planned or in flight.
              </Text>
            ) : (
              <TimelineGrid model={model} onActivate={activateJourney} />
            )}
          </div>
        </div>
      </div>
    </Stack>
  );
}
