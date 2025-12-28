import { useState, useEffect, useMemo, useCallback } from 'react';
import { Button } from '@mantine/core';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import {
  useStrategyPillarsConfig,
  useBatchUpdateStrategyPillars,
} from '../../../hooks/useStrategyPillarsSettings';
import { ApiError, type StrategyPillar } from '../../../api/types';
import type { PillarChange } from '../../../api/metadata';
import './StrategyPillarsSettings.css';

interface EditablePillar {
  id: string;
  name: string;
  description: string;
  active: boolean;
  isNew: boolean;
  markedForDeletion: boolean;
}

interface ValidationErrors {
  [index: number]: {
    name?: string;
  };
}

const MAX_PILLARS = 20;

function isConflictError(err: unknown): boolean {
  return err instanceof ApiError && (err.statusCode === 409 || err.statusCode === 412);
}

function buildPillarChanges(
  editedPillars: EditablePillar[],
  originalPillars: StrategyPillar[]
): PillarChange[] {
  const changes: PillarChange[] = [];

  for (const pillar of editedPillars) {
    const change = buildSinglePillarChange(pillar, originalPillars);
    if (change) {
      changes.push(change);
    }
  }

  return changes;
}

function isNewPillarToAdd(pillar: EditablePillar): boolean {
  return pillar.isNew && !pillar.markedForDeletion;
}

function isExistingPillarToRemove(pillar: EditablePillar): boolean {
  return pillar.markedForDeletion && !pillar.isNew;
}

function isExistingPillarToUpdate(pillar: EditablePillar): boolean {
  return !pillar.isNew && pillar.active;
}

function hasPillarChanged(pillar: EditablePillar, original: StrategyPillar | undefined): boolean {
  if (!original) return false;
  return original.name !== pillar.name.trim() || original.description !== pillar.description.trim();
}

function buildSinglePillarChange(
  pillar: EditablePillar,
  originalPillars: StrategyPillar[]
): PillarChange | null {
  if (isNewPillarToAdd(pillar)) {
    return { operation: 'add', name: pillar.name.trim(), description: pillar.description.trim() };
  }

  if (isExistingPillarToRemove(pillar)) {
    return { operation: 'remove', id: pillar.id };
  }

  if (isExistingPillarToUpdate(pillar)) {
    const original = originalPillars.find((p) => p.id === pillar.id);
    if (hasPillarChanged(pillar, original)) {
      return { operation: 'update', id: pillar.id, name: pillar.name.trim(), description: pillar.description.trim() };
    }
  }

  return null;
}

export function StrategyPillarsSettings() {
  const { data: config, isLoading, error, refetch } = useStrategyPillarsConfig();
  const batchUpdateMutation = useBatchUpdateStrategyPillars();

  const [isEditing, setIsEditing] = useState(false);
  const [editedPillars, setEditedPillars] = useState<EditablePillar[]>([]);
  const [validationErrors, setValidationErrors] = useState<ValidationErrors>({});
  const [showRefreshDialog, setShowRefreshDialog] = useState(false);
  const [conflictError, setConflictError] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  const activePillars = useMemo(() => {
    if (!config?.data) return [];
    return config.data.filter((p) => p.active);
  }, [config]);

  useEffect(() => {
    if (config?.data) {
      setEditedPillars(
        config.data.map((p) => ({
          ...p,
          isNew: false,
          markedForDeletion: false,
        }))
      );
    }
  }, [config]);

  const validatePillars = useCallback((pillars: EditablePillar[]): ValidationErrors => {
    const errors: ValidationErrors = {};
    const activePillarNames = new Set<string>();

    pillars.forEach((pillar, index) => {
      if (pillar.markedForDeletion && !pillar.isNew) return;
      if (!pillar.active && !pillar.isNew) return;

      const trimmedName = pillar.name.trim();

      if (!trimmedName) {
        errors[index] = { name: 'Name cannot be empty' };
        return;
      }

      if (trimmedName.length > 100) {
        errors[index] = { name: 'Name must be 100 characters or less' };
        return;
      }

      const lowerName = trimmedName.toLowerCase();
      if (activePillarNames.has(lowerName)) {
        errors[index] = { name: 'Name must be unique' };
        return;
      }
      activePillarNames.add(lowerName);
    });

    return errors;
  }, []);

  const handleEdit = () => {
    setIsEditing(true);
    setConflictError(false);
  };

  const handleCancel = () => {
    if (config?.data) {
      setEditedPillars(
        config.data.map((p) => ({
          ...p,
          isNew: false,
          markedForDeletion: false,
        }))
      );
    }
    setValidationErrors({});
    setIsEditing(false);
    setConflictError(false);
  };

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
        await batchUpdateMutation.mutateAsync({
          request: { changes },
          version: config.version,
        });
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

  const handleNameChange = (index: number, newName: string) => {
    const updated = [...editedPillars];
    updated[index] = { ...updated[index], name: newName };
    setEditedPillars(updated);

    const errors = validatePillars(updated);
    setValidationErrors(errors);
  };

  const handleDescriptionChange = (index: number, newDescription: string) => {
    const updated = [...editedPillars];
    updated[index] = { ...updated[index], description: newDescription };
    setEditedPillars(updated);
  };

  const handleAddPillar = () => {
    const activeCount = editedPillars.filter(
      (p) => (p.active || p.isNew) && !p.markedForDeletion
    ).length;
    if (activeCount >= MAX_PILLARS) return;

    const newPillar: EditablePillar = {
      id: `new-${Date.now()}`,
      name: '',
      description: '',
      active: true,
      isNew: true,
      markedForDeletion: false,
    };

    const updated = [...editedPillars, newPillar];
    setEditedPillars(updated);

    const errors = validatePillars(updated);
    setValidationErrors(errors);
  };

  const handleDeletePillar = (index: number) => {
    const updated = [...editedPillars];
    const pillar = updated[index];

    if (pillar.isNew) {
      updated.splice(index, 1);
    } else {
      updated[index] = { ...pillar, markedForDeletion: true };
    }

    setEditedPillars(updated);

    const errors = validatePillars(updated);
    setValidationErrors(errors);
  };

  const handleRestorePillar = (index: number) => {
    const updated = [...editedPillars];
    updated[index] = { ...updated[index], markedForDeletion: false };
    setEditedPillars(updated);

    const errors = validatePillars(updated);
    setValidationErrors(errors);
  };

  const getActiveCount = () => {
    return editedPillars.filter(
      (p) => (p.active || p.isNew) && !p.markedForDeletion
    ).length;
  };

  const canDelete = (index: number) => {
    const pillar = editedPillars[index];
    if (pillar.markedForDeletion) return false;
    const activeCount = getActiveCount();
    return activeCount > 1;
  };

  if (isLoading) {
    return (
      <div className="strategy-pillars-settings">
        <div className="loading-state">
          <div className="loading-spinner" />
          <p>Loading strategy pillars configuration...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="strategy-pillars-settings">
        <div className="error-message">
          {error instanceof Error ? error.message : 'Failed to load strategy pillars configuration'}
        </div>
      </div>
    );
  }

  const hasValidationErrors = Object.keys(validationErrors).length > 0;
  const displayPillars = isEditing ? editedPillars : activePillars;
  const activeCount = getActiveCount();

  let orderCounter = 0;

  return (
    <div className="strategy-pillars-settings">
      <div className="strategy-pillars-header">
        <div>
          <h2 className="strategy-pillars-title">Strategy Pillars</h2>
          <p className="strategy-pillars-description">
            Define the strategic pillars used to categorize capabilities across your organization.
          </p>
        </div>
        {!isEditing && (
          <div className="strategy-pillars-actions">
            <Button onClick={handleEdit} data-testid="edit-pillars-btn">
              Edit
            </Button>
          </div>
        )}
      </div>

      {conflictError && (
        <div className="conflict-message">
          Configuration was modified by another user. Please refresh and try again.
        </div>
      )}

      <div className="pillars-list">
        {displayPillars.length === 0 && !isEditing && (
          <div className="empty-state" data-testid="empty-pillars-state">
            No strategy pillars configured yet. Click Edit to add pillars.
          </div>
        )}
        {displayPillars.map((pillar, index) => {
          const isMarkedForDeletion = isEditing && 'markedForDeletion' in pillar && pillar.markedForDeletion;
          const shouldShowOrder = !isMarkedForDeletion && (pillar.active || ('isNew' in pillar && pillar.isNew));
          if (shouldShowOrder) orderCounter++;

          return (
            <div
              key={pillar.id}
              className={`pillar-row ${isMarkedForDeletion ? 'pillar-marked-for-deletion' : ''}`}
              data-testid={`pillar-row-${index}`}
            >
              <span className="pillar-order">{shouldShowOrder ? `${orderCounter}.` : ''}</span>
              <div className="pillar-content">
                {isEditing ? (
                  <>
                    <input
                      type="text"
                      className={`pillar-name-input ${validationErrors[index]?.name ? 'input-error' : ''}`}
                      value={'name' in pillar ? pillar.name : ''}
                      onChange={(e) => handleNameChange(index, e.target.value)}
                      placeholder="Pillar name"
                      data-testid={`pillar-name-input-${index}`}
                      maxLength={100}
                      disabled={isMarkedForDeletion}
                    />
                    {validationErrors[index]?.name && (
                      <div className="validation-error" role="alert">
                        {validationErrors[index].name}
                      </div>
                    )}
                    <input
                      type="text"
                      className="pillar-description-input"
                      value={'description' in pillar ? pillar.description : ''}
                      onChange={(e) => handleDescriptionChange(index, e.target.value)}
                      placeholder="Description (optional)"
                      data-testid={`pillar-description-input-${index}`}
                      maxLength={500}
                      disabled={isMarkedForDeletion}
                    />
                  </>
                ) : (
                  <>
                    <span className="pillar-name">{pillar.name}</span>
                    {pillar.description && (
                      <span className="pillar-description-view">{pillar.description}</span>
                    )}
                  </>
                )}
              </div>
              {isEditing && (
                <div className="pillar-actions">
                  {isMarkedForDeletion ? (
                    <button
                      type="button"
                      className="restore-pillar-btn"
                      onClick={() => handleRestorePillar(index)}
                      aria-label={`Restore ${pillar.name}`}
                      data-testid={`restore-pillar-btn-${index}`}
                    >
                      &#8634;
                    </button>
                  ) : (
                    <button
                      type="button"
                      className="delete-pillar-btn"
                      onClick={() => handleDeletePillar(index)}
                      disabled={!canDelete(index)}
                      aria-label={`Delete ${pillar.name}`}
                      data-testid={`delete-pillar-btn-${index}`}
                    >
                      &#128465;
                    </button>
                  )}
                </div>
              )}
            </div>
          );
        })}
      </div>

      {isEditing && (
        <div className="add-pillar-section">
          <button
            type="button"
            className="add-pillar-btn"
            onClick={handleAddPillar}
            disabled={activeCount >= MAX_PILLARS}
            data-testid="add-pillar-btn"
          >
            + Add Pillar
          </button>
          <p className="max-pillars-notice">
            Maximum 20 pillars allowed. Currently {activeCount} of {MAX_PILLARS}.
          </p>
        </div>
      )}

      {isEditing && (
        <div className="edit-actions">
          <Button
            variant="outline"
            onClick={handleCancel}
            disabled={isSaving}
            data-testid="cancel-pillars-btn"
          >
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            disabled={hasValidationErrors || isSaving}
            loading={isSaving}
            data-testid="save-pillars-btn"
          >
            Save Changes
          </Button>
        </div>
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
