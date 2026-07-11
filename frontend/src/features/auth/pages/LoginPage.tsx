import { Alert, Button, Center, Group, Paper, Stack, Text, TextInput, Title } from '@mantine/core';
import { type FC, type FormEvent, useEffect, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { resetLoginRedirectFlag } from '../../../api';
import { authApi } from '../api/authApi';
import classes from './LoginPage.module.css';

function getReturnUrlFromParams(searchParams: URLSearchParams): string | undefined {
  return searchParams.get('returnUrl') ?? undefined;
}

function isExternalHttps(url: URL): boolean {
  return url.protocol === 'https:' && url.origin !== window.location.origin;
}

function isDevLocalhost(url: URL): boolean {
  if (!import.meta.env.DEV) return false;
  if (url.protocol !== 'http:') return false;
  return url.hostname === 'localhost' || url.hostname === '127.0.0.1';
}

function isAllowedAuthorizeUrl(url: URL): boolean {
  return isExternalHttps(url) || isDevLocalhost(url);
}

function sanitizeAuthorizeUrl(untrustedUrl: string): string | null {
  let parsed: URL;
  try {
    parsed = new URL(untrustedUrl);
  } catch {
    return null;
  }
  if (isAllowedAuthorizeUrl(parsed)) {
    return parsed.href;
  }
  return null;
}

export const LoginPage: FC = () => {
  const [searchParams] = useSearchParams();
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const returnUrl = useMemo(() => getReturnUrlFromParams(searchParams), [searchParams]);

  useEffect(() => {
    resetLoginRedirectFlag();
  }, []);

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (!email.trim()) {
      setError('Email is required');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await authApi.initiateLogin(email, returnUrl);
      const sanitizedUrl = sanitizeAuthorizeUrl(response._links.authorize);
      if (sanitizedUrl === null) {
        setLoading(false);
        setError('Invalid authorization URL received');
        return;
      }
      window.location.href = sanitizedUrl;
    } catch (err) {
      setLoading(false);
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('An unexpected error occurred');
      }
    }
  };

  return (
    <Center className={classes.page} p="lg">
      <Stack className={classes.column} gap="xl" align="center">
        <Stack gap="sm" align="center">
          <Group gap="sm" align="center">
            <div className={classes.tick} />
            <Title order={1} className={classes.wordmark}>
              easi
            </Title>
          </Group>
          <Text size="sm" className={classes.tagline}>
            Map capabilities, applications, and the journeys between them.
          </Text>
        </Stack>

        <Paper withBorder radius="lg" shadow="xs" p="lg" w="100%" className={classes.card}>
          <form onSubmit={handleSubmit}>
            <Stack gap="lg">
              <TextInput
                id="email"
                type="email"
                label="Email Address"
                placeholder="john@acme.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                disabled={loading}
                autoFocus
              />

              {error && <Alert color="red">{error}</Alert>}

              <Button type="submit" fullWidth size="md" loading={loading}>
                Continue with SSO
              </Button>
            </Stack>
          </form>
        </Paper>

        <Text size="xs" className={classes.footer}>
          EASI — enterprise architecture
        </Text>
      </Stack>
    </Center>
  );
};
