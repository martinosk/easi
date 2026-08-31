import type {
  CapabilityLevel,
  EnterpriseCapabilityId,
  EnterpriseStrategicImportanceId,
  HATEOASLink,
  HATEOASLinks,
} from '../../api/types';

export type { EnterpriseCapabilityId, EnterpriseStrategicImportanceId };

export interface EnterpriseCapability {
  id: EnterpriseCapabilityId;
  name: string;
  description: string;
  category: string;
  active: boolean;
  targetMaturity?: number;
  createdAt: string;
  updatedAt?: string;
  _links: HATEOASLinks & {
    self: HATEOASLink;
    edit?: HATEOASLink;
    delete?: HATEOASLink;
    'x-strategic-importance'?: HATEOASLink;
  };
}

export type IncludedCapabilityRole = 'source' | 'implicit' | 'carved-out';

export interface CarvedOutBy {
  enterpriseCapabilityId: string;
  enterpriseCapabilityName: string;
}

export interface IncludedCapabilityItem {
  capabilityId: string;
  name: string;
  level: CapabilityLevel;
  businessDomainId?: string | null;
  businessDomainName?: string | null;
  role: IncludedCapabilityRole;
  carvedOutBy?: CarvedOutBy | null;
  _links: HATEOASLinks & {
    self?: HATEOASLink;
    'x-exclude'?: HATEOASLink;
    'x-owning-ec'?: HATEOASLink;
  };
}

export interface CompositionDomainGroup {
  businessDomainId: string | null;
  businessDomainName: string | null;
  items: IncludedCapabilityItem[];
}

export interface CompositionMeta {
  sourceCount: number;
  includedCount: number;
  carvedOutCount: number;
  domainCount: number;
}

export interface CompositionResponse {
  data: CompositionDomainGroup[];
  meta: CompositionMeta;
  _links: HATEOASLinks & {
    self?: HATEOASLink;
    up?: HATEOASLink;
    'x-direction'?: HATEOASLink;
    'x-capture-direction'?: HATEOASLink;
  };
}

export interface CompositionSummary {
  enterpriseCapabilityId: string;
  sourceCount: number;
  includedCount: number;
  carvedOutCount: number;
  domainCount: number;
  hasDirection: boolean;
  directionStatus?: string;
  _links: HATEOASLinks & {
    'x-enterprise-capability'?: HATEOASLink;
    'x-composition'?: HATEOASLink;
    'x-direction'?: HATEOASLink;
  };
}

export interface CompositionSummariesResponse {
  data: CompositionSummary[];
  _links: {
    self: HATEOASLink;
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
