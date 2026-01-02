import { useState, useEffect } from 'react';
import { Button } from '@mantine/core';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import { HelpTooltip } from '../../../components/shared/HelpTooltip';
import { useMaturityScale, useUpdateMaturityScale, useResetMaturityScale } from '../../../hooks/useMaturityScale';
import type { MaturityScaleSection } from '../../../api/types';
import { ApiError } from '../../../api/types';
import './MaturityScaleSettings.css';

interface ValidationErrors {
  [key: number]: {
    name?: string;
    boundary?: string;
  };
}

export function MaturityScaleSettings() {
  const { data: config, isLoading, error, refetch } = useMaturityScale();
  const updateMutation = useUpdateMaturityScale();
  const resetMutation = useResetMaturityScale();

  const [isEditing, setIsEditing] = useState(false);
  const [editedSections, setEditedSections] = useState<MaturityScaleSection[]>([]);
  const [validationErrors, setValidationErrors] = useState<ValidationErrors>({});
  const [showResetDialog, setShowResetDialog] = useState(false);
  const [showRefreshDialog, setShowRefreshDialog] = useState(false);
  const [conflictError, setConflictError] = useState(false);

  useEffect(() => {
    if (config) {
      setEditedSections([...config.sections]);
    }
  }, [config]);

  const validateSections = (sections: MaturityScaleSection[]): ValidationErrors => {
    const errors: ValidationErrors = {};

    sections.forEach((section, index) => {
      if (!section.name.trim()) {
        errors[index] = { ...errors[index], name: 'Section name cannot be empty' };
      }
      if (section.name.length > 50) {
        errors[index] = { ...errors[index], name: 'Section name must be 50 characters or less' };
      }

      if (index > 0) {
        const prevSection = sections[index - 1];
        if (section.minValue !== prevSection.maxValue + 1) {
          errors[index] = {
            ...errors[index],
            boundary: 'Sections must be contiguous',
          };
        }
      }
    });

    return errors;
  };

  const handleEdit = () => {
    setIsEditing(true);
    setConflictError(false);
  };

  const handleCancel = () => {
    if (config) {
      setEditedSections([...config.sections]);
    }
    setValidationErrors({});
    setIsEditing(false);
    setConflictError(false);
  };

  const handleSave = async () => {
    if (!config) return;

    const errors = validateSections(editedSections);
    if (Object.keys(errors).length > 0) {
      setValidationErrors(errors);
      return;
    }

    try {
      await updateMutation.mutateAsync({
        sections: editedSections,
        version: config.version,
      });
      setIsEditing(false);
      setValidationErrors({});
      setConflictError(false);
    } catch (err) {
      if (err instanceof ApiError && err.statusCode === 409) {
        setConflictError(true);
        setShowRefreshDialog(true);
      }
    }
  };

  const handleReset = async () => {
    await resetMutation.mutateAsync();
    setShowResetDialog(false);
    setIsEditing(false);
    setValidationErrors({});
    setConflictError(false);
  };

  const handleRefresh = async () => {
    await refetch();
    setShowRefreshDialog(false);
    setIsEditing(false);
    setConflictError(false);
  };

  const handleNameChange = (index: number, newName: string) => {
    const updated = [...editedSections];
    updated[index] = { ...updated[index], name: newName };
    setEditedSections(updated);

    const errors = validateSections(updated);
    setValidationErrors(errors);
  };

  const handleBoundaryChange = (index: number, newEndValue: number) => {
    if (index >= editedSections.length - 1) return;

    const updated = [...editedSections];
    updated[index] = { ...updated[index], maxValue: newEndValue };
    updated[index + 1] = { ...updated[index + 1], minValue: newEndValue + 1 };
    setEditedSections(updated);

    const errors = validateSections(updated);
    setValidationErrors(errors);
  };

  if (isLoading) {
    return (
      <div className="maturity-scale-settings">
        <div className="loading-state">
          <div className="loading-spinner" />
          <p>Loading maturity scale configuration...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="maturity-scale-settings">
        <div className="error-message">
          {error instanceof Error ? error.message : 'Failed to load maturity scale configuration'}
        </div>
      </div>
    );
  }

  if (!config) {
    return null;
  }

  const hasValidationErrors = Object.keys(validationErrors).length > 0;
  const sections = isEditing ? editedSections : config.sections;
  const totalRange = 100;

  return (
    <div className="maturity-scale-settings">
      <div className="maturity-scale-header">
        <div>
          <h2 className="maturity-scale-title">
            Maturity Scale Configuration
            <HelpTooltip
              content="Define how capability maturity is categorized. Each section represents a stage of evolution from experimental (Genesis) to fully commoditized (Commodity)."
              iconOnly
            />
          </h2>
          <p className="maturity-scale-description">
            Configure the names and boundaries of maturity sections (0-99 range).
          </p>
        </div>
        {!isEditing && (
          <div className="maturity-scale-actions">
            {!config.isDefault && (
              <Button
                variant="outline"
                onClick={() => setShowResetDialog(true)}
                disabled={resetMutation.isPending}
              >
                Reset to Defaults
              </Button>
            )}
            <Button onClick={handleEdit}>Edit</Button>
          </div>
        )}
      </div>

      {config.isDefault && (
        <div className="default-badge">
          Using default configuration
        </div>
      )}

      {conflictError && (
        <div className="conflict-message">
          Configuration was modified by another user. Please refresh and try again.
        </div>
      )}

      <div className="maturity-scale-visualization">
        <div className="scale-bar">
          {sections.map((section, index) => {
            const width = ((section.maxValue - section.minValue + 1) / totalRange) * 100;
            return (
              <div
                key={index}
                className="scale-section"
                style={{ width: `${width}%` }}
              >
                {isEditing ? (
                  <div className="scale-section-edit">
                    <input
                      type="text"
                      className={`section-name-input ${validationErrors[index]?.name ? 'input-error' : ''}`}
                      value={section.name}
                      onChange={(e) => handleNameChange(index, e.target.value)}
                      aria-label={`Section ${index + 1} name`}
                      maxLength={50}
                    />
                    <div className="section-range">
                      {section.minValue}-{section.maxValue}
                    </div>
                    {validationErrors[index]?.name && (
                      <div className="validation-error" role="alert">
                        {validationErrors[index].name}
                      </div>
                    )}
                  </div>
                ) : (
                  <div className="scale-section-view">
                    <div className="section-name">{section.name}</div>
                    <div className="section-range">
                      {section.minValue}-{section.maxValue}
                    </div>
                  </div>
                )}
              </div>
            );
          })}
        </div>

        {isEditing && (
          <div className="boundary-controls">
            {sections.map((section, index) => {
              const isLastSection = index === sections.length - 1;
              return (
                <div
                  key={index}
                  className="boundary-control-slot"
                >
                  {!isLastSection && (
                    <div className="boundary-control">
                      <label className="boundary-label">
                        End of {section.name}:
                      </label>
                      <input
                        type="number"
                        className="boundary-input"
                        min={section.minValue + 1}
                        max={sections[index + 1].maxValue - 1}
                        value={section.maxValue}
                        onChange={(e) => handleBoundaryChange(index, parseInt(e.target.value) || section.maxValue)}
                        aria-label={`End boundary for ${section.name}`}
                      />
                      {validationErrors[index + 1]?.boundary && (
                        <div className="validation-error" role="alert">
                          {validationErrors[index + 1].boundary}
                        </div>
                      )}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>

      {isEditing && (
        <div className="edit-actions">
          <Button
            variant="outline"
            onClick={handleCancel}
            disabled={updateMutation.isPending}
          >
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            disabled={hasValidationErrors || updateMutation.isPending}
            loading={updateMutation.isPending}
          >
            Save Changes
          </Button>
        </div>
      )}

      {showResetDialog && (
        <ConfirmationDialog
          title="Reset to Default Configuration"
          message="Are you sure you want to reset the maturity scale to default values (Genesis, Custom Built, Product, Commodity with equal ranges)?"
          confirmText="Reset"
          cancelText="Cancel"
          onConfirm={handleReset}
          onCancel={() => setShowResetDialog(false)}
          isLoading={resetMutation.isPending}
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
