import { Drawer } from '@mantine/core';
import { CapabilityDetailsPanel } from '../../capabilities/components/CapabilityDetailsPanel';
import { ComponentDetailsPanel } from '../../components/components/ComponentDetailsPanel';
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

function SubjectDetailPanel({ subjectType, subjectId }: { subjectType: OnePagerSubjectType; subjectId: string }) {
  switch (subjectType) {
    case 'capability':
      return <CapabilityDetailsPanel capabilityId={subjectId} />;
    case 'application':
      return <ComponentDetailsPanel componentId={subjectId} />;
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
      <SubjectDetailPanel subjectType={subjectType} subjectId={subjectId} />
    </Drawer>
  );
}
