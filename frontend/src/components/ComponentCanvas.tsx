import React, { useCallback, useImperativeHandle, forwardRef } from 'react';
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
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { useAppStore } from '../store/appStore';

interface ComponentCanvasProps {
  onConnect: (source: string, target: string) => void;
}

export interface ComponentCanvasRef {
  centerOnNode: (nodeId: string) => void;
}

interface ComponentNodeData {
  label: string;
  description?: string;
  isSelected: boolean;
}

const ComponentNode: React.FC<{ data: ComponentNodeData }> = ({ data }) => {
  return (
    <div
      className={`component-node ${data.isSelected ? 'component-node-selected' : ''}`}
    >
      {/* Connection handles for creating relations - all handles are both source and target */}
      <Handle
        type="source"
        position={Position.Top}
        id="top-source"
        className="component-handle component-handle-top"
      />
      <Handle
        type="target"
        position={Position.Top}
        id="top-target"
        className="component-handle component-handle-top"
      />

      <Handle
        type="source"
        position={Position.Left}
        id="left-source"
        className="component-handle component-handle-left"
      />
      <Handle
        type="target"
        position={Position.Left}
        id="left-target"
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
        id="right-source"
        className="component-handle component-handle-right"
      />
      <Handle
        type="target"
        position={Position.Right}
        id="right-target"
        className="component-handle component-handle-right"
      />

      <Handle
        type="source"
        position={Position.Bottom}
        id="bottom-source"
        className="component-handle component-handle-bottom"
      />
      <Handle
        type="target"
        position={Position.Bottom}
        id="bottom-target"
        className="component-handle component-handle-bottom"
      />
    </div>
  );
};

const nodeTypes: NodeTypes = {
  component: ComponentNode,
};

const ComponentCanvasInner = forwardRef<ComponentCanvasRef, ComponentCanvasProps>(
  ({ onConnect }, ref) => {
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

  const [nodes, setNodes] = React.useState<Node[]>([]);
  const [edges, setEdges] = React.useState<Edge[]>([]);

  // Build nodes from components and view positions
  React.useEffect(() => {
    if (!currentView) return;

    const newNodes: Node[] = components.map((component) => {
      const viewComponent = currentView.components.find(
        (vc) => vc.componentId === component.id
      );

      const position = viewComponent
        ? { x: viewComponent.x, y: viewComponent.y }
        : { x: 400, y: 300 }; // Default center position

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
    const newEdges: Edge[] = relations.map((relation) => {
      const isSelected = selectedEdgeId === relation.id;
      const isTriggers = relation.relationType === 'Triggers';

      return {
        id: relation.id,
        source: relation.sourceComponentId,
        target: relation.targetComponentId,
        label: relation.name || relation.relationType,
        type: 'default',
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
  }, [relations, selectedEdgeId]);

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
  }, [clearSelection]);

  const onNodeDragStop = useCallback(
    (_event: React.MouseEvent, node: Node) => {
      updatePosition(node.id, node.position.x, node.position.y);
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
    <div className="canvas-container">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
        onEdgeClick={onEdgeClick}
        onPaneClick={onPaneClick}
        onNodeDragStop={onNodeDragStop}
        onConnect={onConnectHandler}
        nodeTypes={nodeTypes}
        fitView
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
