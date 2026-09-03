import type {
  AcquiredEntity,
  AcquiredEntityId,
  BusinessDomain,
  Capability,
  CapabilityId,
  CapabilityRealization,
  Component,
  ComponentId,
  InternalTeam,
  InternalTeamId,
  OriginRelationship,
  Relation,
  RelationId,
  Vendor,
  VendorId,
  View,
  ViewId,
} from '../../api/types';
import {
  buildCapability,
  buildCapabilityRealization,
  buildComponent,
  buildRelation,
  buildView,
  resetIdCounter,
} from '../helpers/entityBuilders';

export interface MockDatabase {
  businessDomains: BusinessDomain[];
  components: Component[];
  capabilities: Capability[];
  capabilityRealizations: CapabilityRealization[];
  views: View[];
  relations: Relation[];
  acquiredEntities: AcquiredEntity[];
  vendors: Vendor[];
  internalTeams: InternalTeam[];
  originRelationships: OriginRelationship[];
}

let db: MockDatabase = createEmptyDb();

function createEmptyDb(): MockDatabase {
  return {
    businessDomains: [],
    components: [],
    capabilities: [],
    capabilityRealizations: [],
    views: [],
    relations: [],
    acquiredEntities: [],
    vendors: [],
    internalTeams: [],
    originRelationships: [],
  };
}

export function resetDb(): void {
  resetIdCounter();
  db = createEmptyDb();
}

export function seedDb(data: Partial<MockDatabase>): void {
  for (const [collection, items] of Object.entries(data)) {
    if (items) (db as unknown as Record<string, unknown[]>)[collection] = items;
  }
}

export function getDb(): MockDatabase {
  return db;
}

export function getComponents(): Component[] {
  return db.components;
}

export function getComponent(id: ComponentId): Component | undefined {
  return db.components.find((c) => c.id === id);
}

export function addComponent(component: Partial<Component> = {}): Component {
  const newComponent = buildComponent(component);
  db.components.push(newComponent);
  return newComponent;
}

export function updateComponent(id: ComponentId, updates: Partial<Component>): Component | undefined {
  const index = db.components.findIndex((c) => c.id === id);
  if (index < 0) return undefined;
  db.components[index] = { ...db.components[index], ...updates };
  return db.components[index];
}

export function getCapabilities(): Capability[] {
  return db.capabilities;
}

export function getCapability(id: CapabilityId): Capability | undefined {
  return db.capabilities.find((c) => c.id === id);
}

export function addCapability(capability: Partial<Capability> = {}): Capability {
  const newCapability = buildCapability(capability);
  db.capabilities.push(newCapability);
  return newCapability;
}

export function updateCapability(id: CapabilityId, updates: Partial<Capability>): Capability | undefined {
  const index = db.capabilities.findIndex((c) => c.id === id);
  if (index < 0) return undefined;
  db.capabilities[index] = { ...db.capabilities[index], ...updates };
  return db.capabilities[index];
}

export function getCapabilityRealizations(): CapabilityRealization[] {
  return db.capabilityRealizations;
}

export function getRealizationsByCapability(capabilityId: CapabilityId): CapabilityRealization[] {
  return db.capabilityRealizations.filter((r) => r.capabilityId === capabilityId);
}

export function getRealizationsByComponent(componentId: ComponentId): CapabilityRealization[] {
  return db.capabilityRealizations.filter((r) => r.componentId === componentId);
}

export function addCapabilityRealization(realization: Partial<CapabilityRealization> = {}): CapabilityRealization {
  const newRealization = buildCapabilityRealization(realization);
  db.capabilityRealizations.push(newRealization);
  return newRealization;
}

export function getViews(): View[] {
  return db.views;
}

export function getView(id: ViewId): View | undefined {
  return db.views.find((v) => v.id === id);
}

export function addView(view: Partial<View> = {}): View {
  const newView = buildView(view);
  db.views.push(newView);
  return newView;
}

export function updateView(id: ViewId, updates: Partial<View>): View | undefined {
  const index = db.views.findIndex((v) => v.id === id);
  if (index === -1) return undefined;
  db.views[index] = { ...db.views[index], ...updates };
  return db.views[index];
}

export function getRelations(): Relation[] {
  return db.relations;
}

export function getRelation(id: RelationId): Relation | undefined {
  return db.relations.find((r) => r.id === id);
}

export function addRelation(relation: Partial<Relation> = {}): Relation {
  const newRelation = buildRelation(relation);
  db.relations.push(newRelation);
  return newRelation;
}

export function getBusinessDomains(): BusinessDomain[] {
  return db.businessDomains;
}

function updateIn<T extends { id: string }>(items: T[], id: string, updates: Partial<T>): T | undefined {
  const index = items.findIndex((item) => item.id === id);
  if (index < 0) return undefined;
  items[index] = { ...items[index], ...updates };
  return items[index];
}

export function getAcquiredEntities(): AcquiredEntity[] {
  return db.acquiredEntities;
}

export function getAcquiredEntity(id: AcquiredEntityId): AcquiredEntity | undefined {
  return db.acquiredEntities.find((entity) => entity.id === id);
}

export function updateAcquiredEntity(
  id: AcquiredEntityId,
  updates: Partial<AcquiredEntity>,
): AcquiredEntity | undefined {
  return updateIn(db.acquiredEntities, id, updates);
}

export function getVendors(): Vendor[] {
  return db.vendors;
}

export function getVendor(id: VendorId): Vendor | undefined {
  return db.vendors.find((vendor) => vendor.id === id);
}

export function updateVendor(id: VendorId, updates: Partial<Vendor>): Vendor | undefined {
  return updateIn(db.vendors, id, updates);
}

export function getInternalTeams(): InternalTeam[] {
  return db.internalTeams;
}

export function getInternalTeam(id: InternalTeamId): InternalTeam | undefined {
  return db.internalTeams.find((team) => team.id === id);
}

export function updateInternalTeam(id: InternalTeamId, updates: Partial<InternalTeam>): InternalTeam | undefined {
  return updateIn(db.internalTeams, id, updates);
}

export function getOriginRelationships(): OriginRelationship[] {
  return db.originRelationships;
}
