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

type EventInit = { clientX: number; clientY: number; button?: number };

type ClickSequence = {
  threshold?: number;
  downId: string;
  upId: string;
  down: EventInit;
  up?: EventInit;
};

const fireClickSequence = ({ threshold, downId, upId, down, up = down }: ClickSequence) => {
  const onHandleClick = vi.fn();
  const { getByTestId } = render(<Harness onHandleClick={onHandleClick} threshold={threshold} />);
  fireEvent.mouseDown(getByTestId(downId), down);
  fireEvent.mouseUp(getByTestId(upId), up);
  return onHandleClick;
};

describe('useHandleClickDetection', () => {
  const firesCases: ReadonlyArray<{
    name: string;
    sequence: ClickSequence;
    expected: { nodeId: string; side: string; clientX: number; clientY: number };
  }> = [
    {
      name: 'fires onHandleClick when mousedown and mouseup happen on the same handle without movement',
      sequence: { downId: 'h-right-source', upId: 'h-right-source', down: { clientX: 50, clientY: 60 } },
      expected: { nodeId: 'comp-1', side: 'right', clientX: 50, clientY: 60 },
    },
    {
      name: 'fires when mouseup lands on the sibling source/target handle on the same side',
      sequence: { downId: 'h-right-source', upId: 'h-right-target', down: { clientX: 50, clientY: 60 } },
      expected: { nodeId: 'comp-1', side: 'right', clientX: 50, clientY: 60 },
    },
    {
      name: 'reports the correct side for each handle position',
      sequence: { downId: 'h-top-2', upId: 'h-top-2', down: { clientX: 70, clientY: 80 } },
      expected: { nodeId: 'comp-2', side: 'top', clientX: 70, clientY: 80 },
    },
    {
      name: 'reads nodeId from data-nodeid on the handle when present (React Flow style)',
      sequence: { downId: 'rf-handle', upId: 'rf-handle', down: { clientX: 5, clientY: 5 } },
      expected: { nodeId: 'rf-comp-99', side: 'left', clientX: 5, clientY: 5 },
    },
  ];

  it.each(firesCases)('$name', ({ sequence, expected }) => {
    const onHandleClick = fireClickSequence(sequence);

    expect(onHandleClick).toHaveBeenCalledWith(expected);
  });

  const ignoresCases: ReadonlyArray<{ name: string; sequence: ClickSequence }> = [
    {
      name: 'does not fire when movement exceeds the threshold (drag)',
      sequence: {
        threshold: 5,
        downId: 'h-right-source',
        upId: 'h-right-source',
        down: { clientX: 50, clientY: 60 },
        up: { clientX: 100, clientY: 60 },
      },
    },
    {
      name: 'does not fire when mousedown was not on a handle',
      sequence: { downId: 'not-a-handle', upId: 'not-a-handle', down: { clientX: 10, clientY: 10 } },
    },
    {
      name: 'does not fire when mouseup happens on a handle on a different side',
      sequence: {
        downId: 'h-right-source',
        upId: 'h-left',
        down: { clientX: 50, clientY: 60 },
        up: { clientX: 51, clientY: 61 },
      },
    },
    {
      name: 'ignores non-primary mouse buttons (right-click)',
      sequence: {
        downId: 'h-right-source',
        upId: 'h-right-source',
        down: { button: 2, clientX: 50, clientY: 60 },
      },
    },
  ];

  it.each(ignoresCases)('$name', ({ sequence }) => {
    const onHandleClick = fireClickSequence(sequence);

    expect(onHandleClick).not.toHaveBeenCalled();
  });
});
