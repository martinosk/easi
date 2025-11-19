import type { ViewportState } from '../types/storeTypes';

const VIEWPORT_STATES_KEY = 'viewportStates';

export function loadViewportStatesFromStorage(): Record<string, ViewportState> {
  try {
    const stored = localStorage.getItem(VIEWPORT_STATES_KEY);
    return stored ? JSON.parse(stored) : {};
  } catch {
    return {};
  }
}

export function saveViewportStatesToStorage(
  states: Record<string, ViewportState>
): void {
  try {
    localStorage.setItem(VIEWPORT_STATES_KEY, JSON.stringify(states));
  } catch (error) {
    console.error('Failed to save viewport states:', error);
  }
}
