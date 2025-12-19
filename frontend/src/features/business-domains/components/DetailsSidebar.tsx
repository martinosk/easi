import { useState, useCallback } from 'react';
import type { Capability, ComponentId } from '../../../api/types';
import { useAppStore } from '../../../store/appStore';
import { useComponentDetails } from '../hooks/useComponentDetails';
import { ComponentDetailsContent } from '../../components/components/ComponentDetails';
import { EditComponentDialog } from '../../components/components/EditComponentDialog';

interface DetailsSidebarProps {
  selectedCapability: Capability | null;
  selectedComponentId: ComponentId | null;
  onCloseCapability: () => void;
  onCloseApplication: () => void;
}

function EmptyState() {
  return (
    <div className="detail-panel">
      <div className="detail-header">
        <h3 className="detail-title">Details</h3>
      </div>
      <div className="detail-content" style={{ textAlign: 'center', padding: '2rem' }}>
        <p style={{ color: 'var(--color-gray-500)', margin: 0 }}>
          Select a capability or application to view details
        </p>
      </div>
    </div>
  );
}

function CapabilityContent({ capability, onClose }: { capability: Capability; onClose: () => void }) {
  return (
    <div className="detail-panel">
      <div className="detail-header">
        <h3 className="detail-title">Capability Details</h3>
        <button className="detail-close" onClick={onClose} aria-label="Close details">
          x
        </button>
      </div>
      <div className="detail-content">
        <div className="detail-field">
          <span className="detail-label">Name</span>
          <span className="detail-value">{capability.name}</span>
        </div>
        <div className="detail-field">
          <span className="detail-label">Level</span>
          <span className="detail-value">{capability.level}</span>
        </div>
        {capability.description && (
          <div className="detail-field">
            <span className="detail-label">Description</span>
            <span className="detail-value">{capability.description}</span>
          </div>
        )}
      </div>
    </div>
  );
}

interface ApplicationContentProps {
  componentId: ComponentId;
  onClose: () => void;
}

function ApplicationContent({ componentId, onClose }: ApplicationContentProps) {
  const [editDialogOpen, setEditDialogOpen] = useState(false);

  const storeComponents = useAppStore((state) => state.components);
  const capabilities = useAppStore((state) => state.capabilities);
  const capabilityRealizations = useAppStore((state) => state.capabilityRealizations);

  const componentFromStore = storeComponents.find((c) => c.id === componentId);
  const { component: componentFromApi, isLoading, error } = useComponentDetails(
    componentFromStore ? null : componentId
  );

  const component = componentFromStore || componentFromApi;

  const handleEdit = useCallback(() => {
    setEditDialogOpen(true);
  }, []);

  const handleCloseEditDialog = useCallback(() => {
    setEditDialogOpen(false);
  }, []);

  if (isLoading) {
    return (
      <div className="detail-panel">
        <div className="detail-header">
          <h3 className="detail-title">Application Details</h3>
          <button className="detail-close" onClick={onClose} aria-label="Close details">
            x
          </button>
        </div>
        <div className="detail-content" style={{ textAlign: 'center', padding: '2rem' }}>
          <p style={{ color: 'var(--color-gray-500)', margin: 0 }}>Loading...</p>
        </div>
      </div>
    );
  }

  if (error || !component) {
    return (
      <div className="detail-panel">
        <div className="detail-header">
          <h3 className="detail-title">Application Details</h3>
          <button className="detail-close" onClick={onClose} aria-label="Close details">
            x
          </button>
        </div>
        <div className="detail-content" style={{ textAlign: 'center', padding: '2rem' }}>
          <p style={{ color: 'var(--color-red-500)', margin: 0 }}>
            Failed to load application details
          </p>
        </div>
      </div>
    );
  }

  const componentRealizations = capabilityRealizations.filter((r) => r.componentId === component.id);

  return (
    <>
      <ComponentDetailsContent
        component={component}
        realizations={componentRealizations}
        capabilities={capabilities}
        onEdit={handleEdit}
        onClose={onClose}
      />
      <EditComponentDialog
        isOpen={editDialogOpen}
        onClose={handleCloseEditDialog}
        component={component}
      />
    </>
  );
}

export function DetailsSidebar({
  selectedCapability,
  selectedComponentId,
  onCloseCapability,
  onCloseApplication,
}: DetailsSidebarProps) {
  const hasSelection = selectedCapability || selectedComponentId;

  return (
    <aside
      style={{
        width: '100%',
        height: '100%',
        backgroundColor: 'white',
        overflow: 'auto',
      }}
    >
      {!hasSelection && <EmptyState />}
      {selectedCapability && (
        <CapabilityContent capability={selectedCapability} onClose={onCloseCapability} />
      )}
      {selectedComponentId && !selectedCapability && (
        <ApplicationContent componentId={selectedComponentId} onClose={onCloseApplication} />
      )}
    </aside>
  );
}
