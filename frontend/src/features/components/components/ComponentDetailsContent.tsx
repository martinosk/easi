import { Anchor, Badge, Group, Stack, Text } from '@mantine/core';
import React, { useState } from 'react';
import type { Capability, CapabilityRealization, Component } from '../../../api/types';
import { DetailField } from '../../../components/shared/DetailField';
import type { DetailGroup } from '../../../components/shared/DetailGroups';
import { DetailsShell } from '../../../components/shared/DetailsShell';
import { InlineTextField } from '../../../components/shared/InlineTextField';
import { componentDescriptionSchema, componentNameSchema } from '../../../lib/schemas';
import { hasLink } from '../../../utils/hateoas';
import { AuditHistorySection } from '../../audit';
import { OnePagerActionButton } from '../../one-pagers/components/OnePagerActionButton';
import { useUpdateComponent } from '../hooks/useComponents';
import { AddComponentExpertDialog } from './AddComponentExpertDialog';
import { ComponentContainmentSection } from './ComponentContainmentSection';
import { ComponentExpertsList } from './ComponentExpertsList';
import { ComponentFitScores } from './ComponentFitScores';
import { ComponentHostingSection } from './ComponentHostingSection';
import { ComponentOriginsSection } from './ComponentOriginsSection';
import { ComponentOwnershipSection } from './ComponentOwnershipSection';

const LAYOUT_KEY = 'application-details-layout';

interface ComponentDetailsContentProps {
  component: Component;
  realizations: CapabilityRealization[];
  capabilities: Capability[];
  viewMembership?: React.ReactNode;
}

const getLevelBadge = (level: string): string => {
  const badges: Record<string, string> = {
    Full: '100%',
    Partial: 'Partial',
    Planned: 'Planned',
  };
  return badges[level] || level;
};

const getCapabilityName = (capabilities: Capability[], capabilityId: string): string => {
  const cap = capabilities.find((c) => c.id === capabilityId);
  return cap ? `${cap.level}: ${cap.name}` : 'Unknown';
};

interface RealizationListProps {
  realizations: CapabilityRealization[];
  capabilities: Capability[];
  origin: 'Direct' | 'Inherited';
}

const RealizationListItems: React.FC<RealizationListProps> = ({ realizations, capabilities, origin }) => {
  const isInherited = origin === 'Inherited';
  return (
    <>
      {realizations.map((r) => (
        <Group key={r.id} gap="sm" wrap="nowrap" opacity={isInherited ? 0.7 : 1} justify="space-between">
          <Text size="sm">{getCapabilityName(capabilities, r.capabilityId)}</Text>
          <Group gap="xs" wrap="nowrap">
            <Badge color="green" variant="filled" size="sm">
              {getLevelBadge(r.realizationLevel)}
            </Badge>
            <Badge color={isInherited ? 'gray' : 'blue'} variant="light" size="sm">
              {origin.toLowerCase()}
            </Badge>
          </Group>
        </Group>
      ))}
    </>
  );
};

interface RealizationsFieldProps {
  realizations: CapabilityRealization[];
  capabilities: Capability[];
}

const RealizationsField: React.FC<RealizationsFieldProps> = ({ realizations, capabilities }) => {
  if (realizations.length === 0) {
    return (
      <Text size="sm" c="dimmed" fs="italic">
        realises no capability
      </Text>
    );
  }

  const directRealizations = realizations.filter((r) => r.origin === 'Direct');
  const inheritedRealizations = realizations.filter((r) => r.origin === 'Inherited');

  return (
    <Stack gap="sm">
      <RealizationListItems realizations={directRealizations} capabilities={capabilities} origin="Direct" />
      <RealizationListItems realizations={inheritedRealizations} capabilities={capabilities} origin="Inherited" />
    </Stack>
  );
};

interface TypeFieldProps {
  referenceUrl: string | undefined;
}

const TypeField: React.FC<TypeFieldProps> = ({ referenceUrl }) => {
  const hasReference = referenceUrl && referenceUrl.trim() !== '';

  return (
    <DetailField label="Type">
      {hasReference ? (
        <Anchor href={referenceUrl} target="_blank" rel="noopener noreferrer">
          Application Component
        </Anchor>
      ) : (
        'Application Component'
      )}
    </DetailField>
  );
};

interface FieldProps {
  component: Component;
}

function useComponentRecordEdit(component: Component) {
  const updateMutation = useUpdateComponent();
  const canEdit = hasLink(component, 'edit');
  const save = (patch: { name?: string; description?: string }) =>
    updateMutation.mutateAsync({
      component,
      request: { name: component.name, description: component.description, ...patch },
    });
  return { canEdit, save };
}

const NameField: React.FC<FieldProps> = ({ component }) => {
  const { canEdit, save } = useComponentRecordEdit(component);

  return (
    <InlineTextField
      value={component.name}
      canEdit={canEdit}
      schema={componentNameSchema}
      onSave={(name) => save({ name })}
      editLabel="Edit name"
      testId="component-name"
    />
  );
};

const DescriptionField: React.FC<FieldProps> = ({ component }) => {
  const { canEdit, save } = useComponentRecordEdit(component);

  return (
    <InlineTextField
      label="Description"
      value={component.description ?? ''}
      canEdit={canEdit}
      schema={componentDescriptionSchema}
      onSave={(description) => save({ description: description || undefined })}
      editLabel="Edit description"
      emptyPrompt="Add a description"
      multiline
      testId="component-description"
    />
  );
};

const CreatedField: React.FC<FieldProps> = ({ component }) => (
  <DetailField label="Created">
    <Text size="sm" c="dimmed">
      {new Date(component.createdAt).toLocaleString()}
    </Text>
  </DetailField>
);

interface ExpertsSectionProps {
  component: Component;
}

const ExpertsSection: React.FC<ExpertsSectionProps> = ({ component }) => {
  const [addExpertOpen, setAddExpertOpen] = useState(false);

  return (
    <>
      <ComponentExpertsList
        componentId={component.id}
        experts={component.experts}
        canAddExpert={hasLink(component, 'x-add-expert')}
        onAddClick={() => setAddExpertOpen(true)}
      />
      <AddComponentExpertDialog
        isOpen={addExpertOpen}
        onClose={() => setAddExpertOpen(false)}
        componentId={component.id}
      />
    </>
  );
};

function buildGroups({ component, realizations, capabilities }: ComponentDetailsContentProps): DetailGroup[] {
  return [
    {
      id: 'description',
      title: 'Description',
      content: (
        <Stack gap="sm">
          <DescriptionField component={component} />
          <TypeField referenceUrl={component._links.describedby?.href} />
        </Stack>
      ),
    },
    {
      id: 'ownership',
      title: 'Ownership',
      content: (
        <Stack gap="sm">
          <ComponentOwnershipSection component={component} />
          <ComponentHostingSection component={component} />
        </Stack>
      ),
    },
    { id: 'composition', title: 'Composition', content: <ComponentContainmentSection component={component} /> },
    {
      id: 'metadata',
      title: 'Metadata',
      content: (
        <Stack gap="sm">
          <ExpertsSection component={component} />
          <CreatedField component={component} />
        </Stack>
      ),
    },
    {
      id: 'realisations',
      title: 'Realises capabilities',
      content: <RealizationsField realizations={realizations} capabilities={capabilities} />,
    },
    { id: 'origins', title: 'Origins', content: <ComponentOriginsSection componentId={component.id} /> },
    { id: 'fit', title: 'Fit scores', content: <ComponentFitScores componentId={component.id} /> },
  ];
}

export const ComponentDetailsContent: React.FC<ComponentDetailsContentProps> = (props) => {
  const { component, viewMembership } = props;

  return (
    <DetailsShell
      layoutKey={LAYOUT_KEY}
      heading={<NameField component={component} />}
      groups={buildGroups(props)}
      viewMembership={viewMembership}
      footer={
        <>
          <OnePagerActionButton subject={component} subjectType="application" subjectId={component.id} />
          <AuditHistorySection aggregateId={component.id} />
        </>
      }
    />
  );
};
