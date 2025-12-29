import React, { useState, useEffect } from 'react';
import { Modal, Button, Group, Stack, Alert, Textarea, Radio, Text, Slider } from '@mantine/core';
import { useStrategyPillarsConfig } from '../../../hooks/useStrategyPillarsSettings';
import {
  useSetStrategyImportance,
  useUpdateStrategyImportance,
  useRemoveStrategyImportance,
} from '../hooks/useStrategyImportance';
import type {
  BusinessDomainId,
  CapabilityId,
  StrategyImportance,
  StrategyImportanceId,
} from '../../../api/types';

interface SetImportanceDialogProps {
  isOpen: boolean;
  onClose: () => void;
  domainId: BusinessDomainId;
  domainName: string;
  capabilityId: CapabilityId;
  capabilityName: string;
  existingImportance?: StrategyImportance;
  existingPillarIds?: string[];
}

interface FormState {
  pillarId: string;
  importance: number;
  rationale: string;
}

interface StrategyPillar {
  id: string;
  name: string;
  description?: string;
  active: boolean;
}

const IMPORTANCE_LABELS: Record<number, string> = {
  1: 'Low',
  2: 'Below Average',
  3: 'Average',
  4: 'Above Average',
  5: 'Critical',
};

const IMPORTANCE_MARKS = [
  { value: 1, label: 'Low' },
  { value: 3, label: 'Average' },
  { value: 5, label: 'Critical' },
];

const DEFAULT_FORM_STATE: FormState = { pillarId: '', importance: 3, rationale: '' };

function getInitialFormState(existingImportance?: StrategyImportance): FormState {
  if (!existingImportance) return DEFAULT_FORM_STATE;
  return {
    pillarId: existingImportance.pillarId,
    importance: existingImportance.importance,
    rationale: existingImportance.rationale || '',
  };
}

function getPillarDisplayName(
  existingImportance: StrategyImportance | undefined,
  activePillars: StrategyPillar[]
): string {
  if (existingImportance?.pillarName) return existingImportance.pillarName;
  return activePillars.find((p) => p.id === existingImportance?.pillarId)?.name || '';
}

function getAvailablePillars(
  activePillars: StrategyPillar[],
  existingPillarIds: string[],
  isEditMode: boolean
): StrategyPillar[] {
  return isEditMode ? activePillars : activePillars.filter((p) => !existingPillarIds.includes(p.id));
}

export const SetImportanceDialog: React.FC<SetImportanceDialogProps> = ({
  isOpen,
  onClose,
  domainId,
  domainName,
  capabilityId,
  capabilityName,
  existingImportance,
  existingPillarIds = [],
}) => {
  const isEditMode = !!existingImportance;
  const { data: pillarsConfig } = useStrategyPillarsConfig();
  const setImportanceMutation = useSetStrategyImportance();
  const updateImportanceMutation = useUpdateStrategyImportance();
  const removeImportanceMutation = useRemoveStrategyImportance();

  const [form, setForm] = useState<FormState>(DEFAULT_FORM_STATE);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen) {
      setForm(getInitialFormState(existingImportance));
      setError(null);
    }
  }, [isOpen, existingImportance]);

  const activePillars = pillarsConfig?.data.filter((p) => p.active) || [];
  const availablePillars = getAvailablePillars(activePillars, existingPillarIds, isEditMode);
  const existingPillarName = getPillarDisplayName(existingImportance, activePillars);

  const handleClose = () => {
    setForm(DEFAULT_FORM_STATE);
    setError(null);
    onClose();
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!form.pillarId && !isEditMode) {
      setError('Please select a strategy pillar');
      return;
    }

    try {
      if (isEditMode && existingImportance) {
        await updateImportanceMutation.mutateAsync({
          domainId,
          capabilityId,
          importanceId: existingImportance.id,
          request: { importance: form.importance, rationale: form.rationale || undefined },
        });
      } else {
        await setImportanceMutation.mutateAsync({
          domainId,
          capabilityId,
          request: { pillarId: form.pillarId, importance: form.importance, rationale: form.rationale || undefined },
        });
      }
      handleClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save importance');
    }
  };

  const handleRemove = async () => {
    if (!existingImportance) return;
    setError(null);

    try {
      await removeImportanceMutation.mutateAsync({
        domainId,
        capabilityId,
        importanceId: existingImportance.id as StrategyImportanceId,
      });
      handleClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to remove importance');
    }
  };

  const isPending =
    setImportanceMutation.isPending || updateImportanceMutation.isPending || removeImportanceMutation.isPending;
  const selectedPillarName = activePillars.find((p) => p.id === form.pillarId)?.name;
  const pillarLabel = isEditMode ? existingPillarName : selectedPillarName || 'this pillar';

  return (
    <Modal
      opened={isOpen}
      onClose={handleClose}
      title={isEditMode ? 'Edit Strategic Importance' : 'Set Strategic Importance'}
      centered
      size="md"
      data-testid="set-importance-dialog"
      styles={{ body: { overflow: 'hidden' } }}
    >
      <form onSubmit={handleSubmit}>
        <Stack gap="md">
          <LabeledField label="Capability" value={capabilityName} />
          <LabeledField label="Domain" value={domainName} />

          {!isEditMode && (
            <PillarSelector
              availablePillars={availablePillars}
              selectedPillarId={form.pillarId}
              onChange={(pillarId) => setForm((prev) => ({ ...prev, pillarId }))}
            />
          )}

          {isEditMode && <LabeledField label="Strategy Pillar" value={existingPillarName} />}

          <ImportanceSlider
            value={form.importance}
            onChange={(importance) => setForm((prev) => ({ ...prev, importance }))}
            pillarLabel={pillarLabel}
            disabled={isPending}
          />

          <Textarea
            label="Why? (optional)"
            placeholder="Explain why this capability has this importance level..."
            value={form.rationale}
            onChange={(e) => setForm((prev) => ({ ...prev, rationale: e.currentTarget.value }))}
            maxLength={500}
            minRows={2}
            maxRows={4}
            disabled={isPending}
            data-testid="rationale-input"
          />

          {error && (
            <Alert color="red" data-testid="importance-error">
              {error}
            </Alert>
          )}

          <DialogActions
            isEditMode={isEditMode}
            isPending={isPending}
            canSubmit={!!form.pillarId || isEditMode}
            onClose={handleClose}
            onRemove={handleRemove}
            isRemoving={removeImportanceMutation.isPending}
            isSaving={setImportanceMutation.isPending || updateImportanceMutation.isPending}
          />
        </Stack>
      </form>
    </Modal>
  );
};

interface LabeledFieldProps {
  label: string;
  value: string;
}

const LabeledField: React.FC<LabeledFieldProps> = ({ label, value }) => (
  <div>
    <Text size="sm" c="dimmed">
      {label}
    </Text>
    <Text fw={500}>{value}</Text>
  </div>
);

interface PillarSelectorProps {
  availablePillars: StrategyPillar[];
  selectedPillarId: string;
  onChange: (pillarId: string) => void;
}

const PillarSelector: React.FC<PillarSelectorProps> = ({ availablePillars, selectedPillarId, onChange }) => (
  <Radio.Group label="Strategy Pillar" value={selectedPillarId} onChange={onChange} required withAsterisk>
    <Stack gap="xs" mt="xs">
      {availablePillars.length === 0 ? (
        <Text size="sm" c="dimmed" fs="italic">
          All active pillars have been rated for this capability
        </Text>
      ) : (
        availablePillars.map((pillar) => (
          <Radio
            key={pillar.id}
            value={pillar.id}
            label={pillar.name}
            description={pillar.description}
            data-testid={`pillar-option-${pillar.id}`}
          />
        ))
      )}
    </Stack>
  </Radio.Group>
);

interface ImportanceSliderProps {
  value: number;
  onChange: (value: number) => void;
  pillarLabel: string;
  disabled: boolean;
}

const ImportanceSlider: React.FC<ImportanceSliderProps> = ({ value, onChange, pillarLabel, disabled }) => (
  <div>
    <Text size="sm" fw={500} mb="xs">
      How important is this capability for &quot;{pillarLabel}&quot;?
    </Text>
    <Slider
      value={value}
      onChange={onChange}
      min={1}
      max={5}
      step={1}
      marks={IMPORTANCE_MARKS}
      label={(v) => IMPORTANCE_LABELS[v]}
      disabled={disabled}
      data-testid="importance-slider"
      px="md"
      mb="lg"
    />
  </div>
);

interface DialogActionsProps {
  isEditMode: boolean;
  isPending: boolean;
  canSubmit: boolean;
  onClose: () => void;
  onRemove: () => void;
  isRemoving: boolean;
  isSaving: boolean;
}

const DialogActions: React.FC<DialogActionsProps> = ({
  isEditMode,
  isPending,
  canSubmit,
  onClose,
  onRemove,
  isRemoving,
  isSaving,
}) => (
  <Group justify="space-between">
    {isEditMode ? (
      <Button
        variant="subtle"
        color="red"
        onClick={onRemove}
        loading={isRemoving}
        disabled={isSaving}
        data-testid="remove-importance-btn"
      >
        Remove
      </Button>
    ) : (
      <div />
    )}
    <Group gap="sm">
      <Button variant="default" onClick={onClose} disabled={isPending} data-testid="cancel-btn">
        Cancel
      </Button>
      <Button type="submit" loading={isSaving} disabled={!canSubmit || isRemoving} data-testid="save-importance-btn">
        {isEditMode ? 'Update' : 'Save'}
      </Button>
    </Group>
  </Group>
);
