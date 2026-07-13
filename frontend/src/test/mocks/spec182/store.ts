import type { JourneyKind, MilestoneStatus, TargetPeriod } from '../../../features/journeys/types';

export interface StubApplicationRef {
  componentId: string;
  componentName: string;
  stale: boolean;
}

export interface StubMilestone {
  id: string;
  label: string;
  targetPeriod: TargetPeriod | null;
  status: MilestoneStatus;
}

export interface StubMove {
  targetDomainId: string;
  targetDomainName: string;
  targetDomainStale: boolean;
  targetParentId: string | null;
  targetParentName: string;
  targetParentStale: boolean;
  resultingName: string;
}

export type StubJourneyStatus = 'planned' | 'in-flight' | 'done' | 'abandoned';

export interface StubJourney {
  id: string;
  capabilityId: string;
  capabilityName: string;
  capabilityStale: boolean;
  kind: JourneyKind;
  status: StubJourneyStatus;
  progress: number | null;
  targetPeriod: TargetPeriod | null;
  note: string;
  fromApplications: StubApplicationRef[];
  toApplication: StubApplicationRef;
  move?: StubMove;
  milestones: StubMilestone[];
  plannedBy: string;
  plannedByName: string;
  plannedAt: string;
  updatedAt: string;
  startedAt: string | null;
  completedAt: string | null;
  abandonedAt: string | null;
}

interface Spec182Db {
  journeys: StubJourney[];
  canWrite: boolean;
  nextId: number;
}

function emptyDb(): Spec182Db {
  return { journeys: [], canWrite: true, nextId: 1 };
}

let db: Spec182Db = emptyDb();

export function resetSpec182Db(): void {
  db = emptyDb();
}

export function seedSpec182Db(data: { journeys?: StubJourney[]; canWrite?: boolean }): void {
  if (data.journeys) db.journeys = data.journeys;
  if (data.canWrite !== undefined) db.canWrite = data.canWrite;
}

export function canWriteJourneys(): boolean {
  return db.canWrite;
}

export function nextJourneyId(): string {
  return `journey-${db.nextId++}`;
}

export function nextMilestoneId(): string {
  return `milestone-${db.nextId++}`;
}

export function isActiveStatus(status: StubJourneyStatus): boolean {
  return status === 'planned' || status === 'in-flight';
}

export function findJourney(journeyId: string): StubJourney | undefined {
  return db.journeys.find((j) => j.id === journeyId);
}

export function findActiveJourneyForCapability(capabilityId: string): StubJourney | undefined {
  return db.journeys.find((j) => j.capabilityId === capabilityId && isActiveStatus(j.status));
}

export function getHistoryForCapability(capabilityId: string): StubJourney[] {
  return db.journeys
    .filter((j) => j.capabilityId === capabilityId)
    .slice()
    .sort((a, b) => b.plannedAt.localeCompare(a.plannedAt));
}

export function getCurrentByCapabilityIds(capabilityIds: string[]): StubJourney[] {
  const result: StubJourney[] = [];
  for (const capabilityId of capabilityIds) {
    const active = findActiveJourneyForCapability(capabilityId);
    if (active) {
      result.push(active);
      continue;
    }
    const history = getHistoryForCapability(capabilityId);
    if (history.length > 0) result.push(history[0]);
  }
  return result;
}

export function addJourney(journey: StubJourney): void {
  db.journeys.push(journey);
}

export function replaceJourney(journey: StubJourney): void {
  const index = db.journeys.findIndex((j) => j.id === journey.id);
  if (index === -1) db.journeys.push(journey);
  else db.journeys[index] = journey;
}
