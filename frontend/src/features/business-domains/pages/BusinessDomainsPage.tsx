import { useState, useMemo, useCallback, useRef, useEffect } from 'react';
import toast from 'react-hot-toast';
import { DndContext, DragOverlay, useSensor, useSensors, PointerSensor, type DragEndEvent, type DragStartEvent } from '@dnd-kit/core';
import { arrayMove } from '@dnd-kit/sortable';
import { DomainList } from '../components/DomainList';
import { DomainForm } from '../components/DomainForm';
import { DomainGrid } from '../components/DomainGrid';
import { NestedCapabilityGrid } from '../components/NestedCapabilityGrid';
import { DepthSelector, type DepthLevel } from '../components/DepthSelector';
import { CapabilityExplorer } from '../components/CapabilityExplorer';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import { ContextMenu, type ContextMenuItem } from '../../../components/shared/ContextMenu';
import { useBusinessDomains } from '../hooks/useBusinessDomains';
import { useDomainCapabilities } from '../hooks/useDomainCapabilities';
import { useCapabilityTree } from '../hooks/useCapabilityTree';
import { useGridPositions } from '../hooks/useGridPositions';
import type { BusinessDomain, Capability, CapabilityId } from '../../../api/types';
import '../components/visualization.css';

type DialogMode = 'create' | 'edit' | null;

interface DomainContextMenuState {
  x: number;
  y: number;
  domain: BusinessDomain;
}

export function BusinessDomainsPage() {
  const { domains, isLoading, error, createDomain, updateDomain, deleteDomain } = useBusinessDomains();
  const { tree, isLoading: treeLoading } = useCapabilityTree();
  const [dialogMode, setDialogMode] = useState<DialogMode>(null);
  const [selectedDomain, setSelectedDomain] = useState<BusinessDomain | null>(null);
  const [domainToDelete, setDomainToDelete] = useState<BusinessDomain | null>(null);
  const [visualizedDomain, setVisualizedDomain] = useState<BusinessDomain | null>(null);
  const [selectedCapability, setSelectedCapability] = useState<Capability | null>(null);
  const [activeCapability, setActiveCapability] = useState<Capability | null>(null);
  const [contextMenu, setContextMenu] = useState<DomainContextMenuState | null>(null);
  const [depth, setDepth] = useState<DepthLevel>(1);
  const [isDomainsSidebarCollapsed, setIsDomainsSidebarCollapsed] = useState(false);
  const [isExplorerSidebarCollapsed, setIsExplorerSidebarCollapsed] = useState(false);
  const dialogRef = useRef<HTMLDialogElement>(null);

  const { positions, updatePosition } = useGridPositions(visualizedDomain?.id ?? null);

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    })
  );

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;

    if (dialogMode) {
      dialog.showModal();
    } else {
      dialog.close();
    }
  }, [dialogMode]);

  const allCapabilities = useMemo(() => {
    const flatten = (nodes: typeof tree): Capability[] => {
      return nodes.flatMap((node) => [node.capability, ...flatten(node.children)]);
    };
    return flatten(tree);
  }, [tree]);

  const capabilitiesLink = visualizedDomain?._links.capabilities;

  const {
    capabilities,
    isLoading: capabilitiesLoading,
    associateCapability,
    refetch: refetchCapabilities,
  } = useDomainCapabilities(capabilitiesLink);

  const assignedCapabilityIds = useMemo(
    () => new Set<CapabilityId>(capabilities.map((c) => c.id)),
    [capabilities]
  );

  const capabilitiesWithDescendants = useMemo(() => {
    if (capabilities.length === 0 || tree.length === 0) return capabilities;

    const assignedL1Ids = new Set(capabilities.filter((c) => c.level === 'L1').map((c) => c.id));
    const result: Capability[] = [];

    const collectDescendants = (nodes: typeof tree) => {
      for (const node of nodes) {
        if (assignedL1Ids.has(node.capability.id)) {
          const addAll = (n: typeof tree[0]) => {
            result.push(n.capability);
            n.children.forEach(addAll);
          };
          addAll(node);
        } else {
          collectDescendants(node.children);
        }
      }
    };

    collectDescendants(tree);
    return result;
  }, [capabilities, tree]);

  const handleCreateClick = () => {
    setSelectedDomain(null);
    setDialogMode('create');
  };

  const handleVisualizeClick = (domain: BusinessDomain) => {
    setVisualizedDomain(domain);
    setSelectedCapability(null);
  };

  const handleContextMenu = (e: React.MouseEvent, domain: BusinessDomain) => {
    setContextMenu({ x: e.clientX, y: e.clientY, domain });
  };

  const getContextMenuItems = (menu: DomainContextMenuState): ContextMenuItem[] => {
    const items: ContextMenuItem[] = [];

    if (menu.domain._links.update) {
      items.push({
        label: 'Edit',
        onClick: () => {
          setSelectedDomain(menu.domain);
          setDialogMode('edit');
        },
      });
    }

    const canDelete = menu.domain.capabilityCount === 0 && menu.domain._links.delete;
    if (canDelete) {
      items.push({
        label: 'Delete',
        onClick: () => setDomainToDelete(menu.domain),
        isDanger: true,
      });
    }

    return items;
  };

  const handleCapabilityClick = (capability: Capability) => {
    setSelectedCapability(capability);
  };

  const handleDragStart = useCallback((event: DragStartEvent) => {
    const capability = event.active.data.current?.capability as Capability | undefined;
    if (capability) {
      setActiveCapability(capability);
    }
  }, []);

  const handleDragEnd = useCallback(
    async (event: DragEndEvent) => {
      setActiveCapability(null);

      const { active, over } = event;
      if (!over || !visualizedDomain) return;

      const droppedOnId = over.id as string;
      const isDroppedOnGrid = droppedOnId === 'domain-grid-droppable' || droppedOnId === 'nested-grid-droppable';

      if (active.id !== over.id && !isDroppedOnGrid) {
        const l1Caps = capabilities.filter((c) => c.level === 'L1');
        const oldIndex = l1Caps.findIndex((c) => c.id === active.id);
        const newIndex = l1Caps.findIndex((c) => c.id === over.id);

        if (oldIndex !== -1 && newIndex !== -1) {
          const newOrder = arrayMove(l1Caps, oldIndex, newIndex);
          newOrder.forEach((cap, index) => {
            updatePosition(cap.id, index, 0);
          });
          return;
        }
      }

      const capability = active.data.current?.capability as Capability | undefined;
      if (!capability || capability.level !== 'L1') return;

      if (assignedCapabilityIds.has(capability.id)) return;

      try {
        await associateCapability(capability.id, capability);
        await refetchCapabilities();
        const currentCount = capabilities.filter((c) => c.level === 'L1').length;
        await updatePosition(capability.id, currentCount, 0);
      } catch (err) {
        console.error('Failed to assign capability:', err);
      }
    },
    [visualizedDomain, associateCapability, refetchCapabilities, assignedCapabilityIds, capabilities, updatePosition]
  );

  const handleFormSubmit = async (name: string, description: string) => {
    if (dialogMode === 'create') {
      await createDomain(name, description);
      toast.success('Domain created successfully');
    } else if (dialogMode === 'edit' && selectedDomain) {
      await updateDomain(selectedDomain, name, description);
      toast.success('Domain updated successfully');
    }
    setDialogMode(null);
    setSelectedDomain(null);
  };

  const handleFormCancel = () => {
    setDialogMode(null);
    setSelectedDomain(null);
  };

  const handleConfirmDelete = async () => {
    if (domainToDelete) {
      try {
        await deleteDomain(domainToDelete);
        toast.success('Domain deleted successfully');
        setDomainToDelete(null);
        if (visualizedDomain?.id === domainToDelete.id) {
          setVisualizedDomain(null);
        }
      } catch (err) {
        toast.error(err instanceof Error ? err.message : 'Failed to delete domain');
      }
    }
  };

  const handleCancelDelete = () => {
    setDomainToDelete(null);
  };

  if (isLoading && domains.length === 0) {
    return (
      <div className="page-container">
        <div className="loading-message">Loading business domains...</div>
      </div>
    );
  }

  if (error && domains.length === 0) {
    return (
      <div className="page-container">
        <div className="error-message" data-testid="domains-error">
          {error.message}
        </div>
      </div>
    );
  }

  return (
    <DndContext sensors={sensors} onDragStart={handleDragStart} onDragEnd={handleDragEnd}>
      <div className="business-domains-layout" data-testid="business-domains-page" style={{ display: 'flex', height: '100vh', position: 'relative' }}>
        <aside className={`collapsible-sidebar ${isDomainsSidebarCollapsed ? 'closed' : 'open'}`}>
          {!isDomainsSidebarCollapsed && (
            <div className="sidebar-content">
              <div className="sidebar-header">
                <h3>Business Domains</h3>
                <button
                  type="button"
                  className="sidebar-toggle-btn"
                  onClick={() => setIsDomainsSidebarCollapsed(true)}
                  aria-label="Collapse sidebar"
                >
                  ‹
                </button>
              </div>
              <div style={{ marginBottom: '1rem' }}>
                <button
                  type="button"
                  className="btn btn-primary"
                  onClick={handleCreateClick}
                  data-testid="create-domain-button"
                >
                  Create Domain
                </button>
              </div>
              <div className="sidebar-scrollable">
                <DomainList
                  domains={domains}
                  onVisualize={handleVisualizeClick}
                  onContextMenu={handleContextMenu}
                  selectedDomainId={visualizedDomain?.id}
                />
              </div>
            </div>
          )}
        </aside>

        {isDomainsSidebarCollapsed && (
          <button
            type="button"
            className="sidebar-toggle-btn-collapsed left"
            onClick={() => setIsDomainsSidebarCollapsed(false)}
            aria-label="Expand sidebar"
          >
            ›
          </button>
        )}

        <main className="business-domains-main" style={{ flex: 1, padding: '1rem', overflow: 'auto' }}>
          {!visualizedDomain ? (
            <div style={{ textAlign: 'center', marginTop: '4rem' }}>
              <h2>Grid Visualization</h2>
              <p style={{ color: '#6b7280', marginTop: '1rem' }}>
                Click a domain to see its capabilities
              </p>
            </div>
          ) : capabilitiesLoading ? (
            <div className="loading-message">Loading capabilities...</div>
          ) : (
            <div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
                <h2>{visualizedDomain.name}</h2>
                <DepthSelector value={depth} onChange={setDepth} />
              </div>
              {depth === 1 ? (
                <DomainGrid capabilities={capabilities} onCapabilityClick={handleCapabilityClick} positions={positions} />
              ) : (
                <NestedCapabilityGrid
                  capabilities={capabilitiesWithDescendants}
                  depth={depth}
                  onCapabilityClick={handleCapabilityClick}
                  positions={positions}
                />
              )}
            </div>
          )}
        </main>

        <aside className={`collapsible-sidebar right ${isExplorerSidebarCollapsed ? 'closed' : 'open narrow'}`}>
          {!isExplorerSidebarCollapsed && (
            <div className="sidebar-content">
              <div className="sidebar-header">
                <h3>Capability Explorer</h3>
                <button
                  type="button"
                  className="sidebar-toggle-btn"
                  onClick={() => setIsExplorerSidebarCollapsed(true)}
                  aria-label="Collapse sidebar"
                >
                  ›
                </button>
              </div>
              <p style={{ fontSize: '0.875rem', color: '#6b7280', marginBottom: '1rem' }}>
                {visualizedDomain
                  ? 'Drag L1 capabilities to the grid to assign them'
                  : 'Select a domain to visualize, then drag capabilities to assign them'}
              </p>
              <div className="sidebar-scrollable">
                <CapabilityExplorer
                  capabilities={allCapabilities}
                  assignedCapabilityIds={assignedCapabilityIds}
                  isLoading={treeLoading}
                />
              </div>
            </div>
          )}
        </aside>

        {isExplorerSidebarCollapsed && (
          <button
            type="button"
            className="sidebar-toggle-btn-collapsed right"
            onClick={() => setIsExplorerSidebarCollapsed(false)}
            aria-label="Expand sidebar"
          >
            ‹
          </button>
        )}

        {selectedCapability && (
          <aside style={{ width: '300px', borderLeft: '1px solid #e5e7eb', padding: '1rem', overflow: 'auto' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <h3>Capability Details</h3>
              <button
                type="button"
                onClick={() => setSelectedCapability(null)}
                style={{ background: 'none', border: 'none', cursor: 'pointer', fontSize: '1.5rem' }}
              >
                &times;
              </button>
            </div>
            <div style={{ marginTop: '1rem' }}>
              <p><strong>Name:</strong> {selectedCapability.name}</p>
              <p><strong>Level:</strong> {selectedCapability.level}</p>
              {selectedCapability.description && (
                <p><strong>Description:</strong> {selectedCapability.description}</p>
              )}
            </div>
          </aside>
        )}
      </div>

      <DragOverlay>
        {activeCapability && (
          <div
            style={{
              backgroundColor: '#3b82f6',
              color: 'white',
              padding: '0.75rem 1rem',
              borderRadius: '0.5rem',
              boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)',
              fontWeight: 500,
            }}
          >
            {activeCapability.name}
          </div>
        )}
      </DragOverlay>

      {contextMenu && (
        <ContextMenu
          x={contextMenu.x}
          y={contextMenu.y}
          items={getContextMenuItems(contextMenu)}
          onClose={() => setContextMenu(null)}
        />
      )}

      <dialog ref={dialogRef} className="dialog" onClose={handleFormCancel} data-testid="domain-dialog">
        <div className="dialog-content">
          <h2 className="dialog-title">{dialogMode === 'create' ? 'Create Domain' : 'Edit Domain'}</h2>
          <DomainForm
            mode={dialogMode || 'create'}
            domain={selectedDomain || undefined}
            onSubmit={handleFormSubmit}
            onCancel={handleFormCancel}
          />
        </div>
      </dialog>

      {domainToDelete && (
        <ConfirmationDialog
          title="Delete Domain"
          message={`Are you sure you want to delete "${domainToDelete.name}"?`}
          confirmText="Delete"
          cancelText="Cancel"
          onConfirm={handleConfirmDelete}
          onCancel={handleCancelDelete}
        />
      )}

    </DndContext>
  );
}
