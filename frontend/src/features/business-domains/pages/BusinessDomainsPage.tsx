import { useState, useMemo, useCallback } from 'react';
import toast from 'react-hot-toast';
import { DndContext, DragOverlay, type DragEndEvent, type DragStartEvent } from '@dnd-kit/core';
import { DomainList } from '../components/DomainList';
import { DomainForm } from '../components/DomainForm';
import { DomainGrid } from '../components/DomainGrid';
import { CapabilityExplorer } from '../components/CapabilityExplorer';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import { ContextMenu, type ContextMenuItem } from '../../../components/shared/ContextMenu';
import { useBusinessDomains } from '../hooks/useBusinessDomains';
import { useDomainCapabilities } from '../hooks/useDomainCapabilities';
import { useCapabilityTree } from '../hooks/useCapabilityTree';
import type { BusinessDomain, Capability, CapabilityId } from '../../../api/types';

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

      if (!event.over || !visualizedDomain) return;

      const capability = event.active.data.current?.capability as Capability | undefined;
      if (!capability || capability.level !== 'L1') return;

      if (assignedCapabilityIds.has(capability.id)) return;

      try {
        await associateCapability(capability.id, capability);
        await refetchCapabilities();
      } catch (err) {
        console.error('Failed to assign capability:', err);
      }
    },
    [visualizedDomain, associateCapability, refetchCapabilities, assignedCapabilityIds]
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
    <DndContext onDragStart={handleDragStart} onDragEnd={handleDragEnd}>
      <div className="business-domains-layout" data-testid="business-domains-page" style={{ display: 'flex', height: '100vh' }}>
        <aside className="business-domains-sidebar" style={{ width: '320px', borderRight: '1px solid #e5e7eb', padding: '1rem', overflow: 'auto' }}>
          <div className="page-header" style={{ marginBottom: '1rem' }}>
            <h1 style={{ fontSize: '1.5rem', marginBottom: '0.5rem' }}>Business Domains</h1>
            <button
              type="button"
              className="btn btn-primary"
              onClick={handleCreateClick}
              data-testid="create-domain-button"
            >
              Create Domain
            </button>
          </div>

          <DomainList
            domains={domains}
            onVisualize={handleVisualizeClick}
            onContextMenu={handleContextMenu}
            selectedDomainId={visualizedDomain?.id}
          />
        </aside>

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
              <h2 style={{ marginBottom: '1rem' }}>{visualizedDomain.name}</h2>
              <DomainGrid capabilities={capabilities} onCapabilityClick={handleCapabilityClick} />
            </div>
          )}
        </main>

        <aside className="capability-explorer-sidebar" style={{ width: '300px', borderLeft: '1px solid #e5e7eb', padding: '1rem', overflow: 'auto' }}>
          <h3 style={{ marginBottom: '1rem' }}>Capability Explorer</h3>
          <p style={{ fontSize: '0.875rem', color: '#6b7280', marginBottom: '1rem' }}>
            {visualizedDomain
              ? 'Drag L1 capabilities to the grid to assign them'
              : 'Select a domain to visualize, then drag capabilities to assign them'}
          </p>
          <CapabilityExplorer
            capabilities={allCapabilities}
            assignedCapabilityIds={assignedCapabilityIds}
            isLoading={treeLoading}
          />
        </aside>

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

      {dialogMode && (
        <dialog open className="dialog" data-testid="domain-dialog">
          <div className="dialog-content">
            <h2 className="dialog-title">{dialogMode === 'create' ? 'Create Domain' : 'Edit Domain'}</h2>
            <DomainForm
              mode={dialogMode}
              domain={selectedDomain || undefined}
              onSubmit={handleFormSubmit}
              onCancel={handleFormCancel}
            />
          </div>
        </dialog>
      )}

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
