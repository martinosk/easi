import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, within, fireEvent, waitFor } from '@testing-library/react';
import { NavigationTree } from './NavigationTree';
import { createMantineTestWrapper, seedDb } from '../../../test/helpers';
import { useAppStore } from '../../../store/appStore';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import type { View, ComponentId, CapabilityId } from '../../../api/types';

vi.mock('../../../store/appStore', () => ({
  useAppStore: vi.fn(),
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
  id: 'view-1',
  name: 'Architecture View',
  description: 'Main view',
  isDefault: true,
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
    vi.mocked(useAppStore).mockImplementation((selector: (state: unknown) => unknown) => selector(createMockStore()));
    vi.mocked(useCurrentView).mockReturnValue({
      currentView,
      currentViewId: currentView?.id ?? null,
      isLoading: false,
      error: null,
    });
    const { Wrapper } = createMantineTestWrapper();
    return render(<NavigationTree />, { wrapper: Wrapper });
  };

  describe('Component Color Indicators', () => {
    it('should show custom color indicator next to component when colorScheme is "custom" and component has customColor', async () => {
      const currentView = createViewWithColorScheme(
        'custom',
        [
          { componentId: 'comp-1', customColor: '#FF5733' },
          { componentId: 'comp-2' },
        ],
        []
      );

      renderNavigationTree(currentView);

      await waitFor(() => {
        expect(screen.getByText('Payment Service')).toBeInTheDocument();
      });

      const paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      const colorIndicator = within(paymentServiceItem!).getByTestId('custom-color-indicator');
      expect(colorIndicator).toBeInTheDocument();
      expect(colorIndicator).toHaveStyle({ backgroundColor: '#FF5733' });
    });

    it('should NOT show color indicator next to component when colorScheme is "maturity" even if customColor exists', async () => {
      const currentView = createViewWithColorScheme(
        'maturity',
        [{ componentId: 'comp-1', customColor: '#FF5733' }],
        []
      );

      renderNavigationTree(currentView);

      await waitFor(() => {
        expect(screen.getByText('Payment Service')).toBeInTheDocument();
      });

      const paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      const colorIndicator = within(paymentServiceItem!).queryByTestId('custom-color-indicator');
      expect(colorIndicator).not.toBeInTheDocument();
    });

    it('should NOT show color indicator next to component when colorScheme is "classic" even if customColor exists', async () => {
      const currentView = createViewWithColorScheme(
        'classic',
        [{ componentId: 'comp-1', customColor: '#00AA00' }],
        []
      );

      renderNavigationTree(currentView);

      await waitFor(() => {
        expect(screen.getByText('Payment Service')).toBeInTheDocument();
      });

      const paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      const colorIndicator = within(paymentServiceItem!).queryByTestId('custom-color-indicator');
      expect(colorIndicator).not.toBeInTheDocument();
    });

    it('should NOT show color indicator next to component when colorScheme is "custom" but customColor is null', async () => {
      const currentView = createViewWithColorScheme(
        'custom',
        [{ componentId: 'comp-1' }],
        []
      );

      renderNavigationTree(currentView);

      await waitFor(() => {
        expect(screen.getByText('Payment Service')).toBeInTheDocument();
      });

      const paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      const colorIndicator = within(paymentServiceItem!).queryByTestId('custom-color-indicator');
      expect(colorIndicator).not.toBeInTheDocument();
    });

    it('should show correct color for each component with custom colors', async () => {
      const currentView = createViewWithColorScheme(
        'custom',
        [
          { componentId: 'comp-1', customColor: '#FF5733' },
          { componentId: 'comp-2', customColor: '#33AAFF' },
        ],
        []
      );

      renderNavigationTree(currentView);

      await waitFor(() => {
        expect(screen.getByText('Payment Service')).toBeInTheDocument();
      });

      const paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      const orderServiceItem = screen.getByText('Order Service').closest('.tree-item');

      const paymentColorIndicator = within(paymentServiceItem!).getByTestId('custom-color-indicator');
      const orderColorIndicator = within(orderServiceItem!).getByTestId('custom-color-indicator');

      expect(paymentColorIndicator).toHaveStyle({ backgroundColor: '#FF5733' });
      expect(orderColorIndicator).toHaveStyle({ backgroundColor: '#33AAFF' });
    });
  });

  describe('Capability Color Indicators', () => {
    it('should show custom color indicator next to capability when colorScheme is "custom" and capability has customColor', async () => {
      const currentView = createViewWithColorScheme(
        'custom',
        [],
        [
          { capabilityId: 'cap-1', customColor: '#AA00FF' },
          { capabilityId: 'cap-3' },
        ]
      );

      createMockStore(currentView, [
        { capabilityId: 'cap-1', x: 100, y: 200 },
        { capabilityId: 'cap-3', x: 100, y: 200 },
      ]);

      renderNavigationTree(currentView);

      await waitFor(() => {
        expect(screen.getByText('Customer Management')).toBeInTheDocument();
      });

      const customerMgmtItem = screen.getByText('Customer Management').closest('.capability-tree-item');
      const colorIndicator = within(customerMgmtItem!).getByTestId('custom-color-indicator');
      expect(colorIndicator).toBeInTheDocument();
      expect(colorIndicator).toHaveStyle({ backgroundColor: '#AA00FF' });
    });

    it('should NOT show color indicator next to capability when colorScheme is "maturity" even if customColor exists', async () => {
      const currentView = createViewWithColorScheme(
        'maturity',
        [],
        [{ capabilityId: 'cap-1', customColor: '#AA00FF' }]
      );

      createMockStore(currentView, [{ capabilityId: 'cap-1', x: 100, y: 200 }]);

      renderNavigationTree(currentView);

      await waitFor(() => {
        expect(screen.getByText('Customer Management')).toBeInTheDocument();
      });

      const customerMgmtItem = screen.getByText('Customer Management').closest('.capability-tree-item');
      const colorIndicator = within(customerMgmtItem!).queryByTestId('custom-color-indicator');
      expect(colorIndicator).not.toBeInTheDocument();
    });

    it('should NOT show color indicator next to capability when colorScheme is "custom" but customColor is null', async () => {
      const currentView = createViewWithColorScheme(
        'custom',
        [],
        [{ capabilityId: 'cap-1' }]
      );

      createMockStore(currentView, [{ capabilityId: 'cap-1', x: 100, y: 200 }]);

      renderNavigationTree(currentView);

      await waitFor(() => {
        expect(screen.getByText('Customer Management')).toBeInTheDocument();
      });

      const customerMgmtItem = screen.getByText('Customer Management').closest('.capability-tree-item');
      const colorIndicator = within(customerMgmtItem!).queryByTestId('custom-color-indicator');
      expect(colorIndicator).not.toBeInTheDocument();
    });

    it('should show correct color for each capability with custom colors', async () => {
      const currentView = createViewWithColorScheme(
        'custom',
        [],
        [
          { capabilityId: 'cap-1', customColor: '#FF00AA' },
          { capabilityId: 'cap-3', customColor: '#00FFAA' },
        ]
      );

      createMockStore(currentView, [
        { capabilityId: 'cap-1', x: 100, y: 200 },
        { capabilityId: 'cap-3', x: 100, y: 200 },
      ]);

      renderNavigationTree(currentView);

      await waitFor(() => {
        expect(screen.getByText('Customer Management')).toBeInTheDocument();
      });

      const customerMgmtItem = screen.getByText('Customer Management').closest('.capability-tree-item');
      const shippingItem = screen.getByText('Shipping').closest('.capability-tree-item');

      const customerColorIndicator = within(customerMgmtItem!).getByTestId('custom-color-indicator');
      const shippingColorIndicator = within(shippingItem!).getByTestId('custom-color-indicator');

      expect(customerColorIndicator).toHaveStyle({ backgroundColor: '#FF00AA' });
      expect(shippingColorIndicator).toHaveStyle({ backgroundColor: '#00FFAA' });
    });

    it('should show color indicator for child capabilities with custom colors', async () => {
      const currentView = createViewWithColorScheme(
        'custom',
        [],
        [
          { capabilityId: 'cap-1' },
          { capabilityId: 'cap-2', customColor: '#AABBCC' },
        ]
      );

      createMockStore(currentView, [
        { capabilityId: 'cap-1', x: 100, y: 200 },
        { capabilityId: 'cap-2', x: 100, y: 200 },
      ]);

      renderNavigationTree(currentView);

      await waitFor(() => {
        expect(screen.getByText('Customer Management')).toBeInTheDocument();
      });

      const customerMgmtItem = screen.getByText('Customer Management').closest('.capability-tree-item');
      const expandButton = within(customerMgmtItem!).getByRole('button');
      fireEvent.click(expandButton);

      const orderProcessingItem = screen.getByText('Order Processing').closest('.capability-tree-item');
      expect(orderProcessingItem).toBeInTheDocument();

      const colorIndicator = within(orderProcessingItem!).getByTestId('custom-color-indicator');
      expect(colorIndicator).toBeInTheDocument();
      expect(colorIndicator).toHaveStyle({ backgroundColor: '#AABBCC' });
    });
  });

  describe('Color Scheme Switching', () => {
    it('should show indicators when switching from maturity to custom scheme', async () => {
      const currentViewMaturity = createViewWithColorScheme(
        'maturity',
        [{ componentId: 'comp-1', customColor: '#FF5733' }],
        []
      );

      const { Wrapper } = createMantineTestWrapper();
      vi.mocked(useAppStore).mockImplementation((selector: (state: unknown) => unknown) =>
        selector(createMockStore())
      );
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: currentViewMaturity,
        currentViewId: currentViewMaturity.id,
        isLoading: false,
        error: null,
      });

      const { rerender } = render(<NavigationTree />, { wrapper: Wrapper });

      await waitFor(() => {
        expect(screen.getByText('Payment Service')).toBeInTheDocument();
      });

      let paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      let colorIndicator = within(paymentServiceItem!).queryByTestId('custom-color-indicator');
      expect(colorIndicator).not.toBeInTheDocument();

      const currentViewCustom = createViewWithColorScheme(
        'custom',
        [{ componentId: 'comp-1', customColor: '#FF5733' }],
        []
      );

      vi.mocked(useCurrentView).mockReturnValue({
        currentView: currentViewCustom,
        currentViewId: currentViewCustom.id,
        isLoading: false,
        error: null,
      });

      rerender(<NavigationTree />);

      paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      colorIndicator = within(paymentServiceItem!).getByTestId('custom-color-indicator');
      expect(colorIndicator).toBeInTheDocument();
      expect(colorIndicator).toHaveStyle({ backgroundColor: '#FF5733' });
    });

    it('should hide indicators when switching from custom to classic scheme', async () => {
      const currentViewCustom = createViewWithColorScheme(
        'custom',
        [],
        [{ capabilityId: 'cap-1', customColor: '#AA00FF' }]
      );

      const { Wrapper } = createMantineTestWrapper();
      vi.mocked(useAppStore).mockImplementation((selector: (state: unknown) => unknown) =>
        selector(createMockStore())
      );
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: currentViewCustom,
        currentViewId: currentViewCustom.id,
        isLoading: false,
        error: null,
      });

      const { rerender } = render(<NavigationTree />, { wrapper: Wrapper });

      await waitFor(() => {
        expect(screen.getByText('Customer Management')).toBeInTheDocument();
      });

      let customerMgmtItem = screen.getByText('Customer Management').closest('.capability-tree-item');
      let colorIndicator = within(customerMgmtItem!).getByTestId('custom-color-indicator');
      expect(colorIndicator).toBeInTheDocument();

      const currentViewClassic = createViewWithColorScheme(
        'classic',
        [],
        [{ capabilityId: 'cap-1', customColor: '#AA00FF' }]
      );

      vi.mocked(useCurrentView).mockReturnValue({
        currentView: currentViewClassic,
        currentViewId: currentViewClassic.id,
        isLoading: false,
        error: null,
      });

      rerender(<NavigationTree />);

      customerMgmtItem = screen.getByText('Customer Management').closest('.capability-tree-item');
      colorIndicator = within(customerMgmtItem!).queryByTestId('custom-color-indicator');
      expect(colorIndicator).not.toBeInTheDocument();
    });
  });

  describe('Edge Cases', () => {
    it('should handle null currentView gracefully', async () => {
      renderNavigationTree(null);

      await waitFor(() => {
        expect(screen.getByText('Payment Service')).toBeInTheDocument();
      });

      const paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      const colorIndicator = within(paymentServiceItem!).queryByTestId('custom-color-indicator');
      expect(colorIndicator).not.toBeInTheDocument();
    });

    it('should handle undefined colorScheme gracefully', async () => {
      const currentView = createViewWithColorScheme(
        'maturity',
        [{ componentId: 'comp-1', customColor: '#FF5733' }],
        []
      );
      currentView.colorScheme = undefined;

      renderNavigationTree(currentView);

      await waitFor(() => {
        expect(screen.getByText('Payment Service')).toBeInTheDocument();
      });

      const paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      const colorIndicator = within(paymentServiceItem!).queryByTestId('custom-color-indicator');
      expect(colorIndicator).not.toBeInTheDocument();
    });

    it('should handle mixed scenario with some elements having colors and others not', async () => {
      const currentView = createViewWithColorScheme(
        'custom',
        [
          { componentId: 'comp-1', customColor: '#FF5733' },
          { componentId: 'comp-2' },
        ],
        [
          { capabilityId: 'cap-1', customColor: '#AA00FF' },
          { capabilityId: 'cap-3' },
        ]
      );

      createMockStore(currentView, [
        { capabilityId: 'cap-1', x: 100, y: 200 },
        { capabilityId: 'cap-3', x: 100, y: 200 },
      ]);

      renderNavigationTree(currentView);

      await waitFor(() => {
        expect(screen.getByText('Payment Service')).toBeInTheDocument();
      });

      const paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      const orderServiceItem = screen.getByText('Order Service').closest('.tree-item');
      const customerMgmtItem = screen.getByText('Customer Management').closest('.capability-tree-item');
      const shippingItem = screen.getByText('Shipping').closest('.capability-tree-item');

      expect(within(paymentServiceItem!).getByTestId('custom-color-indicator')).toBeInTheDocument();
      expect(within(orderServiceItem!).queryByTestId('custom-color-indicator')).not.toBeInTheDocument();
      expect(within(customerMgmtItem!).getByTestId('custom-color-indicator')).toBeInTheDocument();
      expect(within(shippingItem!).queryByTestId('custom-color-indicator')).not.toBeInTheDocument();
    });
  });
});
