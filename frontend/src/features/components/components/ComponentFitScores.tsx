import React, { useState, useMemo } from 'react';
import { useComponentFitScores, useSetFitScore, useDeleteFitScore } from '../hooks/useFitScores';
import { useStrategyPillarsConfig } from '../../../hooks/useStrategyPillarsSettings';
import { canEdit, canDelete, canCreate } from '../../../utils/hateoas';
import type { ComponentId, StrategyPillar, ApplicationFitScore } from '../../../api/types';
import './ComponentFitScores.css';

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
// Note: Labels must match backend FitScore.GetLabel() for consistency.
// Used only during editing; API response scoreLabel is authoritative for display.

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
}) => {
  return (
    <div className="fit-score-row" data-testid={`fit-score-row-${pillar.id}`}>
      <div className="fit-score-pillar">
        <span className="fit-score-pillar-name">{pillar.name}</span>
        {pillar.fitCriteria && (
          <span className="fit-score-criteria">{pillar.fitCriteria}</span>
        )}
      </div>
      {isEditing ? (
        <div className="fit-score-edit">
          <div className="fit-score-selector">
            {SCORE_RANGE.map((s) => (
              <button
                key={s}
                type="button"
                className={`fit-score-btn ${editScore === s ? 'selected' : ''}`}
                onClick={() => onScoreChange(s)}
                disabled={isSaving}
                data-testid={`fit-score-btn-${s}`}
              >
                {s}
              </button>
            ))}
          </div>
          <span className="fit-score-label">{editScore ? SCORE_LABELS[editScore] : 'Select score'}</span>
          <textarea
            className="fit-score-rationale-input"
            placeholder="Rationale (optional)"
            value={editRationale}
            onChange={(e) => onRationaleChange(e.target.value)}
            maxLength={500}
            disabled={isSaving}
            data-testid="fit-score-rationale-input"
          />
          <div className="fit-score-edit-actions">
            <button
              type="button"
              className="btn btn-secondary btn-small"
              onClick={onCancel}
              disabled={isSaving}
            >
              Cancel
            </button>
            <button
              type="button"
              className="btn btn-primary btn-small"
              onClick={onSave}
              disabled={!editScore || isSaving}
            >
              {isSaving ? 'Saving...' : 'Save'}
            </button>
          </div>
        </div>
      ) : (
        <div className="fit-score-display">
          {score ? (
            <>
              <div className="fit-score-value">
                <span className="fit-score-dots">
                  {SCORE_RANGE.map((s) => (
                    <span
                      key={s}
                      className={`fit-score-dot ${s <= score.score ? 'filled' : ''}`}
                    />
                  ))}
                </span>
                <span className="fit-score-number">{score.score}/5</span>
                <span className="fit-score-label">{score.scoreLabel}</span>
              </div>
              {score.rationale && (
                <span className="fit-score-rationale">"{score.rationale}"</span>
              )}
              {(canEditScore || canDeleteScore) && (
                <div className="fit-score-actions">
                  {canEditScore && (
                    <button
                      type="button"
                      className="btn btn-link btn-small"
                      onClick={onEdit}
                    >
                      Edit
                    </button>
                  )}
                  {canDeleteScore && (
                    <button
                      type="button"
                      className="btn btn-link btn-small btn-danger"
                      onClick={onDelete}
                    >
                      Remove
                    </button>
                  )}
                </div>
              )}
            </>
          ) : canAddScore ? (
            <button
              type="button"
              className="btn btn-link btn-small"
              onClick={onEdit}
              aria-label={`Add fit score for ${pillar.name}`}
            >
              + Add Score
            </button>
          ) : null}
        </div>
      )}
    </div>
  );
};

export const ComponentFitScores: React.FC<ComponentFitScoresProps> = ({ componentId }) => {
  const { data: pillarsConfig } = useStrategyPillarsConfig();
  const { data: fitScoresResponse } = useComponentFitScores(componentId);
  const setFitScoreMutation = useSetFitScore();
  const deleteFitScoreMutation = useDeleteFitScore();

  const [editingPillarId, setEditingPillarId] = useState<string | null>(null);
  const [editScore, setEditScore] = useState<number | null>(null);
  const [editRationale, setEditRationale] = useState('');

  const fitScores = fitScoresResponse?.data ?? [];
  const collectionLinks = fitScoresResponse?._links;

  const enabledPillars = useMemo(() => {
    if (!pillarsConfig?.data) return [];
    return pillarsConfig.data.filter((p) => p.active && p.fitScoringEnabled);
  }, [pillarsConfig]);

  const scoresByPillar = useMemo(() => {
    return new Map(fitScores.map((s) => [s.pillarId, s]));
  }, [fitScores]);

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
      `Are you sure you want to remove the fit score for "${pillar?.name || 'this pillar'}"?`
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
    <div className="component-fit-scores">
      <h4 className="fit-scores-title">Strategic Fit Scores</h4>
      <p className="fit-scores-description">
        Rate how well this application supports each strategic pillar
      </p>
      <div className="fit-scores-list">
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
      </div>
    </div>
  );
};

export default ComponentFitScores;
