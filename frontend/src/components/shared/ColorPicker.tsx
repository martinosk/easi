import { Box, Popover, TextInput, Tooltip, UnstyledButton } from '@mantine/core';
import { useEffect, useRef, useState } from 'react';
import { HexColorPicker } from 'react-colorful';
import classes from './ColorPicker.module.css';

const DEFAULT_COLOR = '#E0E0E0';
const COMMIT_DEBOUNCE_MS = 300;
const HEX_COLOR = /^#[0-9A-Fa-f]{6}$/;

interface ColorPickerProps {
  color: string | null;
  onChange: (color: string) => void;
  disabled: boolean;
  disabledTooltip?: string;
}

function useDraftColor(color: string | null, onChange: (color: string) => void) {
  const committed = (color ?? DEFAULT_COLOR).toUpperCase();
  const [draft, setDraft] = useState(committed);
  const [syncedFrom, setSyncedFrom] = useState(committed);
  const pendingCommit = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  if (syncedFrom !== committed) {
    setSyncedFrom(committed);
    setDraft(committed);
  }

  const cancelPendingCommit = () => {
    clearTimeout(pendingCommit.current);
    pendingCommit.current = undefined;
  };

  useEffect(() => () => clearTimeout(pendingCommit.current), []);

  const commit = (value: string) => {
    cancelPendingCommit();
    const normalized = value.toUpperCase();
    if (normalized !== committed) onChange(normalized);
  };

  const stageFromPicker = (value: string) => {
    setDraft(value.toUpperCase());
    cancelPendingCommit();
    pendingCommit.current = setTimeout(() => commit(value), COMMIT_DEBOUNCE_MS);
  };

  return { draft, setDraft, commit, stageFromPicker, commitDraft: () => commit(draft) };
}

export function ColorPicker({ color, onChange, disabled, disabledTooltip }: ColorPickerProps) {
  const [opened, setOpened] = useState(false);
  const { draft, setDraft, commit, stageFromPicker, commitDraft } = useDraftColor(color, onChange);

  const handleOpenedChange = (nextOpened: boolean) => {
    if (!nextOpened) commitDraft();
    setOpened(nextOpened);
  };

  const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const value = event.currentTarget.value;
    setDraft(value);
    if (HEX_COLOR.test(value)) {
      commit(value);
      setOpened(false);
    }
  };

  const target = (
    <Box display="inline-block" data-testid="color-picker-target">
      <UnstyledButton
        type="button"
        data-testid="color-picker-button"
        disabled={disabled}
        onClick={() => handleOpenedChange(!opened)}
        className={classes.button}
      >
        <Box data-testid="color-picker-display" className={classes.swatch} style={{ backgroundColor: draft }} />
        <span>{draft}</span>
      </UnstyledButton>
    </Box>
  );

  if (disabled) {
    return (
      <Tooltip label={disabledTooltip} disabled={!disabledTooltip}>
        {target}
      </Tooltip>
    );
  }

  return (
    <Popover
      opened={opened}
      onChange={handleOpenedChange}
      withinPortal
      position="bottom-start"
      trapFocus={false}
      shadow="lg"
      radius="md"
    >
      <Popover.Target>{target}</Popover.Target>
      <Popover.Dropdown data-testid="color-picker-popover" p="md">
        <HexColorPicker color={draft} onChange={stageFromPicker} />
        <TextInput
          data-testid="color-picker-input"
          value={draft}
          onChange={handleInputChange}
          mt="xs"
          classNames={{ input: classes.hexInput }}
        />
      </Popover.Dropdown>
    </Popover>
  );
}
