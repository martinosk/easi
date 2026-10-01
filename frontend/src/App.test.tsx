import { screen } from '@testing-library/react';
import toast from 'react-hot-toast';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { toViewId } from './api/types';
import App from './App';
import { DialogProvider } from './contexts/dialogs';
import { useAppStore } from './store/appStore';
import { useUserStore } from './store/userStore';
import { renderWithProviders } from './test/helpers';
import { buildView } from './test/helpers/entityBuilders';
import { seedDb } from './test/mocks/db';
import { server } from './test/mocks/server';

vi.mock('./contexts/releases/store/useReleaseNotes', () => ({
  useReleaseNotes: () => ({ showOverlay: false, release: null, dismiss: () => {} }),
}));

vi.mock('./features/canvas/CanvasContainer', () => ({
  default: () => <div data-testid="canvas-container" />,
}));

function renderApp(view: 'home' | 'canvas') {
  return renderWithProviders(
    <DialogProvider>
      <App view={view} />
    </DialogProvider>,
  );
}

function recordViewRequests(): string[] {
  const requests: string[] = [];
  server.events.on('request:start', ({ request }) => {
    if (new URL(request.url).pathname.startsWith('/api/v1/views')) requests.push(request.method);
  });
  return requests;
}

describe('App', () => {
  const initialUserState = useUserStore.getState();
  let viewRequests: string[];

  beforeEach(() => {
    useUserStore.setState({ ...initialUserState, isAuthenticated: true, isLoading: false }, true);
    useAppStore.setState({ currentViewId: null, isInitialized: false, openViewIds: [] });
    viewRequests = recordViewRequests();
    toast.remove();
  });

  afterEach(() => {
    server.events.removeAllListeners();
  });

  it('renders Home without initialising the canvas', async () => {
    renderApp('home');

    expect(await screen.findByRole('heading', { name: 'Home' })).toBeInTheDocument();
    expect(await screen.findByTestId('home-tiles')).toBeInTheDocument();

    expect(viewRequests).toEqual([]);
    expect(useAppStore.getState().isInitialized).toBe(false);
    expect(useAppStore.getState().currentViewId).toBeNull();
    expect(screen.queryByText('Data loaded successfully')).not.toBeInTheDocument();
    expect(screen.queryByText('Created default view')).not.toBeInTheDocument();
  });

  it('initialises the canvas when the canvas view is shown', async () => {
    seedDb({ views: [buildView({ id: toViewId('view-1'), name: 'Default View', isDefault: true })] });
    renderApp('canvas');

    expect(await screen.findByTestId('canvas-container')).toBeInTheDocument();

    expect(viewRequests).toContain('GET');
    expect(useAppStore.getState().isInitialized).toBe(true);
    expect(useAppStore.getState().currentViewId).toBe('view-1');
    expect(await screen.findByText('Data loaded successfully')).toBeInTheDocument();
  });
});
