import { ActionIcon, Badge, Checkbox, Group, Stack, Text } from '@mantine/core';
import { IconPencil } from '@tabler/icons-react';
import { hasLink } from '../../../utils/hateoas';
import type { BuiltInField, CustomField, SelectionOption } from '../types';
import { NumberFieldBoundsEditor } from './NumberFieldBoundsEditor';
import { SelectionOptionsEditor } from './SelectionOptionsEditor';

export function isBuiltInField(field: BuiltInField | CustomField): field is BuiltInField {
  return 'included' in field;
}

export interface FieldRowActions {
  onMoveUp: (index: number) => void;
  onMoveDown: (index: number) => void;
  onRename: (field: CustomField) => void;
  onToggleRequired: (field: CustomField, required: boolean) => void;
  onRetireCustom: (field: CustomField) => void;
  onExcludeBuiltIn: (field: BuiltInField) => void;
  onToggleBuiltInRequired: (field: BuiltInField, required: boolean) => void;
  onAddOption: (field: CustomField, label: string) => void;
  onRetireOption: (option: SelectionOption) => void;
  onSetBounds: (field: CustomField, min: number | undefined, max: number | undefined) => void;
}

interface FieldRowProps {
  field: BuiltInField | CustomField;
  index: number;
  isFirst: boolean;
  isLast: boolean;
  canReorder: boolean;
  actions: FieldRowActions;
}

function ReorderControls({ index, isFirst, isLast, canReorder, actions }: Omit<FieldRowProps, 'field'>) {
  if (!canReorder) return null;
  return (
    <Group gap={2}>
      <ActionIcon
        size="sm"
        variant="subtle"
        disabled={isFirst}
        aria-label="Move up"
        onClick={() => actions.onMoveUp(index)}
        data-testid={`one-pager-move-up-${index}`}
      >
        ↑
      </ActionIcon>
      <ActionIcon
        size="sm"
        variant="subtle"
        disabled={isLast}
        aria-label="Move down"
        onClick={() => actions.onMoveDown(index)}
        data-testid={`one-pager-move-down-${index}`}
      >
        ↓
      </ActionIcon>
    </Group>
  );
}

function BuiltInRow({ field, actions }: { field: BuiltInField; actions: FieldRowActions }) {
  return (
    <Group justify="space-between" flex={1}>
      <Group gap="xs">
        <Text fw={500}>{field.label}</Text>
        <Badge variant="outline" color="gray" size="sm">
          Built-in
        </Badge>
      </Group>
      <Group gap="xs">
        {hasLink(field, 'x-set-requirement') && (
          <Checkbox
            size="xs"
            label="Required"
            checked={field.required}
            onChange={(e) => actions.onToggleBuiltInRequired(field, e.currentTarget.checked)}
            data-testid={`one-pager-builtin-required-${field.id}`}
          />
        )}
        {hasLink(field, 'x-exclude') && (
          <ActionIcon
            variant="subtle"
            color="red"
            aria-label={`Exclude ${field.label}`}
            onClick={() => actions.onExcludeBuiltIn(field)}
            data-testid={`one-pager-exclude-${field.id}`}
          >
            −
          </ActionIcon>
        )}
      </Group>
    </Group>
  );
}

function CustomRow({ field, actions }: { field: CustomField; actions: FieldRowActions }) {
  return (
    <Stack gap="xs" flex={1}>
      <Group justify="space-between">
        <Group gap="xs">
          <Text fw={500}>{field.name}</Text>
          <Badge variant="light" color="teal" size="sm">
            Custom · {field.type}
          </Badge>
          {field.helpText && (
            <Text size="xs" c="dimmed">
              {field.helpText}
            </Text>
          )}
        </Group>
        <Group gap="xs">
          {hasLink(field, 'x-set-requirement') && (
            <Checkbox
              size="xs"
              label="Required"
              checked={field.required}
              onChange={(e) => actions.onToggleRequired(field, e.currentTarget.checked)}
              data-testid={`one-pager-required-${field.id}`}
            />
          )}
          {hasLink(field, 'x-rename') && (
            <ActionIcon
              variant="subtle"
              aria-label={`Rename ${field.name}`}
              onClick={() => actions.onRename(field)}
              data-testid={`one-pager-rename-${field.id}`}
            >
              <IconPencil size={16} stroke={1.75} />
            </ActionIcon>
          )}
          {hasLink(field, 'x-retire') && (
            <ActionIcon
              variant="subtle"
              color="red"
              aria-label={`Retire ${field.name}`}
              onClick={() => actions.onRetireCustom(field)}
              data-testid={`one-pager-retire-${field.id}`}
            >
              −
            </ActionIcon>
          )}
        </Group>
      </Group>
      {field.type === 'selection' && (
        <SelectionOptionsEditor
          field={field}
          onAddOption={actions.onAddOption}
          onRetireOption={actions.onRetireOption}
        />
      )}
      {field.type === 'number' && <NumberFieldBoundsEditor field={field} onSave={actions.onSetBounds} />}
    </Stack>
  );
}

export function FieldRow({ field, index, isFirst, isLast, canReorder, actions }: FieldRowProps) {
  return (
    <Group align="flex-start" wrap="nowrap" data-testid={`one-pager-field-row-${index}`}>
      <ReorderControls index={index} isFirst={isFirst} isLast={isLast} canReorder={canReorder} actions={actions} />
      {isBuiltInField(field) ? (
        <BuiltInRow field={field} actions={actions} />
      ) : (
        <CustomRow field={field} actions={actions} />
      )}
    </Group>
  );
}
