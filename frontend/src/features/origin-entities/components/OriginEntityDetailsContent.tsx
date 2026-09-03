import { Badge, Stack, Text } from '@mantine/core';
import type React from 'react';
import type {
  AcquiredEntity,
  InternalTeam,
  OriginRelationship,
  UpdateAcquiredEntityRequest,
  UpdateInternalTeamRequest,
  UpdateVendorRequest,
  Vendor,
} from '../../../api/types';
import { DetailField } from '../../../components/shared/DetailField';
import { InlineDateField } from '../../../components/shared/InlineDateField';
import { InlineSelectField } from '../../../components/shared/InlineSelectField';
import { InlineTextField } from '../../../components/shared/InlineTextField';
import type { OriginEntityType } from '../../../constants/entityIdentifiers';
import { originEntityNameSchema } from '../../../lib/schemas/originEntity';
import { hasLink, type ResourceWithLinks } from '../../../utils/hateoas';
import { AuditHistorySection } from '../../audit';
import { OnePagerActionButton } from '../../one-pagers/components/OnePagerActionButton';
import type { OnePagerSubjectType } from '../../one-pagers/types';
import { useUpdateAcquiredEntity } from '../hooks/useAcquiredEntities';
import { useUpdateInternalTeam } from '../hooks/useInternalTeams';
import { useUpdateVendor } from '../hooks/useVendors';
import { OriginEntityRelationshipsList } from './OriginEntityRelationshipsList';
import {
  acquiredEntityDefinition,
  type FieldDefinition,
  integrationStatusMeta,
  internalTeamDefinition,
  type OriginEntityDefinition,
  vendorDefinition,
} from './originEntityFields';

const TEST_ID = 'origin-entity';

const SUBJECT_TYPES: Record<OriginEntityType, OnePagerSubjectType> = {
  acquired: 'acquired-entity',
  vendor: 'vendor',
  team: 'internal-team',
};

const RELATIONSHIP_LABELS: Record<OriginEntityType, string> = {
  acquired: 'Acquired via',
  vendor: 'Purchased from',
  team: 'Built by',
};

function testIdFor(key: string): string {
  return `${TEST_ID}-${key.replace(/[A-Z]/g, (letter) => `-${letter.toLowerCase()}`)}`;
}

interface FieldProps<E> {
  field: FieldDefinition<E>;
  entity: E;
  canEdit: boolean;
  onSave: (key: string, value: string) => Promise<unknown>;
}

function Field<E>({ field, entity, canEdit, onSave }: FieldProps<E>) {
  const value = field.read(entity);
  const save = (next: string) => onSave(field.key, next);
  const testId = testIdFor(field.key);

  switch (field.kind) {
    case 'select':
      return (
        <InlineSelectField
          label={field.label}
          value={value}
          options={field.options}
          canEdit={canEdit}
          onSave={save}
          editLabel={field.editLabel}
          emptyPrompt={field.emptyPrompt}
          testId={testId}
          renderValue={(status) => {
            const meta = integrationStatusMeta(status);
            return (
              <Badge color={meta.color} variant="dot" size="sm">
                {meta.label}
              </Badge>
            );
          }}
        />
      );
    case 'date':
      return (
        <InlineDateField
          label={field.label}
          value={value}
          canEdit={canEdit}
          onSave={save}
          editLabel={field.editLabel}
          emptyPrompt={field.emptyPrompt}
          testId={testId}
        />
      );
    default:
      return (
        <InlineTextField
          label={field.label}
          value={value}
          canEdit={canEdit}
          schema={field.schema}
          onSave={save}
          editLabel={field.editLabel}
          emptyPrompt={field.emptyPrompt}
          multiline={field.kind === 'multiline'}
          testId={testId}
        />
      );
  }
}

interface EntityFieldsProps<E extends ResourceWithLinks & { name: string }, R> {
  entity: E;
  definition: OriginEntityDefinition<E, R>;
  save: (request: R) => Promise<unknown>;
}

function EntityFields<E extends ResourceWithLinks & { name: string }, R extends object>({
  entity,
  definition,
  save,
}: EntityFieldsProps<E, R>) {
  const canEdit = hasLink(entity, 'edit');
  const saveField = (key: string, value: string) =>
    save({ ...definition.toRequest(entity), [key]: value || undefined } as R);

  return (
    <>
      <InlineTextField
        value={entity.name}
        canEdit={canEdit}
        schema={originEntityNameSchema}
        onSave={(name) => save({ ...definition.toRequest(entity), name } as R)}
        editLabel="Edit name"
        testId={`${TEST_ID}-name`}
      />
      {definition.fields.map((field) => (
        <Field key={field.key} field={field} entity={entity} canEdit={canEdit} onSave={saveField} />
      ))}
    </>
  );
}

export type OriginEntity = AcquiredEntity | Vendor | InternalTeam;

type SaveOriginEntity = (entity: OriginEntity, request: object) => Promise<unknown>;

function useSaveOriginEntity(entityType: OriginEntityType): SaveOriginEntity {
  const acquired = useUpdateAcquiredEntity();
  const vendor = useUpdateVendor();
  const team = useUpdateInternalTeam();

  switch (entityType) {
    case 'acquired':
      return (entity, request) =>
        acquired.mutateAsync({ entity: entity as AcquiredEntity, request: request as UpdateAcquiredEntityRequest });
    case 'vendor':
      return (entity, request) =>
        vendor.mutateAsync({ vendor: entity as Vendor, request: request as UpdateVendorRequest });
    case 'team':
      return (entity, request) =>
        team.mutateAsync({ team: entity as InternalTeam, request: request as UpdateInternalTeamRequest });
  }
}

function TypedFields({ entityType, entity }: { entityType: OriginEntityType; entity: OriginEntity }) {
  const save = useSaveOriginEntity(entityType);
  const saveFor = (target: OriginEntity) => (request: object) => save(target, request);

  switch (entityType) {
    case 'acquired':
      return (
        <EntityFields entity={entity as AcquiredEntity} definition={acquiredEntityDefinition} save={saveFor(entity)} />
      );
    case 'vendor':
      return <EntityFields entity={entity as Vendor} definition={vendorDefinition} save={saveFor(entity)} />;
    case 'team':
      return (
        <EntityFields entity={entity as InternalTeam} definition={internalTeamDefinition} save={saveFor(entity)} />
      );
  }
}

const TYPE_LABELS: Record<OriginEntityType, string> = {
  acquired: acquiredEntityDefinition.typeLabel,
  vendor: vendorDefinition.typeLabel,
  team: internalTeamDefinition.typeLabel,
};

export interface OriginEntityDetailsContentProps {
  entityType: OriginEntityType;
  entity: OriginEntity;
  relationships: OriginRelationship[];
  viewMembership?: React.ReactNode;
}

export const OriginEntityDetailsContent: React.FC<OriginEntityDetailsContentProps> = ({
  entityType,
  entity,
  relationships,
  viewMembership,
}) => (
  <Stack gap="sm">
    <TypedFields entityType={entityType} entity={entity} />
    <DetailField label="Created">
      <Text size="sm" c="dimmed">
        {new Date(entity.createdAt).toLocaleString()}
      </Text>
    </DetailField>
    <DetailField label="Type">{TYPE_LABELS[entityType]}</DetailField>
    <OriginEntityRelationshipsList relationships={relationships} relationshipLabel={RELATIONSHIP_LABELS[entityType]} />
    {viewMembership}
    <OnePagerActionButton subject={entity} subjectType={SUBJECT_TYPES[entityType]} subjectId={entity.id} />
    <AuditHistorySection aggregateId={entity.id} />
  </Stack>
);
