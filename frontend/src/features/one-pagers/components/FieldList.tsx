import { Stack, Text } from '@mantine/core';
import type { SubjectAttribute } from '../../../api/types';
import { hasLink } from '../../../utils/hateoas';
import type { BuiltInField, CustomField, FieldRef, OnePagerConfiguration } from '../types';
import { FieldRow, type FieldRowActions } from './FieldRow';

interface FieldListProps {
  configuration: OnePagerConfiguration;
  attributesById: Map<string, SubjectAttribute>;
  actions: FieldRowActions;
}

function resolveField(configuration: OnePagerConfiguration, ref: FieldRef): BuiltInField | CustomField | undefined {
  if (ref.kind === 'builtIn') return configuration.builtInFields.find((f) => f.id === ref.id);
  return configuration.customFields.find((f) => f.id === ref.id);
}

export function FieldList({ configuration, attributesById, actions }: FieldListProps) {
  const canReorder = hasLink(configuration, 'x-reorder');
  const rows = configuration.displayOrder
    .map((ref, index) => ({ ref, index, field: resolveField(configuration, ref) }))
    .filter(
      (row): row is { ref: FieldRef; index: number; field: BuiltInField | CustomField } => row.field !== undefined,
    );

  if (rows.length === 0) {
    return (
      <Text c="dimmed" size="sm" data-testid="one-pager-field-list-empty">
        No fields are shown on this one-pager yet.
      </Text>
    );
  }

  return (
    <Stack gap="sm" data-testid="one-pager-field-list">
      {rows.map(({ ref, field, index }) => (
        <FieldRow
          key={field.id}
          field={field}
          attribute={ref.kind === 'custom' ? attributesById.get(ref.id) : undefined}
          index={index}
          isFirst={index === 0}
          isLast={index === rows.length - 1}
          canReorder={canReorder}
          actions={actions}
        />
      ))}
    </Stack>
  );
}
