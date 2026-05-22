import { Alert, Button, Center, Group, Stack, Text, Title } from '@mantine/core';
import { Component, type ErrorInfo, type ReactNode } from 'react';

interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode | ((error: Error, reset: () => void) => ReactNode);
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    this.props.onError?.(error, errorInfo);
  }

  reset = (): void => {
    this.setState({ hasError: false, error: null });
  };

  render(): ReactNode {
    if (this.state.hasError && this.state.error) {
      const { fallback } = this.props;

      if (typeof fallback === 'function') {
        return fallback(this.state.error, this.reset);
      }

      if (fallback) {
        return fallback;
      }

      return <DefaultErrorFallback error={this.state.error} onReset={this.reset} />;
    }

    return this.props.children;
  }
}

interface DefaultErrorFallbackProps {
  error: Error;
  onReset: () => void;
}

function DefaultErrorFallback({ error, onReset }: DefaultErrorFallbackProps) {
  return (
    <Center p="xl">
      <Alert color="red" title="Something went wrong" maw={480}>
        <Stack gap="md">
          <Text size="sm">{error.message}</Text>
          <Button onClick={onReset} variant="filled" color="red" size="sm">
            Try again
          </Button>
        </Stack>
      </Alert>
    </Center>
  );
}

interface FeatureErrorFallbackProps {
  featureName: string;
  error: Error;
  onReset: () => void;
}

export function FeatureErrorFallback({ featureName, error, onReset }: FeatureErrorFallbackProps) {
  return (
    <Center p="xl">
      <Stack align="center" gap="md" maw={520}>
        <Title order={3} c="red">
          {featureName} encountered an error
        </Title>
        <Text size="sm" c="dimmed" ta="center">
          {error.message}
        </Text>
        <Group gap="sm">
          <Button onClick={onReset} color="red">
            Try again
          </Button>
          <Button variant="default" onClick={() => window.location.reload()}>
            Reload page
          </Button>
        </Group>
      </Stack>
    </Center>
  );
}
