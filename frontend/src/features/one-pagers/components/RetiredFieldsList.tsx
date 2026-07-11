import { ActionIcon, Badge, Group, Stack, Text, Title } from '@mantine/core';
import { hasLink } from '../../../utils/hateoas';
import type { CustomField } from '../types';

interface RetiredFieldsListProps {
  fields: CustomField[];
  onReactivate: (field: CustomField) => void;
}

export function RetiredFieldsList({ fields, onReactivate }: RetiredFieldsListProps) {
  const retired = fields.filter((field) => !field.active);
  if (retired.length === 0) return null;

  return (
    <Stack gap="xs" data-testid="one-pager-retired-fields">
      <Title order={5}>Retired custom fields</Title>
      {retired.map((field) => (
        <Group key={field.id} justify="space-between">
          <Group gap="xs">
            <Text size="sm">{field.name}</Text>
            <Badge variant="outline" color="gray" size="sm">
              {field.type}
            </Badge>
          </Group>
          {hasLink(field, 'x-reactivate') && (
            <ActionIcon
              variant="subtle"
              color="green"
              aria-label={`Reactivate ${field.name}`}
              onClick={() => onReactivate(field)}
              data-testid={`one-pager-reactivate-${field.id}`}
            >
              ↺
            </ActionIcon>
          )}
        </Group>
      ))}
    </Stack>
  );
}
