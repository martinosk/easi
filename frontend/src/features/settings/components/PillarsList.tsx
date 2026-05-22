import { ActionIcon, Checkbox, Group, NativeSelect, TextInput } from '@mantine/core';
import type { FitType, StrategyPillar } from '../../../api/types';
import { HelpTooltip } from '../../../components/shared/HelpTooltip';
import type { EditablePillar, ValidationErrors } from './pillarChanges';

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
      <div className="pillars-list">
        <div className="pillars-empty-state" data-testid="empty-pillars-state">
          No strategy pillars configured yet. Click Edit to add pillars.
        </div>
      </div>
    );
  }
  return (
    <div className="pillars-list">
      {describeRows(pillars, isEditing).map((row) => (
        <PillarRow
          key={row.pillar.id}
          {...row}
          validationError={validationErrors[row.index]?.name}
          activeCount={activeCount}
          handlers={handlers}
        />
      ))}
    </div>
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
    <div
      className={`pillar-row ${markedForDeletion ? 'pillar-marked-for-deletion' : ''}`}
      data-testid={`pillar-row-${index}`}
    >
      <span className="pillar-order">{orderLabel}</span>
      <div className="pillar-content">
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
      </div>
    </div>
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
        className="pillar-name-input"
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
        className="pillar-description-input"
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
    <div className="pillar-fit-config">
      <Group gap="xs" className="fit-scoring-toggle">
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
          <Group gap="xs" className="fit-type-selector">
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
            className="pillar-fit-criteria-input"
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
    </div>
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
    <div className="pillar-actions">
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
          🗑
        </ActionIcon>
      )}
    </div>
  );
}

type PillarView = Pick<StrategyPillar, 'name' | 'description' | 'fitScoringEnabled' | 'fitType' | 'fitCriteria'>;

function PillarViewRow({ pillar }: { pillar: PillarView }) {
  return (
    <>
      <span className="pillar-name">{pillar.name}</span>
      {pillar.description && <span className="pillar-description-view">{pillar.description}</span>}
      {pillar.fitScoringEnabled && (
        <div className="pillar-fit-info">
          <span className="fit-scoring-badge">Fit Scoring Enabled</span>
          {pillar.fitType && (
            <span className="fit-type-badge">{pillar.fitType === 'TECHNICAL' ? 'Technical' : 'Functional'}</span>
          )}
          {pillar.fitCriteria && <span className="fit-criteria-view">{pillar.fitCriteria}</span>}
        </div>
      )}
    </>
  );
}
