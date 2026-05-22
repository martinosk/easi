import { Anchor, Card, Loader, Stack, Text, Title } from '@mantine/core';
import { useMyEditGrants } from '../hooks/useEditGrants';
import type { EditGrant } from '../types';

function GrantCard({ grant }: { grant: EditGrant }) {
  const name = grant.artifactName || 'Deleted artifact';
  const href = grant._links?.artifact?.href;

  return (
    <Card withBorder padding="md" radius="md" data-testid={`my-grant-${grant.id}`}>
      <Stack gap="xs">
        {href ? (
          <Anchor href={href} fw={600}>
            {name}
          </Anchor>
        ) : (
          <Text fw={600}>{name}</Text>
        )}
        <Text size="sm" c="dimmed">
          Granted by {grant.grantorEmail}
        </Text>
        <Text size="xs" c="dimmed">
          Expires {new Date(grant.expiresAt).toLocaleDateString()}
        </Text>
        {grant.reason && <Text size="sm">{grant.reason}</Text>}
      </Stack>
    </Card>
  );
}

export function MyEditGrants() {
  const { data: grants, isLoading } = useMyEditGrants();

  if (isLoading) {
    return <Loader data-testid="my-edit-grants-loading" />;
  }

  const activeGrants = grants?.filter((g) => g.status === 'active') ?? [];

  if (activeGrants.length === 0) {
    return null;
  }

  return (
    <Stack gap="md" data-testid="my-edit-grants">
      <Title order={3}>Your Edit Access</Title>
      {activeGrants.map((grant) => (
        <GrantCard key={grant.id} grant={grant} />
      ))}
    </Stack>
  );
}
