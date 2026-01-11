import {
  toComponentId,
  toRelationId,
  toViewId,
  toCapabilityId,
  toCapabilityDependencyId,
  toRealizationId,
  toBusinessDomainId,
} from '../../api/types';
import type {
  Component,
  Relation,
  View,
  ViewComponent,
  ViewCapability,
  Capability,
  CapabilityDependency,
  CapabilityRealization,
  BusinessDomain,
  HATEOASLinks,
  CapabilityLevel,
  DependencyType,
  RealizationLevel,
  Expert,
} from '../../api/types';

let idCounter = 0;
function nextId(prefix: string): string {
  return `${prefix}-${++idCounter}`;
}

export function resetIdCounter(): void {
  idCounter = 0;
}

function buildLinks(self: string): HATEOASLinks {
  return {
    self,
    update: self,
    delete: self,
  };
}

export function buildComponent(overrides: Partial<Component> = {}): Component {
  const id = overrides.id ?? toComponentId(nextId('comp'));
  return {
    id,
    name: `Component ${id}`,
    description: 'Test component description',
    createdAt: '2024-01-01T00:00:00Z',
    _links: buildLinks(`/api/v1/components/${id}`),
    ...overrides,
  };
}

export function buildRelation(overrides: Partial<Relation> = {}): Relation {
  const id = overrides.id ?? toRelationId(nextId('rel'));
  return {
    id,
    sourceComponentId: toComponentId('comp-1'),
    targetComponentId: toComponentId('comp-2'),
    relationType: 'Triggers',
    name: 'Test Relation',
    description: 'Test relation description',
    createdAt: '2024-01-01T00:00:00Z',
    _links: buildLinks(`/api/v1/relations/${id}`),
    ...overrides,
  };
}

export function buildViewComponent(overrides: Partial<ViewComponent> = {}): ViewComponent {
  return {
    componentId: toComponentId('comp-1'),
    x: 100,
    y: 100,
    ...overrides,
  };
}

export function buildViewCapability(overrides: Partial<ViewCapability> = {}): ViewCapability {
  return {
    capabilityId: toCapabilityId('cap-1'),
    x: 200,
    y: 200,
    ...overrides,
  };
}

export function buildView(overrides: Partial<View> = {}): View {
  const id = overrides.id ?? toViewId(nextId('view'));
  return {
    id,
    name: `View ${id}`,
    description: 'Test view description',
    isDefault: false,
    components: [],
    capabilities: [],
    edgeType: 'smoothstep',
    colorScheme: 'default',
    createdAt: '2024-01-01T00:00:00Z',
    _links: buildLinks(`/api/v1/views/${id}`),
    ...overrides,
  };
}

export function buildExpert(overrides: Partial<Expert> = {}): Expert {
  return {
    name: 'John Doe',
    role: 'Tech Lead',
    contact: 'john.doe@example.com',
    addedAt: '2024-01-01T00:00:00Z',
    ...overrides,
  };
}

export function buildCapability(overrides: Partial<Capability> = {}): Capability {
  const id = overrides.id ?? toCapabilityId(nextId('cap'));
  return {
    id,
    name: `Capability ${id}`,
    description: 'Test capability description',
    level: 'L1' as CapabilityLevel,
    createdAt: '2024-01-01T00:00:00Z',
    _links: buildLinks(`/api/v1/capabilities/${id}`),
    ...overrides,
  };
}

export function buildCapabilityDependency(
  overrides: Partial<CapabilityDependency> = {}
): CapabilityDependency {
  const id = overrides.id ?? toCapabilityDependencyId(nextId('dep'));
  return {
    id,
    sourceCapabilityId: toCapabilityId('cap-1'),
    targetCapabilityId: toCapabilityId('cap-2'),
    dependencyType: 'Requires' as DependencyType,
    description: 'Test dependency',
    createdAt: '2024-01-01T00:00:00Z',
    _links: buildLinks(`/api/v1/capability-dependencies/${id}`),
    ...overrides,
  };
}

export function buildCapabilityRealization(
  overrides: Partial<CapabilityRealization> = {}
): CapabilityRealization {
  const id = overrides.id ?? toRealizationId(nextId('real'));
  return {
    id,
    capabilityId: toCapabilityId('cap-1'),
    componentId: toComponentId('comp-1'),
    componentName: 'Component 1',
    realizationLevel: 'Full' as RealizationLevel,
    origin: 'Direct',
    linkedAt: '2024-01-01T00:00:00Z',
    _links: buildLinks(`/api/v1/capability-realizations/${id}`),
    ...overrides,
  };
}

export function buildBusinessDomain(overrides: Partial<BusinessDomain> = {}): BusinessDomain {
  const id = overrides.id ?? toBusinessDomainId(nextId('domain'));
  return {
    id,
    name: `Business Domain ${id}`,
    description: 'Test business domain description',
    capabilityCount: 0,
    createdAt: '2024-01-01T00:00:00Z',
    _links: {
      ...buildLinks(`/api/v1/business-domains/${id}`),
      capabilities: `/api/v1/business-domains/${id}/capabilities`,
      associate: `/api/v1/business-domains/${id}/capabilities`,
    },
    ...overrides,
  };
}
