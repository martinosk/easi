import { Drawer, Text } from '@mantine/core';
import type { BusinessDomain, Capability, CapabilityId, CapabilityRealization, ComponentId } from '../../../api/types';
import { CapabilityDetailsPanel } from '../../capabilities/components/CapabilityDetailsPanel';
import type { CapabilityHierarchyJourneys } from '../lens/hierarchyJourneys';
import classes from './CapabilityDrawer.module.css';
import { JourneySection } from './JourneySection';
import { StrategicImportanceSection } from './StrategicImportanceSection';

export interface CapabilityDrawerProps {
  capability: Capability | null;
  domain: BusinessDomain | null;
  l1Name: string | null;
  getRealizationsForCapability: (capabilityId: CapabilityId) => CapabilityRealization[];
  hierarchyJourneys: CapabilityHierarchyJourneys;
  onClose: () => void;
  onChipClick: (componentId: ComponentId) => void;
  onNavigateToCapability: (capabilityId: string) => void;
}

export function CapabilityDrawer({
  capability,
  domain,
  l1Name,
  getRealizationsForCapability,
  hierarchyJourneys,
  onClose,
  onChipClick,
  onNavigateToCapability,
}: CapabilityDrawerProps) {
  return (
    <Drawer
      opened={capability !== null}
      onClose={onClose}
      position="right"
      size="md"
      data-testid="capability-drawer"
      title={
        domain && l1Name ? (
          <Text className={classes.breadcrumb}>
            {domain.name} · {l1Name}
          </Text>
        ) : undefined
      }
    >
      {capability && (
        <CapabilityDetailsPanel
          capabilityId={capability.id}
          onApplicationClick={onChipClick}
          transition={
            <JourneySection
              capability={capability}
              realizations={getRealizationsForCapability(capability.id)}
              hierarchyJourneys={hierarchyJourneys}
              onNavigateToCapability={onNavigateToCapability}
            />
          }
          strategicImportance={domain && <StrategicImportanceSection domain={domain} capabilityId={capability.id} />}
        />
      )}
    </Drawer>
  );
}
