import React, { useMemo } from 'react';
import { Box, Center, List, Paper, Stack, Text, Title } from '@mantine/core';
import type { Release } from '../../../api/types';
import { formatInlineMarkdown, formatReleaseDate, getSectionStyle, parseMarkdownSections } from './releaseNotesUtils';

interface ReleaseNotesContentProps {
  selectedRelease: Release | null;
}

export const ReleaseNotesContent: React.FC<ReleaseNotesContentProps> = ({ selectedRelease }) => {
  const sections = useMemo(() => {
    if (!selectedRelease) return [];
    return parseMarkdownSections(selectedRelease.notes);
  }, [selectedRelease]);

  if (!selectedRelease) {
    return (
      <Center py="xl">
        <Text c="dimmed">No release selected</Text>
      </Center>
    );
  }

  return (
    <Stack gap="md">
      <Box>
        <Title order={3}>Version {selectedRelease.version}</Title>
        <Text size="sm" c="dimmed">
          {formatReleaseDate(selectedRelease.releaseDate)}
        </Text>
      </Box>

      {sections.length > 0 ? (
        <Stack gap="md">
          {sections.map((section) => {
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
          })}
        </Stack>
      ) : (
        <Stack gap={0}>
          {selectedRelease.notes.split('\n').map((line, index) => (
            // biome-ignore lint/suspicious/noArrayIndexKey: release notes are immutable, line order is stable
            <Text key={index} size="sm">
              {formatInlineMarkdown(line) || ' '}
            </Text>
          ))}
        </Stack>
      )}
    </Stack>
  );
};
