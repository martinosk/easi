import { act, renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Capability, CapabilityId } from '../../../api/types';
import { useCapabilityContextMenu } from './useCapabilityContextMenu';

describe('useCapabilityContextMenu', () => {
  const createCapability = (id: string, name: string, level: 'L1' | 'L2', parentId?: string): Capability => ({
    id: id as CapabilityId,
    name,
    level,
    parentId: parentId as CapabilityId | undefined,
    createdAt: '2024-01-01',
    _links: { self: { href: `/api/v1/capabilities/${id}`, method: 'GET' } },
  });

  const mockCapabilities: Capability[] = [
    {
      ...createCapability('l1-1', 'Finance', 'L1'),
      _links: {
        self: { href: '/api/v1/capabilities/l1-1', method: 'GET' },
        delete: { href: '/api/v1/capabilities/l1-1', method: 'DELETE' },
      },
    },
    {
      ...createCapability('l2-1', 'Accounting', 'L2', 'l1-1'),
      _links: {
        self: { href: '/api/v1/capabilities/l2-1', method: 'GET' },
        delete: { href: '/api/v1/capabilities/l2-1', method: 'DELETE' },
      },
    },
  ];

  const mockDomainCapabilities: Capability[] = [
    {
      ...createCapability('l1-1', 'Finance', 'L1'),
      _links: {
        self: { href: '/api/v1/capabilities/l1-1', method: 'GET' },
        'x-remove-from-domain': { href: '/api/v1/business-domains/domain-1/capabilities/l1-1', method: 'DELETE' },
      },
    },
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

  it('provides two menu items: remove and delete when allowed', () => {
    const { result } = renderHook(() => useCapabilityContextMenu(defaultProps));

    const mockEvent = { preventDefault: vi.fn(), clientX: 100, clientY: 200 } as unknown as React.MouseEvent;

    act(() => {
      result.current.handleCapabilityContextMenu(mockCapabilities[0], mockEvent);
    });

    expect(result.current.contextMenuItems).toHaveLength(2);
    expect(result.current.contextMenuItems[0].label).toBe('Remove from Business Domain');
    expect(result.current.contextMenuItems[1].label).toBe('Delete from Model');
  });

  it('hides Remove from Business Domain for capabilities without the link', () => {
    const { result } = renderHook(() => useCapabilityContextMenu(defaultProps));

    const mockEvent = { preventDefault: vi.fn(), clientX: 100, clientY: 200 } as unknown as React.MouseEvent;

    act(() => {
      result.current.handleCapabilityContextMenu(mockCapabilities[1], mockEvent); // L2 capability
    });

    const labels = result.current.contextMenuItems.map((item) => item.label);
    expect(labels).not.toContain('Remove from Business Domain');
    expect(labels).toContain('Delete from Model');
  });

  it('deletes the clicked capability itself, not its L1 ancestor', () => {
    const { result } = renderHook(() => useCapabilityContextMenu(defaultProps));

    const mockEvent = { preventDefault: vi.fn(), clientX: 100, clientY: 200 } as unknown as React.MouseEvent;

    act(() => {
      result.current.handleCapabilityContextMenu(mockCapabilities[1], mockEvent); // L2 capability
    });

    const deleteItem = result.current.contextMenuItems.find((item) => item.label === 'Delete from Model');
    act(() => {
      deleteItem?.onClick();
    });

    expect(result.current.capabilityToDelete).toEqual(expect.objectContaining({ id: 'l2-1', level: 'L2' }));
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

    expect(result.current.capabilityToDelete).toEqual(expect.objectContaining({ id: 'l1-1' }));
  });
});
