import { Badge, Button, Group } from '@mantine/core';
import type { CapabilityId, ComponentId } from '../../../api/types';
import { hasLink } from '../../../utils/hateoas';
import { useAssignRealizationRole, useClearRealizationRole } from '../../architecture-direction/hooks/useRealizationRoles';
import type { RealizationRole, RealizationRoleAssignment } from '../../architecture-direction/types';

const ROLE_LABEL: Record<RealizationRole, string> = {
  standard: 'standard',
  legacy: 'legacy',
};

const ROLE_COLOR: Record<RealizationRole, string> = {
  standard: 'teal',
  legacy: 'orange',
};

function otherRole(role: RealizationRole): RealizationRole {
  return role === 'standard' ? 'legacy' : 'standard';
}

function UnclassifiedControls({
  componentId,
  isAssigning,
  onAssign,
}: {
  componentId: string;
  isAssigning: boolean;
  onAssign: (role: RealizationRole) => void;
}) {
  return (
    <Group gap="xs" data-testid={`role-${componentId}`}>
      <Button
        variant="subtle"
        size="compact-xs"
        onClick={() => onAssign('standard')}
        loading={isAssigning}
        data-testid={`assign-standard-btn-${componentId}`}
      >
        Mark standard
      </Button>
      <Button
        variant="subtle"
        size="compact-xs"
        onClick={() => onAssign('legacy')}
        loading={isAssigning}
        data-testid={`assign-legacy-btn-${componentId}`}
      >
        Mark legacy
      </Button>
    </Group>
  );
}

function ClassifiedControls({
  componentId,
  role,
  canEdit,
  canDelete,
  isAssigning,
  isClearing,
  onAssign,
  onClear,
}: {
  componentId: string;
  role: RealizationRoleAssignment;
  canEdit: boolean;
  canDelete: boolean;
  isAssigning: boolean;
  isClearing: boolean;
  onAssign: (role: RealizationRole) => void;
  onClear: () => void;
}) {
  const next = otherRole(role.role);

  return (
    <Group gap="xs" data-testid={`role-${componentId}`}>
      <Badge color={ROLE_COLOR[role.role]} variant="light" data-testid={`role-badge-${componentId}`}>
        {ROLE_LABEL[role.role]}
      </Badge>
      {canEdit && (
        <Button
          variant="subtle"
          size="compact-xs"
          onClick={() => onAssign(next)}
          loading={isAssigning}
          data-testid={`assign-${next}-btn-${componentId}`}
        >
          Mark {next}
        </Button>
      )}
      {canDelete && (
        <Button
          variant="subtle"
          color="red"
          size="compact-xs"
          onClick={onClear}
          loading={isClearing}
          data-testid={`clear-role-btn-${componentId}`}
        >
          Clear
        </Button>
      )}
    </Group>
  );
}

export interface RealizationRoleControlProps {
  capabilityId: CapabilityId | string;
  componentId: ComponentId | string;
  role: RealizationRoleAssignment | undefined;
  canAssign: boolean;
}

export function RealizationRoleControl({ capabilityId, componentId, role, canAssign }: RealizationRoleControlProps) {
  const assignMutation = useAssignRealizationRole();
  const clearMutation = useClearRealizationRole();

  const id = String(componentId);

  const handleAssign = (nextRole: RealizationRole) =>
    assignMutation.mutateAsync({
      capabilityId: String(capabilityId),
      componentId: id,
      request: { role: nextRole },
    });

  const handleClear = () => {
    if (!role) return;
    clearMutation.mutateAsync({ role });
  };

  if (!role) {
    if (!canAssign) return null;
    return <UnclassifiedControls componentId={id} isAssigning={assignMutation.isPending} onAssign={handleAssign} />;
  }

  return (
    <ClassifiedControls
      componentId={id}
      role={role}
      canEdit={hasLink(role, 'edit')}
      canDelete={hasLink(role, 'delete')}
      isAssigning={assignMutation.isPending}
      isClearing={clearMutation.isPending}
      onAssign={handleAssign}
      onClear={handleClear}
    />
  );
}
