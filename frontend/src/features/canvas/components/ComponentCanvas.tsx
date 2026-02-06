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
import { OriginEntityNode } from '../../../components/canvas/OriginEntityNode';
import { useCanvasNodes } from '../hooks/useCanvasNodes';
import { useCanvasEdges } from '../hooks/useCanvasEdges';
import { useCanvasSelection } from '../hooks/useCanvasSelection';
import { useCanvasViewport } from '../hooks/useCanvasViewport';
import { useCanvasDragDrop } from '../hooks/useCanvasDragDrop';
import { useCanvasConnection } from '../hooks/useCanvasConnection';
import { useContextMenu } from '../hooks/useContextMenu';
import { useDeleteConfirmation } from '../hooks/useDeleteConfirmation';
import { useBulkOperations } from '../hooks/useBulkOperations';
import { AutoLayoutButton } from './AutoLayoutButton';
import { NodeContextMenu } from './context-menus/NodeContextMenu';
import { EdgeContextMenu } from './context-menus/EdgeContextMenu';
import { MultiSelectContextMenu } from './context-menus/MultiSelectContextMenu';
import { DeleteConfirmationWrapper } from './dialogs/DeleteConfirmationWrapper';
import { BulkConfirmationDialog } from './dialogs/BulkConfirmationDialog';

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
  originEntity: OriginEntityNode,
};

const multiSelectionKeys = ['Meta', 'Control', 'Shift'];

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

  const [internalNodes, setInternalNodes] = React.useState(nodes);

  const prevNodesRef = React.useRef<string>('');
  React.useEffect(() => {
    const nodesKey = nodes.map(n => `${n.id}:${n.position.x}:${n.position.y}:${n.data?.isSelected}:${n.data?.label}:${n.data?.customColor}:${n.data?.maturityValue}`).join('|');
    if (prevNodesRef.current !== nodesKey) {
      prevNodesRef.current = nodesKey;
      setInternalNodes(prev => {
        const selectedIds = new Set(prev.filter(n => n.selected).map(n => n.id));
        if (selectedIds.size === 0) return nodes;
        return nodes.map(n => selectedIds.has(n.id) ? { ...n, selected: true } : n);
      });
    }
  }, [nodes]);

  const {
    nodeContextMenu,
    edgeContextMenu,
    multiSelectMenu,
    onNodeContextMenu,
    onSelectionContextMenu,
    onEdgeContextMenu,
    closeMenus,
  } = useContextMenu(internalNodes);
  const {
    deleteTarget,
    isDeleting,
    setDeleteTarget,
    handleDeleteConfirm,
    handleDeleteCancel,
  } = useDeleteConfirmation();
  const {
    bulkOperation,
    isExecuting,
    result: bulkResult,
    setBulkOperation,
    handleBulkConfirm,
    handleBulkCancel,
  } = useBulkOperations();

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

  const handleMiniMapClick = useCallback((_event: React.MouseEvent, position: { x: number; y: number }) => {
    const { zoom } = reactFlowInstance.getViewport();
    reactFlowInstance.setCenter(position.x, position.y, { zoom, duration: 300 });
  }, [reactFlowInstance]);

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
        onSelectionContextMenu={onSelectionContextMenu}
        onEdgeContextMenu={onEdgeContextMenu}
        onConnect={onConnectHandler}
        onMoveEnd={onMoveEnd}
        nodeTypes={nodeTypes}
        connectionMode={ConnectionMode.Loose}
        multiSelectionKeyCode={multiSelectionKeys}
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
          pannable
          zoomable
          onClick={handleMiniMapClick}
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

      <AutoLayoutButton />

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

      <MultiSelectContextMenu
        menu={multiSelectMenu}
        onClose={closeMenus}
        onRequestBulkOperation={setBulkOperation}
      />

      <DeleteConfirmationWrapper
        deleteTarget={deleteTarget}
        isDeleting={isDeleting}
        onConfirm={handleDeleteConfirm}
        onCancel={handleDeleteCancel}
      />

      <BulkConfirmationDialog
        bulkOperation={bulkOperation}
        isExecuting={isExecuting}
        result={bulkResult}
        onConfirm={handleBulkConfirm}
        onCancel={handleBulkCancel}
      />
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
