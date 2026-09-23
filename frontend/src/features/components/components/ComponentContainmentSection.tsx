import { Badge, Group, Stack, Text } from '@mantine/core';
import type React from 'react';
import type { Component, ContainmentKind } from '../../../api/types';
import { CONTAINMENT_KIND_LABELS } from '../utils/containment';

const CONTAINMENT_KIND_COLORS: Record<ContainmentKind, string> = {
  composition: 'indigo',
  aggregation: 'teal',
};

interface ContainmentKindBadgeProps {
  kind: ContainmentKind;
}

const ContainmentKindBadge: React.FC<ContainmentKindBadgeProps> = ({ kind }) => (
  <Badge color={CONTAINMENT_KIND_COLORS[kind]} variant="light" size="sm" data-testid="containment-kind-badge">
    {CONTAINMENT_KIND_LABELS[kind]}
  </Badge>
);

interface PartOfLineProps {
  component: Component;
}

const PartOfLine: React.FC<PartOfLineProps> = ({ component }) => {
  if (!component.partOf) return null;

  return (
    <Group gap="sm" data-testid="part-of-line">
      <Text size="sm">Part of {component.partOf.name}</Text>
      <ContainmentKindBadge kind={component.partOf.kind} />
    </Group>
  );
};

interface PartsListProps {
  component: Component;
}

const PartsList: React.FC<PartsListProps> = ({ component }) => {
  const parts = component.parts ?? [];
  if (parts.length === 0) return null;

  return (
    <Stack gap="xs" data-testid="parts-list">
      <Text size="xs" c="dimmed">
        Contains
      </Text>
      {parts.map((part) => (
        <Group key={part.id} gap="sm" justify="space-between" wrap="nowrap">
          <Text size="sm">{part.name}</Text>
          <ContainmentKindBadge kind={part.kind} />
        </Group>
      ))}
    </Stack>
  );
};

interface ComponentContainmentSectionProps {
  component: Component;
}

export const ComponentContainmentSection: React.FC<ComponentContainmentSectionProps> = ({ component }) => {
  const isStandalone = !component.partOf && (component.parts ?? []).length === 0;

  return (
    <Stack gap="xs" data-testid="containment-section">
      <PartOfLine component={component} />
      <PartsList component={component} />
      {isStandalone && (
        <Text size="sm" c="dimmed">
          Standalone
        </Text>
      )}
    </Stack>
  );
};
