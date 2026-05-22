import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers/renderWithProviders';
import { LoginPage } from './LoginPage';

const mockInitiateLogin = vi.fn();

vi.mock('../api/authApi', () => ({
  authApi: {
    initiateLogin: (...args: unknown[]) => mockInitiateLogin(...args),
  },
}));

vi.mock('../../../api', () => ({
  resetLoginRedirectFlag: vi.fn(),
}));

function mockAuthorize(url: string) {
  mockInitiateLogin.mockResolvedValue({ _links: { authorize: url } });
}

function submitLogin({ email = 'user@example.com', route = '/login' }: { email?: string; route?: string } = {}) {
  renderWithProviders(<LoginPage />, { routerProps: { initialEntries: [route] } });
  if (email) {
    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: email } });
  }
  fireEvent.click(screen.getByRole('button', { name: /continue/i }));
}

describe('LoginPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    delete (window as { location?: unknown }).location;
    (window as { location: { href: string } }).location = { href: '' };
  });

  describe('returnUrl parameter handling', () => {
    it.each([
      {
        description: 'passes returnUrl to initiateLogin when present',
        route: '/login?returnUrl=https%3A%2F%2Fapp.example.com%2Fdashboard',
        expectedReturnUrl: 'https://app.example.com/dashboard',
      },
      {
        description: 'does NOT double-decode returnUrl (security)',
        route: '/login?returnUrl=https%253A%252F%252Fevil.com',
        expectedReturnUrl: 'https%3A%2F%2Fevil.com',
      },
      {
        description: 'calls initiateLogin without returnUrl when not present',
        route: '/login',
        expectedReturnUrl: undefined,
      },
    ])('$description', async ({ route, expectedReturnUrl }) => {
      mockAuthorize('https://idp.example.com/authorize');

      submitLogin({ route });

      await waitFor(() => {
        expect(mockInitiateLogin).toHaveBeenCalledWith('user@example.com', expectedReturnUrl);
      });
    });
  });

  describe('form validation', () => {
    it('shows error when email is empty', async () => {
      submitLogin({ email: '' });

      expect(await screen.findByText('Email is required')).toBeInTheDocument();
      expect(mockInitiateLogin).not.toHaveBeenCalled();
    });
  });

  describe('authorize URL validation', () => {
    async function expectAuthorizeUrlRejected(url: string) {
      mockAuthorize(url);
      submitLogin();
      expect(await screen.findByText('Invalid authorization URL received')).toBeInTheDocument();
    }

    it('rejects non-http(s) authorize URLs', async () => {
      await expectAuthorizeUrlRejected('javascript:alert(1)');
    });

    it('rejects same-origin authorize URLs (open redirect prevention)', async () => {
      await expectAuthorizeUrlRejected(`${window.location.origin}/malicious-path`);
    });

    describe('HTTP localhost handling', () => {
      const originalDev = import.meta.env.DEV;
      afterEach(() => {
        import.meta.env.DEV = originalDev;
      });

      it('rejects HTTP URLs including localhost in production', async () => {
        import.meta.env.DEV = false;
        mockAuthorize('http://localhost:8080/authorize');

        submitLogin();

        expect(await screen.findByText('Invalid authorization URL received')).toBeInTheDocument();
      });

      it('allows HTTP localhost in development mode', async () => {
        import.meta.env.DEV = true;
        mockAuthorize('http://localhost:8080/authorize');

        submitLogin();

        await waitFor(() => {
          expect(window.location.href).toBe('http://localhost:8080/authorize');
        });
      });
    });
  });
});
