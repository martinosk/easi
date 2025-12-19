import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import React from 'react';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';

vi.mock('../../../store/appStore', () => ({
  useAppStore: vi.fn(),
}));

vi.mock('../../../api/client', () => ({
  apiClient: {
    getMaturityLevels: vi.fn(),
    getStatuses: vi.fn(),
    getOwnershipModels: vi.fn(),
    getStrategyPillars: vi.fn(),
  },
  default: {
    getMaturityLevels: vi.fn(),
    getStatuses: vi.fn(),
    getOwnershipModels: vi.fn(),
    getStrategyPillars: vi.fn(),
  },
}));

import { useAppStore } from '../../../store/appStore';
import type { Capability, View, Component } from '../../../api/types';

const mockCapabilities: Capability[] = [
  {
    id: 'cap-1',
    name: 'Customer Management',
    level: 'L1',
    description: 'Manages customer data',
    maturityLevel: 'Product',
    createdAt: '2024-01-01T00:00:00Z',
    _links: { self: { href: '/api/v1/capabilities/cap-1' } },
  },
  {
    id: 'cap-2',
    name: 'Order Processing',
    level: 'L2',
    parentId: 'cap-1',
    description: 'Processes orders',
    maturityLevel: 'Genesis',
    createdAt: '2024-01-01T00:00:00Z',
    _links: { self: { href: '/api/v1/capabilities/cap-2' } },
  },
  {
    id: 'cap-3',
    name: 'Inventory Control',
    level: 'L1',
    description: 'Controls inventory',
    maturityLevel: 'Commodity',
    createdAt: '2024-01-01T00:00:00Z',
    _links: { self: { href: '/api/v1/capabilities/cap-3' } },
  },
];

const mockComponents: Component[] = [
  {
    id: 'comp-1',
    name: 'Component A',
    description: 'Test component',
    createdAt: '2024-01-01T00:00:00Z',
    _links: { self: { href: '/api/v1/components/comp-1' } },
  },
];

const mockCurrentView: View = {
  id: 'view-1',
  name: 'Main View',
  description: 'Default view',
  isDefault: true,
  components: [{ componentId: 'comp-1', x: 100, y: 100 }],
  capabilities: [{ capabilityId: 'cap-1', x: 200, y: 200 }],
  createdAt: '2024-01-01T00:00:00Z',
  _links: { self: { href: '/api/v1/views/view-1' } },
};

const createMockStore = (overrides: Record<string, unknown> = {}) => ({
  capabilities: mockCapabilities,
  components: mockComponents,
  currentView: mockCurrentView,
  views: [mockCurrentView],
  relations: [],
  selectedNodeId: null,
  selectedCapabilityId: null,
  canvasCapabilities: mockCurrentView.capabilities,
  loadCapabilities: vi.fn(),
  loadViews: vi.fn(),
  updateComponent: vi.fn(),
  deleteComponent: vi.fn(),
  updateCapability: vi.fn(),
  updateCapabilityMetadata: vi.fn(),
  addCapabilityExpert: vi.fn(),
  addCapabilityTag: vi.fn(),
  selectCapability: vi.fn(),
  selectNode: vi.fn(),
  clearSelection: vi.fn(),
  ...overrides,
});

describe('Capability UI Consistency', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Dialog Management', () => {
    describe('EditCapabilityDialog should be managed via DialogManager pattern', () => {
      it('should render dialog as a modal overlay when opened', async () => {
        const mockStore = createMockStore();
        vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
          selector(mockStore)
        );

        const { apiClient } = await import('../../../api/client');
        vi.mocked(apiClient.getMaturityLevels).mockResolvedValue(['Genesis', 'Custom Build', 'Product', 'Commodity']);
        vi.mocked(apiClient.getStatuses).mockResolvedValue([
          { value: 'Active', displayName: 'Active', sortOrder: 1 },
        ]);
        vi.mocked(apiClient.getOwnershipModels).mockResolvedValue([]);
        vi.mocked(apiClient.getStrategyPillars).mockResolvedValue([]);

        const { EditCapabilityDialog } = await import('./EditCapabilityDialog');
        const capability = mockCapabilities[0];

        render(<EditCapabilityDialog isOpen={true} onClose={vi.fn()} capability={capability} />, { wrapper: MantineTestWrapper });

        await waitFor(() => {
          expect(screen.getByText('Edit Capability')).toBeInTheDocument();
        });
      });

      it('should not show modal when isOpen is false', async () => {
        const mockStore = createMockStore();
        vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
          selector(mockStore)
        );

        const { EditCapabilityDialog } = await import('./EditCapabilityDialog');

        render(<EditCapabilityDialog isOpen={false} onClose={vi.fn()} capability={null} />, { wrapper: MantineTestWrapper });

        expect(screen.queryByText('Edit Capability')).not.toBeInTheDocument();
      });

      it('should call onClose when cancel button is clicked', async () => {
        const mockStore = createMockStore();
        vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
          selector(mockStore)
        );

        const { apiClient } = await import('../../../api/client');
        vi.mocked(apiClient.getMaturityLevels).mockResolvedValue(['Genesis', 'Custom Build', 'Product', 'Commodity']);
        vi.mocked(apiClient.getStatuses).mockResolvedValue([
          { value: 'Active', displayName: 'Active', sortOrder: 1 },
        ]);
        vi.mocked(apiClient.getOwnershipModels).mockResolvedValue([]);
        vi.mocked(apiClient.getStrategyPillars).mockResolvedValue([]);

        const { EditCapabilityDialog } = await import('./EditCapabilityDialog');
        const mockOnClose = vi.fn();
        const capability = mockCapabilities[0];

        render(<EditCapabilityDialog isOpen={true} onClose={mockOnClose} capability={capability} />, { wrapper: MantineTestWrapper });

        await waitFor(() => {
          expect(screen.getByTestId('edit-capability-cancel')).toBeInTheDocument();
        });

        fireEvent.click(screen.getByTestId('edit-capability-cancel'));

        expect(mockOnClose).toHaveBeenCalled();
      });

      it('should follow same pattern as EditComponentDialog for dialog opening', async () => {
        const mockStore = createMockStore();
        vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
          selector(mockStore)
        );

        const { apiClient } = await import('../../../api/client');
        vi.mocked(apiClient.getMaturityLevels).mockResolvedValue(['Genesis', 'Custom Build', 'Product', 'Commodity']);
        vi.mocked(apiClient.getStatuses).mockResolvedValue([
          { value: 'Active', displayName: 'Active', sortOrder: 1 },
        ]);
        vi.mocked(apiClient.getOwnershipModels).mockResolvedValue([]);
        vi.mocked(apiClient.getStrategyPillars).mockResolvedValue([]);

        const { EditComponentDialog } = await import('../../components/components/EditComponentDialog');
        const { EditCapabilityDialog } = await import('./EditCapabilityDialog');
        const component = mockComponents[0];
        const capability = mockCapabilities[0];

        const { rerender } = render(
          <EditComponentDialog isOpen={true} onClose={vi.fn()} component={component} />,
          { wrapper: MantineTestWrapper }
        );

        await waitFor(() => {
          expect(screen.getByText('Edit Application')).toBeInTheDocument();
        });

        rerender(
          <EditCapabilityDialog isOpen={true} onClose={vi.fn()} capability={capability} />
        );

        await waitFor(() => {
          expect(screen.getByText('Edit Capability')).toBeInTheDocument();
          expect(screen.queryByText('Edit Application')).not.toBeInTheDocument();
        });
      });
    });
  });

  describe('Treeview Visibility', () => {
    describe('Capabilities should show visual distinction when not in view', () => {
      it('should render all capabilities in tree regardless of view presence', async () => {
        const mockStore = createMockStore({
          canvasCapabilities: [{ capabilityId: 'cap-1', x: 200, y: 200 }],
        });
        vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
          selector(mockStore)
        );

        const { NavigationTree } = await import('../../navigation/components/NavigationTree');
        render(<NavigationTree />);

        await waitFor(() => {
          expect(screen.getByText('Customer Management')).toBeInTheDocument();
        });
        expect(screen.getByText('Inventory Control')).toBeInTheDocument();
      });

      it('should allow capabilities not in view to remain draggable', async () => {
        const mockStore = createMockStore({
          canvasCapabilities: [{ capabilityId: 'cap-1', x: 200, y: 200 }],
        });
        vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
          selector(mockStore)
        );

        const { NavigationTree } = await import('../../navigation/components/NavigationTree');
        render(<NavigationTree />);

        await waitFor(() => {
          const capabilityItem = screen.getByText('Inventory Control').closest('.capability-tree-item');
          expect(capabilityItem).toHaveAttribute('draggable', 'true');
        });
      });

      it('should set correct data transfer on drag start for capabilities', async () => {
        const mockStore = createMockStore({
          canvasCapabilities: [],
        });
        vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
          selector(mockStore)
        );

        const { NavigationTree } = await import('../../navigation/components/NavigationTree');
        render(<NavigationTree />);

        await waitFor(() => {
          expect(screen.getByText('Customer Management')).toBeInTheDocument();
        });

        const capabilityItem = screen.getByText('Customer Management').closest('.capability-tree-item');
        expect(capabilityItem).toBeTruthy();

        const mockDataTransfer = {
          setData: vi.fn(),
          effectAllowed: '',
        };

        fireEvent.dragStart(capabilityItem!, {
          dataTransfer: mockDataTransfer,
        });

        expect(mockDataTransfer.setData).toHaveBeenCalledWith('capabilityId', 'cap-1');
        expect(mockDataTransfer.effectAllowed).toBe('copy');
      });
    });

    describe('Components visual distinction pattern (for reference)', () => {
      it('should apply not-in-view class to components not in current view', async () => {
        const mockStoreWithComponentOutOfView = createMockStore({
          components: [
            ...mockComponents,
            {
              id: 'comp-2',
              name: 'Component B',
              createdAt: '2024-01-01T00:00:00Z',
              _links: { self: { href: '/api/v1/components/comp-2' } },
            },
          ],
          currentView: {
            ...mockCurrentView,
            components: [{ componentId: 'comp-1', x: 100, y: 100 }],
          },
        });
        vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStoreWithComponentOutOfView) => unknown) =>
          selector(mockStoreWithComponentOutOfView)
        );

        const { NavigationTree } = await import('../../navigation/components/NavigationTree');
        render(<NavigationTree />);

        await waitFor(() => {
          const compBButton = screen.getByText('Component B').closest('button');
          expect(compBButton).toHaveClass('not-in-view');
        });
      });

      it('should show tooltip suffix for components not in current view', async () => {
        const mockStoreWithComponentOutOfView = createMockStore({
          components: [
            ...mockComponents,
            {
              id: 'comp-2',
              name: 'Component B',
              createdAt: '2024-01-01T00:00:00Z',
              _links: { self: { href: '/api/v1/components/comp-2' } },
            },
          ],
          currentView: {
            ...mockCurrentView,
            components: [{ componentId: 'comp-1', x: 100, y: 100 }],
          },
        });
        vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStoreWithComponentOutOfView) => unknown) =>
          selector(mockStoreWithComponentOutOfView)
        );

        const { NavigationTree } = await import('../../navigation/components/NavigationTree');
        render(<NavigationTree />);

        await waitFor(() => {
          const compBButton = screen.getByText('Component B').closest('button');
          expect(compBButton).toHaveAttribute('title', 'Component B (not in current view)');
        });
      });
    });
  });

  describe('View Focus on Selection', () => {
    it('should call onCapabilitySelect when capability is clicked in tree', async () => {
      const mockOnCapabilitySelect = vi.fn();
      const mockStore = createMockStore();
      vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
        selector(mockStore)
      );

      const { NavigationTree } = await import('../../navigation/components/NavigationTree');
      render(<NavigationTree onCapabilitySelect={mockOnCapabilitySelect} />);

      await waitFor(() => {
        expect(screen.getByText('Customer Management')).toBeInTheDocument();
      });

      const capabilityItem = screen.getByText('Customer Management').closest('.capability-tree-item');
      fireEvent.click(capabilityItem!);

      expect(mockOnCapabilitySelect).toHaveBeenCalledWith('cap-1');
    });

    it('should call onComponentSelect when component is clicked in tree', async () => {
      const mockOnComponentSelect = vi.fn();
      const mockStore = createMockStore();
      vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
        selector(mockStore)
      );

      const { NavigationTree } = await import('../../navigation/components/NavigationTree');
      render(<NavigationTree onComponentSelect={mockOnComponentSelect} />);

      await waitFor(() => {
        expect(screen.getByText('Component A')).toBeInTheDocument();
      });

      const componentButton = screen.getByText('Component A').closest('button');
      fireEvent.click(componentButton!);

      expect(mockOnComponentSelect).toHaveBeenCalledWith('comp-1');
    });
  });

  describe('Context Menu Consistency', () => {
    describe('Tree Context Menu for Capabilities', () => {
      it('should show Edit option in capability tree context menu', async () => {
        const mockStore = createMockStore();
        vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
          selector(mockStore)
        );

        const { NavigationTree } = await import('../../navigation/components/NavigationTree');
        render(<NavigationTree />);

        await waitFor(() => {
          expect(screen.getByText('Customer Management')).toBeInTheDocument();
        });

        const capabilityItem = screen.getByText('Customer Management').closest('.capability-tree-item');
        fireEvent.contextMenu(capabilityItem!);

        await waitFor(() => {
          expect(screen.getByRole('menuitem', { name: 'Edit' })).toBeInTheDocument();
        });
      });

      it('should show Delete from Model option in capability tree context menu', async () => {
        const mockStore = createMockStore();
        vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
          selector(mockStore)
        );

        const { NavigationTree } = await import('../../navigation/components/NavigationTree');
        render(<NavigationTree />);

        await waitFor(() => {
          expect(screen.getByText('Customer Management')).toBeInTheDocument();
        });

        const capabilityItem = screen.getByText('Customer Management').closest('.capability-tree-item');
        fireEvent.contextMenu(capabilityItem!);

        await waitFor(() => {
          expect(screen.getByRole('menuitem', { name: /Delete capability from model/i })).toBeInTheDocument();
        });
      });
    });

    describe('Tree Context Menu for Components', () => {
      it('should show Edit option in component tree context menu', async () => {
        const mockStore = createMockStore();
        vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
          selector(mockStore)
        );

        const { NavigationTree } = await import('../../navigation/components/NavigationTree');
        render(<NavigationTree />);

        await waitFor(() => {
          expect(screen.getByText('Component A')).toBeInTheDocument();
        });

        const componentButton = screen.getByText('Component A').closest('button');
        fireEvent.contextMenu(componentButton!);

        await waitFor(() => {
          expect(screen.getByRole('menuitem', { name: 'Edit' })).toBeInTheDocument();
        });
      });

      it('should show Delete from Model option in component tree context menu', async () => {
        const mockStore = createMockStore();
        vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
          selector(mockStore)
        );

        const { NavigationTree } = await import('../../navigation/components/NavigationTree');
        render(<NavigationTree />);

        await waitFor(() => {
          expect(screen.getByText('Component A')).toBeInTheDocument();
        });

        const componentButton = screen.getByText('Component A').closest('button');
        fireEvent.contextMenu(componentButton!);

        await waitFor(() => {
          expect(screen.getByRole('menuitem', { name: /Delete application from model/i })).toBeInTheDocument();
        });
      });
    });

    describe('Context Menu item structure comparison', () => {
      it('should have matching menu structure for capability and component tree items', async () => {
        const mockStore = createMockStore();
        vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
          selector(mockStore)
        );

        const { NavigationTree } = await import('../../navigation/components/NavigationTree');
        const { rerender } = render(<NavigationTree />);

        await waitFor(() => {
          expect(screen.getByText('Customer Management')).toBeInTheDocument();
        });

        const capabilityItem = screen.getByText('Customer Management').closest('.capability-tree-item');
        fireEvent.contextMenu(capabilityItem!);

        await waitFor(() => {
          expect(screen.getByRole('menu')).toBeInTheDocument();
        });

        const capabilityMenuItems = screen.getAllByRole('menuitem');
        const capabilityMenuLabels = capabilityMenuItems.map(item => item.textContent);

        fireEvent.keyDown(document, { key: 'Escape' });

        rerender(<NavigationTree />);

        await waitFor(() => {
          expect(screen.queryByRole('menu')).not.toBeInTheDocument();
        });

        const componentButton = screen.getByText('Component A').closest('button');
        fireEvent.contextMenu(componentButton!);

        await waitFor(() => {
          expect(screen.getByRole('menu')).toBeInTheDocument();
        });

        const componentMenuItems = screen.getAllByRole('menuitem');
        const componentMenuLabels = componentMenuItems.map(item => item.textContent);

        expect(capabilityMenuLabels).toContain('Edit');
        expect(componentMenuLabels).toContain('Edit');
        expect(capabilityMenuLabels.some(label => label?.includes('Delete'))).toBe(true);
        expect(componentMenuLabels.some(label => label?.includes('Delete'))).toBe(true);
      });
    });
  });
});

describe('Capability Tree Item Selection', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should apply selected class when capability is selected', async () => {
    const mockStore = createMockStore({
      selectedCapabilityId: null,
    });
    vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
      selector(mockStore)
    );

    const { NavigationTree } = await import('../../navigation/components/NavigationTree');
    render(<NavigationTree />);

    await waitFor(() => {
      const capabilityItem = screen.getByText('Customer Management').closest('.capability-tree-item');
      fireEvent.click(capabilityItem!);
    });
  });
});

describe('Capability Expand/Collapse in Tree', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should show expand button for capabilities with children', async () => {
    const mockStore = createMockStore();
    vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
      selector(mockStore)
    );

    const { NavigationTree } = await import('../../navigation/components/NavigationTree');
    render(<NavigationTree />);

    await waitFor(() => {
      const customerManagementItem = screen.getByText('Customer Management').closest('.capability-tree-item');
      expect(customerManagementItem?.querySelector('.capability-expand-btn')).toBeInTheDocument();
    });
  });

  it('should not show expand button for capabilities without children', async () => {
    const mockStore = createMockStore();
    vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
      selector(mockStore)
    );

    const { NavigationTree } = await import('../../navigation/components/NavigationTree');
    render(<NavigationTree />);

    await waitFor(() => {
      const inventoryControlItem = screen.getByText('Inventory Control').closest('.capability-tree-item');
      expect(inventoryControlItem?.querySelector('.capability-expand-btn')).not.toBeInTheDocument();
    });
  });

  it('should toggle children visibility when expand button is clicked', async () => {
    const mockStore = createMockStore();
    vi.mocked(useAppStore).mockImplementation((selector: (state: typeof mockStore) => unknown) =>
      selector(mockStore)
    );

    const { NavigationTree } = await import('../../navigation/components/NavigationTree');
    render(<NavigationTree />);

    await waitFor(() => {
      expect(screen.getByText('Customer Management')).toBeInTheDocument();
    });

    const customerManagementItem = screen.getByText('Customer Management').closest('.capability-tree-item');
    const expandBtn = customerManagementItem?.querySelector('.capability-expand-btn');
    expect(expandBtn).toBeTruthy();

    expect(screen.queryByText('Order Processing')).not.toBeInTheDocument();

    fireEvent.click(expandBtn!);

    await waitFor(() => {
      expect(screen.getByText('Order Processing')).toBeInTheDocument();
    });
  });
});
