import { Button, Group, Stack } from '@mantine/core';
import type { BoardViewMode } from '../hooks/useMapViewState';
import type { BoardLens } from '../lens/boardLens';
import type { SummaryCounts } from '../lens/journeyIndex';
import { BoardLegend } from './BoardLegend';
import classes from './BoardToolbar.module.css';
import { LensSwitcher } from './LensSwitcher';
import { ChangesOnlySwitch, ToolbarSearchInput } from './ToolbarControls';
import { ViewModeToggle } from './ViewModeToggle';

export interface BoardToolbarProps {
  searchQuery: string;
  onSearchChange: (value: string) => void;
  canCreateDomain: boolean;
  onCreateDomain: () => void;
  showAssignToggle: boolean;
  assignRailOpen: boolean;
  onToggleAssignRail: () => void;
  lens: BoardLens;
  onLensChange: (lens: BoardLens) => void;
  changesOnly: boolean;
  onChangesOnlyChange: (value: boolean) => void;
  summary: SummaryCounts;
  viewMode: BoardViewMode;
  onViewModeChange: (mode: BoardViewMode) => void;
}

function BoardSummary({ summary }: { summary: SummaryCounts }) {
  return (
    <Group gap="sm" className={classes.summary} data-testid="board-summary">
      <span className={[classes.stat, classes.statSettled].join(' ')}>
        <b>{summary.settled}</b> settled
      </span>
      <span className={[classes.stat, classes.statFlight].join(' ')}>
        <b>{summary.inFlight}</b> in flight
      </span>
      <span className={[classes.stat, classes.statIdle].join(' ')}>
        <b>{summary.notStarted}</b> not started
      </span>
    </Group>
  );
}

export function BoardToolbar({
  searchQuery,
  onSearchChange,
  canCreateDomain,
  onCreateDomain,
  showAssignToggle,
  assignRailOpen,
  onToggleAssignRail,
  lens,
  onLensChange,
  changesOnly,
  onChangesOnlyChange,
  summary,
  viewMode,
  onViewModeChange,
}: BoardToolbarProps) {
  return (
    <Stack gap="sm">
      <Group justify="space-between" wrap="wrap" gap="md">
        <Group gap="sm">
          <ViewModeToggle value={viewMode} onChange={onViewModeChange} />
          <LensSwitcher lens={lens} onLensChange={onLensChange} />
        </Group>
        <Group gap="sm">
          <ToolbarSearchInput value={searchQuery} onChange={onSearchChange} />
          {showAssignToggle && lens === 'now' && (
            <Button
              variant={assignRailOpen ? 'filled' : 'default'}
              onClick={onToggleAssignRail}
              data-testid="assign-rail-toggle"
            >
              Assign capabilities
            </Button>
          )}
          {canCreateDomain && (
            <Button onClick={onCreateDomain} data-testid="create-domain-button">
              New domain
            </Button>
          )}
        </Group>
      </Group>

      {lens !== 'now' && <ChangesOnlySwitch checked={changesOnly} onChange={onChangesOnlyChange} />}
      {lens === 'journey' && <BoardSummary summary={summary} />}
      <BoardLegend lens={lens} />
    </Stack>
  );
}
