import { Box, Button, Group, Loader, SegmentedControl, Stack, Text, Textarea, Title } from '@mantine/core';
import React, { useMemo, useState } from 'react';
import type { BusinessDomain, CapabilityId, StrategyImportance } from '../../../api/types';
import { useStrategyPillarsConfig } from '../../../hooks/useStrategyPillarsSettings';
import { canCreate } from '../../../utils/hateoas';
import {
  useRemoveStrategyImportance,
  useSetStrategyImportance,
  useStrategyImportanceByDomainAndCapability,
  useUpdateStrategyImportance,
} from '../hooks/useStrategyImportance';

interface StrategicImportanceSectionProps {
  domain: BusinessDomain;
  capabilityId: CapabilityId;
}

const SCORE_RANGE = [1, 2, 3, 4, 5] as const;

const IMPORTANCE_LABELS: Record<number, string> = {
  1: 'Low',
  2: 'Below Average',
  3: 'Average',
  4: 'Above Average',
  5: 'Critical',
};

interface StrategyPillar {
  id: string;
  name: string;
  description?: string;
  active: boolean;
}

interface ScoreDotsProps {
  value: number;
}

function ScoreDots({ value }: ScoreDotsProps) {
  return (
    <Group gap={4}>
      {SCORE_RANGE.map((s) => (
        <Box
          key={s}
          w={10}
          h={10}
          style={{
            borderRadius: '50%',
            background:
              s <= value
                ? 'var(--mantine-color-blue-6)'
                : 'var(--mantine-color-gray-3)',
          }}
        />
      ))}
    </Group>
  );
}

interface ImportanceEditFormProps {
  editScore: number | null;
  editRationale: string;
  onScoreChange: (score: number) => void;
  onRationaleChange: (rationale: string) => void;
  onCancel: () => void;
  onSave: () => void;
  isSaving: boolean;
}

function ImportanceEditForm({
  editScore,
  editRationale,
  onScoreChange,
  onRationaleChange,
  onCancel,
  onSave,
  isSaving,
}: ImportanceEditFormProps) {
  const segmentedData = SCORE_RANGE.map((s) => ({
    value: String(s),
    label: String(s),
  }));

  return (
    <Stack gap="xs">
      <SegmentedControl
        data={segmentedData}
        value={editScore ? String(editScore) : ''}
        onChange={(value) => onScoreChange(Number(value))}
        disabled={isSaving}
      />
      <Text size="sm" c="dimmed">
        {editScore ? IMPORTANCE_LABELS[editScore] : 'Select importance'}
      </Text>
      <Textarea
        placeholder="Rationale (optional)"
        value={editRationale}
        onChange={(e) => onRationaleChange(e.currentTarget.value)}
        maxLength={2000}
        disabled={isSaving}
        autosize
        minRows={2}
        data-testid="importance-rationale-input"
      />
      <Group justify="flex-end" gap="xs">
        <Button variant="default" size="xs" onClick={onCancel} disabled={isSaving}>
          Cancel
        </Button>
        <Button size="xs" onClick={onSave} disabled={!editScore || isSaving} loading={isSaving}>
          Save
        </Button>
      </Group>
    </Stack>
  );
}

interface ImportanceDisplayProps {
  pillarId: string;
  pillarName: string;
  importance: StrategyImportance | undefined;
  canAddImportance: boolean;
  onEdit: () => void;
  onDelete: () => void;
}

function ImportanceDisplay({
  pillarId,
  pillarName,
  importance,
  canAddImportance,
  onEdit,
  onDelete,
}: ImportanceDisplayProps) {
  if (importance) {
    return (
      <Stack gap={4}>
        <Group gap="sm" align="center">
          <ScoreDots value={importance.importance} />
          <Text size="sm" fw={600}>
            {importance.importance}/5
          </Text>
          <Text size="sm" c="dimmed">
            {importance.importanceLabel}
          </Text>
        </Group>
        {importance.rationale && (
          <Text size="sm" fs="italic" c="dimmed">
            "{importance.rationale}"
          </Text>
        )}
        <Group gap="xs">
          {importance._links?.edit && (
            <Button
              variant="subtle"
              size="compact-xs"
              onClick={onEdit}
              data-testid={`edit-importance-${pillarId}`}
            >
              Edit
            </Button>
          )}
          {importance._links?.delete && (
            <Button variant="subtle" color="red" size="compact-xs" onClick={onDelete}>
              Remove
            </Button>
          )}
        </Group>
      </Stack>
    );
  }

  if (canAddImportance) {
    return (
      <Button
        variant="subtle"
        size="compact-xs"
        onClick={onEdit}
        aria-label={`Add importance for ${pillarName}`}
        data-testid="add-importance-btn"
      >
        + Add Importance
      </Button>
    );
  }

  return null;
}

interface ImportanceRowProps {
  pillar: StrategyPillar;
  importance: StrategyImportance | undefined;
  canAddImportance: boolean;
  isEditing: boolean;
  editScore: number | null;
  editRationale: string;
  onEdit: () => void;
  onCancel: () => void;
  onScoreChange: (score: number) => void;
  onRationaleChange: (rationale: string) => void;
  onSave: () => void;
  onDelete: () => void;
  isSaving: boolean;
}

const ImportanceRow: React.FC<ImportanceRowProps> = ({
  pillar,
  importance,
  canAddImportance,
  isEditing,
  editScore,
  editRationale,
  onEdit,
  onCancel,
  onScoreChange,
  onRationaleChange,
  onSave,
  onDelete,
  isSaving,
}) => {
  return (
    <Stack gap="xs" data-testid={`importance-row-${pillar.id}`}>
      <Stack gap={2}>
        <Text size="sm" fw={600}>
          {pillar.name}
        </Text>
        {pillar.description && (
          <Text size="xs" c="dimmed">
            {pillar.description}
          </Text>
        )}
      </Stack>
      {isEditing ? (
        <ImportanceEditForm
          editScore={editScore}
          editRationale={editRationale}
          onScoreChange={onScoreChange}
          onRationaleChange={onRationaleChange}
          onCancel={onCancel}
          onSave={onSave}
          isSaving={isSaving}
        />
      ) : (
        <ImportanceDisplay
          pillarId={pillar.id}
          pillarName={pillar.name}
          importance={importance}
          canAddImportance={canAddImportance}
          onEdit={onEdit}
          onDelete={onDelete}
        />
      )}
    </Stack>
  );
};

function useImportanceEditing() {
  const [editingPillarId, setEditingPillarId] = useState<string | null>(null);
  const [editScore, setEditScore] = useState<number | null>(null);
  const [editRationale, setEditRationale] = useState('');

  const startEditing = (pillarId: string, existing: StrategyImportance | undefined) => {
    setEditingPillarId(pillarId);
    setEditScore(existing?.importance ?? null);
    setEditRationale(existing?.rationale ?? '');
  };

  const stopEditing = () => {
    setEditingPillarId(null);
    setEditScore(null);
    setEditRationale('');
  };

  return { editingPillarId, editScore, editRationale, setEditScore, setEditRationale, startEditing, stopEditing };
}

export function StrategicImportanceSection({ domain, capabilityId }: StrategicImportanceSectionProps) {
  const { data: importanceResponse, isLoading } = useStrategyImportanceByDomainAndCapability(domain.id, capabilityId);
  const { data: pillarsConfig } = useStrategyPillarsConfig();
  const setImportanceMutation = useSetStrategyImportance();
  const updateImportanceMutation = useUpdateStrategyImportance();
  const removeImportanceMutation = useRemoveStrategyImportance();

  const editing = useImportanceEditing();

  const activePillars = useMemo(() => {
    if (!pillarsConfig?.data) return [];
    return pillarsConfig.data.filter((p) => p.active);
  }, [pillarsConfig]);

  const importanceRatings = useMemo(() => importanceResponse?.data ?? [], [importanceResponse?.data]);
  const canAddImportance = useMemo(
    () => canCreate({ _links: importanceResponse?._links }),
    [importanceResponse?._links],
  );

  const importanceByPillar = useMemo(() => {
    return new Map(importanceRatings.map((r) => [r.pillarId, r]));
  }, [importanceRatings]);

  const getImportanceForPillar = (pillarId: string): StrategyImportance | undefined => {
    return importanceByPillar.get(pillarId);
  };

  const handleSave = async (pillarId: string) => {
    if (!editing.editScore) return;
    const existing = getImportanceForPillar(pillarId);

    if (existing) {
      await updateImportanceMutation.mutateAsync({
        domainId: domain.id,
        capabilityId,
        importanceId: existing.id,
        request: {
          importance: editing.editScore,
          rationale: editing.editRationale.trim() || undefined,
        },
      });
    } else {
      await setImportanceMutation.mutateAsync({
        domainId: domain.id,
        capabilityId,
        request: {
          pillarId,
          importance: editing.editScore,
          rationale: editing.editRationale.trim() || undefined,
        },
      });
    }
    editing.stopEditing();
  };

  const handleDelete = async (pillarId: string) => {
    const existing = getImportanceForPillar(pillarId);
    if (!existing) return;

    const pillar = activePillars.find((p) => p.id === pillarId);
    const confirmed = window.confirm(
      `Are you sure you want to remove the importance rating for "${pillar?.name || 'this pillar'}"?`,
    );
    if (!confirmed) return;

    await removeImportanceMutation.mutateAsync({
      domainId: domain.id,
      capabilityId,
      importanceId: existing.id,
    });
  };

  if (activePillars.length === 0) {
    return null;
  }

  const isSaving = setImportanceMutation.isPending || updateImportanceMutation.isPending;

  return (
    <Stack gap="sm" mt="md">
      <Title order={5}>Strategic Importance</Title>
      <Text size="sm" c="dimmed">
        Rate how important this capability is for {domain.name}
      </Text>
      {isLoading ? (
        <Group gap="xs">
          <Loader size="xs" />
          <Text size="sm" c="dimmed">
            Loading...
          </Text>
        </Group>
      ) : (
        <Stack gap="md">
          {activePillars.map((pillar) => (
            <ImportanceRow
              key={pillar.id}
              pillar={pillar}
              importance={getImportanceForPillar(pillar.id)}
              canAddImportance={canAddImportance}
              isEditing={editing.editingPillarId === pillar.id}
              editScore={editing.editScore}
              editRationale={editing.editRationale}
              onEdit={() => editing.startEditing(pillar.id, getImportanceForPillar(pillar.id))}
              onCancel={editing.stopEditing}
              onScoreChange={editing.setEditScore}
              onRationaleChange={editing.setEditRationale}
              onSave={() => handleSave(pillar.id)}
              onDelete={() => handleDelete(pillar.id)}
              isSaving={isSaving}
            />
          ))}
        </Stack>
      )}
    </Stack>
  );
}
