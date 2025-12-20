import { vi, type Mock } from 'vitest';

export interface MockComponentsApi {
  getAll: Mock;
  getById: Mock;
  create: Mock;
  update: Mock;
  delete: Mock;
}

export interface MockRelationsApi {
  getAll: Mock;
  getById: Mock;
  create: Mock;
  update: Mock;
  delete: Mock;
}

export interface MockViewsApi {
  getAll: Mock;
  getById: Mock;
  create: Mock;
  delete: Mock;
  rename: Mock;
  setDefault: Mock;
  updateEdgeType: Mock;
  updateColorScheme: Mock;
  getComponents: Mock;
  addComponent: Mock;
  removeComponent: Mock;
  updateComponentPosition: Mock;
  updateComponentColor: Mock;
  clearComponentColor: Mock;
  updateMultiplePositions: Mock;
  addCapability: Mock;
  removeCapability: Mock;
  updateCapabilityPosition: Mock;
  updateCapabilityColor: Mock;
  clearCapabilityColor: Mock;
}

export interface MockCapabilitiesApi {
  getAll: Mock;
  getById: Mock;
  getChildren: Mock;
  create: Mock;
  update: Mock;
  updateMetadata: Mock;
  addExpert: Mock;
  addTag: Mock;
  delete: Mock;
  changeParent: Mock;
  getAllDependencies: Mock;
  getOutgoingDependencies: Mock;
  getIncomingDependencies: Mock;
  createDependency: Mock;
  deleteDependency: Mock;
  getSystemsByCapability: Mock;
  getCapabilitiesByComponent: Mock;
  linkSystem: Mock;
  updateRealization: Mock;
  deleteRealization: Mock;
}

export interface MockBusinessDomainsApi {
  getAll: Mock;
  getById: Mock;
  create: Mock;
  update: Mock;
  delete: Mock;
  getCapabilities: Mock;
  associateCapability: Mock;
  dissociateCapability: Mock;
  getCapabilityRealizations: Mock;
}

export interface MockLayoutsApi {
  get: Mock;
  upsert: Mock;
  delete: Mock;
  updatePreferences: Mock;
  upsertElement: Mock;
  deleteElement: Mock;
  batchUpdateElements: Mock;
}

export interface MockMetadataApi {
  getMaturityLevels: Mock;
  getStatuses: Mock;
  getOwnershipModels: Mock;
  getStrategyPillars: Mock;
  getVersion: Mock;
  getLatestRelease: Mock;
  getReleaseByVersion: Mock;
  getReleases: Mock;
}

export function createMockComponentsApi(
  overrides?: Partial<MockComponentsApi>
): MockComponentsApi {
  return {
    getAll: vi.fn().mockResolvedValue([]),
    getById: vi.fn().mockResolvedValue(null),
    create: vi.fn().mockResolvedValue({ id: 'comp-1', name: 'Test' }),
    update: vi.fn().mockResolvedValue({ id: 'comp-1', name: 'Updated' }),
    delete: vi.fn().mockResolvedValue(undefined),
    ...overrides,
  };
}

export function createMockRelationsApi(
  overrides?: Partial<MockRelationsApi>
): MockRelationsApi {
  return {
    getAll: vi.fn().mockResolvedValue([]),
    getById: vi.fn().mockResolvedValue(null),
    create: vi.fn().mockResolvedValue({ id: 'rel-1' }),
    update: vi.fn().mockResolvedValue({ id: 'rel-1' }),
    delete: vi.fn().mockResolvedValue(undefined),
    ...overrides,
  };
}

export function createMockViewsApi(overrides?: Partial<MockViewsApi>): MockViewsApi {
  return {
    getAll: vi.fn().mockResolvedValue([]),
    getById: vi.fn().mockResolvedValue(null),
    create: vi.fn().mockResolvedValue({ id: 'view-1', name: 'Test View', components: [] }),
    delete: vi.fn().mockResolvedValue(undefined),
    rename: vi.fn().mockResolvedValue(undefined),
    setDefault: vi.fn().mockResolvedValue(undefined),
    updateEdgeType: vi.fn().mockResolvedValue(undefined),
    updateColorScheme: vi.fn().mockResolvedValue(undefined),
    getComponents: vi.fn().mockResolvedValue([]),
    addComponent: vi.fn().mockResolvedValue(undefined),
    removeComponent: vi.fn().mockResolvedValue(undefined),
    updateComponentPosition: vi.fn().mockResolvedValue(undefined),
    updateComponentColor: vi.fn().mockResolvedValue(undefined),
    clearComponentColor: vi.fn().mockResolvedValue(undefined),
    updateMultiplePositions: vi.fn().mockResolvedValue(undefined),
    addCapability: vi.fn().mockResolvedValue(undefined),
    removeCapability: vi.fn().mockResolvedValue(undefined),
    updateCapabilityPosition: vi.fn().mockResolvedValue(undefined),
    updateCapabilityColor: vi.fn().mockResolvedValue(undefined),
    clearCapabilityColor: vi.fn().mockResolvedValue(undefined),
    ...overrides,
  };
}

export function createMockCapabilitiesApi(
  overrides?: Partial<MockCapabilitiesApi>
): MockCapabilitiesApi {
  return {
    getAll: vi.fn().mockResolvedValue([]),
    getById: vi.fn().mockResolvedValue(null),
    getChildren: vi.fn().mockResolvedValue([]),
    create: vi.fn().mockResolvedValue({ id: 'cap-1', name: 'Test', level: 'L1' }),
    update: vi.fn().mockResolvedValue({ id: 'cap-1', name: 'Updated' }),
    updateMetadata: vi.fn().mockResolvedValue({ id: 'cap-1' }),
    addExpert: vi.fn().mockResolvedValue(undefined),
    addTag: vi.fn().mockResolvedValue(undefined),
    delete: vi.fn().mockResolvedValue(undefined),
    changeParent: vi.fn().mockResolvedValue(undefined),
    getAllDependencies: vi.fn().mockResolvedValue([]),
    getOutgoingDependencies: vi.fn().mockResolvedValue([]),
    getIncomingDependencies: vi.fn().mockResolvedValue([]),
    createDependency: vi.fn().mockResolvedValue({ id: 'dep-1' }),
    deleteDependency: vi.fn().mockResolvedValue(undefined),
    getSystemsByCapability: vi.fn().mockResolvedValue([]),
    getCapabilitiesByComponent: vi.fn().mockResolvedValue([]),
    linkSystem: vi.fn().mockResolvedValue({ id: 'real-1' }),
    updateRealization: vi.fn().mockResolvedValue({ id: 'real-1' }),
    deleteRealization: vi.fn().mockResolvedValue(undefined),
    ...overrides,
  };
}

export function createMockBusinessDomainsApi(
  overrides?: Partial<MockBusinessDomainsApi>
): MockBusinessDomainsApi {
  return {
    getAll: vi.fn().mockResolvedValue([]),
    getById: vi.fn().mockResolvedValue(null),
    create: vi.fn().mockResolvedValue({ id: 'domain-1', name: 'Test Domain' }),
    update: vi.fn().mockResolvedValue({ id: 'domain-1', name: 'Updated' }),
    delete: vi.fn().mockResolvedValue(undefined),
    getCapabilities: vi.fn().mockResolvedValue([]),
    associateCapability: vi.fn().mockResolvedValue(undefined),
    dissociateCapability: vi.fn().mockResolvedValue(undefined),
    getCapabilityRealizations: vi.fn().mockResolvedValue([]),
    ...overrides,
  };
}

export function createMockLayoutsApi(overrides?: Partial<MockLayoutsApi>): MockLayoutsApi {
  return {
    get: vi.fn().mockResolvedValue(null),
    upsert: vi.fn().mockResolvedValue({ id: 'layout-1' }),
    delete: vi.fn().mockResolvedValue(undefined),
    updatePreferences: vi.fn().mockResolvedValue({ id: 'layout-1' }),
    upsertElement: vi.fn().mockResolvedValue({ elementId: 'elem-1', x: 0, y: 0 }),
    deleteElement: vi.fn().mockResolvedValue(undefined),
    batchUpdateElements: vi.fn().mockResolvedValue({ updated: 0, elements: [] }),
    ...overrides,
  };
}

export function createMockMetadataApi(overrides?: Partial<MockMetadataApi>): MockMetadataApi {
  return {
    getMaturityLevels: vi.fn().mockResolvedValue(['Initial', 'Developing', 'Defined', 'Managed', 'Optimizing']),
    getStatuses: vi.fn().mockResolvedValue([
      { value: 'Active', displayName: 'Active', sortOrder: 1 },
      { value: 'Planned', displayName: 'Planned', sortOrder: 2 },
      { value: 'Deprecated', displayName: 'Deprecated', sortOrder: 3 },
    ]),
    getOwnershipModels: vi.fn().mockResolvedValue([
      { value: 'Centralized', displayName: 'Centralized' },
      { value: 'Distributed', displayName: 'Distributed' },
    ]),
    getStrategyPillars: vi.fn().mockResolvedValue([
      { value: 'Growth', displayName: 'Growth' },
      { value: 'Efficiency', displayName: 'Efficiency' },
    ]),
    getVersion: vi.fn().mockResolvedValue('1.0.0'),
    getLatestRelease: vi.fn().mockResolvedValue(null),
    getReleaseByVersion: vi.fn().mockResolvedValue(null),
    getReleases: vi.fn().mockResolvedValue([]),
    ...overrides,
  };
}

export interface AllMockApis {
  componentsApi: MockComponentsApi;
  relationsApi: MockRelationsApi;
  viewsApi: MockViewsApi;
  capabilitiesApi: MockCapabilitiesApi;
  businessDomainsApi: MockBusinessDomainsApi;
  layoutsApi: MockLayoutsApi;
  metadataApi: MockMetadataApi;
}

export function createAllMockApis(overrides?: Partial<AllMockApis>): AllMockApis {
  return {
    componentsApi: createMockComponentsApi(overrides?.componentsApi),
    relationsApi: createMockRelationsApi(overrides?.relationsApi),
    viewsApi: createMockViewsApi(overrides?.viewsApi),
    capabilitiesApi: createMockCapabilitiesApi(overrides?.capabilitiesApi),
    businessDomainsApi: createMockBusinessDomainsApi(overrides?.businessDomainsApi),
    layoutsApi: createMockLayoutsApi(overrides?.layoutsApi),
    metadataApi: createMockMetadataApi(overrides?.metadataApi),
  };
}
