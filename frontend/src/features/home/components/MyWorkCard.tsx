import { Badge, Card, Group, Stack, Text } from '@mantine/core';
import { generatePath, Link } from 'react-router-dom';
import { ROUTES } from '../../../routes/routePaths';
import { hasLink } from '../../../utils/hateoas';
import { TimeGradeBadge } from '../../architecture-direction/components/TimeGradeBadge';
import { normalizeTimeGrade } from '../../architecture-direction/utils/timeGrade';
import type { MyWorkItem, MyWorkRelation, MyWorkSubjectType } from '../types';
import { formatCalendarDate } from '../utils/homeText';

const RELATION_LABEL: Record<MyWorkRelation, string> = {
  'ea-owner': 'EA owner',
  owner: 'Owner',
  nominated: 'Nominated owner',
  'edit-grant': 'Edit access',
};

const SUBJECT_LABEL: Record<MyWorkSubjectType, string> = {
  capability: 'Capability',
  application: 'Application',
};

function MyWorkCardBody({ item }: { item: MyWorkItem }) {
  const grade = normalizeTimeGrade(item.dominantGrade);
  return (
    <Stack gap="xs">
      <Group justify="space-between" wrap="nowrap" gap="sm">
        <Text fw={600} truncate>
          {item.name}
        </Text>
        {grade && <TimeGradeBadge grade={grade} labelled />}
      </Group>
      <Group gap="xs">
        <Badge variant="light" color="gray" size="sm">
          {SUBJECT_LABEL[item.subjectType]}
        </Badge>
        {item.level && (
          <Badge variant="outline" color="gray" size="sm">
            {item.level}
          </Badge>
        )}
        <Badge variant="light" size="sm">
          {RELATION_LABEL[item.relation]}
        </Badge>
      </Group>
      {item.grantExpiresOn && (
        <Text size="xs" c="dimmed">
          {`Edit access until ${formatCalendarDate(item.grantExpiresOn)}`}
        </Text>
      )}
    </Stack>
  );
}

export function MyWorkCard({ item }: { item: MyWorkItem }) {
  if (!hasLink(item, 'x-one-pager')) {
    return (
      <Card withBorder radius="md" padding="md" data-testid="my-work-card">
        <MyWorkCardBody item={item} />
      </Card>
    );
  }
  const to = generatePath(ROUTES.ONE_PAGER_DETAIL, { subjectType: item.subjectType, subjectId: item.id });
  return (
    <Card component={Link} to={to} withBorder radius="md" padding="md" data-testid="my-work-card">
      <MyWorkCardBody item={item} />
    </Card>
  );
}
