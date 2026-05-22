import React from 'react';
import { Badge, Group, Stack, Text, UnstyledButton } from '@mantine/core';
import type { Release } from '../../../api/types';
import { formatShortDate } from './releaseNotesUtils';

interface ReleaseNotesSidebarProps {
  releases: Release[];
  selectedRelease: Release | null;
  currentVersion: string | null;
  onSelectRelease: (release: Release) => void;
}

export const ReleaseNotesSidebar: React.FC<ReleaseNotesSidebarProps> = ({
  releases,
  selectedRelease,
  currentVersion,
  onSelectRelease,
}) => {
  return (
    <Stack gap="xs" component="aside">
      <Text fw={600} size="sm" px="sm">
        Releases
      </Text>
      <Stack gap={2}>
        {releases.map((release) => {
          const isSelected = selectedRelease?.version === release.version;
          const isCurrent = release.version === currentVersion;
          return (
            <UnstyledButton
              key={release.version}
              onClick={() => onSelectRelease(release)}
              data-testid="release-browser-item"
              data-selected={isSelected || undefined}
              data-current={isCurrent || undefined}
              p="sm"
              bg={isSelected ? 'blue.0' : undefined}
            >
              <Stack gap={2}>
                <Group gap="xs" wrap="nowrap">
                  <Text size="sm" fw={isSelected ? 600 : 500}>
                    v{release.version}
                  </Text>
                  {isCurrent && (
                    <Badge size="xs" variant="light" color="blue">
                      current
                    </Badge>
                  )}
                </Group>
                <Text size="xs" c="dimmed">
                  {formatShortDate(release.releaseDate)}
                </Text>
              </Stack>
            </UnstyledButton>
          );
        })}
      </Stack>
    </Stack>
  );
};
