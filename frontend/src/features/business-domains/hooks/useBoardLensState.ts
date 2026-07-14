import { useCallback, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { type BoardLens, DEFAULT_LENS, isBoardLens } from '../lens/boardLens';

export const LENS_PARAM = 'lens';

export interface BoardLensState {
  lens: BoardLens;
  setLens: (lens: BoardLens) => void;
  changesOnly: boolean;
  setChangesOnly: (changesOnly: boolean) => void;
}

export function useBoardLensState(): BoardLensState {
  const [searchParams, setSearchParams] = useSearchParams();
  const [changesOnly, setChangesOnly] = useState(false);

  const rawLens = searchParams.get(LENS_PARAM);
  const lens: BoardLens = isBoardLens(rawLens) ? rawLens : DEFAULT_LENS;

  const setLens = useCallback(
    (next: BoardLens) => {
      setSearchParams(
        (prev) => {
          const params = new URLSearchParams(prev);
          if (next === DEFAULT_LENS) params.delete(LENS_PARAM);
          else params.set(LENS_PARAM, next);
          return params;
        },
        { replace: true },
      );
    },
    [setSearchParams],
  );

  return { lens, setLens, changesOnly, setChangesOnly };
}
