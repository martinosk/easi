import { Button, Group, Modal, Stack, Text } from '@mantine/core';
import { useState } from 'react';
import type { Capability, CapabilityRealization } from '../../../api/types';
import { hasLink } from '../../../utils/hateoas';
import { CaptureJourneyForm } from '../../journeys/components/CaptureJourneyForm';
import { JourneyActions } from '../../journeys/components/JourneyActions';
import { JourneyMilestones } from '../../journeys/components/JourneyMilestones';
import { JourneyProgressBar } from '../../journeys/components/JourneyProgressBar';
import { JourneyTransitionTable } from '../../journeys/components/JourneyTransitionTable';
import { useJourneyForCapability, useJourneyHistory } from '../../journeys/hooks/useJourneys';
import type { CapabilityJourney, CapabilityJourneyResponse, JourneyKind } from '../../journeys/types';
import type { CapabilityHierarchyJourneys } from '../lens/hierarchyJourneys';
import { DrawerSectionHeader } from './DrawerSectionHeader';
import { AncestorJourneys, SubCapabilityJourneys } from './HierarchyJourneys';
import classes from './JourneySection.module.css';

export interface JourneySectionProps {
  capability: Capability;
  realizations: CapabilityRealization[];
  hierarchyJourneys: CapabilityHierarchyJourneys;
  onNavigateToCapability: (capabilityId: string) => void;
}

function useDisplayJourneys(capabilityId: string) {
  const wrapperQuery = useJourneyForCapability(capabilityId);
  const wrapper = wrapperQuery.data;
  const activeJourneys = wrapper?.journeys ?? [];
  const historyQuery = useJourneyHistory(wrapper && activeJourneys.length === 0 ? capabilityId : undefined);
  const mostRecent = historyQuery.data?.data[0];
  const displayJourneys = activeJourneys.length > 0 ? activeJourneys : mostRecent ? [mostRecent] : [];
  return { wrapper, displayJourneys };
}

function availableCaptureKinds(wrapper: CapabilityJourneyResponse): JourneyKind[] {
  const kinds: JourneyKind[] = [];
  if (hasLink(wrapper, 'x-capture')) kinds.push('migration', 'consolidation', 'carve-out', 'move');
  if (hasLink(wrapper, 'x-capture-maturity')) kinds.push('maturity');
  return kinds;
}

function JourneyContent({
  journey,
  realizations,
}: {
  journey: CapabilityJourney;
  realizations: CapabilityRealization[];
}) {
  const showMilestones = journey.milestones.length > 0 || hasLink(journey, 'x-add-milestone');
  return (
    <>
      <JourneyTransitionTable journey={journey} />
      <JourneyProgressBar journey={journey} />
      <JourneyActions journey={journey} realizations={realizations} />
      {journey.note && (
        <>
          <DrawerSectionHeader>Plan summary</DrawerSectionHeader>
          <Text className={classes.note} data-testid="journey-note">
            {journey.note}
          </Text>
        </>
      )}
      {showMilestones && (
        <>
          <DrawerSectionHeader>Milestones</DrawerSectionHeader>
          <JourneyMilestones journey={journey} />
        </>
      )}
    </>
  );
}

function CaptureAffordance({
  wrapper,
  capability,
  realizations,
}: {
  wrapper: CapabilityJourneyResponse;
  capability: Capability;
  realizations: CapabilityRealization[];
}) {
  const [open, setOpen] = useState(false);
  const availableKinds = availableCaptureKinds(wrapper);
  if (availableKinds.length === 0) return null;

  const close = () => setOpen(false);

  return (
    <>
      <Group>
        <Button variant="default" size="xs" onClick={() => setOpen(true)} data-testid="plan-journey-btn">
          Plan journey
        </Button>
      </Group>
      {open && (
        <Modal opened onClose={close} title="Plan journey" size="lg" data-testid="capture-journey-modal">
          <CaptureJourneyForm
            capability={capability}
            realizations={realizations}
            availableKinds={availableKinds}
            onCaptured={close}
            onCancel={close}
          />
        </Modal>
      )}
    </>
  );
}

export function JourneySection({
  capability,
  realizations,
  hierarchyJourneys,
  onNavigateToCapability,
}: JourneySectionProps) {
  const { wrapper, displayJourneys } = useDisplayJourneys(String(capability.id));

  if (!wrapper) return null;

  return (
    <Stack gap="xs" data-testid="journey-section">
      <AncestorJourneys journeys={hierarchyJourneys.ancestors} onNavigate={onNavigateToCapability} />
      <DrawerSectionHeader>Transition</DrawerSectionHeader>
      {displayJourneys.length === 0 && <Text className={classes.empty}>No change planned.</Text>}
      <CaptureAffordance wrapper={wrapper} capability={capability} realizations={realizations} />
      {displayJourneys.map((journey) => (
        <JourneyContent key={journey.id} journey={journey} realizations={realizations} />
      ))}
      <SubCapabilityJourneys journeys={hierarchyJourneys.descendants} onNavigate={onNavigateToCapability} />
    </Stack>
  );
}
