import { componentsApi } from '../features/components/api';
import { relationsApi } from '../features/relations/api';
import { viewsApi } from '../features/views/api';
import { capabilitiesApi } from '../features/capabilities/api';
import { businessDomainsApi } from '../features/business-domains/api';
import { layoutsApi } from '../features/canvas/api';
import { metadataApi } from './metadata';
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
  StatusOption,
  OwnershipModelOption,
  StrategyPillarOption,
  Release,
  ReleaseVersion,
  BusinessDomain,
  BusinessDomainId,
  CreateBusinessDomainRequest,
  UpdateBusinessDomainRequest,
  AssociateCapabilityRequest,
  LayoutContextType,
  LayoutContainer,
  LayoutContainerSummary,
  ElementPosition,
  UpsertLayoutRequest,
  ElementPositionInput,
  BatchUpdateItem,
  BatchUpdateResponse,
  Position,
} from './types';

class ApiClient {
  async getComponents(): Promise<Component[]> {
    return componentsApi.getAll();
  }

  async getComponentById(id: ComponentId): Promise<Component> {
    return componentsApi.getById(id);
  }

  async createComponent(request: CreateComponentRequest): Promise<Component> {
    return componentsApi.create(request);
  }

  async updateComponent(id: ComponentId, request: CreateComponentRequest): Promise<Component> {
    return componentsApi.update(id, request);
  }

  async deleteComponent(id: ComponentId): Promise<void> {
    return componentsApi.delete(id);
  }

  async getRelations(): Promise<Relation[]> {
    return relationsApi.getAll();
  }

  async getRelationById(id: RelationId): Promise<Relation> {
    return relationsApi.getById(id);
  }

  async createRelation(request: CreateRelationRequest): Promise<Relation> {
    return relationsApi.create(request);
  }

  async updateRelation(id: RelationId, request: Partial<CreateRelationRequest>): Promise<Relation> {
    return relationsApi.update(id, request);
  }

  async deleteRelation(id: RelationId): Promise<void> {
    return relationsApi.delete(id);
  }

  async getViews(): Promise<View[]> {
    return viewsApi.getAll();
  }

  async getViewById(id: ViewId): Promise<View> {
    return viewsApi.getById(id);
  }

  async createView(request: CreateViewRequest): Promise<View> {
    return viewsApi.create(request);
  }

  async getViewComponents(viewId: ViewId): Promise<ViewComponent[]> {
    return viewsApi.getComponents(viewId);
  }

  async addComponentToView(viewId: ViewId, request: AddComponentToViewRequest): Promise<void> {
    return viewsApi.addComponent(viewId, request);
  }

  async updateComponentPosition(
    viewId: ViewId,
    componentId: ComponentId,
    request: UpdatePositionRequest
  ): Promise<void> {
    return viewsApi.updateComponentPosition(viewId, componentId, request);
  }

  async updateMultiplePositions(viewId: ViewId, request: UpdateMultiplePositionsRequest): Promise<void> {
    return viewsApi.updateMultiplePositions(viewId, request);
  }

  async renameView(viewId: ViewId, request: RenameViewRequest): Promise<void> {
    return viewsApi.rename(viewId, request);
  }

  async deleteView(viewId: ViewId): Promise<void> {
    return viewsApi.delete(viewId);
  }

  async removeComponentFromView(viewId: ViewId, componentId: ComponentId): Promise<void> {
    return viewsApi.removeComponent(viewId, componentId);
  }

  async setDefaultView(viewId: ViewId): Promise<void> {
    return viewsApi.setDefault(viewId);
  }

  async updateViewEdgeType(viewId: ViewId, request: UpdateViewEdgeTypeRequest): Promise<void> {
    return viewsApi.updateEdgeType(viewId, request);
  }

  async updateViewColorScheme(viewId: ViewId, request: UpdateViewColorSchemeRequest): Promise<void> {
    return viewsApi.updateColorScheme(viewId, request);
  }

  async addCapabilityToView(viewId: ViewId, request: AddCapabilityToViewRequest): Promise<void> {
    return viewsApi.addCapability(viewId, request);
  }

  async updateCapabilityPositionInView(viewId: ViewId, capabilityId: CapabilityId, position: Position): Promise<void>;
  async updateCapabilityPositionInView(viewId: ViewId, capabilityId: CapabilityId, x: number, y: number): Promise<void>;
  async updateCapabilityPositionInView(viewId: ViewId, capabilityId: CapabilityId, xOrPosition: number | Position, y?: number): Promise<void> {
    const position = typeof xOrPosition === 'number' ? { x: xOrPosition, y: y! } : xOrPosition;
    return viewsApi.updateCapabilityPosition(viewId, capabilityId, position);
  }

  async removeCapabilityFromView(viewId: ViewId, capabilityId: CapabilityId): Promise<void> {
    return viewsApi.removeCapability(viewId, capabilityId);
  }

  async updateCapabilityColor(viewId: ViewId, capabilityId: CapabilityId, color: string): Promise<void> {
    return viewsApi.updateCapabilityColor(viewId, capabilityId, color);
  }

  async clearCapabilityColor(viewId: ViewId, capabilityId: CapabilityId): Promise<void> {
    return viewsApi.clearCapabilityColor(viewId, capabilityId);
  }

  async updateComponentColor(viewId: ViewId, componentId: ComponentId, color: string): Promise<void> {
    return viewsApi.updateComponentColor(viewId, componentId, color);
  }

  async clearComponentColor(viewId: ViewId, componentId: ComponentId): Promise<void> {
    return viewsApi.clearComponentColor(viewId, componentId);
  }

  async getCapabilities(): Promise<Capability[]> {
    return capabilitiesApi.getAll();
  }

  async getCapabilityById(id: CapabilityId): Promise<Capability> {
    return capabilitiesApi.getById(id);
  }

  async getCapabilityChildren(id: CapabilityId): Promise<Capability[]> {
    return capabilitiesApi.getChildren(id);
  }

  async createCapability(request: CreateCapabilityRequest): Promise<Capability> {
    return capabilitiesApi.create(request);
  }

  async updateCapability(id: CapabilityId, request: UpdateCapabilityRequest): Promise<Capability> {
    return capabilitiesApi.update(id, request);
  }

  async updateCapabilityMetadata(id: CapabilityId, request: UpdateCapabilityMetadataRequest): Promise<Capability> {
    return capabilitiesApi.updateMetadata(id, request);
  }

  async addCapabilityExpert(id: CapabilityId, request: AddCapabilityExpertRequest): Promise<void> {
    return capabilitiesApi.addExpert(id, request);
  }

  async addCapabilityTag(id: CapabilityId, request: AddCapabilityTagRequest): Promise<void> {
    return capabilitiesApi.addTag(id, request);
  }

  async deleteCapability(id: CapabilityId): Promise<void> {
    return capabilitiesApi.delete(id);
  }

  async changeCapabilityParent(id: CapabilityId, parentId: CapabilityId | null): Promise<void> {
    return capabilitiesApi.changeParent(id, parentId);
  }

  async getCapabilityDependencies(): Promise<CapabilityDependency[]> {
    return capabilitiesApi.getAllDependencies();
  }

  async getOutgoingDependencies(capabilityId: CapabilityId): Promise<CapabilityDependency[]> {
    return capabilitiesApi.getOutgoingDependencies(capabilityId);
  }

  async getIncomingDependencies(capabilityId: CapabilityId): Promise<CapabilityDependency[]> {
    return capabilitiesApi.getIncomingDependencies(capabilityId);
  }

  async createCapabilityDependency(request: CreateCapabilityDependencyRequest): Promise<CapabilityDependency> {
    return capabilitiesApi.createDependency(request);
  }

  async deleteCapabilityDependency(id: CapabilityDependencyId): Promise<void> {
    return capabilitiesApi.deleteDependency(id);
  }

  async getSystemsByCapability(capabilityId: CapabilityId): Promise<CapabilityRealization[]> {
    return capabilitiesApi.getSystemsByCapability(capabilityId);
  }

  async getCapabilitiesByComponent(componentId: ComponentId): Promise<CapabilityRealization[]> {
    return capabilitiesApi.getCapabilitiesByComponent(componentId);
  }

  async linkSystemToCapability(capabilityId: CapabilityId, request: LinkSystemToCapabilityRequest): Promise<CapabilityRealization> {
    return capabilitiesApi.linkSystem(capabilityId, request);
  }

  async updateRealization(id: RealizationId, request: UpdateRealizationRequest): Promise<CapabilityRealization> {
    return capabilitiesApi.updateRealization(id, request);
  }

  async deleteRealization(id: RealizationId): Promise<void> {
    return capabilitiesApi.deleteRealization(id);
  }

  async getMaturityLevels(): Promise<string[]> {
    return metadataApi.getMaturityLevels();
  }

  async getStatuses(): Promise<StatusOption[]> {
    return metadataApi.getStatuses();
  }

  async getOwnershipModels(): Promise<OwnershipModelOption[]> {
    return metadataApi.getOwnershipModels();
  }

  async getStrategyPillars(): Promise<StrategyPillarOption[]> {
    return metadataApi.getStrategyPillars();
  }

  async getVersion(): Promise<string> {
    return metadataApi.getVersion();
  }

  async getLatestRelease(): Promise<Release | null> {
    return metadataApi.getLatestRelease();
  }

  async getReleaseByVersion(version: ReleaseVersion): Promise<Release | null> {
    return metadataApi.getReleaseByVersion(version);
  }

  async getReleases(): Promise<Release[]> {
    return metadataApi.getReleases();
  }

  async getBusinessDomains(): Promise<BusinessDomain[]> {
    return businessDomainsApi.getAll();
  }

  async getBusinessDomainById(id: BusinessDomainId): Promise<BusinessDomain> {
    return businessDomainsApi.getById(id);
  }

  async createBusinessDomain(request: CreateBusinessDomainRequest): Promise<BusinessDomain> {
    return businessDomainsApi.create(request);
  }

  async updateBusinessDomain(id: BusinessDomainId, request: UpdateBusinessDomainRequest): Promise<BusinessDomain> {
    return businessDomainsApi.update(id, request);
  }

  async deleteBusinessDomain(id: BusinessDomainId): Promise<void> {
    return businessDomainsApi.delete(id);
  }

  async getDomainCapabilities(capabilitiesLink: string): Promise<Capability[]> {
    return businessDomainsApi.getCapabilities(capabilitiesLink);
  }

  async associateCapabilityWithDomain(associateLink: string, request: AssociateCapabilityRequest): Promise<void> {
    return businessDomainsApi.associateCapability(associateLink, request);
  }

  async dissociateCapabilityFromDomain(dissociateLink: string): Promise<void> {
    return businessDomainsApi.dissociateCapability(dissociateLink);
  }

  async getCapabilityRealizationsByDomain(
    domainId: BusinessDomainId,
    depth: number = 4
  ): Promise<CapabilityRealizationsGroup[]> {
    return businessDomainsApi.getCapabilityRealizations(domainId, depth);
  }

  async getLayout(contextType: LayoutContextType, contextRef: string): Promise<LayoutContainer | null> {
    return layoutsApi.get(contextType, contextRef);
  }

  async upsertLayout(
    contextType: LayoutContextType,
    contextRef: string,
    request: UpsertLayoutRequest = {}
  ): Promise<LayoutContainer> {
    return layoutsApi.upsert(contextType, contextRef, request);
  }

  async deleteLayout(contextType: LayoutContextType, contextRef: string): Promise<void> {
    return layoutsApi.delete(contextType, contextRef);
  }

  async updateLayoutPreferences(
    contextType: LayoutContextType,
    contextRef: string,
    preferences: Record<string, unknown>,
    version: number
  ): Promise<LayoutContainerSummary> {
    return layoutsApi.updatePreferences(contextType, contextRef, preferences, version);
  }

  async upsertElementPosition(
    contextType: LayoutContextType,
    contextRef: string,
    elementId: string,
    position: ElementPositionInput
  ): Promise<ElementPosition> {
    return layoutsApi.upsertElement(contextType, contextRef, elementId, position);
  }

  async deleteElementPosition(
    contextType: LayoutContextType,
    contextRef: string,
    elementId: string
  ): Promise<void> {
    return layoutsApi.deleteElement(contextType, contextRef, elementId);
  }

  async batchUpdateElements(
    contextType: LayoutContextType,
    contextRef: string,
    updates: BatchUpdateItem[]
  ): Promise<BatchUpdateResponse> {
    return layoutsApi.batchUpdateElements(contextType, contextRef, updates);
  }
}

export const apiClient = new ApiClient();
export default apiClient;
