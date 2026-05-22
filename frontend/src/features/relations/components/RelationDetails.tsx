import { Anchor, Badge, Button, Group, Stack, Text, Title } from '@mantine/core';
import React from 'react';
import type { Component, Relation } from '../../../api/types';
import { DetailField } from '../../../components/shared/DetailField';
import { useAppStore } from '../../../store/appStore';
import { AuditHistorySection } from '../../audit';
import { useComponents } from '../../components/hooks/useComponents';
import { useRelations } from '../hooks/useRelations';

interface RelationDetailsProps {
  onEdit: () => void;
}

interface RelationData {
  relation: Relation;
  sourceComponent: Component | undefined;
  targetComponent: Component | undefined;
  referenceLink: string | undefined;
  formattedDate: string;
}

const RELATION_TYPE_COLOR: Record<Relation['relationType'], string> = {
  Triggers: 'orange',
  Serves: 'blue',
};

const useRelationData = (selectedEdgeId: string | null): RelationData | null => {
  const { data: relations = [] } = useRelations();
  const { data: components = [] } = useComponents();

  if (!selectedEdgeId) {
    return null;
  }

  const relation = relations.find((r) => r.id === selectedEdgeId);

  if (!relation) {
    return null;
  }

  const sourceComponent = components.find((c) => c.id === relation.sourceComponentId);
  const targetComponent = components.find((c) => c.id === relation.targetComponentId);
  const referenceLink = relation._links.describedby?.href;
  const formattedDate = new Date(relation.createdAt).toLocaleString();

  return { relation, sourceComponent, targetComponent, referenceLink, formattedDate };
};

interface ReferenceLinkProps {
  href: string | undefined;
}

const ReferenceLink: React.FC<ReferenceLinkProps> = ({ href }) => {
  if (!href) return null;

  return (
    <Anchor href={href} target="_blank" rel="noopener noreferrer" size="sm">
      <Group gap="xs">
        <Text component="span" aria-hidden>
          📚
        </Text>
        <Text component="span">Reference Documentation</Text>
      </Group>
    </Anchor>
  );
};

interface EditActionProps {
  canEdit: boolean;
  onEdit: () => void;
}

const EditAction: React.FC<EditActionProps> = ({ canEdit, onEdit }) => {
  if (!canEdit) return null;

  return (
    <Group gap="sm">
      <Button variant="default" size="xs" onClick={onEdit}>
        Edit
      </Button>
    </Group>
  );
};

export const RelationDetails: React.FC<RelationDetailsProps> = ({ onEdit }) => {
  const selectedEdgeId = useAppStore((state) => state.selectedEdgeId);

  const data = useRelationData(selectedEdgeId);

  if (!data) {
    return null;
  }

  const { relation, sourceComponent, targetComponent, referenceLink, formattedDate } = data;
  const canEdit = relation._links?.edit !== undefined;

  return (
    <Stack gap="sm" p="md">
      <Title order={4}>Relation Details</Title>

      <EditAction canEdit={canEdit} onEdit={onEdit} />

      {relation.name && <DetailField label="Name">{relation.name}</DetailField>}

      <DetailField label="Type">
        <Badge color={RELATION_TYPE_COLOR[relation.relationType] ?? 'gray'} variant="light" size="sm">
          {relation.relationType}
        </Badge>
      </DetailField>

      <DetailField label="Source">{sourceComponent?.name || relation.sourceComponentId}</DetailField>

      <DetailField label="Target">{targetComponent?.name || relation.targetComponentId}</DetailField>

      {relation.description && <DetailField label="Description">{relation.description}</DetailField>}

      <DetailField label="Created">
        <Text size="sm" c="dimmed">
          {formattedDate}
        </Text>
      </DetailField>

      <DetailField label="ID">
        <Text size="xs" ff="monospace" c="gray.5" style={{ wordBreak: 'break-all' }}>
          {relation.id}
        </Text>
      </DetailField>

      <ReferenceLink href={referenceLink} />

      <AuditHistorySection aggregateId={relation.id} />
    </Stack>
  );
};
