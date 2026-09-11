import { ActionIcon, Badge, Group, Stack, Text, Title } from '@mantine/core';
import type { SubjectAttribute } from '../../../api/types';
import { hasLink } from '../../../utils/hateoas';

interface RetiredFieldsListProps {
  attributes: SubjectAttribute[];
  onReactivate: (attribute: SubjectAttribute) => void;
}

export function RetiredFieldsList({ attributes, onReactivate }: RetiredFieldsListProps) {
  const retired = attributes.filter((attribute) => !attribute.active);
  if (retired.length === 0) return null;

  return (
    <Stack gap="xs" data-testid="one-pager-retired-fields">
      <Title order={5}>Retired custom fields</Title>
      {retired.map((attribute) => (
        <Group key={attribute.id} justify="space-between">
          <Group gap="xs">
            <Text size="sm">{attribute.name}</Text>
            <Badge variant="outline" color="gray" size="sm">
              {attribute.type}
            </Badge>
          </Group>
          {hasLink(attribute, 'x-reactivate') && (
            <ActionIcon
              variant="subtle"
              color="green"
              aria-label={`Reactivate ${attribute.name}`}
              onClick={() => onReactivate(attribute)}
              data-testid={`one-pager-reactivate-${attribute.id}`}
            >
              ↺
            </ActionIcon>
          )}
        </Group>
      ))}
    </Stack>
  );
}
