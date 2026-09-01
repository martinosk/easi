import { Badge, Button, Group, Stack, Text } from '@mantine/core';
import React, { useState } from 'react';
import type { Component, OwnershipState } from '../../../api/types';
import { DetailField } from '../../../components/shared/DetailField';
import { hasLink } from '../../../utils/hateoas';
import { useClearComponentOwnership, useConfirmComponentOwnership } from '../hooks/useComponentOwnership';
import { ComponentOwnerDialog, type OwnerDialogMode } from './ComponentOwnerDialog';

export const OWNERSHIP_STATE_LABELS: Record<OwnershipState, string> = {
  unknown: 'Unknown',
  nominated: 'Nominated',
  owned: 'Owned',
  managed: 'Managed',
};

const OWNERSHIP_STATE_COLORS: Record<OwnershipState, string> = {
  unknown: 'gray',
  nominated: 'yellow',
  owned: 'blue',
  managed: 'teal',
};

interface OwnershipStateBadgeProps {
  state: OwnershipState;
}

export const OwnershipStateBadge: React.FC<OwnershipStateBadgeProps> = ({ state }) => (
  <Badge color={OWNERSHIP_STATE_COLORS[state]} variant="light" size="sm" data-testid="ownership-state-badge">
    {OWNERSHIP_STATE_LABELS[state]}
  </Badge>
);

interface OwnerLineProps {
  component: Component;
}

const OwnerLine: React.FC<OwnerLineProps> = ({ component }) => {
  if (!component.owner) return null;
  const kindLabel = component.owner.kind === 'team' ? 'Internal Team' : 'User';
  return (
    <Group gap="xs">
      <Text size="sm">{component.owner.name || component.owner.id}</Text>
      <Text size="xs" c="dimmed">
        {kindLabel}
      </Text>
    </Group>
  );
};

interface OwnershipActionsProps {
  component: Component;
  onOpenDialog: (mode: OwnerDialogMode) => void;
}

const OwnershipActions: React.FC<OwnershipActionsProps> = ({ component, onOpenDialog }) => {
  const confirmMutation = useConfirmComponentOwnership();
  const clearMutation = useClearComponentOwnership();

  const actions = [
    hasLink(component, 'x-nominate-owner') && (
      <Button key="nominate" variant="default" size="xs" onClick={() => onOpenDialog('nominate')}>
        Nominate Owner
      </Button>
    ),
    hasLink(component, 'x-assign-owner') && (
      <Button key="assign" variant="default" size="xs" onClick={() => onOpenDialog('assign')}>
        Assign Owner
      </Button>
    ),
    hasLink(component, 'x-confirm-owner') && (
      <Button
        key="confirm"
        variant="default"
        size="xs"
        loading={confirmMutation.isPending}
        onClick={() => confirmMutation.mutate(component)}
      >
        Confirm
      </Button>
    ),
    hasLink(component, 'x-clear-owner') && (
      <Button
        key="clear"
        variant="default"
        size="xs"
        loading={clearMutation.isPending}
        onClick={() => clearMutation.mutate(component)}
      >
        Clear
      </Button>
    ),
  ].filter(Boolean);

  if (actions.length === 0) return null;
  return <Group gap="sm">{actions}</Group>;
};

interface ComponentOwnershipSectionProps {
  component: Component;
}

export const ComponentOwnershipSection: React.FC<ComponentOwnershipSectionProps> = ({ component }) => {
  const [dialogMode, setDialogMode] = useState<OwnerDialogMode | null>(null);

  return (
    <DetailField label="Ownership">
      <Stack gap="xs" data-testid="ownership-section">
        <Group gap="sm">
          <OwnershipStateBadge state={component.ownershipState} />
          <OwnerLine component={component} />
        </Group>
        <OwnershipActions component={component} onOpenDialog={setDialogMode} />
        {dialogMode && (
          <ComponentOwnerDialog
            mode={dialogMode}
            component={component}
            isOpen
            onClose={() => setDialogMode(null)}
          />
        )}
      </Stack>
    </DetailField>
  );
};
