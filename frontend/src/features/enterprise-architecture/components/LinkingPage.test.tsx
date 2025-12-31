import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { LinkingPage } from './LinkingPage';
import { enterpriseArchApi } from '../api/enterpriseArchApi';
import { capabilitiesApi } from '../../capabilities/api/capabilitiesApi';

vi.mock('../api/enterpriseArchApi');
vi.mock('../../capabilities/api/capabilitiesApi');
vi.mock('react-hot-toast', () => ({
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

function createTestQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
      mutations: {
        retry: false,
      },
    },
  });
}

function renderWithQueryClient(ui: React.ReactNode) {
  const queryClient = createTestQueryClient();
  return render(
    <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>
  );
}

describe('LinkingPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Data Loading', () => {
    it('loads enterprise capabilities on mount', async () => {
      const mockEnterpriseCapabilities = [
        {
          id: 'ec-1',
          name: 'Customer Management',
          category: 'Business',
          linkCount: 0,
          domainCount: 0,
          createdAt: '2025-01-01T00:00:00Z',
          updatedAt: '2025-01-01T00:00:00Z',
          _links: { self: '/api/v1/enterprise-capabilities/ec-1', links: '', strategicImportance: '' },
        },
      ];

      vi.mocked(enterpriseArchApi.getAll).mockResolvedValue(mockEnterpriseCapabilities);
      vi.mocked(capabilitiesApi.getAll).mockResolvedValue([]);
      vi.mocked(enterpriseArchApi.getBatchLinkStatus).mockResolvedValue([]);

      renderWithQueryClient(<LinkingPage />);

      await waitFor(() => {
        expect(enterpriseArchApi.getAll).toHaveBeenCalledTimes(1);
      });
    });

    it('loads domain capabilities on mount', async () => {
      const mockDomainCapabilities = [
        {
          id: 'cap-1',
          name: 'Payment Processing',
          level: 'L1' as const,
          status: 'active' as const,
          createdAt: '2025-01-01T00:00:00Z',
          updatedAt: '2025-01-01T00:00:00Z',
          _links: { self: '/api/v1/capabilities/cap-1' },
        },
      ];

      vi.mocked(enterpriseArchApi.getAll).mockResolvedValue([]);
      vi.mocked(capabilitiesApi.getAll).mockResolvedValue(mockDomainCapabilities);
      vi.mocked(enterpriseArchApi.getBatchLinkStatus).mockResolvedValue([]);

      renderWithQueryClient(<LinkingPage />);

      await waitFor(() => {
        expect(capabilitiesApi.getAll).toHaveBeenCalledTimes(1);
      });
    });

    it('loads link statuses after loading capabilities', async () => {
      const mockDomainCapabilities = [
        {
          id: 'cap-1',
          name: 'Payment Processing',
          level: 'L1' as const,
          status: 'active' as const,
          createdAt: '2025-01-01T00:00:00Z',
          updatedAt: '2025-01-01T00:00:00Z',
          _links: { self: '/api/v1/capabilities/cap-1' },
        },
      ];

      vi.mocked(enterpriseArchApi.getAll).mockResolvedValue([]);
      vi.mocked(capabilitiesApi.getAll).mockResolvedValue(mockDomainCapabilities);
      vi.mocked(enterpriseArchApi.getBatchLinkStatus).mockResolvedValue([
        { capabilityId: 'cap-1', status: 'available' },
      ]);

      renderWithQueryClient(<LinkingPage />);

      await waitFor(() => {
        expect(enterpriseArchApi.getBatchLinkStatus).toHaveBeenCalledWith(['cap-1']);
      });
    });

    it('shows loading state initially', () => {
      vi.mocked(enterpriseArchApi.getAll).mockImplementation(
        () => new Promise(() => {})
      );
      vi.mocked(capabilitiesApi.getAll).mockImplementation(
        () => new Promise(() => {})
      );

      renderWithQueryClient(<LinkingPage />);

      expect(screen.queryByText('Loading domain capabilities...')).toBeInTheDocument();
    });

    it('displays loaded enterprise capabilities', async () => {
      const mockEnterpriseCapabilities = [
        {
          id: 'ec-1',
          name: 'Customer Management',
          category: 'Business',
          linkCount: 0,
          domainCount: 0,
          createdAt: '2025-01-01T00:00:00Z',
          updatedAt: '2025-01-01T00:00:00Z',
          _links: { self: '/api/v1/enterprise-capabilities/ec-1', links: '', strategicImportance: '' },
        },
      ];

      vi.mocked(enterpriseArchApi.getAll).mockResolvedValue(mockEnterpriseCapabilities);
      vi.mocked(capabilitiesApi.getAll).mockResolvedValue([]);
      vi.mocked(enterpriseArchApi.getBatchLinkStatus).mockResolvedValue([]);

      renderWithQueryClient(<LinkingPage />);

      await waitFor(() => {
        expect(screen.getByText('Customer Management')).toBeInTheDocument();
      });
    });

    it('displays loaded domain capabilities', async () => {
      const mockDomainCapabilities = [
        {
          id: 'cap-1',
          name: 'Payment Processing',
          level: 'L1' as const,
          status: 'active' as const,
          createdAt: '2025-01-01T00:00:00Z',
          updatedAt: '2025-01-01T00:00:00Z',
          _links: { self: '/api/v1/capabilities/cap-1' },
        },
      ];

      vi.mocked(enterpriseArchApi.getAll).mockResolvedValue([]);
      vi.mocked(capabilitiesApi.getAll).mockResolvedValue(mockDomainCapabilities);
      vi.mocked(enterpriseArchApi.getBatchLinkStatus).mockResolvedValue([
        { capabilityId: 'cap-1', status: 'available' },
      ]);

      renderWithQueryClient(<LinkingPage />);

      await waitFor(() => {
        expect(screen.getByText('Payment Processing')).toBeInTheDocument();
      });
    });
  });

  describe('Error Handling', () => {
    it('does not crash when link status loading fails', async () => {
      const mockDomainCapabilities = [
        {
          id: 'cap-1',
          name: 'Payment Processing',
          level: 'L1' as const,
          status: 'active' as const,
          createdAt: '2025-01-01T00:00:00Z',
          updatedAt: '2025-01-01T00:00:00Z',
          _links: { self: '/api/v1/capabilities/cap-1' },
        },
      ];

      vi.mocked(enterpriseArchApi.getAll).mockResolvedValue([]);
      vi.mocked(capabilitiesApi.getAll).mockResolvedValue(mockDomainCapabilities);
      vi.mocked(enterpriseArchApi.getBatchLinkStatus).mockRejectedValue(
        new Error('Status loading failed')
      );

      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

      renderWithQueryClient(<LinkingPage />);

      await waitFor(() => {
        expect(screen.getByText('Payment Processing')).toBeInTheDocument();
      });

      consoleErrorSpy.mockRestore();
    });
  });

  describe('Layout', () => {
    it('renders in split-panel layout', async () => {
      vi.mocked(enterpriseArchApi.getAll).mockResolvedValue([]);
      vi.mocked(capabilitiesApi.getAll).mockResolvedValue([]);
      vi.mocked(enterpriseArchApi.getBatchLinkStatus).mockResolvedValue([]);

      const { container } = renderWithQueryClient(<LinkingPage />);

      await waitFor(() => {
        const splitPanels = container.querySelectorAll('div[style*="width: 50%"]');
        expect(splitPanels.length).toBeGreaterThanOrEqual(2);
      });
    });
  });
});
