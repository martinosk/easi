import type { Node } from '@xyflow/react';
import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { View } from '../../../api/types';
import { useAppStore } from '../../../store/appStore';
import { useAutoLayout } from './useAutoLayout';

const mocks = vi.hoisted(() => ({
  nodes: [] as Node[],
  fitView: vi.fn(),
  updateMultiplePositions: vi.fn(),
  updateCapabilityPosition: vi.fn(),
  updateOriginEntityPosition: vi.fn(),
}));

vi.mock('@xyflow/react', () => ({
  useReactFlow: () => ({
    getNodes: () => mocks.nodes,
    getEdges: () => [],
    getViewport: () => ({ x: 0, y: 0, zoom: 1 }),
    fitView: mocks.fitView,
  }),
}));

vi.mock('../../../utils/autoLayout', () => ({
  calculateAutoLayout: (nodes: Node[]) => nodes,
}));

const editableView = {
  id: 'v1',
  name: 'Test',
  components: [],
  capabilities: [],
  originEntities: [],
  _links: {
    self: { href: '/api/v1/views/v1', method: 'GET' },
    edit: { href: '/api/v1/views/v1', method: 'PATCH' },
  },
} as unknown as View;

vi.mock('../../views/hooks/useCurrentView', () => ({
  useCurrentView: () => ({ currentView: editableView, currentViewId: 'v1' }),
}));

vi.mock('../../views/hooks/useViews', () => ({
  useUpdateMultiplePositions: () => ({ mutateAsync: mocks.updateMultiplePositions }),
  useUpdateCapabilityPosition: () => ({ mutateAsync: mocks.updateCapabilityPosition }),
  useUpdateOriginEntityPosition: () => ({ mutateAsync: mocks.updateOriginEntityPosition }),
}));

function node(id: string, type: string, x: number, y: number): Node {
  return { id, type, position: { x, y }, data: {} };
}

describe('useAutoLayout', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mocks.updateMultiplePositions.mockResolvedValue(undefined);
    mocks.updateCapabilityPosition.mockResolvedValue(undefined);
    mocks.updateOriginEntityPosition.mockResolvedValue(undefined);
    mocks.nodes = [
      node('comp-1', 'component', 1, 2),
      node('cap-cap-1', 'capability', 3, 4),
      node('acq-oe-1', 'originEntity', 9, 10),
    ];
    act(() => {
      useAppStore.setState({
        dynamicViewId: null,
        dynamicEntities: [],
        dynamicPositions: {},
        dynamicOriginal: null,
        draftsByView: {},
      });
    });
  });

  it('applies positions to the draft when a draft is active', async () => {
    act(() => {
      useAppStore.setState({ dynamicViewId: 'v1' });
    });
    const { result } = renderHook(() => useAutoLayout());

    await act(() => result.current.applyAutoLayout());

    const positions = useAppStore.getState().dynamicPositions;
    expect(positions['comp-1']).toEqual({ x: 1, y: 2 });
    expect(positions['cap-1']).toEqual({ x: 3, y: 4 });
    expect(positions['oe-1']).toEqual({ x: 9, y: 10 });
    expect(mocks.updateMultiplePositions).not.toHaveBeenCalled();
    expect(mocks.updateCapabilityPosition).not.toHaveBeenCalled();
    expect(mocks.updateOriginEntityPosition).not.toHaveBeenCalled();
  });

  it('writes positions to the view when no draft is active', async () => {
    const { result } = renderHook(() => useAutoLayout());

    await act(() => result.current.applyAutoLayout());

    expect(mocks.updateMultiplePositions).toHaveBeenCalledWith({
      viewId: 'v1',
      request: { positions: [{ componentId: 'comp-1', x: 1, y: 2 }] },
    });
    expect(mocks.updateCapabilityPosition).toHaveBeenCalledWith({
      viewId: 'v1',
      capabilityId: 'cap-1',
      position: { x: 3, y: 4 },
    });
    expect(mocks.updateOriginEntityPosition).toHaveBeenCalledWith({
      viewId: 'v1',
      originEntityId: 'oe-1',
      position: { x: 9, y: 10 },
    });
  });

  it('skips the component batch call when auto-layout moves no components', async () => {
    mocks.nodes = [node('cap-cap-1', 'capability', 3, 4)];
    const { result } = renderHook(() => useAutoLayout());

    await act(() => result.current.applyAutoLayout());

    expect(mocks.updateMultiplePositions).not.toHaveBeenCalled();
    expect(mocks.updateCapabilityPosition).toHaveBeenCalledTimes(1);
  });
});
