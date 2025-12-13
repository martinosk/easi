import axios, { type AxiosError, type AxiosInstance } from 'axios';
import type {
  Component,
  ComponentId,
  Relation,
  RelationId,
  View,
  ViewId,
  ViewComponent,
  CreateComponentRequest,
  CreateRelationRequest,
  CreateViewRequest,
  AddComponentToViewRequest,
  AddCapabilityToViewRequest,
  UpdatePositionRequest,
  UpdateMultiplePositionsRequest,
  RenameViewRequest,
  UpdateViewEdgeTypeRequest,
  UpdateViewColorSchemeRequest,
  PaginatedResponse,
  CollectionResponse,
  Capability,
  CapabilityId,
  CapabilityDependency,
  CapabilityDependencyId,
  CapabilityRealization,
  CapabilityRealizationsGroup,
  RealizationId,
  CreateCapabilityRequest,
  UpdateCapabilityRequest,
  UpdateCapabilityMetadataRequest,
  AddCapabilityExpertRequest,
  AddCapabilityTagRequest,
  CreateCapabilityDependencyRequest,
  LinkSystemToCapabilityRequest,
  UpdateRealizationRequest,
  MaturityLevelsResponse,
  StatusOption,
  StatusesResponse,
  OwnershipModelOption,
  OwnershipModelsResponse,
  StrategyPillarOption,
  StrategyPillarsResponse,
  VersionResponse,
  Release,
  ReleasesResponse,
  ReleaseVersion,
  Position,
  BusinessDomain,
  BusinessDomainId,
  CreateBusinessDomainRequest,
  UpdateBusinessDomainRequest,
  AssociateCapabilityRequest,
  BusinessDomainsResponse,
  LayoutContextType,
  LayoutContainer,
  LayoutContainerSummary,
  ElementPosition,
  UpsertLayoutRequest,
  ElementPositionInput,
  BatchUpdateItem,
  BatchUpdateResponse,
} from './types';
import { ApiError } from './types';

let isRedirectingToLogin = false;

class ApiClient {
  private client: AxiosInstance;

  constructor(baseURL: string = import.meta.env.VITE_API_URL || 'http://localhost:8080') {
    this.client = axios.create({
      baseURL,
      headers: {
        'Content-Type': 'application/json',
      },
      withCredentials: true,
    });

    this.client.interceptors.response.use(
      (response) => response,
      (error: AxiosError) => {
        const statusCode = error.response?.status || 500;

        if (statusCode === 401 && !window.location.pathname.endsWith('/login') && !isRedirectingToLogin) {
          isRedirectingToLogin = true;
          const basePath = import.meta.env.BASE_URL || '/';
          window.location.href = `${basePath}login`;
          return Promise.reject(error);
        }

        const message = this.extractErrorMessage(error);
        throw new ApiError(message, statusCode, error.response?.data);
      }
    );
  }

  private extractErrorMessage(error: AxiosError): string {
    const responseMessage = this.extractResponseMessage(error.response?.data);
    return responseMessage ?? error.message ?? 'An unknown error occurred';
  }

  private extractResponseMessage(data: unknown): string | null {
    if (!data || typeof data !== 'object') return null;

    const errorData = data as { message?: string; error?: string; details?: Record<string, string> };
    if (errorData.message) return errorData.message;
    if (errorData.error) return errorData.error;

    const detailMessages = errorData.details ? Object.values(errorData.details).filter(Boolean) : [];
    return detailMessages.length > 0 ? detailMessages.join(', ') : 'An error occurred';
  }

  async getComponents(): Promise<Component[]> {
    const response = await this.client.get<PaginatedResponse<Component>>('/api/v1/components');
    return response.data.data || [];
  }

  async getComponentById(id: ComponentId): Promise<Component> {
    const response = await this.client.get<Component>(`/api/v1/components/${id}`);
    return response.data;
  }

  async createComponent(request: CreateComponentRequest): Promise<Component> {
    const response = await this.client.post<Component>('/api/v1/components', request);
    return response.data;
  }

  async updateComponent(id: ComponentId, request: CreateComponentRequest): Promise<Component> {
    const response = await this.client.put<Component>(`/api/v1/components/${id}`, request);
    return response.data;
  }

  async deleteComponent(id: ComponentId): Promise<void> {
    await this.client.delete(`/api/v1/components/${id}`);
  }

  async getRelations(): Promise<Relation[]> {
    const response = await this.client.get<PaginatedResponse<Relation>>('/api/v1/relations');
    return response.data.data || [];
  }

  async getRelationById(id: RelationId): Promise<Relation> {
    const response = await this.client.get<Relation>(`/api/v1/relations/${id}`);
    return response.data;
  }

  async createRelation(request: CreateRelationRequest): Promise<Relation> {
    const response = await this.client.post<Relation>('/api/v1/relations', request);
    return response.data;
  }

  async updateRelation(id: RelationId, request: Partial<CreateRelationRequest>): Promise<Relation> {
    const response = await this.client.put<Relation>(`/api/v1/relations/${id}`, request);
    return response.data;
  }

  async deleteRelation(id: RelationId): Promise<void> {
    await this.client.delete(`/api/v1/relations/${id}`);
  }

  async getViews(): Promise<View[]> {
    const response = await this.client.get<PaginatedResponse<View>>('/api/v1/views');
    return response.data.data || [];
  }

  async getViewById(id: ViewId): Promise<View> {
    const response = await this.client.get<View>(`/api/v1/views/${id}`);
    return response.data;
  }

  async createView(request: CreateViewRequest): Promise<View> {
    const response = await this.client.post<View>('/api/v1/views', request);
    return response.data;
  }

  async getViewComponents(viewId: ViewId): Promise<ViewComponent[]> {
    const response = await this.client.get<CollectionResponse<ViewComponent>>(
      `/api/v1/views/${viewId}/components`
    );
    return response.data.data || [];
  }

  async addComponentToView(viewId: ViewId, request: AddComponentToViewRequest): Promise<void> {
    await this.client.post(`/api/v1/views/${viewId}/components`, request);
  }

  async updateComponentPosition(
    viewId: ViewId,
    componentId: ComponentId,
    request: UpdatePositionRequest
  ): Promise<void> {
    await this.client.patch(
      `/api/v1/views/${viewId}/components/${componentId}/position`,
      request
    );
  }

  async updateMultiplePositions(viewId: ViewId, request: UpdateMultiplePositionsRequest): Promise<void> {
    await this.client.patch(`/api/v1/views/${viewId}/layout`, request);
  }

  async renameView(viewId: ViewId, request: RenameViewRequest): Promise<void> {
    await this.client.patch(`/api/v1/views/${viewId}/name`, request);
  }

  async deleteView(viewId: ViewId): Promise<void> {
    await this.client.delete(`/api/v1/views/${viewId}`);
  }

  async removeComponentFromView(viewId: ViewId, componentId: ComponentId): Promise<void> {
    await this.client.delete(`/api/v1/views/${viewId}/components/${componentId}`);
  }

  async setDefaultView(viewId: ViewId): Promise<void> {
    await this.client.put(`/api/v1/views/${viewId}/default`);
  }

  async updateViewEdgeType(viewId: ViewId, request: UpdateViewEdgeTypeRequest): Promise<void> {
    await this.client.patch(`/api/v1/views/${viewId}/edge-type`, request);
  }

  async updateViewColorScheme(viewId: ViewId, request: UpdateViewColorSchemeRequest): Promise<void> {
    await this.client.patch(`/api/v1/views/${viewId}/color-scheme`, request);
  }

  async addCapabilityToView(viewId: ViewId, request: AddCapabilityToViewRequest): Promise<void> {
    await this.client.post(`/api/v1/views/${viewId}/capabilities`, request);
  }

  async updateCapabilityPositionInView(viewId: ViewId, capabilityId: CapabilityId, position: Position): Promise<void>;
  async updateCapabilityPositionInView(viewId: ViewId, capabilityId: CapabilityId, x: number, y: number): Promise<void>;
  async updateCapabilityPositionInView(viewId: ViewId, capabilityId: CapabilityId, xOrPosition: number | Position, y?: number): Promise<void> {
    const position = typeof xOrPosition === 'number' ? { x: xOrPosition, y: y! } : xOrPosition;
    await this.client.patch(`/api/v1/views/${viewId}/capabilities/${capabilityId}/position`, position);
  }

  async removeCapabilityFromView(viewId: ViewId, capabilityId: CapabilityId): Promise<void> {
    await this.client.delete(`/api/v1/views/${viewId}/capabilities/${capabilityId}`);
  }

  async updateCapabilityColor(viewId: ViewId, capabilityId: CapabilityId, color: string): Promise<void> {
    await this.client.patch(`/api/v1/views/${viewId}/capabilities/${capabilityId}/color`, { color });
  }

  async clearCapabilityColor(viewId: ViewId, capabilityId: CapabilityId): Promise<void> {
    await this.client.delete(`/api/v1/views/${viewId}/capabilities/${capabilityId}/color`);
  }

  async updateComponentColor(viewId: ViewId, componentId: ComponentId, color: string): Promise<void> {
    await this.client.patch(`/api/v1/views/${viewId}/components/${componentId}/color`, { color });
  }

  async clearComponentColor(viewId: ViewId, componentId: ComponentId): Promise<void> {
    await this.client.delete(`/api/v1/views/${viewId}/components/${componentId}/color`);
  }

  async getCapabilities(): Promise<Capability[]> {
    const response = await this.client.get<CollectionResponse<Capability>>('/api/v1/capabilities');
    return response.data.data || [];
  }

  async getCapabilityById(id: CapabilityId): Promise<Capability> {
    const response = await this.client.get<Capability>(`/api/v1/capabilities/${id}`);
    return response.data;
  }

  async getCapabilityChildren(id: CapabilityId): Promise<Capability[]> {
    const response = await this.client.get<CollectionResponse<Capability>>(`/api/v1/capabilities/${id}/children`);
    return response.data.data || [];
  }

  async createCapability(request: CreateCapabilityRequest): Promise<Capability> {
    const response = await this.client.post<Capability>('/api/v1/capabilities', request);
    return response.data;
  }

  async updateCapability(id: CapabilityId, request: UpdateCapabilityRequest): Promise<Capability> {
    const response = await this.client.put<Capability>(`/api/v1/capabilities/${id}`, request);
    return response.data;
  }

  async updateCapabilityMetadata(id: CapabilityId, request: UpdateCapabilityMetadataRequest): Promise<Capability> {
    const response = await this.client.put<Capability>(`/api/v1/capabilities/${id}/metadata`, request);
    return response.data;
  }

  async addCapabilityExpert(id: CapabilityId, request: AddCapabilityExpertRequest): Promise<void> {
    await this.client.post(`/api/v1/capabilities/${id}/experts`, request);
  }

  async addCapabilityTag(id: CapabilityId, request: AddCapabilityTagRequest): Promise<void> {
    await this.client.post(`/api/v1/capabilities/${id}/tags`, request);
  }

  async deleteCapability(id: CapabilityId): Promise<void> {
    await this.client.delete(`/api/v1/capabilities/${id}`);
  }

  async changeCapabilityParent(id: CapabilityId, parentId: CapabilityId | null): Promise<void> {
    await this.client.patch(`/api/v1/capabilities/${id}/parent`, {
      parentId: parentId || '',
    });
  }

  async getCapabilityDependencies(): Promise<CapabilityDependency[]> {
    const response = await this.client.get<CollectionResponse<CapabilityDependency>>('/api/v1/capability-dependencies');
    return response.data.data || [];
  }

  async getOutgoingDependencies(capabilityId: CapabilityId): Promise<CapabilityDependency[]> {
    const response = await this.client.get<CollectionResponse<CapabilityDependency>>(
      `/api/v1/capabilities/${capabilityId}/dependencies/outgoing`
    );
    return response.data.data || [];
  }

  async getIncomingDependencies(capabilityId: CapabilityId): Promise<CapabilityDependency[]> {
    const response = await this.client.get<CollectionResponse<CapabilityDependency>>(
      `/api/v1/capabilities/${capabilityId}/dependencies/incoming`
    );
    return response.data.data || [];
  }

  async createCapabilityDependency(request: CreateCapabilityDependencyRequest): Promise<CapabilityDependency> {
    const response = await this.client.post<CapabilityDependency>('/api/v1/capability-dependencies', request);
    return response.data;
  }

  async deleteCapabilityDependency(id: CapabilityDependencyId): Promise<void> {
    await this.client.delete(`/api/v1/capability-dependencies/${id}`);
  }

  async getSystemsByCapability(capabilityId: CapabilityId): Promise<CapabilityRealization[]> {
    const response = await this.client.get<CollectionResponse<CapabilityRealization>>(
      `/api/v1/capabilities/${capabilityId}/systems`
    );
    return response.data.data || [];
  }

  async getCapabilitiesByComponent(componentId: ComponentId): Promise<CapabilityRealization[]> {
    const response = await this.client.get<CollectionResponse<CapabilityRealization>>(
      `/api/v1/capability-realizations/by-component/${componentId}`
    );
    return response.data.data || [];
  }

  async linkSystemToCapability(capabilityId: CapabilityId, request: LinkSystemToCapabilityRequest): Promise<CapabilityRealization> {
    const response = await this.client.post<CapabilityRealization>(
      `/api/v1/capabilities/${capabilityId}/systems`,
      request
    );
    return response.data;
  }

  async updateRealization(id: RealizationId, request: UpdateRealizationRequest): Promise<CapabilityRealization> {
    const response = await this.client.put<CapabilityRealization>(
      `/api/v1/capability-realizations/${id}`,
      request
    );
    return response.data;
  }

  async deleteRealization(id: RealizationId): Promise<void> {
    await this.client.delete(`/api/v1/capability-realizations/${id}`);
  }

  async getMaturityLevels(): Promise<string[]> {
    const response = await this.client.get<MaturityLevelsResponse>(
      '/api/v1/capabilities/metadata/maturity-levels'
    );
    return response.data.data.map((level) => level.value);
  }

  async getStatuses(): Promise<StatusOption[]> {
    const response = await this.client.get<StatusesResponse>(
      '/api/v1/capabilities/metadata/statuses'
    );
    return response.data.data;
  }

  async getOwnershipModels(): Promise<OwnershipModelOption[]> {
    const response = await this.client.get<OwnershipModelsResponse>(
      '/api/v1/capabilities/metadata/ownership-models'
    );
    return response.data.data;
  }

  async getStrategyPillars(): Promise<StrategyPillarOption[]> {
    const response = await this.client.get<StrategyPillarsResponse>(
      '/api/v1/capabilities/metadata/strategy-pillars'
    );
    return response.data.data;
  }

  async getVersion(): Promise<string> {
    const response = await this.client.get<VersionResponse>('/api/v1/version');
    return response.data.version;
  }

  async getLatestRelease(): Promise<Release | null> {
    try {
      const response = await this.client.get<Release>('/api/v1/releases/latest');
      return response.data;
    } catch {
      return null;
    }
  }

  async getReleaseByVersion(version: ReleaseVersion): Promise<Release | null> {
    try {
      const response = await this.client.get<Release>(`/api/v1/releases/${version}`);
      return response.data;
    } catch {
      return null;
    }
  }

  async getReleases(): Promise<Release[]> {
    const response = await this.client.get<ReleasesResponse>('/api/v1/releases');
    return response.data.data || [];
  }

  async getBusinessDomains(): Promise<BusinessDomain[]> {
    const response = await this.client.get<BusinessDomainsResponse>('/api/v1/business-domains');
    return response.data.data || [];
  }

  async getBusinessDomainById(id: BusinessDomainId): Promise<BusinessDomain> {
    const response = await this.client.get<BusinessDomain>(`/api/v1/business-domains/${id}`);
    return response.data;
  }

  async createBusinessDomain(request: CreateBusinessDomainRequest): Promise<BusinessDomain> {
    const response = await this.client.post<BusinessDomain>('/api/v1/business-domains', request);
    return response.data;
  }

  async updateBusinessDomain(id: BusinessDomainId, request: UpdateBusinessDomainRequest): Promise<BusinessDomain> {
    const response = await this.client.put<BusinessDomain>(`/api/v1/business-domains/${id}`, request);
    return response.data;
  }

  async deleteBusinessDomain(id: BusinessDomainId): Promise<void> {
    await this.client.delete(`/api/v1/business-domains/${id}`);
  }

  async getDomainCapabilities(capabilitiesLink: string): Promise<Capability[]> {
    const response = await this.client.get<CollectionResponse<Capability>>(capabilitiesLink);
    return response.data.data || [];
  }

  async associateCapabilityWithDomain(associateLink: string, request: AssociateCapabilityRequest): Promise<void> {
    await this.client.post(associateLink, request);
  }

  async dissociateCapabilityFromDomain(dissociateLink: string): Promise<void> {
    await this.client.delete(dissociateLink);
  }

  async getCapabilityRealizationsByDomain(
    domainId: BusinessDomainId,
    depth: number = 4
  ): Promise<CapabilityRealizationsGroup[]> {
    const response = await this.client.get<CollectionResponse<CapabilityRealizationsGroup>>(
      `/api/v1/business-domains/${domainId}/capability-realizations?depth=${depth}`
    );
    return response.data.data || [];
  }

  async getLayout(contextType: LayoutContextType, contextRef: string): Promise<LayoutContainer | null> {
    try {
      const response = await this.client.get<LayoutContainer>(
        `/api/v1/layouts/${contextType}/${contextRef}`
      );
      return response.data;
    } catch (error) {
      if (error instanceof ApiError && error.statusCode === 404) {
        return null;
      }
      throw error;
    }
  }

  async upsertLayout(
    contextType: LayoutContextType,
    contextRef: string,
    request: UpsertLayoutRequest = {}
  ): Promise<LayoutContainer> {
    const response = await this.client.put<LayoutContainer>(
      `/api/v1/layouts/${contextType}/${contextRef}`,
      request
    );
    return response.data;
  }

  async deleteLayout(contextType: LayoutContextType, contextRef: string): Promise<void> {
    await this.client.delete(`/api/v1/layouts/${contextType}/${contextRef}`);
  }

  async updateLayoutPreferences(
    contextType: LayoutContextType,
    contextRef: string,
    preferences: Record<string, unknown>,
    version: number
  ): Promise<LayoutContainerSummary> {
    const response = await this.client.patch<LayoutContainerSummary>(
      `/api/v1/layouts/${contextType}/${contextRef}/preferences`,
      { preferences },
      {
        headers: {
          'If-Match': `"${version}"`,
        },
      }
    );
    return response.data;
  }

  async upsertElementPosition(
    contextType: LayoutContextType,
    contextRef: string,
    elementId: string,
    position: ElementPositionInput
  ): Promise<ElementPosition> {
    const response = await this.client.put<ElementPosition>(
      `/api/v1/layouts/${contextType}/${contextRef}/elements/${elementId}`,
      position
    );
    return response.data;
  }

  async deleteElementPosition(
    contextType: LayoutContextType,
    contextRef: string,
    elementId: string
  ): Promise<void> {
    await this.client.delete(
      `/api/v1/layouts/${contextType}/${contextRef}/elements/${elementId}`
    );
  }

  async batchUpdateElements(
    contextType: LayoutContextType,
    contextRef: string,
    updates: BatchUpdateItem[]
  ): Promise<BatchUpdateResponse> {
    const response = await this.client.patch<BatchUpdateResponse>(
      `/api/v1/layouts/${contextType}/${contextRef}/elements`,
      { updates }
    );
    return response.data;
  }
}

export const apiClient = new ApiClient();
export default apiClient;
