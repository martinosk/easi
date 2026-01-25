import { useMemo } from 'react';
import { useAppStore } from '../../../store/appStore';
import { useCapabilities, useRealizationsForComponents } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../../components/hooks/useComponents';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import type { CapabilityRealization, Capability, Component } from '../../../api/types';

const REALIZATION_PREFIX = 'realization-';

export interface RealizationData {
  realization: CapabilityRealization;
  capability: Capability | undefined;
  component: Component | undefined;
  formattedDate: string;
  isInherited: boolean;
}

const isRealizationEdge = (edgeId: string | null): boolean =>
  edgeId !== null && edgeId.startsWith(REALIZATION_PREFIX);

const extractRealizationId = (edgeId: string): string =>
  edgeId.replace(REALIZATION_PREFIX, '');

const getRealizationData = (
  selectedEdgeId: string | null,
  capabilityRealizations: CapabilityRealization[],
  capabilities: Capability[],
  components: Component[]
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
  const { currentView } = useCurrentView();
  const { data: components = [] } = useComponents();
  const { data: capabilities = [] } = useCapabilities();

  const componentIdsInView = useMemo(
    () => currentView?.components.map((vc) => vc.componentId) || [],
    [currentView?.components]
  );
  const { data: capabilityRealizations = [] } = useRealizationsForComponents(componentIdsInView);

  return useMemo(
    () => getRealizationData(selectedEdgeId, capabilityRealizations, capabilities, components),
    [selectedEdgeId, capabilityRealizations, capabilities, components]
  );
};
