# Spec 124: Architecture Canvas Auto-Layout

## Status
pending

## Overview
Enable automatic layout of architecture canvas views with intelligent positioning based on entity types and relationships. The auto-layout algorithm arranges capabilities hierarchically (L1-L4), positions application components based on their realization relationships, and places origin entities according to their relationships with applications. The architecture must support extensibility for future entity types.

## Business Context
As architecture models grow in complexity, manually positioning all entities becomes time-consuming and error-prone. Users need:
- Quick organization of canvas elements into logical layouts
- Hierarchical visualization of capabilities (L1 at top, L2 below, etc.)
- Application components positioned near the capabilities they realize
- Origin entities (acquired, vendor, internal team) positioned near their related applications
- Consistent layout that reflects the natural structure of the architecture

This feature builds upon the existing Dagre integration (Spec 021) but introduces **stratified layout** where different entity types occupy distinct vertical zones.

## Functional Requirements

### Auto-Layout Button
- [x] "Auto Layout" button appears in the canvas toolbar/controls
- [x] Button only enabled when current view has at least one entity
- [x] Clicking triggers automatic repositioning of all entities in the current view
- [x] Layout preserves viewport zoom level
- [x] Loading indicator shows during calculation
- [x] Toast notification confirms completion or shows errors

### Layout Behavior

**Stratified Vertical Zones:**

The canvas is divided into vertical zones from top to bottom:

1. **Capability Zone** (top): L1 capabilities → L2 → L3 → L4
2. **Application Zone** (middle): Application components
3. **Origin Entity Zone** (bottom): Acquired entities, vendors, internal teams

**Within Each Zone:**

**Capability Zone:**
- L1 capabilities arranged horizontally at the top
- L2 capabilities positioned below their L1 parents
- L3 capabilities positioned below their L2 parents
- L4 capabilities positioned below their L3 parents
- Horizontal spacing keeps children near their parents
- Hierarchical layout algorithm (Dagre or similar) respects parent-child edges

**Application Zone:**
- Application components positioned based on:
  - Realization relationships to capabilities (components cluster near realized capabilities)
  - Component-to-component relationships (dependencies, integrations)
- Components with no capability realizations positioned at the start of the zone
- Horizontal ordering minimizes edge crossings

**Origin Entity Zone:**
- Origin entities (acquired, vendor, team) positioned based on:
  - Relationships to application components (acquired via, purchased from, built by)
- Entities cluster near their related applications
- Entities with no relationships positioned at the start of the zone
- Horizontal ordering minimizes edge crossings

**Cross-Zone Relationships:**
- Capability-to-application edges (realization)
- Application-to-origin-entity edges (origin relationships)
- These edges may span multiple zones but layout minimizes length and crossings

### Layout Algorithm Selection

**Recommended Library:** Use a hierarchical layout library compatible with React Flow:
- **Dagre** (already integrated): Good foundation but may need custom layer assignment
- **ELK (Eclipse Layout Kernel)** via `elkjs`: More sophisticated, supports layered graphs with constraints
- **Cytoscape.js layouts** via `cytoscape`: Alternative with multiple algorithms

**Algorithm Characteristics:**
- Hierarchical/layered layout (top-to-bottom direction)
- Support for custom layer/rank assignment per node
- Minimize edge crossings
- Respect edge directions
- Handle disconnected components gracefully

### User Experience

**Before Layout:**
- User sees scattered entities or manually positioned entities
- Relationships may be hard to follow due to edge crossings

**After Layout:**
- Clear visual hierarchy: capabilities at top, applications in middle, origin entities at bottom
- Within capabilities: L1 at top, lower levels cascade downward
- Applications cluster near the capabilities they realize
- Origin entities cluster near the applications they relate to
- Minimal edge crossings
- Viewport automatically adjusts to fit all entities (fitView)

**Manual Adjustments:**
- After auto-layout, users can manually drag entities to fine-tune positions
- Manual positions persist to backend (existing behavior)
- Re-running auto-layout resets all positions based on current relationships

## Technical Requirements

### Frontend Architecture

#### 1. Extensible Entity Type Registry

Create a **layout strategy pattern** to support future entity types:

```typescript
// frontend/src/utils/autoLayout.ts

export interface EntityLayoutMetadata {
  nodeId: string;
  entityType: EntityType; // 'capability' | 'component' | 'originEntity' | ...
  layer: number; // Vertical zone (0 = top, 1 = middle, 2 = bottom, etc.)
  sublayer?: number; // For hierarchical ordering within a zone (e.g., capability level)
  weight?: number; // For horizontal positioning within a layer
}

export type EntityType = 'capability' | 'component' | 'originEntity';

export interface LayoutStrategy {
  /**
   * Extracts layout metadata from a React Flow node
   */
  extractMetadata(node: Node): EntityLayoutMetadata;

  /**
   * Determines layer assignment for this entity type
   */
  getLayer(): number;
}
```

#### 2. Strategy Implementations

**CapabilityLayoutStrategy:**
```typescript
class CapabilityLayoutStrategy implements LayoutStrategy {
  getLayer(): number {
    return 0; // Top zone
  }

  extractMetadata(node: Node): EntityLayoutMetadata {
    const level = node.data?.level ?? 1; // L1, L2, L3, L4
    return {
      nodeId: node.id,
      entityType: 'capability',
      layer: 0,
      sublayer: level - 1, // L1 = 0, L2 = 1, L3 = 2, L4 = 3
    };
  }
}
```

**ComponentLayoutStrategy:**
```typescript
class ComponentLayoutStrategy implements LayoutStrategy {
  getLayer(): number {
    return 1; // Middle zone
  }

  extractMetadata(node: Node): EntityLayoutMetadata {
    return {
      nodeId: node.id,
      entityType: 'component',
      layer: 1,
    };
  }
}
```

**OriginEntityLayoutStrategy:**
```typescript
class OriginEntityLayoutStrategy implements LayoutStrategy {
  getLayer(): number {
    return 2; // Bottom zone
  }

  extractMetadata(node: Node): EntityLayoutMetadata {
    return {
      nodeId: node.id,
      entityType: 'originEntity',
      layer: 2,
    };
  }
}
```

#### 3. Strategy Registry

```typescript
// frontend/src/utils/autoLayout.ts

const LAYOUT_STRATEGIES = new Map<EntityType, LayoutStrategy>([
  ['capability', new CapabilityLayoutStrategy()],
  ['component', new ComponentLayoutStrategy()],
  ['originEntity', new OriginEntityLayoutStrategy()],
]);

function getStrategyForNode(node: Node): LayoutStrategy {
  // Detect entity type from node
  if (node.type === 'capability') return LAYOUT_STRATEGIES.get('capability')!;
  if (node.type === 'component') return LAYOUT_STRATEGIES.get('component')!;
  if (node.type === 'originEntity') return LAYOUT_STRATEGIES.get('originEntity')!;
  
  throw new Error(`No layout strategy for node type: ${node.type}`);
}
```

#### 4. Auto-Layout Function

```typescript
// frontend/src/utils/autoLayout.ts

export interface AutoLayoutOptions {
  nodeSpacing?: number; // Horizontal spacing between nodes
  layerSpacing?: number; // Vertical spacing between layers
  sublayerSpacing?: number; // Vertical spacing between sublayers (e.g., L1 to L2)
}

export function calculateAutoLayout(
  nodes: Node[],
  edges: Edge[],
  options: AutoLayoutOptions = {}
): Node[] {
  const {
    nodeSpacing = 120,
    layerSpacing = 200,
    sublayerSpacing = 100,
  } = options;

  // 1. Extract metadata for all nodes using strategies
  const metadataMap = new Map<string, EntityLayoutMetadata>();
  for (const node of nodes) {
    const strategy = getStrategyForNode(node);
    const metadata = strategy.extractMetadata(node);
    metadataMap.set(node.id, metadata);
  }

  // 2. Use hierarchical layout library (Dagre, ELK, etc.)
  //    Configure library to respect layer/sublayer assignments
  //    Minimize edge crossings within constraints
  
  // Example using Dagre:
  const graph = new dagre.graphlib.Graph();
  graph.setGraph({
    rankdir: 'TB',
    nodesep: nodeSpacing,
    ranksep: layerSpacing,
    ranker: 'tight-tree', // or 'network-simplex'
  });
  
  graph.setDefaultEdgeLabel(() => ({}));
  
  // Add nodes with rank constraints
  nodes.forEach((node) => {
    const metadata = metadataMap.get(node.id)!;
    const rank = metadata.layer * 10 + (metadata.sublayer ?? 0);
    
    graph.setNode(node.id, {
      width: 180,
      height: 80,
      rank, // Force layer assignment
    });
  });
  
  // Add edges
  edges.forEach((edge) => {
    graph.setEdge(edge.source, edge.target);
  });
  
  dagre.layout(graph);
  
  // 3. Extract positions and update nodes
  const layoutedNodes = nodes.map((node) => {
    const nodeWithPosition = graph.node(node.id);
    return {
      ...node,
      position: {
        x: nodeWithPosition.x - 90,
        y: nodeWithPosition.y - 40,
      },
    };
  });
  
  return layoutedNodes;
}
```

**Alternative: Using ELK for more control:**

```typescript
import ELK from 'elkjs/lib/elk.bundled.js';

const elk = new ELK();

export async function calculateAutoLayoutELK(
  nodes: Node[],
  edges: Edge[],
  options: AutoLayoutOptions = {}
): Promise<Node[]> {
  const { nodeSpacing = 120, layerSpacing = 200 } = options;
  
  // Extract metadata
  const metadataMap = new Map<string, EntityLayoutMetadata>();
  for (const node of nodes) {
    const strategy = getStrategyForNode(node);
    metadataMap.set(node.id, strategy.extractMetadata(node));
  }
  
  // Build ELK graph
  const elkNodes = nodes.map((node) => {
    const metadata = metadataMap.get(node.id)!;
    return {
      id: node.id,
      width: 180,
      height: 80,
      properties: {
        'org.eclipse.elk.layered.layering.layer': metadata.layer,
        'org.eclipse.elk.layered.crossingMinimization.semiInteractive': true,
      },
    };
  });
  
  const elkEdges = edges.map((edge) => ({
    id: edge.id,
    sources: [edge.source],
    targets: [edge.target],
  }));
  
  const graph = {
    id: 'root',
    layoutOptions: {
      'elk.algorithm': 'layered',
      'elk.direction': 'DOWN',
      'elk.spacing.nodeNode': nodeSpacing.toString(),
      'elk.layered.spacing.nodeNodeBetweenLayers': layerSpacing.toString(),
    },
    children: elkNodes,
    edges: elkEdges,
  };
  
  const layout = await elk.layout(graph);
  
  // Map positions back to React Flow nodes
  const layoutedNodes = nodes.map((node) => {
    const elkNode = layout.children?.find((n) => n.id === node.id);
    return {
      ...node,
      position: {
        x: elkNode?.x ?? node.position.x,
        y: elkNode?.y ?? node.position.y,
      },
    };
  });
  
  return layoutedNodes;
}
```

#### 5. Integration with Canvas

```typescript
// frontend/src/features/canvas/hooks/useAutoLayout.ts

import { useCallback, useState } from 'react';
import { useReactFlow } from '@xyflow/react';
import { calculateAutoLayout } from '../../../utils/autoLayout';
import { useCanvasLayoutContext } from '../context/CanvasLayoutContext';
import { toast } from 'react-hot-toast';

export function useAutoLayout() {
  const reactFlowInstance = useReactFlow();
  const { batchUpdatePositions } = useCanvasLayoutContext();
  const [isLayouting, setIsLayouting] = useState(false);
  
  const applyAutoLayout = useCallback(async () => {
    if (!reactFlowInstance) return;
    
    setIsLayouting(true);
    
    try {
      const nodes = reactFlowInstance.getNodes();
      const edges = reactFlowInstance.getEdges();
      
      if (nodes.length === 0) {
        toast.error('No entities to layout');
        return;
      }
      
      // Calculate layout
      const layoutedNodes = calculateAutoLayout(nodes, edges);
      
      // Prepare batch update
      const updates = layoutedNodes.map((node) => ({
        elementId: node.id,
        x: node.position.x,
        y: node.position.y,
      }));
      
      // Persist to backend
      await batchUpdatePositions(updates);
      
      // Fit view to show all nodes
      window.requestAnimationFrame(() => {
        reactFlowInstance.fitView({ padding: 0.2, duration: 800 });
      });
      
      toast.success('Layout applied successfully');
    } catch (error) {
      console.error('Auto-layout failed:', error);
      toast.error('Failed to apply layout');
    } finally {
      setIsLayouting(false);
    }
  }, [reactFlowInstance, batchUpdatePositions]);
  
  return { applyAutoLayout, isLayouting };
}
```

#### 6. UI Component

```tsx
// frontend/src/features/canvas/components/AutoLayoutButton.tsx

import React from 'react';
import { useAutoLayout } from '../hooks/useAutoLayout';

export function AutoLayoutButton() {
  const { applyAutoLayout, isLayouting } = useAutoLayout();
  
  return (
    <button
      onClick={applyAutoLayout}
      disabled={isLayouting}
      className="auto-layout-button"
      title="Auto Layout"
    >
      {isLayouting ? (
        <>
          <span className="spinner" />
          Layouting...
        </>
      ) : (
        <>
          <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
            <path d="M2 2h12v2H2V2zm0 4h12v2H2V6zm0 4h12v2H2v-2z" />
          </svg>
          Auto Layout
        </>
      )}
    </button>
  );
}
```

### Extensibility for Future Entity Types

**Adding a New Entity Type (Example: "Infrastructure"):**

1. Define the entity type:
```typescript
export type EntityType = 'capability' | 'component' | 'originEntity' | 'infrastructure';
```

2. Create a layout strategy:
```typescript
class InfrastructureLayoutStrategy implements LayoutStrategy {
  getLayer(): number {
    return 3; // New zone below origin entities
  }

  extractMetadata(node: Node): EntityLayoutMetadata {
    return {
      nodeId: node.id,
      entityType: 'infrastructure',
      layer: 3,
    };
  }
}
```

3. Register the strategy:
```typescript
LAYOUT_STRATEGIES.set('infrastructure', new InfrastructureLayoutStrategy());
```

4. Update node type detection in `getStrategyForNode`:
```typescript
if (node.type === 'infrastructure') return LAYOUT_STRATEGIES.get('infrastructure')!;
```

**No changes required to the core layout algorithm** — it automatically respects the new layer assignments.

### Dependencies

**Recommended Library: `elkjs`**

ELK provides more control over layered layouts with constraints:

```bash
npm install elkjs
```

**Alternative: Dagre (already installed)**

Dagre works but may require custom layer preprocessing. If choosing Dagre, use the `rank` property to enforce layer constraints.

**Type Definitions:**

```bash
npm install --save-dev @types/elkjs
# or
npm install --save-dev @types/dagre
```

### Performance Considerations

- **Layout calculation** should complete within 2 seconds for diagrams up to 200 entities
- Use `requestAnimationFrame` for position updates to avoid UI blocking
- Batch position updates to minimize API calls (use existing `batchUpdatePositions`)
- Consider Web Worker for large graphs (>500 entities) if needed in the future

## Testing Requirements

### Unit Tests

**Layout Strategy Tests:**
- Test metadata extraction for each entity type
- Test layer assignment for capabilities (L1-L4 → sublayers 0-3)
- Test layer assignment for components (layer 1)
- Test layer assignment for origin entities (layer 2)

**Auto-Layout Function Tests:**
- Test layout with only capabilities (verify hierarchical ordering)
- Test layout with only components (verify clustering)
- Test layout with mixed entities (verify stratified zones)
- Test edge crossing minimization
- Test disconnected components (no edges)
- Test empty graph (no nodes)

**Strategy Registry Tests:**
- Test strategy retrieval by entity type
- Test error handling for unknown entity types

### Integration Tests

**Canvas Integration:**
- Test auto-layout button triggers layout calculation
- Test positions persist to backend after layout
- Test viewport fits to show all entities after layout
- Test loading indicator displays during layout
- Test toast notifications (success/error)

**E2E Tests (Playwright):**
- User clicks "Auto Layout" button
- Canvas rearranges entities into zones
- Capabilities appear at top in hierarchical order
- Applications appear in middle near realized capabilities
- Origin entities appear at bottom near related applications
- Positions persist after page refresh

## User Experience Requirements

### Auto-Layout Button
- Button in canvas toolbar or React Flow Controls panel
- Icon: grid/layout icon (e.g., Material Icons: `auto_awesome_mosaic` or `account_tree`)
- Label: "Auto Layout"
- Disabled state when no entities in view
- Loading state with spinner during calculation

### Layout Quality Metrics
- Minimize edge crossings (especially within zones)
- Minimize total edge length
- Balance horizontal distribution (avoid clustering on one side)
- Respect visual hierarchy (top-to-bottom)

### Accessibility
- Button keyboard accessible (Tab, Enter/Space)
- Loading state announced to screen readers
- Toast notifications announced to screen readers

## Business Rules / Invariants

- Auto-layout only affects positions, not entity properties
- Auto-layout applies to all entities in the current view
- Manual positions are overwritten by auto-layout
- Re-running auto-layout recalculates from current relationships (not from previous layout)
- Auto-layout does not create or delete entities

## Migration & Compatibility

- No backend changes required (uses existing position update APIs)
- No data migration needed
- Existing manual layouts remain unchanged until user triggers auto-layout

## Open Questions

1. **Layout Library Choice:** Dagre (simple, already integrated) vs ELK (sophisticated, better layer control)?
   - **Recommendation:** Start with Dagre, migrate to ELK if layer control is insufficient

2. **Capability Parent-Child Layout:** Should capabilities use parent-child edges for layout, or just level-based layers?
   - **Recommendation:** Use both (level for vertical position, parent edges for horizontal clustering)

3. **Cross-Zone Edge Routing:** Should we use custom edge routing for capability-to-application and application-to-origin edges?
   - **Recommendation:** Let layout algorithm handle initially, consider custom routing in future iteration

4. **Layout Options UI:** Should users configure spacing/options, or use sensible defaults?
   - **Recommendation:** Use sensible defaults initially, add options panel in future spec if needed

## Checklist
- [ ] Specification reviewed and approved
- [ ] Layout strategy pattern implemented
- [ ] Entity type strategies implemented (capability, component, originEntity)
- [ ] Auto-layout function implemented with chosen library
- [ ] Auto-layout button UI component created
- [ ] Integration with canvas and position persistence
- [ ] Unit tests for strategies and layout function
- [ ] Integration tests for canvas behavior
- [ ] E2E tests for user workflow
- [ ] Performance validated (<2s for 200 entities)
- [ ] Documentation updated (frontend patterns, user guide)
- [ ] User sign-off

## References

- **Spec 021:** Canvas Layout Improvements (Dagre foundation)
- **React Flow Documentation:** https://reactflow.dev/
- **Dagre:** https://github.com/dagrejs/dagre
- **ELK (Eclipse Layout Kernel):** https://www.eclipse.org/elk/
- **elkjs:** https://github.com/kieler/elkjs
