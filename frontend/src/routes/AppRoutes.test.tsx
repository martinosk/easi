import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MemoryRouter, useHref, useLocation, useNavigate } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useUserStore } from '../store/userStore';
import { AppRoutes } from './AppRoutes';

const shell = vi.hoisted(() => ({ mounts: 0 }));

vi.mock('../App', async () => {
  const { useEffect } = await import('react');
  return {
    default: ({ view }: { view: string }) => {
      useEffect(() => {
        shell.mounts += 1;
      }, []);
      return <div data-testid="app-view">{view}</div>;
    },
  };
});

vi.mock('../features/auth/pages/LoginPage', () => ({
  LoginPage: () => <div data-testid="login-page" />,
}));

function LocationProbe() {
  const location = useLocation();
  const navigate = useNavigate();
  const href = useHref(location);
  return (
    <>
      <div data-testid="location">{`${location.pathname}${location.search}${location.hash}`}</div>
      <div data-testid="href">{href}</div>
      <button type="button" onClick={() => navigate('/')}>
        go home
      </button>
      <button type="button" onClick={() => navigate('/canvas')}>
        go canvas
      </button>
    </>
  );
}

function renderAt(path: string, basename?: string) {
  return render(
    <MemoryRouter initialEntries={[path]} basename={basename}>
      <AppRoutes />
      <LocationProbe />
    </MemoryRouter>,
  );
}

function currentView(): string | null {
  return screen.getByTestId('app-view').textContent;
}

function currentLocation(): string | null {
  return screen.getByTestId('location').textContent;
}

describe('AppRoutes', () => {
  const initialState = useUserStore.getState();

  beforeEach(() => {
    shell.mounts = 0;
    useUserStore.setState({ ...initialState, isAuthenticated: true, isLoading: false }, true);
  });

  it('renders Home at /', () => {
    renderAt('/');

    expect(currentView()).toBe('home');
    expect(currentLocation()).toBe('/');
  });

  it('renders the canvas at /canvas', () => {
    renderAt('/canvas');

    expect(currentView()).toBe('canvas');
    expect(currentLocation()).toBe('/canvas');
  });

  it('redirects a legacy view link to the canvas with the query preserved', () => {
    renderAt('/?view=v1&x=1');

    expect(currentView()).toBe('canvas');
    expect(currentLocation()).toBe('/canvas?view=v1&x=1');
  });

  it('keeps the hash of a legacy view link', () => {
    renderAt('/?view=v1#notes');

    expect(currentView()).toBe('canvas');
    expect(currentLocation()).toBe('/canvas?view=v1#notes');
  });

  it('keeps a hostile view value as query on the same-origin canvas path', () => {
    renderAt('/?view=//evil.example');

    expect(currentView()).toBe('canvas');
    expect(currentLocation()).toBe('/canvas?view=//evil.example');
  });

  it('redirects a legacy view link under the router basename', () => {
    renderAt('/easi/?view=//evil.example&x=1', '/easi');

    expect(currentView()).toBe('canvas');
    expect(screen.getByTestId('href').textContent).toBe('/easi/canvas?view=//evil.example&x=1');
  });

  it('leaves a view parameter alone on other pages', () => {
    renderAt('/business-domains?view=v1');

    expect(currentView()).toBe('business-domains');
    expect(currentLocation()).toBe('/business-domains?view=v1');
  });

  it('leads an unknown path to Home', () => {
    renderAt('/no/such/page');

    expect(currentView()).toBe('home');
    expect(currentLocation()).toBe('/');
  });

  it('keeps the other views on their routes', () => {
    renderAt('/business-domains');

    expect(currentView()).toBe('business-domains');
  });

  it('keeps the same App shell mounted from Home to the canvas and back', async () => {
    const user = userEvent.setup();
    renderAt('/');
    expect(shell.mounts).toBe(1);

    await user.click(screen.getByRole('button', { name: 'go canvas' }));
    expect(currentView()).toBe('canvas');

    await user.click(screen.getByRole('button', { name: 'go home' }));
    expect(currentView()).toBe('home');

    expect(shell.mounts).toBe(1);
  });
});
