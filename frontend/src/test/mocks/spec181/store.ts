import type { RealizationRole } from '../../../features/architecture-direction/types';

export interface StubRealizationRole {
  capabilityId: string;
  capabilityName: string;
  componentId: string;
  componentName: string;
  role: RealizationRole;
  assignedBy: string;
  assignedAt: string;
}

interface Spec181Db {
  roles: StubRealizationRole[];
  canWrite: boolean;
}

function emptyDb(): Spec181Db {
  return { roles: [], canWrite: true };
}

let db: Spec181Db = emptyDb();

export function resetSpec181Db(): void {
  db = emptyDb();
}

export function seedSpec181Db(data: { roles?: StubRealizationRole[]; canWrite?: boolean }): void {
  if (data.roles) db.roles = data.roles;
  if (data.canWrite !== undefined) db.canWrite = data.canWrite;
}

export function canWriteRoles(): boolean {
  return db.canWrite;
}

export function findRole(capabilityId: string, componentId: string): StubRealizationRole | undefined {
  return db.roles.find((r) => r.capabilityId === capabilityId && r.componentId === componentId);
}

export function rolesForCapabilities(capabilityIds: string[] | null): StubRealizationRole[] {
  if (capabilityIds === null) return [...db.roles];
  return db.roles.filter((r) => capabilityIds.includes(r.capabilityId));
}

export function upsertRole(role: StubRealizationRole): void {
  const index = db.roles.findIndex((r) => r.capabilityId === role.capabilityId && r.componentId === role.componentId);
  if (index === -1) db.roles.push(role);
  else db.roles[index] = role;
}

export function removeRole(capabilityId: string, componentId: string): void {
  db.roles = db.roles.filter((r) => !(r.capabilityId === capabilityId && r.componentId === componentId));
}
