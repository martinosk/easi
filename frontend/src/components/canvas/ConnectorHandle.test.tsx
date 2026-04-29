import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ConnectorHandle } from './ConnectorHandle';

vi.mock('@xyflow/react', () => ({
  Handle: ({
    onMouseDown,
    onMouseUp,
    children,
    ...props
  }: {
    onMouseDown?: React.MouseEventHandler;
    onMouseUp?: React.MouseEventHandler;
    children?: React.ReactNode;
    [key: string]: unknown;
  }) => (
    <div data-testid="handle" data-position={props.position} onMouseDown={onMouseDown} onMouseUp={onMouseUp}>
      {children}
    </div>
  ),
  Position: { Top: 'top', Right: 'right', Bottom: 'bottom', Left: 'left' },
}));

describe('ConnectorHandle', () => {
  const onConnectorClick = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders the handle element', () => {
    render(
      <ConnectorHandle
        type="source"
        position={'right' as never}
        nodeId="node-1"
        onConnectorClick={onConnectorClick}
      />,
    );
    expect(screen.getByTestId('handle')).toBeInTheDocument();
  });

  it('calls onConnectorClick on quick click (mousedown+mouseup <200ms, <5px)', () => {
    render(
      <ConnectorHandle
        type="source"
        position={'right' as never}
        nodeId="node-1"
        onConnectorClick={onConnectorClick}
      />,
    );
    const handle = screen.getByTestId('handle');

    fireEvent.mouseDown(handle, { clientX: 100, clientY: 100 });
    fireEvent.mouseUp(handle, { clientX: 102, clientY: 101 });

    expect(onConnectorClick).toHaveBeenCalledWith({
      nodeId: 'node-1',
      handlePosition: 'right',
    });
  });

  it('does NOT call onConnectorClick when mouse moves >5px (drag)', () => {
    render(
      <ConnectorHandle
        type="source"
        position={'right' as never}
        nodeId="node-1"
        onConnectorClick={onConnectorClick}
      />,
    );
    const handle = screen.getByTestId('handle');

    fireEvent.mouseDown(handle, { clientX: 100, clientY: 100 });
    fireEvent.mouseUp(handle, { clientX: 120, clientY: 100 });

    expect(onConnectorClick).not.toHaveBeenCalled();
  });

  it('does NOT call onConnectorClick when elapsed time exceeds 200ms', () => {
    vi.useFakeTimers();

    render(
      <ConnectorHandle
        type="source"
        position={'right' as never}
        nodeId="node-1"
        onConnectorClick={onConnectorClick}
      />,
    );
    const handle = screen.getByTestId('handle');

    fireEvent.mouseDown(handle, { clientX: 100, clientY: 100 });
    vi.advanceTimersByTime(250);
    fireEvent.mouseUp(handle, { clientX: 101, clientY: 100 });

    expect(onConnectorClick).not.toHaveBeenCalled();
    vi.useRealTimers();
  });

  it('uses handleProps.id as handlePosition when id is provided', () => {
    render(
      <ConnectorHandle
        type="source"
        position={'right' as never}
        id="custom-handle"
        nodeId="node-1"
        onConnectorClick={onConnectorClick}
      />,
    );
    const handle = screen.getByTestId('handle');

    fireEvent.mouseDown(handle, { clientX: 100, clientY: 100 });
    fireEvent.mouseUp(handle, { clientX: 100, clientY: 100 });

    expect(onConnectorClick).toHaveBeenCalledWith({
      nodeId: 'node-1',
      handlePosition: 'custom-handle',
    });
  });

  it('does nothing when onConnectorClick is not provided', () => {
    render(
      <ConnectorHandle
        type="source"
        position={'right' as never}
        nodeId="node-1"
      />,
    );
    const handle = screen.getByTestId('handle');

    fireEvent.mouseDown(handle, { clientX: 100, clientY: 100 });
    fireEvent.mouseUp(handle, { clientX: 100, clientY: 100 });
  });
});
