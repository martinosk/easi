import React, { useCallback, useImperativeHandle, forwardRef, useState } from 'react';
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
import toast from 'react-hot-toast';

interface ComponentCanvasProps {
  onConnect: (source: string, target: string) => void;
  onComponentDrop?: (componentId: string, x: number, y: number) => void;
}

export interface ComponentCanvasRef {
  centerOnNode: (nodeId: string) => void;
}

interface ComponentNodeData {
  label: string;
  description?: string;
  isSelected: boolean;
}

const ComponentNode: React.FC<{ data: ComponentNodeData; id: string }> = ({ data, id }) => {
  return (
    <div
      className={`component-node ${data.isSelected ? 'component-node-selected' : ''}`}
      data-component-id={id}
    >
      <Handle
        type="source"
        position={Position.Top}
        id="top"
        className="component-handle component-handle-top"
      />
      <Handle
        type="target"
        position={Position.Top}
        id="top"
        className="component-handle component-handle-top"
      />

      <Handle
        type="source"
        position={Position.Left}
        id="left"
        className="component-handle component-handle-left"
      />
      <Handle
        type="target"
        position={Position.Left}
        id="left"
        className="component-handle component-handle-left"
      />

      <div className="component-node-content">
        <div className="component-node-header">{data.label}</div>
        {data.description && (
          <div className="component-node-description">{data.description}</div>
        )}
      </div>

      <Handle
        type="source"
        position={Position.Right}
        id="right"
        className="component-handle component-handle-right"
      />
      <Handle
        type="target"
        position={Position.Right}
        id="right"
        className="component-handle component-handle-right"
      />

      <Handle
        type="source"
        position={Position.Bottom}
        id="bottom"
        className="component-handle component-handle-bottom"
      />
      <Handle
        type="target"
        position={Position.Bottom}
        id="bottom"
        className="component-handle component-handle-bottom"
      />
    </div>
  );
};

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
    edgeType: 'relation' | 'parent';
  } | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<{
    type: 'component-from-view' | 'component-from-model' | 'relation-from-model' | 'capability-from-canvas' | 'parent-relation';
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
          },
        };
      });

    // Build capability nodes from canvas capabilities
    const capabilityNodes: Node[] = canvasCapabilities
      .map((cc) => {
        const capability = capabilities.find((c) => c.id === cc.capabilityId);
        if (!capability) return null;

        return {
          id: `cap-${capability.id}`,
          type: 'capability' as const,
          position: { x: cc.x, y: cc.y },
          data: {
            label: capability.name,
            level: capability.level,
            maturityLevel: capability.maturityLevel,
            isSelected: selectedCapabilityId === capability.id,
          },
        };
      })
      .filter((n) => n !== null) as Node[];

    setNodes([...componentNodes, ...capabilityNodes]);
  }, [components, currentView, selectedNodeId, canvasCapabilities, capabilities, selectedCapabilityId]);

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

    setEdges([...relationEdges, ...parentEdges]);
  }, [relations, selectedEdgeId, currentView?.edgeType, nodes, canvasCapabilities, capabilities]);

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
    [relations]
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
      }
    },
    [onConnect, changeCapabilityParent]
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
      } else if (deleteTarget.type === 'parent-relation' && deleteTarget.childId) {
        await changeCapabilityParent(deleteTarget.childId, null);
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
          label: 'Remove from Canvas',
          onClick: () => {
            removeCapabilityFromCanvas(nodeContextMenu.nodeId);
            setNodeContextMenu(null);
          },
        },
      ];
    }

    return [
      {
        label: 'Delete from View',
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
              : deleteTarget.type === 'parent-relation'
              ? 'Remove Parent Relationship'
              : 'Delete Relation from Model'
          }
          message={
            deleteTarget.type === 'component-from-model'
              ? 'This will delete the component from the entire model, remove it from ALL views, and delete ALL relations involving this component.'
              : deleteTarget.type === 'parent-relation'
              ? 'This will remove the parent-child relationship. The child capability will become a top-level (L1) capability.'
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
