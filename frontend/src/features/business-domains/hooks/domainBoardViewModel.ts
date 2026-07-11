import type {
  BusinessDomain,
  Capability,
  CapabilityId,
  CapabilityRealization,
  CapabilityRealizationsGroup,
  ComponentId,
} from '../../../api/types';
import type { CapabilityTreeNode } from '../../capabilities/hooks/useCapabilityTree';

export interface DomainBoardL1Group {
  node: CapabilityTreeNode;
  distinctAppCount: number;
}

export interface DomainBoardViewModel {
  domain: BusinessDomain;
  l1Groups: DomainBoardL1Group[];
  assignedCapabilities: Capability[];
  totalCapabilityCount: number;
  totalAppCount: number;
  isLoading: boolean;
  getRealizationsForCapability: (capabilityId: CapabilityId) => CapabilityRealization[];
}

function countNode(node: CapabilityTreeNode): number {
  return 1 + node.children.reduce((sum, child) => sum + countNode(child), 0);
}

function collectComponentIds(
  node: CapabilityTreeNode,
  realizationsByCapabilityId: Map<CapabilityId, CapabilityRealization[]>,
  acc: Set<ComponentId>,
): void {
  for (const realization of realizationsByCapabilityId.get(node.capability.id) ?? []) {
    acc.add(realization.componentId);
  }
  for (const child of node.children) {
    collectComponentIds(child, realizationsByCapabilityId, acc);
  }
}

export function buildRealizationsByCapabilityId(
  groups: CapabilityRealizationsGroup[],
): Map<CapabilityId, CapabilityRealization[]> {
  const map = new Map<CapabilityId, CapabilityRealization[]>();
  for (const group of groups) {
    map.set(group.capabilityId, group.realizations);
  }
  return map;
}

export function flattenViewModelCapabilities(viewModel: DomainBoardViewModel): Capability[] {
  const result: Capability[] = [];
  const visit = (node: CapabilityTreeNode) => {
    result.push(node.capability);
    node.children.forEach(visit);
  };
  viewModel.l1Groups.forEach((group) => {
    visit(group.node);
  });
  return result;
}

export interface BuildDomainBoardViewModelParams {
  domain: BusinessDomain;
  assignedCapabilities: Capability[];
  tree: CapabilityTreeNode[];
  realizationGroups: CapabilityRealizationsGroup[];
  isLoading: boolean;
}

export function buildDomainBoardViewModel(params: BuildDomainBoardViewModelParams): DomainBoardViewModel {
  const { domain, assignedCapabilities, tree, realizationGroups, isLoading } = params;
  const assignedL1Ids = new Set(assignedCapabilities.filter((c) => c.level === 'L1').map((c) => c.id));
  const realizationsByCapabilityId = buildRealizationsByCapabilityId(realizationGroups);
  const l1Nodes = tree.filter((node) => assignedL1Ids.has(node.capability.id));

  const domainAppIds = new Set<ComponentId>();
  const l1Groups: DomainBoardL1Group[] = l1Nodes.map((node) => {
    const appIds = new Set<ComponentId>();
    collectComponentIds(node, realizationsByCapabilityId, appIds);
    for (const id of appIds) domainAppIds.add(id);
    return { node, distinctAppCount: appIds.size };
  });

  const totalCapabilityCount = l1Nodes.reduce((sum, node) => sum + countNode(node), 0);

  const getRealizationsForCapability = (capabilityId: CapabilityId): CapabilityRealization[] =>
    realizationsByCapabilityId.get(capabilityId) ?? [];

  return {
    domain,
    l1Groups,
    assignedCapabilities,
    totalCapabilityCount,
    totalAppCount: domainAppIds.size,
    isLoading,
    getRealizationsForCapability,
  };
}
