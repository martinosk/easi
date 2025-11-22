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
  id: string;
  name: string;
  description?: string;
  createdAt: string;
  _links: HATEOASLinks;
}

export interface Relation {
  id: string;
  sourceComponentId: string;
  targetComponentId: string;
  relationType: 'Triggers' | 'Serves';
  name?: string;
  description?: string;
  createdAt: string;
  _links: HATEOASLinks;
}

export interface ViewComponent {
  componentId: string;
  x: number;
  y: number;
}

export interface View {
  id: string;
  name: string;
  description?: string;
  isDefault: boolean;
  components: ViewComponent[];
  edgeType?: string;
  layoutDirection?: string;
  createdAt: string;
  _links: HATEOASLinks;
}

export interface CreateComponentRequest {
  name: string;
  description?: string;
}

export interface CreateRelationRequest {
  sourceComponentId: string;
  targetComponentId: string;
  relationType: 'Triggers' | 'Serves';
  name?: string;
  description?: string;
}

export interface CreateViewRequest {
  name: string;
  description?: string;
}

export interface AddComponentToViewRequest {
  componentId: string;
  x: number;
  y: number;
}

export interface UpdatePositionRequest {
  x: number;
  y: number;
}

export interface PositionUpdate {
  componentId: string;
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
  id: string;
  name: string;
  description?: string;
  parentId?: string;
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
  id: string;
  sourceCapabilityId: string;
  targetCapabilityId: string;
  dependencyType: DependencyType;
  description?: string;
  createdAt: string;
  _links: HATEOASLinks;
}

export interface CapabilityRealization {
  id: string;
  capabilityId: string;
  componentId: string;
  realizationLevel: RealizationLevel;
  notes?: string;
  linkedAt: string;
  _links: HATEOASLinks;
}

export interface CreateCapabilityRequest {
  name: string;
  description?: string;
  parentId?: string;
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
  sourceCapabilityId: string;
  targetCapabilityId: string;
  dependencyType: DependencyType;
  description?: string;
}

export interface LinkSystemToCapabilityRequest {
  componentId: string;
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

export type MaturityLevelsResponse = CollectionResponse<MaturityLevelOption>;
