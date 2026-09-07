import { Badge, Stack, Text } from '@mantine/core';
import type React from 'react';
import type { CapabilityRealization, RealizationLevel } from '../../../api/types';
import { DetailField } from '../../../components/shared/DetailField';
import { InlineSelectField } from '../../../components/shared/InlineSelectField';
import { InlineTextField } from '../../../components/shared/InlineTextField';
import { realizationNotesSchema } from '../../../lib/schemas/relation';
import { hasLink } from '../../../utils/hateoas';
import { AuditHistorySection } from '../../audit';
import { useUpdateRealization } from '../../capabilities/hooks/useCapabilities';
import type { RealizationData } from '../hooks/useRealizationDetails';
import { InheritedRealizationInfo } from './InheritedRealizationInfo';
import { OriginBadge } from './OriginBadge';

const REALIZATION_LEVEL_OPTIONS: { value: RealizationLevel; label: string }[] = [
  { value: 'Full', label: 'Full (100%)' },
  { value: 'Partial', label: 'Partial' },
  { value: 'Planned', label: 'Planned' },
];

interface LevelAndNotesProps {
  realization: CapabilityRealization;
  canEdit: boolean;
}

const LevelAndNotes: React.FC<LevelAndNotesProps> = ({ realization, canEdit }) => {
  const updateMutation = useUpdateRealization();
  const save = (patch: { realizationLevel?: RealizationLevel; notes?: string }) =>
    updateMutation.mutateAsync({
      realization,
      request: { realizationLevel: realization.realizationLevel, notes: realization.notes || undefined, ...patch },
    });

  return (
    <>
      <InlineSelectField
        label="Realization Level"
        value={realization.realizationLevel}
        options={REALIZATION_LEVEL_OPTIONS}
        canEdit={canEdit}
        onSave={(level) => save({ realizationLevel: level as RealizationLevel })}
        editLabel="Edit realization level"
        testId="realization-level"
        renderValue={(_, label) => (
          <Badge color="gray" variant="filled" size="sm">
            {label}
          </Badge>
        )}
      />
      <OriginBadge origin={realization.origin} isInherited={realization.origin === 'Inherited'} />
      <InlineTextField
        label="Notes"
        value={realization.notes ?? ''}
        canEdit={canEdit}
        schema={realizationNotesSchema}
        onSave={(notes) => save({ notes: notes || undefined })}
        editLabel="Edit notes"
        emptyPrompt="Add notes"
        multiline
        testId="realization-notes"
      />
    </>
  );
};

interface RealizationDetailsContentProps {
  data: RealizationData;
}

export const RealizationDetailsContent: React.FC<RealizationDetailsContentProps> = ({ data }) => {
  const { realization, capability, component, formattedDate, isInherited } = data;
  const canEdit = !isInherited && hasLink(realization, 'edit');

  return (
    <Stack gap="sm" p="md">
      <DetailField label="Capability">{capability?.name || 'Unknown'}</DetailField>
      <DetailField label="Application">{component?.name || 'Unknown'}</DetailField>
      <LevelAndNotes realization={realization} canEdit={canEdit} />
      <DetailField label="Linked">
        <Text size="sm" c="dimmed">
          {formattedDate}
        </Text>
      </DetailField>
      <InheritedRealizationInfo isInherited={isInherited} />
      <AuditHistorySection aggregateId={realization.id} />
    </Stack>
  );
};
