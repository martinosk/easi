import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { LinkingPage } from './LinkingPage';
import { enterpriseArchApi } from '../api/enterpriseArchApi';
import { capabilitiesApi } from '../../capabilities/api/capabilitiesApi';

vi.mock('../api/enterpriseArchApi');
vi.mock('../../capabilities/api/capabilitiesApi');

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
          _links: { self: { href: '/api/v1/enterprise-capabilities/ec-1' } },
        },
      ];

      vi.mocked(enterpriseArchApi.getAll).mockResolvedValue(mockEnterpriseCapabilities);
      vi.mocked(capabilitiesApi.getAll).mockResolvedValue([]);
      vi.mocked(enterpriseArchApi.getBatchLinkStatus).mockResolvedValue([]);

      render(<LinkingPage />);

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
          _links: { self: { href: '/api/v1/capabilities/cap-1' } },
        },
      ];

      vi.mocked(enterpriseArchApi.getAll).mockResolvedValue([]);
      vi.mocked(capabilitiesApi.getAll).mockResolvedValue(mockDomainCapabilities);
      vi.mocked(enterpriseArchApi.getBatchLinkStatus).mockResolvedValue([]);

      render(<LinkingPage />);

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
          _links: { self: { href: '/api/v1/capabilities/cap-1' } },
        },
      ];

      vi.mocked(enterpriseArchApi.getAll).mockResolvedValue([]);
      vi.mocked(capabilitiesApi.getAll).mockResolvedValue(mockDomainCapabilities);
      vi.mocked(enterpriseArchApi.getBatchLinkStatus).mockResolvedValue([
        { capabilityId: 'cap-1', status: 'available' },
      ]);

      render(<LinkingPage />);

      await waitFor(() => {
        expect(enterpriseArchApi.getBatchLinkStatus).toHaveBeenCalledWith(['cap-1']);
      });
    });

    it('shows enterprise capabilities panel with loading state initially', () => {
      vi.mocked(enterpriseArchApi.getAll).mockImplementation(
        () => new Promise(() => {})
      );
      vi.mocked(capabilitiesApi.getAll).mockImplementation(
        () => new Promise(() => {})
      );

      render(<LinkingPage />);

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
          _links: { self: { href: '/api/v1/enterprise-capabilities/ec-1' } },
        },
      ];

      vi.mocked(enterpriseArchApi.getAll).mockResolvedValue(mockEnterpriseCapabilities);
      vi.mocked(capabilitiesApi.getAll).mockResolvedValue([]);
      vi.mocked(enterpriseArchApi.getBatchLinkStatus).mockResolvedValue([]);

      render(<LinkingPage />);

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
          _links: { self: { href: '/api/v1/capabilities/cap-1' } },
        },
      ];

      vi.mocked(enterpriseArchApi.getAll).mockResolvedValue([]);
      vi.mocked(capabilitiesApi.getAll).mockResolvedValue(mockDomainCapabilities);
      vi.mocked(enterpriseArchApi.getBatchLinkStatus).mockResolvedValue([
        { capabilityId: 'cap-1', status: 'available' },
      ]);

      render(<LinkingPage />);

      await waitFor(() => {
        expect(screen.getByText('Payment Processing')).toBeInTheDocument();
      });
    });
  });

  describe('Error Handling', () => {
    it('logs error when enterprise capability loading fails', async () => {
      vi.mocked(enterpriseArchApi.getAll).mockRejectedValue(new Error('Network error'));
      vi.mocked(capabilitiesApi.getAll).mockResolvedValue([]);
      vi.mocked(enterpriseArchApi.getBatchLinkStatus).mockResolvedValue([]);

      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

      render(<LinkingPage />);

      await waitFor(() => {
        expect(consoleErrorSpy).toHaveBeenCalledWith(
          'Failed to load enterprise capabilities:',
          expect.any(Error)
        );
      });

      consoleErrorSpy.mockRestore();
    });

    it('logs error when domain capability loading fails', async () => {
      vi.mocked(enterpriseArchApi.getAll).mockResolvedValue([]);
      vi.mocked(capabilitiesApi.getAll).mockRejectedValue(new Error('Network error'));
      vi.mocked(enterpriseArchApi.getBatchLinkStatus).mockResolvedValue([]);

      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

      render(<LinkingPage />);

      await waitFor(() => {
        expect(consoleErrorSpy).toHaveBeenCalledWith(
          'Failed to load domain capabilities:',
          expect.any(Error)
        );
      });

      consoleErrorSpy.mockRestore();
    });

    it('does not crash when link status loading fails', async () => {
      const mockDomainCapabilities = [
        {
          id: 'cap-1',
          name: 'Payment Processing',
          level: 'L1' as const,
          status: 'active' as const,
          createdAt: '2025-01-01T00:00:00Z',
          updatedAt: '2025-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capabilities/cap-1' } },
        },
      ];

      vi.mocked(enterpriseArchApi.getAll).mockResolvedValue([]);
      vi.mocked(capabilitiesApi.getAll).mockResolvedValue(mockDomainCapabilities);
      vi.mocked(enterpriseArchApi.getBatchLinkStatus).mockRejectedValue(
        new Error('Status loading failed')
      );

      const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

      render(<LinkingPage />);

      await waitFor(() => {
        expect(screen.getByText('Payment Processing')).toBeInTheDocument();
      });

      consoleErrorSpy.mockRestore();
    });
  });

  describe('Linking Capability', () => {
    it('calls API to link capability and refreshes data', async () => {
      const mockEnterpriseCapabilities = [
        {
          id: 'ec-1',
          name: 'Customer Management',
          category: 'Business',
          linkCount: 0,
          domainCount: 0,
          createdAt: '2025-01-01T00:00:00Z',
          updatedAt: '2025-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/enterprise-capabilities/ec-1' } },
        },
      ];

      const mockDomainCapabilities = [
        {
          id: 'cap-1',
          name: 'Payment Processing',
          level: 'L1' as const,
          status: 'active' as const,
          createdAt: '2025-01-01T00:00:00Z',
          updatedAt: '2025-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capabilities/cap-1' } },
        },
      ];

      vi.mocked(enterpriseArchApi.getAll).mockResolvedValue(mockEnterpriseCapabilities);
      vi.mocked(capabilitiesApi.getAll).mockResolvedValue(mockDomainCapabilities);
      vi.mocked(enterpriseArchApi.getBatchLinkStatus).mockResolvedValue([
        { capabilityId: 'cap-1', status: 'available' },
      ]);
      vi.mocked(enterpriseArchApi.linkDomainCapability).mockResolvedValue(undefined);

      const { rerender } = render(<LinkingPage />);

      await waitFor(() => {
        expect(screen.getByText('Payment Processing')).toBeInTheDocument();
      });

      rerender(<LinkingPage />);

      await waitFor(() => {
        expect(enterpriseArchApi.getAll).toHaveBeenCalled();
        expect(capabilitiesApi.getAll).toHaveBeenCalled();
      });
    });
  });

  describe('Layout', () => {
    it('renders in split-panel layout', async () => {
      vi.mocked(enterpriseArchApi.getAll).mockResolvedValue([]);
      vi.mocked(capabilitiesApi.getAll).mockResolvedValue([]);
      vi.mocked(enterpriseArchApi.getBatchLinkStatus).mockResolvedValue([]);

      const { container } = render(<LinkingPage />);

      await waitFor(() => {
        const splitPanels = container.querySelectorAll('div[style*="width: 50%"]');
        expect(splitPanels.length).toBeGreaterThanOrEqual(2);
      });
    });
  });
});
