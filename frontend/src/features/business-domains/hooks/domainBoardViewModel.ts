import type {
  BusinessDomain,
  Capability,
  CapabilityId,
  CapabilityRealization,
  CapabilityRealizationsGroup,
  ComponentId,
  TimeGrade,
} from '../../../api/types';
import type { RealizationRole, RealizationRoleAssignment, TimeAssessment } from '../../architecture-direction/types';
import type { CapabilityTreeNode } from '../../capabilities/hooks/useCapabilityTree';

export type AssessedRealization = CapabilityRealization & { timeGrade?: TimeGrade; role?: RealizationRole };

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
  getRealizationsForCapability: (capabilityId: CapabilityId) => AssessedRealization[];
}

function countNode(node: CapabilityTreeNode): number {
  return 1 + node.children.reduce((sum, child) => sum + countNode(child), 0);
}

function collectComponentIds(
  node: CapabilityTreeNode,
  realizationsByCapabilityId: Map<CapabilityId, AssessedRealization[]>,
  acc: Set<ComponentId>,
): void {
  for (const realization of realizationsByCapabilityId.get(node.capability.id) ?? []) {
    acc.add(realization.componentId);
  }
  for (const child of node.children) {
    collectComponentIds(child, realizationsByCapabilityId, acc);
  }
}

function buildGradeByPair(assessments: TimeAssessment[]): Map<string, TimeAssessment['grade']> {
  const map = new Map<string, TimeAssessment['grade']>();
  for (const assessment of assessments) {
    map.set(`${assessment.capabilityId}:${assessment.componentId}`, assessment.grade);
  }
  return map;
}

function buildRoleByPair(roles: RealizationRoleAssignment[]): Map<string, RealizationRole> {
  const map = new Map<string, RealizationRole>();
  for (const role of roles) {
    map.set(`${role.capabilityId}:${role.componentId}`, role.role);
  }
  return map;
}

function withAssessedGrade(
  realization: CapabilityRealization,
  gradeByPair: Map<string, TimeAssessment['grade']>,
): AssessedRealization {
  if (realization.origin !== 'Direct') return realization;
  const grade = gradeByPair.get(`${realization.capabilityId}:${realization.componentId}`);
  return grade ? { ...realization, timeGrade: grade } : realization;
}

function withRole(realization: AssessedRealization, roleByPair: Map<string, RealizationRole>): AssessedRealization {
  if (realization.origin !== 'Direct') return realization;
  const role = roleByPair.get(`${realization.capabilityId}:${realization.componentId}`);
  return role ? { ...realization, role } : realization;
}

export function buildRealizationsByCapabilityId(
  groups: CapabilityRealizationsGroup[],
  assessments: TimeAssessment[] = [],
  roles: RealizationRoleAssignment[] = [],
): Map<CapabilityId, AssessedRealization[]> {
  const gradeByPair = buildGradeByPair(assessments);
  const roleByPair = buildRoleByPair(roles);
  const map = new Map<CapabilityId, AssessedRealization[]>();
  for (const group of groups) {
    map.set(
      group.capabilityId,
      group.realizations.map((realization) => withRole(withAssessedGrade(realization, gradeByPair), roleByPair)),
    );
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
  assessments?: TimeAssessment[];
  roles?: RealizationRoleAssignment[];
}

export function buildDomainBoardViewModel(params: BuildDomainBoardViewModelParams): DomainBoardViewModel {
  const { domain, assignedCapabilities, tree, realizationGroups, isLoading, assessments = [], roles = [] } = params;
  const assignedL1Ids = new Set(assignedCapabilities.filter((c) => c.level === 'L1').map((c) => c.id));
  const realizationsByCapabilityId = buildRealizationsByCapabilityId(realizationGroups, assessments, roles);
  const l1Nodes = tree.filter((node) => assignedL1Ids.has(node.capability.id));

  const domainAppIds = new Set<ComponentId>();
  const l1Groups: DomainBoardL1Group[] = l1Nodes.map((node) => {
    const appIds = new Set<ComponentId>();
    collectComponentIds(node, realizationsByCapabilityId, appIds);
    for (const id of appIds) domainAppIds.add(id);
    return { node, distinctAppCount: appIds.size };
  });

  const totalCapabilityCount = l1Nodes.reduce((sum, node) => sum + countNode(node), 0);

  const getRealizationsForCapability = (capabilityId: CapabilityId): AssessedRealization[] =>
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
