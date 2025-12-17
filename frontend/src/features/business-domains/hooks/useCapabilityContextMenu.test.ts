import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { useCapabilityContextMenu } from './useCapabilityContextMenu';
import type { Capability, CapabilityId } from '../../../api/types';

describe('useCapabilityContextMenu', () => {
  const createCapability = (id: string, name: string, level: 'L1' | 'L2', parentId?: string): Capability => ({
    id: id as CapabilityId,
    name,
    level,
    parentId: parentId as CapabilityId | undefined,
    createdAt: '2024-01-01',
    _links: { self: { href: `/api/v1/capabilities/${id}` } },
  });

  const mockCapabilities: Capability[] = [
    createCapability('l1-1', 'Finance', 'L1'),
    createCapability('l2-1', 'Accounting', 'L2', 'l1-1'),
  ];

  const mockDomainCapabilities: Capability[] = [
    { ...createCapability('l1-1', 'Finance', 'L1'), _links: { self: { href: '/api/v1/capabilities/l1-1' }, dissociate: '/api/v1/dissociate/l1-1' } },
  ];

  const defaultProps = {
    capabilities: mockCapabilities,
    domainCapabilities: mockDomainCapabilities,
    dissociateCapability: vi.fn().mockResolvedValue(undefined),
    refetch: vi.fn().mockResolvedValue(undefined),
    selectedCapabilities: new Set<CapabilityId>(),
    setSelectedCapabilities: vi.fn(),
  };

  it('opens context menu at click position', () => {
    const { result } = renderHook(() => useCapabilityContextMenu(defaultProps));

    const mockEvent = {
      preventDefault: vi.fn(),
      clientX: 100,
      clientY: 200,
    } as unknown as React.MouseEvent;

    act(() => {
      result.current.handleCapabilityContextMenu(mockCapabilities[0], mockEvent);
    });

    expect(result.current.contextMenu).toEqual({
      x: 100,
      y: 200,
      capability: mockCapabilities[0],
    });
  });

  it('closes context menu', () => {
    const { result } = renderHook(() => useCapabilityContextMenu(defaultProps));

    const mockEvent = { preventDefault: vi.fn(), clientX: 100, clientY: 200 } as unknown as React.MouseEvent;

    act(() => {
      result.current.handleCapabilityContextMenu(mockCapabilities[0], mockEvent);
    });
    expect(result.current.contextMenu).not.toBeNull();

    act(() => {
      result.current.closeContextMenu();
    });
    expect(result.current.contextMenu).toBeNull();
  });

  it('provides two menu items: remove and delete', () => {
    const { result } = renderHook(() => useCapabilityContextMenu(defaultProps));

    expect(result.current.contextMenuItems).toHaveLength(2);
    expect(result.current.contextMenuItems[0].label).toBe('Remove from Business Domain');
    expect(result.current.contextMenuItems[1].label).toBe('Delete from Model');
  });

  it('resolves L1 ancestor when removing L2 capability', async () => {
    const dissociateCapability = vi.fn().mockResolvedValue(undefined);
    const { result } = renderHook(() => useCapabilityContextMenu({
      ...defaultProps,
      dissociateCapability,
    }));

    const mockEvent = { preventDefault: vi.fn(), clientX: 100, clientY: 200 } as unknown as React.MouseEvent;

    act(() => {
      result.current.handleCapabilityContextMenu(mockCapabilities[1], mockEvent); // L2 capability
    });

    await act(async () => {
      await result.current.contextMenuItems[0].onClick(); // Remove from Business Domain
    });

    expect(dissociateCapability).toHaveBeenCalledWith(
      expect.objectContaining({ id: 'l1-1', level: 'L1', _links: expect.objectContaining({ dissociate: '/api/v1/dissociate/l1-1' }) })
    );
  });

  it('sets capability to delete when clicking delete option', () => {
    const { result } = renderHook(() => useCapabilityContextMenu(defaultProps));

    const mockEvent = { preventDefault: vi.fn(), clientX: 100, clientY: 200 } as unknown as React.MouseEvent;

    act(() => {
      result.current.handleCapabilityContextMenu(mockCapabilities[0], mockEvent);
    });

    act(() => {
      result.current.contextMenuItems[1].onClick(); // Delete from Model
    });

    expect(result.current.capabilityToDelete).toEqual(
      expect.objectContaining({ id: 'l1-1' })
    );
  });
});
