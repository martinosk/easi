import type { CapabilityLevel, CapabilityId } from '../../../api/types';

export interface RealizingSystem {
  componentId: string;
  componentName: string;
  componentType: string;
}

export interface CapabilityTreeNode {
  id: CapabilityId;
  code: string;
  name: string;
  description?: string;
  level: CapabilityLevel;
  parentId?: CapabilityId;
  children: CapabilityTreeNode[];
  realizingSystems: RealizingSystem[];
  associatedDomains: string[];
  isOrphaned: boolean;
}
