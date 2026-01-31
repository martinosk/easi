import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, within, fireEvent, waitFor } from '@testing-library/react';
import { NavigationTree } from './NavigationTree';
import { createMantineTestWrapper, seedDb } from '../../../test/helpers';
import { useAppStore } from '../../../store/appStore';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import type { View, ComponentId, CapabilityId } from '../../../api/types';
import { toViewId } from '../../../api/types';
import type { AppStore } from '../../../store/appStore';

vi.mock('../../../store/appStore', () => ({
  useAppStore: vi.fn(),
}));

vi.mock('../../canvas/context/CanvasLayoutContext', () => ({
  useCanvasLayoutContext: vi.fn(() => ({
    positions: {},
    isLoading: false,
    error: null,
    updateComponentPosition: vi.fn(),
    updateCapabilityPosition: vi.fn(),
    batchUpdatePositions: vi.fn(),
    getPositionForElement: vi.fn(),
    refetch: vi.fn(),
  })),
}));

vi.mock('../../views/hooks/useCurrentView', () => ({
  useCurrentView: vi.fn(),
}));

const mockComponents = [
  {
    id: 'comp-1' as ComponentId,
    name: 'Payment Service',
    description: 'Handles payments',
    createdAt: '2024-01-01',
    _links: { self: { href: '/api/v1/components/comp-1', method: 'GET' as const } },
  },
  {
    id: 'comp-2' as ComponentId,
    name: 'Order Service',
    description: 'Handles orders',
    createdAt: '2024-01-01',
    _links: { self: { href: '/api/v1/components/comp-2', method: 'GET' as const } },
  },
];

const mockCapabilities = [
  {
    id: 'cap-1' as CapabilityId,
    name: 'Customer Management',
    description: 'Manage customers',
    level: 'L1' as const,
    maturityLevel: 'Product' as const,
    createdAt: '2024-01-01',
    _links: { self: { href: '/api/v1/capabilities/cap-1', method: 'GET' as const } },
  },
  {
    id: 'cap-2' as CapabilityId,
    name: 'Order Processing',
    description: 'Process orders',
    level: 'L2' as const,
    parentId: 'cap-1' as CapabilityId,
    maturityLevel: 'Custom Build' as const,
    createdAt: '2024-01-01',
    _links: { self: { href: '/api/v1/capabilities/cap-2', method: 'GET' as const } },
  },
  {
    id: 'cap-3' as CapabilityId,
    name: 'Shipping',
    description: 'Shipping operations',
    level: 'L1' as const,
    maturityLevel: 'Genesis' as const,
    createdAt: '2024-01-01',
    _links: { self: { href: '/api/v1/capabilities/cap-3', method: 'GET' as const } },
  },
];

const createViewWithColorScheme = (
  colorScheme: string,
  components: Array<{ componentId: string; customColor?: string }>,
  capabilities: Array<{ capabilityId: string; customColor?: string }>
): View => ({
  id: toViewId('view-1'),
  name: 'Architecture View',
  description: 'Main view',
  isDefault: true,
  isPrivate: false,
  colorScheme,
  components: components.map(c => ({
    componentId: c.componentId as ComponentId,
    x: 100,
    y: 200,
    customColor: c.customColor,
  })),
  capabilities: capabilities.map(c => ({
    capabilityId: c.capabilityId as CapabilityId,
    x: 150,
    y: 250,
    customColor: c.customColor,
  })),
  originEntities: [],
  createdAt: '2024-01-01',
  _links: { self: { href: '/api/v1/views/view-1', method: 'GET' as const } },
});

const createMockStore = () => ({
  selectedNodeId: null,
});

describe('NavigationTree - Custom Color Indicators', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    seedDb({
      components: mockComponents,
      capabilities: mockCapabilities,
    });
  });

  const renderNavigationTree = (currentView: View | null) => {
    vi.mocked(useAppStore).mockImplementation((selector: (state: AppStore) => unknown) => selector(createMockStore() as unknown as AppStore));
    vi.mocked(useCurrentView).mockReturnValue({
      currentView,
      currentViewId: currentView?.id ?? null,
      isLoading: false,
      error: null,
    });
    const { Wrapper } = createMantineTestWrapper();
    return render(<NavigationTree />, { wrapper: Wrapper });
  };

  const renderWithRerender = (currentView: View) => {
    const { Wrapper } = createMantineTestWrapper();
    vi.mocked(useAppStore).mockImplementation((selector: (state: AppStore) => unknown) =>
      selector(createMockStore() as unknown as AppStore)
    );
    vi.mocked(useCurrentView).mockReturnValue({
      currentView,
      currentViewId: currentView.id,
      isLoading: false,
      error: null,
    });
    return { ...render(<NavigationTree />, { wrapper: Wrapper }), Wrapper };
  };

  const switchView = (nextView: View) => {
    vi.mocked(useCurrentView).mockReturnValue({
      currentView: nextView,
      currentViewId: nextView.id,
      isLoading: false,
      error: null,
    });
  };

  const waitForText = async (text: string) => {
    await waitFor(() => {
      expect(screen.getByText(text)).toBeInTheDocument();
    });
  };

  const findTreeItem = (text: string, selector: string) =>
    screen.getByText(text).closest(selector) as HTMLElement;

  const expectIndicatorWithColor = (container: HTMLElement, color: string) => {
    const indicator = within(container).getByTestId('custom-color-indicator');
    expect(indicator).toBeInTheDocument();
    expect(indicator).toHaveStyle({ backgroundColor: color });
  };

  const expectNoIndicator = (container: HTMLElement) => {
    expect(within(container).queryByTestId('custom-color-indicator')).not.toBeInTheDocument();
  };

  const renderAndWait = async (view: View | null, waitText: string) => {
    renderNavigationTree(view);
    await waitForText(waitText);
  };

  describe('Component Color Indicators', () => {
    it('should show custom color indicator when colorScheme is "custom" and component has customColor', async () => {
      await renderAndWait(
        createViewWithColorScheme('custom', [{ componentId: 'comp-1', customColor: '#FF5733' }, { componentId: 'comp-2' }], []),
        'Payment Service'
      );
      expectIndicatorWithColor(findTreeItem('Payment Service', '.tree-item'), '#FF5733');
    });

    it.each([
      { scheme: 'maturity', color: '#FF5733' },
      { scheme: 'classic', color: '#00AA00' },
    ])('should NOT show color indicator when colorScheme is "$scheme" even if customColor exists', async ({ scheme, color }) => {
      await renderAndWait(createViewWithColorScheme(scheme, [{ componentId: 'comp-1', customColor: color }], []), 'Payment Service');
      expectNoIndicator(findTreeItem('Payment Service', '.tree-item'));
    });

    it('should NOT show color indicator when colorScheme is "custom" but customColor is null', async () => {
      await renderAndWait(createViewWithColorScheme('custom', [{ componentId: 'comp-1' }], []), 'Payment Service');
      expectNoIndicator(findTreeItem('Payment Service', '.tree-item'));
    });

    it('should show correct color for each component with custom colors', async () => {
      await renderAndWait(
        createViewWithColorScheme('custom', [{ componentId: 'comp-1', customColor: '#FF5733' }, { componentId: 'comp-2', customColor: '#33AAFF' }], []),
        'Payment Service'
      );
      expectIndicatorWithColor(findTreeItem('Payment Service', '.tree-item'), '#FF5733');
      expectIndicatorWithColor(findTreeItem('Order Service', '.tree-item'), '#33AAFF');
    });
  });

  describe('Capability Color Indicators', () => {
    it('should show custom color indicator when colorScheme is "custom" and capability has customColor', async () => {
      await renderAndWait(
        createViewWithColorScheme('custom', [], [{ capabilityId: 'cap-1', customColor: '#AA00FF' }, { capabilityId: 'cap-3' }]),
        'Customer Management'
      );
      expectIndicatorWithColor(findTreeItem('Customer Management', '.capability-tree-item'), '#AA00FF');
    });

    it('should NOT show color indicator when colorScheme is "maturity" even if customColor exists', async () => {
      await renderAndWait(createViewWithColorScheme('maturity', [], [{ capabilityId: 'cap-1', customColor: '#AA00FF' }]), 'Customer Management');
      expectNoIndicator(findTreeItem('Customer Management', '.capability-tree-item'));
    });

    it('should NOT show color indicator when colorScheme is "custom" but customColor is null', async () => {
      await renderAndWait(createViewWithColorScheme('custom', [], [{ capabilityId: 'cap-1' }]), 'Customer Management');
      expectNoIndicator(findTreeItem('Customer Management', '.capability-tree-item'));
    });

    it('should show correct color for each capability with custom colors', async () => {
      await renderAndWait(
        createViewWithColorScheme('custom', [], [{ capabilityId: 'cap-1', customColor: '#FF00AA' }, { capabilityId: 'cap-3', customColor: '#00FFAA' }]),
        'Customer Management'
      );
      expectIndicatorWithColor(findTreeItem('Customer Management', '.capability-tree-item'), '#FF00AA');
      expectIndicatorWithColor(findTreeItem('Shipping', '.capability-tree-item'), '#00FFAA');
    });

    it('should show color indicator for child capabilities with custom colors', async () => {
      await renderAndWait(
        createViewWithColorScheme('custom', [], [{ capabilityId: 'cap-1' }, { capabilityId: 'cap-2', customColor: '#AABBCC' }]),
        'Customer Management'
      );
      fireEvent.click(within(findTreeItem('Customer Management', '.capability-tree-item')).getByRole('button'));
      expectIndicatorWithColor(findTreeItem('Order Processing', '.capability-tree-item'), '#AABBCC');
    });
  });

  describe('Color Scheme Switching', () => {
    it('should show indicators when switching from maturity to custom scheme', async () => {
      const initialView = createViewWithColorScheme('maturity', [{ componentId: 'comp-1', customColor: '#FF5733' }], []);
      const { rerender } = renderWithRerender(initialView);
      await waitForText('Payment Service');

      expectNoIndicator(findTreeItem('Payment Service', '.tree-item'));

      const customView = createViewWithColorScheme('custom', [{ componentId: 'comp-1', customColor: '#FF5733' }], []);
      switchView(customView);
      rerender(<NavigationTree />);

      expectIndicatorWithColor(findTreeItem('Payment Service', '.tree-item'), '#FF5733');
    });

    it('should hide indicators when switching from custom to classic scheme', async () => {
      const initialView = createViewWithColorScheme('custom', [], [{ capabilityId: 'cap-1', customColor: '#AA00FF' }]);
      const { rerender } = renderWithRerender(initialView);
      await waitForText('Customer Management');

      expectIndicatorWithColor(findTreeItem('Customer Management', '.capability-tree-item'), '#AA00FF');

      const classicView = createViewWithColorScheme('classic', [], [{ capabilityId: 'cap-1', customColor: '#AA00FF' }]);
      switchView(classicView);
      rerender(<NavigationTree />);

      expectNoIndicator(findTreeItem('Customer Management', '.capability-tree-item'));
    });
  });

  describe('Edge Cases', () => {
    it('should handle null currentView gracefully', async () => {
      await renderAndWait(null, 'Payment Service');
      expectNoIndicator(findTreeItem('Payment Service', '.tree-item'));
    });

    it('should handle undefined colorScheme gracefully', async () => {
      const currentView = createViewWithColorScheme('maturity', [{ componentId: 'comp-1', customColor: '#FF5733' }], []);
      currentView.colorScheme = undefined;
      await renderAndWait(currentView, 'Payment Service');
      expectNoIndicator(findTreeItem('Payment Service', '.tree-item'));
    });

    it('should handle mixed scenario with some elements having colors and others not', async () => {
      await renderAndWait(
        createViewWithColorScheme(
          'custom',
          [{ componentId: 'comp-1', customColor: '#FF5733' }, { componentId: 'comp-2' }],
          [{ capabilityId: 'cap-1', customColor: '#AA00FF' }, { capabilityId: 'cap-3' }]
        ),
        'Payment Service'
      );
      expectIndicatorWithColor(findTreeItem('Payment Service', '.tree-item'), '#FF5733');
      expectNoIndicator(findTreeItem('Order Service', '.tree-item'));
      expectIndicatorWithColor(findTreeItem('Customer Management', '.capability-tree-item'), '#AA00FF');
      expectNoIndicator(findTreeItem('Shipping', '.capability-tree-item'));
    });
  });
});
