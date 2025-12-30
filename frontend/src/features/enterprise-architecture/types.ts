import type {
  EnterpriseCapabilityId,
  EnterpriseCapabilityLinkId,
  EnterpriseStrategicImportanceId,
  CapabilityId
} from '../../api/types';

export type { EnterpriseCapabilityId, EnterpriseCapabilityLinkId, EnterpriseStrategicImportanceId };

export interface EnterpriseCapability {
  id: EnterpriseCapabilityId;
  name: string;
  description: string;
  category: string;
  active: boolean;
  linkCount: number;
  domainCount: number;
  createdAt: string;
  updatedAt?: string;
  _links: {
    self: string;
    links: string;
    strategicImportance: string;
  };
}

export interface EnterpriseCapabilityLink {
  id: EnterpriseCapabilityLinkId;
  enterpriseCapabilityId: EnterpriseCapabilityId;
  domainCapabilityId: CapabilityId;
  linkedBy: string;
  linkedAt: string;
  _links: {
    self: string;
    enterpriseCapability: string;
  };
}

export interface StrategicImportance {
  id: EnterpriseStrategicImportanceId;
  enterpriseCapabilityId: EnterpriseCapabilityId;
  pillarId: string;
  pillarName: string;
  importance: number;
  rationale?: string;
  setAt: string;
  updatedAt?: string;
  _links: {
    self: string;
    enterpriseCapability: string;
  };
}

export interface CreateEnterpriseCapabilityRequest {
  name: string;
  description?: string;
  category?: string;
}

export interface UpdateEnterpriseCapabilityRequest {
  name: string;
  description?: string;
  category?: string;
}

export interface LinkCapabilityRequest {
  domainCapabilityId: CapabilityId;
}

export interface SetStrategicImportanceRequest {
  pillarId: string;
  pillarName: string;
  importance: number;
  rationale?: string;
}

export interface UpdateStrategicImportanceRequest {
  importance: number;
  rationale?: string;
}

export interface EnterpriseCapabilitiesListResponse {
  data: EnterpriseCapability[];
  _links: {
    self: string;
  };
}

export interface DomainCapabilityLinkStatus {
  linked: boolean;
  enterpriseCapabilityId?: EnterpriseCapabilityId;
  enterpriseCapabilityName?: string;
  linkId?: EnterpriseCapabilityLinkId;
  _links: {
    self: string;
    enterpriseCapability?: string;
    unlink?: string;
  };
}
