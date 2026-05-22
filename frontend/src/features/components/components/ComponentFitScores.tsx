import { Box, Button, ColorSwatch, Divider, Group, Paper, Stack, Text, Textarea, Title } from '@mantine/core';
import React, { useMemo, useState } from 'react';
import type { ApplicationFitScore, ComponentId, StrategyPillar } from '../../../api/types';
import { useStrategyPillarsConfig } from '../../../hooks/useStrategyPillarsSettings';
import { canCreate, canDelete, canEdit } from '../../../utils/hateoas';
import { useComponentFitScores, useDeleteFitScore, useSetFitScore } from '../hooks/useFitScores';

interface ComponentFitScoresProps {
  componentId: ComponentId;
}

const SCORE_RANGE = [1, 2, 3, 4, 5] as const;

const SCORE_LABELS: Record<number, string> = {
  1: 'Critical',
  2: 'Poor',
  3: 'Adequate',
  4: 'Good',
  5: 'Excellent',
};

interface FitScoreRowProps {
  pillar: StrategyPillar;
  score: ApplicationFitScore | undefined;
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
  canEditScore: boolean;
  canDeleteScore: boolean;
  canAddScore: boolean;
}

interface FitScoreEditFormProps {
  editScore: number | null;
  editRationale: string;
  onScoreChange: (score: number) => void;
  onRationaleChange: (rationale: string) => void;
  onSave: () => void;
  onCancel: () => void;
  isSaving: boolean;
}

const FitScoreEditForm: React.FC<FitScoreEditFormProps> = ({
  editScore,
  editRationale,
  onScoreChange,
  onRationaleChange,
  onSave,
  onCancel,
  isSaving,
}) => (
  <Stack gap="xs">
    <Group gap="xs">
      {SCORE_RANGE.map((s) => (
        <Button
          key={s}
          variant={editScore === s ? 'filled' : 'default'}
          size="xs"
          onClick={() => onScoreChange(s)}
          disabled={isSaving}
          data-testid={`fit-score-btn-${s}`}
        >
          {s}
        </Button>
      ))}
    </Group>
    <Text size="xs" c="dimmed">
      {editScore ? SCORE_LABELS[editScore] : 'Select score'}
    </Text>
    <Textarea
      placeholder="Rationale (optional)"
      value={editRationale}
      onChange={(e) => onRationaleChange(e.target.value)}
      maxLength={2000}
      disabled={isSaving}
      autosize
      minRows={2}
      data-testid="fit-score-rationale-input"
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

interface FitScoreDotsProps {
  score: number;
}

const FitScoreDots: React.FC<FitScoreDotsProps> = ({ score }) => (
  <Group gap="xs">
    {SCORE_RANGE.map((s) => (
      <ColorSwatch
        key={s}
        size="xs"
        color={s <= score ? 'var(--mantine-color-blue-6)' : 'var(--mantine-color-gray-3)'}
        withShadow={false}
      />
    ))}
  </Group>
);

interface FitScoreDisplayProps {
  score: ApplicationFitScore | undefined;
  onEdit: () => void;
  onDelete: () => void;
  canEditScore: boolean;
  canDeleteScore: boolean;
  canAddScore: boolean;
  pillarName: string;
}

const FitScoreDisplay: React.FC<FitScoreDisplayProps> = ({
  score,
  onEdit,
  onDelete,
  canEditScore,
  canDeleteScore,
  canAddScore,
  pillarName,
}) => {
  if (score) {
    return (
      <Stack gap="xs">
        <Group gap="sm" wrap="wrap">
          <FitScoreDots score={score.score} />
          <Text size="sm" fw={600}>
            {score.score}/5
          </Text>
          <Text size="xs" c="dimmed">
            {score.scoreLabel}
          </Text>
        </Group>
        {score.rationale && (
          <Text size="xs" c="dimmed" fs="italic">
            "{score.rationale}"
          </Text>
        )}
        {(canEditScore || canDeleteScore) && (
          <Group gap="sm">
            {canEditScore && (
              <Button variant="subtle" size="compact-xs" onClick={onEdit}>
                Edit
              </Button>
            )}
            {canDeleteScore && (
              <Button variant="subtle" color="red" size="compact-xs" onClick={onDelete}>
                Remove
              </Button>
            )}
          </Group>
        )}
      </Stack>
    );
  }

  if (canAddScore) {
    return (
      <Button variant="subtle" size="compact-xs" onClick={onEdit} aria-label={`Add fit score for ${pillarName}`}>
        + Add Score
      </Button>
    );
  }

  return null;
};

const FitScoreRow: React.FC<FitScoreRowProps> = ({
  pillar,
  score,
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
  canEditScore,
  canDeleteScore,
  canAddScore,
}) => (
  <Paper withBorder p="sm" radius="sm" bg="gray.0" data-testid={`fit-score-row-${pillar.id}`}>
    <Stack gap="xs">
      <Box>
        <Text size="sm" fw={500}>
          {pillar.name}
        </Text>
        {pillar.fitCriteria && (
          <Text size="xs" c="dimmed" fs="italic">
            {pillar.fitCriteria}
          </Text>
        )}
      </Box>
      {isEditing ? (
        <FitScoreEditForm
          editScore={editScore}
          editRationale={editRationale}
          onScoreChange={onScoreChange}
          onRationaleChange={onRationaleChange}
          onSave={onSave}
          onCancel={onCancel}
          isSaving={isSaving}
        />
      ) : (
        <FitScoreDisplay
          score={score}
          onEdit={onEdit}
          onDelete={onDelete}
          canEditScore={canEditScore}
          canDeleteScore={canDeleteScore}
          canAddScore={canAddScore}
          pillarName={pillar.name}
        />
      )}
    </Stack>
  </Paper>
);

export const ComponentFitScores: React.FC<ComponentFitScoresProps> = ({ componentId }) => {
  const { data: pillarsConfig } = useStrategyPillarsConfig();
  const { data: fitScoresResponse } = useComponentFitScores(componentId);
  const setFitScoreMutation = useSetFitScore();
  const deleteFitScoreMutation = useDeleteFitScore();

  const [editingPillarId, setEditingPillarId] = useState<string | null>(null);
  const [editScore, setEditScore] = useState<number | null>(null);
  const [editRationale, setEditRationale] = useState('');

  const collectionLinks = fitScoresResponse?._links;

  const enabledPillars = useMemo(() => {
    if (!pillarsConfig?.data) return [];
    return pillarsConfig.data.filter((p) => p.active && p.fitScoringEnabled);
  }, [pillarsConfig]);

  const scoresByPillar = useMemo(() => {
    const fitScores = fitScoresResponse?.data ?? [];
    return new Map(fitScores.map((s) => [s.pillarId, s]));
  }, [fitScoresResponse?.data]);

  const getScoreForPillar = (pillarId: string): ApplicationFitScore | undefined => {
    return scoresByPillar.get(pillarId);
  };

  const handleEdit = (pillar: StrategyPillar) => {
    const existingScore = getScoreForPillar(pillar.id);
    setEditingPillarId(pillar.id);
    setEditScore(existingScore?.score ?? null);
    setEditRationale(existingScore?.rationale ?? '');
  };

  const handleCancel = () => {
    setEditingPillarId(null);
    setEditScore(null);
    setEditRationale('');
  };

  const handleSave = async (pillarId: string) => {
    if (!editScore) return;

    await setFitScoreMutation.mutateAsync({
      componentId,
      pillarId,
      request: {
        score: editScore,
        rationale: editRationale.trim() || undefined,
      },
    });
    handleCancel();
  };

  const handleDelete = async (pillarId: string) => {
    const pillar = enabledPillars.find((p) => p.id === pillarId);
    const confirmed = window.confirm(
      `Are you sure you want to remove the fit score for "${pillar?.name || 'this pillar'}"?`,
    );
    if (!confirmed) return;

    await deleteFitScoreMutation.mutateAsync({
      componentId,
      pillarId,
    });
  };

  if (enabledPillars.length === 0) {
    return null;
  }

  return (
    <Stack gap="xs" mt="xs">
      <Divider />
      <Title order={6} c="dimmed" tt="uppercase">
        Strategic Fit Scores
      </Title>
      <Text size="xs" c="dimmed">
        Rate how well this application supports each strategic pillar
      </Text>
      <Stack gap="xs">
        {enabledPillars.map((pillar) => {
          const score = getScoreForPillar(pillar.id);
          return (
            <FitScoreRow
              key={pillar.id}
              pillar={pillar}
              score={score}
              isEditing={editingPillarId === pillar.id}
              editScore={editScore}
              editRationale={editRationale}
              onEdit={() => handleEdit(pillar)}
              onCancel={handleCancel}
              onScoreChange={setEditScore}
              onRationaleChange={setEditRationale}
              onSave={() => handleSave(pillar.id)}
              onDelete={() => handleDelete(pillar.id)}
              isSaving={setFitScoreMutation.isPending}
              canEditScore={canEdit(score)}
              canDeleteScore={canDelete(score)}
              canAddScore={canCreate({ _links: collectionLinks })}
            />
          );
        })}
      </Stack>
    </Stack>
  );
};

export default ComponentFitScores;
