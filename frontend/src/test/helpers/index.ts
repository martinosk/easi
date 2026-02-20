export {
  renderWithProviders,
  TestProviders,
  createTestQueryClient,
  type RenderWithProvidersOptions,
} from './renderWithProviders';

export {
  buildComponent,
  buildRelation,
  buildViewComponent,
  buildViewCapability,
  buildView,
  buildExpert,
  buildCapability,
  buildCapabilityDependency,
  buildCapabilityRealization,
  buildBusinessDomain,
  buildAcquiredEntity,
  buildVendor,
  buildInternalTeam,
  buildOriginRelationship,
  resetIdCounter,
} from './entityBuilders';

export {
  createMockComponentsApi,
  createMockRelationsApi,
  createMockViewsApi,
  createMockCapabilitiesApi,
  createMockBusinessDomainsApi,
  createMockLayoutsApi,
  createMockMetadataApi,
  createAllMockApis,
  type MockComponentsApi,
  type MockRelationsApi,
  type MockViewsApi,
  type MockCapabilitiesApi,
  type MockBusinessDomainsApi,
  type MockLayoutsApi,
  type MockMetadataApi,
  type AllMockApis,
} from './mockApiClient';

export { MantineTestWrapper, createMantineTestWrapper } from './mantineTestWrapper';

export { server } from '../mocks/server';
export { resetDb, seedDb, getDb, addComponent, addCapability, addView, addRelation, addCapabilityRealization } from '../mocks/db';
export type { MockDatabase } from '../mocks/db';
