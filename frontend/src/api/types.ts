declare const __brand: unique symbol;
type Brand<B> = { [__brand]: B };
type Branded<T, B> = T & Brand<B>;

export type ComponentId = Branded<string, 'ComponentId'>;
export type RelationId = Branded<string, 'RelationId'>;
export type ViewId = Branded<string, 'ViewId'>;
export type CapabilityId = Branded<string, 'CapabilityId'>;
export type CapabilityDependencyId = Branded<string, 'CapabilityDependencyId'>;
export type RealizationId = Branded<string, 'RealizationId'>;
export type ReleaseVersion = Branded<string, 'ReleaseVersion'>;
export type BusinessDomainId = Branded<string, 'BusinessDomainId'>;
export type EnterpriseCapabilityId = Branded<string, 'EnterpriseCapabilityId'>;
export type EnterpriseCapabilityLinkId = Branded<string, 'EnterpriseCapabilityLinkId'>;
export type EnterpriseStrategicImportanceId = Branded<string, 'EnterpriseStrategicImportanceId'>;

export interface Position {
  x: number;
  y: number;
}

export interface HATEOASLinks {
  self?: string;
  next?: string;
  reference?: string;
  update?: string;
  delete?: string;
  components?: string;
  relations?: string;
  views?: string;
  addComponent?: string;
  updatePosition?: string;
  capabilities?: string;
  collection?: string;
  associate?: string;
  dissociate?: string;
  removeFromDomain?: string;
  children?: string;
  businessDomains?: string;
  [key: string]: string | undefined;
}

export interface Component {
  id: ComponentId;
  name: string;
  description?: string;
  createdAt: string;
  _links: HATEOASLinks;
}

export interface Relation {
  id: RelationId;
  sourceComponentId: ComponentId;
  targetComponentId: ComponentId;
  relationType: 'Triggers' | 'Serves';
  name?: string;
  description?: string;
  createdAt: string;
  _links: HATEOASLinks;
}

export interface ViewComponent {
  componentId: ComponentId;
  x: number;
  y: number;
  customColor?: string;
}

export interface ViewCapability {
  capabilityId: CapabilityId;
  x: number;
  y: number;
  customColor?: string;
}

export interface View {
  id: ViewId;
  name: string;
  description?: string;
  isDefault: boolean;
  components: ViewComponent[];
  capabilities: ViewCapability[];
  edgeType?: string;
  colorScheme?: string;
  createdAt: string;
  _links: HATEOASLinks;
}

export interface AddCapabilityToViewRequest {
  capabilityId: CapabilityId;
  x: number;
  y: number;
}

export interface CreateComponentRequest {
  name: string;
  description?: string;
}

export interface CreateRelationRequest {
  sourceComponentId: ComponentId;
  targetComponentId: ComponentId;
  relationType: 'Triggers' | 'Serves';
  name?: string;
  description?: string;
}

export interface CreateViewRequest {
  name: string;
  description?: string;
}

export interface AddComponentToViewRequest {
  componentId: ComponentId;
  x: number;
  y: number;
}

export interface UpdatePositionRequest {
  x: number;
  y: number;
}

export interface PositionUpdate {
  componentId: ComponentId;
  x: number;
  y: number;
}

export interface UpdateMultiplePositionsRequest {
  positions: PositionUpdate[];
}

export interface RenameViewRequest {
  name: string;
}

export interface UpdateViewEdgeTypeRequest {
  edgeType: string;
}

export interface UpdateViewColorSchemeRequest {
  colorScheme: string;
}

export interface PaginationInfo {
  hasMore: boolean;
  limit: number;
  cursor?: string;
}

export interface CollectionResponse<T> {
  data: T[];
  _links: HATEOASLinks;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: PaginationInfo;
  _links: HATEOASLinks;
}

export interface ErrorDetails {
  [key: string]: string;
}

export interface ErrorResponse {
  error: string;
  message: string;
  details?: ErrorDetails;
}

export class ApiError extends Error {
  public statusCode: number;
  public response?: unknown;

  constructor(message: string, statusCode: number, response?: unknown) {
    super(message);
    this.name = 'ApiError';
    this.statusCode = statusCode;
    this.response = response;
  }
}

export type CapabilityLevel = 'L1' | 'L2' | 'L3' | 'L4';
export type DependencyType = 'Requires' | 'Enables' | 'Supports';
export type RealizationLevel = 'Full' | 'Partial' | 'Planned';

export interface Expert {
  name: string;
  role: string;
  contact: string;
  addedAt: string;
}

export interface MaturitySection {
  name: string;
  order: number;
  minValue: number;
  maxValue: number;
}

export interface Capability {
  id: CapabilityId;
  name: string;
  description?: string;
  parentId?: CapabilityId;
  level: CapabilityLevel;
  strategyPillar?: string;
  pillarWeight?: number;
  maturityLevel?: string;
  maturityValue?: number;
  maturitySection?: MaturitySection;
  ownershipModel?: string;
  primaryOwner?: string;
  eaOwner?: string;
  status?: string;
  experts?: Expert[];
  tags?: string[];
  createdAt: string;
  _links: HATEOASLinks;
}

export interface CapabilityDependency {
  id: CapabilityDependencyId;
  sourceCapabilityId: CapabilityId;
  targetCapabilityId: CapabilityId;
  dependencyType: DependencyType;
  description?: string;
  createdAt: string;
  _links: HATEOASLinks;
}

export interface CapabilityRealization {
  id: RealizationId;
  capabilityId: CapabilityId;
  componentId: ComponentId;
  componentName?: string;
  realizationLevel: RealizationLevel;
  notes?: string;
  origin: 'Direct' | 'Inherited';
  sourceRealizationId?: RealizationId;
  sourceCapabilityId?: CapabilityId;
  sourceCapabilityName?: string;
  linkedAt: string;
  _links: HATEOASLinks;
}

export interface CreateCapabilityRequest {
  name: string;
  description?: string;
  parentId?: CapabilityId;
  level: CapabilityLevel;
}

export interface UpdateCapabilityRequest {
  name: string;
  description?: string;
}

export interface UpdateCapabilityMetadataRequest {
  strategyPillar?: string;
  pillarWeight?: number;
  maturityLevel?: string;
  maturityValue?: number;
  ownershipModel?: string;
  primaryOwner?: string;
  eaOwner?: string;
  status: string;
}

export interface AddCapabilityExpertRequest {
  expertName: string;
  expertRole: string;
  contactInfo: string;
}

export interface AddCapabilityTagRequest {
  tag: string;
}

export interface CreateCapabilityDependencyRequest {
  sourceCapabilityId: CapabilityId;
  targetCapabilityId: CapabilityId;
  dependencyType: DependencyType;
  description?: string;
}

export interface LinkSystemToCapabilityRequest {
  componentId: ComponentId;
  realizationLevel: RealizationLevel;
  notes?: string;
}

export interface UpdateRealizationRequest {
  realizationLevel: RealizationLevel;
  notes?: string;
}

export interface MaturityLevelOption {
  value: string;
  numericValue: number;
}

export interface StatusOption {
  value: string;
  displayName: string;
  sortOrder: number;
}

export interface OwnershipModelOption {
  value: string;
  displayName: string;
}

export interface StrategyPillarOption {
  value: string;
  displayName: string;
}

export interface MaturityScaleSection {
  name: string;
  order: number;
  minValue: number;
  maxValue: number;
}

export interface MaturityScale {
  sections: MaturityScaleSection[];
  version: number;
  isDefault: boolean;
  _links: HATEOASLinks;
}

export type MaturityScaleConfiguration = MaturityScale;

export interface UpdateMaturityScaleRequest {
  sections: MaturityScaleSection[];
  version: number;
}

export type MaturityLevelsResponse = CollectionResponse<MaturityLevelOption>;
export type StatusesResponse = CollectionResponse<StatusOption>;
export type OwnershipModelsResponse = CollectionResponse<OwnershipModelOption>;
export type StrategyPillarsResponse = CollectionResponse<StrategyPillarOption>;

export interface VersionResponse {
  version: string;
}

export interface Release {
  version: ReleaseVersion;
  releaseDate: string;
  notes: string;
  _links: HATEOASLinks;
}

export type ReleasesResponse = CollectionResponse<Release>;

export interface BusinessDomain {
  id: BusinessDomainId;
  name: string;
  description: string;
  capabilityCount: number;
  createdAt: string;
  updatedAt?: string;
  _links: HATEOASLinks;
}

export interface CreateBusinessDomainRequest {
  name: string;
  description?: string;
}

export interface UpdateBusinessDomainRequest {
  name: string;
  description?: string;
}

export interface AssociateCapabilityRequest {
  capabilityId: CapabilityId;
}

export type BusinessDomainsResponse = CollectionResponse<BusinessDomain>;

export type LayoutContainerId = Branded<string, 'LayoutContainerId'>;
export type LayoutContextType = 'architecture-canvas' | 'business-domain-grid';

export interface LayoutLink {
  href: string;
  method?: string;
}

export interface LayoutLinks {
  self?: LayoutLink;
  updatePreferences?: LayoutLink;
  batchUpdate?: LayoutLink;
  delete?: LayoutLink;
  layout?: LayoutLink;
  update?: LayoutLink;
}

export interface ElementPositionLinks {
  self?: LayoutLink;
  layout?: LayoutLink;
  update?: LayoutLink;
  delete?: LayoutLink;
}

export interface ElementPosition {
  elementId: string;
  x: number;
  y: number;
  width?: number;
  height?: number;
  customColor?: string;
  sortOrder?: number;
  _links: ElementPositionLinks;
}

export interface LayoutContainer {
  id: LayoutContainerId;
  contextType: LayoutContextType;
  contextRef: string;
  preferences: Record<string, unknown>;
  elements: ElementPosition[];
  version: number;
  createdAt: string;
  updatedAt: string;
  _links: LayoutLinks;
}

export interface LayoutContainerSummary {
  id: LayoutContainerId;
  contextType: LayoutContextType;
  contextRef: string;
  preferences: Record<string, unknown>;
  version: number;
  _links: LayoutLinks;
}

export interface UpsertLayoutRequest {
  preferences?: Record<string, unknown>;
}

export interface UpdatePreferencesRequest {
  preferences: Record<string, unknown>;
}

export interface ElementPositionInput {
  x: number;
  y: number;
  width?: number;
  height?: number;
  customColor?: string;
  sortOrder?: number;
}

export interface BatchUpdateItem {
  elementId: string;
  x: number;
  y: number;
  width?: number;
  height?: number;
  customColor?: string;
  sortOrder?: number;
}

export interface BatchUpdateRequest {
  updates: BatchUpdateItem[];
}

export interface BatchUpdateResponse {
  updated: number;
  elements: ElementPosition[];
  _links: LayoutLinks;
}

export interface CapabilityRealizationsGroup {
  capabilityId: CapabilityId;
  capabilityName: string;
  level: CapabilityLevel;
  realizations: CapabilityRealization[];
}

export interface StrategyPillar {
  id: string;
  name: string;
  description: string;
  active: boolean;
  _links: HATEOASLinks;
}

export interface StrategyPillarsConfiguration {
  data: StrategyPillar[];
  _links: HATEOASLinks;
}

export interface CreateStrategyPillarRequest {
  name: string;
  description: string;
}

export interface UpdateStrategyPillarRequest {
  name: string;
  description: string;
}

export type StrategyImportanceId = Branded<string, 'StrategyImportanceId'>;

export interface StrategyImportance {
  id: StrategyImportanceId;
  businessDomainId: BusinessDomainId;
  businessDomainName: string;
  capabilityId: CapabilityId;
  capabilityName: string;
  pillarId: string;
  pillarName: string;
  importance: number;
  importanceLabel: string;
  rationale?: string;
  _links: HATEOASLinks;
}

export interface SetStrategyImportanceRequest {
  pillarId: string;
  importance: number;
  rationale?: string;
}

export interface UpdateStrategyImportanceRequest {
  importance: number;
  rationale?: string;
}
