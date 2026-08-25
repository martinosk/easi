import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../test/helpers';
import { AppNavigation } from './AppNavigation';

const mockState = {
  hasPermission: (_permission: string) => true,
  sessionLinks: null as Record<string, string> | null,
  tenant: null,
  user: null,
};

vi.mock('../../store/userStore', () => ({
  useUserStore: <T,>(selector: (state: typeof mockState) => T): T => selector(mockState),
}));

const MEASURED = { full: 100, compact: 40, more: 40 };
let navWidth = 1000;

function installMeasurement() {
  Object.defineProperty(HTMLElement.prototype, 'offsetWidth', {
    configurable: true,
    get(this: HTMLElement) {
      const measure = this.dataset.measure as keyof typeof MEASURED | undefined;
      if (measure) return MEASURED[measure];
      if (this.dataset.testid === 'app-primary-nav') return navWidth;
      return 0;
    },
  });
  class ImmediateResizeObserver {
    private readonly callback: ResizeObserverCallback;
    constructor(callback: ResizeObserverCallback) {
      this.callback = callback;
    }
    observe() {
      this.callback([], this as unknown as ResizeObserver);
    }
    unobserve() {}
    disconnect() {}
  }
  vi.stubGlobal('ResizeObserver', ImmediateResizeObserver);
}

describe('AppNavigation', () => {
  const originalOffsetWidth = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'offsetWidth');

  beforeEach(() => {
    mockState.sessionLinks = { 'x-one-pager-quality': '/api/v1/one-pager-quality' };
    navWidth = 1000;
    installMeasurement();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    if (originalOffsetWidth) Object.defineProperty(HTMLElement.prototype, 'offsetWidth', originalOffsetWidth);
    else Reflect.deleteProperty(HTMLElement.prototype, 'offsetWidth');
  });

  it('shows the One-Pager Quality nav entry when the session link is present', () => {
    renderWithProviders(<AppNavigation currentView="canvas" />);

    expect(screen.getByTestId('nav-one-pager-quality')).toBeInTheDocument();
  });

  it('hides the One-Pager Quality nav entry when the session link is absent', () => {
    mockState.sessionLinks = {};

    renderWithProviders(<AppNavigation currentView="canvas" />);

    expect(screen.queryByTestId('nav-one-pager-quality')).not.toBeInTheDocument();
  });

  it('hides the One-Pager Quality nav entry when sessionLinks is null', () => {
    mockState.sessionLinks = null;

    renderWithProviders(<AppNavigation currentView="canvas" />);

    expect(screen.queryByTestId('nav-one-pager-quality')).not.toBeInTheDocument();
  });

  it('does not show Invitations or Settings in the primary navigation', () => {
    renderWithProviders(<AppNavigation currentView="canvas" />);

    expect(screen.queryByTestId('nav-invitations')).not.toBeInTheDocument();
    expect(screen.queryByTestId('nav-settings')).not.toBeInTheDocument();
    expect(screen.getByTestId('nav-users')).toBeInTheDocument();
  });

  it('highlights the Users entry while viewing invitations', () => {
    renderWithProviders(<AppNavigation currentView="invitations" />);

    expect(screen.getByTestId('nav-users')).toHaveAttribute('aria-current', 'page');
  });

  it('shows icon and label for every entry when all labels fit', () => {
    navWidth = 1000;

    renderWithProviders(<AppNavigation currentView="canvas" />);

    expect(screen.getByTestId('nav-users')).toHaveTextContent('Users');
    expect(screen.queryByTestId('nav-more')).not.toBeInTheDocument();
  });

  it('shows a tooltip with the entry name on hover', async () => {
    navWidth = 1000;
    const user = userEvent.setup();
    renderWithProviders(<AppNavigation currentView="canvas" />);

    await user.hover(screen.getByTestId('nav-business-domains'));

    expect(await screen.findByRole('tooltip')).toHaveTextContent('Business Domains');
  });

  it('shows icons only when labels do not fit but icons do', async () => {
    navWidth = 300;
    const user = userEvent.setup();
    renderWithProviders(<AppNavigation currentView="canvas" />);

    expect(screen.getByTestId('nav-users')).not.toHaveTextContent('Users');
    expect(screen.getByTestId('nav-users')).toHaveAttribute('aria-label', 'Users');
    expect(screen.queryByTestId('nav-more')).not.toBeInTheDocument();

    await user.hover(screen.getByTestId('nav-users'));
    expect(await screen.findByRole('tooltip')).toHaveTextContent('Users');
  });

  it('overflows trailing entries into a More menu when icons do not fit', async () => {
    navWidth = 130;
    const user = userEvent.setup();
    renderWithProviders(<AppNavigation currentView="canvas" />);

    expect(screen.getByTestId('nav-canvas')).toBeInTheDocument();
    expect(screen.getByTestId('nav-business-domains')).toBeInTheDocument();
    expect(screen.queryByTestId('nav-users')).not.toBeInTheDocument();

    await user.click(screen.getByTestId('nav-more'));
    const menu = await screen.findByTestId('nav-more-menu');
    expect(within(menu).getByTestId('nav-users-overflow')).toHaveTextContent('Users');
    expect(within(menu).getByTestId('nav-value-streams-overflow')).toBeInTheDocument();
    expect(within(menu).queryByTestId('nav-canvas-overflow')).not.toBeInTheDocument();
  });

  it('marks the More button active when the current view is overflowed', () => {
    navWidth = 130;

    renderWithProviders(<AppNavigation currentView="users" />);

    expect(screen.getByTestId('nav-more')).toHaveAttribute('aria-current', 'page');
  });

  it("gives the What's New action a tooltip", async () => {
    const user = userEvent.setup();
    renderWithProviders(<AppNavigation currentView="canvas" onOpenReleaseNotes={vi.fn()} />);

    await user.hover(screen.getByTestId('nav-whats-new'));

    expect(await screen.findByRole('tooltip')).toHaveTextContent("What's New");
  });
});
