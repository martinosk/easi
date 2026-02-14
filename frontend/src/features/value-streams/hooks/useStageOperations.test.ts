import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useStageOperations } from './useStageOperations';
import type { ValueStreamDetail, ValueStreamId, StageId, HttpMethod } from '../../../api/types';

const mockAddStageMutateAsync = vi.fn();
const mockUpdateStageMutateAsync = vi.fn();
const mockDeleteStageMutateAsync = vi.fn();
const mockReorderStagesMutateAsync = vi.fn();
const mockAddCapabilityMutate = vi.fn();

vi.mock('./useValueStreamStages', () => ({
  useAddStage: () => ({ mutateAsync: mockAddStageMutateAsync }),
  useUpdateStage: () => ({ mutateAsync: mockUpdateStageMutateAsync }),
  useDeleteStage: () => ({ mutateAsync: mockDeleteStageMutateAsync }),
  useReorderStages: () => ({ mutateAsync: mockReorderStagesMutateAsync }),
  useAddStageCapability: () => ({ mutate: mockAddCapabilityMutate }),
}));

function createDetail(stageCount = 2): ValueStreamDetail {
  const stages = Array.from({ length: stageCount }, (_, i) => ({
    id: `stage-${i + 1}` as StageId,
    valueStreamId: 'vs-1' as ValueStreamId,
    name: `Stage ${i + 1}`,
    position: i + 1,
    _links: {
      edit: { href: `/api/v1/value-streams/vs-1/stages/stage-${i + 1}`, method: 'PUT' as HttpMethod },
      delete: { href: `/api/v1/value-streams/vs-1/stages/stage-${i + 1}`, method: 'DELETE' as HttpMethod },
    },
  }));

  return {
    id: 'vs-1' as ValueStreamId,
    name: 'Test VS',
    description: '',
    stageCount,
    createdAt: '2024-01-01T00:00:00Z',
    stages,
    stageCapabilities: [],
    _links: {
      self: { href: '/api/v1/value-streams/vs-1', method: 'GET' as HttpMethod },
      'x-add-stage': { href: '/api/v1/value-streams/vs-1/stages', method: 'POST' as HttpMethod },
      'x-reorder-stages': { href: '/api/v1/value-streams/vs-1/stages/positions', method: 'PUT' as HttpMethod },
    },
  };
}

describe('useStageOperations', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockAddStageMutateAsync.mockResolvedValue(createDetail());
    mockUpdateStageMutateAsync.mockResolvedValue(createDetail());
  });

  it('should open add form without position (append)', () => {
    const { result } = renderHook(() => useStageOperations(createDetail()));

    act(() => result.current.openAddForm());

    expect(result.current.isFormOpen).toBe(true);
    expect(result.current.editingStage).toBeNull();
  });

  it('should open add form with insert position', () => {
    const { result } = renderHook(() => useStageOperations(createDetail()));

    act(() => result.current.openAddForm(2));

    expect(result.current.isFormOpen).toBe(true);
  });

  it.each([
    { scenario: 'appending', position: undefined, name: 'New Stage', description: '', expectedPosition: undefined },
    { scenario: 'inserting between stages', position: 2, name: 'Inserted Stage', description: 'desc', expectedPosition: 2 },
  ])('should submit correctly when $scenario', async ({ position, name, description, expectedPosition }) => {
    const detail = createDetail();
    const { result } = renderHook(() => useStageOperations(detail));

    act(() => result.current.openAddForm(position));
    act(() => result.current.setFormData({ name, description }));

    await act(() => result.current.submitForm());

    expect(mockAddStageMutateAsync).toHaveBeenCalledWith({
      valueStream: detail,
      request: { name, description: description || undefined, position: expectedPosition },
    });
  });

  it('should clear insert position when closing form', () => {
    const { result } = renderHook(() => useStageOperations(createDetail()));

    act(() => result.current.openAddForm(2));
    expect(result.current.isFormOpen).toBe(true);

    act(() => result.current.closeForm());
    expect(result.current.isFormOpen).toBe(false);
  });

  it('should clear insert position after successful submission', async () => {
    const detail = createDetail();
    const { result } = renderHook(() => useStageOperations(detail));

    act(() => result.current.openAddForm(2));
    act(() => result.current.setFormData({ name: 'Test', description: '' }));
    await act(() => result.current.submitForm());

    expect(result.current.isFormOpen).toBe(false);

    act(() => result.current.openAddForm());
    act(() => result.current.setFormData({ name: 'Next Stage', description: '' }));
    await act(() => result.current.submitForm());

    expect(mockAddStageMutateAsync).toHaveBeenLastCalledWith({
      valueStream: detail,
      request: { name: 'Next Stage', description: undefined, position: undefined },
    });
  });

  it('should not include position when editing an existing stage', async () => {
    const detail = createDetail();
    const stage = detail.stages[0];
    const { result } = renderHook(() => useStageOperations(detail));

    act(() => result.current.openEditForm(stage));
    act(() => result.current.setFormData({ name: 'Renamed', description: '' }));
    await act(() => result.current.submitForm());

    expect(mockUpdateStageMutateAsync).toHaveBeenCalledWith({
      stage,
      request: { name: 'Renamed', description: undefined },
    });
    expect(mockAddStageMutateAsync).not.toHaveBeenCalled();
  });
});
