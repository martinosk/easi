import { Badge, Group, Stack, Text } from '@mantine/core';
import React, { useState } from 'react';
import type { Capability, UpdateCapabilityMetadataRequest } from '../../../api/types';
import { DetailField } from '../../../components/shared/DetailField';
import { InlineSelectField } from '../../../components/shared/InlineSelectField';
import { InlineTextField } from '../../../components/shared/InlineTextField';
import { useMaturityScale } from '../../../hooks/useMaturityScale';
import {
  addTagSchema,
  capabilityDescriptionSchema,
  capabilityNameSchema,
  capabilityPrimaryOwnerSchema,
} from '../../../lib/schemas/capability';
import { hasLink } from '../../../utils/hateoas';
import { deriveLegacyMaturityValue, getDefaultSections } from '../../../utils/maturity';
import { useAddCapabilityTag, useUpdateCapability, useUpdateCapabilityMetadata } from '../hooks/useCapabilities';
import { useCapabilityMetadataOptions } from '../hooks/useCapabilityMetadataOptions';
import { AddExpertDialog } from './AddExpertDialog';
import { CapabilityExpertsList } from './CapabilityExpertsList';
import { InlineMaturityField } from './InlineMaturityField';

const DEFAULT_LEGACY_MATURITY = 12;
const DEFAULT_STATUS = 'Active';

function currentMaturityValue(capability: Capability, sections = getDefaultSections()): number {
  if (capability.maturityValue !== undefined) return capability.maturityValue;
  if (capability.maturityLevel) return deriveLegacyMaturityValue(capability.maturityLevel, sections);
  return DEFAULT_LEGACY_MATURITY;
}

function metadataRequest(
  capability: Capability,
  maturityValue: number,
  patch: Partial<UpdateCapabilityMetadataRequest>,
): UpdateCapabilityMetadataRequest {
  return {
    status: capability.status ?? DEFAULT_STATUS,
    maturityValue,
    ownershipModel: capability.ownershipModel || undefined,
    primaryOwner: capability.primaryOwner || undefined,
    eaOwner: capability.eaOwner || undefined,
    ...patch,
  };
}

function useMetadataCommit(capability: Capability) {
  const updateMetadata = useUpdateCapabilityMetadata();
  const { data: maturityScale } = useMaturityScale();
  const sections = maturityScale?.sections?.length ? maturityScale.sections : getDefaultSections();
  const maturityValue = currentMaturityValue(capability, sections);

  const save = (patch: Partial<UpdateCapabilityMetadataRequest>) =>
    updateMetadata.mutateAsync({ capability, request: metadataRequest(capability, maturityValue, patch) });

  return { canEdit: hasLink(capability, 'x-update-metadata'), maturityValue, save };
}

export interface SectionProps {
  capability: Capability;
}

function useCapabilityRecordEdit(capability: Capability) {
  const updateMutation = useUpdateCapability();
  const canEdit = hasLink(capability, 'edit');
  const save = (patch: { name?: string; description?: string }) =>
    updateMutation.mutateAsync({
      capability,
      request: { name: capability.name, description: capability.description, ...patch },
    });
  return { canEdit, save };
}

export const NameField: React.FC<SectionProps> = ({ capability }) => {
  const { canEdit, save } = useCapabilityRecordEdit(capability);

  return (
    <InlineTextField
      value={capability.name}
      canEdit={canEdit}
      schema={capabilityNameSchema}
      onSave={(name) => save({ name })}
      editLabel="Edit name"
      testId="capability-name"
    />
  );
};

export const DescriptionField: React.FC<SectionProps> = ({ capability }) => {
  const { canEdit, save } = useCapabilityRecordEdit(capability);

  return (
    <InlineTextField
      label="Description"
      value={capability.description ?? ''}
      canEdit={canEdit}
      schema={capabilityDescriptionSchema}
      onSave={(description) => save({ description: description || undefined })}
      editLabel="Edit description"
      emptyPrompt="Add a description"
      multiline
      testId="capability-description"
    />
  );
};

export const LevelField: React.FC<SectionProps> = ({ capability }) => (
  <DetailField label="Level">
    <Badge color="dark" variant="filled" size="sm">
      {capability.level}
    </Badge>
  </DetailField>
);

export const StatusField: React.FC<SectionProps> = ({ capability }) => {
  const { canEdit, save } = useMetadataCommit(capability);
  const { statusOptions } = useCapabilityMetadataOptions();

  return (
    <InlineSelectField
      label="Status"
      value={capability.status ?? ''}
      options={statusOptions}
      canEdit={canEdit}
      onSave={(status) => save({ status })}
      editLabel="Edit status"
      emptyPrompt="Set a status"
      testId="capability-status"
    />
  );
};

export const MaturityField: React.FC<SectionProps> = ({ capability }) => {
  const { canEdit, maturityValue, save } = useMetadataCommit(capability);

  return (
    <InlineMaturityField value={maturityValue} canEdit={canEdit} onSave={(value) => save({ maturityValue: value })} />
  );
};

export const OwnershipFields: React.FC<SectionProps> = ({ capability }) => {
  const { canEdit, save } = useMetadataCommit(capability);
  const { ownershipOptions, eaOwnerOptions } = useCapabilityMetadataOptions();

  return (
    <>
      <InlineSelectField
        label="Ownership Model"
        value={capability.ownershipModel ?? ''}
        options={ownershipOptions}
        canEdit={canEdit}
        onSave={(ownershipModel) => save({ ownershipModel })}
        editLabel="Edit ownership model"
        emptyPrompt="Set an ownership model"
        testId="capability-ownership-model"
      />
      <InlineTextField
        label="Primary Owner"
        value={capability.primaryOwner ?? ''}
        canEdit={canEdit}
        schema={capabilityPrimaryOwnerSchema}
        onSave={(primaryOwner) => save({ primaryOwner })}
        editLabel="Edit primary owner"
        emptyPrompt="Set a primary owner"
        testId="capability-primary-owner"
      />
      <InlineSelectField
        label="EA Owner"
        value={capability.eaOwner ?? ''}
        options={eaOwnerOptions}
        canEdit={canEdit}
        onSave={(eaOwner) => save({ eaOwner })}
        editLabel="Edit EA owner"
        emptyPrompt="Set an EA owner"
        testId="capability-ea-owner"
        searchable
        renderValue={(_, label) => <Text size="sm">{capability.eaOwnerName ?? label}</Text>}
      />
    </>
  );
};

const tagSchema = addTagSchema.shape.tag;

export const TagsSection: React.FC<SectionProps> = ({ capability }) => {
  const addTag = useAddCapabilityTag();
  const canAddTag = hasLink(capability, 'x-add-tag');
  const tags = capability.tags ?? [];

  if (tags.length === 0 && !canAddTag) return null;

  return (
    <DetailField label="Tags">
      <Stack gap="xs">
        {tags.length > 0 && (
          <Group gap="xs">
            {tags.map((tag) => (
              <Badge key={tag} variant="light">
                {tag}
              </Badge>
            ))}
          </Group>
        )}
        <InlineTextField
          value=""
          canEdit={canAddTag}
          schema={tagSchema}
          onSave={(tag) => addTag.mutateAsync({ capability, request: { tag } })}
          editLabel="Add a tag"
          emptyPrompt="Add a tag"
          testId="capability-tag"
        />
      </Stack>
    </DetailField>
  );
};

export const ExpertsSection: React.FC<SectionProps> = ({ capability }) => {
  const [addExpertOpen, setAddExpertOpen] = useState(false);

  return (
    <>
      <CapabilityExpertsList
        capabilityId={capability.id}
        experts={capability.experts}
        canAddExpert={hasLink(capability, 'x-add-expert')}
        onAddClick={() => setAddExpertOpen(true)}
      />
      <AddExpertDialog isOpen={addExpertOpen} onClose={() => setAddExpertOpen(false)} capabilityId={capability.id} />
    </>
  );
};

export const CreatedField: React.FC<SectionProps> = ({ capability }) => (
  <DetailField label="Created">
    <Text size="sm" c="dimmed">
      {new Date(capability.createdAt).toLocaleString()}
    </Text>
  </DetailField>
);
