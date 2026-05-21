import { Button, Group, NumberInput, Slider, Stack, Text } from '@mantine/core';
import { useState } from 'react';

interface SetTargetMaturityModalProps {
  currentValue: number | null;
  onClose: () => void;
  onSave: (value: number) => void;
  isSaving: boolean;
  getColorForValue: (value: number) => string;
  getSectionNameForValue: (value: number) => string;
  bounds: { min: number; max: number };
}

export function SetTargetMaturityModal({
  currentValue,
  onClose,
  onSave,
  isSaving,
  getColorForValue,
  getSectionNameForValue,
  bounds,
}: SetTargetMaturityModalProps) {
  const [value, setValue] = useState<number>(currentValue ?? Math.floor((bounds.min + bounds.max) / 2));
  const section = getSectionNameForValue(value);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSave(value);
  };

  const clamp = (next: number) => Math.min(bounds.max, Math.max(bounds.min, next));

  return (
    <form onSubmit={handleSubmit}>
      <Stack gap="md">
        <Stack gap="xs">
          <Text size="sm" fw={500}>
            Target Maturity Value
          </Text>
          <Slider min={bounds.min} max={bounds.max} value={value} onChange={setValue} disabled={isSaving} />
          <Group gap="xs" align="center">
            <Text size="lg" fw={700}>
              {value}
            </Text>
            <Text size="sm" fw={600} style={{ color: getColorForValue(value) }}>
              {section}
            </Text>
          </Group>
          <NumberInput
            value={value}
            onChange={(next) => setValue(clamp(typeof next === 'number' ? next : Number(next) || bounds.min))}
            min={bounds.min}
            max={bounds.max}
            disabled={isSaving}
            aria-label="Target maturity value"
          />
        </Stack>
        <Group justify="flex-end" gap="sm">
          <Button variant="default" onClick={onClose} disabled={isSaving}>
            Cancel
          </Button>
          <Button type="submit" loading={isSaving}>
            Save
          </Button>
        </Group>
      </Stack>
    </form>
  );
}
