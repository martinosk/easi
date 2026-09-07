import { Badge, Button, Group, SegmentedControl, Stack, Text, Textarea } from '@mantine/core';
import { useState } from 'react';
import type { CapabilityId, ComponentId } from '../../../api/types';
import { hasLink } from '../../../utils/hateoas';
import { useAssessRealization, useRemoveTimeAssessment } from '../../architecture-direction/hooks/useTimeAssessments';
import type {
  TimeAssessment,
  TimeAssessmentGradeCounts,
  TimeGrade,
  TimeSuggestion,
} from '../../architecture-direction/types';
import { TIME_GRADES } from '../../architecture-direction/utils/timeGrade';

function formatAssessedDate(dateString: string): string {
  return new Date(dateString).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
}

function SuggestionLine({ componentId, suggestion }: { componentId: string; suggestion: TimeSuggestion | null }) {
  if (!suggestion?.grade) return null;

  return (
    <Text size="xs" c="dimmed" data-testid={`assessment-suggestion-${componentId}`}>
      Suggested: {suggestion.grade} — {suggestion.confidence.toLowerCase()} confidence
    </Text>
  );
}

function RollupLine({ componentId, rollup }: { componentId: string; rollup: TimeAssessmentGradeCounts }) {
  return (
    <Text size="xs" c="dimmed" data-testid={`assessment-rollup-${componentId}`}>
      Across landscape: I×{rollup.Invest} · T×{rollup.Tolerate} · M×{rollup.Migrate} · E×{rollup.Eliminate}
    </Text>
  );
}

interface AssessmentFormProps {
  componentId: string;
  initialGrade: TimeGrade | null;
  suggestion: TimeSuggestion | null;
  isSaving: boolean;
  onCancel: () => void;
  onSave: (grade: TimeGrade, rationale: string) => void;
}

function AssessmentForm({ componentId, initialGrade, suggestion, isSaving, onCancel, onSave }: AssessmentFormProps) {
  const [grade, setGrade] = useState<TimeGrade | null>(initialGrade);
  const [rationale, setRationale] = useState('');

  return (
    <Stack gap="xs" data-testid={`assessment-form-${componentId}`}>
      <SuggestionLine componentId={componentId} suggestion={suggestion} />
      <SegmentedControl
        data={TIME_GRADES.map((g) => ({ value: g, label: g }))}
        value={grade ?? ''}
        onChange={(value) => setGrade(value as TimeGrade)}
        disabled={isSaving}
      />
      <Textarea
        placeholder="Rationale (optional)"
        value={rationale}
        onChange={(e) => setRationale(e.currentTarget.value)}
        maxLength={2000}
        disabled={isSaving}
        autosize
        minRows={2}
      />
      <Group justify="flex-end" gap="xs">
        <Button variant="default" size="xs" onClick={onCancel} disabled={isSaving}>
          Cancel
        </Button>
        <Button
          size="xs"
          onClick={() => grade && onSave(grade, rationale)}
          disabled={!grade || isSaving}
          loading={isSaving}
        >
          Save
        </Button>
      </Group>
    </Stack>
  );
}

function UnassessedState({
  componentId,
  canAssess,
  suggestion,
  onAssess,
}: {
  componentId: string;
  canAssess: boolean;
  suggestion: TimeSuggestion | null;
  onAssess: () => void;
}) {
  return (
    <Stack gap={4} data-testid={`assessment-${componentId}`}>
      <Text size="sm" c="dimmed">
        unassessed
      </Text>
      <SuggestionLine componentId={componentId} suggestion={suggestion} />
      {canAssess && (
        <Group>
          <Button variant="subtle" size="compact-xs" onClick={onAssess} data-testid={`assess-btn-${componentId}`}>
            Assess
          </Button>
        </Group>
      )}
    </Stack>
  );
}

interface AssessmentControlsProps {
  componentId: string;
  canEdit: boolean;
  canDelete: boolean;
  isRemoving: boolean;
  onReassess: () => void;
  onRemove: () => void;
}

function AssessmentControls({ componentId, canEdit, canDelete, isRemoving, onReassess, onRemove }: AssessmentControlsProps) {
  if (!canEdit && !canDelete) return null;

  return (
    <Group gap="xs">
      {canEdit && (
        <Button variant="subtle" size="compact-xs" onClick={onReassess} data-testid={`reassess-btn-${componentId}`}>
          Re-assess
        </Button>
      )}
      {canDelete && (
        <Button
          variant="subtle"
          color="red"
          size="compact-xs"
          onClick={onRemove}
          loading={isRemoving}
          data-testid={`remove-assessment-btn-${componentId}`}
        >
          Remove
        </Button>
      )}
    </Group>
  );
}

interface AssessmentSummaryProps {
  componentId: string;
  assessment: TimeAssessment;
  rollup: TimeAssessmentGradeCounts | undefined;
  suggestion: TimeSuggestion | null;
  isRemoving: boolean;
  onReassess: () => void;
  onRemove: () => void;
}

function AssessmentSummary({
  componentId,
  assessment,
  rollup,
  suggestion,
  isRemoving,
  onReassess,
  onRemove,
}: AssessmentSummaryProps) {
  return (
    <Stack gap={4} data-testid={`assessment-${componentId}`}>
      <Group gap="xs">
        <Text size="sm">{assessment.grade} — for this capability</Text>
        {assessment.stale && (
          <Badge color="orange" variant="light" size="xs" data-testid={`assessment-stale-${componentId}`}>
            stale
          </Badge>
        )}
      </Group>
      {assessment.assessedAt && (
        <Text size="xs" c="dimmed">
          Assessed {formatAssessedDate(assessment.assessedAt)} by {assessment.assessedByName ?? assessment.assessedBy}
        </Text>
      )}
      <SuggestionLine componentId={componentId} suggestion={suggestion} />
      {rollup && <RollupLine componentId={componentId} rollup={rollup} />}
      <AssessmentControls
        componentId={componentId}
        canEdit={hasLink(assessment, 'edit')}
        canDelete={hasLink(assessment, 'delete')}
        isRemoving={isRemoving}
        onReassess={onReassess}
        onRemove={onRemove}
      />
    </Stack>
  );
}

export interface RealizationAssessmentProps {
  capabilityId: CapabilityId | string;
  componentId: ComponentId | string;
  assessment: TimeAssessment | undefined;
  rollup: TimeAssessmentGradeCounts | undefined;
  canAssess: boolean;
  suggestion: TimeSuggestion | null;
}

export function RealizationAssessment({
  capabilityId,
  componentId,
  assessment,
  rollup,
  canAssess,
  suggestion,
}: RealizationAssessmentProps) {
  const [editing, setEditing] = useState(false);
  const assessMutation = useAssessRealization();
  const removeMutation = useRemoveTimeAssessment();

  const id = String(componentId);

  const handleSave = async (grade: TimeGrade, rationale: string) => {
    await assessMutation.mutateAsync({
      capabilityId: String(capabilityId),
      componentId: id,
      request: { grade, rationale: rationale.trim() || undefined },
    });
    setEditing(false);
  };

  const handleRemove = async () => {
    if (!assessment) return;
    await removeMutation.mutateAsync({ assessment });
  };

  if (editing) {
    return (
      <AssessmentForm
        componentId={id}
        initialGrade={assessment?.grade ?? null}
        suggestion={suggestion}
        isSaving={assessMutation.isPending}
        onCancel={() => setEditing(false)}
        onSave={handleSave}
      />
    );
  }

  if (!assessment) {
    return (
      <UnassessedState
        componentId={id}
        canAssess={canAssess}
        suggestion={suggestion}
        onAssess={() => setEditing(true)}
      />
    );
  }

  return (
    <AssessmentSummary
      componentId={id}
      assessment={assessment}
      rollup={rollup}
      suggestion={suggestion}
      isRemoving={removeMutation.isPending}
      onReassess={() => setEditing(true)}
      onRemove={handleRemove}
    />
  );
}
