import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render } from '@testing-library/react';
import React from 'react';
import { ReactFlowProvider } from '@xyflow/react';

vi.mock('../../features/views/hooks/useCurrentView', () => ({
  useCurrentView: vi.fn(),
}));

vi.mock('../../hooks/useMaturityScale', () => ({
  useMaturityScale: vi.fn(() => ({
    data: {
      sections: [
        { name: 'Genesis', order: 1, minValue: 0, maxValue: 24 },
        { name: 'Custom Built', order: 2, minValue: 25, maxValue: 49 },
        { name: 'Product', order: 3, minValue: 50, maxValue: 74 },
        { name: 'Commodity', order: 4, minValue: 75, maxValue: 99 },
      ],
      version: 1,
      isDefault: true,
    },
    isLoading: false,
    error: null,
  })),
}));

import { CapabilityNode, type CapabilityNodeData } from './CapabilityNode';
import { useCurrentView } from '../../features/views/hooks/useCurrentView';
import type { View } from '../../api/types';
import { toViewId, toCapabilityId } from '../../api/types';

const createMockView = (colorScheme: string, capabilitiesWithColors?: Array<{ capabilityId: string; customColor?: string }>): View => ({
  id: toViewId('view-1'),
  name: 'Test View',
  description: 'Test view description',
  isDefault: true,
  isPrivate: false,
  components: [],
  capabilities: capabilitiesWithColors?.map(cap => ({
    capabilityId: toCapabilityId(cap.capabilityId),
    x: 100,
    y: 200,
    customColor: cap.customColor,
  })) || [],
  originEntities: [],
  colorScheme,
  createdAt: '2024-01-01T00:00:00Z',
  _links: { self: { href: '/api/v1/views/view-1', method: 'GET' } },
});

const createCapabilityNodeData = (
  maturityLevel?: string,
  isSelected: boolean = false,
  customColor?: string,
  maturityValue?: number
): CapabilityNodeData => ({
  label: 'Customer Management',
  level: 'L1',
  maturityLevel,
  maturityValue,
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

const containsColor = (styleValue: string, hex: string): boolean => {
  const rgbValue = hexToRgb(hex);
  const upperHex = hex.toUpperCase();
  const lowerHex = hex.toLowerCase();
  return styleValue.includes(rgbValue) || styleValue.includes(upperHex) || styleValue.includes(lowerHex);
};

const colorMatches = (styleValue: string, hex: string): boolean => {
  const rgbValue = hexToRgb(hex);
  const upperHex = hex.toUpperCase();
  const lowerHex = hex.toLowerCase();
  return styleValue === rgbValue || styleValue === upperHex || styleValue === lowerHex;
};

const defaultViewCapabilities = [{ capabilityId: 'cap-1', customColor: '#FF5733' }];

const mockCurrentView = (
  colorScheme: string | undefined,
  capabilities: Array<{ capabilityId: string; customColor?: string }> = defaultViewCapabilities,
) => {
  const mockView = createMockView(colorScheme ?? 'maturity', capabilities);
  if (colorScheme === undefined) mockView.colorScheme = undefined;
  vi.mocked(useCurrentView).mockReturnValue({
    currentView: mockView,
    currentViewId: mockView.id,
    isLoading: false,
    error: null,
  });
  return mockView;
};

interface NodeOptions {
  nodeId?: string;
  maturityLevel?: string;
  isSelected?: boolean;
  customColor?: string;
  maturityValue?: number;
}

const renderAndGetNode = (options: NodeOptions = {}) => {
  const { nodeId = 'cap-1', maturityLevel, isSelected = false, customColor, maturityValue } = options;
  const nodeData = createCapabilityNodeData(maturityLevel, isSelected, customColor, maturityValue);
  const result = renderWithProvider(
    <CapabilityNode data={nodeData} id={nodeId} />
  );
  const node = result.container.querySelector('.capability-node') as HTMLElement;
  return { node, nodeData, ...result };
};

describe('CapabilityNode Custom Color Rendering', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Custom Color Scheme with Custom Color Set', () => {
    it('should use customColor when colorScheme is "custom" and customColor is provided', () => {
      mockCurrentView('custom');
      const { node } = renderAndGetNode({ maturityLevel: 'Product', customColor: '#FF5733' });
      expect(node).toBeTruthy();
      expect(containsColor(node.style.background, '#FF5733')).toBe(true);
    });

    it('should apply custom color as gradient with opacity when colorScheme is "custom"', () => {
      mockCurrentView('custom');
      const { node } = renderAndGetNode({ maturityLevel: 'Genesis' });
      expect(node.style.background).toMatch(/linear-gradient/);
    });

    it('should use customColor for border color when colorScheme is "custom" and element is not selected', () => {
      mockCurrentView('custom', [{ capabilityId: 'cap-1', customColor: '#22AA88' }]);
      const { node } = renderAndGetNode({ maturityLevel: 'Product', customColor: '#22AA88' });
      expect(colorMatches(node.style.borderColor, '#22AA88')).toBe(true);
    });
  });

  describe('Custom Color Scheme without Custom Color (Neutral Default)', () => {
    it('should use neutral default color #E0E0E0 when colorScheme is "custom" and customColor is null', () => {
      mockCurrentView('custom', [{ capabilityId: 'cap-2', customColor: undefined }]);
      const { node } = renderAndGetNode({ nodeId: 'cap-2', maturityLevel: 'Product' });
      expect(containsColor(node.style.background, '#E0E0E0')).toBe(true);
    });

    it('should use neutral default for border when colorScheme is "custom" and customColor is null', () => {
      mockCurrentView('custom', [{ capabilityId: 'cap-2', customColor: undefined }]);
      const { node } = renderAndGetNode({ nodeId: 'cap-2', maturityLevel: 'Genesis' });
      expect(colorMatches(node.style.borderColor, '#E0E0E0')).toBe(true);
    });

    it('should use neutral default when colorScheme is "custom" and customColor is undefined', () => {
      mockCurrentView('custom', [{ capabilityId: 'cap-3', customColor: undefined }]);
      const { node } = renderAndGetNode({ nodeId: 'cap-3', maturityLevel: 'Commodity' });
      expect(containsColor(node.style.background, '#E0E0E0')).toBe(true);
    });
  });

  describe('Non-Custom Color Schemes Ignore Custom Colors', () => {
    it('should use maturity-based color when colorScheme is "maturity", ignoring customColor', () => {
      mockCurrentView('maturity');
      const { node } = renderAndGetNode({ maturityLevel: 'Product', customColor: '#FF5733', maturityValue: 62 });
      expect(node.style.background).toMatch(/linear-gradient/);
      expect(containsColor(node.style.background, '#FF5733')).toBe(false);
    });

    it.each([
      { maturityLevel: 'Genesis', maturityValue: 12, expectedColor: '#f89191' },
      { maturityLevel: 'Custom Build', maturityValue: 37, expectedColor: '#fdb774' },
      { maturityLevel: 'Commodity', maturityValue: 87, expectedColor: '#5befb1' },
    ])('should use maturity color for $maturityLevel when colorScheme is "maturity"', ({ maturityLevel, maturityValue, expectedColor }) => {
      mockCurrentView('maturity');
      const { node } = renderAndGetNode({ maturityLevel, maturityValue });
      expect(node.style.background).toMatch(/linear-gradient/);
      expect(containsColor(node.style.background, expectedColor)).toBe(true);
    });
  });

  describe('Default Color Scheme Behavior', () => {
    it('should use maturity color when colorScheme is undefined', () => {
      mockCurrentView(undefined);
      const { node } = renderAndGetNode({ maturityLevel: 'Product', maturityValue: 62 });
      expect(node.style.background).toMatch(/linear-gradient/);
    });

    it('should use maturity color when currentView is null', () => {
      vi.mocked(useCurrentView).mockReturnValue({
        currentView: null,
        currentViewId: null,
        isLoading: false,
        error: null,
      });
      const { node } = renderAndGetNode({ maturityLevel: 'Commodity', maturityValue: 87 });
      expect(node.style.background).toMatch(/linear-gradient/);
    });
  });

  describe('Color Reactivity and Dynamic Updates', () => {
    it('should update color when customColor changes in custom scheme', () => {
      const view = mockCurrentView('custom');
      const { container, rerender } = renderAndGetNode({ maturityLevel: 'Product', customColor: '#FF5733' });

      let node = container.querySelector('.capability-node') as HTMLElement;
      const initialBackground = node.style.background;
      expect(containsColor(initialBackground, '#FF5733')).toBe(true);

      view.capabilities[0].customColor = '#33AAFF';
      const updatedNodeData = createCapabilityNodeData('Product', false, '#33AAFF');
      rerender(<CapabilityNode data={updatedNodeData} id="cap-1" />);

      node = container.querySelector('.capability-node') as HTMLElement;
      expect(containsColor(node.style.background, '#33AAFF')).toBe(true);
      expect(node.style.background).not.toBe(initialBackground);
    });

    it('should switch from custom color to maturity color when scheme changes from "custom" to "maturity"', () => {
      mockCurrentView('custom');
      const { container, rerender } = renderAndGetNode({ maturityLevel: 'Product', customColor: '#FF5733', maturityValue: 62 });

      let node = container.querySelector('.capability-node') as HTMLElement;
      expect(containsColor(node.style.background, '#FF5733')).toBe(true);

      mockCurrentView('maturity');
      const updatedNodeData = createCapabilityNodeData('Product', false, '#FF5733', 62);
      rerender(<CapabilityNode data={updatedNodeData} id="cap-1" />);

      node = container.querySelector('.capability-node') as HTMLElement;
      expect(node.style.background).toMatch(/linear-gradient/);
      expect(containsColor(node.style.background, '#FF5733')).toBe(false);
    });

    it('should switch from maturity color to custom color when scheme changes from "maturity" to "custom"', () => {
      mockCurrentView('maturity');
      const { container, rerender } = renderAndGetNode({ maturityLevel: 'Genesis', customColor: '#FF5733', maturityValue: 12 });

      let node = container.querySelector('.capability-node') as HTMLElement;
      expect(node.style.background).toMatch(/linear-gradient/);

      mockCurrentView('custom');
      const updatedNodeData = createCapabilityNodeData('Genesis', false, '#FF5733', 12);
      rerender(<CapabilityNode data={updatedNodeData} id="cap-1" />);

      node = container.querySelector('.capability-node') as HTMLElement;
      expect(containsColor(node.style.background, '#FF5733')).toBe(true);
    });

    it('should update to neutral default when custom color is removed in custom scheme', () => {
      mockCurrentView('custom', [
        { capabilityId: 'cap-1', customColor: '#FF5733' },
        { capabilityId: 'cap-2', customColor: undefined },
      ]);
      const { container, rerender } = renderAndGetNode({ maturityLevel: 'Product', customColor: '#FF5733' });

      let node = container.querySelector('.capability-node') as HTMLElement;
      expect(containsColor(node.style.background, '#FF5733')).toBe(true);

      const updatedNodeData = createCapabilityNodeData('Product', false, undefined);
      rerender(<CapabilityNode data={updatedNodeData} id="cap-2" />);

      node = container.querySelector('.capability-node') as HTMLElement;
      expect(containsColor(node.style.background, '#E0E0E0')).toBe(true);
      expect(containsColor(node.style.background, '#FF5733')).toBe(false);
    });
  });

  describe('Border Color with Selection State', () => {
    it('should use selected border color when element is selected, regardless of color scheme', () => {
      mockCurrentView('custom');
      const { node } = renderAndGetNode({ maturityLevel: 'Product', isSelected: true });
      expect(colorMatches(node.style.borderColor, '#374151')).toBe(true);
    });

    it('should use custom color for border when element is not selected in custom scheme', () => {
      mockCurrentView('custom');
      const { node } = renderAndGetNode({ maturityLevel: 'Genesis', customColor: '#FF5733' });
      expect(colorMatches(node.style.borderColor, '#FF5733')).toBe(true);
      expect(colorMatches(node.style.borderColor, '#374151')).toBe(false);
    });

    it('should use maturity color for border when element is not selected in maturity scheme', () => {
      mockCurrentView('maturity');
      const { node } = renderAndGetNode({ maturityLevel: 'Commodity', maturityValue: 87 });
      expect(node.style.borderColor).toBeTruthy();
    });
  });

  describe('Edge Cases', () => {
    it.each([
      { maturityLevel: 'UnknownLevel', maturityValue: 0 },
      { maturityLevel: undefined, maturityValue: undefined },
    ])('should handle $maturityLevel maturity level with gradient fallback', ({ maturityLevel, maturityValue }) => {
      mockCurrentView('maturity');
      const { node } = renderAndGetNode({ maturityLevel, maturityValue });
      expect(node.style.background).toMatch(/linear-gradient/);
    });

    it('should handle empty string customColor as null in custom scheme', () => {
      mockCurrentView('custom', [{ capabilityId: 'cap-4', customColor: '' }]);
      const { node } = renderAndGetNode({ nodeId: 'cap-4', maturityLevel: 'Product' });
      expect(containsColor(node.style.background, '#E0E0E0')).toBe(true);
    });
  });
});
