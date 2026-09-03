import { Anchor, Badge, Box, Group, Stack, Text } from '@mantine/core';
import { IconExternalLink } from '@tabler/icons-react';
import type React from 'react';
import type { Component, Relation } from '../../../api/types';
import { DetailField } from '../../../components/shared/DetailField';
import { InlineTextField } from '../../../components/shared/InlineTextField';
import { relationDescriptionSchema, relationNameSchema } from '../../../lib/schemas/relation';
import { useAppStore } from '../../../store/appStore';
import { hasLink } from '../../../utils/hateoas';
import { AuditHistorySection } from '../../audit';
import { useComponents } from '../../components/hooks/useComponents';
import { useRelations, useUpdateRelation } from '../hooks/useRelations';

interface RelationData {
  relation: Relation;
  sourceComponent: Component | undefined;
  targetComponent: Component | undefined;
}

const RELATION_TYPE_COLOR: Record<Relation['relationType'], string> = {
  Triggers: 'orange',
  Serves: 'blue',
};

const useRelationData = (selectedEdgeId: string | null): RelationData | null => {
  const { data: relations = [] } = useRelations();
  const { data: components = [] } = useComponents();

  const relation = relations.find((r) => r.id === selectedEdgeId);
  if (!relation) return null;

  return {
    relation,
    sourceComponent: components.find((c) => c.id === relation.sourceComponentId),
    targetComponent: components.find((c) => c.id === relation.targetComponentId),
  };
};

const ReferenceLink: React.FC<{ href: string | undefined }> = ({ href }) => {
  if (!href) return null;

  return (
    <Anchor href={href} target="_blank" rel="noopener noreferrer" size="sm">
      <Group gap="xs">
        <Box component="span" aria-hidden>
          <IconExternalLink size={16} stroke={1.75} />
        </Box>
        <Text component="span">Reference Documentation</Text>
      </Group>
    </Anchor>
  );
};

interface RelationFieldProps {
  relation: Relation;
}

const RelationName: React.FC<RelationFieldProps> = ({ relation }) => {
  const updateMutation = useUpdateRelation();

  return (
    <InlineTextField
      value={relation.name ?? ''}
      canEdit={hasLink(relation, 'edit')}
      schema={relationNameSchema}
      onSave={(name) =>
        updateMutation.mutateAsync({
          relation,
          request: { name: name || undefined, description: relation.description },
        })
      }
      editLabel="Edit name"
      emptyPrompt="Add a name"
      testId="relation-name"
    />
  );
};

const RelationDescription: React.FC<RelationFieldProps> = ({ relation }) => {
  const updateMutation = useUpdateRelation();

  return (
    <InlineTextField
      label="Description"
      value={relation.description ?? ''}
      canEdit={hasLink(relation, 'edit')}
      schema={relationDescriptionSchema}
      onSave={(description) =>
        updateMutation.mutateAsync({
          relation,
          request: { name: relation.name, description: description || undefined },
        })
      }
      editLabel="Edit description"
      emptyPrompt="Add a description"
      multiline
      testId="relation-description"
    />
  );
};

export const RelationDetails: React.FC = () => {
  const selectedEdgeId = useAppStore((state) => state.selectedEdgeId);
  const data = useRelationData(selectedEdgeId);
  if (!data) return null;

  const { relation, sourceComponent, targetComponent } = data;

  return (
    <Stack gap="sm" p="md">
      <RelationName relation={relation} />
      <DetailField label="Type">
        <Badge color={RELATION_TYPE_COLOR[relation.relationType] ?? 'gray'} variant="light" size="sm">
          {relation.relationType}
        </Badge>
      </DetailField>
      <DetailField label="Source">{sourceComponent?.name || relation.sourceComponentId}</DetailField>
      <DetailField label="Target">{targetComponent?.name || relation.targetComponentId}</DetailField>
      <RelationDescription relation={relation} />
      <DetailField label="Created">
        <Text size="sm" c="dimmed">
          {new Date(relation.createdAt).toLocaleString()}
        </Text>
      </DetailField>
      <ReferenceLink href={relation._links.describedby?.href} />
      <AuditHistorySection aggregateId={relation.id} />
    </Stack>
  );
};
