import { Alert, Button, Center, Container, Group, Loader, Stack, Text } from '@mantine/core';
import type { ReactNode } from 'react';
import { MyWorkSection } from '../components/MyWorkSection';
import { PortfolioTiles } from '../components/PortfolioTiles';
import { ScopeHeader } from '../components/ScopeHeader';
import { useHome } from '../hooks/useHome';
import type { HomeResponse } from '../types';
import classes from './HomePage.module.css';

function PageShell({ children }: { children: ReactNode }) {
  return (
    <div className={classes.page} data-testid="home-page">
      <Container size="xl" py="xl">
        {children}
      </Container>
    </div>
  );
}

function HomeLoading() {
  return (
    <Center py="xl">
      <Stack align="center" gap="sm">
        <Loader />
        <Text c="dimmed">Loading your home...</Text>
      </Stack>
    </Center>
  );
}

function HomeError({ onRetry }: { onRetry: () => void }) {
  return (
    <Alert color="red" title="Could not load your home" role="alert">
      <Stack gap="sm">
        <Text size="sm">Something went wrong while loading your home.</Text>
        <Group justify="flex-start">
          <Button variant="default" size="xs" onClick={onRetry}>
            Retry
          </Button>
        </Group>
      </Stack>
    </Alert>
  );
}

function HomeContent({ home }: { home: HomeResponse }) {
  return (
    <Stack gap="xl">
      <ScopeHeader scope={home.scope} />
      {home.portfolio && <PortfolioTiles portfolio={home.portfolio} />}
      {home.myWork && <MyWorkSection myWork={home.myWork} />}
    </Stack>
  );
}

function HomeBody() {
  const { data, isLoading, refetch } = useHome();
  if (data) return <HomeContent home={data} />;
  if (isLoading) return <HomeLoading />;
  return <HomeError onRetry={() => refetch()} />;
}

export function HomePage() {
  return (
    <PageShell>
      <HomeBody />
    </PageShell>
  );
}
