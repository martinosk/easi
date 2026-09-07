import type { CapabilityJourney } from '../../journeys/types';

export type BoardLens = 'now' | 'journey' | 'target';

export const BOARD_LENSES: readonly BoardLens[] = ['now', 'journey', 'target'];

export const DEFAULT_LENS: BoardLens = 'now';

export function isBoardLens(value: string | null | undefined): value is BoardLens {
  return value === 'now' || value === 'journey' || value === 'target';
}

export const LENS_LABELS: Record<BoardLens, string> = {
  now: 'Now',
  journey: 'Journey',
  target: 'Target',
};

export type BoardJourneyStatus = 'steady' | 'not-started' | 'in-flight' | 'done' | 'planned-move';

function isActiveJourney(journey: CapabilityJourney): boolean {
  return journey.status === 'planned' || journey.status === 'in-flight';
}

export function selectBoardJourneys(journeys: CapabilityJourney[]): CapabilityJourney[] {
  const active = journeys.filter(isActiveJourney);
  if (active.length > 0) return active;
  const done = journeys.find((journey) => journey.status === 'done');
  return done ? [done] : [];
}

export function selectBoardJourney(journeys: CapabilityJourney[]): CapabilityJourney | undefined {
  const board = selectBoardJourneys(journeys);
  return board.find((journey) => journey.kind !== 'maturity') ?? board[0];
}

export function capabilityJourneyStatus(journey: CapabilityJourney | undefined): BoardJourneyStatus {
  if (!journey) return 'steady';
  if (journey.kind === 'move') {
    if (journey.status === 'done') return 'done';
    if (journey.status === 'in-flight') return 'in-flight';
    return 'planned-move';
  }
  if (journey.status === 'done') return 'done';
  if (journey.status === 'in-flight') return 'in-flight';
  return 'not-started';
}

export function isMoveJourney(journey: CapabilityJourney | undefined): journey is CapabilityJourney {
  return journey?.kind === 'move';
}
