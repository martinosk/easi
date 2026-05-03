import { fireEvent, render } from '@testing-library/react';
import React from 'react';
import { describe, expect, it, vi } from 'vitest';
import { useHandleClickDetection } from './useHandleClickDetection';

type ClickFn = Parameters<typeof useHandleClickDetection>[0];

const Harness: React.FC<{ onHandleClick: ClickFn; threshold?: number }> = ({
  onHandleClick,
  threshold,
}) => {
  useHandleClickDetection(onHandleClick, threshold);
  return (
    <div>
      <div data-id="comp-1">
        <div className="react-flow__handle react-flow__handle-right" data-handlepos="right" data-testid="h-right-source" />
        <div className="react-flow__handle react-flow__handle-right" data-handlepos="right" data-testid="h-right-target" />
        <div className="react-flow__handle react-flow__handle-left" data-handlepos="left" data-testid="h-left" />
      </div>
      <div data-id="comp-2">
        <div className="react-flow__handle react-flow__handle-top" data-handlepos="top" data-testid="h-top-2" />
      </div>
      <div
        className="react-flow__handle"
        data-nodeid="rf-comp-99"
        data-handlepos="left"
        data-testid="rf-handle"
      />
      <div data-testid="not-a-handle" />
    </div>
  );
};

describe('useHandleClickDetection', () => {
  it('fires onHandleClick when mousedown and mouseup happen on the same handle without movement', () => {
    const onHandleClick = vi.fn();
    const { getByTestId } = render(<Harness onHandleClick={onHandleClick} />);
    const handle = getByTestId('h-right-source');

    fireEvent.mouseDown(handle, { clientX: 50, clientY: 60 });
    fireEvent.mouseUp(handle, { clientX: 50, clientY: 60 });

    expect(onHandleClick).toHaveBeenCalledTimes(1);
    expect(onHandleClick).toHaveBeenCalledWith({
      nodeId: 'comp-1',
      side: 'right',
      clientX: 50,
      clientY: 60,
    });
  });

  it('fires when mouseup lands on the sibling source/target handle on the same side', () => {
    const onHandleClick = vi.fn();
    const { getByTestId } = render(<Harness onHandleClick={onHandleClick} />);

    fireEvent.mouseDown(getByTestId('h-right-source'), { clientX: 50, clientY: 60 });
    fireEvent.mouseUp(getByTestId('h-right-target'), { clientX: 50, clientY: 60 });

    expect(onHandleClick).toHaveBeenCalledTimes(1);
    expect(onHandleClick).toHaveBeenCalledWith({
      nodeId: 'comp-1',
      side: 'right',
      clientX: 50,
      clientY: 60,
    });
  });

  it('does not fire when movement exceeds the threshold (drag)', () => {
    const onHandleClick = vi.fn();
    const { getByTestId } = render(<Harness onHandleClick={onHandleClick} threshold={5} />);
    const handle = getByTestId('h-right-source');

    fireEvent.mouseDown(handle, { clientX: 50, clientY: 60 });
    fireEvent.mouseUp(handle, { clientX: 100, clientY: 60 });

    expect(onHandleClick).not.toHaveBeenCalled();
  });

  it('does not fire when mousedown was not on a handle', () => {
    const onHandleClick = vi.fn();
    const { getByTestId } = render(<Harness onHandleClick={onHandleClick} />);
    const notAHandle = getByTestId('not-a-handle');

    fireEvent.mouseDown(notAHandle, { clientX: 10, clientY: 10 });
    fireEvent.mouseUp(notAHandle, { clientX: 10, clientY: 10 });

    expect(onHandleClick).not.toHaveBeenCalled();
  });

  it('does not fire when mouseup happens on a handle on a different side', () => {
    const onHandleClick = vi.fn();
    const { getByTestId } = render(<Harness onHandleClick={onHandleClick} />);

    fireEvent.mouseDown(getByTestId('h-right-source'), { clientX: 50, clientY: 60 });
    fireEvent.mouseUp(getByTestId('h-left'), { clientX: 51, clientY: 61 });

    expect(onHandleClick).not.toHaveBeenCalled();
  });

  it('ignores non-primary mouse buttons (right-click)', () => {
    const onHandleClick = vi.fn();
    const { getByTestId } = render(<Harness onHandleClick={onHandleClick} />);
    const handle = getByTestId('h-right-source');

    fireEvent.mouseDown(handle, { button: 2, clientX: 50, clientY: 60 });
    fireEvent.mouseUp(handle, { button: 2, clientX: 50, clientY: 60 });

    expect(onHandleClick).not.toHaveBeenCalled();
  });

  it('reports the correct side for each handle position', () => {
    const onHandleClick = vi.fn();
    const { getByTestId } = render(<Harness onHandleClick={onHandleClick} />);

    fireEvent.mouseDown(getByTestId('h-top-2'), { clientX: 70, clientY: 80 });
    fireEvent.mouseUp(getByTestId('h-top-2'), { clientX: 70, clientY: 80 });

    expect(onHandleClick).toHaveBeenCalledWith({
      nodeId: 'comp-2',
      side: 'top',
      clientX: 70,
      clientY: 80,
    });
  });

  it('reads nodeId from data-nodeid on the handle when present (React Flow style)', () => {
    const onHandleClick = vi.fn();
    const { getByTestId } = render(<Harness onHandleClick={onHandleClick} />);
    fireEvent.mouseDown(getByTestId('rf-handle'), { clientX: 5, clientY: 5 });
    fireEvent.mouseUp(getByTestId('rf-handle'), { clientX: 5, clientY: 5 });
    expect(onHandleClick).toHaveBeenCalledWith({ nodeId: 'rf-comp-99', side: 'left', clientX: 5, clientY: 5 });
  });
});
