import { Button, Group, TextInput } from '@mantine/core';
import { IconSearch } from '@tabler/icons-react';
import classes from './BoardToolbar.module.css';

export interface BoardToolbarProps {
  searchQuery: string;
  onSearchChange: (value: string) => void;
  canCreateDomain: boolean;
  onCreateDomain: () => void;
  showAssignToggle: boolean;
  assignRailOpen: boolean;
  onToggleAssignRail: () => void;
}

const LEGEND_ITEMS: { swatchClass: keyof typeof classes; label: string }[] = [
  { swatchClass: 'swatchFull', label: 'Full' },
  { swatchClass: 'swatchPartial', label: 'Partial' },
  { swatchClass: 'swatchPlanned', label: 'Planned' },
  { swatchClass: 'swatchInherited', label: 'Inherited' },
];

function Legend() {
  return (
    <div className={classes.legend} data-testid="board-legend">
      {LEGEND_ITEMS.map((item) => (
        <span key={item.label} className={classes.legendItem}>
          <span className={[classes.swatch, classes[item.swatchClass]].join(' ')} />
          {item.label}
        </span>
      ))}
    </div>
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
}: BoardToolbarProps) {
  return (
    <Group justify="space-between" wrap="wrap" gap="md">
      <Group gap="lg" wrap="wrap">
        <TextInput
          value={searchQuery}
          onChange={(e) => onSearchChange(e.currentTarget.value)}
          placeholder="Filter capabilities or apps..."
          leftSection={<IconSearch size={14} />}
          data-testid="board-search-input"
          className={classes.searchInput}
        />
        <Legend />
      </Group>
      <Group gap="sm">
        {showAssignToggle && (
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
  );
}
