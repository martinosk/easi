import type {
  Component,
  ComponentId,
  Capability,
  CapabilityId,
  CapabilityRealization,
  View,
  ViewId,
  Relation,
  RelationId,
} from '../../api/types';
import {
  buildComponent,
  buildCapability,
  buildView,
  buildRelation,
  buildCapabilityRealization,
  resetIdCounter,
} from '../helpers/entityBuilders';

export interface MockDatabase {
  components: Component[];
  capabilities: Capability[];
  capabilityRealizations: CapabilityRealization[];
  views: View[];
  relations: Relation[];
}

let db: MockDatabase = createEmptyDb();

function createEmptyDb(): MockDatabase {
  return {
    components: [],
    capabilities: [],
    capabilityRealizations: [],
    views: [],
    relations: [],
  };
}

export function resetDb(): void {
  resetIdCounter();
  db = createEmptyDb();
}

export function seedDb(data: Partial<MockDatabase>): void {
  if (data.components) db.components = data.components;
  if (data.capabilities) db.capabilities = data.capabilities;
  if (data.capabilityRealizations) db.capabilityRealizations = data.capabilityRealizations;
  if (data.views) db.views = data.views;
  if (data.relations) db.relations = data.relations;
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
