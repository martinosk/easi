import { Badge, Button, Divider, Group, Popover, Stack, Text, UnstyledButton } from '@mantine/core';
import type { ReactNode } from 'react';
import type { DynamicFilters, EdgeType, UnexpandedByEdgeType } from '../utils/dynamicMode';

interface ExpandPopoverProps {
  entityName: string;
  breakdown: UnexpandedByEdgeType;
  enabledEdgeTypes: DynamicFilters['edges'];
  opened: boolean;
  onClose: () => void;
  onExpandEdgeType: (edge: EdgeType) => void;
  onExpandAll: () => void;
  children: ReactNode;
}

const EDGE_LABEL: Record<EdgeType, string> = {
  relation: 'Triggers / Serves',
  realization: 'Realization',
  parentage: 'Capability parentage',
  origin: 'Origin',
};

const EDGE_ORDER: EdgeType[] = ['relation', 'realization', 'parentage', 'origin'];

export function ExpandPopover({
  entityName,
  breakdown,
  enabledEdgeTypes,
  opened,
  onClose,
  onExpandEdgeType,
  onExpandAll,
  children,
}: ExpandPopoverProps) {
  const total = EDGE_ORDER.reduce((acc, et) => (enabledEdgeTypes[et] ? acc + breakdown[et].length : acc), 0);

  return (
    <Popover opened={opened} onChange={(o) => !o && onClose()} position="right-start" withArrow shadow="md">
      <Popover.Target>{children}</Popover.Target>
      <Popover.Dropdown>
        <Stack gap="xs" miw={220}>
          <Text size="xs" tt="uppercase" c="dimmed">
            Expand from {entityName}
          </Text>
          {EDGE_ORDER.filter((et) => enabledEdgeTypes[et]).map((et) => {
            const count = breakdown[et].length;
            return (
              <UnstyledButton
                key={et}
                aria-label={EDGE_LABEL[et]}
                disabled={count === 0}
                onClick={() => onExpandEdgeType(et)}
                style={{ opacity: count === 0 ? 0.5 : 1 }}
              >
                <Group justify="space-between" wrap="nowrap">
                  <Text size="sm">{EDGE_LABEL[et]}</Text>
                  <Badge size="sm" variant="light">+{count}</Badge>
                </Group>
              </UnstyledButton>
            );
          })}
          {total > 0 && (
            <>
              <Divider />
              <Button size="xs" onClick={onExpandAll}>
                Expand all (+{total})
              </Button>
            </>
          )}
        </Stack>
      </Popover.Dropdown>
    </Popover>
  );
}
