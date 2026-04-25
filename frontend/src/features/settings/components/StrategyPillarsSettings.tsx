import { Button } from '@mantine/core';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import { HelpTooltip } from '../../../components/shared/HelpTooltip';
import { useBatchUpdateStrategyPillars, useStrategyPillarsConfig } from '../../../hooks/useStrategyPillarsSettings';
import {
  buildPillarChanges,
  countActive,
  deleteOrMarkAt,
  type EditablePillar,
  emptyEditablePillar,
  isConflictError,
  MAX_PILLARS,
  patchPillarAt,
  toEditable,
  type ValidationErrors,
  validatePillars,
} from './pillarChanges';
import { type PillarHandlers, PillarsList } from './PillarsList';
import './StrategyPillarsSettings.css';

export function StrategyPillarsSettings() {
  const { data: config, isLoading, error, refetch } = useStrategyPillarsConfig();
  const batchUpdate = useBatchUpdateStrategyPillars();

  const [isEditing, setIsEditing] = useState(false);
  const [editedPillars, setEditedPillars] = useState<EditablePillar[]>([]);
  const [validationErrors, setValidationErrors] = useState<ValidationErrors>({});
  const [showRefreshDialog, setShowRefreshDialog] = useState(false);
  const [conflictError, setConflictError] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  const activePillars = useMemo(() => config?.data?.filter((p) => p.active) ?? [], [config]);

  useEffect(() => {
    if (config?.data) setEditedPillars(toEditable(config.data));
  }, [config]);

  const mutate = useCallback((mutator: (prev: EditablePillar[]) => EditablePillar[], revalidate: boolean) => {
    setEditedPillars((prev) => {
      const next = mutator(prev);
      if (revalidate) setValidationErrors(validatePillars(next));
      return next;
    });
  }, []);

  const handlers: PillarHandlers = useMemo(
    () => ({
      onNameChange: (i, name) => mutate((p) => patchPillarAt(p, i, { name }), true),
      onDescriptionChange: (i, description) => mutate((p) => patchPillarAt(p, i, { description }), false),
      onFitScoringEnabledChange: (i, fitScoringEnabled) => mutate((p) => patchPillarAt(p, i, { fitScoringEnabled }), false),
      onFitCriteriaChange: (i, fitCriteria) => mutate((p) => patchPillarAt(p, i, { fitCriteria }), false),
      onFitTypeChange: (i, fitType) => mutate((p) => patchPillarAt(p, i, { fitType }), false),
      onDelete: (i) => mutate((p) => deleteOrMarkAt(p, i), true),
      onRestore: (i) => mutate((p) => patchPillarAt(p, i, { markedForDeletion: false }), true),
    }),
    [mutate],
  );

  const handleEdit = () => {
    setIsEditing(true);
    setConflictError(false);
  };

  const handleCancel = () => {
    if (config?.data) setEditedPillars(toEditable(config.data));
    setValidationErrors({});
    setIsEditing(false);
    setConflictError(false);
  };

  const handleAddPillar = () =>
    mutate((prev) => (countActive(prev) >= MAX_PILLARS ? prev : [...prev, emptyEditablePillar()]), true);

  const handleSave = async () => {
    if (!config) return;
    const errors = validatePillars(editedPillars);
    if (Object.keys(errors).length > 0) {
      setValidationErrors(errors);
      return;
    }
    setIsSaving(true);
    try {
      const changes = buildPillarChanges(editedPillars, config.data);
      if (changes.length > 0) {
        await batchUpdate.mutateAsync({ request: { changes }, version: config.version });
      }
      setIsEditing(false);
      setValidationErrors({});
      setConflictError(false);
    } catch (err) {
      if (isConflictError(err)) {
        setConflictError(true);
        setShowRefreshDialog(true);
      }
    } finally {
      setIsSaving(false);
    }
  };

  const handleRefresh = async () => {
    await refetch();
    setShowRefreshDialog(false);
    setIsEditing(false);
    setConflictError(false);
  };

  if (isLoading) return <LoadingState />;
  if (error) return <ErrorState error={error} />;

  const activeCount = countActive(editedPillars);
  const hasErrors = Object.keys(validationErrors).length > 0;

  return (
    <div className="strategy-pillars-settings">
      <Header isEditing={isEditing} onEdit={handleEdit} />
      {conflictError && (
        <div className="conflict-message">
          Configuration was modified by another user. Please refresh and try again.
        </div>
      )}
      <PillarsList
        pillars={isEditing ? editedPillars : activePillars}
        isEditing={isEditing}
        validationErrors={validationErrors}
        activeCount={activeCount}
        handlers={handlers}
      />
      {isEditing && (
        <EditFooter
          activeCount={activeCount}
          disabled={hasErrors || isSaving}
          isSaving={isSaving}
          onAdd={handleAddPillar}
          onCancel={handleCancel}
          onSave={handleSave}
        />
      )}
      {showRefreshDialog && (
        <ConfirmationDialog
          title="Configuration Conflict"
          message="The configuration was modified by another user. Please refresh to see the latest version and try again."
          confirmText="Refresh"
          cancelText="Cancel"
          onConfirm={handleRefresh}
          onCancel={() => setShowRefreshDialog(false)}
        />
      )}
    </div>
  );
}

function Header({ isEditing, onEdit }: { isEditing: boolean; onEdit: () => void }) {
  return (
    <div className="strategy-pillars-header">
      <div>
        <h2 className="strategy-pillars-title">
          Strategy Pillars
          <HelpTooltip
            content="Strategic pillars represent your organization's key strategic themes. Use them to align capabilities with business strategy and measure strategic fit."
            iconOnly
          />
        </h2>
        <p className="strategy-pillars-description">
          Define the strategic pillars used to categorize capabilities across your organization.
        </p>
      </div>
      {!isEditing && (
        <div className="strategy-pillars-actions">
          <Button onClick={onEdit} data-testid="edit-pillars-btn">
            Edit
          </Button>
        </div>
      )}
    </div>
  );
}

interface EditFooterProps {
  activeCount: number;
  disabled: boolean;
  isSaving: boolean;
  onAdd: () => void;
  onCancel: () => void;
  onSave: () => void;
}

function EditFooter({ activeCount, disabled, isSaving, onAdd, onCancel, onSave }: EditFooterProps) {
  return (
    <>
      <div className="add-pillar-section">
        <button
          type="button"
          className="add-pillar-btn"
          onClick={onAdd}
          disabled={activeCount >= MAX_PILLARS}
          data-testid="add-pillar-btn"
        >
          + Add Pillar
        </button>
        <p className="max-pillars-notice">
          Maximum 20 pillars allowed. Currently {activeCount} of {MAX_PILLARS}.
        </p>
      </div>
      <div className="edit-actions">
        <Button variant="outline" onClick={onCancel} disabled={isSaving} data-testid="cancel-pillars-btn">
          Cancel
        </Button>
        <Button onClick={onSave} disabled={disabled} loading={isSaving} data-testid="save-pillars-btn">
          Save Changes
        </Button>
      </div>
    </>
  );
}

function LoadingState() {
  return (
    <div className="strategy-pillars-settings">
      <div className="loading-state">
        <div className="loading-spinner" />
        <p>Loading strategy pillars configuration...</p>
      </div>
    </div>
  );
}

function ErrorState({ error }: { error: unknown }) {
  return (
    <div className="strategy-pillars-settings">
      <div className="error-message">
        {error instanceof Error ? error.message : 'Failed to load strategy pillars configuration'}
      </div>
    </div>
  );
}
