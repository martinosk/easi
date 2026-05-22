import { Anchor, Center, Container, Loader, Paper, Stack, Table, Text, Title } from '@mantine/core';
import { Link } from 'react-router-dom';
import { useMyEditGrants } from '../hooks/useEditGrants';
import type { EditGrant } from '../types';
import classes from './MyEditAccessPage.module.css';

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

function ArtifactCell({ grant }: { grant: EditGrant }) {
  const name = grant.artifactName || 'Deleted artifact';
  const href = grant._links?.artifact?.href;

  if (href) {
    return (
      <Anchor component={Link} to={href} fw={500}>
        {name}
      </Anchor>
    );
  }
  return <Text component="span">{name}</Text>;
}

function PageShell({ children }: { children: React.ReactNode }) {
  return (
    <div className={classes.page}>
      <Container size="xl" py="xl">
        {children}
      </Container>
    </div>
  );
}

function EmptyState() {
  return (
    <Stack align="center" gap="md" py="xl">
      <Text size="lg" c="dimmed">
        You have no active edit access grants
      </Text>
    </Stack>
  );
}

function GrantsTable({ grants }: { grants: EditGrant[] }) {
  return (
    <Paper shadow="sm" radius="lg" withBorder>
      <Table striped highlightOnHover verticalSpacing="sm">
        <Table.Thead>
          <Table.Tr>
            <Table.Th>Artifact</Table.Th>
            <Table.Th>Granted by</Table.Th>
            <Table.Th>Reason</Table.Th>
            <Table.Th>Expires</Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {grants.map((grant) => (
            <Table.Tr key={grant.id}>
              <Table.Td>
                <ArtifactCell grant={grant} />
              </Table.Td>
              <Table.Td>{grant.grantorEmail}</Table.Td>
              <Table.Td>{grant.reason || '—'}</Table.Td>
              <Table.Td>
                <Text c="dimmed" size="xs">
                  {formatDate(grant.expiresAt)}
                </Text>
              </Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
    </Paper>
  );
}

export function MyEditAccessPage() {
  const { data: grants, isLoading } = useMyEditGrants();
  const activeGrants = grants?.filter((g) => g.status === 'active') ?? [];

  if (isLoading) {
    return (
      <PageShell>
        <Center py="xl">
          <Stack align="center" gap="md">
            <Loader />
            <Text>Loading edit access...</Text>
          </Stack>
        </Center>
      </PageShell>
    );
  }

  return (
    <PageShell>
      <Stack gap="xs" mb="xl">
        <Title order={1}>My Edit Access</Title>
        <Text c="dimmed">Artifacts you have been granted write access to.</Text>
      </Stack>
      {activeGrants.length === 0 ? <EmptyState /> : <GrantsTable grants={activeGrants} />}
    </PageShell>
  );
}

export default MyEditAccessPage;
