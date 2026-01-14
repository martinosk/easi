import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
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

function renderLoginPage(initialRoute = '/login') {
  return render(
    <MemoryRouter initialEntries={[initialRoute]}>
      <LoginPage />
    </MemoryRouter>
  );
}

describe('LoginPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    delete (window as { location?: unknown }).location;
    (window as { location: { href: string } }).location = { href: '' };
  });

  describe('returnUrl parameter handling', () => {
    it('should pass returnUrl to initiateLogin when present', async () => {
      mockInitiateLogin.mockResolvedValue({
        _links: { authorize: 'https://idp.example.com/authorize' },
      });

      renderLoginPage('/login?returnUrl=https%3A%2F%2Fapp.example.com%2Fdashboard');

      fireEvent.change(screen.getByLabelText(/email/i), {
        target: { value: 'user@example.com' },
      });
      fireEvent.click(screen.getByRole('button', { name: /continue/i }));

      await waitFor(() => {
        expect(mockInitiateLogin).toHaveBeenCalledWith(
          'user@example.com',
          'https://app.example.com/dashboard'
        );
      });
    });

    it('should NOT double-decode returnUrl (security)', async () => {
      mockInitiateLogin.mockResolvedValue({
        _links: { authorize: 'https://idp.example.com/authorize' },
      });

      renderLoginPage('/login?returnUrl=https%253A%252F%252Fevil.com');

      fireEvent.change(screen.getByLabelText(/email/i), {
        target: { value: 'user@example.com' },
      });
      fireEvent.click(screen.getByRole('button', { name: /continue/i }));

      await waitFor(() => {
        expect(mockInitiateLogin).toHaveBeenCalledWith(
          'user@example.com',
          'https%3A%2F%2Fevil.com'
        );
      });
    });

    it('should call initiateLogin without returnUrl when not present', async () => {
      mockInitiateLogin.mockResolvedValue({
        _links: { authorize: 'https://idp.example.com/authorize' },
      });

      renderLoginPage('/login');

      fireEvent.change(screen.getByLabelText(/email/i), {
        target: { value: 'user@example.com' },
      });
      fireEvent.click(screen.getByRole('button', { name: /continue/i }));

      await waitFor(() => {
        expect(mockInitiateLogin).toHaveBeenCalledWith('user@example.com', undefined);
      });
    });
  });

  describe('form validation', () => {
    it('should show error when email is empty', async () => {
      renderLoginPage();

      fireEvent.click(screen.getByRole('button', { name: /continue/i }));

      expect(await screen.findByText('Email is required')).toBeInTheDocument();
      expect(mockInitiateLogin).not.toHaveBeenCalled();
    });
  });

  describe('authorize URL validation', () => {
    it('should reject non-http(s) authorize URLs', async () => {
      mockInitiateLogin.mockResolvedValue({
        _links: { authorize: 'javascript:alert(1)' },
      });

      renderLoginPage();

      fireEvent.change(screen.getByLabelText(/email/i), {
        target: { value: 'user@example.com' },
      });
      fireEvent.click(screen.getByRole('button', { name: /continue/i }));

      expect(await screen.findByText('Invalid authorization URL received')).toBeInTheDocument();
    });

    it('should reject same-origin authorize URLs (open redirect prevention)', async () => {
      mockInitiateLogin.mockResolvedValue({
        _links: { authorize: `${window.location.origin}/malicious-path` },
      });

      renderLoginPage();

      fireEvent.change(screen.getByLabelText(/email/i), {
        target: { value: 'user@example.com' },
      });
      fireEvent.click(screen.getByRole('button', { name: /continue/i }));

      expect(await screen.findByText('Invalid authorization URL received')).toBeInTheDocument();
    });

    it('should reject HTTP URLs including localhost in production', async () => {
      const originalDev = import.meta.env.DEV;
      import.meta.env.DEV = false;

      mockInitiateLogin.mockResolvedValue({
        _links: { authorize: 'http://localhost:8080/authorize' },
      });

      renderLoginPage();

      fireEvent.change(screen.getByLabelText(/email/i), {
        target: { value: 'user@example.com' },
      });
      fireEvent.click(screen.getByRole('button', { name: /continue/i }));

      expect(await screen.findByText('Invalid authorization URL received')).toBeInTheDocument();

      import.meta.env.DEV = originalDev;
    });

    it('should allow HTTP localhost in development mode', async () => {
      const originalDev = import.meta.env.DEV;
      import.meta.env.DEV = true;

      mockInitiateLogin.mockResolvedValue({
        _links: { authorize: 'http://localhost:8080/authorize' },
      });

      renderLoginPage();

      fireEvent.change(screen.getByLabelText(/email/i), {
        target: { value: 'user@example.com' },
      });
      fireEvent.click(screen.getByRole('button', { name: /continue/i }));

      await waitFor(() => {
        expect(window.location.href).toBe('http://localhost:8080/authorize');
      });

      import.meta.env.DEV = originalDev;
    });
  });
});
