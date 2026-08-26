import { Box, Button, Group, NumberInput, Paper, Stack, Text, TextInput } from '@mantine/core';
import { useLayoutEffect, useState } from 'react';
import type { MaturityScaleSection } from '../../../api/types';
import { ApiError } from '../../../api/types';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import { useMaturityScale, useResetMaturityScale, useUpdateMaturityScale } from '../../../hooks/useMaturityScale';
import classes from './MaturityScaleSettings.module.css';
import {
  SettingsConflictNotice,
  SettingsSection,
  SettingsSectionError,
  SettingsSectionFooter,
  SettingsSectionHeader,
  SettingsSectionLoading,
} from './SettingsSection';

interface ValidationErrors {
  [key: number]: {
    name?: string;
    boundary?: string;
  };
}

const TOTAL_RANGE = 100;

function validateSections(sections: MaturityScaleSection[]): ValidationErrors {
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
}

function sectionWidthPercent(section: MaturityScaleSection): number {
  return ((section.maxValue - section.minValue + 1) / TOTAL_RANGE) * 100;
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

  useLayoutEffect(() => {
    if (config) {
      const next = [...config.sections];
      if (JSON.stringify(editedSections) !== JSON.stringify(next)) {
        queueMicrotask(() => setEditedSections(next));
      }
    }
  }, [config, editedSections]);

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

  if (isLoading) return <SettingsSectionLoading message="Loading maturity scale configuration..." />;
  if (error) return <SettingsSectionError error={error} fallback="Failed to load maturity scale configuration" />;
  if (!config) return null;

  const hasValidationErrors = Object.keys(validationErrors).length > 0;
  const sections = isEditing ? editedSections : config.sections;

  return (
    <SettingsSection>
      <SettingsSectionHeader
        title="Maturity Scale Configuration"
        description="Configure the names and boundaries of maturity sections (0-99 range)."
        help="Define how capability maturity is categorized. Each section represents a stage of evolution from experimental (Genesis) to fully commoditized (Commodity)."
        actions={
          !isEditing && (
            <Group gap="sm">
              {!config.isDefault && (
                <Button variant="outline" onClick={() => setShowResetDialog(true)} disabled={resetMutation.isPending}>
                  Reset to Defaults
                </Button>
              )}
              <Button onClick={handleEdit}>Edit</Button>
            </Group>
          )
        }
      />

      {config.isDefault && <Text className={classes.defaultBadge}>Using default configuration</Text>}

      {conflictError && <SettingsConflictNotice />}

      <ScaleBar
        sections={sections}
        isEditing={isEditing}
        validationErrors={validationErrors}
        onNameChange={handleNameChange}
      />

      {isEditing && (
        <BoundaryControls
          sections={sections}
          validationErrors={validationErrors}
          onBoundaryChange={handleBoundaryChange}
        />
      )}

      {isEditing && (
        <SettingsSectionFooter>
          <Button variant="outline" onClick={handleCancel} disabled={updateMutation.isPending}>
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            disabled={hasValidationErrors || updateMutation.isPending}
            loading={updateMutation.isPending}
          >
            Save Changes
          </Button>
        </SettingsSectionFooter>
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
    </SettingsSection>
  );
}

interface ScaleBarProps {
  sections: MaturityScaleSection[];
  isEditing: boolean;
  validationErrors: ValidationErrors;
  onNameChange: (index: number, name: string) => void;
}

function ScaleBar({ sections, isEditing, validationErrors, onNameChange }: ScaleBarProps) {
  return (
    <div className={classes.scaleBar}>
      {sections.map((section, index) => (
        <div key={section.order} className={classes.scaleSection} style={{ width: `${sectionWidthPercent(section)}%` }}>
          <Stack align="center" gap="sm" w="100%">
            {isEditing ? (
              <TextInput
                w="100%"
                size="xs"
                classNames={{ input: classes.centeredInput }}
                value={section.name}
                onChange={(e) => onNameChange(index, e.currentTarget.value)}
                aria-label={`Section ${index + 1} name`}
                maxLength={50}
                error={validationErrors[index]?.name}
              />
            ) : (
              <Text fw={600} c="gray.8" ta="center">
                {section.name}
              </Text>
            )}
            <Text size="xs" c="dimmed" fw={500}>
              {section.minValue}-{section.maxValue}
            </Text>
          </Stack>
        </div>
      ))}
    </div>
  );
}

interface BoundaryControlsProps {
  sections: MaturityScaleSection[];
  validationErrors: ValidationErrors;
  onBoundaryChange: (index: number, endValue: number) => void;
}

function BoundaryControls({ sections, validationErrors, onBoundaryChange }: BoundaryControlsProps) {
  return (
    <Paper bg="gray.0" p="md" radius="md">
      <Group gap="md" align="flex-start">
        {sections.map((section, index) => {
          const isLastSection = index === sections.length - 1;
          return (
            <Box key={section.order} flex="1 1 auto" className={classes.boundarySlot}>
              {!isLastSection && (
                <Stack gap="xs">
                  <NumberInput
                    label={`End of ${section.name}:`}
                    classNames={{ input: classes.centeredInput }}
                    min={section.minValue + 1}
                    max={sections[index + 1].maxValue - 1}
                    value={section.maxValue}
                    onChange={(v) => onBoundaryChange(index, typeof v === 'number' ? v : section.maxValue)}
                    aria-label={`End boundary for ${section.name}`}
                    size="xs"
                  />
                  {validationErrors[index + 1]?.boundary && (
                    <Text size="xs" ta="center" className={classes.validationError} role="alert">
                      {validationErrors[index + 1].boundary}
                    </Text>
                  )}
                </Stack>
              )}
            </Box>
          );
        })}
      </Group>
    </Paper>
  );
}
