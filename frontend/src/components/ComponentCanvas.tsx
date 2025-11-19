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
};

const getBestHandles = (
  sourceNode: Node | undefined,
  targetNode: Node | undefined
): { sourceHandle: string; targetHandle: string } => {
  if (!sourceNode || !targetNode) {
    return { sourceHandle: 'top', targetHandle: 'top' };
  }

  const sourceX = sourceNode.position.x + (sourceNode.width || 150) / 2;
  const sourceY = sourceNode.position.y + (sourceNode.height || 100) / 2;
  const targetX = targetNode.position.x + (targetNode.width || 150) / 2;
  const targetY = targetNode.position.y + (targetNode.height || 100) / 2;

  const dx = targetX - sourceX;
  const dy = targetY - sourceY;

  const angle = Math.atan2(dy, dx) * (180 / Math.PI);

  let sourceHandle = 'right';
  let targetHandle = 'left';

  if (angle >= -45 && angle < 45) {
    sourceHandle = 'right';
    targetHandle = 'left';
  } else if (angle >= 45 && angle < 135) {
    sourceHandle = 'bottom';
    targetHandle = 'top';
  } else if (angle >= 135 || angle < -135) {
    sourceHandle = 'left';
    targetHandle = 'right';
  } else {
    sourceHandle = 'top';
    targetHandle = 'bottom';
  }

  return { sourceHandle, targetHandle };
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

  const [nodes, setNodes] = React.useState<Node[]>([]);
  const [edges, setEdges] = React.useState<Edge[]>([]);
  const [isFirstLoad, setIsFirstLoad] = React.useState(true);
  const [nodeContextMenu, setNodeContextMenu] = useState<{
    x: number;
    y: number;
    nodeId: string;
    nodeName: string;
  } | null>(null);
  const [edgeContextMenu, setEdgeContextMenu] = useState<{
    x: number;
    y: number;
    edgeId: string;
    edgeName: string;
  } | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<{
    type: 'component-from-view' | 'component-from-model' | 'relation-from-model';
    id: string;
    name: string;
  } | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  // Build nodes from components and view positions
  React.useEffect(() => {
    if (!currentView) return;

    // Only show components that are in the current view
    const newNodes: Node[] = components
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
          : { x: 400, y: 300 }; // Default center position (shouldn't happen after filter)

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

    setNodes(newNodes);
  }, [components, currentView, selectedNodeId]);

  // Build edges from relations
  React.useEffect(() => {
    const edgeType = currentView?.edgeType || 'default';

    const newEdges: Edge[] = relations.map((relation) => {
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

    setEdges(newEdges);
  }, [relations, selectedEdgeId, currentView?.edgeType, nodes]);

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
      selectNode(node.id);
    },
    [selectNode]
  );

  const onEdgeClick = useCallback(
    (_event: React.MouseEvent, edge: Edge) => {
      selectEdge(edge.id);
    },
    [selectEdge]
  );

  const onPaneClick = useCallback(() => {
    clearSelection();
    setNodeContextMenu(null);
    setEdgeContextMenu(null);
  }, [clearSelection]);

  const onNodeContextMenu = useCallback(
    (event: React.MouseEvent, node: Node) => {
      event.preventDefault();
      const component = components.find(c => c.id === node.id);
      if (component) {
        setNodeContextMenu({
          x: event.clientX,
          y: event.clientY,
          nodeId: node.id,
          nodeName: component.name,
        });
      }
    },
    [components]
  );

  const onEdgeContextMenu = useCallback(
    (event: React.MouseEvent, edge: Edge) => {
      event.preventDefault();
      const relation = relations.find(r => r.id === edge.id);
      if (relation) {
        setEdgeContextMenu({
          x: event.clientX,
          y: event.clientY,
          edgeId: edge.id,
          edgeName: relation.name || relation.relationType,
        });
      }
    },
    [relations]
  );

  const onNodeDragStop = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      updatePosition(node.id, node.position);
    },
    [updatePosition]
  );

  const onConnectHandler = useCallback(
    (connection: Connection) => {
      if (connection.source && connection.target) {
        // Swap source and target because React Flow's connection.source/target
        // are inverted from our domain model (connection.target is where you start dragging)
        onConnect(connection.target, connection.source);
      }
    },
    [onConnect]
  );

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'copy';
  }, []);

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault();

      const componentId = event.dataTransfer.getData('componentId');
      if (!componentId || !onComponentDrop || !reactFlowInstance) return;

      // Get the position where the drop occurred
      const bounds = (event.target as HTMLElement).getBoundingClientRect();
      const position = reactFlowInstance.screenToFlowPosition({
        x: event.clientX - bounds.left,
        y: event.clientY - bounds.top,
      });

      onComponentDrop(componentId, position.x, position.y);
    },
    [onComponentDrop, reactFlowInstance]
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
              : 'Delete Relation from Model'
          }
          message={
            deleteTarget.type === 'component-from-model'
              ? 'This will delete the component from the entire model, remove it from ALL views, and delete ALL relations involving this component.'
              : 'This will delete the relation from the entire model and remove it from ALL views.'
          }
          itemName={deleteTarget.name}
          confirmText="Delete"
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
