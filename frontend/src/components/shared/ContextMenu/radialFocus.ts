export interface FocusState {
  current: number | null;
  count: number;
}

export interface KeyIntent {
  key: string;
  shiftKey: boolean;
}

function cycle(state: FocusState, delta: number): number {
  if (state.current === null) return delta > 0 ? 0 : state.count - 1;
  return (state.current + delta + state.count) % state.count;
}

const DIRECTIONAL_KEY_DELTAS: Record<string, number> = {
  ArrowRight: 1,
  ArrowDown: 1,
  ArrowLeft: -1,
  ArrowUp: -1,
};

const ABSOLUTE_KEYS: Record<string, (count: number) => number> = {
  Home: () => 0,
  End: (count) => count - 1,
};

export function nextFocusForKey(intent: KeyIntent, state: FocusState): number | null {
  const absolute = ABSOLUTE_KEYS[intent.key];
  if (absolute) return absolute(state.count);

  const directionalDelta = DIRECTIONAL_KEY_DELTAS[intent.key];
  if (directionalDelta !== undefined) return cycle(state, directionalDelta);

  if (intent.key === 'Tab') return cycle(state, intent.shiftKey ? -1 : 1);

  return null;
}
