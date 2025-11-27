import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, within, fireEvent } from '@testing-library/react';
import { NavigationTree } from './NavigationTree';
import type { Component, Capability, View } from '../../../api/types';

vi.mock('../../../store/appStore', () => ({
  useAppStore: vi.fn(),
}));

vi.mock('../../../api/client');

import { useAppStore } from '../../../store/appStore';

const mockUseAppStore = useAppStore as unknown as ReturnType<typeof vi.fn>;

describe('NavigationTree - Custom Color Indicators', () => {
  const mockComponents: Component[] = [
    {
      id: 'comp-1',
      name: 'Payment Service',
      description: 'Handles payments',
      createdAt: '2024-01-01',
      _links: { self: { href: '/api/v1/components/comp-1' } },
    },
    {
      id: 'comp-2',
      name: 'Order Service',
      description: 'Handles orders',
      createdAt: '2024-01-01',
      _links: { self: { href: '/api/v1/components/comp-2' } },
    },
  ];

  const mockCapabilities: Capability[] = [
    {
      id: 'cap-1',
      name: 'Customer Management',
      description: 'Manage customers',
      level: 'L1',
      maturityLevel: 'Product',
      createdAt: '2024-01-01',
      _links: { self: { href: '/api/v1/capabilities/cap-1' } },
    },
    {
      id: 'cap-2',
      name: 'Order Processing',
      description: 'Process orders',
      level: 'L2',
      parentId: 'cap-1',
      maturityLevel: 'Custom Build',
      createdAt: '2024-01-01',
      _links: { self: { href: '/api/v1/capabilities/cap-2' } },
    },
    {
      id: 'cap-3',
      name: 'Shipping',
      description: 'Shipping operations',
      level: 'L1',
      maturityLevel: 'Genesis',
      createdAt: '2024-01-01',
      _links: { self: { href: '/api/v1/capabilities/cap-3' } },
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
      componentId: c.componentId,
      x: 100,
      y: 200,
      customColor: c.customColor,
    })),
    capabilities: capabilities.map(c => ({
      capabilityId: c.capabilityId,
      x: 150,
      y: 250,
      customColor: c.customColor,
    })),
    createdAt: '2024-01-01',
    _links: { self: { href: '/api/v1/views/view-1' } },
  });

  beforeEach(() => {
    vi.clearAllMocks();
    mockUseAppStore.mockImplementation((selector) => {
      const state = {
        components: mockComponents,
        capabilities: mockCapabilities,
        currentView: null,
        views: [],
        selectedNodeId: null,
        canvasCapabilities: [],
        loadViews: vi.fn(),
        loadCapabilities: vi.fn(),
        updateComponent: vi.fn(),
        deleteComponent: vi.fn(),
      };
      return selector(state);
    });
  });

  describe('Component Color Indicators', () => {
    it('should show custom color indicator next to component when colorScheme is "custom" and component has customColor', () => {
      const currentView = createViewWithColorScheme(
        'custom',
        [
          { componentId: 'comp-1', customColor: '#FF5733' },
          { componentId: 'comp-2' },
        ],
        []
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      render(<NavigationTree />);

      const paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      expect(paymentServiceItem).toBeInTheDocument();

      const colorIndicator = within(paymentServiceItem!).getByTestId('custom-color-indicator');
      expect(colorIndicator).toBeInTheDocument();
      expect(colorIndicator).toHaveStyle({ backgroundColor: '#FF5733' });
    });

    it('should NOT show color indicator next to component when colorScheme is "maturity" even if customColor exists', () => {
      const currentView = createViewWithColorScheme(
        'maturity',
        [{ componentId: 'comp-1', customColor: '#FF5733' }],
        []
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      render(<NavigationTree />);

      const paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      expect(paymentServiceItem).toBeInTheDocument();

      const colorIndicator = within(paymentServiceItem!).queryByTestId('custom-color-indicator');
      expect(colorIndicator).not.toBeInTheDocument();
    });

    it('should NOT show color indicator next to component when colorScheme is "classic" even if customColor exists', () => {
      const currentView = createViewWithColorScheme(
        'classic',
        [{ componentId: 'comp-1', customColor: '#00AA00' }],
        []
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      render(<NavigationTree />);

      const paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      expect(paymentServiceItem).toBeInTheDocument();

      const colorIndicator = within(paymentServiceItem!).queryByTestId('custom-color-indicator');
      expect(colorIndicator).not.toBeInTheDocument();
    });

    it('should NOT show color indicator next to component when colorScheme is "custom" but customColor is null', () => {
      const currentView = createViewWithColorScheme(
        'custom',
        [{ componentId: 'comp-1' }],
        []
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      render(<NavigationTree />);

      const paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      expect(paymentServiceItem).toBeInTheDocument();

      const colorIndicator = within(paymentServiceItem!).queryByTestId('custom-color-indicator');
      expect(colorIndicator).not.toBeInTheDocument();
    });

    it('should show correct color for each component with custom colors', () => {
      const currentView = createViewWithColorScheme(
        'custom',
        [
          { componentId: 'comp-1', customColor: '#FF5733' },
          { componentId: 'comp-2', customColor: '#33AAFF' },
        ],
        []
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      render(<NavigationTree />);

      const paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      const orderServiceItem = screen.getByText('Order Service').closest('.tree-item');

      const paymentColorIndicator = within(paymentServiceItem!).getByTestId('custom-color-indicator');
      const orderColorIndicator = within(orderServiceItem!).getByTestId('custom-color-indicator');

      expect(paymentColorIndicator).toHaveStyle({ backgroundColor: '#FF5733' });
      expect(orderColorIndicator).toHaveStyle({ backgroundColor: '#33AAFF' });
    });
  });

  describe('Capability Color Indicators', () => {
    it('should show custom color indicator next to capability when colorScheme is "custom" and capability has customColor', () => {
      const currentView = createViewWithColorScheme(
        'custom',
        [],
        [
          { capabilityId: 'cap-1', customColor: '#AA00FF' },
          { capabilityId: 'cap-3' },
        ]
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [
            { capabilityId: 'cap-1', x: 100, y: 200 },
            { capabilityId: 'cap-3', x: 100, y: 200 },
          ],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      render(<NavigationTree />);

      const customerMgmtItem = screen.getByText('Customer Management').closest('.capability-tree-item');
      expect(customerMgmtItem).toBeInTheDocument();

      const colorIndicator = within(customerMgmtItem!).getByTestId('custom-color-indicator');
      expect(colorIndicator).toBeInTheDocument();
      expect(colorIndicator).toHaveStyle({ backgroundColor: '#AA00FF' });
    });

    it('should NOT show color indicator next to capability when colorScheme is "maturity" even if customColor exists', () => {
      const currentView = createViewWithColorScheme(
        'maturity',
        [],
        [{ capabilityId: 'cap-1', customColor: '#AA00FF' }]
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [{ capabilityId: 'cap-1', x: 100, y: 200 }],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      render(<NavigationTree />);

      const customerMgmtItem = screen.getByText('Customer Management').closest('.capability-tree-item');
      expect(customerMgmtItem).toBeInTheDocument();

      const colorIndicator = within(customerMgmtItem!).queryByTestId('custom-color-indicator');
      expect(colorIndicator).not.toBeInTheDocument();
    });

    it('should NOT show color indicator next to capability when colorScheme is "custom" but customColor is null', () => {
      const currentView = createViewWithColorScheme(
        'custom',
        [],
        [{ capabilityId: 'cap-1' }]
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [{ capabilityId: 'cap-1', x: 100, y: 200 }],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      render(<NavigationTree />);

      const customerMgmtItem = screen.getByText('Customer Management').closest('.capability-tree-item');
      expect(customerMgmtItem).toBeInTheDocument();

      const colorIndicator = within(customerMgmtItem!).queryByTestId('custom-color-indicator');
      expect(colorIndicator).not.toBeInTheDocument();
    });

    it('should show correct color for each capability with custom colors', () => {
      const currentView = createViewWithColorScheme(
        'custom',
        [],
        [
          { capabilityId: 'cap-1', customColor: '#FF00AA' },
          { capabilityId: 'cap-3', customColor: '#00FFAA' },
        ]
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [
            { capabilityId: 'cap-1', x: 100, y: 200 },
            { capabilityId: 'cap-3', x: 100, y: 200 },
          ],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      render(<NavigationTree />);

      const customerMgmtItem = screen.getByText('Customer Management').closest('.capability-tree-item');
      const shippingItem = screen.getByText('Shipping').closest('.capability-tree-item');

      const customerColorIndicator = within(customerMgmtItem!).getByTestId('custom-color-indicator');
      const shippingColorIndicator = within(shippingItem!).getByTestId('custom-color-indicator');

      expect(customerColorIndicator).toHaveStyle({ backgroundColor: '#FF00AA' });
      expect(shippingColorIndicator).toHaveStyle({ backgroundColor: '#00FFAA' });
    });

    it('should show color indicator for child capabilities with custom colors', () => {
      const currentView = createViewWithColorScheme(
        'custom',
        [],
        [
          { capabilityId: 'cap-1' },
          { capabilityId: 'cap-2', customColor: '#AABBCC' },
        ]
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [
            { capabilityId: 'cap-1', x: 100, y: 200 },
            { capabilityId: 'cap-2', x: 100, y: 200 },
          ],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      render(<NavigationTree />);

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
    it('should show indicators when switching from maturity to custom scheme', () => {
      const currentViewMaturity = createViewWithColorScheme(
        'maturity',
        [{ componentId: 'comp-1', customColor: '#FF5733' }],
        []
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView: currentViewMaturity,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      const { rerender } = render(<NavigationTree />);

      let paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      let colorIndicator = within(paymentServiceItem!).queryByTestId('custom-color-indicator');
      expect(colorIndicator).not.toBeInTheDocument();

      const currentViewCustom = createViewWithColorScheme(
        'custom',
        [{ componentId: 'comp-1', customColor: '#FF5733' }],
        []
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView: currentViewCustom,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      rerender(<NavigationTree />);

      paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      colorIndicator = within(paymentServiceItem!).getByTestId('custom-color-indicator');
      expect(colorIndicator).toBeInTheDocument();
      expect(colorIndicator).toHaveStyle({ backgroundColor: '#FF5733' });
    });

    it('should hide indicators when switching from custom to classic scheme', () => {
      const currentViewCustom = createViewWithColorScheme(
        'custom',
        [],
        [{ capabilityId: 'cap-1', customColor: '#AA00FF' }]
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView: currentViewCustom,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [{ capabilityId: 'cap-1', x: 100, y: 200 }],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      const { rerender } = render(<NavigationTree />);

      let customerMgmtItem = screen.getByText('Customer Management').closest('.capability-tree-item');
      let colorIndicator = within(customerMgmtItem!).getByTestId('custom-color-indicator');
      expect(colorIndicator).toBeInTheDocument();

      const currentViewClassic = createViewWithColorScheme(
        'classic',
        [],
        [{ capabilityId: 'cap-1', customColor: '#AA00FF' }]
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView: currentViewClassic,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [{ capabilityId: 'cap-1', x: 100, y: 200 }],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      rerender(<NavigationTree />);

      customerMgmtItem = screen.getByText('Customer Management').closest('.capability-tree-item');
      colorIndicator = within(customerMgmtItem!).queryByTestId('custom-color-indicator');
      expect(colorIndicator).not.toBeInTheDocument();
    });
  });

  describe('Color Change Reactivity', () => {
    it('should update component color indicator when custom color changes', () => {
      const currentView = createViewWithColorScheme(
        'custom',
        [{ componentId: 'comp-1', customColor: '#FF5733' }],
        []
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      const { rerender } = render(<NavigationTree />);

      let paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      let colorIndicator = within(paymentServiceItem!).getByTestId('custom-color-indicator');
      expect(colorIndicator).toHaveStyle({ backgroundColor: '#FF5733' });

      const updatedView = createViewWithColorScheme(
        'custom',
        [{ componentId: 'comp-1', customColor: '#00AA00' }],
        []
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView: updatedView,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      rerender(<NavigationTree />);

      paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      colorIndicator = within(paymentServiceItem!).getByTestId('custom-color-indicator');
      expect(colorIndicator).toHaveStyle({ backgroundColor: '#00AA00' });
    });

    it('should update capability color indicator when custom color changes', () => {
      const currentView = createViewWithColorScheme(
        'custom',
        [],
        [{ capabilityId: 'cap-1', customColor: '#AA00FF' }]
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [{ capabilityId: 'cap-1', x: 100, y: 200 }],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      const { rerender } = render(<NavigationTree />);

      let customerMgmtItem = screen.getByText('Customer Management').closest('.capability-tree-item');
      let colorIndicator = within(customerMgmtItem!).getByTestId('custom-color-indicator');
      expect(colorIndicator).toHaveStyle({ backgroundColor: '#AA00FF' });

      const updatedView = createViewWithColorScheme(
        'custom',
        [],
        [{ capabilityId: 'cap-1', customColor: '#FFAA00' }]
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView: updatedView,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [{ capabilityId: 'cap-1', x: 100, y: 200 }],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      rerender(<NavigationTree />);

      customerMgmtItem = screen.getByText('Customer Management').closest('.capability-tree-item');
      colorIndicator = within(customerMgmtItem!).getByTestId('custom-color-indicator');
      expect(colorIndicator).toHaveStyle({ backgroundColor: '#FFAA00' });
    });

    it('should remove indicator when custom color is cleared', () => {
      const currentView = createViewWithColorScheme(
        'custom',
        [{ componentId: 'comp-1', customColor: '#FF5733' }],
        []
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      const { rerender } = render(<NavigationTree />);

      let paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      let colorIndicator = within(paymentServiceItem!).getByTestId('custom-color-indicator');
      expect(colorIndicator).toBeInTheDocument();

      const updatedView = createViewWithColorScheme(
        'custom',
        [{ componentId: 'comp-1' }],
        []
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView: updatedView,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      rerender(<NavigationTree />);

      paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      colorIndicator = within(paymentServiceItem!).queryByTestId('custom-color-indicator');
      expect(colorIndicator).not.toBeInTheDocument();
    });

    it('should add indicator when custom color is newly assigned', () => {
      const currentView = createViewWithColorScheme(
        'custom',
        [{ componentId: 'comp-1' }],
        []
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      const { rerender } = render(<NavigationTree />);

      let paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      let colorIndicator = within(paymentServiceItem!).queryByTestId('custom-color-indicator');
      expect(colorIndicator).not.toBeInTheDocument();

      const updatedView = createViewWithColorScheme(
        'custom',
        [{ componentId: 'comp-1', customColor: '#FF5733' }],
        []
      );

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView: updatedView,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      rerender(<NavigationTree />);

      paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      colorIndicator = within(paymentServiceItem!).getByTestId('custom-color-indicator');
      expect(colorIndicator).toBeInTheDocument();
      expect(colorIndicator).toHaveStyle({ backgroundColor: '#FF5733' });
    });
  });

  describe('Edge Cases', () => {
    it('should handle null currentView gracefully', () => {
      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView: null,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      render(<NavigationTree />);

      const paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      const colorIndicator = within(paymentServiceItem!).queryByTestId('custom-color-indicator');
      expect(colorIndicator).not.toBeInTheDocument();
    });

    it('should handle undefined colorScheme gracefully', () => {
      const currentView = createViewWithColorScheme(
        'maturity',
        [{ componentId: 'comp-1', customColor: '#FF5733' }],
        []
      );
      currentView.colorScheme = undefined;

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      render(<NavigationTree />);

      const paymentServiceItem = screen.getByText('Payment Service').closest('.tree-item');
      const colorIndicator = within(paymentServiceItem!).queryByTestId('custom-color-indicator');
      expect(colorIndicator).not.toBeInTheDocument();
    });

    it('should handle mixed scenario with some elements having colors and others not', () => {
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

      mockUseAppStore.mockImplementation((selector) => {
        const state = {
          components: mockComponents,
          capabilities: mockCapabilities,
          currentView,
          views: [],
          selectedNodeId: null,
          canvasCapabilities: [
            { capabilityId: 'cap-1', x: 100, y: 200 },
            { capabilityId: 'cap-3', x: 100, y: 200 },
          ],
          loadViews: vi.fn(),
          loadCapabilities: vi.fn(),
          updateComponent: vi.fn(),
          deleteComponent: vi.fn(),
        };
        return selector(state);
      });

      render(<NavigationTree />);

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
