import { ActionIcon, Box, Checkbox, Group, NativeSelect, Stack, Text, TextInput } from '@mantine/core';
import { IconTrash } from '@tabler/icons-react';
import type { FitType, StrategyPillar } from '../../../api/types';
import { HelpTooltip } from '../../../components/shared/HelpTooltip';
import type { EditablePillar, ValidationErrors } from './pillarChanges';
import classes from './PillarsList.module.css';

export interface PillarHandlers {
  onNameChange: (index: number, name: string) => void;
  onDescriptionChange: (index: number, description: string) => void;
  onFitScoringEnabledChange: (index: number, enabled: boolean) => void;
  onFitCriteriaChange: (index: number, criteria: string) => void;
  onFitTypeChange: (index: number, fitType: FitType) => void;
  onDelete: (index: number) => void;
  onRestore: (index: number) => void;
}

interface PillarsListProps {
  pillars: EditablePillar[] | StrategyPillar[];
  isEditing: boolean;
  validationErrors: ValidationErrors;
  activeCount: number;
  handlers: PillarHandlers;
}

interface RowDescriptor {
  pillar: EditablePillar | StrategyPillar;
  editable: EditablePillar | null;
  index: number;
  orderLabel: string;
  markedForDeletion: boolean;
}

function describeRows(pillars: EditablePillar[] | StrategyPillar[], isEditing: boolean): RowDescriptor[] {
  let order = 0;
  return pillars.map((pillar, index) => {
    const editable = isEditing ? (pillar as EditablePillar) : null;
    const markedForDeletion = editable?.markedForDeletion === true;
    const showOrder = !markedForDeletion && (pillar.active || editable?.isNew === true);
    return { pillar, editable, index, orderLabel: showOrder ? `${++order}.` : '', markedForDeletion };
  });
}

export function PillarsList({ pillars, isEditing, validationErrors, activeCount, handlers }: PillarsListProps) {
  if (pillars.length === 0 && !isEditing) {
    return (
      <Text p="xl" ta="center" c="gray.5" className={classes.emptyState} data-testid="empty-pillars-state">
        No strategy pillars configured yet. Click Edit to add pillars.
      </Text>
    );
  }
  return (
    <Stack gap="sm">
      {describeRows(pillars, isEditing).map((row) => (
        <PillarRow
          key={row.pillar.id}
          {...row}
          validationError={validationErrors[row.index]?.name}
          activeCount={activeCount}
          handlers={handlers}
        />
      ))}
    </Stack>
  );
}

interface PillarRowProps {
  pillar: EditablePillar | StrategyPillar;
  editable: EditablePillar | null;
  index: number;
  orderLabel: string;
  markedForDeletion: boolean;
  validationError: string | undefined;
  activeCount: number;
  handlers: PillarHandlers;
}

function PillarRow({
  pillar,
  editable,
  index,
  orderLabel,
  markedForDeletion,
  validationError,
  activeCount,
  handlers,
}: PillarRowProps) {
  return (
    <Group
      align="flex-start"
      gap="md"
      p="md"
      wrap="nowrap"
      className={classes.row}
      mod={{ 'marked-for-deletion': markedForDeletion }}
      data-testid={`pillar-row-${index}`}
    >
      <Text size="lg" fw={600} c="gray.5" miw="xl" pt="xs">
        {orderLabel}
      </Text>
      <Stack flex={1} gap="xs">
        {editable ? (
          <PillarEditRow
            pillar={editable}
            index={index}
            validationError={validationError}
            canDelete={!markedForDeletion && activeCount > 1}
            handlers={handlers}
          />
        ) : (
          <PillarViewRow pillar={pillar} />
        )}
      </Stack>
    </Group>
  );
}

interface PillarEditRowProps {
  pillar: EditablePillar;
  index: number;
  validationError: string | undefined;
  canDelete: boolean;
  handlers: PillarHandlers;
}

function PillarEditRow({ pillar, index, validationError, canDelete, handlers }: PillarEditRowProps) {
  const disabled = pillar.markedForDeletion;
  return (
    <>
      <TextInput
        value={pillar.name}
        onChange={(e) => handlers.onNameChange(index, e.currentTarget.value)}
        placeholder="Pillar name"
        data-testid={`pillar-name-input-${index}`}
        maxLength={100}
        disabled={disabled}
        error={validationError}
        fw={600}
      />
      <TextInput
        value={pillar.description}
        onChange={(e) => handlers.onDescriptionChange(index, e.currentTarget.value)}
        placeholder="Description (optional)"
        data-testid={`pillar-description-input-${index}`}
        maxLength={500}
        disabled={disabled}
      />
      <FitConfigEditor pillar={pillar} index={index} disabled={disabled} handlers={handlers} />
      <PillarRowActions
        pillar={pillar}
        index={index}
        canDelete={canDelete}
        onDelete={handlers.onDelete}
        onRestore={handlers.onRestore}
      />
    </>
  );
}

interface FitConfigEditorProps {
  pillar: EditablePillar;
  index: number;
  disabled: boolean;
  handlers: Pick<PillarHandlers, 'onFitScoringEnabledChange' | 'onFitCriteriaChange' | 'onFitTypeChange'>;
}

const FIT_TYPE_OPTIONS = [
  { value: '', label: 'Select fit type' },
  { value: 'TECHNICAL', label: 'Technical' },
  { value: 'FUNCTIONAL', label: 'Functional' },
];

function FitConfigEditor({ pillar, index, disabled, handlers }: FitConfigEditorProps) {
  return (
    <Box mt="sm" pt="sm" className={classes.divider}>
      <Group gap="xs">
        <Checkbox
          checked={pillar.fitScoringEnabled}
          onChange={(e) => handlers.onFitScoringEnabledChange(index, e.currentTarget.checked)}
          disabled={disabled}
          label="Enable fit scoring for realizations"
          data-testid={`pillar-fit-scoring-checkbox-${index}`}
        />
        <HelpTooltip
          content="When enabled, realizations can be scored on how well they support this strategic pillar"
          iconOnly
        />
      </Group>
      {pillar.fitScoringEnabled && (
        <>
          <Group gap="xs" mt="sm">
            <NativeSelect
              label="Fit Type"
              data={FIT_TYPE_OPTIONS}
              value={pillar.fitType}
              onChange={(e) => handlers.onFitTypeChange(index, e.currentTarget.value as FitType)}
              disabled={disabled}
              data-testid={`pillar-fit-type-select-${index}`}
              size="xs"
            />
            <HelpTooltip
              content="Technical fit measures how well the application supports technical aspects of this pillar. Functional fit measures business functionality alignment."
              iconOnly
            />
          </Group>
          <TextInput
            value={pillar.fitCriteria}
            onChange={(e) => handlers.onFitCriteriaChange(index, e.currentTarget.value)}
            placeholder="Fit criteria (e.g., Reliability, uptime SLA, disaster recovery)"
            data-testid={`pillar-fit-criteria-input-${index}`}
            maxLength={500}
            disabled={disabled}
            mt="xs"
          />
        </>
      )}
    </Box>
  );
}

interface PillarRowActionsProps {
  pillar: EditablePillar;
  index: number;
  canDelete: boolean;
  onDelete: (index: number) => void;
  onRestore: (index: number) => void;
}

function PillarRowActions({ pillar, index, canDelete, onDelete, onRestore }: PillarRowActionsProps) {
  return (
    <Group align="flex-start" pt="xs">
      {pillar.markedForDeletion ? (
        <ActionIcon
          variant="subtle"
          color="green"
          onClick={() => onRestore(index)}
          aria-label={`Restore ${pillar.name}`}
          data-testid={`restore-pillar-btn-${index}`}
        >
          ↺
        </ActionIcon>
      ) : (
        <ActionIcon
          variant="subtle"
          color="red"
          onClick={() => onDelete(index)}
          disabled={!canDelete}
          aria-label={`Delete ${pillar.name}`}
          data-testid={`delete-pillar-btn-${index}`}
        >
          <IconTrash size={16} stroke={1.75} />
        </ActionIcon>
      )}
    </Group>
  );
}

type PillarView = Pick<StrategyPillar, 'name' | 'description' | 'fitScoringEnabled' | 'fitType' | 'fitCriteria'>;

function PillarViewRow({ pillar }: { pillar: PillarView }) {
  return (
    <>
      <Text size="lg" fw={600} c="gray.9">
        {pillar.name}
      </Text>
      {pillar.description && <Text c="dimmed">{pillar.description}</Text>}
      {pillar.fitScoringEnabled && (
        <Stack gap="xs" mt="sm" pt="sm" className={classes.divider}>
          <span className={`${classes.fitBadge} ${classes.fitBadgeScoring}`}>Fit Scoring Enabled</span>
          {pillar.fitType && (
            <span className={`${classes.fitBadge} ${classes.fitBadgeType}`}>
              {pillar.fitType === 'TECHNICAL' ? 'Technical' : 'Functional'}
            </span>
          )}
          {pillar.fitCriteria && (
            <Text size="sm" c="gray.5" fs="italic">
              {pillar.fitCriteria}
            </Text>
          )}
        </Stack>
      )}
    </>
  );
}
