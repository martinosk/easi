import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render } from '@testing-library/react';
import React from 'react';
import { ReactFlowProvider } from '@xyflow/react';

vi.mock('../../features/views/hooks/useCurrentView', () => ({
  useCurrentView: vi.fn(),
}));

import { useCurrentView } from '../../features/views/hooks/useCurrentView';
import type { View } from '../../api/types';
import { toViewId, toComponentId } from '../../api/types';
import { ComponentNode, type ComponentNodeData } from './ComponentNode';

type HexColor = `#${string}`;

interface ViewComponent {
  componentId: string;
  customColor?: HexColor | string;
}

const createMockView = (colorScheme: string, componentsWithColors?: ViewComponent[]): View => ({
  id: toViewId('view-1'),
  name: 'Test View',
  description: 'Test view description',
  isDefault: true,
  isPrivate: false,
  components: componentsWithColors?.map(comp => ({
    componentId: toComponentId(comp.componentId),
    x: 100,
    y: 200,
    customColor: comp.customColor,
  })) || [],
  capabilities: [],
  originEntities: [],
  colorScheme,
  createdAt: '2024-01-01T00:00:00Z',
  _links: { self: { href: '/api/v1/views/view-1', method: 'GET' } },
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

const hexToRgb = (hex: HexColor): string => {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  if (!result) return hex;
  const r = parseInt(result[1], 16);
  const g = parseInt(result[2], 16);
  const b = parseInt(result[3], 16);
  return `rgb(${r}, ${g}, ${b})`;
};

const toColorVariants = (hex: HexColor) => ({
  rgb: hexToRgb(hex),
  upper: hex.toUpperCase(),
  lower: hex.toLowerCase(),
});

const containsColor = (styleValue: string, hex: HexColor): boolean => {
  const { rgb, upper, lower } = toColorVariants(hex);
  return styleValue.includes(rgb) || styleValue.includes(upper) || styleValue.includes(lower);
};

const colorMatches = (styleValue: string, hex: HexColor): boolean => {
  const { rgb, upper, lower } = toColorVariants(hex);
  return styleValue === rgb || styleValue === upper || styleValue === lower;
};

const mockCurrentView = (view: View | null) => {
  vi.mocked(useCurrentView).mockReturnValue({
    currentView: view,
    currentViewId: view?.id ?? null,
    isLoading: false,
    error: null,
  });
  return view;
};

interface RenderNodeOptions {
  colorScheme: string;
  componentId: string;
  nodeData: ComponentNodeData;
  components?: ViewComponent[];
}

const renderNode = ({ colorScheme, componentId, nodeData, components }: RenderNodeOptions) => {
  const mockView = createMockView(colorScheme, components ?? [{ componentId, customColor: nodeData.customColor }]);
  mockCurrentView(mockView);
  return {
    mockView,
    ...renderWithProvider(<ComponentNode data={nodeData} id={componentId} />),
    getNode: (container: HTMLElement) => container.querySelector('.component-node') as HTMLElement,
  };
};

describe('ComponentNode Custom Color Rendering', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Custom Color Scheme with Custom Color Set', () => {
    it('should use customColor when colorScheme is "custom" and customColor is provided', () => {
      const { container, getNode } = renderNode({ colorScheme: 'custom', componentId: 'comp-1', nodeData: createComponentNodeData(false, '#FF5733') });
      const node = getNode(container);
      expect(node).toBeTruthy();
      expect(containsColor(node.style.background, '#FF5733')).toBe(true);
    });

    it('should apply custom color as gradient with opacity when colorScheme is "custom"', () => {
      const { container, getNode } = renderNode({ colorScheme: 'custom', componentId: 'comp-1', nodeData: createComponentNodeData(false, '#FF5733') });
      expect(getNode(container).style.background).toMatch(/linear-gradient/);
    });

    it('should use customColor for border color when colorScheme is "custom" and element is not selected', () => {
      const { container, getNode } = renderNode({ colorScheme: 'custom', componentId: 'comp-1', nodeData: createComponentNodeData(false, '#22AA88'),
        components: [{ componentId: 'comp-1', customColor: '#22AA88' }] });
      expect(colorMatches(getNode(container).style.borderColor, '#22AA88')).toBe(true);
    });
  });

  describe('Custom Color Scheme without Custom Color (Neutral Default)', () => {
    it('should use neutral default color #E0E0E0 when colorScheme is "custom" and customColor is null', () => {
      const { container, getNode } = renderNode({ colorScheme: 'custom', componentId: 'comp-2', nodeData: createComponentNodeData(),
        components: [{ componentId: 'comp-2', customColor: undefined }] });
      expect(containsColor(getNode(container).style.background, '#E0E0E0')).toBe(true);
    });

    it('should use neutral default for border when colorScheme is "custom" and customColor is null', () => {
      const { container, getNode } = renderNode({ colorScheme: 'custom', componentId: 'comp-2', nodeData: createComponentNodeData(false),
        components: [{ componentId: 'comp-2', customColor: undefined }] });
      expect(colorMatches(getNode(container).style.borderColor, '#E0E0E0')).toBe(true);
    });

    it('should use neutral default when colorScheme is "custom" and customColor is undefined', () => {
      const { container, getNode } = renderNode({ colorScheme: 'custom', componentId: 'comp-3', nodeData: createComponentNodeData(),
        components: [{ componentId: 'comp-3', customColor: undefined }] });
      expect(containsColor(getNode(container).style.background, '#E0E0E0')).toBe(true);
    });
  });

  describe('Non-Custom Color Schemes Ignore Custom Colors', () => {
    it.each(['maturity', 'classic'] as const)('should ignore customColor when colorScheme is "%s"', (scheme) => {
      const { container, getNode } = renderNode({ colorScheme: scheme, componentId: 'comp-1', nodeData: createComponentNodeData(),
        components: [{ componentId: 'comp-1', customColor: '#FF5733' }] });
      const node = getNode(container);
      expect(node.style.background).toBeTruthy();
      expect(containsColor(node.style.background, '#FF5733')).toBe(false);
    });
  });

  describe('Default Color Scheme Behavior', () => {
    it('should apply scheme-based color when colorScheme is undefined', () => {
      const mockView = createMockView('maturity', [
        { componentId: 'comp-1', customColor: '#FF5733' },
      ]);
      mockView.colorScheme = undefined;
      mockCurrentView(mockView);

      const { container } = renderWithProvider(
        <ComponentNode data={createComponentNodeData()} id="comp-1" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toBeTruthy();
      expect(containsColor(node.style.background, '#FF5733')).toBe(false);
    });

    it('should apply scheme-based color when currentView is null', () => {
      mockCurrentView(null);

      const { container } = renderWithProvider(
        <ComponentNode data={createComponentNodeData()} id="comp-1" />
      );

      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node.style.background).toBeTruthy();
    });
  });

  describe('Color Reactivity and Dynamic Updates', () => {
    it('should update color when customColor changes in custom scheme', () => {
      const { mockView, container, rerender, getNode } = renderNode({ colorScheme: 'custom', componentId: 'comp-1', nodeData: createComponentNodeData(false, '#FF5733') });

      const initialBackground = getNode(container).style.background;
      expect(containsColor(initialBackground, '#FF5733')).toBe(true);

      mockView.components[0].customColor = '#33AAFF';
      rerender(<ComponentNode data={createComponentNodeData(false, '#33AAFF')} id="comp-1" />);

      const node = getNode(container);
      expect(containsColor(node.style.background, '#33AAFF')).toBe(true);
      expect(node.style.background).not.toBe(initialBackground);
    });

    it('should switch from custom color to scheme color when scheme changes from "custom" to "maturity"', () => {
      const { container, rerender, getNode } = renderNode({ colorScheme: 'custom', componentId: 'comp-1', nodeData: createComponentNodeData(false, '#FF5733') });
      expect(containsColor(getNode(container).style.background, '#FF5733')).toBe(true);

      mockCurrentView(createMockView('maturity', [{ componentId: 'comp-1', customColor: '#FF5733' }]));
      rerender(<ComponentNode data={createComponentNodeData(false, '#FF5733')} id="comp-1" />);

      const node = getNode(container);
      expect(node.style.background).toBeTruthy();
      expect(containsColor(node.style.background, '#FF5733')).toBe(false);
    });

    it.each([
      ['maturity', 'custom'],
      ['classic', 'custom'],
    ] as const)('should switch from scheme color to custom color when scheme changes from "%s" to "%s"', (fromScheme, toScheme) => {
      const { container, rerender, getNode } = renderNode({ colorScheme: fromScheme, componentId: 'comp-1', nodeData: createComponentNodeData(false, '#FF5733') });

      const schemeColor = getNode(container).style.background;
      expect(schemeColor).toBeTruthy();
      expect(containsColor(schemeColor, '#FF5733')).toBe(false);

      mockCurrentView(createMockView(toScheme, [{ componentId: 'comp-1', customColor: '#FF5733' }]));
      rerender(<ComponentNode data={createComponentNodeData(false, '#FF5733')} id="comp-1" />);

      const node = getNode(container);
      expect(containsColor(node.style.background, '#FF5733')).toBe(true);
      expect(node.style.background).not.toBe(schemeColor);
    });

    it('should update to neutral default when custom color is removed in custom scheme', () => {
      const { container, rerender, getNode } = renderNode({ colorScheme: 'custom', componentId: 'comp-1', nodeData: createComponentNodeData(false, '#FF5733'),
        components: [{ componentId: 'comp-1', customColor: '#FF5733' }, { componentId: 'comp-2', customColor: undefined }] });

      expect(containsColor(getNode(container).style.background, '#FF5733')).toBe(true);

      rerender(<ComponentNode data={createComponentNodeData(false, undefined)} id="comp-2" />);

      const node = getNode(container);
      expect(containsColor(node.style.background, '#E0E0E0')).toBe(true);
      expect(containsColor(node.style.background, '#FF5733')).toBe(false);
    });
  });

  describe('Border Color with Selection State', () => {
    it('should use selected border color when element is selected, regardless of color scheme', () => {
      const { container, getNode } = renderNode({ colorScheme: 'custom', componentId: 'comp-1', nodeData: createComponentNodeData(true),
        components: [{ componentId: 'comp-1', customColor: '#FF5733' }] });
      expect(colorMatches(getNode(container).style.borderColor, '#374151')).toBe(true);
    });

    it('should use custom color for border when element is not selected in custom scheme', () => {
      const { container, getNode } = renderNode({ colorScheme: 'custom', componentId: 'comp-1', nodeData: createComponentNodeData(false, '#FF5733') });
      const node = getNode(container);
      expect(colorMatches(node.style.borderColor, '#FF5733')).toBe(true);
      expect(colorMatches(node.style.borderColor, '#374151')).toBe(false);
    });

    it('should use scheme color for border when element is not selected in non-custom scheme', () => {
      const { container, getNode } = renderNode({ colorScheme: 'maturity', componentId: 'comp-1', nodeData: createComponentNodeData(false),
        components: [{ componentId: 'comp-1', customColor: '#FF5733' }] });
      const node = getNode(container);
      expect(node.style.borderColor).toBeTruthy();
      expect(colorMatches(node.style.borderColor, '#FF5733')).toBe(false);
      expect(colorMatches(node.style.borderColor, '#374151')).toBe(false);
    });
  });

  describe('Edge Cases', () => {
    it('should handle empty string customColor as null in custom scheme', () => {
      const { container, getNode } = renderNode({ colorScheme: 'custom', componentId: 'comp-4', nodeData: createComponentNodeData(),
        components: [{ componentId: 'comp-4', customColor: '' }] });
      expect(containsColor(getNode(container).style.background, '#E0E0E0')).toBe(true);
    });

    it('should handle component not in view with default color in custom scheme', () => {
      const { container, getNode } = renderNode({ colorScheme: 'custom', componentId: 'comp-999', nodeData: createComponentNodeData(),
        components: [{ componentId: 'comp-1', customColor: '#FF5733' }] });
      expect(containsColor(getNode(container).style.background, '#E0E0E0')).toBe(true);
    });

    it('should handle multiple components with different custom colors in same view', () => {
      const components = [
        { componentId: 'comp-1', customColor: '#FF5733' },
        { componentId: 'comp-2', customColor: '#33FF57' },
        { componentId: 'comp-3', customColor: '#3357FF' },
      ];
      const mockView = createMockView('custom', components);
      mockCurrentView(mockView);

      for (const { componentId, customColor } of components) {
        const { container } = renderWithProvider(
          <ComponentNode data={createComponentNodeData(false, customColor)} id={componentId} />
        );
        const node = container.querySelector('.component-node') as HTMLElement;
        expect(containsColor(node.style.background, customColor as HexColor)).toBe(true);
      }
    });
  });
});
