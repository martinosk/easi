import { Stack } from '@mantine/core';
import type React from 'react';
import type { Capability, ComponentId } from '../../../api/types';
import type { DetailGroup } from '../../../components/shared/DetailGroups';
import { DetailsShell } from '../../../components/shared/DetailsShell';
import { AuditHistorySection } from '../../audit';
import { OnePagerActionButton } from '../../one-pagers/components/OnePagerActionButton';
import {
  CreatedField,
  DescriptionField,
  ExpertsSection,
  LevelField,
  MaturityField,
  NameField,
  OwnershipFields,
  StatusField,
  TagsSection,
} from './CapabilityFieldSections';
import { CapabilityRealizationsSection } from './CapabilityRealizationsSection';

const LAYOUT_KEY = 'capability-details-layout';

export interface CapabilityDetailsContentProps {
  capability: Capability;
  transition?: React.ReactNode;
  strategicImportance?: React.ReactNode;
  viewMembership?: React.ReactNode;
  onApplicationClick?: (componentId: ComponentId) => void;
}

function buildGroups({
  capability,
  transition,
  strategicImportance,
  onApplicationClick,
}: CapabilityDetailsContentProps): DetailGroup[] {
  return [
    {
      id: 'description',
      title: 'Description',
      content: (
        <Stack gap="sm">
          <DescriptionField capability={capability} />
          <LevelField capability={capability} />
          <StatusField capability={capability} />
        </Stack>
      ),
    },
    { id: 'transition', title: 'Transition', content: transition },
    {
      id: 'fitness',
      title: 'Fitness',
      content: (
        <Stack gap="sm">
          {strategicImportance}
          <MaturityField capability={capability} />
        </Stack>
      ),
    },
    {
      id: 'metadata',
      title: 'Metadata',
      content: (
        <Stack gap="sm">
          <OwnershipFields capability={capability} />
          <TagsSection capability={capability} />
          <ExpertsSection capability={capability} />
          <CreatedField capability={capability} />
        </Stack>
      ),
    },
    {
      id: 'realisations',
      title: 'Realising applications',
      content: <CapabilityRealizationsSection capability={capability} onApplicationClick={onApplicationClick} />,
    },
  ];
}

export const CapabilityDetailsContent: React.FC<CapabilityDetailsContentProps> = (props) => {
  const { capability, viewMembership } = props;

  return (
    <DetailsShell
      layoutKey={LAYOUT_KEY}
      heading={<NameField capability={capability} />}
      groups={buildGroups(props)}
      viewMembership={viewMembership}
      footer={
        <>
          <OnePagerActionButton subject={capability} subjectType="capability" subjectId={capability.id} />
          <AuditHistorySection aggregateId={capability.id} />
        </>
      }
    />
  );
};
