import React, { useEffect, useState } from 'react';
import { Alert, Box, Button, Center, Grid, Group, Loader, Modal, ScrollArea, Stack, Text } from '@mantine/core';
import { apiClient } from '../../../api/client';
import type { Release } from '../../../api/types';
import { ReleaseNotesContent } from './ReleaseNotesContent';
import { ReleaseNotesSidebar } from './ReleaseNotesSidebar';
import { compareSemver } from './releaseNotesUtils';

interface ReleaseNotesBrowserProps {
  isOpen: boolean;
  onClose: () => void;
}

interface ReleaseNotesState {
  releases: Release[];
  currentVersion: string | null;
  selectedRelease: Release | null;
  isLoading: boolean;
  error: string | null;
}

function useReleaseNotesData(isOpen: boolean) {
  const [state, setState] = useState<ReleaseNotesState>({
    releases: [],
    currentVersion: null,
    selectedRelease: null,
    isLoading: true,
    error: null,
  });

  useEffect(() => {
    if (!isOpen) return;
    let cancelled = false;
    async function fetchData() {
      try {
        setState((s) => ({ ...s, isLoading: true, error: null }));
        const [releasesData, version] = await Promise.all([apiClient.getReleases(), apiClient.getVersion()]);
        if (cancelled) return;
        const sortedReleases = [...releasesData].sort((a, b) => compareSemver(b.version, a.version));
        setState({
          releases: sortedReleases,
          currentVersion: version,
          selectedRelease: sortedReleases[0] ?? null,
          isLoading: false,
          error: null,
        });
      } catch {
        if (cancelled) return;
        setState((s) => ({ ...s, isLoading: false, error: 'Failed to load release notes' }));
      }
    }
    fetchData();
    return () => {
      cancelled = true;
    };
  }, [isOpen]);

  const selectRelease = (release: Release) => setState((s) => ({ ...s, selectedRelease: release }));

  return { ...state, selectRelease };
}

export const ReleaseNotesBrowser: React.FC<ReleaseNotesBrowserProps> = ({ isOpen, onClose }) => {
  const { releases, currentVersion, selectedRelease, isLoading, error, selectRelease } = useReleaseNotesData(isOpen);

  return (
    <Modal
      opened={isOpen}
      onClose={onClose}
      title={
        <Group gap="md">
          <Text fw={600} size="lg">
            Release Notes
          </Text>
          {currentVersion && (
            <Text size="sm" c="dimmed">
              Current: v{currentVersion}
            </Text>
          )}
        </Group>
      }
      size="80%"
      centered
      data-testid="release-notes-browser"
    >
      <Stack gap="md" mih={500}>
        {isLoading && (
          <Center py="xl">
            <Group gap="sm">
              <Loader size="sm" />
              <Text c="dimmed">Loading release notes...</Text>
            </Group>
          </Center>
        )}

        {error && (
          <Alert color="red" variant="light">
            {error}
          </Alert>
        )}

        {!isLoading && !error && (
          <Grid gutter="md">
            <Grid.Col span={{ base: 12, sm: 4 }}>
              <ScrollArea h={500}>
                <ReleaseNotesSidebar
                  releases={releases}
                  selectedRelease={selectedRelease}
                  currentVersion={currentVersion}
                  onSelectRelease={selectRelease}
                />
              </ScrollArea>
            </Grid.Col>
            <Grid.Col span={{ base: 12, sm: 8 }}>
              <ScrollArea h={500}>
                <Box component="main">
                  <ReleaseNotesContent selectedRelease={selectedRelease} />
                </Box>
              </ScrollArea>
            </Grid.Col>
          </Grid>
        )}

        <Group justify="flex-end" gap="sm">
          <Button onClick={onClose}>Close</Button>
        </Group>
      </Stack>
    </Modal>
  );
};
