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
export type LayoutContainerId = Branded<string, 'LayoutContainerId'>;
export type StrategyImportanceId = Branded<string, 'StrategyImportanceId'>;
export type AcquiredEntityId = Branded<string, 'AcquiredEntityId'>;
export type VendorId = Branded<string, 'VendorId'>;
export type InternalTeamId = Branded<string, 'InternalTeamId'>;
export type OriginRelationshipId = Branded<string, 'OriginRelationshipId'>;
export type ValueStreamId = Branded<string, 'ValueStreamId'>;
export type StageId = Branded<string, 'StageId'>;

function isNonEmptyString(value: unknown): value is string {
  return typeof value === 'string' && value.length > 0;
}

function createBrandedFactory<T extends string>(typeName: string) {
  return (value: unknown): T => {
    if (!isNonEmptyString(value)) {
      throw new Error(`Invalid ${typeName}: expected non-empty string, got ${typeof value}`);
    }
    return value as T;
  };
}

function createBrandedTypeGuard<T extends string>() {
  return (value: unknown): value is T => isNonEmptyString(value);
}

export const toComponentId = createBrandedFactory<ComponentId>('ComponentId');
export const toRelationId = createBrandedFactory<RelationId>('RelationId');
export const toViewId = createBrandedFactory<ViewId>('ViewId');
export const toCapabilityId = createBrandedFactory<CapabilityId>('CapabilityId');
export const toCapabilityDependencyId = createBrandedFactory<CapabilityDependencyId>('CapabilityDependencyId');
export const toRealizationId = createBrandedFactory<RealizationId>('RealizationId');
export const toReleaseVersion = createBrandedFactory<ReleaseVersion>('ReleaseVersion');
export const toBusinessDomainId = createBrandedFactory<BusinessDomainId>('BusinessDomainId');
export const toEnterpriseCapabilityId = createBrandedFactory<EnterpriseCapabilityId>('EnterpriseCapabilityId');
export const toEnterpriseCapabilityLinkId = createBrandedFactory<EnterpriseCapabilityLinkId>('EnterpriseCapabilityLinkId');
export const toEnterpriseStrategicImportanceId = createBrandedFactory<EnterpriseStrategicImportanceId>('EnterpriseStrategicImportanceId');
export const toLayoutContainerId = createBrandedFactory<LayoutContainerId>('LayoutContainerId');
export const toStrategyImportanceId = createBrandedFactory<StrategyImportanceId>('StrategyImportanceId');
export const toAcquiredEntityId = createBrandedFactory<AcquiredEntityId>('AcquiredEntityId');
export const toVendorId = createBrandedFactory<VendorId>('VendorId');
export const toInternalTeamId = createBrandedFactory<InternalTeamId>('InternalTeamId');
export const toOriginRelationshipId = createBrandedFactory<OriginRelationshipId>('OriginRelationshipId');
export const toValueStreamId = createBrandedFactory<ValueStreamId>('ValueStreamId');
export const toStageId = createBrandedFactory<StageId>('StageId');

export const isComponentId = createBrandedTypeGuard<ComponentId>();
export const isRelationId = createBrandedTypeGuard<RelationId>();
export const isViewId = createBrandedTypeGuard<ViewId>();
export const isCapabilityId = createBrandedTypeGuard<CapabilityId>();
export const isCapabilityDependencyId = createBrandedTypeGuard<CapabilityDependencyId>();
export const isRealizationId = createBrandedTypeGuard<RealizationId>();
export const isReleaseVersion = createBrandedTypeGuard<ReleaseVersion>();
export const isBusinessDomainId = createBrandedTypeGuard<BusinessDomainId>();
export const isEnterpriseCapabilityId = createBrandedTypeGuard<EnterpriseCapabilityId>();
export const isEnterpriseCapabilityLinkId = createBrandedTypeGuard<EnterpriseCapabilityLinkId>();
export const isEnterpriseStrategicImportanceId = createBrandedTypeGuard<EnterpriseStrategicImportanceId>();
export const isLayoutContainerId = createBrandedTypeGuard<LayoutContainerId>();
export const isStrategyImportanceId = createBrandedTypeGuard<StrategyImportanceId>();
export const isAcquiredEntityId = createBrandedTypeGuard<AcquiredEntityId>();
export const isVendorId = createBrandedTypeGuard<VendorId>();
export const isInternalTeamId = createBrandedTypeGuard<InternalTeamId>();
export const isOriginRelationshipId = createBrandedTypeGuard<OriginRelationshipId>();
export const isValueStreamId = createBrandedTypeGuard<ValueStreamId>();
export const isStageId = createBrandedTypeGuard<StageId>();

export interface Position {
  x: number;
  y: number;
}

export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';

export interface HATEOASLink {
  href: string;
  method: HttpMethod;
}

export interface HATEOASLinks {
  self?: HATEOASLink;
  edit?: HATEOASLink;
  delete?: HATEOASLink;
  create?: HATEOASLink;
  collection?: HATEOASLink;
  up?: HATEOASLink;
  describedby?: HATEOASLink;
  next?: HATEOASLink;
  [key: string]: HATEOASLink | undefined;
}

export interface Component {
  id: ComponentId;
  name: string;
  description?: string;
  experts?: Expert[];
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
  _links?: HATEOASLinks;
}

export interface ViewCapability {
  capabilityId: CapabilityId;
  x: number;
  y: number;
  customColor?: string;
  _links?: HATEOASLinks;
}

export interface ViewOriginEntity {
  originEntityId: string;
  x: number;
  y: number;
  _links?: HATEOASLinks;
}

export interface View {
  id: ViewId;
  name: string;
  description?: string;
  isDefault: boolean;
  isPrivate: boolean;
  ownerUserId?: string;
  ownerEmail?: string;
  components: ViewComponent[];
  capabilities: ViewCapability[];
  originEntities: ViewOriginEntity[];
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

export interface AddOriginEntityToViewRequest {
  originEntityId: string;
  x: number;
  y: number;
}

export interface CreateComponentRequest {
  name: string;
  description?: string;
}

export interface AddComponentExpertRequest {
  expertName: string;
  expertRole: string;
  contactInfo: string;
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
  _links?: HATEOASLinks;
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


export interface MaturityScaleSection {
  name: string;
  order: number;
  minValue: number;
  maxValue: number;
}

export interface MaturityBounds {
  min: number;
  max: number;
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
  domainArchitectId?: string;
  capabilityCount: number;
  createdAt: string;
  updatedAt?: string;
  _links: HATEOASLinks;
}

export interface CreateBusinessDomainRequest {
  name: string;
  description?: string;
  domainArchitectId?: string;
}

export interface UpdateBusinessDomainRequest {
  name: string;
  description?: string;
  domainArchitectId?: string;
}

export interface AssociateCapabilityRequest {
  capabilityId: CapabilityId;
}

export type BusinessDomainsResponse = CollectionResponse<BusinessDomain>;

export interface ValueStream {
  id: ValueStreamId;
  name: string;
  description: string;
  stageCount: number;
  createdAt: string;
  updatedAt?: string;
  _links: HATEOASLinks;
}

export interface CreateValueStreamRequest {
  name: string;
  description?: string;
}

export interface UpdateValueStreamRequest {
  name: string;
  description?: string;
}

export type ValueStreamsResponse = CollectionResponse<ValueStream>;

export interface ValueStreamStage {
  id: StageId;
  valueStreamId: ValueStreamId;
  name: string;
  description?: string;
  position: number;
  _links?: HATEOASLinks;
}

export interface StageCapabilityMapping {
  stageId: StageId;
  capabilityId: string;
  capabilityName?: string;
  _links?: HATEOASLinks;
}

export interface ValueStreamDetail extends ValueStream {
  stages: ValueStreamStage[];
  stageCapabilities: StageCapabilityMapping[];
}

export interface CreateStageRequest {
  name: string;
  description?: string;
  position?: number;
}

export interface UpdateStageRequest {
  name: string;
  description?: string;
}

export interface ReorderStagesRequest {
  positions: { stageId: string; position: number }[];
}

export interface AddStageCapabilityRequest {
  capabilityId: string;
}

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

export type FitType = 'TECHNICAL' | 'FUNCTIONAL' | '';

export interface StrategyPillar {
  id: string;
  name: string;
  description: string;
  active: boolean;
  fitScoringEnabled: boolean;
  fitCriteria: string;
  fitType: FitType;
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

export interface ApplicationFitScore {
  id: string;
  componentId: ComponentId;
  componentName: string;
  pillarId: string;
  pillarName: string;
  score: number;
  scoreLabel: string;
  rationale?: string;
  scoredAt: string;
  scoredBy: string;
  _links: HATEOASLinks;
}

export interface ApplicationFitScoresResponse {
  data: ApplicationFitScore[];
  _links: HATEOASLinks;
}

export interface SetApplicationFitScoreRequest {
  score: number;
  rationale?: string;
}

export type FitCategory = 'liability' | 'concern' | 'aligned';

export interface FitComparison {
  pillarId: string;
  pillarName: string;
  fitScore: number;
  fitScoreLabel: string;
  importance: number;
  importanceLabel: string;
  gap: number;
  category: FitCategory;
  fitRationale?: string;
}

export interface FitComparisonsResponse {
  data: FitComparison[];
  _links: HATEOASLinks;
}

export interface RealizationFit {
  realizationId: string;
  componentId: ComponentId;
  componentName: string;
  capabilityId: CapabilityId;
  capabilityName: string;
  businessDomainId?: BusinessDomainId;
  businessDomainName?: string;
  importance: number;
  importanceLabel: string;
  importanceSourceCapabilityId?: CapabilityId;
  importanceSourceCapabilityName?: string;
  isImportanceInherited: boolean;
  importanceRationale?: string;
  fitScore: number;
  fitScoreLabel: string;
  gap: number;
  fitRationale?: string;
  category: FitCategory;
}

export interface StrategicFitSummary {
  totalRealizations: number;
  scoredRealizations: number;
  liabilityCount: number;
  concernCount: number;
  alignedCount: number;
  averageGap: number;
}

export interface StrategicFitAnalysis {
  pillarId: string;
  pillarName: string;
  summary: StrategicFitSummary;
  liabilities: RealizationFit[];
  concerns: RealizationFit[];
  aligned: RealizationFit[];
  _links: HATEOASLinks;
}

export interface AuditEntry {
  eventId: number;
  aggregateId: string;
  eventType: string;
  displayName: string;
  eventData: Record<string, unknown>;
  occurredAt: string;
  version: number;
  actorId: string;
  actorEmail: string;
}

export interface AuditPaginationInfo {
  hasMore: boolean;
  nextCursor?: string;
}

export interface AuditHistoryResponse {
  entries: AuditEntry[];
  pagination?: AuditPaginationInfo;
  _links: HATEOASLinks;
}

export type IntegrationStatus = 'NOT_STARTED' | 'IN_PROGRESS' | 'COMPLETED';
export type OriginRelationshipType = 'AcquiredVia' | 'PurchasedFrom' | 'BuiltBy';

export interface AcquiredEntity {
  id: AcquiredEntityId;
  name: string;
  acquisitionDate?: string;
  integrationStatus: IntegrationStatus;
  notes?: string;
  componentCount: number;
  createdAt: string;
  updatedAt?: string;
  _links: HATEOASLinks;
}

export interface CreateAcquiredEntityRequest {
  name: string;
  acquisitionDate?: string;
  integrationStatus?: IntegrationStatus;
  notes?: string;
}

export interface UpdateAcquiredEntityRequest {
  name: string;
  acquisitionDate?: string;
  integrationStatus?: IntegrationStatus;
  notes?: string;
}

export interface Vendor {
  id: VendorId;
  name: string;
  implementationPartner?: string;
  notes?: string;
  componentCount: number;
  createdAt: string;
  updatedAt?: string;
  _links: HATEOASLinks;
}

export interface CreateVendorRequest {
  name: string;
  implementationPartner?: string;
  notes?: string;
}

export interface UpdateVendorRequest {
  name: string;
  implementationPartner?: string;
  notes?: string;
}

export interface InternalTeam {
  id: InternalTeamId;
  name: string;
  department?: string;
  contactPerson?: string;
  notes?: string;
  componentCount: number;
  createdAt: string;
  updatedAt?: string;
  _links: HATEOASLinks;
}

export interface CreateInternalTeamRequest {
  name: string;
  department?: string;
  contactPerson?: string;
  notes?: string;
}

export interface UpdateInternalTeamRequest {
  name: string;
  department?: string;
  contactPerson?: string;
  notes?: string;
}

export interface OriginRelationship {
  id: OriginRelationshipId;
  componentId: ComponentId;
  componentName: string;
  relationshipType: OriginRelationshipType;
  originEntityId: string;
  originEntityName: string;
  notes?: string;
  createdAt: string;
  _links: HATEOASLinks;
}

export interface CreateOriginRelationshipRequest {
  componentId: ComponentId;
  notes?: string;
}

export type AcquiredEntitiesResponse = CollectionResponse<AcquiredEntity>;
export type VendorsResponse = CollectionResponse<Vendor>;
export type InternalTeamsResponse = CollectionResponse<InternalTeam>;
export type OriginRelationshipsResponse = CollectionResponse<OriginRelationship>;

export interface AcquiredViaRelationshipDTO {
  id: string;
  acquiredEntityId: string;
  acquiredEntityName: string;
  componentId: string;
  componentName: string;
  notes?: string;
  createdAt: string;
  _links: HATEOASLinks;
}

export interface PurchasedFromRelationshipDTO {
  id: string;
  vendorId: string;
  vendorName: string;
  componentId: string;
  componentName: string;
  notes?: string;
  createdAt: string;
  _links: HATEOASLinks;
}

export interface BuiltByRelationshipDTO {
  id: string;
  internalTeamId: string;
  internalTeamName: string;
  componentId: string;
  componentName: string;
  notes?: string;
  createdAt: string;
  _links: HATEOASLinks;
}

export interface AllOriginRelationshipsResponse {
  acquiredVia: AcquiredViaRelationshipDTO[];
  purchasedFrom: PurchasedFromRelationshipDTO[];
  builtBy: BuiltByRelationshipDTO[];
  _links: HATEOASLinks;
}

export interface RelationshipConflictError {
  error: string;
  existingRelationshipId: string;
  componentId: string;
  originEntityId: string;
  originEntityName: string;
  relationshipType: string;
}
