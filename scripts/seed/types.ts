export interface Component {
  id: string;
  name: string;
  description: string;
}

export interface Capability {
  id: string;
  name: string;
  description: string;
  level: string;
  parentId?: string;
}

export interface BusinessDomain {
  id: string;
  name: string;
  description: string;
}

export interface EnterpriseCapability {
  id: string;
  name: string;
  description: string;
  category: string;
}

export interface View {
  id: string;
  name: string;
  description: string;
}

export interface StrategyPillar {
  id: string;
  name: string;
  description: string;
  active: boolean;
  fitScoringEnabled: boolean;
}

export interface AcquiredEntity {
  id: string;
  name: string;
  acquisitionDate?: string;
  integrationStatus: string;
  notes?: string;
}

export interface Vendor {
  id: string;
  name: string;
  implementationPartner?: string;
  notes?: string;
}

export interface InternalTeam {
  id: string;
  name: string;
  department?: string;
  contactPerson?: string;
  notes?: string;
}

export interface CapNode {
  name: string;
  description: string;
  level: string;
  children?: CapNode[];
}
