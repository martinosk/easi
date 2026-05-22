import React, { useMemo } from 'react';
import { Button, Group, List, Modal, Paper, Stack, Text, Title } from '@mantine/core';
import type { Release } from '../../../api/types';
import { formatInlineMarkdown, formatReleaseDate, getSectionStyle, parseMarkdownSections } from './releaseNotesUtils';

interface ReleaseNotesOverlayProps {
  isOpen: boolean;
  release: Release;
  onDismiss: (mode: 'forever' | 'untilNext') => void;
}

export const ReleaseNotesOverlay: React.FC<ReleaseNotesOverlayProps> = ({ isOpen, release, onDismiss }) => {
  const sections = useMemo(() => parseMarkdownSections(release.notes), [release.notes]);
  const formattedDate = useMemo(() => formatReleaseDate(release.releaseDate), [release.releaseDate]);

  return (
    <Modal
      opened={isOpen}
      onClose={() => onDismiss('untilNext')}
      title={
        <Stack gap={0}>
          <Title order={3}>What's New in {release.version}</Title>
          <Text size="sm" c="dimmed">
            {formattedDate}
          </Text>
        </Stack>
      }
      size="lg"
      centered
      data-testid="release-notes-overlay"
    >
      <Stack gap="md">
        {sections.length > 0 ? (
          sections.map((section) => {
            const sectionStyle = getSectionStyle(section.title);
            return (
              <Paper
                key={section.title}
                p="md"
                bg={`${sectionStyle.color}.0`}
                style={{ borderLeft: `4px solid var(--mantine-color-${sectionStyle.color}-6)` }}
              >
                <Title order={4} mb="sm">
                  <Text component="span" data-testid="release-notes-section-icon" mr="xs">
                    {sectionStyle.icon}
                  </Text>
                  {section.title}
                </Title>
                <List size="sm" spacing="xs">
                  {section.items.map((item) => (
                    <List.Item key={item}>{formatInlineMarkdown(item)}</List.Item>
                  ))}
                </List>
              </Paper>
            );
          })
        ) : (
          <Stack gap={0}>
            {release.notes.split('\n').map((line, index) => (
              // biome-ignore lint/suspicious/noArrayIndexKey: release notes are immutable, line order is stable
              <Text key={index} size="sm">
                {formatInlineMarkdown(line) || ' '}
              </Text>
            ))}
          </Stack>
        )}

        <Group justify="flex-end" gap="sm">
          <Button variant="default" onClick={() => onDismiss('forever')} data-testid="release-notes-hide-forever">
            Don't show again
          </Button>
          <Button onClick={() => onDismiss('untilNext')} data-testid="release-notes-dismiss">
            Got it
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
};
