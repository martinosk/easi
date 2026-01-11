import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render } from '@testing-library/react';
import React from 'react';
import { ReactFlowProvider } from '@xyflow/react';

vi.mock('../../features/views/hooks/useCurrentView', () => ({
  useCurrentView: vi.fn(),
}));

import { useCurrentView } from '../../features/views/hooks/useCurrentView';
import type { View } from '../../api/types';
import { ComponentNode, type ComponentNodeData } from './ComponentNode';

const createMockView = (colorScheme: string, componentsWithColors?: Array<{ componentId: string; customColor?: string }>): View => ({
  id: 'view-1',
  name: 'Test View',
  description: 'Test view description',
  isDefault: true,
  components: componentsWithColors?.map(comp => ({
    componentId: comp.componentId,
    x: 100,
    y: 200,
    customColor: comp.customColor,
  })) || [],
  capabilities: [],
  colorScheme,
  createdAt: '2024-01-01T00:00:00Z',
  _links: { self: { href: '/api/v1/views/view-1' } },
});

const createComponentNodeData = (
  isSelected: boolean = false,
  customColor?: string
): ComponentNodeData => ({
  label: 'Payment Service',
  description: 'Handles payment processing',
  isSelected,
  customColor,
});

const renderWithProvider = (component: React.ReactElement) => {
  const result = render(
    <ReactFlowProvider>
      {component}
    </ReactFlowProvider>
  );

  return {
    ...result,
    rerender: (newComponent: React.ReactElement) => {
      return result.rerender(
        <ReactFlowProvider>
          {newComponent}
        </ReactFlowProvider>
      );
    },
  };
};

const hexToRgb = (hex: string): string => {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  if (!result) return hex;
  const r = parseInt(result[1], 16);
  const g = parseInt(result[2], 16);
  const b = parseInt(result[3], 16);
  return `rgb(${r}, ${g}, ${b})`;
};

describe('ComponentNode Custom Color Rendering', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Custom Color Scheme with Custom Color Set', () => {
    it('should use customColor when colorScheme is "custom" and customColor is provided', () => {
      const mockView = createMockView('custom', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const nodeData = createComponentNodeData(false, '#FF5733');
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node).toBeTruthy();
      expect(node.style.background).toContain(hexToRgb('#FF5733'));
    });

    it('should apply custom color as gradient with opacity when colorScheme is "custom"', () => {
      const mockView = createMockView('custom', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const nodeData = createComponentNodeData(false, '#FF5733');
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toMatch(/linear-gradient.*rgb\(\d+,\s*\d+,\s*\d+\)/);
    });

    it('should use customColor for border color when colorScheme is "custom" and element is not selected', () => {
      const mockView = createMockView('custom', [
        { componentId: 'comp-1', customColor: '#22AA88' },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const nodeData = createComponentNodeData(false, '#22AA88');
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.borderColor).toBe(hexToRgb('#22AA88'));
    });
  });

  describe('Custom Color Scheme without Custom Color (Neutral Default)', () => {
    it('should use neutral default color #E0E0E0 when colorScheme is "custom" and customColor is null', () => {
      const mockView = createMockView('custom', [
        { componentId: 'comp-2', customColor: undefined },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const nodeData = createComponentNodeData();
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-2" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#E0E0E0'));
    });

    it('should use neutral default for border when colorScheme is "custom" and customColor is null', () => {
      const mockView = createMockView('custom', [
        { componentId: 'comp-2', customColor: undefined },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const nodeData = createComponentNodeData(false);
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-2" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.borderColor).toBe(hexToRgb('#E0E0E0'));
    });

    it('should use neutral default when colorScheme is "custom" and customColor is undefined', () => {
      const mockView = createMockView('custom', [
        { componentId: 'comp-3', customColor: undefined },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const nodeData = createComponentNodeData();
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-3" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#E0E0E0'));
    });
  });

  describe('Non-Custom Color Schemes Ignore Custom Colors', () => {
    it('should ignore customColor when colorScheme is "maturity"', () => {
      const mockView = createMockView('maturity', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const nodeData = createComponentNodeData();
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toMatch(/rgb\(\d+,\s*\d+,\s*\d+\)/);
      expect(node.style.background).not.toContain(hexToRgb('#FF5733'));
    });

    it('should ignore customColor when colorScheme is "classic"', () => {
      const mockView = createMockView('classic', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const nodeData = createComponentNodeData();
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toMatch(/rgb\(\d+,\s*\d+,\s*\d+\)/);
      expect(node.style.background).not.toContain(hexToRgb('#FF5733'));
    });
  });
 
  describe('Default Color Scheme Behavior', () => {
    it('should apply scheme-based color when colorScheme is undefined', () => {
      const mockView = createMockView('maturity', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      mockView.colorScheme = undefined;
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const nodeData = createComponentNodeData();
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toMatch(/rgb\(\d+,\s*\d+,\s*\d+\)/);
      expect(node.style.background).not.toContain(hexToRgb('#FF5733'));
    });

    it('should apply scheme-based color when currentView is null', () => {
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: null,
        currentViewId: null,
        isLoading: false,
        error: null,
      });

      const nodeData = createComponentNodeData();
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toMatch(/rgb\(\d+,\s*\d+,\s*\d+\)/);
    });
  });

  describe('Color Reactivity and Dynamic Updates', () => {
    it('should update color when customColor changes in custom scheme', () => {
      const mockView = createMockView('custom', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const nodeData = createComponentNodeData(false, '#FF5733');
      const { container, rerender } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      let node = container.querySelector('.component-node') as HTMLElement;
      const initialBackground = node.style.background;
      expect(initialBackground).toContain(hexToRgb('#FF5733'));

      mockView.components[0].customColor = '#33AAFF';
      const updatedNodeData = createComponentNodeData(false, '#33AAFF');
      rerender(<ComponentNode data={updatedNodeData} id="comp-1" />);

      node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#33AAFF'));
      expect(node.style.background).not.toBe(initialBackground);
    });

    it('should switch from custom color to scheme color when scheme changes from "custom" to "maturity"', () => {
      let mockView = createMockView('custom', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const nodeData = createComponentNodeData(false, '#FF5733');
      const { container, rerender } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      let node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#FF5733'));

      mockView = createMockView('maturity', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const updatedNodeData = createComponentNodeData(false, '#FF5733');
      rerender(<ComponentNode data={updatedNodeData} id="comp-1" />);

      node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toMatch(/rgb\(\d+,\s*\d+,\s*\d+\)/);
      expect(node.style.background).not.toContain(hexToRgb('#FF5733'));
    });

    it('should switch from scheme color to custom color when scheme changes from "maturity" to "custom"', () => {
      let mockView = createMockView('maturity', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const nodeData = createComponentNodeData(false, '#FF5733');
      const { container, rerender } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      let node = container.querySelector('.component-node') as HTMLElement;
      const schemeColor = node.style.background;
      expect(schemeColor).toMatch(/rgb\(\d+,\s*\d+,\s*\d+\)/);
      expect(schemeColor).not.toContain(hexToRgb('#FF5733'));

      mockView = createMockView('custom', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const updatedNodeData = createComponentNodeData(false, '#FF5733');
      rerender(<ComponentNode data={updatedNodeData} id="comp-1" />);

      node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#FF5733'));
      expect(node.style.background).not.toBe(schemeColor);
    });

    it('should switch from scheme color to custom color when scheme changes from "classic" to "custom"', () => {
      let mockView = createMockView('classic', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const nodeData = createComponentNodeData(false, '#FF5733');
      const { container, rerender } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      let node = container.querySelector('.component-node') as HTMLElement;
      const schemeColor = node.style.background;
      expect(schemeColor).toMatch(/rgb\(\d+,\s*\d+,\s*\d+\)/);
      expect(schemeColor).not.toContain(hexToRgb('#FF5733'));

      mockView = createMockView('custom', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const updatedNodeData = createComponentNodeData(false, '#FF5733');
      rerender(<ComponentNode data={updatedNodeData} id="comp-1" />);

      node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#FF5733'));
      expect(node.style.background).not.toBe(schemeColor);
    });

    it('should update to neutral default when custom color is removed in custom scheme', () => {
      const mockView = createMockView('custom', [
        { componentId: 'comp-1', customColor: '#FF5733' },
        { componentId: 'comp-2', customColor: undefined },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const nodeData = createComponentNodeData(false, '#FF5733');
      const { container, rerender } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      let node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#FF5733'));

      const updatedNodeData = createComponentNodeData(false, undefined);
      rerender(<ComponentNode data={updatedNodeData} id="comp-2" />);

      node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#E0E0E0'));
      expect(node.style.background).not.toContain(hexToRgb('#FF5733'));
    });
  });

  describe('Border Color with Selection State', () => {
    it('should use selected border color when element is selected, regardless of color scheme', () => {
      const mockView = createMockView('custom', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const nodeData = createComponentNodeData(true);
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.borderColor).toBe(hexToRgb('#374151'));
    });

    it('should use custom color for border when element is not selected in custom scheme', () => {
      const mockView = createMockView('custom', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const nodeData = createComponentNodeData(false, '#FF5733');
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.borderColor).toBe(hexToRgb('#FF5733'));
      expect(node.style.borderColor).not.toBe(hexToRgb('#374151'));
    });

    it('should use scheme color for border when element is not selected in non-custom scheme', () => {
      const mockView = createMockView('maturity', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const nodeData = createComponentNodeData(false);
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.borderColor).toMatch(/rgb\(\d+,\s*\d+,\s*\d+\)/);
      expect(node.style.borderColor).not.toBe(hexToRgb('#FF5733'));
      expect(node.style.borderColor).not.toBe(hexToRgb('#374151'));
    });
  });

  describe('Edge Cases', () => {
    it('should handle empty string customColor as null in custom scheme', () => {
      const mockView = createMockView('custom', [
        { componentId: 'comp-4', customColor: '' },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const nodeData = createComponentNodeData();
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-4" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#E0E0E0'));
    });

    it('should handle component not in view with default color in custom scheme', () => {
      const mockView = createMockView('custom', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const nodeData = createComponentNodeData();
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-999" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#E0E0E0'));
    });

    it('should handle multiple components with different custom colors in same view', () => {
      const mockView = createMockView('custom', [
        { componentId: 'comp-1', customColor: '#FF5733' },
        { componentId: 'comp-2', customColor: '#33FF57' },
        { componentId: 'comp-3', customColor: '#3357FF' },
      ]);
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: mockView,
        currentViewId: mockView.id,
        isLoading: false,
        error: null,
      });

      const nodeData1 = createComponentNodeData(false, '#FF5733');
      const { container: container1 } = renderWithProvider(
        <ComponentNode data={nodeData1} id="comp-1" />
      );
      const node1 = container1.querySelector('.component-node') as HTMLElement;
      expect(node1.style.background).toContain(hexToRgb('#FF5733'));

      const nodeData2 = createComponentNodeData(false, '#33FF57');
      const { container: container2 } = renderWithProvider(
        <ComponentNode data={nodeData2} id="comp-2" />
      );
      const node2 = container2.querySelector('.component-node') as HTMLElement;
      expect(node2.style.background).toContain(hexToRgb('#33FF57'));

      const nodeData3 = createComponentNodeData(false, '#3357FF');
      const { container: container3 } = renderWithProvider(
        <ComponentNode data={nodeData3} id="comp-3" />
      );
      const node3 = container3.querySelector('.component-node') as HTMLElement;
      expect(node3.style.background).toContain(hexToRgb('#3357FF'));
    });
  });
});
