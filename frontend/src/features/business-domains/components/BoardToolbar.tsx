import { Button, Group, Stack, Switch, TextInput } from '@mantine/core';
import { IconSearch } from '@tabler/icons-react';
import type { BoardLens } from '../lens/boardLens';
import type { SummaryCounts } from '../lens/journeyIndex';
import { BoardLegend } from './BoardLegend';
import classes from './BoardToolbar.module.css';
import { LensSwitcher } from './LensSwitcher';

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
}: BoardToolbarProps) {
  return (
    <Stack gap="sm">
      <Group justify="space-between" wrap="wrap" gap="md">
        <LensSwitcher lens={lens} onLensChange={onLensChange} />
        <Group gap="sm">
          <TextInput
            value={searchQuery}
            onChange={(e) => onSearchChange(e.currentTarget.value)}
            placeholder="Filter capabilities or apps..."
            leftSection={<IconSearch size={14} />}
            data-testid="board-search-input"
            className={classes.searchInput}
          />
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

      {lens !== 'now' && (
        <Switch
          checked={changesOnly}
          onChange={(e) => onChangesOnlyChange(e.currentTarget.checked)}
          label="Highlight only what changed"
          data-testid="changes-only-toggle"
        />
      )}
      {lens === 'journey' && <BoardSummary summary={summary} />}
      <BoardLegend lens={lens} />
    </Stack>
  );
}
