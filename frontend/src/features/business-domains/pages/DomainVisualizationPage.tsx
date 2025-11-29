import { useState, useMemo, useCallback } from 'react';
import { DndContext, DragOverlay, type DragEndEvent, type DragStartEvent } from '@dnd-kit/core';
import { useBusinessDomains } from '../hooks/useBusinessDomains';
import { useDomainCapabilities } from '../hooks/useDomainCapabilities';
import { useCapabilityTree } from '../hooks/useCapabilityTree';
import { DomainFilter } from '../components/DomainFilter';
import { DomainGrid } from '../components/DomainGrid';
import { CapabilityExplorer } from '../components/CapabilityExplorer';
import type { BusinessDomainId, Capability, CapabilityId } from '../../../api/types';

interface DomainVisualizationPageProps {
  initialDomainId?: BusinessDomainId;
}

export function DomainVisualizationPage({ initialDomainId }: DomainVisualizationPageProps) {
  const [selectedDomainId, setSelectedDomainId] = useState<BusinessDomainId | null>(initialDomainId ?? null);
  const [selectedCapability, setSelectedCapability] = useState<Capability | null>(null);
  const [activeCapability, setActiveCapability] = useState<Capability | null>(null);
  const { domains, isLoading: domainsLoading } = useBusinessDomains();
  const { tree, isLoading: treeLoading } = useCapabilityTree();

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

  const handleDomainSelect = (domainId: BusinessDomainId | null) => {
    setSelectedDomainId(domainId);
    setSelectedCapability(null);
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

      if (!event.over || !selectedDomainId) return;

      const capability = event.active.data.current?.capability as Capability | undefined;
      if (!capability || capability.level !== 'L1') return;

      if (assignedCapabilityIds.has(capability.id)) return;

      try {
        await associateCapability(capability.id, capability);
        await refetchCapabilities();
      } catch (error) {
        console.error('Failed to assign capability:', error);
      }
    },
    [selectedDomainId, associateCapability, refetchCapabilities, assignedCapabilityIds]
  );

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
              <h1 style={{ marginBottom: '1rem' }}>{selectedDomain?.name}</h1>
              <DomainGrid capabilities={capabilities} onCapabilityClick={handleCapabilityClick} />
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
    </DndContext>
  );
}
