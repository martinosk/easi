import { Drawer, Stack, Text } from '@mantine/core';
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

interface DomainContextProps {
  capability: Capability;
  domain: BusinessDomain | null;
  realizations: CapabilityRealization[];
  hierarchyJourneys: CapabilityHierarchyJourneys;
  onNavigateToCapability: (capabilityId: string) => void;
}

function DomainContext({
  capability,
  domain,
  realizations,
  hierarchyJourneys,
  onNavigateToCapability,
}: DomainContextProps) {
  return (
    <Stack gap="md">
      <JourneySection
        capability={capability}
        realizations={realizations}
        hierarchyJourneys={hierarchyJourneys}
        onNavigateToCapability={onNavigateToCapability}
      />
      {domain && <StrategicImportanceSection domain={domain} capabilityId={capability.id} />}
    </Stack>
  );
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
          domainContext={
            <DomainContext
              capability={capability}
              domain={domain}
              realizations={getRealizationsForCapability(capability.id)}
              hierarchyJourneys={hierarchyJourneys}
              onNavigateToCapability={onNavigateToCapability}
            />
          }
        />
      )}
    </Drawer>
  );
}
