import React, { useCallback, useImperativeHandle, forwardRef, useState, useEffect } from 'react';
import {
  ReactFlow,
  ReactFlowProvider,
  type Node,
  type Edge,
  Background,
  Controls,
  MiniMap,
  type Connection,
  type NodeChange,
  type EdgeChange,
  applyNodeChanges,
  applyEdgeChanges,
  type NodeTypes,
  BackgroundVariant,
  MarkerType,
  Handle,
  Position,
  useReactFlow,
  ConnectionMode,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { useAppStore } from '../store/appStore';
import { ContextMenu, type ContextMenuItem } from './ContextMenu';
import { ConfirmationDialog } from './ConfirmationDialog';
import { CapabilityNode } from './CapabilityNode';
import { ComponentNode } from './ComponentNode';
import toast from 'react-hot-toast';
import type { CapabilityRealization } from '../api/types';

interface ComponentCanvasProps {
  onConnect: (source: string, target: string) => void;
  onComponentDrop?: (componentId: string, x: number, y: number) => void;
}

export interface ComponentCanvasRef {
  centerOnNode: (nodeId: string) => void;
}

const nodeTypes: NodeTypes = {
  component: ComponentNode,
  capability: CapabilityNode,
};

type HandlePair = { sourceHandle: string; targetHandle: string };

const HANDLE_PAIRS: HandlePair[] = [
  { sourceHandle: 'right', targetHandle: 'left' },
  { sourceHandle: 'bottom', targetHandle: 'top' },
  { sourceHandle: 'left', targetHandle: 'right' },
  { sourceHandle: 'top', targetHandle: 'bottom' },
];

const DEFAULT_HANDLES: HandlePair = { sourceHandle: 'top', targetHandle: 'top' };

const getNodeCenter = (node: Node): { x: number; y: number } => ({
  x: node.position.x + (node.width || 150) / 2,
  y: node.position.y + (node.height || 100) / 2,
});

const angleToHandleIndex = (angleDegrees: number): number => {
  const normalized = ((angleDegrees % 360) + 360) % 360;
  if (normalized < 45 || normalized >= 315) return 0;
  if (normalized < 135) return 1;
  if (normalized < 225) return 2;
  return 3;
};

const getBestHandles = (sourceNode: Node | undefined, targetNode: Node | undefined): HandlePair => {
  if (!sourceNode || !targetNode) return DEFAULT_HANDLES;

  const source = getNodeCenter(sourceNode);
  const target = getNodeCenter(targetNode);
  const angle = Math.atan2(target.y - source.y, target.x - source.x) * (180 / Math.PI);

  return HANDLE_PAIRS[angleToHandleIndex(angle)];
};

const ComponentCanvasInner = forwardRef<ComponentCanvasRef, ComponentCanvasProps>(
  ({ onConnect, onComponentDrop }, ref) => {
  const reactFlowInstance = useReactFlow();
  const components = useAppStore((state) => state.components);
  const relations = useAppStore((state) => state.relations);
  const currentView = useAppStore((state) => state.currentView);
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const selectedEdgeId = useAppStore((state) => state.selectedEdgeId);
  const selectNode = useAppStore((state) => state.selectNode);
  const selectEdge = useAppStore((state) => state.selectEdge);
  const clearSelection = useAppStore((state) => state.clearSelection);
  const updatePosition = useAppStore((state) => state.updatePosition);
  const deleteComponent = useAppStore((state) => state.deleteComponent);
  const deleteRelation = useAppStore((state) => state.deleteRelation);
  const removeComponentFromView = useAppStore((state) => state.removeComponentFromView);
  const saveViewportState = useAppStore((state) => state.saveViewportState);
  const getViewportState = useAppStore((state) => state.getViewportState);

  const capabilities = useAppStore((state) => state.capabilities);
  const canvasCapabilities = useAppStore((state) => state.canvasCapabilities);
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
  const addCapabilityToCanvas = useAppStore((state) => state.addCapabilityToCanvas);
  const removeCapabilityFromCanvas = useAppStore((state) => state.removeCapabilityFromCanvas);
  const updateCapabilityPosition = useAppStore((state) => state.updateCapabilityPosition);
  const selectCapability = useAppStore((state) => state.selectCapability);
  const changeCapabilityParent = useAppStore((state) => state.changeCapabilityParent);
  const deleteCapability = useAppStore((state) => state.deleteCapability);

  const capabilityRealizations = useAppStore((state) => state.capabilityRealizations);
  const loadRealizationsByComponent = useAppStore((state) => state.loadRealizationsByComponent);
  const deleteRealization = useAppStore((state) => state.deleteRealization);

  const [nodes, setNodes] = React.useState<Node[]>([]);
  const [edges, setEdges] = React.useState<Edge[]>([]);
  const [isFirstLoad, setIsFirstLoad] = React.useState(true);
  const [nodeContextMenu, setNodeContextMenu] = useState<{
    x: number;
    y: number;
    nodeId: string;
    nodeName: string;
    nodeType: 'component' | 'capability';
  } | null>(null);
  const [edgeContextMenu, setEdgeContextMenu] = useState<{
    x: number;
    y: number;
    edgeId: string;
    edgeName: string;
    edgeType: 'relation' | 'parent' | 'realization';
    realizationId?: string;
  } | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<{
    type: 'component-from-view' | 'component-from-model' | 'relation-from-model' | 'capability-from-canvas' | 'capability-from-model' | 'parent-relation' | 'realization';
    id: string;
    name: string;
    childId?: string;
  } | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  // Build nodes from components and capabilities
  React.useEffect(() => {
    if (!currentView) return;

    // Build component nodes from view
    const componentNodes: Node[] = components
      .filter((component) => {
        return currentView.components.some(
          (vc) => vc.componentId === component.id
        );
      })
      .map((component) => {
        const viewComponent = currentView.components.find(
          (vc) => vc.componentId === component.id
        );

        const position = viewComponent
          ? { x: viewComponent.x, y: viewComponent.y }
          : { x: 400, y: 300 };

        return {
          id: component.id,
          type: 'component',
          position,
          data: {
            label: component.name,
            description: component.description,
            isSelected: selectedNodeId === component.id,
            customColor: viewComponent?.customColor,
          },
        };
      });

    // Build capability nodes from canvas capabilities
    const capabilityNodes: Node[] = canvasCapabilities
      .map((cc) => {
        const capability = capabilities.find((c) => c.id === cc.capabilityId);
        if (!capability) return null;

        const viewCapability = currentView.capabilities.find((vc) => vc.capabilityId === capability.id);

        return {
          id: `cap-${capability.id}`,
          type: 'capability' as const,
          position: { x: cc.x, y: cc.y },
          data: {
            label: capability.name,
            level: capability.level,
            maturityLevel: capability.maturityLevel,
            isSelected: selectedCapabilityId === capability.id,
            customColor: viewCapability?.customColor,
          },
        };
      })
      .filter((n) => n !== null) as Node[];

    setNodes([...componentNodes, ...capabilityNodes]);
  }, [components, currentView, selectedNodeId, canvasCapabilities, capabilities, selectedCapabilityId]);

  useEffect(() => {
    if (!currentView) return;
    const componentIdsOnCanvas = currentView.components.map((vc) => vc.componentId);
    componentIdsOnCanvas.forEach((componentId) => {
      loadRealizationsByComponent(componentId);
    });
  }, [currentView?.id, currentView?.components.length, loadRealizationsByComponent]);

  // Build edges from relations and capability parent relationships
  React.useEffect(() => {
    const edgeType = currentView?.edgeType || 'default';

    // Build relation edges
    const relationEdges: Edge[] = relations.map((relation) => {
      const isSelected = selectedEdgeId === relation.id;
      const isTriggers = relation.relationType === 'Triggers';

      const sourceNode = nodes.find(n => n.id === relation.sourceComponentId);
      const targetNode = nodes.find(n => n.id === relation.targetComponentId);
      const { sourceHandle, targetHandle } = getBestHandles(sourceNode, targetNode);

      return {
        id: relation.id,
        source: relation.sourceComponentId,
        target: relation.targetComponentId,
        sourceHandle,
        targetHandle,
        label: relation.name || relation.relationType,
        type: edgeType,
        animated: isSelected,
        style: {
          stroke: isTriggers ? '#f97316' : '#3b82f6',
          strokeWidth: isSelected ? 3 : 2,
        },
        markerEnd: {
          type: MarkerType.ArrowClosed,
          color: isTriggers ? '#f97316' : '#3b82f6',
        },
        labelStyle: {
          fill: isTriggers ? '#f97316' : '#3b82f6',
          fontWeight: isSelected ? 700 : 500,
        },
        labelBgStyle: {
          fill: '#ffffff',
        },
      };
    });

    // Build parent edges from capabilities on canvas
    const canvasCapabilityIds = new Set(canvasCapabilities.map((cc) => cc.capabilityId));
    const parentEdges: Edge[] = canvasCapabilities
      .map((cc) => {
        const capability = capabilities.find((c) => c.id === cc.capabilityId);
        if (!capability || !capability.parentId) return null;

        // Only show edge if parent is also on canvas
        if (!canvasCapabilityIds.has(capability.parentId)) return null;

        const childNodeId = `cap-${capability.id}`;
        const parentNodeId = `cap-${capability.parentId}`;
        const edgeId = `parent-${capability.parentId}-${capability.id}`;
        const isSelected = selectedEdgeId === edgeId;

        const parentNode = nodes.find((n) => n.id === parentNodeId);
        const childNode = nodes.find((n) => n.id === childNodeId);
        const { sourceHandle, targetHandle } = getBestHandles(parentNode, childNode);

        return {
          id: edgeId,
          source: parentNodeId,
          target: childNodeId,
          sourceHandle,
          targetHandle,
          label: 'Parent',
          type: 'default' as const,
          animated: isSelected,
          style: {
            stroke: '#374151',
            strokeWidth: 3,
          },
          markerEnd: {
            type: MarkerType.ArrowClosed,
            color: '#374151',
          },
          labelStyle: {
            fill: '#374151',
            fontWeight: isSelected ? 700 : 600,
          },
          labelBgStyle: {
            fill: '#ffffff',
          },
        };
      })
      .filter((e) => e !== null) as Edge[];

    const visibleCapabilityIds = new Set(canvasCapabilities.map((cc) => cc.capabilityId));
    const componentIdsOnCanvas = new Set(
      currentView?.components.map((vc) => vc.componentId) || []
    );

    const shouldShowRealizationEdge = (realization: CapabilityRealization): boolean => {
      if (!componentIdsOnCanvas.has(realization.componentId)) return false;
      if (!visibleCapabilityIds.has(realization.capabilityId)) return false;

      if (realization.origin === 'Direct') {
        return true;
      }

      if (realization.origin === 'Inherited' && realization.sourceRealizationId) {
        const sourceRealization = capabilityRealizations.find(
          (r) => r.id === realization.sourceRealizationId
        );
        if (sourceRealization) {
          return !visibleCapabilityIds.has(sourceRealization.capabilityId);
        }
      }
      return false;
    };

    const realizationEdges = capabilityRealizations
      .filter(shouldShowRealizationEdge)
      .map((realization) => {
        const edgeId = `realization-${realization.id}`;
        const isSelected = selectedEdgeId === edgeId;
        const isInherited = realization.origin === 'Inherited';

        const sourceNodeId = realization.componentId;
        const targetNodeId = `cap-${realization.capabilityId}`;

        const sourceNode = nodes.find((n) => n.id === sourceNodeId);
        const targetNode = nodes.find((n) => n.id === targetNodeId);
        const { sourceHandle, targetHandle } = getBestHandles(sourceNode, targetNode);

        return {
          id: edgeId,
          source: sourceNodeId,
          target: targetNodeId,
          sourceHandle,
          targetHandle,
          label: isInherited ? 'Realizes (inherited)' : 'Realizes',
          type: 'default' as const,
          animated: isSelected,
          className: isInherited ? 'realization-edge inherited' : 'realization-edge',
          style: {
            stroke: '#10B981',
            strokeWidth: isSelected ? 3 : 2,
            strokeDasharray: '5,5',
            opacity: isInherited ? 0.6 : 1.0,
          },
          markerEnd: {
            type: MarkerType.ArrowClosed,
            color: '#10B981',
          },
          labelStyle: {
            fill: '#10B981',
            fontWeight: isSelected ? 700 : 500,
            opacity: isInherited ? 0.8 : 1.0,
          },
          labelBgStyle: {
            fill: '#ffffff',
          },
        };
      });

    setEdges([...relationEdges, ...parentEdges, ...realizationEdges]);
  }, [relations, selectedEdgeId, currentView?.edgeType, currentView?.components, nodes, canvasCapabilities, capabilities, capabilityRealizations]);

  const onNodesChange = useCallback(
    (changes: NodeChange[]) => {
      setNodes((nds) => applyNodeChanges(changes, nds));
    },
    []
  );

  const onEdgesChange = useCallback(
    (changes: EdgeChange[]) => {
      setEdges((eds) => applyEdgeChanges(changes, eds));
    },
    []
  );

  const onNodeClick = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      if (node.type === 'capability') {
        const capId = node.id.replace('cap-', '');
        selectCapability(capId);
        selectNode(null);
      } else {
        selectNode(node.id);
        selectCapability(null);
      }
    },
    [selectNode, selectCapability]
  );

  const onEdgeClick = useCallback(
    (_event: React.MouseEvent, edge: Edge) => {
      selectEdge(edge.id);
    },
    [selectEdge]
  );

  const onPaneClick = useCallback(() => {
    clearSelection();
    selectCapability(null);
    setNodeContextMenu(null);
    setEdgeContextMenu(null);
  }, [clearSelection, selectCapability]);

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

  const onNodeDragStop = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      if (node.type === 'capability') {
        const capId = node.id.replace('cap-', '');
        updateCapabilityPosition(capId, node.position.x, node.position.y);
      } else {
        updatePosition(node.id, node.position);
      }
    },
    [updatePosition, updateCapabilityPosition]
  );

  const linkSystemToCapability = useAppStore((state) => state.linkSystemToCapability);

  const onConnectHandler = useCallback(
    async (connection: Connection) => {
      if (!connection.source || !connection.target) return;

      const sourceIsCapability = connection.source.startsWith('cap-');
      const targetIsCapability = connection.target.startsWith('cap-');

      if (sourceIsCapability && targetIsCapability) {
        // Capability to capability connection - create parent relationship
        // Source is the parent, target is the child (React Flow inverted)
        const parentId = connection.target.replace('cap-', '');
        const childId = connection.source.replace('cap-', '');

        try {
          await changeCapabilityParent(childId, parentId);
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to create parent relationship';
          if (errorMessage.includes('depth') || errorMessage.includes('L5')) {
            toast.error('Cannot create this parent relationship: would result in hierarchy deeper than L4');
          }
        }
      } else if (!sourceIsCapability && !targetIsCapability) {
        // Component to component connection
        onConnect(connection.target, connection.source);
      } else {
        // Capability to component or component to capability - create realization
        const capabilityId = sourceIsCapability
          ? connection.source.replace('cap-', '')
          : connection.target.replace('cap-', '');
        const componentId = sourceIsCapability ? connection.target : connection.source;

        try {
          await linkSystemToCapability(capabilityId, {
            componentId,
            realizationLevel: 'Full',
          });
        } catch (error) {
          const errorMessage = error instanceof Error ? error.message : 'Failed to create realization';
          toast.error(errorMessage);
        }
      }
    },
    [onConnect, changeCapabilityParent, linkSystemToCapability]
  );

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'copy';
  }, []);

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault();

      const componentId = event.dataTransfer.getData('componentId');
      const capabilityId = event.dataTransfer.getData('capabilityId');

      if (!reactFlowInstance) return;

      // Get the position where the drop occurred
      const bounds = (event.target as HTMLElement).getBoundingClientRect();
      const position = reactFlowInstance.screenToFlowPosition({
        x: event.clientX - bounds.left,
        y: event.clientY - bounds.top,
      });

      if (componentId && onComponentDrop) {
        onComponentDrop(componentId, position.x, position.y);
      } else if (capabilityId) {
        addCapabilityToCanvas(capabilityId, position.x, position.y);
      }
    },
    [onComponentDrop, reactFlowInstance, addCapabilityToCanvas]
  );

  // Restore viewport state when view changes
  React.useEffect(() => {
    if (!currentView || !reactFlowInstance) return;

    const savedViewport = getViewportState(currentView.id);
    if (savedViewport) {
      // Restore saved viewport
      reactFlowInstance.setViewport(savedViewport, { duration: 300 });
      setIsFirstLoad(false);
    } else if (isFirstLoad && nodes.length > 0) {
      // On first load with no saved state, fit view
      setTimeout(() => {
        reactFlowInstance.fitView({ padding: 0.2, duration: 300 });
        setIsFirstLoad(false);
      }, 100);
    }
  }, [currentView?.id, reactFlowInstance, getViewportState, nodes.length, isFirstLoad]);

  // Save viewport state when canvas is moved or zoomed
  const onMoveEnd = useCallback(() => {
    if (!currentView || !reactFlowInstance) return;

    const viewport = reactFlowInstance.getViewport();
    saveViewportState(currentView.id, viewport);
  }, [currentView, reactFlowInstance, saveViewportState]);

  const handleDeleteConfirm = async () => {
    if (!deleteTarget) return;

    setIsDeleting(true);
    try {
      if (deleteTarget.type === 'component-from-view') {
        await removeComponentFromView(deleteTarget.id);
      } else if (deleteTarget.type === 'component-from-model') {
        await deleteComponent(deleteTarget.id);
      } else if (deleteTarget.type === 'relation-from-model') {
        await deleteRelation(deleteTarget.id);
      } else if (deleteTarget.type === 'capability-from-canvas') {
        removeCapabilityFromCanvas(deleteTarget.id);
      } else if (deleteTarget.type === 'capability-from-model') {
        await deleteCapability(deleteTarget.id);
      } else if (deleteTarget.type === 'parent-relation' && deleteTarget.childId) {
        await changeCapabilityParent(deleteTarget.childId, null);
      } else if (deleteTarget.type === 'realization') {
        await deleteRealization(deleteTarget.id);
      }
      setDeleteTarget(null);
    } catch (error) {
      console.error('Failed to delete:', error);
    } finally {
      setIsDeleting(false);
    }
  };

  const getNodeContextMenuItems = (): ContextMenuItem[] => {
    if (!nodeContextMenu) return [];

    if (nodeContextMenu.nodeType === 'capability') {
      return [
        {
          label: 'Remove from View',
          onClick: () => {
            removeCapabilityFromCanvas(nodeContextMenu.nodeId);
            setNodeContextMenu(null);
          },
        },
        {
          label: 'Delete from Model',
          onClick: () => {
            setDeleteTarget({
              type: 'capability-from-model',
              id: nodeContextMenu.nodeId,
              name: nodeContextMenu.nodeName,
            });
            setNodeContextMenu(null);
          },
          isDanger: true,
          ariaLabel: 'Delete capability from entire model',
        },
      ];
    }

    return [
      {
        label: 'Remove from View',
        onClick: () => {
          removeComponentFromView(nodeContextMenu.nodeId);
          setNodeContextMenu(null);
        },
      },
      {
        label: 'Delete from Model',
        onClick: () => {
          setDeleteTarget({
            type: 'component-from-model',
            id: nodeContextMenu.nodeId,
            name: nodeContextMenu.nodeName,
          });
          setNodeContextMenu(null);
        },
        isDanger: true,
        ariaLabel: 'Delete component from entire model',
      },
    ];
  };

  const getEdgeContextMenuItems = (): ContextMenuItem[] => {
    if (!edgeContextMenu) return [];

    if (edgeContextMenu.edgeType === 'parent') {
      const edgeId = edgeContextMenu.edgeId;
      const parentIdStart = edgeId.indexOf('-') + 1;
      const parentIdEnd = edgeId.indexOf('-', parentIdStart + 36);
      const childId = edgeId.substring(parentIdEnd + 1);

      return [
        {
          label: 'Remove Parent Relationship',
          onClick: () => {
            setDeleteTarget({
              type: 'parent-relation',
              id: edgeContextMenu.edgeId,
              name: 'Parent relationship',
              childId,
            });
            setEdgeContextMenu(null);
          },
          isDanger: true,
        },
      ];
    }

    if (edgeContextMenu.edgeType === 'realization' && edgeContextMenu.realizationId) {
      const realization = capabilityRealizations.find(
        (r) => r.id === edgeContextMenu.realizationId
      );
      const isInherited = realization?.origin === 'Inherited';

      if (isInherited) {
        return [];
      }

      return [
        {
          label: 'Delete Realization',
          onClick: () => {
            setDeleteTarget({
              type: 'realization',
              id: edgeContextMenu.realizationId!,
              name: edgeContextMenu.edgeName,
            });
            setEdgeContextMenu(null);
          },
          isDanger: true,
          ariaLabel: 'Delete realization link',
        },
      ];
    }

    return [
      {
        label: 'Delete from Model',
        onClick: () => {
          setDeleteTarget({
            type: 'relation-from-model',
            id: edgeContextMenu.edgeId,
            name: edgeContextMenu.edgeName,
          });
          setEdgeContextMenu(null);
        },
        isDanger: true,
        ariaLabel: 'Delete relation from entire model',
      },
    ];
  };

  // Expose method to center on a node
  useImperativeHandle(ref, () => ({
    centerOnNode: (nodeId: string) => {
      const node = nodes.find(n => n.id === nodeId);
      if (node && reactFlowInstance) {
        reactFlowInstance.setCenter(node.position.x + 75, node.position.y + 50, {
          zoom: 1,
          duration: 800,
        });
      }
    },
  }));

  return (
    <div
      className="canvas-container"
      onDragOver={onDragOver}
      onDrop={onDrop}
      data-testid="canvas-loaded"
    >
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
        onEdgeClick={onEdgeClick}
        onPaneClick={onPaneClick}
        onNodeDragStop={onNodeDragStop}
        onNodeContextMenu={onNodeContextMenu}
        onEdgeContextMenu={onEdgeContextMenu}
        onConnect={onConnectHandler}
        onMoveEnd={onMoveEnd}
        nodeTypes={nodeTypes}
        connectionMode={ConnectionMode.Loose}
        minZoom={0.1}
        maxZoom={2}
        defaultEdgeOptions={{
          type: 'default',
          animated: false,
        }}
      >
        <Background variant={BackgroundVariant.Dots} gap={16} size={1} />
        <Controls />
        <MiniMap
          nodeColor={(node) => {
            if (node.type === 'capability') {
              const capId = node.id.replace('cap-', '');
              return capId === selectedCapabilityId ? '#1f2937' : '#374151';
            }
            return node.id === selectedNodeId ? '#8b5cf6' : '#3b82f6';
          }}
          maskColor="rgba(0, 0, 0, 0.1)"
        />
      </ReactFlow>

      {nodeContextMenu && (
        <ContextMenu
          x={nodeContextMenu.x}
          y={nodeContextMenu.y}
          items={getNodeContextMenuItems()}
          onClose={() => setNodeContextMenu(null)}
        />
      )}

      {edgeContextMenu && (
        <ContextMenu
          x={edgeContextMenu.x}
          y={edgeContextMenu.y}
          items={getEdgeContextMenuItems()}
          onClose={() => setEdgeContextMenu(null)}
        />
      )}

      {deleteTarget && (
        <ConfirmationDialog
          title={
            deleteTarget.type === 'component-from-model'
              ? 'Delete Component from Model'
              : deleteTarget.type === 'capability-from-model'
              ? 'Delete Capability from Model'
              : deleteTarget.type === 'parent-relation'
              ? 'Remove Parent Relationship'
              : deleteTarget.type === 'realization'
              ? 'Delete Realization'
              : 'Delete Relation from Model'
          }
          message={
            deleteTarget.type === 'component-from-model'
              ? 'This will delete the component from the entire model, remove it from ALL views, and delete ALL relations involving this component.'
              : deleteTarget.type === 'capability-from-model'
              ? 'This will delete the capability from the entire model, remove it from ALL views, and affect any child capabilities.'
              : deleteTarget.type === 'parent-relation'
              ? 'This will remove the parent-child relationship. The child capability will become a top-level (L1) capability.'
              : deleteTarget.type === 'realization'
              ? 'This will remove the link between this capability and application. Any inherited realizations will also be removed.'
              : 'This will delete the relation from the entire model and remove it from ALL views.'
          }
          itemName={deleteTarget.name}
          confirmText={deleteTarget.type === 'parent-relation' ? 'Remove' : 'Delete'}
          cancelText="Cancel"
          onConfirm={handleDeleteConfirm}
          onCancel={() => setDeleteTarget(null)}
          isLoading={isDeleting}
        />
      )}
    </div>
  );
});

ComponentCanvasInner.displayName = 'ComponentCanvasInner';

export const ComponentCanvas = forwardRef<ComponentCanvasRef, ComponentCanvasProps>(
  (props, ref) => {
    return (
      <ReactFlowProvider>
        <ComponentCanvasInner {...props} ref={ref} />
      </ReactFlowProvider>
    );
  }
);

ComponentCanvas.displayName = 'ComponentCanvas';
