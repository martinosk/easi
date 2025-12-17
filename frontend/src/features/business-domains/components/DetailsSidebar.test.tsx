import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { DetailsSidebar } from './DetailsSidebar';
import { useAppStore } from '../../../store/appStore';
import { useComponentDetails } from '../hooks/useComponentDetails';
import type { Capability, CapabilityId, ComponentId, Component } from '../../../api/types';

vi.mock('../../../store/appStore', () => ({
  useAppStore: vi.fn(),
}));

vi.mock('../hooks/useComponentDetails', () => ({
  useComponentDetails: vi.fn(),
}));

vi.mock('../../components/components/EditComponentDialog', () => ({
  EditComponentDialog: ({ isOpen, onClose }: { isOpen: boolean; onClose: () => void }) =>
    isOpen ? (
      <div data-testid="edit-dialog">
        <button onClick={onClose}>Close Dialog</button>
      </div>
    ) : null,
}));

describe('DetailsSidebar', () => {
  const mockCapability: Capability = {
    id: 'cap-1' as CapabilityId,
    name: 'Financial Management',
    level: 'L1',
    description: 'Manage financial operations',
    createdAt: '2024-01-01',
    _links: { self: { href: '/api/v1/capabilities/cap-1' } },
  };

  const mockComponent: Component = {
    id: 'comp-1' as ComponentId,
    name: 'SAP Finance',
    description: 'Financial system',
    createdAt: '2024-01-01T00:00:00Z',
    _links: { self: '/api/v1/components/comp-1' },
  };

  const defaultProps = {
    selectedCapability: null,
    selectedComponentId: null,
    onCloseCapability: vi.fn(),
    onCloseApplication: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useAppStore).mockImplementation((selector) => {
      const state = {
        components: [mockComponent],
        capabilities: [],
        capabilityRealizations: [],
      };
      return selector(state as any);
    });
    vi.mocked(useComponentDetails).mockReturnValue({
      component: null,
      isLoading: false,
      error: null,
    });
  });

  describe('empty state', () => {
    it('shows placeholder message when nothing is selected', () => {
      render(<DetailsSidebar {...defaultProps} />);

      expect(screen.getByText('Select a capability or application to view details')).toBeInTheDocument();
    });

    it('shows Details title in empty state', () => {
      render(<DetailsSidebar {...defaultProps} />);

      expect(screen.getByText('Details')).toBeInTheDocument();
    });
  });

  describe('capability details', () => {
    it('shows capability details when capability is selected', () => {
      render(<DetailsSidebar {...defaultProps} selectedCapability={mockCapability} />);

      expect(screen.getByText('Capability Details')).toBeInTheDocument();
      expect(screen.getByText('Financial Management')).toBeInTheDocument();
      expect(screen.getByText('L1')).toBeInTheDocument();
    });

    it('shows capability description when available', () => {
      render(<DetailsSidebar {...defaultProps} selectedCapability={mockCapability} />);

      expect(screen.getByText('Manage financial operations')).toBeInTheDocument();
    });

    it('calls onCloseCapability when close button is clicked', () => {
      const onCloseCapability = vi.fn();
      render(
        <DetailsSidebar
          {...defaultProps}
          selectedCapability={mockCapability}
          onCloseCapability={onCloseCapability}
        />
      );

      fireEvent.click(screen.getByRole('button', { name: /close/i }));

      expect(onCloseCapability).toHaveBeenCalledTimes(1);
    });
  });

  describe('application details', () => {
    it('shows application details when componentId is selected and component is in store', () => {
      render(
        <DetailsSidebar {...defaultProps} selectedComponentId={'comp-1' as ComponentId} />
      );

      expect(screen.getByText('Application Details')).toBeInTheDocument();
      expect(screen.getByText('SAP Finance')).toBeInTheDocument();
    });

    it('shows loading state when fetching component from API', () => {
      vi.mocked(useAppStore).mockImplementation((selector) => {
        const state = {
          components: [],
          capabilities: [],
          capabilityRealizations: [],
        };
        return selector(state as any);
      });
      vi.mocked(useComponentDetails).mockReturnValue({
        component: null,
        isLoading: true,
        error: null,
      });

      render(
        <DetailsSidebar {...defaultProps} selectedComponentId={'comp-2' as ComponentId} />
      );

      expect(screen.getByText('Loading...')).toBeInTheDocument();
    });

    it('shows error state when component fetch fails', () => {
      vi.mocked(useAppStore).mockImplementation((selector) => {
        const state = {
          components: [],
          capabilities: [],
          capabilityRealizations: [],
        };
        return selector(state as any);
      });
      vi.mocked(useComponentDetails).mockReturnValue({
        component: null,
        isLoading: false,
        error: new Error('Network error'),
      });

      render(
        <DetailsSidebar {...defaultProps} selectedComponentId={'comp-2' as ComponentId} />
      );

      expect(screen.getByText('Failed to load application details')).toBeInTheDocument();
    });

    it('calls onCloseApplication when close button is clicked', () => {
      const onCloseApplication = vi.fn();
      render(
        <DetailsSidebar
          {...defaultProps}
          selectedComponentId={'comp-1' as ComponentId}
          onCloseApplication={onCloseApplication}
        />
      );

      fireEvent.click(screen.getByRole('button', { name: /close/i }));

      expect(onCloseApplication).toHaveBeenCalledTimes(1);
    });

    it('opens edit dialog when Edit button is clicked', () => {
      render(
        <DetailsSidebar {...defaultProps} selectedComponentId={'comp-1' as ComponentId} />
      );

      fireEvent.click(screen.getByText('Edit'));

      expect(screen.getByTestId('edit-dialog')).toBeInTheDocument();
    });

    it('closes edit dialog when dialog close is triggered', () => {
      render(
        <DetailsSidebar {...defaultProps} selectedComponentId={'comp-1' as ComponentId} />
      );

      fireEvent.click(screen.getByText('Edit'));
      expect(screen.getByTestId('edit-dialog')).toBeInTheDocument();

      fireEvent.click(screen.getByText('Close Dialog'));
      expect(screen.queryByTestId('edit-dialog')).not.toBeInTheDocument();
    });
  });

  describe('priority of selection', () => {
    it('shows capability details when both capability and component are selected', () => {
      render(
        <DetailsSidebar
          {...defaultProps}
          selectedCapability={mockCapability}
          selectedComponentId={'comp-1' as ComponentId}
        />
      );

      expect(screen.getByText('Capability Details')).toBeInTheDocument();
      expect(screen.queryByText('Application Details')).not.toBeInTheDocument();
    });
  });

  describe('data fetching optimization', () => {
    it('uses component from store when available instead of fetching', () => {
      render(
        <DetailsSidebar {...defaultProps} selectedComponentId={'comp-1' as ComponentId} />
      );

      expect(useComponentDetails).toHaveBeenCalledWith(null);
    });

    it('fetches from API when component not in store', () => {
      vi.mocked(useAppStore).mockImplementation((selector) => {
        const state = {
          components: [],
          capabilities: [],
          capabilityRealizations: [],
        };
        return selector(state as any);
      });
      vi.mocked(useComponentDetails).mockReturnValue({
        component: mockComponent,
        isLoading: false,
        error: null,
      });

      render(
        <DetailsSidebar {...defaultProps} selectedComponentId={'comp-1' as ComponentId} />
      );

      expect(useComponentDetails).toHaveBeenCalledWith('comp-1');
    });
  });
});
