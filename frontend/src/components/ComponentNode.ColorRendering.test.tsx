import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render } from '@testing-library/react';
import React from 'react';
import { ReactFlowProvider } from '@xyflow/react';

vi.mock('../store/appStore', () => ({
  useAppStore: vi.fn(),
}));

import { useAppStore } from '../store/appStore';
import type { View } from '../api/types';
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
  isSelected: boolean = false
): ComponentNodeData => ({
  label: 'Payment Service',
  description: 'Handles payment processing',
  isSelected,
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
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

      const nodeData = createComponentNodeData();
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
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

      const nodeData = createComponentNodeData();
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
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

      const nodeData = createComponentNodeData(false);
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
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

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
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

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
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

      const nodeData = createComponentNodeData();
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-3" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#E0E0E0'));
    });
  });

  describe('Non-Custom Color Schemes Ignore Custom Colors', () => {
    it('should use default component color when colorScheme is "maturity", ignoring customColor', () => {
      const mockView = createMockView('maturity', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

      const nodeData = createComponentNodeData();
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#3b82f6'));
      expect(node.style.background).not.toContain(hexToRgb('#FF5733'));
    });

    it('should use archimate color when colorScheme is "archimate", ignoring customColor', () => {
      const mockView = createMockView('archimate', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

      const nodeData = createComponentNodeData();
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#B5FFFF'));
      expect(node.style.background).not.toContain(hexToRgb('#FF5733'));
    });

    it('should use archimate-classic color when colorScheme is "archimate-classic", ignoring customColor', () => {
      const mockView = createMockView('archimate-classic', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

      const nodeData = createComponentNodeData();
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#B5FFFF'));
      expect(node.style.background).not.toContain(hexToRgb('#FF5733'));
    });
  });

  describe('Default Color Scheme Behavior', () => {
    it('should use default component color when colorScheme is undefined', () => {
      const mockView = createMockView('maturity', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      mockView.colorScheme = undefined;
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

      const nodeData = createComponentNodeData();
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#3b82f6'));
    });

    it('should use default component color when currentView is null', () => {
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: null })
      );

      const nodeData = createComponentNodeData();
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#3b82f6'));
    });
  });

  describe('Color Reactivity and Dynamic Updates', () => {
    it('should update color when customColor changes in custom scheme', () => {
      const mockView = createMockView('custom', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

      const nodeData = createComponentNodeData();
      const { container, rerender } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      let node = container.querySelector('.component-node') as HTMLElement;
      const initialBackground = node.style.background;
      expect(initialBackground).toContain(hexToRgb('#FF5733'));

      mockView.components[0].customColor = '#33AAFF';
      rerender(<ComponentNode data={nodeData} id="comp-1" />);

      node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#33AAFF'));
      expect(node.style.background).not.toBe(initialBackground);
    });

    it('should switch from custom color to default color when scheme changes from "custom" to "maturity"', () => {
      let mockView = createMockView('custom', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

      const nodeData = createComponentNodeData();
      const { container, rerender } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      let node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#FF5733'));

      mockView = createMockView('maturity', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

      rerender(<ComponentNode data={nodeData} id="comp-1" />);

      node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#3b82f6'));
      expect(node.style.background).not.toContain(hexToRgb('#FF5733'));
    });

    it('should switch from default color to custom color when scheme changes from "maturity" to "custom"', () => {
      let mockView = createMockView('maturity', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

      const nodeData = createComponentNodeData();
      const { container, rerender } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      let node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#3b82f6'));

      mockView = createMockView('custom', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

      rerender(<ComponentNode data={nodeData} id="comp-1" />);

      node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#FF5733'));
      expect(node.style.background).not.toContain(hexToRgb('#3b82f6'));
    });

    it('should switch from archimate color to custom color when scheme changes', () => {
      let mockView = createMockView('archimate', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

      const nodeData = createComponentNodeData();
      const { container, rerender } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      let node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#B5FFFF'));

      mockView = createMockView('custom', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

      rerender(<ComponentNode data={nodeData} id="comp-1" />);

      node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#FF5733'));
    });

    it('should update to neutral default when custom color is removed in custom scheme', () => {
      const mockView = createMockView('custom', [
        { componentId: 'comp-1', customColor: '#FF5733' },
        { componentId: 'comp-2', customColor: undefined },
      ]);
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

      const nodeData = createComponentNodeData();
      const { container, rerender } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      let node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toContain(hexToRgb('#FF5733'));

      rerender(<ComponentNode data={nodeData} id="comp-2" />);

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
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

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
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

      const nodeData = createComponentNodeData(false);
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.borderColor).toBe(hexToRgb('#FF5733'));
      expect(node.style.borderColor).not.toBe(hexToRgb('#374151'));
    });

    it('should use default component color for border when element is not selected in maturity scheme', () => {
      const mockView = createMockView('maturity', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

      const nodeData = createComponentNodeData(false);
      const { container } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.borderColor).toBe(hexToRgb('#3b82f6'));
    });
  });

  describe('Edge Cases', () => {
    it('should handle empty string customColor as null in custom scheme', () => {
      const mockView = createMockView('custom', [
        { componentId: 'comp-4', customColor: '' },
      ]);
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

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
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

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
      vi.mocked(useAppStore).mockImplementation((selector: (state: { currentView: View | null }) => unknown) =>
        selector({ currentView: mockView })
      );

      const nodeData = createComponentNodeData();

      const { container: container1 } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-1" />
      );
      const node1 = container1.querySelector('.component-node') as HTMLElement;
      expect(node1.style.background).toContain(hexToRgb('#FF5733'));

      const { container: container2 } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-2" />
      );
      const node2 = container2.querySelector('.component-node') as HTMLElement;
      expect(node2.style.background).toContain(hexToRgb('#33FF57'));

      const { container: container3 } = renderWithProvider(
        <ComponentNode data={nodeData} id="comp-3" />
      );
      const node3 = container3.querySelector('.component-node') as HTMLElement;
      expect(node3.style.background).toContain(hexToRgb('#3357FF'));
    });
  });
});
