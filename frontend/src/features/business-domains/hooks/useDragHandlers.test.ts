import { act, renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { BusinessDomainId, Capability, CapabilityId } from '../../../api/types';
import { useDragHandlers } from './useDragHandlers';

vi.mock('react-hot-toast', () => ({ default: { error: vi.fn() } }));

const domainA = 'domain-a' as BusinessDomainId;
const domainB = 'domain-b' as BusinessDomainId;

const l1Capability: Capability = {
  id: 'cap-l1' as CapabilityId,
  name: 'Finance',
  level: 'L1',
  createdAt: '2024-01-01',
  _links: {},
};

function buildDropEvent(payload: unknown): React.DragEvent {
  return {
    preventDefault: vi.fn(),
    dataTransfer: {
      getData: () => (payload === undefined ? '' : JSON.stringify(payload)),
      dropEffect: 'none',
    },
  } as unknown as React.DragEvent;
}

describe('useDragHandlers', () => {
  it('associates a dropped L1 capability with the domain of the drop target', async () => {
    const associateCapability = vi.fn().mockResolvedValue(undefined);
    const refetchDomain = vi.fn().mockResolvedValue(undefined);
    const { result } = renderHook(() =>
      useDragHandlers({
        associateCapability,
        isCapabilityAssignedToDomain: () => false,
        refetchDomain,
      }),
    );

    await act(async () => {
      await result.current.handleDrop(domainB, buildDropEvent(l1Capability));
    });

    expect(associateCapability).toHaveBeenCalledWith(domainB, l1Capability.id);
    expect(refetchDomain).toHaveBeenCalledWith(domainB);
  });

  it('ignores drops of non-L1 capabilities', async () => {
    const associateCapability = vi.fn().mockResolvedValue(undefined);
    const { result } = renderHook(() =>
      useDragHandlers({
        associateCapability,
        isCapabilityAssignedToDomain: () => false,
        refetchDomain: vi.fn().mockResolvedValue(undefined),
      }),
    );

    await act(async () => {
      await result.current.handleDrop(domainA, buildDropEvent({ ...l1Capability, level: 'L2' }));
    });

    expect(associateCapability).not.toHaveBeenCalled();
  });

  it('skips association when the capability is already assigned to the target domain', async () => {
    const associateCapability = vi.fn().mockResolvedValue(undefined);
    const { result } = renderHook(() =>
      useDragHandlers({
        associateCapability,
        isCapabilityAssignedToDomain: (domainId, capabilityId) =>
          domainId === domainA && capabilityId === l1Capability.id,
        refetchDomain: vi.fn().mockResolvedValue(undefined),
      }),
    );

    await act(async () => {
      await result.current.handleDrop(domainA, buildDropEvent(l1Capability));
    });

    expect(associateCapability).not.toHaveBeenCalled();
  });

  it('tracks which domain is currently being dragged over, and clears it on drag end / leave', () => {
    const { result } = renderHook(() =>
      useDragHandlers({
        associateCapability: vi.fn(),
        isCapabilityAssignedToDomain: () => false,
        refetchDomain: vi.fn(),
      }),
    );

    act(() => {
      result.current.handleDragOver(domainA, {
        preventDefault: vi.fn(),
        dataTransfer: { dropEffect: 'none' },
      } as unknown as React.DragEvent);
    });
    expect(result.current.dragOverDomainId).toBe(domainA);

    act(() => result.current.handleDragLeave());
    expect(result.current.dragOverDomainId).toBeNull();
  });

  it('shows an error toast and clears drag state when the drop payload cannot be parsed', async () => {
    const toast = (await import('react-hot-toast')).default;
    const { result } = renderHook(() =>
      useDragHandlers({
        associateCapability: vi.fn(),
        isCapabilityAssignedToDomain: () => false,
        refetchDomain: vi.fn(),
      }),
    );

    const badEvent = {
      preventDefault: vi.fn(),
      dataTransfer: { getData: () => '{not-json', dropEffect: 'none' },
    } as unknown as React.DragEvent;

    await act(async () => {
      await result.current.handleDrop(domainA, badEvent);
    });

    expect(toast.error).toHaveBeenCalledWith('Failed to assign capability');
    expect(result.current.activeCapability).toBeNull();
  });
});
