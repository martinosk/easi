import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { buildCapabilityJourney, buildJourneyMilestone, renderWithProviders } from '../../../test/helpers';
import type { CapabilityJourney, TargetPeriod } from '../../journeys/types';
import { currentTargetPeriod, formatTargetPeriodCompact } from '../../journeys/utils/period';
import { buildJourneyIndex } from '../lens/journeyIndex';
import { buildHookData } from '../testkit/hookData';
import { TimelineView } from './TimelineView';

const CURRENT = currentTargetPeriod(new Date());

function addQuarters(period: TargetPeriod, count: number): TargetPeriod {
  const rank = period.year * 4 + period.quarter + count;
  const quarter = ((rank - 1) % 4) + 1;
  return { year: (rank - quarter) / 4, quarter };
}

function renderTimeline(journeys: CapabilityJourney[], domains: Record<string, string>, overrides = {}) {
  const hookData = buildHookData({
    journeys,
    journeyIndex: buildJourneyIndex({ journeys, capabilityDomainNames: new Map(Object.entries(domains)) }),
    ...overrides,
  });
  const onViewModeChange = vi.fn();
  renderWithProviders(<TimelineView hookData={hookData} viewMode="timeline" onViewModeChange={onViewModeChange} />);
  return { hookData, onViewModeChange };
}

describe('TimelineView rows', () => {
  it('groups journey rows under their business domain with kind, status, and progress', () => {
    renderTimeline(
      [
        buildCapabilityJourney({
          id: 'j1',
          capabilityId: 'cap-1',
          capabilityName: 'Ferry booking',
          kind: 'migration',
          status: 'in-flight',
          progress: 25,
          targetPeriod: addQuarters(CURRENT, 4),
        }),
        buildCapabilityJourney({
          id: 'j2',
          capabilityId: 'cap-2',
          capabilityName: 'Expense handling',
          kind: 'consolidation',
          status: 'planned',
          progress: null,
          targetPeriod: addQuarters(CURRENT, 2),
        }),
      ],
      { 'cap-1': 'Ferry freight', 'cap-2': 'Group functions' },
    );

    expect(screen.getByTestId('business-domains-timeline-view')).toBeInTheDocument();
    expect(screen.getByText('Ferry freight')).toBeInTheDocument();
    expect(screen.getByText('Group functions')).toBeInTheDocument();

    const ferryRow = screen.getByTestId('timeline-journey-j1');
    expect(within(ferryRow).getByText('Ferry booking')).toBeInTheDocument();
    expect(within(ferryRow).getByText(/migration/)).toBeInTheDocument();
    expect(within(ferryRow).getByText('in flight (25%)')).toBeInTheDocument();

    const expenseRow = screen.getByTestId('timeline-journey-j2');
    expect(within(expenseRow).getByText('planned')).toBeInTheDocument();
  });

  it('marks the journey target period at its quarter', () => {
    const target = addQuarters(CURRENT, 4);
    renderTimeline(
      [buildCapabilityJourney({ id: 'j1', capabilityId: 'cap-1', targetPeriod: target })],
      { 'cap-1': 'Ferry freight' },
    );

    expect(screen.getByTestId('timeline-target-j1')).toHaveTextContent(formatTargetPeriodCompact(target));
  });

  it('renders each dated milestone as a labelled row with status and period', () => {
    const period = addQuarters(CURRENT, 1);
    renderTimeline(
      [
        buildCapabilityJourney({
          id: 'j1',
          capabilityId: 'cap-1',
          milestones: [buildJourneyMilestone({ id: 'm1', label: 'Pilot live', status: 'in-flight', targetPeriod: period })],
        }),
      ],
      { 'cap-1': 'Ferry freight' },
    );

    const row = screen.getByTestId('timeline-milestone-m1');
    expect(within(row).getByText('Pilot live')).toBeInTheDocument();
    expect(within(row).getByText(formatTargetPeriodCompact(period))).toBeInTheDocument();
  });

  it('marks overdue journeys and milestones', () => {
    const past = addQuarters(CURRENT, -2);
    renderTimeline(
      [
        buildCapabilityJourney({
          id: 'j1',
          capabilityId: 'cap-1',
          status: 'in-flight',
          targetPeriod: past,
          milestones: [buildJourneyMilestone({ id: 'm1', status: 'planned', targetPeriod: past })],
        }),
      ],
      { 'cap-1': 'Ferry freight' },
    );

    expect(screen.getByTestId('timeline-journey-j1')).toHaveAttribute('data-overdue', 'true');
    const milestone = screen.getByTestId('timeline-milestone-m1');
    expect(within(milestone).getByText('overdue')).toBeInTheDocument();
  });

  it('does not mark a done milestone with a past period as overdue', () => {
    const past = addQuarters(CURRENT, -2);
    renderTimeline(
      [
        buildCapabilityJourney({
          id: 'j1',
          capabilityId: 'cap-1',
          targetPeriod: addQuarters(CURRENT, 2),
          milestones: [buildJourneyMilestone({ id: 'm1', status: 'done', targetPeriod: past })],
        }),
      ],
      { 'cap-1': 'Ferry freight' },
    );

    expect(within(screen.getByTestId('timeline-milestone-m1')).queryByText('overdue')).toBeNull();
  });

  it('renders undated milestones last, marked as having no date and never overdue', () => {
    renderTimeline(
      [
        buildCapabilityJourney({
          id: 'j1',
          capabilityId: 'cap-1',
          milestones: [
            buildJourneyMilestone({ id: 'm-undated', label: 'Legacy exit', targetPeriod: null }),
            buildJourneyMilestone({ id: 'm-dated', label: 'Core live', targetPeriod: addQuarters(CURRENT, 1) }),
          ],
        }),
      ],
      { 'cap-1': 'Ferry freight' },
    );

    const undated = screen.getByTestId('timeline-milestone-m-undated');
    expect(within(undated).getByText('no date')).toBeInTheDocument();
    expect(within(undated).queryByText('overdue')).toBeNull();

    const rows = screen.getAllByTestId(/^timeline-milestone-/);
    expect(rows.map((row) => row.dataset.testid)).toEqual(['timeline-milestone-m-dated', 'timeline-milestone-m-undated']);
  });
});

describe('TimelineView axis and summary', () => {
  it('marks the current quarter on the axis', () => {
    renderTimeline(
      [buildCapabilityJourney({ id: 'j1', capabilityId: 'cap-1', targetPeriod: addQuarters(CURRENT, 2) })],
      { 'cap-1': 'Ferry freight' },
    );

    const axis = screen.getByTestId('timeline-axis');
    const todayCell = within(axis).getByText(formatTargetPeriodCompact(CURRENT));
    expect(todayCell).toHaveAttribute('data-today', 'true');
  });

  it('spans the axis from the earliest to the latest shown period', () => {
    const earliest = addQuarters(CURRENT, -3);
    const latest = addQuarters(CURRENT, 5);
    renderTimeline(
      [
        buildCapabilityJourney({
          id: 'j1',
          capabilityId: 'cap-1',
          targetPeriod: latest,
          milestones: [buildJourneyMilestone({ id: 'm1', targetPeriod: earliest })],
        }),
      ],
      { 'cap-1': 'Ferry freight' },
    );

    const axis = screen.getByTestId('timeline-axis');
    expect(within(axis).getByText(formatTargetPeriodCompact(earliest))).toBeInTheDocument();
    expect(within(axis).getByText(formatTargetPeriodCompact(latest))).toBeInTheDocument();
  });

  it('summarises schedule health including overdue counts', () => {
    const past = addQuarters(CURRENT, -1);
    renderTimeline(
      [
        buildCapabilityJourney({
          id: 'j1',
          capabilityId: 'cap-1',
          capabilityName: 'A',
          status: 'in-flight',
          targetPeriod: past,
          milestones: [
            buildJourneyMilestone({ id: 'm1', status: 'planned', targetPeriod: past }),
            buildJourneyMilestone({ id: 'm2', status: 'in-flight', targetPeriod: past }),
          ],
        }),
        buildCapabilityJourney({
          id: 'j2',
          capabilityId: 'cap-2',
          capabilityName: 'B',
          status: 'planned',
          targetPeriod: past,
          milestones: [buildJourneyMilestone({ id: 'm3', status: 'planned', targetPeriod: past })],
        }),
      ],
      { 'cap-1': 'Ferry freight', 'cap-2': 'Group functions' },
    );

    const summary = screen.getByTestId('timeline-summary');
    expect(within(summary).getByText('in flight').previousSibling).toHaveTextContent('1');
    expect(within(summary).getByText('planned').previousSibling).toHaveTextContent('1');
    expect(within(summary).getByText('journeys overdue').previousSibling).toHaveTextContent('2');
    expect(within(summary).getByText('milestones overdue').previousSibling).toHaveTextContent('3');
  });

  it('shows an empty state when no active journeys exist', () => {
    renderTimeline([buildCapabilityJourney({ id: 'j1', capabilityId: 'cap-1', status: 'done' })], {
      'cap-1': 'Ferry freight',
    });

    expect(screen.getByTestId('timeline-empty')).toHaveTextContent('No journeys are planned or in flight.');
    expect(screen.queryByTestId('timeline-axis')).toBeNull();
  });
});

describe('TimelineView interactions', () => {
  it('returns to the board and opens the capability drawer when a journey row is activated', async () => {
    const user = userEvent.setup();
    const { hookData, onViewModeChange } = renderTimeline(
      [buildCapabilityJourney({ id: 'j1', capabilityId: 'cap-1', targetPeriod: addQuarters(CURRENT, 2) })],
      { 'cap-1': 'Ferry freight' },
    );

    await user.click(screen.getByTestId('timeline-journey-j1'));

    expect(onViewModeChange).toHaveBeenCalledWith('board');
    expect(hookData.openCapabilityById).toHaveBeenCalledWith('cap-1');
  });

  it('carries no write affordances', () => {
    renderTimeline(
      [
        buildCapabilityJourney({
          id: 'j1',
          capabilityId: 'cap-1',
          milestones: [buildJourneyMilestone({ id: 'm1' })],
        }),
      ],
      { 'cap-1': 'Ferry freight' },
    );

    expect(screen.queryByRole('button', { name: /edit|delete|add|create|assign/i })).toBeNull();
  });
});
