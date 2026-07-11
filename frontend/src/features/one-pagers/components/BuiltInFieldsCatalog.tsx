import { ActionIcon, Group, Stack, Text, Title } from '@mantine/core';
import { hasLink } from '../../../utils/hateoas';
import type { BuiltInField } from '../types';

interface BuiltInFieldsCatalogProps {
  fields: BuiltInField[];
  onInclude: (field: BuiltInField) => void;
}

export function BuiltInFieldsCatalog({ fields, onInclude }: BuiltInFieldsCatalogProps) {
  const excluded = fields.filter((field) => !field.included);
  if (excluded.length === 0) return null;

  return (
    <Stack gap="xs" data-testid="one-pager-excluded-catalog">
      <Title order={5}>Excluded built-in fields</Title>
      {excluded.map((field) => (
        <Group key={field.id} justify="space-between">
          <Text size="sm">{field.label}</Text>
          {hasLink(field, 'x-include') && (
            <ActionIcon
              variant="subtle"
              color="green"
              aria-label={`Include ${field.label}`}
              onClick={() => onInclude(field)}
              data-testid={`one-pager-include-${field.id}`}
            >
              +
            </ActionIcon>
          )}
        </Group>
      ))}
    </Stack>
  );
}
