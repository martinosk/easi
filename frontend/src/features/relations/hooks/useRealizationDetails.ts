import { useMemo } from 'react';
import type { Capability, CapabilityRealization, Component } from '../../../api/types';
import { useAppStore } from '../../../store/appStore';
import { useCapabilities, useRealizations } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../../components/hooks/useComponents';

const REALIZATION_PREFIX = 'realization-';

export interface RealizationData {
  realization: CapabilityRealization;
  capability: Capability | undefined;
  component: Component | undefined;
  formattedDate: string;
  isInherited: boolean;
}

const isRealizationEdge = (edgeId: string | null): boolean => edgeId !== null && edgeId.startsWith(REALIZATION_PREFIX);

const extractRealizationId = (edgeId: string): string => edgeId.replace(REALIZATION_PREFIX, '');

const getRealizationData = (
  selectedEdgeId: string | null,
  capabilityRealizations: CapabilityRealization[],
  capabilities: Capability[],
  components: Component[],
): RealizationData | null => {
  if (!isRealizationEdge(selectedEdgeId)) {
    return null;
  }

  const realizationId = extractRealizationId(selectedEdgeId!);
  const realization = capabilityRealizations.find((r) => r.id === realizationId);

  if (!realization) {
    return null;
  }

  const capability = capabilities.find((c) => c.id === realization.capabilityId);
  const component = components.find((c) => c.id === realization.componentId);
  const formattedDate = new Date(realization.linkedAt).toLocaleString();
  const isInherited = realization.origin === 'Inherited';

  return { realization, capability, component, formattedDate, isInherited };
};

export const useRealizationDetails = (): RealizationData | null => {
  const selectedEdgeId = useAppStore((state) => state.selectedEdgeId);
  const { data: components = [] } = useComponents();
  const { data: capabilities = [] } = useCapabilities();

  const { data: capabilityRealizations = [] } = useRealizations();

  return useMemo(
    () => getRealizationData(selectedEdgeId, capabilityRealizations, capabilities, components),
    [selectedEdgeId, capabilityRealizations, capabilities, components],
  );
};
