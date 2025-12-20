import React, { useCallback, useImperativeHandle, forwardRef } from 'react';
import {
  ReactFlow,
  ReactFlowProvider,
  type NodeChange,
  applyNodeChanges,
  type NodeTypes,
  Background,
  Controls,
  MiniMap,
  BackgroundVariant,
  useReactFlow,
  ConnectionMode,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { useAppStore } from '../../../store/appStore';
import { CapabilityNode } from '../../../components/canvas/CapabilityNode';
import { ComponentNode } from '../../../components/canvas/ComponentNode';
import { useCanvasNodes } from '../hooks/useCanvasNodes';
import { useCanvasEdges } from '../hooks/useCanvasEdges';
import { useCanvasSelection } from '../hooks/useCanvasSelection';
import { useCanvasViewport } from '../hooks/useCanvasViewport';
import { useCanvasDragDrop } from '../hooks/useCanvasDragDrop';
import { useCanvasConnection } from '../hooks/useCanvasConnection';
import { useContextMenu } from '../hooks/useContextMenu';
import { useDeleteConfirmation } from '../hooks/useDeleteConfirmation';
import { NodeContextMenu } from './context-menus/NodeContextMenu';
import { EdgeContextMenu } from './context-menus/EdgeContextMenu';
import { DeleteConfirmationWrapper } from './dialogs/DeleteConfirmationWrapper';
import { CanvasLayoutProvider } from '../context/CanvasLayoutContext';

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

const ComponentCanvasInner = forwardRef<ComponentCanvasRef, ComponentCanvasProps>(
  ({ onConnect, onComponentDrop }, ref) => {
  const reactFlowInstance = useReactFlow();
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);

  const nodes = useCanvasNodes();
  const edges = useCanvasEdges(nodes);
  const { onNodeClick, onEdgeClick, onPaneClick, onNodeDragStop } = useCanvasSelection();
  const { onDragOver, onDrop } = useCanvasDragDrop(reactFlowInstance, onComponentDrop);
  const { onConnectHandler } = useCanvasConnection(onConnect);
  const { onMoveEnd } = useCanvasViewport(reactFlowInstance, nodes);
  const {
    nodeContextMenu,
    edgeContextMenu,
    onNodeContextMenu,
    onEdgeContextMenu,
    closeMenus,
  } = useContextMenu();
  const {
    deleteTarget,
    isDeleting,
    setDeleteTarget,
    handleDeleteConfirm,
    handleDeleteCancel,
  } = useDeleteConfirmation();

  const [internalNodes, setInternalNodes] = React.useState(nodes);

  const nodesRef = React.useRef(nodes);
  React.useEffect(() => {
    if (nodesRef.current !== nodes) {
      nodesRef.current = nodes;
      setInternalNodes(nodes);
    }
  }, [nodes]);

  const onNodesChange = useCallback(
    (changes: NodeChange[]) => {
      setInternalNodes((nds) => applyNodeChanges(changes, nds));
    },
    []
  );


  const handlePaneClick = useCallback(() => {
    onPaneClick();
    closeMenus();
  }, [onPaneClick, closeMenus]);

  useImperativeHandle(ref, () => ({
    centerOnNode: (nodeId: string) => {
      const node = internalNodes.find(n => n.id === nodeId);
      if (node && reactFlowInstance) {
        reactFlowInstance.setCenter(node.position.x + 75, node.position.y + 50, {
          zoom: 1,
          duration: 800,
        });
      }
    },
  }), [internalNodes, reactFlowInstance]);

  return (
    <div
      className="canvas-container"
      onDragOver={onDragOver}
      onDrop={onDrop}
      data-testid="canvas-loaded"
    >
      <ReactFlow
        nodes={internalNodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onNodeClick={onNodeClick}
        onEdgeClick={onEdgeClick}
        onPaneClick={handlePaneClick}
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

      <NodeContextMenu
        menu={nodeContextMenu}
        onClose={closeMenus}
        onRequestDelete={setDeleteTarget}
      />

      <EdgeContextMenu
        menu={edgeContextMenu}
        onClose={closeMenus}
        onRequestDelete={setDeleteTarget}
      />

      <DeleteConfirmationWrapper
        deleteTarget={deleteTarget}
        isDeleting={isDeleting}
        onConfirm={handleDeleteConfirm}
        onCancel={handleDeleteCancel}
      />
    </div>
  );
});

ComponentCanvasInner.displayName = 'ComponentCanvasInner';

export const ComponentCanvas = forwardRef<ComponentCanvasRef, ComponentCanvasProps>(
  (props, ref) => {
    return (
      <ReactFlowProvider>
        <CanvasLayoutProvider>
          <ComponentCanvasInner {...props} ref={ref} />
        </CanvasLayoutProvider>
      </ReactFlowProvider>
    );
  }
);

ComponentCanvas.displayName = 'ComponentCanvas';
