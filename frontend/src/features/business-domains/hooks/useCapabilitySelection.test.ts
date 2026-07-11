import { act, renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Capability, CapabilityId } from '../../../api/types';
import { useCapabilitySelection } from './useCapabilitySelection';

describe('useCapabilitySelection', () => {
  const createCapability = (id: string, name: string, level: 'L1' | 'L2'): Capability => ({
    id: id as CapabilityId,
    name,
    level,
    createdAt: '2024-01-01',
    _links: { self: { href: `/api/v1/capabilities/${id}`, method: 'GET' } },
  });

  const mockCapabilities: Capability[] = [
    createCapability('l1-1', 'Finance', 'L1'),
    createCapability('l1-2', 'HR', 'L1'),
    createCapability('l2-1', 'Accounting', 'L2'),
  ];

  const clickEvent = (shiftKey: boolean): React.MouseEvent =>
    ({
      shiftKey,
      preventDefault: vi.fn(),
      stopPropagation: vi.fn(),
    }) as unknown as React.MouseEvent;

  const modifierClickEvent = (mod: 'ctrlKey' | 'metaKey'): React.MouseEvent =>
    ({
      [mod]: true,
      preventDefault: vi.fn(),
      stopPropagation: vi.fn(),
    }) as unknown as React.MouseEvent;

  const renderSelection = () => {
    const onRegularClick = vi.fn();
    const { result } = renderHook(() => useCapabilitySelection(mockCapabilities, onRegularClick));
    return { onRegularClick, result };
  };

  const shiftClick = (result: ReturnType<typeof renderSelection>['result'], ...capabilities: Capability[]) => {
    act(() => {
      for (const capability of capabilities) {
        result.current.handleCapabilityClick(capability, clickEvent(true));
      }
    });
  };

  const expectSelected = (result: ReturnType<typeof renderSelection>['result'], id: string, selected: boolean) => {
    expect(result.current.selectedCapabilities.has(id as CapabilityId)).toBe(selected);
  };

  it('starts with empty selection', () => {
    const { result } = renderSelection();

    expect(result.current.selectedCapabilities.size).toBe(0);
  });

  it('calls onRegularClick on normal click (no shift)', () => {
    const { onRegularClick, result } = renderSelection();

    act(() => {
      result.current.handleCapabilityClick(mockCapabilities[0], clickEvent(false));
    });

    expect(onRegularClick).toHaveBeenCalledWith(mockCapabilities[0]);
    expect(result.current.selectedCapabilities.size).toBe(0);
  });

  it('toggles selection on shift-click', () => {
    const { result } = renderSelection();

    shiftClick(result, mockCapabilities[0]);
    expectSelected(result, 'l1-1', true);

    shiftClick(result, mockCapabilities[0]);
    expectSelected(result, 'l1-1', false);
  });

  it('allows multi-selection with shift-click', () => {
    const { result } = renderSelection();

    shiftClick(result, mockCapabilities[0], mockCapabilities[1]);

    expect(result.current.selectedCapabilities.size).toBe(2);
    expectSelected(result, 'l1-1', true);
    expectSelected(result, 'l1-2', true);
  });

  it('clears selection on normal click', () => {
    const { result } = renderSelection();

    shiftClick(result, mockCapabilities[0], mockCapabilities[1]);
    expect(result.current.selectedCapabilities.size).toBe(2);

    act(() => {
      result.current.handleCapabilityClick(mockCapabilities[0], clickEvent(false));
    });
    expect(result.current.selectedCapabilities.size).toBe(0);
  });

  it('toggles selection on ctrl-click and cmd-click, matching the Domain Board multi-select gesture', () => {
    const { result } = renderSelection();

    act(() => {
      result.current.handleCapabilityClick(mockCapabilities[0], modifierClickEvent('ctrlKey'));
    });
    expectSelected(result, 'l1-1', true);

    act(() => {
      result.current.handleCapabilityClick(mockCapabilities[1], modifierClickEvent('metaKey'));
    });
    expectSelected(result, 'l1-1', true);
    expectSelected(result, 'l1-2', true);
  });

  it('selectAllL1Capabilities selects only L1 capabilities', () => {
    const { result } = renderSelection();

    act(() => {
      result.current.selectAllL1Capabilities();
    });

    expect(result.current.selectedCapabilities.size).toBe(2);
    expectSelected(result, 'l1-1', true);
    expectSelected(result, 'l1-2', true);
    expectSelected(result, 'l2-1', false);
  });

  it('clearSelection clears all selections', () => {
    const { result } = renderSelection();

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
