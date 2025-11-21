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
