import { useMemo } from 'react';
import type { Capability, ComponentId } from '../../../api/types';
import { hasLink } from '../../../utils/hateoas';
import { useRealizationRolesByCapabilityIds } from '../../architecture-direction/hooks/useRealizationRoles';
import type { RealizationRoleAssignment } from '../../architecture-direction/types';

export interface CapabilityRoles {
  getRole: (componentId: ComponentId | string) => RealizationRoleAssignment | undefined;
  canAssign: boolean;
}

export function useCapabilityRoles(capability: Capability | null): CapabilityRoles {
  const capabilityIds = useMemo(() => (capability ? [String(capability.id)] : []), [capability]);
  const rolesQuery = useRealizationRolesByCapabilityIds(capabilityIds);

  const roleByComponentId = useMemo(() => {
    const map = new Map<string, RealizationRoleAssignment>();
    for (const role of rolesQuery.data?.data ?? []) map.set(String(role.componentId), role);
    return map;
  }, [rolesQuery.data]);

  return {
    getRole: (componentId) => roleByComponentId.get(String(componentId)),
    canAssign: hasLink(rolesQuery.data, 'x-assign'),
  };
}
