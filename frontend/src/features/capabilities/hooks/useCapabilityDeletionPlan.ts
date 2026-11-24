import { useCallback, useMemo } from 'react';
import type { Capability, CapabilityDeleteImpact, CapabilityId, ImpactRealization } from '../../../api/types';
import { useCapabilities, useCascadeDeleteCapability, useDeleteImpacts } from './useCapabilities';

export interface PlannedCapabilityDeletion {
  capability: Capability;
  cascade: boolean;
  confirmedRealizationIds: string[];
}

export interface CapabilityDeletionImpact {
  deletions: PlannedCapabilityDeletion[];
  coveredIds: Set<string>;
  realizations: ImpactRealization[];
  capabilityNames: Map<string, string>;
  impacts: CapabilityDeleteImpact[];
}

export interface CapabilityDeletionPlan extends CapabilityDeletionImpact {
  isLoading: boolean;
  isError: boolean;
  ready: boolean;
  retry: () => void;
}

const uniqueById = (realizations: ImpactRealization[]) => [
  ...new Map(realizations.map((r) => [r.id, r] as const)).values(),
];

export function planCapabilityDeletion(
  capabilities: Capability[],
  impacts: CapabilityDeleteImpact[],
): CapabilityDeletionImpact {
  const impactById = new Map(impacts.map((impact) => [impact.capabilityId as string, impact] as const));
  const coveredIds = new Set(impacts.flatMap((impact) => impact.affectedCapabilities.map((c) => c.id as string)));
  const deletions = capabilities
    .filter((capability) => !coveredIds.has(capability.id))
    .flatMap((capability) => {
      const impact = impactById.get(capability.id);
      if (!impact) return [];
      return [
        {
          capability,
          cascade: impact.hasDescendants,
          confirmedRealizationIds: (impact.realizations ?? []).map((r) => r.id),
        },
      ];
    });
  const plannedImpacts = deletions.flatMap((d) => impactById.get(d.capability.id) ?? []);

  return {
    deletions,
    coveredIds,
    realizations: uniqueById(plannedImpacts.flatMap((impact) => impact.realizations ?? [])),
    capabilityNames: new Map([
      ...capabilities.map((c) => [c.id as string, c.name] as const),
      ...impacts.flatMap((impact) => impact.affectedCapabilities.map((c) => [c.id as string, c.name] as const)),
    ]),
    impacts,
  };
}

export function useCapabilityDeletionPlan(capabilities: Capability[]): CapabilityDeletionPlan {
  const { impacts, isLoading, isFetching, isError, retry } = useDeleteImpacts(
    capabilities.map((c) => c.id as CapabilityId),
  );
  const loaded = impacts.length === capabilities.length;
  return {
    ...planCapabilityDeletion(capabilities, impacts),
    isLoading,
    isError,
    ready: loaded && !isFetching && !isError,
    retry,
  };
}

export interface DeletionContext {
  domainId?: string;
}

export function useExecuteCapabilityDeletion() {
  const cascadeDelete = useCascadeDeleteCapability();
  const { mutateAsync } = cascadeDelete;
  return useCallback(
    (deletion: PlannedCapabilityDeletion, context: DeletionContext = {}) =>
      mutateAsync({
        capability: deletion.capability,
        cascade: deletion.cascade,
        confirmedRealizationIds: deletion.confirmedRealizationIds,
        parentId: deletion.capability.parentId ?? undefined,
        domainId: context.domainId,
      }),
    [mutateAsync],
  );
}

export function capabilitiesWithIds(capabilities: Capability[], ids: string[]): Capability[] {
  const wanted = new Set(ids);
  return capabilities.filter((c) => wanted.has(c.id));
}

export async function deletePlannedCapability(
  plan: CapabilityDeletionImpact,
  target: { id: string; name: string },
  execute: (deletion: PlannedCapabilityDeletion) => Promise<void>,
): Promise<void> {
  if (plan.coveredIds.has(target.id)) return;
  const deletion = plan.deletions.find((d) => d.capability.id === target.id);
  if (!deletion) throw new Error(`Capability not found: ${target.name}`);
  await execute(deletion);
}

export function useSelectedCapabilityDeletion(capabilityIds: string[]) {
  const { data: capabilities = [] } = useCapabilities();
  const selected = useMemo(() => capabilitiesWithIds(capabilities, capabilityIds), [capabilities, capabilityIds]);
  const plan = useCapabilityDeletionPlan(selected);
  const execute = useExecuteCapabilityDeletion();
  const deleteCapability = useCallback(
    (target: { id: string; name: string }) => deletePlannedCapability(plan, target, execute),
    [plan, execute],
  );
  return { plan, deleteCapability };
}
