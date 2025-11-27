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

export interface Position {
  x: number;
  y: number;
}

export interface HATEOASLink {
  href: string;
}

export interface HATEOASLinks {
  self?: HATEOASLink;
  next?: HATEOASLink;
  archimate?: HATEOASLink;
  update?: HATEOASLink;
  delete?: HATEOASLink;
  components?: HATEOASLink;
  relations?: HATEOASLink;
  views?: HATEOASLink;
  addComponent?: HATEOASLink;
  updatePosition?: HATEOASLink;
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
  layoutDirection?: string;
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

export interface UpdateViewLayoutDirectionRequest {
  layoutDirection: string;
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

export interface Capability {
  id: CapabilityId;
  name: string;
  description?: string;
  parentId?: CapabilityId;
  level: CapabilityLevel;
  strategyPillar?: string;
  pillarWeight?: number;
  maturityLevel?: string;
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
  realizationLevel: RealizationLevel;
  notes?: string;
  origin: string;
  sourceRealizationId?: RealizationId;
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
  maturityLevel: string;
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
