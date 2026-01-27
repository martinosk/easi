import { useState, useCallback, useMemo } from 'react';
import { EnterpriseArchHeader } from '../components/EnterpriseArchHeader';
import { EnterpriseArchContent } from '../components/EnterpriseArchContent';
import { CreateEnterpriseCapabilityModal } from '../components/CreateEnterpriseCapabilityModal';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import { MaturityAnalysisTab } from '../components/MaturityAnalysisTab';
import { StrategicFitTab } from '../components/StrategicFitTab';
import { TimeSuggestionsTab } from '../components/TimeSuggestionsTab';
import { MaturityGapDetailPanel } from '../components/MaturityGapDetailPanel';
import { useEnterpriseCapabilities } from '../hooks/useEnterpriseCapabilities';
import { useDomainCapabilityLinking } from '../hooks/useDomainCapabilityLinking';
import { getErrorMessage } from '../utils/errorMessages';
import type { EnterpriseCapability, CreateEnterpriseCapabilityRequest, EnterpriseCapabilityId, CapabilityLinkStatusResponse } from '../types';
import type { Capability } from '../../../api/types';
import { useUserStore } from '../../../store/userStore';
import './EnterpriseArchPage.css';

type TabType = 'capabilities' | 'maturity-analysis' | 'strategic-fit' | 'time-suggestions';

const TAB_CONFIG: { id: TabType; label: string }[] = [
  { id: 'capabilities', label: 'Enterprise Capabilities' },
  { id: 'maturity-analysis', label: 'Maturity Analysis' },
  { id: 'strategic-fit', label: 'Strategic Fit' },
  { id: 'time-suggestions', label: 'TIME Suggestions' },
];

function useEnterpriseArchPermissions() {
  const hasPermission = useUserStore((state) => state.hasPermission);
  return useMemo(() => ({
    canRead: hasPermission('enterprise-arch:read'),
    canWrite: hasPermission('enterprise-arch:write'),
    canDelete: hasPermission('enterprise-arch:delete'),
  }), [hasPermission]);
}

interface TabNavigationProps {
  activeTab: TabType;
  onTabChange: (tab: TabType) => void;
}

function TabNavigation({ activeTab, onTabChange }: TabNavigationProps) {
  return (
    <div className="tab-navigation">
      {TAB_CONFIG.map(({ id, label }) => (
        <button
          key={id}
          type="button"
          className={`tab-button ${activeTab === id ? 'active' : ''}`}
          onClick={() => onTabChange(id)}
        >
          {label}
        </button>
      ))}
    </div>
  );
}

interface TabContentProps {
  activeTab: TabType;
  maturityGapDetailId: EnterpriseCapabilityId | null;
  onViewMaturityGapDetail: (id: EnterpriseCapabilityId) => void;
  onBackFromMaturityGapDetail: () => void;
  isLoading: boolean;
  error: string | null;
  capabilities: EnterpriseCapability[];
  selectedCapability: EnterpriseCapability | null;
  canWrite: boolean;
  onSelect: (capability: EnterpriseCapability) => void;
  onDelete: (capability: EnterpriseCapability) => void;
  onCreateNew: () => void;
  isDockPanelOpen: boolean;
  domainCapabilities: Capability[];
  linkStatuses: Map<string, CapabilityLinkStatusResponse>;
  isLoadingDomainCapabilities: boolean;
  onCloseDockPanel: () => void;
  onLinkCapability: (enterpriseCapabilityId: EnterpriseCapabilityId, domainCapability: Capability) => Promise<void>;
}

function TabContent({ activeTab, maturityGapDetailId, onViewMaturityGapDetail, onBackFromMaturityGapDetail, ...contentProps }: TabContentProps) {
  if (activeTab === 'maturity-analysis') {
    if (maturityGapDetailId) {
      return <MaturityGapDetailPanel enterpriseCapabilityId={maturityGapDetailId} onBack={onBackFromMaturityGapDetail} />;
    }
    return <MaturityAnalysisTab onViewDetail={onViewMaturityGapDetail} />;
  }
  if (activeTab === 'strategic-fit') return <StrategicFitTab />;
  if (activeTab === 'time-suggestions') return <TimeSuggestionsTab />;
  return <EnterpriseArchContent {...contentProps} />;
}

function useDeleteCapabilityDialog(
  selectedCapabilityId: EnterpriseCapabilityId | null,
  setSelectedCapabilityId: (id: EnterpriseCapabilityId | null) => void,
  deleteCapability: (id: EnterpriseCapabilityId, name: string) => Promise<void>
) {
  const [capabilityToDelete, setCapabilityToDelete] = useState<EnterpriseCapability | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  const handleDeleteClick = useCallback((capability: EnterpriseCapability) => {
    setCapabilityToDelete(capability);
    setDeleteError(null);
  }, []);

  const handleConfirmDelete = useCallback(async () => {
    if (!capabilityToDelete) return;
    try {
      setDeleteError(null);
      await deleteCapability(capabilityToDelete.id, capabilityToDelete.name);
      if (selectedCapabilityId === capabilityToDelete.id) {
        setSelectedCapabilityId(null);
      }
      setCapabilityToDelete(null);
    } catch (err) {
      setDeleteError(getErrorMessage(err, 'Failed to delete capability'));
    }
  }, [capabilityToDelete, selectedCapabilityId, setSelectedCapabilityId, deleteCapability]);

  const handleCancelDelete = useCallback(() => {
    setCapabilityToDelete(null);
    setDeleteError(null);
  }, []);

  return { capabilityToDelete, deleteError, handleDeleteClick, handleConfirmDelete, handleCancelDelete };
}

export function EnterpriseArchPage() {
  const [activeTab, setActiveTab] = useState<TabType>('capabilities');
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedCapabilityId, setSelectedCapabilityId] = useState<EnterpriseCapabilityId | null>(null);
  const [isDockPanelOpen, setIsDockPanelOpen] = useState(false);
  const [maturityGapDetailId, setMaturityGapDetailId] = useState<EnterpriseCapabilityId | null>(null);

  const { canRead, canWrite } = useEnterpriseArchPermissions();
  const { capabilities, isLoading, error, createCapability, deleteCapability } = useEnterpriseCapabilities();
  const { capabilityToDelete, deleteError, handleDeleteClick, handleConfirmDelete, handleCancelDelete } =
    useDeleteCapabilityDialog(selectedCapabilityId, setSelectedCapabilityId, deleteCapability);

  const selectedCapability = useMemo(
    () => capabilities.find((c) => c.id === selectedCapabilityId) || null,
    [capabilities, selectedCapabilityId]
  );

  const { domainCapabilities, linkStatuses, isLoading: isLoadingDomainCapabilities, linkCapability } =
    useDomainCapabilityLinking(isDockPanelOpen);

  const handleCreateCapability = useCallback(async (request: CreateEnterpriseCapabilityRequest) => {
    await createCapability(request);
    setIsModalOpen(false);
  }, [createCapability]);

  const handleSelectCapability = useCallback((capability: EnterpriseCapability) => {
    setSelectedCapabilityId(capability.id === selectedCapabilityId ? null : capability.id);
  }, [selectedCapabilityId]);

  const handleLinkCapability = useCallback(
    async (enterpriseCapabilityId: EnterpriseCapabilityId, domainCapability: Capability) => {
      await linkCapability(enterpriseCapabilityId, domainCapability);
    },
    [linkCapability]
  );

  const handleTabChange = useCallback((tab: TabType) => {
    setActiveTab(tab);
    setMaturityGapDetailId(null);
  }, []);

  if (!canRead) {
    return (
      <div className="enterprise-arch-page">
        <div className="enterprise-arch-container">
          <div className="error-message">You do not have permission to view enterprise architecture.</div>
        </div>
      </div>
    );
  }

  return (
    <div className="enterprise-arch-page">
      <div className="enterprise-arch-container">
        <EnterpriseArchHeader
          canWrite={canWrite}
          onCreateNew={() => setIsModalOpen(true)}
          isDockPanelOpen={isDockPanelOpen}
          onToggleDockPanel={() => setIsDockPanelOpen((prev) => !prev)}
          activeTab={activeTab}
          showTabActions={activeTab === 'capabilities'}
        />
        <TabNavigation activeTab={activeTab} onTabChange={handleTabChange} />
        <TabContent
          activeTab={activeTab}
          maturityGapDetailId={maturityGapDetailId}
          onViewMaturityGapDetail={setMaturityGapDetailId}
          onBackFromMaturityGapDetail={() => setMaturityGapDetailId(null)}
          isLoading={isLoading}
          error={error?.message || null}
          capabilities={capabilities}
          selectedCapability={selectedCapability}
          canWrite={canWrite}
          onSelect={handleSelectCapability}
          onDelete={handleDeleteClick}
          onCreateNew={() => setIsModalOpen(true)}
          isDockPanelOpen={isDockPanelOpen}
          domainCapabilities={domainCapabilities}
          linkStatuses={linkStatuses}
          isLoadingDomainCapabilities={isLoadingDomainCapabilities}
          onCloseDockPanel={() => setIsDockPanelOpen(false)}
          onLinkCapability={handleLinkCapability}
        />
      </div>
      {canWrite && (
        <CreateEnterpriseCapabilityModal
          isOpen={isModalOpen}
          onClose={() => setIsModalOpen(false)}
          onSubmit={handleCreateCapability}
        />
      )}
      {capabilityToDelete && (
        <ConfirmationDialog
          title="Delete Enterprise Capability"
          message={`Are you sure you want to delete "${capabilityToDelete.name}"?`}
          confirmText="Delete"
          cancelText="Cancel"
          onConfirm={handleConfirmDelete}
          onCancel={handleCancelDelete}
          error={deleteError}
        />
      )}
    </div>
  );
}
