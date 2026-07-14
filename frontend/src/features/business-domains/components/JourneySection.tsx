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
import type { CapabilityJourney, CapabilityJourneyResponse } from '../../journeys/types';
import { DrawerSectionHeader } from './DrawerSectionHeader';
import classes from './JourneySection.module.css';

export interface JourneySectionProps {
  capability: Capability;
  realizations: CapabilityRealization[];
}

function useDisplayJourney(capabilityId: string) {
  const wrapperQuery = useJourneyForCapability(capabilityId);
  const wrapper = wrapperQuery.data;
  const activeJourney = wrapper?.journey ?? null;
  const historyQuery = useJourneyHistory(wrapper && !activeJourney ? capabilityId : undefined);
  const displayJourney = activeJourney ?? historyQuery.data?.data[0] ?? null;
  return { wrapper, displayJourney };
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
  if (!hasLink(wrapper, 'x-capture')) return null;

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
          <CaptureJourneyForm capability={capability} realizations={realizations} onCaptured={close} onCancel={close} />
        </Modal>
      )}
    </>
  );
}

export function JourneySection({ capability, realizations }: JourneySectionProps) {
  const { wrapper, displayJourney } = useDisplayJourney(String(capability.id));

  if (!wrapper) return null;

  return (
    <Stack gap="xs" data-testid="journey-section">
      <DrawerSectionHeader>Transition</DrawerSectionHeader>
      {!displayJourney && <Text className={classes.empty}>No change planned.</Text>}
      <CaptureAffordance wrapper={wrapper} capability={capability} realizations={realizations} />
      {displayJourney && <JourneyContent journey={displayJourney} realizations={realizations} />}
    </Stack>
  );
}
