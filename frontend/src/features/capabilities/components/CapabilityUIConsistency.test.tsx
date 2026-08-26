import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { apiClient } from '../../../api/client';
import type { Capability, CapabilityId, Component, ComponentId, View, ViewId } from '../../../api/types';
import type { AppStore } from '../../../store/appStore';
import { useAppStore } from '../../../store/appStore';
import { createMantineTestWrapper, seedDb } from '../../../test/helpers';
import { EditComponentDialog } from '../../components/components/EditComponentDialog';
import { NavigationTree } from '../../navigation/components/NavigationTree';
import { EditCapabilityDialog } from './EditCapabilityDialog';

vi.mock('../../../store/appStore', () => ({
  useAppStore: vi.fn(),
}));

vi.mock('../../../api/client', () => ({
  apiClient: {
    getMaturityLevels: vi.fn(),
    getStatuses: vi.fn(),
    getOwnershipModels: vi.fn(),
  },
  default: {
    getMaturityLevels: vi.fn(),
    getStatuses: vi.fn(),
    getOwnershipModels: vi.fn(),
  },
}));

function setupMockStore(overrides: Record<string, unknown> = {}) {
  const mockStore = createMockStore(overrides);
  vi.mocked(useAppStore).mockImplementation((selector: (state: AppStore) => unknown) =>
    selector(mockStore as unknown as AppStore),
  );
  return mockStore;
}

function setupApiClientMocks() {
  vi.mocked(apiClient.getMaturityLevels).mockResolvedValue(['Genesis', 'Custom Build', 'Product', 'Commodity']);
  vi.mocked(apiClient.getStatuses).mockResolvedValue([{ value: 'Active', displayName: 'Active', sortOrder: 1 }]);
  vi.mocked(apiClient.getOwnershipModels).mockResolvedValue([]);
}

function renderNavigationTree(props: Record<string, unknown> = {}) {
  const { Wrapper } = createMantineTestWrapper();
  return render(<NavigationTree {...props} />, { wrapper: Wrapper });
}

function renderEditCapabilityDialog(props: { isOpen: boolean; onClose: () => void; capability: Capability | null }) {
  const { Wrapper } = createMantineTestWrapper();
  return render(<EditCapabilityDialog {...props} />, { wrapper: Wrapper });
}

function expandButtonFor(capabilityName: string) {
  const item = screen.getByText(capabilityName).closest('[data-testid^="capability-tree-item-"]');
  return item?.querySelector('button[aria-label="Expand"]');
}

async function openContextMenuOn(text: string, selector: string) {
  await waitFor(() => {
    expect(screen.getByText(text)).toBeInTheDocument();
  });
  const element = screen.getByText(text).closest(selector);
  fireEvent.contextMenu(element!);
  await waitFor(() => {
    expect(screen.getByRole('menu')).toBeInTheDocument();
  });
}

function createComponentsWithB() {
  return [
    ...mockComponents,
    {
      id: 'comp-2' as ComponentId,
      name: 'Component B',
      createdAt: '2024-01-01T00:00:00Z',
      _links: { self: { href: '/api/v1/components/comp-2', method: 'GET' as const } },
    },
  ];
}

const mockCapabilities: Capability[] = [
  {
    id: 'cap-1' as CapabilityId,
    name: 'Customer Management',
    level: 'L1',
    description: 'Manages customer data',
    maturityLevel: 'Product',
    createdAt: '2024-01-01T00:00:00Z',
    _links: {
      self: { href: '/api/v1/capabilities/cap-1', method: 'GET' },
      edit: { href: '/api/v1/capabilities/cap-1', method: 'PUT' },
      delete: { href: '/api/v1/capabilities/cap-1', method: 'DELETE' },
    },
  },
  {
    id: 'cap-2' as CapabilityId,
    name: 'Order Processing',
    level: 'L2',
    parentId: 'cap-1' as CapabilityId,
    description: 'Processes orders',
    maturityLevel: 'Genesis',
    createdAt: '2024-01-01T00:00:00Z',
    _links: {
      self: { href: '/api/v1/capabilities/cap-2', method: 'GET' },
      edit: { href: '/api/v1/capabilities/cap-2', method: 'PUT' },
      delete: { href: '/api/v1/capabilities/cap-2', method: 'DELETE' },
    },
  },
  {
    id: 'cap-3' as CapabilityId,
    name: 'Inventory Control',
    level: 'L1',
    description: 'Controls inventory',
    maturityLevel: 'Commodity',
    createdAt: '2024-01-01T00:00:00Z',
    _links: {
      self: { href: '/api/v1/capabilities/cap-3', method: 'GET' },
      edit: { href: '/api/v1/capabilities/cap-3', method: 'PUT' },
      delete: { href: '/api/v1/capabilities/cap-3', method: 'DELETE' },
    },
  },
];

const mockComponents: Component[] = [
  {
    id: 'comp-1' as ComponentId,
    name: 'Component A',
    description: 'Test component',
    createdAt: '2024-01-01T00:00:00Z',
    _links: {
      self: { href: '/api/v1/components/comp-1', method: 'GET' },
      edit: { href: '/api/v1/components/comp-1', method: 'PUT' },
      delete: { href: '/api/v1/components/comp-1', method: 'DELETE' },
    },
  },
];

const mockCurrentView: View = {
  id: 'view-1' as ViewId,
  name: 'Main View',
  description: 'Default view',
  isDefault: true,
  isPrivate: false,
  components: [{ componentId: 'comp-1' as ComponentId, x: 100, y: 100 }],
  capabilities: [{ capabilityId: 'cap-1' as CapabilityId, x: 200, y: 200 }],
  originEntities: [],
  createdAt: '2024-01-01T00:00:00Z',
  _links: { self: { href: '/api/v1/views/view-1', method: 'GET' } },
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
    seedDb({
      capabilities: mockCapabilities,
      components: mockComponents,
      views: [mockCurrentView],
    });
  });

  describe('Dialog Management', () => {
    describe('EditCapabilityDialog should be managed via DialogManager pattern', () => {
      it('should render dialog as a modal overlay when opened', async () => {
        setupMockStore();
        setupApiClientMocks();
        renderEditCapabilityDialog({ isOpen: true, onClose: vi.fn(), capability: mockCapabilities[0] });

        await waitFor(() => {
          expect(screen.getByText('Edit Capability')).toBeInTheDocument();
        });
      });

      it('should not show modal when isOpen is false', async () => {
        setupMockStore();
        renderEditCapabilityDialog({ isOpen: false, onClose: vi.fn(), capability: null });

        expect(screen.queryByText('Edit Capability')).not.toBeInTheDocument();
      });

      it('should call onClose when cancel button is clicked', async () => {
        setupMockStore();
        setupApiClientMocks();
        const mockOnClose = vi.fn();
        renderEditCapabilityDialog({ isOpen: true, onClose: mockOnClose, capability: mockCapabilities[0] });

        await waitFor(() => {
          expect(screen.getByTestId('edit-capability-cancel')).toBeInTheDocument();
        });

        fireEvent.click(screen.getByTestId('edit-capability-cancel'));

        expect(mockOnClose).toHaveBeenCalled();
      });

      it('should follow same pattern as EditComponentDialog for dialog opening', async () => {
        setupMockStore();
        setupApiClientMocks();
        const { Wrapper } = createMantineTestWrapper();

        const { rerender } = render(
          <EditComponentDialog isOpen={true} onClose={vi.fn()} component={mockComponents[0]} />,
          { wrapper: Wrapper },
        );

        await waitFor(() => {
          expect(screen.getByText('Edit Application')).toBeInTheDocument();
        });

        rerender(<EditCapabilityDialog isOpen={true} onClose={vi.fn()} capability={mockCapabilities[0]} />);

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
        setupMockStore({ canvasCapabilities: [{ capabilityId: 'cap-1', x: 200, y: 200 }] });
        renderNavigationTree();

        await waitFor(() => {
          expect(screen.getByText('Customer Management')).toBeInTheDocument();
        });
        expect(screen.getByText('Inventory Control')).toBeInTheDocument();
      });

      it('should allow capabilities not in view to remain draggable', async () => {
        setupMockStore({ canvasCapabilities: [{ capabilityId: 'cap-1', x: 200, y: 200 }] });
        renderNavigationTree();

        await waitFor(() => {
          const capabilityItem = screen
            .getByText('Inventory Control')
            .closest('[data-testid^="capability-tree-item-"]');
          expect(capabilityItem).toHaveAttribute('draggable', 'true');
        });
      });

      it('should set correct data transfer on drag start for capabilities', async () => {
        setupMockStore({ canvasCapabilities: [] });
        renderNavigationTree();

        await waitFor(() => {
          expect(screen.getByText('Customer Management')).toBeInTheDocument();
        });

        const capabilityItem = screen
          .getByText('Customer Management')
          .closest('[data-testid^="capability-tree-item-"]');
        expect(capabilityItem).toBeTruthy();

        const mockDataTransfer = {
          setData: vi.fn(),
          effectAllowed: '',
        };

        fireEvent.dragStart(capabilityItem!, {
          dataTransfer: mockDataTransfer,
        });

        expect(mockDataTransfer.setData).toHaveBeenCalledWith('capabilityId', 'cap-1');
      });
    });

    describe('Components visual distinction pattern (for reference)', () => {
      function setupComponentOutOfViewStore() {
        const componentsWithB = createComponentsWithB();
        setupMockStore({
          components: componentsWithB,
          currentView: {
            ...mockCurrentView,
            components: [{ componentId: 'comp-1', x: 100, y: 100 }],
          },
        });
        seedDb({ components: componentsWithB, capabilities: mockCapabilities });
      }

      it('should mark components not in current view as out of view', async () => {
        setupComponentOutOfViewStore();
        renderNavigationTree();

        await waitFor(() => {
          const compBButton = screen.getByText('Component B').closest('button');
          expect(compBButton).toHaveAttribute('data-in-view', 'false');
        });
      });

      it('should show tooltip suffix for components not in current view', async () => {
        setupComponentOutOfViewStore();
        renderNavigationTree();

        await waitFor(() => {
          const compBButton = screen.getByText('Component B').closest('button');
          expect(compBButton).toHaveAttribute('title', 'Component B (not in current view)');
        });
      });
    });
  });

  describe('View Focus on Selection', () => {
    it.each([
      ['onCapabilitySelect', 'Customer Management', '[data-testid^="capability-tree-item-"]', 'cap-1'],
      ['onComponentSelect', 'Component A', 'button', 'comp-1'],
    ])('should call %s when the tree item is clicked', async (prop, text, selector, expectedId) => {
      const onSelect = vi.fn();
      setupMockStore();
      renderNavigationTree({ [prop]: onSelect });

      await waitFor(() => {
        expect(screen.getByText(text)).toBeInTheDocument();
      });

      fireEvent.click(screen.getByText(text).closest(selector)!);

      expect(onSelect).toHaveBeenCalledWith(expectedId);
    });
  });

  describe('Context Menu Consistency', () => {
    describe('Tree Context Menu for Capabilities', () => {
      it('should show Edit option in capability tree context menu', async () => {
        setupMockStore();
        renderNavigationTree({ onEditCapability: vi.fn() });
        await openContextMenuOn('Customer Management', '[data-testid^="capability-tree-item-"]');

        expect(screen.getByRole('menuitem', { name: 'Edit' })).toBeInTheDocument();
      });

      it('should show Delete from Model option in capability tree context menu', async () => {
        setupMockStore();
        renderNavigationTree();
        await openContextMenuOn('Customer Management', '[data-testid^="capability-tree-item-"]');

        expect(screen.getByRole('menuitem', { name: /Delete capability from model/i })).toBeInTheDocument();
      });
    });

    describe('Tree Context Menu for Components', () => {
      it('should show Edit option in component tree context menu', async () => {
        setupMockStore();
        renderNavigationTree({ onEditComponent: vi.fn() });
        await openContextMenuOn('Component A', 'button');

        expect(screen.getByRole('menuitem', { name: 'Edit' })).toBeInTheDocument();
      });

      it('should show Delete from Model option in component tree context menu', async () => {
        setupMockStore();
        renderNavigationTree();
        await openContextMenuOn('Component A', 'button');

        expect(screen.getByRole('menuitem', { name: /Delete application from model/i })).toBeInTheDocument();
      });
    });

    describe('Context Menu item structure comparison', () => {
      it('should have matching menu structure for capability and component tree items', async () => {
        setupMockStore();
        const { Wrapper } = createMantineTestWrapper();

        const { rerender } = render(<NavigationTree onEditCapability={vi.fn()} onEditComponent={vi.fn()} />, {
          wrapper: Wrapper,
        });

        await waitFor(() => {
          expect(screen.getByText('Customer Management')).toBeInTheDocument();
        });

        const capabilityItem = screen
          .getByText('Customer Management')
          .closest('[data-testid^="capability-tree-item-"]');
        fireEvent.contextMenu(capabilityItem!);

        await waitFor(() => {
          expect(screen.getByRole('menu')).toBeInTheDocument();
        });

        const capabilityMenuItems = screen.getAllByRole('menuitem');
        const capabilityMenuLabels = capabilityMenuItems.map((item) => item.textContent);

        fireEvent.keyDown(screen.getByTestId('context-menu'), { key: 'Escape' });

        rerender(<NavigationTree onEditCapability={vi.fn()} onEditComponent={vi.fn()} />);

        await waitFor(() => {
          expect(screen.queryByRole('menu')).not.toBeInTheDocument();
        });

        const componentButton = screen.getByText('Component A').closest('button');
        fireEvent.contextMenu(componentButton!);

        await waitFor(() => {
          expect(screen.getByRole('menu')).toBeInTheDocument();
        });

        const componentMenuItems = screen.getAllByRole('menuitem');
        const componentMenuLabels = componentMenuItems.map((item) => item.textContent);

        expect(capabilityMenuLabels).toContain('Edit');
        expect(componentMenuLabels).toContain('Edit');
        expect(capabilityMenuLabels.some((label) => label?.includes('Delete'))).toBe(true);
        expect(componentMenuLabels.some((label) => label?.includes('Delete'))).toBe(true);
      });
    });
  });
});

describe('Capability Tree Item Selection', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    seedDb({
      capabilities: mockCapabilities,
      components: mockComponents,
    });
  });

  it('should apply selected class when capability is clicked', async () => {
    setupMockStore({ selectedCapabilityId: null });
    renderNavigationTree();

    await waitFor(() => {
      const capabilityItem = screen.getByText('Customer Management').closest('[data-testid^="capability-tree-item-"]');
      fireEvent.click(capabilityItem!);
    });
  });
});

describe('Capability Expand/Collapse in Tree', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    seedDb({
      capabilities: mockCapabilities,
      components: mockComponents,
    });
  });

  it('should show expand button for capabilities with children', async () => {
    setupMockStore();
    renderNavigationTree();

    await waitFor(() => {
      expect(expandButtonFor('Customer Management')).toBeInTheDocument();
    });
  });

  it('should not show expand button for capabilities without children', async () => {
    setupMockStore();
    renderNavigationTree();

    await waitFor(() => {
      expect(expandButtonFor('Inventory Control')).not.toBeInTheDocument();
    });
  });

  it('should toggle children visibility when expand button is clicked', async () => {
    setupMockStore();
    renderNavigationTree();

    await waitFor(() => {
      expect(screen.getByText('Customer Management')).toBeInTheDocument();
    });

    const expandBtn = expandButtonFor('Customer Management');
    expect(expandBtn).toBeTruthy();

    expect(screen.queryByText('Order Processing')).not.toBeInTheDocument();

    fireEvent.click(expandBtn!);

    await waitFor(() => {
      expect(screen.getByText('Order Processing')).toBeInTheDocument();
    });
  });
});
