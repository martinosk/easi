import { ColorSwatch, Group, Progress, Stack, Text } from '@mantine/core';
import type { TimeShares, TimeTile as TimeTileData } from '../types';
import { countOf } from '../utils/homeText';
import { Tile } from './Tile';

interface ShareSegment {
  key: keyof TimeShares;
  label: string;
  color: string;
}

const SEGMENTS: readonly ShareSegment[] = [
  { key: 'invest', label: 'Invest', color: 'var(--status-positive)' },
  { key: 'tolerate', label: 'Tolerate', color: 'var(--status-neutral)' },
  { key: 'migrate', label: 'Migrate', color: 'var(--status-progress)' },
  { key: 'eliminate', label: 'Eliminate', color: 'var(--status-danger)' },
  { key: 'notAssessed', label: 'Not assessed', color: 'var(--line-strong)' },
];

function TimeDistribution({ shares }: { shares: TimeShares }) {
  return (
    <Stack gap="xs">
      <Progress.Root size="lg" aria-hidden>
        {SEGMENTS.map((segment) => (
          <Progress.Section key={segment.key} value={shares[segment.key]} color={segment.color} />
        ))}
      </Progress.Root>
      <Group gap="sm">
        {SEGMENTS.map((segment) => (
          <Group key={segment.key} gap="xs" wrap="nowrap">
            <ColorSwatch size="xs" color={segment.color} withShadow={false} />
            <Text size="xs">{`${segment.label} ${shares[segment.key]}%`}</Text>
          </Group>
        ))}
      </Group>
    </Stack>
  );
}

export function TimeTile({ time }: { time: TimeTileData }) {
  const hasShares = time.total > 0 && time.shares !== undefined;
  return (
    <Tile label="TIME" testId="home-tile-time">
      {hasShares && time.shares ? (
        <>
          <Text size="sm" c="dimmed">
            {countOf(time.total, 'realisation', 'realisations')}
          </Text>
          <TimeDistribution shares={time.shares} />
        </>
      ) : (
        <Text size="sm" c="dimmed">
          No realisations to assess
        </Text>
      )}
    </Tile>
  );
}
