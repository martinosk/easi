import { Switch, TextInput } from '@mantine/core';
import { IconSearch } from '@tabler/icons-react';
import classes from './BoardToolbar.module.css';

export interface ToolbarSearchInputProps {
  value: string;
  onChange: (value: string) => void;
}

export function ToolbarSearchInput({ value, onChange }: ToolbarSearchInputProps) {
  return (
    <TextInput
      value={value}
      onChange={(e) => onChange(e.currentTarget.value)}
      placeholder="Filter capabilities or apps..."
      leftSection={<IconSearch size={14} />}
      data-testid="board-search-input"
      className={classes.searchInput}
    />
  );
}

export interface ChangesOnlySwitchProps {
  checked: boolean;
  onChange: (checked: boolean) => void;
}

export function ChangesOnlySwitch({ checked, onChange }: ChangesOnlySwitchProps) {
  return (
    <Switch
      checked={checked}
      onChange={(e) => onChange(e.currentTarget.checked)}
      label="Highlight only what changed"
      data-testid="changes-only-toggle"
    />
  );
}
