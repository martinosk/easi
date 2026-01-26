import type {
  EnterpriseCapabilityId,
  EnterpriseCapabilityLinkId,
  EnterpriseStrategicImportanceId,
  CapabilityId,
  HATEOASLink
} from '../../api/types';

export type { EnterpriseCapabilityId, EnterpriseCapabilityLinkId, EnterpriseStrategicImportanceId };

export interface EnterpriseCapability {
  id: EnterpriseCapabilityId;
  name: string;
  description: string;
  category: string;
  active: boolean;
  targetMaturity?: number;
  linkCount: number;
  domainCount: number;
  createdAt: string;
  updatedAt?: string;
  _links: {
    self: HATEOASLink;
    edit?: HATEOASLink;
    delete?: HATEOASLink;
    'x-links': HATEOASLink;
    'x-create-link'?: HATEOASLink;
    'x-strategic-importance': HATEOASLink;
  };
}

export interface EnterpriseCapabilityLink {
  id: EnterpriseCapabilityLinkId;
  enterpriseCapabilityId: EnterpriseCapabilityId;
  domainCapabilityId: CapabilityId;
  domainCapabilityName?: string;
  businessDomainId?: string;
  businessDomainName?: string;
  linkedBy: string;
  linkedAt: string;
  _links: {
    self: HATEOASLink;
    up: HATEOASLink;
    delete?: HATEOASLink;
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
    self: HATEOASLink;
    up: HATEOASLink;
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
    self: HATEOASLink;
  };
}

export interface DomainCapabilityLinkStatus {
  linked: boolean;
  enterpriseCapabilityId?: EnterpriseCapabilityId;
  enterpriseCapabilityName?: string;
  linkId?: EnterpriseCapabilityLinkId;
  _links: {
    self: HATEOASLink;
    up?: HATEOASLink;
    'x-unlink'?: HATEOASLink;
  };
}

export type CapabilityLinkStatus = 'available' | 'linked' | 'blocked_by_parent' | 'blocked_by_child';

export interface CapabilityLinkStatusResponse {
  capabilityId: string;
  status: CapabilityLinkStatus;
  linkedTo?: { id: string; name: string };
  blockingCapability?: { id: string; name: string };
  blockingEnterpriseCapabilityId?: string;
}

export interface MaturityDistribution {
  genesis: number;
  customBuild: number;
  product: number;
  commodity: number;
}

export interface MaturityAnalysisCandidate {
  enterpriseCapabilityId: string;
  enterpriseCapabilityName: string;
  category?: string;
  targetMaturity: number | null;
  targetMaturitySection?: string;
  implementationCount: number;
  domainCount: number;
  maxMaturity: number;
  minMaturity: number;
  averageMaturity: number;
  maxGap: number;
  maturityDistribution: MaturityDistribution;
  _links: {
    self: HATEOASLink;
    'x-maturity-gap': HATEOASLink;
  };
}

export interface MaturityAnalysisSummary {
  candidateCount: number;
  totalImplementations: number;
  averageGap: number;
}

export interface MaturityAnalysisResponse {
  summary: MaturityAnalysisSummary;
  data: MaturityAnalysisCandidate[];
  _links: {
    self: HATEOASLink;
  };
}

export interface ImplementationDetail {
  domainCapabilityId: string;
  domainCapabilityName: string;
  businessDomainId?: string;
  businessDomainName?: string;
  maturityValue: number;
  maturitySection: string;
  gap: number;
  priority: 'High' | 'Medium' | 'Low' | 'None';
}

export interface InvestmentPriorities {
  high: ImplementationDetail[];
  medium: ImplementationDetail[];
  low: ImplementationDetail[];
  onTarget: ImplementationDetail[];
}

export interface MaturityGapDetail {
  enterpriseCapabilityId: string;
  enterpriseCapabilityName: string;
  category?: string;
  targetMaturity: number | null;
  targetMaturitySection?: string;
  implementations: ImplementationDetail[];
  investmentPriorities: InvestmentPriorities;
  _links: {
    self: HATEOASLink;
    up: HATEOASLink;
    'x-set-target-maturity'?: HATEOASLink;
  };
}

export type TimeClassification = 'Tolerate' | 'Invest' | 'Migrate' | 'Eliminate';
export type TimeSuggestionConfidence = 'High' | 'Medium' | 'Low' | 'Insufficient';

export interface TimeSuggestion {
  capabilityId: string;
  capabilityName: string;
  componentId: string;
  componentName: string;
  suggestedTime: TimeClassification | null;
  technicalGap: number | null;
  functionalGap: number | null;
  confidence: TimeSuggestionConfidence;
}

export interface TimeSuggestionsResponse {
  data: TimeSuggestion[];
  _links: {
    self: HATEOASLink;
  };
}

