import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ComponentOriginsSection } from './ComponentOriginsSection';
import type { ComponentId } from '../../../api/types';
import { httpClient } from '../../../api/core/httpClient';

vi.mock('../../../api/core/httpClient', () => ({
  httpClient: {
    get: vi.fn(),
  },
}));

describe('ComponentOriginsSection', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    vi.clearAllMocks();
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
  });

  const renderComponent = (componentId: ComponentId) => {
    return render(
      <QueryClientProvider client={queryClient}>
        <ComponentOriginsSection componentId={componentId} />
      </QueryClientProvider>
    );
  };

  const createEmptyResponse = () => ({
    data: {
      componentId: 'comp-123',
      acquiredVia: [],
      purchasedFrom: [],
      builtBy: [],
      _links: { self: { href: '/test', method: 'GET' } },
    },
  });

  describe('loading state', () => {
    it('should display loading text while fetching origins', async () => {
      vi.mocked(httpClient.get).mockImplementation(
        () => new Promise(() => {})
      );

      renderComponent('comp-123' as ComponentId);

      expect(screen.getByText('Loading...')).toBeInTheDocument();
    });
  });

  describe('empty state', () => {
    it('should render nothing when no origins exist', async () => {
      vi.mocked(httpClient.get).mockResolvedValue(createEmptyResponse());

      const { container } = renderComponent('comp-123' as ComponentId);

      await waitFor(() => {
        expect(container.textContent).toBe('');
      });
    });
  });

  describe('displaying origins', () => {
    it('should display AcquiredVia relationship correctly', async () => {
      vi.mocked(httpClient.get).mockResolvedValue({
        data: {
          componentId: 'comp-123',
          acquiredVia: [
            {
              id: 'rel-1',
              acquiredEntityId: 'ae-123',
              acquiredEntityName: 'TechCorp',
              componentId: 'comp-123',
              componentName: 'SAP HR',
              createdAt: '2021-01-01T00:00:00Z',
              _links: { self: { href: '/test', method: 'GET' } },
            },
          ],
          purchasedFrom: [],
          builtBy: [],
          _links: { self: { href: '/test', method: 'GET' } },
        },
      });

      renderComponent('comp-123' as ComponentId);

      await waitFor(() => {
        expect(screen.getByText('TechCorp')).toBeInTheDocument();
        expect(screen.getByText('Acquired via')).toBeInTheDocument();
      });
    });

    it('should display PurchasedFrom relationship correctly', async () => {
      vi.mocked(httpClient.get).mockResolvedValue({
        data: {
          componentId: 'comp-123',
          acquiredVia: [],
          purchasedFrom: [
            {
              id: 'rel-2',
              vendorId: 'v-123',
              vendorName: 'SAP',
              componentId: 'comp-123',
              componentName: 'SAP HR',
              createdAt: '2021-01-01T00:00:00Z',
              _links: { self: { href: '/test', method: 'GET' } },
            },
          ],
          builtBy: [],
          _links: { self: { href: '/test', method: 'GET' } },
        },
      });

      renderComponent('comp-123' as ComponentId);

      await waitFor(() => {
        expect(screen.getByText('SAP')).toBeInTheDocument();
        expect(screen.getByText('Purchased from')).toBeInTheDocument();
      });
    });

    it('should display BuiltBy relationship correctly', async () => {
      vi.mocked(httpClient.get).mockResolvedValue({
        data: {
          componentId: 'comp-123',
          acquiredVia: [],
          purchasedFrom: [],
          builtBy: [
            {
              id: 'rel-3',
              internalTeamId: 'it-123',
              internalTeamName: 'Platform Engineering',
              componentId: 'comp-123',
              componentName: 'SAP HR',
              createdAt: '2021-01-01T00:00:00Z',
              _links: { self: { href: '/test', method: 'GET' } },
            },
          ],
          _links: { self: { href: '/test', method: 'GET' } },
        },
      });

      renderComponent('comp-123' as ComponentId);

      await waitFor(() => {
        expect(screen.getByText('Platform Engineering')).toBeInTheDocument();
        expect(screen.getByText('Built by')).toBeInTheDocument();
      });
    });

    it('should display multiple origins', async () => {
      vi.mocked(httpClient.get).mockResolvedValue({
        data: {
          componentId: 'comp-123',
          acquiredVia: [
            {
              id: 'rel-1',
              acquiredEntityId: 'ae-123',
              acquiredEntityName: 'TechCorp',
              componentId: 'comp-123',
              componentName: 'SAP HR',
              createdAt: '2021-01-01T00:00:00Z',
              _links: { self: { href: '/test', method: 'GET' } },
            },
          ],
          purchasedFrom: [],
          builtBy: [
            {
              id: 'rel-2',
              internalTeamId: 'it-123',
              internalTeamName: 'TechCorp Engineering',
              componentId: 'comp-123',
              componentName: 'SAP HR',
              createdAt: '2021-01-01T00:00:00Z',
              _links: { self: { href: '/test', method: 'GET' } },
            },
          ],
          _links: { self: { href: '/test', method: 'GET' } },
        },
      });

      renderComponent('comp-123' as ComponentId);

      await waitFor(() => {
        expect(screen.getByText('TechCorp')).toBeInTheDocument();
        expect(screen.getByText('Acquired via')).toBeInTheDocument();
        expect(screen.getByText('TechCorp Engineering')).toBeInTheDocument();
        expect(screen.getByText('Built by')).toBeInTheDocument();
      });
    });

    it('should display correct icons for each relationship type', async () => {
      vi.mocked(httpClient.get).mockResolvedValue({
        data: {
          componentId: 'comp-123',
          acquiredVia: [
            {
              id: 'rel-1',
              acquiredEntityId: 'ae-123',
              acquiredEntityName: 'TechCorp',
              componentId: 'comp-123',
              componentName: 'SAP HR',
              createdAt: '2021-01-01T00:00:00Z',
              _links: { self: { href: '/test', method: 'GET' } },
            },
          ],
          purchasedFrom: [
            {
              id: 'rel-2',
              vendorId: 'v-123',
              vendorName: 'SAP',
              componentId: 'comp-123',
              componentName: 'SAP HR',
              createdAt: '2021-01-01T00:00:00Z',
              _links: { self: { href: '/test', method: 'GET' } },
            },
          ],
          builtBy: [
            {
              id: 'rel-3',
              internalTeamId: 'it-123',
              internalTeamName: 'Platform Engineering',
              componentId: 'comp-123',
              componentName: 'SAP HR',
              createdAt: '2021-01-01T00:00:00Z',
              _links: { self: { href: '/test', method: 'GET' } },
            },
          ],
          _links: { self: { href: '/test', method: 'GET' } },
        },
      });

      renderComponent('comp-123' as ComponentId);

      await waitFor(() => {
        expect(screen.getByText('Origins')).toBeInTheDocument();
      });
    });
  });

  describe('API integration', () => {
    it('should fetch origins from correct endpoint', async () => {
      vi.mocked(httpClient.get).mockResolvedValue(createEmptyResponse());

      renderComponent('comp-123' as ComponentId);

      await waitFor(() => {
        expect(httpClient.get).toHaveBeenCalledWith('/api/v1/components/comp-123/origins');
      });
    });

    it('should not fetch when componentId is empty', async () => {
      vi.mocked(httpClient.get).mockResolvedValue(createEmptyResponse());

      renderComponent('' as ComponentId);

      await new Promise((resolve) => setTimeout(resolve, 100));

      expect(httpClient.get).not.toHaveBeenCalled();
    });
  });
});
