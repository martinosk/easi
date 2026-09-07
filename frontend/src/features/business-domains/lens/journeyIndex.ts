import type { CapabilityTreeNode } from '../../capabilities/hooks/useCapabilityTree';
import type { CapabilityJourney } from '../../journeys/types';
import { capabilityJourneyStatus, selectBoardJourney, selectBoardJourneys } from './boardLens';

export interface JourneyIndex {
  getJourney: (capabilityId: string) => CapabilityJourney | undefined;
  getJourneys: (capabilityId: string) => CapabilityJourney[];
  getArrivingMovesForParent: (capabilityId: string) => CapabilityJourney[];
  getArrivingMovesForDomain: (domainId: string) => CapabilityJourney[];
  sourceDomainName: (capabilityId: string) => string | undefined;
}

export interface BuildJourneyIndexParams {
  journeys: CapabilityJourney[];
  capabilityDomainNames: Map<string, string>;
}

function pushInto<T>(map: Map<string, T[]>, key: string, value: T): void {
  const list = map.get(key) ?? [];
  list.push(value);
  map.set(key, list);
}

export function buildJourneyIndex({ journeys, capabilityDomainNames }: BuildJourneyIndexParams): JourneyIndex {
  const journeysByCapability = new Map<string, CapabilityJourney[]>();
  for (const journey of journeys) {
    pushInto(journeysByCapability, journey.capabilityId, journey);
  }

  const boardJourney = new Map<string, CapabilityJourney>();
  const boardJourneys = new Map<string, CapabilityJourney[]>();
  const movesByParent = new Map<string, CapabilityJourney[]>();
  const movesByDomain = new Map<string, CapabilityJourney[]>();

  for (const [capabilityId, list] of journeysByCapability) {
    const board = selectBoardJourney(list);
    if (!board) continue;
    boardJourney.set(capabilityId, board);
    boardJourneys.set(capabilityId, selectBoardJourneys(list));
    if (board.kind !== 'move' || !board.move) continue;
    if (board.move.targetParentId) pushInto(movesByParent, board.move.targetParentId, board);
    else pushInto(movesByDomain, board.move.targetDomainId, board);
  }

  return {
    getJourney: (capabilityId) => boardJourney.get(capabilityId),
    getJourneys: (capabilityId) => boardJourneys.get(capabilityId) ?? [],
    getArrivingMovesForParent: (capabilityId) => movesByParent.get(capabilityId) ?? [],
    getArrivingMovesForDomain: (domainId) => movesByDomain.get(domainId) ?? [],
    sourceDomainName: (capabilityId) => capabilityDomainNames.get(capabilityId),
  };
}

export interface SummaryCounts {
  settled: number;
  inFlight: number;
  notStarted: number;
}

export function summaryCounts(
  l1Nodes: CapabilityTreeNode[],
  getJourney: (capabilityId: string) => CapabilityJourney | undefined,
): SummaryCounts {
  const counts: SummaryCounts = { settled: 0, inFlight: 0, notStarted: 0 };

  const tally = (capabilityId: string) => {
    const status = capabilityJourneyStatus(getJourney(capabilityId));
    if (status === 'done' || status === 'steady') counts.settled++;
    else if (status === 'in-flight') counts.inFlight++;
    else counts.notStarted++;
  };

  for (const l1 of l1Nodes) {
    const cards = l1.children.length ? l1.children : [l1];
    for (const card of cards) tally(card.capability.id);
  }

  return counts;
}

export function capabilityHasChange(capabilityId: string, index: JourneyIndex): boolean {
  return index.getJourney(capabilityId) !== undefined || index.getArrivingMovesForParent(capabilityId).length > 0;
}

export function l1HasChange(node: CapabilityTreeNode, index: JourneyIndex): boolean {
  const visit = (current: CapabilityTreeNode): boolean => {
    if (capabilityHasChange(current.capability.id, index)) return true;
    return current.children.some(visit);
  };
  return visit(node);
}
