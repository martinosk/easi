import { render } from '@testing-library/react';
import { ReactFlowProvider } from '@xyflow/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../../features/views/hooks/useCurrentView', () => ({
  useCurrentView: vi.fn(),
}));

const getColorForValue = vi.fn((value: number) => (value >= 50 ? '#CFE8DA' : '#F6D9D5'));
const getSectionNameForValue = vi.fn((value: number) => (value >= 50 ? 'Product' : 'Genesis'));

vi.mock('../../hooks/useMaturityColorScale', () => ({
  useMaturityColorScale: () => ({
    getColorForValue,
    getSectionNameForValue,
    getBaseSectionColor: vi.fn(),
    bounds: { min: 0, max: 99 },
  }),
}));

import type { View } from '../../api/types';
import { toViewId } from '../../api/types';
import { DEFAULT_CUSTOM_COLOR } from '../../constants/maturityColors';
import { useCurrentView } from '../../features/views/hooks/useCurrentView';
import { CapabilityNode, type CapabilityNodeData } from './CapabilityNode';
import { getContrastTextColor } from './contrastText';

const createMockView = (colorScheme: string | undefined): View => ({
  id: toViewId('view-1'),
  name: 'Test View',
  description: 'Test view description',
  isDefault: true,
  isPrivate: false,
  components: [],
  capabilities: [],
  originEntities: [],
  colorScheme,
  createdAt: '2024-01-01T00:00:00Z',
  _links: { self: { href: '/api/v1/views/view-1', method: 'GET' } },
});

const mockCurrentView = (view: View | null) => {
  vi.mocked(useCurrentView).mockReturnValue({
    currentView: view,
    currentViewId: view?.id ?? null,
    isLoading: false,
    error: null,
  });
};

const createCapabilityNodeData = (overrides: Partial<CapabilityNodeData> = {}): CapabilityNodeData => ({
  label: 'Customer Management',
  level: 'L1',
  isSelected: false,
  ...overrides,
});

const renderNode = (data: CapabilityNodeData, id = 'cap-1') => {
  const result = render(
    <ReactFlowProvider>
      <CapabilityNode data={data} id={id} />
    </ReactFlowProvider>,
  );
  const node = result.container.querySelector('.capability-node') as HTMLElement;
  return { ...result, node };
};

describe('CapabilityNode color rendering', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('maturity scheme (default)', () => {
    it('fills the card with the heat-mapped colour from useMaturityColorScale', () => {
      mockCurrentView(createMockView('maturity'));
      const { node } = renderNode(createCapabilityNodeData({ maturityValue: 87 }));
      expect(node).toHaveClass('capability-node--maturity');
      expect(node.style.backgroundColor).toBeTruthy();
      expect(node.style.color).toBe(getContrastTextColor('#CFE8DA'));
      expect(getColorForValue).toHaveBeenCalledWith(87);
    });

    it('ignores customColor', () => {
      mockCurrentView(createMockView('maturity'));
      const { node } = renderNode(createCapabilityNodeData({ maturityValue: 12, customColor: '#FF5733' }));
      expect(node.style.backgroundColor).not.toContain('255, 87, 51');
    });

    it('derives the maturity value from maturityLevel when maturityValue is absent', () => {
      mockCurrentView(createMockView(undefined));
      renderNode(createCapabilityNodeData({ maturityLevel: 'Commodity' }));
      expect(getColorForValue).toHaveBeenCalledWith(87);
    });
  });

  describe('custom scheme', () => {
    beforeEach(() => {
      mockCurrentView(createMockView('custom'));
    });

    it('fills the card with the custom colour and computes contrast-appropriate text', () => {
      const { node } = renderNode(createCapabilityNodeData({ customColor: '#22AA88' }));
      expect(node).toHaveClass('capability-node--custom');
      expect(node.style.backgroundColor.toUpperCase()).toBe('#22AA88');
      expect(node.style.color).toBe(getContrastTextColor('#22AA88'));
    });

    it('falls back to the neutral default fill when no custom colour is set', () => {
      const { node } = renderNode(createCapabilityNodeData());
      expect(node.style.color).toBe(getContrastTextColor(DEFAULT_CUSTOM_COLOR));
    });

    it('falls back to the neutral default fill for an empty-string custom colour', () => {
      const { node } = renderNode(createCapabilityNodeData({ customColor: '' }));
      expect(node.style.color).toBe(getContrastTextColor(DEFAULT_CUSTOM_COLOR));
    });
  });

  describe('classic scheme', () => {
    it('applies no inline fill and carries the classic classes', () => {
      mockCurrentView(createMockView('classic'));
      const { node } = renderNode(createCapabilityNodeData({ customColor: '#FF5733' }));
      expect(node).toHaveClass('capability-node--classic');
      expect(node).toHaveClass('classic-text');
      expect(node.style.backgroundColor).toBe('');
      expect(node.style.color).toBe('');
    });
  });

  it('renders the level tag and section name', () => {
    mockCurrentView(createMockView('maturity'));
    const { container } = renderNode(createCapabilityNodeData({ maturityValue: 12 }));
    expect(container.querySelector('.capability-node-level')?.textContent).toBe('L1:');
    expect(container.querySelector('.capability-node-maturity')?.textContent).toBe('Genesis');
  });

  describe('selection', () => {
    it('adds the selected class regardless of colour scheme', () => {
      mockCurrentView(createMockView('custom'));
      const { node } = renderNode(createCapabilityNodeData({ isSelected: true }));
      expect(node).toHaveClass('capability-node-selected');
    });

    it('omits the selected class when not selected', () => {
      mockCurrentView(createMockView('maturity'));
      const { node } = renderNode(createCapabilityNodeData());
      expect(node).not.toHaveClass('capability-node-selected');
    });
  });

  describe('reactivity', () => {
    it('switches from custom fill to heat-mapped fill when scheme changes from custom to maturity', () => {
      mockCurrentView(createMockView('custom'));
      const { container, rerender } = render(
        <ReactFlowProvider>
          <CapabilityNode data={createCapabilityNodeData({ customColor: '#FF5733' })} id="cap-1" />
        </ReactFlowProvider>,
      );
      const customFill = (container.querySelector('.capability-node') as HTMLElement).style.backgroundColor;

      mockCurrentView(createMockView('maturity'));
      rerender(
        <ReactFlowProvider>
          <CapabilityNode data={createCapabilityNodeData({ customColor: '#FF5733', maturityValue: 87 })} id="cap-1" />
        </ReactFlowProvider>,
      );
      const maturityFill = (container.querySelector('.capability-node') as HTMLElement).style.backgroundColor;

      expect(maturityFill).not.toBe(customFill);
    });
  });
});
