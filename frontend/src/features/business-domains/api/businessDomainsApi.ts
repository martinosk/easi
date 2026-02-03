import { httpClient } from '../../../api/core';
import { followLink } from '../../../utils/hateoas';
import type {
  BusinessDomain,
  BusinessDomainId,
  Capability,
  CapabilityRealizationsGroup,
  CreateBusinessDomainRequest,
  UpdateBusinessDomainRequest,
  AssociateCapabilityRequest,
  BusinessDomainsResponse,
  CollectionResponse,
} from '../../../api/types';

export const businessDomainsApi = {
  async getAll(): Promise<BusinessDomainsResponse> {
    const response = await httpClient.get<BusinessDomainsResponse>('/api/v1/business-domains');
    return response.data;
  },

  async getById(id: BusinessDomainId): Promise<BusinessDomain> {
    const response = await httpClient.get<BusinessDomain>(`/api/v1/business-domains/${id}`);
    return response.data;
  },

  async create(request: CreateBusinessDomainRequest): Promise<BusinessDomain> {
    const response = await httpClient.post<BusinessDomain>('/api/v1/business-domains', request);
    return response.data;
  },

  async update(domain: BusinessDomain, request: UpdateBusinessDomainRequest): Promise<BusinessDomain> {
    const response = await httpClient.put<BusinessDomain>(followLink(domain, 'edit'), request);
    return response.data;
  },

  async delete(domain: BusinessDomain): Promise<void> {
    await httpClient.delete(followLink(domain, 'delete'));
  },

  async getCapabilities(capabilitiesLink: string): Promise<Capability[]> {
    const response = await httpClient.get<CollectionResponse<Capability>>(capabilitiesLink);
    return response.data.data || [];
  },

  async getCapabilitiesByDomainId(domainId: BusinessDomainId): Promise<Capability[]> {
    const response = await httpClient.get<CollectionResponse<Capability>>(
      `/api/v1/business-domains/${domainId}/capabilities`
    );
    return response.data.data || [];
  },

  async associateCapability(associateLink: string, request: AssociateCapabilityRequest): Promise<void> {
    await httpClient.post(associateLink, request);
  },

  async associateCapabilityByDomainId(domainId: BusinessDomainId, request: AssociateCapabilityRequest): Promise<void> {
    await httpClient.post(`/api/v1/business-domains/${domainId}/capabilities`, request);
  },

  async dissociateCapability(dissociateLink: string): Promise<void> {
    await httpClient.delete(dissociateLink);
  },

  async dissociateCapabilityByDomainId(domainId: BusinessDomainId, capabilityId: string): Promise<void> {
    await httpClient.delete(`/api/v1/business-domains/${domainId}/capabilities/${capabilityId}`);
  },

  async getCapabilityRealizations(
    domainId: BusinessDomainId,
    depth: number = 4
  ): Promise<CapabilityRealizationsGroup[]> {
    const response = await httpClient.get<CollectionResponse<CapabilityRealizationsGroup>>(
      `/api/v1/business-domains/${domainId}/capability-realizations?depth=${depth}`
    );
    return response.data.data || [];
  },
};

export default businessDomainsApi;
