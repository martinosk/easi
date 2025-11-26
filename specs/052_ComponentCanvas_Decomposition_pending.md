# ComponentCanvas Decomposition

## Description
Break down the 868-line ComponentCanvas.tsx into smaller, focused components following single responsibility principle.

## Current State Analysis

### File: `/frontend/src/components/ComponentCanvas.tsx`
- **Lines**: 868
- **Responsibilities**:
  1. React Flow setup and configuration
  2. Node rendering (components and capabilities)
  3. Edge rendering (relations, parent edges, realization edges)
  4. Selection state management
  5. Context menu handling (nodes and edges)
  6. Delete confirmation dialogs
  7. Drag and drop handling
  8. Viewport state persistence
  9. Connection handling between nodes
  10. Node position updates

### Current Structure
```typescript
// Helper functions (lines 46-78)
- getBestHandles, getNodeCenter, angleToHandleIndex

// Main component (lines 80-853)
- ComponentCanvasInner with forwardRef
  - 20+ useAppStore selectors
  - 6 useState hooks for local state
  - 5 useCallback handlers
  - 3 useEffect hooks
  - Multiple inline render functions
```

## Target Structure

### Proposed Component Breakdown
```
features/canvas/
├── components/
│   ├── ComponentCanvas.tsx          # Main canvas container (orchestration only)
│   ├── CanvasFlow.tsx               # React Flow wrapper with configuration
│   ├── nodes/
│   │   ├── ComponentNode.tsx        # Component node renderer (exists)
│   │   ├── CapabilityNode.tsx       # Capability node renderer (exists)
│   │   └── index.ts
│   ├── edges/
│   │   ├── useEdgeBuilder.ts        # Hook to build edge arrays
│   │   └── edgeStyles.ts            # Edge styling configuration
│   ├── context-menus/
│   │   ├── NodeContextMenu.tsx      # Context menu for nodes
│   │   ├── EdgeContextMenu.tsx      # Context menu for edges
│   │   └── useContextMenu.ts        # Shared context menu logic
│   ├── dialogs/
│   │   └── DeleteConfirmationWrapper.tsx  # Delete confirmation handling
│   └── index.ts
├── hooks/
│   ├── useCanvasNodes.ts            # Node building logic
│   ├── useCanvasEdges.ts            # Edge building logic
│   ├── useCanvasSelection.ts        # Selection handling
│   ├── useCanvasViewport.ts         # Viewport persistence
│   ├── useCanvasDragDrop.ts         # Drag and drop handling
│   └── useCanvasConnection.ts       # Connection handling
└── utils/
    ├── handleCalculation.ts         # getBestHandles, etc.
    └── nodeUtils.ts                 # Node position utilities
```

## Requirements

### Phase 1: Extract Utility Functions
- [ ] Create `features/canvas/utils/handleCalculation.ts`
  - Move getBestHandles, getNodeCenter, angleToHandleIndex
  - Export as pure functions with proper types

### Phase 2: Extract Hooks
- [ ] Create `useCanvasNodes.ts` hook
  - Extract node building logic from useEffect
  - Return computed nodes array

- [ ] Create `useCanvasEdges.ts` hook
  - Extract edge building logic from useEffect
  - Return computed edges array

- [ ] Create `useCanvasSelection.ts` hook
  - Extract onNodeClick, onEdgeClick, onPaneClick handlers
  - Return selection handlers and state

- [ ] Create `useCanvasViewport.ts` hook
  - Extract viewport save/restore logic
  - Handle view change viewport restoration

- [ ] Create `useCanvasDragDrop.ts` hook
  - Extract onDragOver, onDrop handlers
  - Handle component and capability drops

- [ ] Create `useCanvasConnection.ts` hook
  - Extract onConnectHandler logic
  - Handle different connection types (capability-capability, component-component, mixed)

### Phase 3: Extract Context Menu Components
- [ ] Create `NodeContextMenu.tsx`
  - Extract getNodeContextMenuItems logic
  - Handle both component and capability context menus

- [ ] Create `EdgeContextMenu.tsx`
  - Extract getEdgeContextMenuItems logic
  - Handle relation, parent, and realization edge menus

- [ ] Create `useContextMenu.ts` hook
  - Shared state management for context menus
  - Position calculation
  - Close handling

### Phase 4: Extract Delete Confirmation
- [ ] Create `DeleteConfirmationWrapper.tsx`
  - Extract delete target state
  - Extract handleDeleteConfirm logic
  - Encapsulate ConfirmationDialog usage

### Phase 5: Simplify Main Component
- [ ] Refactor ComponentCanvas.tsx to:
  - Import and compose extracted hooks
  - Import and render extracted components
  - Focus only on React Flow orchestration
  - Target: Under 200 lines

## Example Refactored ComponentCanvas.tsx
```typescript
const ComponentCanvasInner = forwardRef<ComponentCanvasRef, ComponentCanvasProps>(
  ({ onConnect, onComponentDrop }, ref) => {
    const reactFlowInstance = useReactFlow();

    // Composed hooks
    const nodes = useCanvasNodes();
    const edges = useCanvasEdges(nodes);
    const { handleNodeClick, handleEdgeClick, handlePaneClick } = useCanvasSelection();
    const { handleDragOver, handleDrop } = useCanvasDragDrop(reactFlowInstance, onComponentDrop);
    const { handleConnect } = useCanvasConnection(onConnect);
    useCanvasViewport(reactFlowInstance);

    // Context menu state
    const {
      nodeContextMenu,
      edgeContextMenu,
      openNodeMenu,
      openEdgeMenu,
      closeMenus,
    } = useContextMenu();

    // Delete confirmation
    const {
      deleteTarget,
      isDeleting,
      confirmDelete,
      cancelDelete,
      requestDelete,
    } = useDeleteConfirmation();

    useImperativeHandle(ref, () => ({
      centerOnNode: (nodeId: string) => { /* ... */ },
    }));

    return (
      <div className="canvas-container" onDragOver={handleDragOver} onDrop={handleDrop}>
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodeClick={handleNodeClick}
          onEdgeClick={handleEdgeClick}
          onPaneClick={() => { handlePaneClick(); closeMenus(); }}
          onNodeContextMenu={openNodeMenu}
          onEdgeContextMenu={openEdgeMenu}
          onConnect={handleConnect}
          nodeTypes={nodeTypes}
          /* ... other props */
        >
          <Background />
          <Controls />
          <MiniMap />
        </ReactFlow>

        <NodeContextMenu
          menu={nodeContextMenu}
          onClose={closeMenus}
          onDelete={requestDelete}
        />
        <EdgeContextMenu
          menu={edgeContextMenu}
          onClose={closeMenus}
          onDelete={requestDelete}
        />
        <DeleteConfirmation
          target={deleteTarget}
          isDeleting={isDeleting}
          onConfirm={confirmDelete}
          onCancel={cancelDelete}
        />
      </div>
    );
  }
);
```

## Testing Strategy
- Extract hooks should have dedicated unit tests
- Context menu components can be tested in isolation
- Integration test remains on ComponentCanvas to ensure composition works
- Keep existing E2E tests for canvas functionality

## Checklist
- [ ] Specification ready
- [ ] Utility functions extracted
- [ ] Custom hooks created and tested
- [ ] Context menu components extracted
- [ ] Delete confirmation extracted
- [ ] Main component simplified
- [ ] All tests passing
- [ ] Documentation updated
- [ ] User sign-off
