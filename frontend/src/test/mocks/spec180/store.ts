import type { TimeGrade } from '../../../api/types';

export interface StubTimeAssessment {
  id: string;
  capabilityId: string;
  capabilityName: string;
  componentId: string;
  componentName: string;
  grade: TimeGrade;
  rationale: string;
  assessedBy: string;
  assessedByName?: string;
  assessedAt: string;
}

interface Spec180Db {
  assessments: StubTimeAssessment[];
  canWrite: boolean;
}

function emptyDb(): Spec180Db {
  return { assessments: [], canWrite: true };
}

let db: Spec180Db = emptyDb();

export function resetSpec180Db(): void {
  db = emptyDb();
}

export function seedSpec180Db(data: { assessments?: StubTimeAssessment[]; canWrite?: boolean }): void {
  if (data.assessments) db.assessments = data.assessments;
  if (data.canWrite !== undefined) db.canWrite = data.canWrite;
}

export function canWriteAssessments(): boolean {
  return db.canWrite;
}

export function findAssessment(capabilityId: string, componentId: string): StubTimeAssessment | undefined {
  return db.assessments.find((a) => a.capabilityId === capabilityId && a.componentId === componentId);
}

export function assessmentsForCapabilities(capabilityIds: string[]): StubTimeAssessment[] {
  return db.assessments.filter((a) => capabilityIds.includes(a.capabilityId));
}

export function assessmentsForComponents(componentIds: string[]): StubTimeAssessment[] {
  return db.assessments.filter((a) => componentIds.includes(a.componentId));
}

export function upsertAssessment(assessment: StubTimeAssessment): void {
  const index = db.assessments.findIndex(
    (a) => a.capabilityId === assessment.capabilityId && a.componentId === assessment.componentId,
  );
  if (index === -1) db.assessments.push(assessment);
  else db.assessments[index] = assessment;
}

export function removeAssessment(capabilityId: string, componentId: string): void {
  db.assessments = db.assessments.filter((a) => !(a.capabilityId === capabilityId && a.componentId === componentId));
}
