import { MantineProvider } from '@mantine/core';
import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useUserStore } from '../../store/userStore';
import { CanvasViewControls } from './CanvasViewControls';

vi.mock('../../features/business-domains', () => ({
  useBusinessDomains: () => ({ domains: [] }),
}));

vi.mock('../../features/views', () => ({
  EdgeTypeSelector: () => null,
  ColorSchemeSelector: () => null,
}));

function signInWith(links: Record<string, string>) {
  useUserStore.setState({
    sessionLinks: { self: '/api/v1/auth/sessions/current', logout: '', user: '', tenant: '', ...links },
  });
}

function renderControls() {
  return render(
    <MantineProvider>
      <CanvasViewControls />
    </MantineProvider>,
  );
}

describe('CanvasViewControls', () => {
  beforeEach(() => {
    useUserStore.setState({ sessionLinks: null });
  });

  it('offers Import when the session carries the imports link', () => {
    signInWith({ 'x-imports': '/api/v1/imports' });

    renderControls();

    expect(screen.getByTestId('import-button')).toBeInTheDocument();
  });

  it('does not offer Import when the session lacks the imports link', () => {
    signInWith({});

    renderControls();

    expect(screen.queryByTestId('import-button')).not.toBeInTheDocument();
  });
});
