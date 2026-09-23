import { Stack } from '@mantine/core';
import type React from 'react';
import type { Capability, ComponentId } from '../../../api/types';
import { type DetailGroup, DetailGroups } from '../../../components/shared/DetailGroups';
import { useDetailGroupLayout } from '../../../components/shared/useDetailGroupLayout';
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
const DEFAULT_ORDER = ['description', 'transition', 'fitness', 'metadata', 'realisations', 'view'] as const;

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
  viewMembership,
  onApplicationClick,
}: CapabilityDetailsContentProps): DetailGroup[] {
  const groups: DetailGroup[] = [
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
  if (transition) groups.push({ id: 'transition', title: 'Transition', content: transition });
  if (viewMembership) groups.push({ id: 'view', title: 'In this view', content: viewMembership });
  return groups;
}

export const CapabilityDetailsContent: React.FC<CapabilityDetailsContentProps> = (props) => {
  const layout = useDetailGroupLayout(LAYOUT_KEY, DEFAULT_ORDER);
  const { capability } = props;

  return (
    <Stack gap="sm">
      <NameField capability={capability} />
      <DetailGroups groups={buildGroups(props)} layout={layout} />
      <OnePagerActionButton subject={capability} subjectType="capability" subjectId={capability.id} />
      <AuditHistorySection aggregateId={capability.id} />
    </Stack>
  );
};
