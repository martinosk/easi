import { SegmentedControl, Stack, Text, Title } from '@mantine/core';
import { useState } from 'react';
import { getSkin, SKINS, type SkinName, setSkin } from '../../../theme/skin';

export function AppearanceSettings() {
  const [skin, setSkinState] = useState<SkinName>(getSkin());

  const handleChange = (value: string) => {
    const next = SKINS.find((option) => option.value === value);
    if (!next) return;
    setSkin(next.value);
    setSkinState(next.value);
  };

  return (
    <Stack gap="lg">
      <Stack gap={0}>
        <Title order={2}>Appearance</Title>
        <Text c="dimmed" size="sm">
          Choose the chrome colours for this workstation.
        </Text>
      </Stack>
      <Stack gap="xs">
        <Text size="sm" fw={500}>
          Theme skin
        </Text>
        <SegmentedControl
          aria-label="Theme skin"
          data={SKINS.map((option) => ({ value: option.value, label: option.label }))}
          value={skin}
          onChange={handleChange}
        />
        <Text size="xs" c="dimmed">
          Chrome colours only — status colours stay the same for every tenant.
        </Text>
      </Stack>
    </Stack>
  );
}
