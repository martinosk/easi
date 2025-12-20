import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { http, HttpResponse } from 'msw';
import { DetailsSidebar } from './DetailsSidebar';
import { useAppStore } from '../../../store/appStore';
import { createMantineTestWrapper, seedDb, server } from '../../../test/helpers';
import type { Capability, CapabilityId, ComponentId } from '../../../api/types';

const API_BASE = 'http://localhost:8080';

vi.mock('../../../store/appStore', () => ({
  useAppStore: vi.fn(),
}));

vi.mock('../../components/components/EditComponentDialog', () => ({
  EditComponentDialog: ({ isOpen, onClose }: { isOpen: boolean; onClose: () => void }) =>
    isOpen ? (
      <div data-testid="edit-dialog">
        <button onClick={onClose}>Close Dialog</button>
      </div>
    ) : null,
}));

const mockCapability: Capability = {
  id: 'cap-1' as CapabilityId,
  name: 'Financial Management',
  level: 'L1',
  description: 'Manage financial operations',
  createdAt: '2024-01-01',
  _links: { self: '/api/v1/capabilities/cap-1' },
};

const mockComponent = {
  id: 'comp-1' as ComponentId,
  name: 'SAP Finance',
  description: 'Financial system',
  createdAt: '2024-01-01T00:00:00Z',
  _links: { self: '/api/v1/components/comp-1' },
};

const createMockStore = (overrides: Record<string, unknown> = {}) => ({
  capabilities: [],
  ...overrides,
});

describe('DetailsSidebar', () => {
  const defaultProps = {
    selectedCapability: null,
    selectedComponentId: null,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    seedDb({
      components: [mockComponent],
      capabilities: [mockCapability],
    });
    vi.mocked(useAppStore).mockImplementation((selector: (state: unknown) => unknown) =>
      selector(createMockStore())
    );
  });

  const renderSidebar = (props = defaultProps) => {
    const { Wrapper } = createMantineTestWrapper();
    return render(<DetailsSidebar {...props} />, { wrapper: Wrapper });
  };

  describe('empty state', () => {
    it('shows placeholder message when nothing is selected', () => {
      renderSidebar();
      expect(screen.getByText('Select a capability or application to view details')).toBeInTheDocument();
    });

    it('shows Details title in empty state', () => {
      renderSidebar();
      expect(screen.getByText('Details')).toBeInTheDocument();
    });
  });

  describe('capability details', () => {
    it('shows capability details when capability is selected', () => {
      renderSidebar({ ...defaultProps, selectedCapability: mockCapability });

      expect(screen.getByText('Capability Details')).toBeInTheDocument();
      expect(screen.getByText('Financial Management')).toBeInTheDocument();
      expect(screen.getByText('L1')).toBeInTheDocument();
    });

    it('shows capability description when available', () => {
      renderSidebar({ ...defaultProps, selectedCapability: mockCapability });
      expect(screen.getByText('Manage financial operations')).toBeInTheDocument();
    });
  });

  describe('application details', () => {
    it('shows application details when componentId is selected and component is loaded', async () => {
      renderSidebar({ ...defaultProps, selectedComponentId: 'comp-1' as ComponentId });

      await waitFor(() => {
        expect(screen.getByText('Application Details')).toBeInTheDocument();
      });
      expect(screen.getByText('SAP Finance')).toBeInTheDocument();
    });

    it('shows loading state when fetching component from API', async () => {
      server.use(
        http.get(`${API_BASE}/api/v1/components`, async () => {
          await new Promise((resolve) => setTimeout(resolve, 100));
          return HttpResponse.json({ data: [], _links: { self: '/api/v1/components' } });
        })
      );

      renderSidebar({ ...defaultProps, selectedComponentId: 'comp-2' as ComponentId });

      expect(screen.getByText('Loading...')).toBeInTheDocument();
    });

    it('shows error state when component fetch fails', async () => {
      server.use(
        http.get(`${API_BASE}/api/v1/components`, () => {
          return HttpResponse.json({ error: 'Network error' }, { status: 500 });
        })
      );

      renderSidebar({ ...defaultProps, selectedComponentId: 'comp-2' as ComponentId });

      await waitFor(() => {
        expect(screen.getByText('Failed to load application details')).toBeInTheDocument();
      });
    });

    it('opens edit dialog when Edit button is clicked', async () => {
      renderSidebar({ ...defaultProps, selectedComponentId: 'comp-1' as ComponentId });

      await waitFor(() => {
        expect(screen.getByText('SAP Finance')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Edit'));
      expect(screen.getByTestId('edit-dialog')).toBeInTheDocument();
    });

    it('closes edit dialog when dialog close is triggered', async () => {
      renderSidebar({ ...defaultProps, selectedComponentId: 'comp-1' as ComponentId });

      await waitFor(() => {
        expect(screen.getByText('SAP Finance')).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText('Edit'));
      expect(screen.getByTestId('edit-dialog')).toBeInTheDocument();

      fireEvent.click(screen.getByText('Close Dialog'));
      expect(screen.queryByTestId('edit-dialog')).not.toBeInTheDocument();
    });
  });

  describe('priority of selection', () => {
    it('shows capability details when both capability and component are selected', async () => {
      renderSidebar({
        ...defaultProps,
        selectedCapability: mockCapability,
        selectedComponentId: 'comp-1' as ComponentId,
      });

      expect(screen.getByText('Capability Details')).toBeInTheDocument();
      expect(screen.queryByText('Application Details')).not.toBeInTheDocument();
    });
  });
});
