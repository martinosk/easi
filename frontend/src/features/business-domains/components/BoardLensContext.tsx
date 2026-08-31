import { createContext, type ReactNode, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react';
import { type BoardLens, DEFAULT_LENS } from '../lens/boardLens';
import type { JourneyIndex } from '../lens/journeyIndex';

export type TraceEnd = 'source' | 'dest';

const TRACE_HIGHLIGHT_MS = 2600;

const EMPTY_INDEX: JourneyIndex = {
  getJourney: () => undefined,
  getJourneys: () => [],
  getArrivingMovesForParent: () => [],
  getArrivingMovesForDomain: () => [],
  sourceDomainName: () => undefined,
};

export interface BoardLensContextValue {
  lens: BoardLens;
  changesOnly: boolean;
  index: JourneyIndex;
  tracedMoveId: string | null;
  activateTrace: (moveId: string, originEnd: TraceEnd) => void;
  registerTraceRef: (moveId: string, end: TraceEnd, element: HTMLElement | null) => void;
  openCapabilityById: (capabilityId: string) => void;
}

const BoardLensContext = createContext<BoardLensContextValue>({
  lens: DEFAULT_LENS,
  changesOnly: false,
  index: EMPTY_INDEX,
  tracedMoveId: null,
  activateTrace: () => {},
  registerTraceRef: () => {},
  openCapabilityById: () => {},
});

export function useBoardLens(): BoardLensContextValue {
  return useContext(BoardLensContext);
}

export interface BoardLensProviderProps {
  lens: BoardLens;
  changesOnly: boolean;
  index: JourneyIndex;
  openCapabilityById: (capabilityId: string) => void;
  children: ReactNode;
}

export function BoardLensProvider({ lens, changesOnly, index, openCapabilityById, children }: BoardLensProviderProps) {
  const [tracedMoveId, setTracedMoveId] = useState<string | null>(null);
  const refs = useRef(new Map<string, HTMLElement>());
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  const registerTraceRef = useCallback((moveId: string, end: TraceEnd, element: HTMLElement | null) => {
    const key = `${moveId}:${end}`;
    if (element) refs.current.set(key, element);
    else refs.current.delete(key);
  }, []);

  const activateTrace = useCallback((moveId: string, originEnd: TraceEnd) => {
    setTracedMoveId(moveId);
    const otherEnd: TraceEnd = originEnd === 'source' ? 'dest' : 'source';
    refs.current.get(`${moveId}:${otherEnd}`)?.scrollIntoView?.({ behavior: 'smooth', block: 'center' });
    clearTimeout(timeoutRef.current);
    timeoutRef.current = setTimeout(() => setTracedMoveId(null), TRACE_HIGHLIGHT_MS);
  }, []);

  useEffect(() => () => clearTimeout(timeoutRef.current), []);

  const value = useMemo<BoardLensContextValue>(
    () => ({ lens, changesOnly, index, tracedMoveId, activateTrace, registerTraceRef, openCapabilityById }),
    [lens, changesOnly, index, tracedMoveId, activateTrace, registerTraceRef, openCapabilityById],
  );

  return <BoardLensContext.Provider value={value}>{children}</BoardLensContext.Provider>;
}
