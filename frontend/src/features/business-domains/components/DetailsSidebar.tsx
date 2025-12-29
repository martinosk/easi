import { useState, useCallback } from 'react';
import type { BusinessDomain, Capability, CapabilityId, ComponentId } from '../../../api/types';
import { useComponentDetails } from '../hooks/useComponentDetails';
import { ComponentDetailsContent } from '../../components/components/ComponentDetails';
import { EditComponentDialog } from '../../components/components/EditComponentDialog';
import { useCapabilities, useCapabilitiesByComponent } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../../components/hooks/useComponents';
import { StrategicImportanceSection } from './StrategicImportanceSection';

interface DetailsSidebarProps {
  selectedCapability: Capability | null;
  selectedComponentId: ComponentId | null;
  visualizedDomain: BusinessDomain | null;
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

interface CapabilityContentProps {
  capability: Capability;
  domain: BusinessDomain | null;
}

function CapabilityContent({ capability, domain }: CapabilityContentProps) {
  return (
    <div className="detail-panel">
      <div className="detail-header">
        <h3 className="detail-title">Capability Details</h3>
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

        {domain && (
          <StrategicImportanceSection
            domain={domain}
            capabilityId={capability.id as CapabilityId}
            capabilityName={capability.name}
          />
        )}
      </div>
    </div>
  );
}

interface ApplicationContentProps {
  componentId: ComponentId;
}

function ApplicationContent({ componentId }: ApplicationContentProps) {
  const [editDialogOpen, setEditDialogOpen] = useState(false);

  const { data: storeComponents = [] } = useComponents();
  const { data: capabilities = [] } = useCapabilities();
  const { data: componentRealizations = [] } = useCapabilitiesByComponent(componentId);

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
        </div>
        <div className="detail-content" style={{ textAlign: 'center', padding: '2rem' }}>
          <p style={{ color: 'var(--color-red-500)', margin: 0 }}>
            Failed to load application details
          </p>
        </div>
      </div>
    );
  }

  return (
    <>
      <ComponentDetailsContent
        component={component}
        realizations={componentRealizations}
        capabilities={capabilities}
        onEdit={handleEdit}
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
  visualizedDomain,
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
      {selectedCapability && <CapabilityContent capability={selectedCapability} domain={visualizedDomain} />}
      {selectedComponentId && !selectedCapability && (
        <ApplicationContent componentId={selectedComponentId} />
      )}
    </aside>
  );
}
