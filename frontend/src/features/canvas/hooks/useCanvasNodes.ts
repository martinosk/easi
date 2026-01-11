import { useMemo } from 'react';
import type { Node } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';
import { useCanvasLayoutContext } from '../context/CanvasLayoutContext';
import { createComponentNode, createCapabilityNode, isComponentInView } from '../utils/nodeFactory';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../../components/hooks/useComponents';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import type { ViewCapability } from '../../../api/types';

export const useCanvasNodes = (): Node[] => {
  const { data: components = [] } = useComponents();
  const { currentView } = useCurrentView();
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const { data: capabilities = [] } = useCapabilities();
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
  const { positions: layoutPositions } = useCanvasLayoutContext();

  return useMemo(() => {
    if (!currentView) return [];

    const componentNodes = components
      .filter((component) => isComponentInView(component, currentView))
      .map((component) => createComponentNode(component, currentView, layoutPositions, selectedNodeId));

    const capabilityNodes = (currentView.capabilities || [])
      .map((vc: ViewCapability) => {
        const capability = capabilities.find((c) => c.id === vc.capabilityId);
        if (!capability) return null;

        return createCapabilityNode(vc.capabilityId, capability, layoutPositions, vc, selectedCapabilityId);
      })
      .filter((n): n is Node => n !== null);

    return [...componentNodes, ...capabilityNodes];
  }, [components, currentView, selectedNodeId, capabilities, selectedCapabilityId, layoutPositions]);
};
