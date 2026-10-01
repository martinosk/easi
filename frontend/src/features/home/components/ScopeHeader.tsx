import { List, Stack, Text, Title } from '@mantine/core';
import type { HomeScope } from '../types';
import { scopeStatements } from '../utils/homeText';

const TENANT_TEXT = 'No personal scope is set, so the whole portfolio is shown.';
const EMPTY_TEXT =
  'Home fills when you are made a steward, own an application, or are granted edit access to a capability or application.';

function PersonalScope({ scope }: { scope: HomeScope }) {
  return (
    <List size="sm" spacing="xs">
      {scopeStatements(scope).map(({ key, text }) => (
        <List.Item key={key}>{text}</List.Item>
      ))}
    </List>
  );
}

function EmptyScope() {
  return (
    <Stack gap="xs">
      <Title order={3}>Nothing here yet</Title>
      <Text c="dimmed">{EMPTY_TEXT}</Text>
    </Stack>
  );
}

function ScopeBody({ scope }: { scope: HomeScope }) {
  if (scope.kind === 'empty') return <EmptyScope />;
  if (scope.kind === 'tenant') return <Text c="dimmed">{TENANT_TEXT}</Text>;
  return <PersonalScope scope={scope} />;
}

export function ScopeHeader({ scope }: { scope: HomeScope }) {
  return (
    <Stack gap="sm" data-testid="home-scope">
      <Title order={2}>Home</Title>
      <ScopeBody scope={scope} />
    </Stack>
  );
}
