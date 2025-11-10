export interface HATEOASLink {
  href: string;
}

export interface HATEOASLinks {
  self?: HATEOASLink;
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

export interface RenameViewRequest {
  name: string;
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
