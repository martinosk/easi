import { Anchor, Badge, Group, Stack, Text } from '@mantine/core';
import { generatePath, Link } from 'react-router-dom';
import { ROUTES } from '../../../routes/routePaths';
import { formatIsoDate } from '../../../utils/date';
import type { BuiltInExpertView, BuiltInReference, BuiltInValue } from '../types';

function MaturityValueDisplay({ value, section }: { value: number; section?: string }) {
  return (
    <Group gap="xs">
      <Text size="sm">{value}</Text>
      {section && (
        <Badge variant="light" size="sm">
          {section}
        </Badge>
      )}
    </Group>
  );
}

function ExpertsValueDisplay({ experts }: { experts: BuiltInExpertView[] }) {
  return (
    <Stack gap={2}>
      {experts.map((expert) => (
        <Text size="sm" key={`${expert.name}-${expert.contact}`}>
          {expert.name} ({expert.role}), {expert.contact}
        </Text>
      ))}
    </Stack>
  );
}

function ReferencesValueDisplay({ references }: { references: BuiltInReference[] }) {
  return (
    <Stack gap={2}>
      {references.map((reference) =>
        reference.subjectType ? (
          <Anchor
            component={Link}
            size="sm"
            key={reference.id}
            to={generatePath(ROUTES.ONE_PAGER_DETAIL, {
              subjectType: reference.subjectType,
              subjectId: reference.id,
            })}
          >
            {reference.label}
          </Anchor>
        ) : (
          <Text size="sm" key={reference.id}>
            {reference.label}
          </Text>
        ),
      )}
    </Stack>
  );
}

interface BuiltInValueDisplayProps {
  value: BuiltInValue | null;
}

export function BuiltInValueDisplay({ value }: BuiltInValueDisplayProps) {
  if (!value) {
    return (
      <Text size="sm" c="dimmed">
        —
      </Text>
    );
  }
  switch (value.type) {
    case 'text':
      return <Text size="sm">{value.text}</Text>;
    case 'date':
      return <Text size="sm">{formatIsoDate(value.date)}</Text>;
    case 'maturity':
      return <MaturityValueDisplay value={value.maturity.value} section={value.maturity.section} />;
    case 'experts':
      return <ExpertsValueDisplay experts={value.experts} />;
    case 'references':
      return <ReferencesValueDisplay references={value.references} />;
  }
}
