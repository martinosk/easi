import { useState, useCallback } from 'react';
import type { Node, Edge } from '@xyflow/react';
import { useAppStore } from '../../../store/appStore';

export interface NodeContextMenu {
  x: number;
  y: number;
  nodeId: string;
  nodeName: string;
  nodeType: 'component' | 'capability';
}

export interface EdgeContextMenu {
  x: number;
  y: number;
  edgeId: string;
  edgeName: string;
  edgeType: 'relation' | 'parent' | 'realization';
  realizationId?: string;
}

export const useContextMenu = () => {
  const components = useAppStore((state) => state.components);
  const capabilities = useAppStore((state) => state.capabilities);
  const relations = useAppStore((state) => state.relations);
  const capabilityRealizations = useAppStore((state) => state.capabilityRealizations);

  const [nodeContextMenu, setNodeContextMenu] = useState<NodeContextMenu | null>(null);
  const [edgeContextMenu, setEdgeContextMenu] = useState<EdgeContextMenu | null>(null);

  const onNodeContextMenu = useCallback(
    (event: React.MouseEvent, node: Node) => {
      event.preventDefault();
      if (node.type === 'capability') {
        const capId = node.id.replace('cap-', '');
        const capability = capabilities.find((c) => c.id === capId);
        if (capability) {
          setNodeContextMenu({
            x: event.clientX,
            y: event.clientY,
            nodeId: capId,
            nodeName: capability.name,
            nodeType: 'capability',
          });
        }
      } else {
        const component = components.find((c) => c.id === node.id);
        if (component) {
          setNodeContextMenu({
            x: event.clientX,
            y: event.clientY,
            nodeId: node.id,
            nodeName: component.name,
            nodeType: 'component',
          });
        }
      }
    },
    [components, capabilities]
  );

  const onEdgeContextMenu = useCallback(
    (event: React.MouseEvent, edge: Edge) => {
      event.preventDefault();
      if (edge.id.startsWith('parent-')) {
        setEdgeContextMenu({
          x: event.clientX,
          y: event.clientY,
          edgeId: edge.id,
          edgeName: 'Parent',
          edgeType: 'parent',
        });
      } else if (edge.id.startsWith('realization-')) {
        const realizationId = edge.id.replace('realization-', '');
        const realization = capabilityRealizations.find((r) => r.id === realizationId);
        if (realization) {
          const capability = capabilities.find((c) => c.id === realization.capabilityId);
          const component = components.find((c) => c.id === realization.componentId);
          const edgeName = `${capability?.name || 'Capability'} -> ${component?.name || 'Component'}`;
          setEdgeContextMenu({
            x: event.clientX,
            y: event.clientY,
            edgeId: edge.id,
            edgeName,
            edgeType: 'realization',
            realizationId,
          });
        }
      } else {
        const relation = relations.find((r) => r.id === edge.id);
        if (relation) {
          setEdgeContextMenu({
            x: event.clientX,
            y: event.clientY,
            edgeId: edge.id,
            edgeName: relation.name || relation.relationType,
            edgeType: 'relation',
          });
        }
      }
    },
    [relations, capabilityRealizations, capabilities, components]
  );

  const closeMenus = useCallback(() => {
    setNodeContextMenu(null);
    setEdgeContextMenu(null);
  }, []);

  return {
    nodeContextMenu,
    edgeContextMenu,
    onNodeContextMenu,
    onEdgeContextMenu,
    closeMenus,
  };
};
