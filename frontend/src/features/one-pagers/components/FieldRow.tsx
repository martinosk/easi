import { ActionIcon, Badge, Checkbox, Group, Stack, Text } from '@mantine/core';
import { IconPencil } from '@tabler/icons-react';
import type { SubjectAttribute } from '../../../api/types';
import { hasLink } from '../../../utils/hateoas';
import type { AttributeSchemaActions } from '../hooks/useAttributeSchemaActions';
import type { PresentationActions } from '../hooks/useOnePagerFieldActions';
import type { BuiltInField, CustomField } from '../types';
import { NumberFieldBoundsEditor } from './NumberFieldBoundsEditor';
import { SelectionOptionsEditor } from './SelectionOptionsEditor';

export function isBuiltInField(field: BuiltInField | CustomField): field is BuiltInField {
  return 'included' in field && 'label' in field;
}

export interface FieldRowActions {
  presentation: PresentationActions;
  schema: AttributeSchemaActions;
}

interface FieldRowProps {
  field: BuiltInField | CustomField;
  attribute?: SubjectAttribute;
  index: number;
  isFirst: boolean;
  isLast: boolean;
  canReorder: boolean;
  actions: FieldRowActions;
}

function ReorderControls({ index, isFirst, isLast, canReorder, actions }: Omit<FieldRowProps, 'field'>) {
  if (!canReorder) return null;
  return (
    <Group gap={0}>
      <ActionIcon
        size="sm"
        variant="subtle"
        disabled={isFirst}
        aria-label="Move up"
        onClick={() => actions.presentation.onMoveUp(index)}
        data-testid={`one-pager-move-up-${index}`}
      >
        ↑
      </ActionIcon>
      <ActionIcon
        size="sm"
        variant="subtle"
        disabled={isLast}
        aria-label="Move down"
        onClick={() => actions.presentation.onMoveDown(index)}
        data-testid={`one-pager-move-down-${index}`}
      >
        ↓
      </ActionIcon>
    </Group>
  );
}

function BuiltInRow({ field, actions }: { field: BuiltInField; actions: PresentationActions }) {
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

interface CustomRowProps {
  field: CustomField;
  attribute?: SubjectAttribute;
  actions: FieldRowActions;
}

function SchemaControls({ attribute, actions }: { attribute?: SubjectAttribute; actions: AttributeSchemaActions }) {
  if (!attribute) return null;
  return (
    <>
      {hasLink(attribute, 'x-rename') && (
        <ActionIcon
          variant="subtle"
          aria-label={`Rename ${attribute.name}`}
          onClick={() => actions.onRename(attribute)}
          data-testid={`one-pager-rename-${attribute.id}`}
        >
          <IconPencil size={16} stroke={1.75} />
        </ActionIcon>
      )}
      {hasLink(attribute, 'x-retire') && (
        <ActionIcon
          variant="subtle"
          color="red"
          aria-label={`Retire ${attribute.name}`}
          onClick={() => actions.onRetire(attribute)}
          data-testid={`one-pager-retire-${attribute.id}`}
        >
          −
        </ActionIcon>
      )}
    </>
  );
}

function CustomRow({ field, attribute, actions }: CustomRowProps) {
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
              onChange={(e) => actions.presentation.onToggleRequired(field, e.currentTarget.checked)}
              data-testid={`one-pager-required-${field.id}`}
            />
          )}
          <SchemaControls attribute={attribute} actions={actions.schema} />
        </Group>
      </Group>
      {attribute?.type === 'selection' && (
        <SelectionOptionsEditor
          attribute={attribute}
          onAddOption={actions.schema.onAddOption}
          onRetireOption={actions.schema.onRetireOption}
        />
      )}
      {attribute?.type === 'number' && <NumberFieldBoundsEditor attribute={attribute} onSave={actions.schema.onSetBounds} />}
    </Stack>
  );
}

export function FieldRow({ field, attribute, index, isFirst, isLast, canReorder, actions }: FieldRowProps) {
  return (
    <Group align="flex-start" wrap="nowrap" data-testid={`one-pager-field-row-${index}`}>
      <ReorderControls index={index} isFirst={isFirst} isLast={isLast} canReorder={canReorder} actions={actions} />
      {isBuiltInField(field) ? (
        <BuiltInRow field={field} actions={actions.presentation} />
      ) : (
        <CustomRow field={field} attribute={attribute} actions={actions} />
      )}
    </Group>
  );
}
