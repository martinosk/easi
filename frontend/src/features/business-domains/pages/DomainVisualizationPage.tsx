import { useState, useMemo, useCallback } from 'react';
import { DndContext, DragOverlay } from '@dnd-kit/core';
import { useBusinessDomains } from '../hooks/useBusinessDomains';
import { useDomainCapabilities } from '../hooks/useDomainCapabilities';
import { useCapabilityTree } from '../hooks/useCapabilityTree';
import { useGridPositions } from '../hooks/useGridPositions';
import { useDragHandlers, type PendingReassignment } from '../hooks/useDragHandlers';
import { DomainFilter } from '../components/DomainFilter';
import { DomainGrid } from '../components/DomainGrid';
import { NestedCapabilityGrid } from '../components/NestedCapabilityGrid';
import { DepthSelector } from '../components/DepthSelector';
import { CapabilityExplorer } from '../components/CapabilityExplorer';
import { ReassignConfirmDialog } from '../components/ReassignConfirmDialog';
import { usePersistedDepth } from '../hooks/usePersistedDepth';
import { apiClient } from '../../../api/client';
import type { BusinessDomainId, Capability, CapabilityId } from '../../../api/types';

interface DomainVisualizationPageProps {
  initialDomainId?: BusinessDomainId;
}

export function DomainVisualizationPage({ initialDomainId }: DomainVisualizationPageProps) {
  const [selectedDomainId, setSelectedDomainId] = useState<BusinessDomainId | null>(initialDomainId ?? null);
  const [selectedCapability, setSelectedCapability] = useState<Capability | null>(null);
  const [depth, setDepth] = usePersistedDepth();
  const [pendingReassignment, setPendingReassignment] = useState<PendingReassignment | null>(null);
  const [isReassigning, setIsReassigning] = useState(false);
  const { domains, isLoading: domainsLoading } = useBusinessDomains();
  const { tree, isLoading: treeLoading, refetch: refetchTree } = useCapabilityTree();
  const { positions, updatePosition } = useGridPositions(selectedDomainId);

  const allCapabilities = useMemo(() => {
    const flatten = (nodes: typeof tree): Capability[] => {
      return nodes.flatMap((node) => [node.capability, ...flatten(node.children)]);
    };
    return flatten(tree);
  }, [tree]);

  const selectedDomain = useMemo(
    () => domains.find((d) => d.id === selectedDomainId),
    [domains, selectedDomainId]
  );

  const capabilitiesLink = selectedDomain?._links.capabilities;

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

  const { activeCapability, handleDragStart, handleDragEnd } = useDragHandlers({
    domainId: selectedDomainId,
    capabilities,
    assignedCapabilityIds,
    positions,
    updatePosition,
    associateCapability,
    refetchCapabilities,
    allCapabilities,
    onReassignment: setPendingReassignment,
  });

  const handleDomainSelect = (domainId: BusinessDomainId | null) => {
    setSelectedDomainId(domainId);
    setSelectedCapability(null);
  };

  const handleCapabilityClick = (capability: Capability) => {
    setSelectedCapability(capability);
  };

  const handleConfirmReassign = useCallback(async () => {
    if (!pendingReassignment) return;

    setIsReassigning(true);
    try {
      await apiClient.changeCapabilityParent(pendingReassignment.capability.id, pendingReassignment.newParent.id);
      await refetchTree();
      await refetchCapabilities();
      setPendingReassignment(null);
    } catch (error) {
      console.error('Failed to reassign capability:', error);
    } finally {
      setIsReassigning(false);
    }
  }, [pendingReassignment, refetchTree, refetchCapabilities]);

  const handleCancelReassign = useCallback(() => {
    setPendingReassignment(null);
  }, []);

  if (domainsLoading) {
    return (
      <div className="page-container">
        <div className="loading-message">Loading domains...</div>
      </div>
    );
  }

  return (
    <DndContext onDragStart={handleDragStart} onDragEnd={handleDragEnd}>
      <div className="visualization-container" style={{ display: 'flex', height: '100vh' }}>
        <aside style={{ width: '250px', borderRight: '1px solid #e5e7eb', padding: '1rem' }}>
          <h2 style={{ marginBottom: '1rem' }}>Business Domains</h2>
          <DomainFilter
            domains={domains}
            selected={selectedDomainId}
            onSelect={handleDomainSelect}
          />
        </aside>

        <main style={{ flex: 1, padding: '1rem', overflow: 'auto' }}>
          {!selectedDomainId ? (
            <div style={{ textAlign: 'center', marginTop: '4rem' }}>
              <h1>Grid Visualization</h1>
              <p style={{ color: '#6b7280', marginTop: '1rem' }}>
                Select a domain from the left sidebar
              </p>
            </div>
          ) : capabilitiesLoading ? (
            <div className="loading-message">Loading capabilities...</div>
          ) : (
            <div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
                <h1>{selectedDomain?.name}</h1>
                <DepthSelector value={depth} onChange={setDepth} />
              </div>
              {depth === 1 ? (
                <DomainGrid capabilities={capabilities} onCapabilityClick={handleCapabilityClick} positions={positions} />
              ) : (
                <NestedCapabilityGrid
                  capabilities={capabilities}
                  depth={depth}
                  onCapabilityClick={handleCapabilityClick}
                  positions={positions}
                />
              )}
            </div>
          )}
        </main>

        <aside style={{ width: '300px', borderLeft: '1px solid #e5e7eb', padding: '1rem', overflow: 'auto' }}>
          <h3 style={{ marginBottom: '1rem' }}>Capability Explorer</h3>
          <p style={{ fontSize: '0.875rem', color: '#6b7280', marginBottom: '1rem' }}>
            Drag L1 capabilities to the grid to assign them to the selected domain
          </p>
          <CapabilityExplorer
            capabilities={allCapabilities}
            assignedCapabilityIds={assignedCapabilityIds}
            isLoading={treeLoading}
          />
        </aside>

        {selectedCapability && (
          <aside style={{ width: '300px', borderLeft: '1px solid #e5e7eb', padding: '1rem' }}>
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

      <ReassignConfirmDialog
        isOpen={pendingReassignment !== null}
        capability={pendingReassignment?.capability ?? null}
        newParent={pendingReassignment?.newParent ?? null}
        onConfirm={handleConfirmReassign}
        onCancel={handleCancelReassign}
        isLoading={isReassigning}
      />
    </DndContext>
  );
}
