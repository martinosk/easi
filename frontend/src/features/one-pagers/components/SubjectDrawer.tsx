import { Drawer } from '@mantine/core';
import { DetailPanelFailure, DetailPanelLoading } from '../../../components/shared/DetailPanelStatus';
import { CapabilityDetailsPanel } from '../../capabilities/components/CapabilityDetailsPanel';
import { ComponentDetailsPanel } from '../../components/components/ComponentDetailsPanel';
import { EnterpriseCapabilityDetailPanel } from '../../enterprise-architecture/components/EnterpriseCapabilityDetailPanel';
import { useEnterpriseCapability } from '../../enterprise-architecture/hooks/useEnterpriseCapabilities';
import type { EnterpriseCapabilityId } from '../../enterprise-architecture/types';
import { AcquiredEntityDetailsPanel } from '../../origin-entities/components/AcquiredEntityDetailsPanel';
import { InternalTeamDetailsPanel } from '../../origin-entities/components/InternalTeamDetailsPanel';
import { VendorDetailsPanel } from '../../origin-entities/components/VendorDetailsPanel';
import { subjectTypeLabel } from '../subjectTypes';
import type { OnePagerSubjectType } from '../types';

export interface SubjectDrawerProps {
  opened: boolean;
  onClose: () => void;
  subjectType: OnePagerSubjectType;
  subjectId: string;
}

interface SubjectPanelProps {
  subjectId: string;
  onClose: () => void;
}

function EnterpriseCapabilityPanelHost({ subjectId, onClose }: SubjectPanelProps) {
  const query = useEnterpriseCapability(subjectId as EnterpriseCapabilityId);

  if (query.isLoading) return <DetailPanelLoading />;
  if (!query.data) return <DetailPanelFailure message="Failed to load enterprise capability" />;

  return <EnterpriseCapabilityDetailPanel capability={query.data} onClose={onClose} />;
}

function SubjectDetailPanel({ subjectType, subjectId, onClose }: SubjectPanelProps & { subjectType: OnePagerSubjectType }) {
  switch (subjectType) {
    case 'capability':
      return <CapabilityDetailsPanel capabilityId={subjectId} />;
    case 'application':
      return <ComponentDetailsPanel componentId={subjectId} />;
    case 'enterprise-capability':
      return <EnterpriseCapabilityPanelHost subjectId={subjectId} onClose={onClose} />;
    case 'acquired-entity':
      return <AcquiredEntityDetailsPanel entityId={subjectId} />;
    case 'vendor':
      return <VendorDetailsPanel entityId={subjectId} />;
    case 'internal-team':
      return <InternalTeamDetailsPanel entityId={subjectId} />;
  }
}

export function SubjectDrawer({ opened, onClose, subjectType, subjectId }: SubjectDrawerProps) {
  return (
    <Drawer
      opened={opened}
      onClose={onClose}
      position="right"
      size="md"
      title={subjectTypeLabel(subjectType)}
      data-testid="one-pager-subject-drawer"
    >
      <SubjectDetailPanel subjectType={subjectType} subjectId={subjectId} onClose={onClose} />
    </Drawer>
  );
}
