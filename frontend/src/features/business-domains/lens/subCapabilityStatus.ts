import type { CapabilityId, CapabilityRealization } from '../../../api/types';
import type { CapabilityTreeNode } from '../../capabilities/hooks/useCapabilityTree';
import type { CapabilityJourney } from '../../journeys/types';

export type SubCapabilityStatus = 'done' | 'in-flight' | 'not-started';

export interface JourneyApps {
  fromComponentIds: Set<string>;
  toComponentId: string;
}

export interface SubCapabilityRow {
  capability: CapabilityTreeNode['capability'];
  status: SubCapabilityStatus;
  appLabel: string;
}

export function journeyApps(journey: CapabilityJourney): JourneyApps {
  return {
    fromComponentIds: new Set(journey.fromApplications.map((app) => app.componentId)),
    toComponentId: journey.toApplication.componentId,
  };
}

function directRealizations(realizations: CapabilityRealization[]): CapabilityRealization[] {
  return realizations.filter((realization) => realization.origin === 'Direct');
}

export function deriveSubCapabilityStatus(
  realizations: CapabilityRealization[],
  apps: JourneyApps,
): SubCapabilityStatus | null {
  const direct = directRealizations(realizations);
  const toRealization = direct.find((realization) => realization.componentId === apps.toComponentId);
  const hasFrom = direct.some((realization) => apps.fromComponentIds.has(realization.componentId));

  if (!toRealization) return hasFrom ? 'not-started' : null;
  if (toRealization.realizationLevel === 'Full' && !hasFrom) return 'done';
  return 'in-flight';
}

export function deriveSubCapabilityAppLabel(
  realizations: CapabilityRealization[],
  apps: JourneyApps,
  status: SubCapabilityStatus,
): string {
  const direct = directRealizations(realizations);
  const toName = direct.find((realization) => realization.componentId === apps.toComponentId)?.componentName ?? '';
  const fromNames = direct
    .filter((realization) => apps.fromComponentIds.has(realization.componentId))
    .map((realization) => realization.componentName);

  if (status === 'done') return toName;
  if (status === 'not-started') return fromNames.join(', ');
  if (fromNames.length && toName) return `${fromNames.join(', ')} → ${toName}`;
  return toName || fromNames.join(', ');
}

export function buildSubCapabilityBreakdown(
  node: CapabilityTreeNode,
  journey: CapabilityJourney,
  getRealizations: (capabilityId: CapabilityId) => CapabilityRealization[],
): SubCapabilityRow[] {
  const apps = journeyApps(journey);
  const rows: SubCapabilityRow[] = [];

  const visit = (current: CapabilityTreeNode) => {
    for (const child of current.children) {
      const realizations = getRealizations(child.capability.id);
      const status = deriveSubCapabilityStatus(realizations, apps);
      if (status) {
        rows.push({
          capability: child.capability,
          status,
          appLabel: deriveSubCapabilityAppLabel(realizations, apps, status),
        });
      }
      visit(child);
    }
  };

  visit(node);
  return rows;
}
