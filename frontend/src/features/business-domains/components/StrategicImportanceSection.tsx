import React, { useState, useMemo } from 'react';
import type { BusinessDomain, CapabilityId, StrategyImportance } from '../../../api/types';
import {
  useStrategyImportanceByDomainAndCapability,
  useSetStrategyImportance,
  useUpdateStrategyImportance,
  useRemoveStrategyImportance,
} from '../hooks/useStrategyImportance';
import { useStrategyPillarsConfig } from '../../../hooks/useStrategyPillarsSettings';
import { canCreate } from '../../../utils/hateoas';
import '../../../features/components/components/ComponentFitScores.css';

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

interface ImportanceEditFormProps {
  editScore: number | null;
  editRationale: string;
  onScoreChange: (score: number) => void;
  onRationaleChange: (rationale: string) => void;
  onCancel: () => void;
  onSave: () => void;
  isSaving: boolean;
}

function ImportanceEditForm({ editScore, editRationale, onScoreChange, onRationaleChange, onCancel, onSave, isSaving }: ImportanceEditFormProps) {
  return (
    <div className="fit-score-edit">
      <div className="fit-score-selector">
        {SCORE_RANGE.map((s) => (
          <button
            key={s}
            type="button"
            className={`fit-score-btn ${editScore === s ? 'selected' : ''}`}
            onClick={() => onScoreChange(s)}
            disabled={isSaving}
            data-testid={`importance-btn-${s}`}
          >
            {s}
          </button>
        ))}
      </div>
      <span className="fit-score-label">{editScore ? IMPORTANCE_LABELS[editScore] : 'Select importance'}</span>
      <textarea
        className="fit-score-rationale-input"
        placeholder="Rationale (optional)"
        value={editRationale}
        onChange={(e) => onRationaleChange(e.target.value)}
        maxLength={500}
        disabled={isSaving}
        data-testid="importance-rationale-input"
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

function ImportanceDisplay({ pillarId, pillarName, importance, canAddImportance, onEdit, onDelete }: ImportanceDisplayProps) {
  if (importance) {
    return (
      <div className="fit-score-display">
        <div className="fit-score-value">
          <span className="fit-score-dots">
            {SCORE_RANGE.map((s) => (
              <span
                key={s}
                className={`fit-score-dot ${s <= importance.importance ? 'filled' : ''}`}
              />
            ))}
          </span>
          <span className="fit-score-number">{importance.importance}/5</span>
          <span className="fit-score-label">{importance.importanceLabel}</span>
        </div>
        {importance.rationale && (
          <span className="fit-score-rationale">"{importance.rationale}"</span>
        )}
        <div className="fit-score-actions">
          {importance._links?.edit && (
            <button
              type="button"
              className="btn btn-link btn-small"
              onClick={onEdit}
              data-testid={`edit-importance-${pillarId}`}
            >
              Edit
            </button>
          )}
          {importance._links?.delete && (
            <button
              type="button"
              className="btn btn-link btn-small btn-danger"
              onClick={onDelete}
            >
              Remove
            </button>
          )}
        </div>
      </div>
    );
  }

  if (canAddImportance) {
    return (
      <div className="fit-score-display">
        <button
          type="button"
          className="btn btn-link btn-small"
          onClick={onEdit}
          aria-label={`Add importance for ${pillarName}`}
          data-testid="add-importance-btn"
        >
          + Add Importance
        </button>
      </div>
    );
  }

  return <div className="fit-score-display" />;
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
    <div className="fit-score-row" data-testid={`importance-row-${pillar.id}`}>
      <div className="fit-score-pillar">
        <span className="fit-score-pillar-name">{pillar.name}</span>
        {pillar.description && (
          <span className="fit-score-criteria">{pillar.description}</span>
        )}
      </div>
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
    </div>
  );
};

export function StrategicImportanceSection({ domain, capabilityId }: StrategicImportanceSectionProps) {
  const { data: importanceResponse, isLoading } = useStrategyImportanceByDomainAndCapability(
    domain.id,
    capabilityId
  );
  const { data: pillarsConfig } = useStrategyPillarsConfig();
  const setImportanceMutation = useSetStrategyImportance();
  const updateImportanceMutation = useUpdateStrategyImportance();
  const removeImportanceMutation = useRemoveStrategyImportance();

  const [editingPillarId, setEditingPillarId] = useState<string | null>(null);
  const [editScore, setEditScore] = useState<number | null>(null);
  const [editRationale, setEditRationale] = useState('');

  const activePillars = useMemo(() => {
    if (!pillarsConfig?.data) return [];
    return pillarsConfig.data.filter((p) => p.active);
  }, [pillarsConfig]);

  const importanceRatings = useMemo(() => importanceResponse?.data ?? [], [importanceResponse?.data]);
  const canAddImportance = useMemo(
    () => canCreate({ _links: importanceResponse?._links }),
    [importanceResponse?._links]
  );

  const importanceByPillar = useMemo(() => {
    return new Map(importanceRatings.map((r) => [r.pillarId, r]));
  }, [importanceRatings]);

  const getImportanceForPillar = (pillarId: string): StrategyImportance | undefined => {
    return importanceByPillar.get(pillarId);
  };

  const handleEdit = (pillar: StrategyPillar) => {
    const existing = getImportanceForPillar(pillar.id);
    setEditingPillarId(pillar.id);
    setEditScore(existing?.importance ?? null);
    setEditRationale(existing?.rationale ?? '');
  };

  const handleCancel = () => {
    setEditingPillarId(null);
    setEditScore(null);
    setEditRationale('');
  };

  const handleSave = async (pillarId: string) => {
    if (!editScore) return;

    const existing = getImportanceForPillar(pillarId);

    if (existing) {
      await updateImportanceMutation.mutateAsync({
        domainId: domain.id,
        capabilityId,
        importanceId: existing.id,
        request: {
          importance: editScore,
          rationale: editRationale.trim() || undefined,
        },
      });
    } else {
      await setImportanceMutation.mutateAsync({
        domainId: domain.id,
        capabilityId,
        request: {
          pillarId,
          importance: editScore,
          rationale: editRationale.trim() || undefined,
        },
      });
    }
    handleCancel();
  };

  const handleDelete = async (pillarId: string) => {
    const existing = getImportanceForPillar(pillarId);
    if (!existing) return;

    const pillar = activePillars.find((p) => p.id === pillarId);
    const confirmed = window.confirm(
      `Are you sure you want to remove the importance rating for "${pillar?.name || 'this pillar'}"?`
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
    <div className="component-fit-scores">
      <h4 className="fit-scores-title">Strategic Importance</h4>
      <p className="fit-scores-description">
        Rate how important this capability is for {domain.name}
      </p>
      {isLoading ? (
        <div style={{ color: 'var(--color-gray-500)', fontSize: '0.875rem' }}>Loading...</div>
      ) : (
        <div className="fit-scores-list">
          {activePillars.map((pillar) => (
            <ImportanceRow
              key={pillar.id}
              pillar={pillar}
              importance={getImportanceForPillar(pillar.id)}
              canAddImportance={canAddImportance}
              isEditing={editingPillarId === pillar.id}
              editScore={editScore}
              editRationale={editRationale}
              onEdit={() => handleEdit(pillar)}
              onCancel={handleCancel}
              onScoreChange={setEditScore}
              onRationaleChange={setEditRationale}
              onSave={() => handleSave(pillar.id)}
              onDelete={() => handleDelete(pillar.id)}
              isSaving={isSaving}
            />
          ))}
        </div>
      )}
    </div>
  );
}
