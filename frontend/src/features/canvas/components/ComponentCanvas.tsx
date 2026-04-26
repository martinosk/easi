import {
  applyNodeChanges,
  Background,
  BackgroundVariant,
  ConnectionMode,
  Controls,
  MiniMap,
  type NodeChange,
  type NodeTypes,
  ReactFlow,
  ReactFlowProvider,
  useReactFlow,
} from '@xyflow/react';
import React, { forwardRef, useCallback, useImperativeHandle, useState } from 'react';
import '@xyflow/react/dist/style.css';
import { CapabilityNode } from '../../../components/canvas/CapabilityNode';
import { ComponentNode } from '../../../components/canvas/ComponentNode';
import { OriginEntityNode } from '../../../components/canvas/OriginEntityNode';
import { useAppStore } from '../../../store/appStore';
import { useUserStore } from '../../../store/userStore';
import { InviteToEditDialog } from '../../edit-grants/components/InviteToEditDialog';
import { useCreateEditGrant } from '../../edit-grants/hooks/useEditGrants';
import { useBulkOperations } from '../hooks/useBulkOperations';
import { useCanvasConnection } from '../hooks/useCanvasConnection';
import { useCanvasDragDrop } from '../hooks/useCanvasDragDrop';
import { useCanvasEdges } from '../hooks/useCanvasEdges';
import { useCanvasNodes } from '../hooks/useCanvasNodes';
import { useCanvasSelection } from '../hooks/useCanvasSelection';
import { useCanvasViewport } from '../hooks/useCanvasViewport';
import { useContextMenu } from '../hooks/useContextMenu';
import { useCreateDynamicView } from '../hooks/useCreateDynamicView';
import { useDeleteConfirmation } from '../hooks/useDeleteConfirmation';
import { AutoLayoutButton } from './AutoLayoutButton';
import { EdgeContextMenu } from './context-menus/EdgeContextMenu';
import { MultiSelectContextMenu } from './context-menus/MultiSelectContextMenu';
import { type GenerateViewTarget, type InviteTarget, NodeContextMenu } from './context-menus/NodeContextMenu';
import { BulkConfirmationDialog } from './dialogs/BulkConfirmationDialog';
import { DeleteConfirmationWrapper } from './dialogs/DeleteConfirmationWrapper';
import { DynamicModeContainer } from './DynamicModeContainer';
import { withDynamicExpansion } from './withDynamicExpansion';

interface ComponentCanvasProps {
  onConnect: (source: string, target: string) => void;
  onComponentDrop?: (componentId: string, x: number, y: number) => void;
}

export interface ComponentCanvasRef {
  centerOnNode: (nodeId: string) => void;
}

const nodeTypes: NodeTypes = {
  component: withDynamicExpansion(ComponentNode),
  capability: withDynamicExpansion(CapabilityNode),
  originEntity: withDynamicExpansion(OriginEntityNode),
};

const multiSelectionKeys = ['Meta', 'Control', 'Shift'];

const ComponentCanvasInner = forwardRef<ComponentCanvasRef, ComponentCanvasProps>(
  ({ onConnect, onComponentDrop }, ref) => {
    const reactFlowInstance = useReactFlow();
    const selectedNodeId = useAppStore((state) => state.selectedNodeId);
    const selectedCapabilityId = useAppStore((state) => state.selectedCapabilityId);
    const hasPermission = useUserStore((state) => state.hasPermission);
    const canCreateView = hasPermission('views:write');
    const { create: createDynamicView } = useCreateDynamicView();

    const nodes = useCanvasNodes();
    const edges = useCanvasEdges(nodes);
    const { onNodeClick, onEdgeClick, onPaneClick, onNodeDragStop } = useCanvasSelection();
    const { onDragOver, onDrop } = useCanvasDragDrop(reactFlowInstance, onComponentDrop);
    const { onConnectHandler } = useCanvasConnection(onConnect);
    const { onMoveEnd } = useCanvasViewport(reactFlowInstance, nodes);

    const [internalNodes, setInternalNodes] = React.useState(nodes);

    const prevNodesRef = React.useRef<string>('');
    React.useEffect(() => {
      const nodesKey = nodes
        .map(
          (n) =>
            `${n.id}:${n.position.x}:${n.position.y}:${n.data?.isSelected}:${n.data?.label}:${n.data?.customColor}:${n.data?.maturityValue}`,
        )
        .join('|');
      if (prevNodesRef.current !== nodesKey) {
        prevNodesRef.current = nodesKey;
        setInternalNodes((prev) => {
          const selectedIds = new Set(prev.filter((n) => n.selected).map((n) => n.id));
          if (selectedIds.size === 0) return nodes;
          return nodes.map((n) => (selectedIds.has(n.id) ? { ...n, selected: true } : n));
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
    const { deleteTarget, isDeleting, setDeleteTarget, handleDeleteConfirm, handleDeleteCancel } =
      useDeleteConfirmation();
    const [inviteTarget, setInviteTarget] = useState<InviteTarget | null>(null);

    const handleGenerateView = useCallback(
      (target: GenerateViewTarget) => {
        void createDynamicView(target.entityRef, target.entityName);
      },
      [createDynamicView],
    );
    const createGrant = useCreateEditGrant();
    const {
      bulkOperation,
      isExecuting,
      result: bulkResult,
      setBulkOperation,
      handleBulkConfirm,
      handleBulkCancel,
    } = useBulkOperations();

    const onNodesChange = useCallback((changes: NodeChange[]) => {
      setInternalNodes((nds) => applyNodeChanges(changes, nds));
    }, []);

    const handlePaneClick = useCallback(() => {
      onPaneClick();
      closeMenus();
    }, [onPaneClick, closeMenus]);

    const handleMiniMapClick = useCallback(
      (_event: React.MouseEvent, position: { x: number; y: number }) => {
        const { zoom } = reactFlowInstance.getViewport();
        reactFlowInstance.setCenter(position.x, position.y, { zoom, duration: 300 });
      },
      [reactFlowInstance],
    );

    useImperativeHandle(
      ref,
      () => ({
        centerOnNode: (nodeId: string) => {
          const node = internalNodes.find((n) => n.id === nodeId);
          if (node && reactFlowInstance) {
            reactFlowInstance.setCenter(node.position.x + 75, node.position.y + 50, {
              zoom: 1,
              duration: 800,
            });
          }
        },
      }),
      [internalNodes, reactFlowInstance],
    );

    return (
      <div className="canvas-container" onDragOver={onDragOver} onDrop={onDrop} data-testid="canvas-loaded">
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

        <div className="canvas-toolbar">
          <AutoLayoutButton />
          <DynamicModeContainer />
        </div>

        <NodeContextMenu
          menu={nodeContextMenu}
          onClose={closeMenus}
          onRequestDelete={setDeleteTarget}
          onRequestInviteToEdit={setInviteTarget}
          onRequestGenerateView={handleGenerateView}
          canCreateView={canCreateView}
        />

        <EdgeContextMenu menu={edgeContextMenu} onClose={closeMenus} onRequestDelete={setDeleteTarget} />

        <MultiSelectContextMenu menu={multiSelectMenu} onClose={closeMenus} onRequestBulkOperation={setBulkOperation} />

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

        {inviteTarget && (
          <InviteToEditDialog
            isOpen={inviteTarget !== null}
            onClose={() => setInviteTarget(null)}
            onSubmit={async (request) => {
              await createGrant.mutateAsync(request);
            }}
            artifactType={inviteTarget.artifactType}
            artifactId={inviteTarget.id}
          />
        )}
      </div>
    );
  },
);

ComponentCanvasInner.displayName = 'ComponentCanvasInner';

export const ComponentCanvas = forwardRef<ComponentCanvasRef, ComponentCanvasProps>((props, ref) => {
  return (
    <ReactFlowProvider>
      <ComponentCanvasInner {...props} ref={ref} />
    </ReactFlowProvider>
  );
});

ComponentCanvas.displayName = 'ComponentCanvas';
