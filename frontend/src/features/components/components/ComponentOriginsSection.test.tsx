import { MantineProvider } from '@mantine/core';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { httpClient } from '../../../api/core/httpClient';
import type { ComponentId } from '../../../api/types';
import { theme } from '../../../theme/mantine';
import { ComponentOriginsSection } from './ComponentOriginsSection';

vi.mock('../../../api/core/httpClient', () => ({
  httpClient: {
    get: vi.fn(),
  },
}));

const SELF_LINKS = { self: { href: '/test', method: 'GET' } };
const BASE = {
  componentId: 'comp-123',
  componentName: 'SAP HR',
  createdAt: '2021-01-01T00:00:00Z',
  _links: SELF_LINKS,
};

const acquiredVia = (id: string, name: string) => ({ id, acquiredEntityId: `ae-${id}`, acquiredEntityName: name, ...BASE });
const purchasedFrom = (id: string, name: string) => ({ id, vendorId: `v-${id}`, vendorName: name, ...BASE });
const builtBy = (id: string, name: string) => ({ id, internalTeamId: `it-${id}`, internalTeamName: name, ...BASE });

const mockOriginsResponse = (overrides: {
  acquiredVia?: ReturnType<typeof acquiredVia>[];
  purchasedFrom?: ReturnType<typeof purchasedFrom>[];
  builtBy?: ReturnType<typeof builtBy>[];
}) => {
  vi.mocked(httpClient.get).mockResolvedValue({
    data: {
      componentId: 'comp-123',
      acquiredVia: [],
      purchasedFrom: [],
      builtBy: [],
      _links: SELF_LINKS,
      ...overrides,
    },
  });
};

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
        <MantineProvider theme={theme}>
          <ComponentOriginsSection componentId={componentId} />
        </MantineProvider>
      </QueryClientProvider>,
    );
  };

  describe('loading state', () => {
    it('should display loading text while fetching origins', async () => {
      vi.mocked(httpClient.get).mockImplementation(() => new Promise(() => {}));

      renderComponent('comp-123' as ComponentId);

      expect(screen.getByText('Loading...')).toBeInTheDocument();
    });
  });

  describe('empty state', () => {
    it('should render nothing when no origins exist', async () => {
      mockOriginsResponse({});

      renderComponent('comp-123' as ComponentId);

      await waitFor(() => {
        expect(screen.queryByText('Origins')).not.toBeInTheDocument();
      });
    });
  });

  describe('displaying origins', () => {
    const singleRelationshipCases = [
      {
        name: 'AcquiredVia',
        setup: () => mockOriginsResponse({ acquiredVia: [acquiredVia('rel-1', 'TechCorp')] }),
        expectedEntity: 'TechCorp',
        expectedLabel: 'Acquired via',
      },
      {
        name: 'PurchasedFrom',
        setup: () => mockOriginsResponse({ purchasedFrom: [purchasedFrom('rel-2', 'SAP')] }),
        expectedEntity: 'SAP',
        expectedLabel: 'Purchased from',
      },
      {
        name: 'BuiltBy',
        setup: () => mockOriginsResponse({ builtBy: [builtBy('rel-3', 'Platform Engineering')] }),
        expectedEntity: 'Platform Engineering',
        expectedLabel: 'Built by',
      },
    ];

    singleRelationshipCases.forEach(({ name, setup, expectedEntity, expectedLabel }) => {
      it(`should display ${name} relationship correctly`, async () => {
        setup();

        renderComponent('comp-123' as ComponentId);

        await waitFor(() => {
          expect(screen.getByText(expectedEntity)).toBeInTheDocument();
          expect(screen.getByText(expectedLabel)).toBeInTheDocument();
        });
      });
    });

    it('should display multiple origins of different types together', async () => {
      mockOriginsResponse({
        acquiredVia: [acquiredVia('rel-1', 'TechCorp')],
        builtBy: [builtBy('rel-2', 'TechCorp Engineering')],
      });

      renderComponent('comp-123' as ComponentId);

      await waitFor(() => {
        expect(screen.getByText('TechCorp')).toBeInTheDocument();
        expect(screen.getByText('Acquired via')).toBeInTheDocument();
        expect(screen.getByText('TechCorp Engineering')).toBeInTheDocument();
        expect(screen.getByText('Built by')).toBeInTheDocument();
      });
    });

    it('should render section header when origins are present', async () => {
      mockOriginsResponse({
        acquiredVia: [acquiredVia('rel-1', 'TechCorp')],
        purchasedFrom: [purchasedFrom('rel-2', 'SAP')],
        builtBy: [builtBy('rel-3', 'Platform Engineering')],
      });

      renderComponent('comp-123' as ComponentId);

      await waitFor(() => {
        expect(screen.getByText('Origins')).toBeInTheDocument();
      });
    });
  });

  describe('API integration', () => {
    it('should fetch origins from correct endpoint', async () => {
      mockOriginsResponse({});

      renderComponent('comp-123' as ComponentId);

      await waitFor(() => {
        expect(httpClient.get).toHaveBeenCalledWith('/api/v1/components/comp-123/origins');
      });
    });

    it('should not fetch when componentId is empty', async () => {
      mockOriginsResponse({});

      renderComponent('' as ComponentId);

      await new Promise((resolve) => setTimeout(resolve, 100));

      expect(httpClient.get).not.toHaveBeenCalled();
    });
  });
});
