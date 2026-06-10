import { Badge, Button, Group, Loader, Paper, Stack, Text, Title } from '@mantine/core';
import type { CompositionDomainGroup, CompositionResponse, IncludedCapabilityItem } from '../types';

interface IncludedCapabilitiesSectionProps {
  composition: CompositionResponse | undefined;
  isLoading: boolean;
  onExclude: (capabilityId: string) => void;
  isExcluding: boolean;
}

const LEVEL_COLORS: Record<IncludedCapabilityItem['level'], string> = {
  L1: 'indigo.9',
  L2: 'blue.8',
  L3: 'blue.6',
  L4: 'blue.4',
};

export function IncludedCapabilitiesSection({
  composition,
  isLoading,
  onExclude,
  isExcluding,
}: IncludedCapabilitiesSectionProps) {
  return (
    <Stack gap="sm" data-testid="included-capabilities">
      <Group justify="space-between" align="center">
        <Title order={4}>Included capabilities</Title>
        {composition && composition.meta.sourceCount > 0 && (
          <Badge variant="light" color="gray" data-testid="composition-counts">
            {composition.meta.sourceCount} sources · {composition.meta.includedCount} included
          </Badge>
        )}
      </Group>

      {isLoading ? (
        <Group gap="xs" data-testid="included-capabilities-loading">
          <Loader size="sm" />
          <Text size="sm" c="dimmed">
            Loading included capabilities…
          </Text>
        </Group>
      ) : !composition || composition.data.length === 0 ? (
        <EmptyState />
      ) : (
        <>
          {composition.data.map((group) => (
            <DomainGroup
              key={group.businessDomainId ?? '__unassigned__'}
              group={group}
              onExclude={onExclude}
              isExcluding={isExcluding}
            />
          ))}
          <Text size="xs" c="dimmed">
            A source implicitly includes its descendants. A descendant sourced by a more specific direction on another
            EC is carved out and owned there. Excluding a source removes it (and its implicit children) from this
            direction.
          </Text>
        </>
      )}
    </Stack>
  );
}

function EmptyState() {
  return (
    <Paper withBorder p="lg" radius="md" data-testid="included-empty-state">
      <Text size="sm" c="dimmed" ta="center">
        No capabilities included yet. Capture a direction to compose this Enterprise Capability.
      </Text>
    </Paper>
  );
}

interface DomainGroupProps {
  group: CompositionDomainGroup;
  onExclude: (capabilityId: string) => void;
  isExcluding: boolean;
}

function DomainGroup({ group, onExclude, isExcluding }: DomainGroupProps) {
  return (
    <Stack gap={4}>
      <Paper bg="gray.1" px="md" py="xs" radius="sm">
        <Text size="sm" fw={600} c="gray.7">
          {group.businessDomainName ?? 'Unassigned'}
        </Text>
      </Paper>
      <Stack gap={4}>
        {group.items.map((item) => (
          <IncludedRow key={item.capabilityId} item={item} onExclude={onExclude} isExcluding={isExcluding} />
        ))}
      </Stack>
    </Stack>
  );
}

interface IncludedRowProps {
  item: IncludedCapabilityItem;
  onExclude: (capabilityId: string) => void;
  isExcluding: boolean;
}

function IncludedRow({ item, onExclude, isExcluding }: IncludedRowProps) {
  const carved = item.role === 'carved-out';
  const canExclude = item.role === 'source' && !!item._links['x-exclude'];
  return (
    <Group
      justify="space-between"
      gap="sm"
      wrap="nowrap"
      px="md"
      py="xs"
      data-testid={`included-row-${item.capabilityId}`}
    >
      <Group gap="xs" wrap="nowrap">
        <Badge size="xs" variant="filled" color={LEVEL_COLORS[item.level]} radius="sm">
          {item.level}
        </Badge>
        <Text
          size="sm"
          fw={item.role === 'source' ? 600 : 400}
          td={carved ? 'line-through' : undefined}
          c={carved ? 'dimmed' : undefined}
        >
          {item.name}
        </Text>
        <RoleTag item={item} />
      </Group>
      {canExclude && (
        <Button
          size="compact-xs"
          variant="subtle"
          color="red"
          disabled={isExcluding}
          onClick={() => onExclude(item.capabilityId)}
          data-testid={`exclude-${item.capabilityId}`}
        >
          Exclude
        </Button>
      )}
    </Group>
  );
}

function RoleTag({ item }: { item: IncludedCapabilityItem }) {
  if (item.role === 'source') {
    return (
      <Badge size="xs" variant="light" color="blue">
        source
      </Badge>
    );
  }
  if (item.role === 'implicit') {
    return (
      <Badge size="xs" variant="light" color="gray">
        via parent
      </Badge>
    );
  }
  return (
    <Badge size="xs" variant="light" color="red">
      carved out · owned by {item.carvedOutBy?.enterpriseCapabilityName}
    </Badge>
  );
}
