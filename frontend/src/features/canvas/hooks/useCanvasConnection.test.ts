import { renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { useCanvasConnection } from './useCanvasConnection';

const handleCapabilityParentConnection = vi.fn();
const handleCapabilityComponentConnection = vi.fn();
const handleOriginComponentConnection = vi.fn();

vi.mock('./useCapabilityConnection', async () => {
  const actual = await vi.importActual<typeof import('./useCapabilityConnection')>('./useCapabilityConnection');
  return {
    ...actual,
    useCapabilityConnection: () => ({
      handleCapabilityParentConnection,
      handleCapabilityComponentConnection,
    }),
  };
});

vi.mock('./useOriginConnection', () => ({
  useOriginConnection: () => ({ handleOriginComponentConnection }),
}));

describe('useCanvasConnection — self-loop rejection', () => {
  it.each([
    { kind: 'component', id: 'comp-1' },
    { kind: 'capability', id: 'cap-1' },
    { kind: 'origin (acquired)', id: 'acq-1' },
  ])('does not dispatch any handler when source equals target ($kind)', async ({ id }) => {
    const onConnect = vi.fn();
    const { result } = renderHook(() => useCanvasConnection(onConnect));

    await result.current.onConnectHandler({
      source: id,
      target: id,
      sourceHandle: 'right',
      targetHandle: 'left',
    });

    expect(onConnect).not.toHaveBeenCalled();
    expect(handleCapabilityParentConnection).not.toHaveBeenCalled();
    expect(handleCapabilityComponentConnection).not.toHaveBeenCalled();
    expect(handleOriginComponentConnection).not.toHaveBeenCalled();
  });
});
