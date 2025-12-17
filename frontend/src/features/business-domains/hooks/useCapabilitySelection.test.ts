import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { useCapabilitySelection } from './useCapabilitySelection';
import type { Capability, CapabilityId } from '../../../api/types';

describe('useCapabilitySelection', () => {
  const createCapability = (id: string, name: string, level: 'L1' | 'L2'): Capability => ({
    id: id as CapabilityId,
    name,
    level,
    createdAt: '2024-01-01',
    _links: { self: { href: `/api/v1/capabilities/${id}` } },
  });

  const mockCapabilities: Capability[] = [
    createCapability('l1-1', 'Finance', 'L1'),
    createCapability('l1-2', 'HR', 'L1'),
    createCapability('l2-1', 'Accounting', 'L2'),
  ];

  it('starts with empty selection', () => {
    const onRegularClick = vi.fn();
    const { result } = renderHook(() => useCapabilitySelection(mockCapabilities, onRegularClick));

    expect(result.current.selectedCapabilities.size).toBe(0);
  });

  it('calls onRegularClick on normal click (no shift)', () => {
    const onRegularClick = vi.fn();
    const { result } = renderHook(() => useCapabilitySelection(mockCapabilities, onRegularClick));

    const mockEvent = { shiftKey: false, preventDefault: vi.fn(), stopPropagation: vi.fn() } as unknown as React.MouseEvent;

    act(() => {
      result.current.handleCapabilityClick(mockCapabilities[0], mockEvent);
    });

    expect(onRegularClick).toHaveBeenCalledWith(mockCapabilities[0]);
    expect(result.current.selectedCapabilities.size).toBe(0);
  });

  it('toggles selection on shift-click', () => {
    const onRegularClick = vi.fn();
    const { result } = renderHook(() => useCapabilitySelection(mockCapabilities, onRegularClick));

    const shiftEvent = { shiftKey: true, preventDefault: vi.fn(), stopPropagation: vi.fn() } as unknown as React.MouseEvent;

    act(() => {
      result.current.handleCapabilityClick(mockCapabilities[0], shiftEvent);
    });
    expect(result.current.selectedCapabilities.has('l1-1' as CapabilityId)).toBe(true);

    act(() => {
      result.current.handleCapabilityClick(mockCapabilities[0], shiftEvent);
    });
    expect(result.current.selectedCapabilities.has('l1-1' as CapabilityId)).toBe(false);
  });

  it('allows multi-selection with shift-click', () => {
    const onRegularClick = vi.fn();
    const { result } = renderHook(() => useCapabilitySelection(mockCapabilities, onRegularClick));

    const shiftEvent = { shiftKey: true, preventDefault: vi.fn(), stopPropagation: vi.fn() } as unknown as React.MouseEvent;

    act(() => {
      result.current.handleCapabilityClick(mockCapabilities[0], shiftEvent);
      result.current.handleCapabilityClick(mockCapabilities[1], shiftEvent);
    });

    expect(result.current.selectedCapabilities.size).toBe(2);
    expect(result.current.selectedCapabilities.has('l1-1' as CapabilityId)).toBe(true);
    expect(result.current.selectedCapabilities.has('l1-2' as CapabilityId)).toBe(true);
  });

  it('clears selection on normal click', () => {
    const onRegularClick = vi.fn();
    const { result } = renderHook(() => useCapabilitySelection(mockCapabilities, onRegularClick));

    const shiftEvent = { shiftKey: true, preventDefault: vi.fn(), stopPropagation: vi.fn() } as unknown as React.MouseEvent;
    const normalEvent = { shiftKey: false, preventDefault: vi.fn(), stopPropagation: vi.fn() } as unknown as React.MouseEvent;

    act(() => {
      result.current.handleCapabilityClick(mockCapabilities[0], shiftEvent);
      result.current.handleCapabilityClick(mockCapabilities[1], shiftEvent);
    });
    expect(result.current.selectedCapabilities.size).toBe(2);

    act(() => {
      result.current.handleCapabilityClick(mockCapabilities[0], normalEvent);
    });
    expect(result.current.selectedCapabilities.size).toBe(0);
  });

  it('selectAllL1Capabilities selects only L1 capabilities', () => {
    const onRegularClick = vi.fn();
    const { result } = renderHook(() => useCapabilitySelection(mockCapabilities, onRegularClick));

    act(() => {
      result.current.selectAllL1Capabilities();
    });

    expect(result.current.selectedCapabilities.size).toBe(2);
    expect(result.current.selectedCapabilities.has('l1-1' as CapabilityId)).toBe(true);
    expect(result.current.selectedCapabilities.has('l1-2' as CapabilityId)).toBe(true);
    expect(result.current.selectedCapabilities.has('l2-1' as CapabilityId)).toBe(false);
  });

  it('clearSelection clears all selections', () => {
    const onRegularClick = vi.fn();
    const { result } = renderHook(() => useCapabilitySelection(mockCapabilities, onRegularClick));

    act(() => {
      result.current.selectAllL1Capabilities();
    });
    expect(result.current.selectedCapabilities.size).toBe(2);

    act(() => {
      result.current.clearSelection();
    });
    expect(result.current.selectedCapabilities.size).toBe(0);
  });
});
