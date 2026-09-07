import { Anchor, Badge, Group, Stack, Text } from '@mantine/core';
import React, { useState } from 'react';
import type { Capability, CapabilityRealization, Component } from '../../../api/types';
import { DetailField } from '../../../components/shared/DetailField';
import { InlineTextField } from '../../../components/shared/InlineTextField';
import { componentDescriptionSchema, componentNameSchema } from '../../../lib/schemas';
import { hasLink } from '../../../utils/hateoas';
import { AuditHistorySection } from '../../audit';
import { OnePagerActionButton } from '../../one-pagers/components/OnePagerActionButton';
import { useUpdateComponent } from '../hooks/useComponents';
import { AddComponentExpertDialog } from './AddComponentExpertDialog';
import { ComponentExpertsList } from './ComponentExpertsList';
import { ComponentFitScores } from './ComponentFitScores';
import { ComponentHostingSection } from './ComponentHostingSection';
import { ComponentOriginsSection } from './ComponentOriginsSection';
import { ComponentOwnershipSection } from './ComponentOwnershipSection';

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
  if (realizations.length === 0) return null;

  const directRealizations = realizations.filter((r) => r.origin === 'Direct');
  const inheritedRealizations = realizations.filter((r) => r.origin === 'Inherited');

  return (
    <DetailField label="Realizes Capabilities">
      <Stack gap="sm">
        <RealizationListItems realizations={directRealizations} capabilities={capabilities} origin="Direct" />
        <RealizationListItems realizations={inheritedRealizations} capabilities={capabilities} origin="Inherited" />
      </Stack>
    </DetailField>
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

interface NameAndDescriptionProps {
  component: Component;
}

const NameAndDescription: React.FC<NameAndDescriptionProps> = ({ component }) => {
  const updateMutation = useUpdateComponent();
  const canEdit = hasLink(component, 'edit');

  const saveName = (name: string) =>
    updateMutation.mutateAsync({ component, request: { name, description: component.description } });
  const saveDescription = (description: string) =>
    updateMutation.mutateAsync({ component, request: { name: component.name, description: description || undefined } });

  return (
    <>
      <InlineTextField
        value={component.name}
        canEdit={canEdit}
        schema={componentNameSchema}
        onSave={saveName}
        editLabel="Edit name"
        testId="component-name"
      />
      <InlineTextField
        label="Description"
        value={component.description ?? ''}
        canEdit={canEdit}
        schema={componentDescriptionSchema}
        onSave={saveDescription}
        editLabel="Edit description"
        emptyPrompt="Add a description"
        multiline
        testId="component-description"
      />
    </>
  );
};

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
      <AddComponentExpertDialog isOpen={addExpertOpen} onClose={() => setAddExpertOpen(false)} componentId={component.id} />
    </>
  );
};

export const ComponentDetailsContent: React.FC<ComponentDetailsContentProps> = ({
  component,
  realizations,
  capabilities,
  viewMembership,
}) => {
  const formattedDate = new Date(component.createdAt).toLocaleString();

  return (
    <Stack gap="sm" p="md" data-testid="component-details-panel">
      <NameAndDescription component={component} />
      <ComponentOwnershipSection component={component} />
      <ComponentHostingSection component={component} />
      <ExpertsSection component={component} />
      <DetailField label="Created">
        <Text size="sm" c="dimmed">
          {formattedDate}
        </Text>
      </DetailField>
      <TypeField referenceUrl={component._links.describedby?.href} />
      <RealizationsField realizations={realizations} capabilities={capabilities} />
      <ComponentOriginsSection componentId={component.id} />
      <ComponentFitScores componentId={component.id} />
      {viewMembership}
      <OnePagerActionButton subject={component} subjectType="application" subjectId={component.id} />
      <AuditHistorySection aggregateId={component.id} />
    </Stack>
  );
};
