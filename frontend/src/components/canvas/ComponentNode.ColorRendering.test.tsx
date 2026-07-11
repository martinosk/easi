import { render } from '@testing-library/react';
import { ReactFlowProvider } from '@xyflow/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('../../features/views/hooks/useCurrentView', () => ({
  useCurrentView: vi.fn(),
}));

import type { View } from '../../api/types';
import { toViewId } from '../../api/types';
import { DEFAULT_CUSTOM_COLOR } from '../../constants/maturityColors';
import { useCurrentView } from '../../features/views/hooks/useCurrentView';
import { ComponentNode, type ComponentNodeData } from './ComponentNode';
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

const createComponentNodeData = (overrides: Partial<ComponentNodeData> = {}): ComponentNodeData => ({
  label: 'Payment Service',
  description: 'Handles payment processing',
  isSelected: false,
  ...overrides,
});

const renderNode = (data: ComponentNodeData, id = 'comp-1') => {
  const result = render(
    <ReactFlowProvider>
      <ComponentNode data={data} id={id} />
    </ReactFlowProvider>,
  );
  const node = result.container.querySelector('.component-node') as HTMLElement;
  return { ...result, node };
};

describe('ComponentNode color rendering', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('custom scheme', () => {
    beforeEach(() => {
      mockCurrentView(createMockView('custom'));
    });

    it('fills the card with the custom colour and computes contrast-appropriate text', () => {
      const { node } = renderNode(createComponentNodeData({ customColor: '#FF5733' }));
      expect(node).toHaveClass('component-node--custom');
      expect(node.style.backgroundColor.toUpperCase()).toBe('#FF5733');
      expect(node.style.color).toBe(getContrastTextColor('#FF5733'));
    });

    it('falls back to the neutral default fill when no custom colour is set', () => {
      const { node } = renderNode(createComponentNodeData());
      expect(node.style.backgroundColor).toBeTruthy();
      expect(node.style.color).toBe(getContrastTextColor(DEFAULT_CUSTOM_COLOR));
    });

    it('falls back to the neutral default fill for an empty-string custom colour', () => {
      const { node } = renderNode(createComponentNodeData({ customColor: '' }));
      expect(node.style.color).toBe(getContrastTextColor(DEFAULT_CUSTOM_COLOR));
    });

    it('updates the fill when the custom colour changes', () => {
      const { rerender, container } = render(
        <ReactFlowProvider>
          <ComponentNode data={createComponentNodeData({ customColor: '#FF5733' })} id="comp-1" />
        </ReactFlowProvider>,
      );
      const initialColor = (container.querySelector('.component-node') as HTMLElement).style.backgroundColor;

      rerender(
        <ReactFlowProvider>
          <ComponentNode data={createComponentNodeData({ customColor: '#33AAFF' })} id="comp-1" />
        </ReactFlowProvider>,
      );
      const updatedColor = (container.querySelector('.component-node') as HTMLElement).style.backgroundColor;

      expect(updatedColor).not.toBe(initialColor);
    });
  });

  describe.each(['maturity', 'classic'] as const)('%s scheme', (scheme) => {
    beforeEach(() => {
      mockCurrentView(createMockView(scheme));
    });

    it('applies no inline fill, letting canvas.css drive the background', () => {
      const { node } = renderNode(createComponentNodeData({ customColor: '#FF5733' }));
      expect(node.style.backgroundColor).toBe('');
      expect(node.style.color).toBe('');
    });

    it(`carries the component-node--${scheme} class`, () => {
      const { node } = renderNode(createComponentNodeData());
      expect(node).toHaveClass(`component-node--${scheme}`);
    });
  });

  it('applies the classic-text class in classic scheme', () => {
    mockCurrentView(createMockView('classic'));
    const { node } = renderNode(createComponentNodeData());
    expect(node).toHaveClass('classic-text');
  });

  it('defaults to maturity scheme when colorScheme is undefined', () => {
    mockCurrentView(createMockView(undefined));
    const { node } = renderNode(createComponentNodeData());
    expect(node).toHaveClass('component-node--maturity');
  });

  it('defaults to maturity scheme when currentView is null', () => {
    mockCurrentView(null);
    const { node } = renderNode(createComponentNodeData());
    expect(node).toHaveClass('component-node--maturity');
  });

  describe('selection', () => {
    it('adds the selected class when data.isSelected is true, independent of scheme', () => {
      mockCurrentView(createMockView('maturity'));
      const { node } = renderNode(createComponentNodeData({ isSelected: true }));
      expect(node).toHaveClass('component-node-selected');
    });

    it('adds the selected class when the selected prop is true', () => {
      mockCurrentView(createMockView('custom'));
      const { container } = render(
        <ReactFlowProvider>
          <ComponentNode data={createComponentNodeData()} id="comp-1" selected />
        </ReactFlowProvider>,
      );
      const node = container.querySelector('.component-node') as HTMLElement;
      expect(node).toHaveClass('component-node-selected');
    });

    it('omits the selected class when not selected', () => {
      mockCurrentView(createMockView('maturity'));
      const { node } = renderNode(createComponentNodeData());
      expect(node).not.toHaveClass('component-node-selected');
    });
  });
});
