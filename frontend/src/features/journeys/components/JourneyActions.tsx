import { Button, Group, Modal } from '@mantine/core';
import { useState } from 'react';
import type { CapabilityRealization } from '../../../api/types';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import { hasLink } from '../../../utils/hateoas';
import { useAbandonJourney, useCompleteJourney, useStartJourney } from '../hooks/useJourneys';
import type { CapabilityJourney } from '../types';
import { ChangeSourcesForm } from './ChangeSourcesForm';
import { EditJourneyDetailsForm } from './EditJourneyDetailsForm';
import { ProgressEditor } from './ProgressEditor';

interface ActionButtonProps {
  journey: CapabilityJourney;
  rel: string;
  testId: string;
  label: string;
  loading?: boolean;
  onClick: () => void;
}

function ActionButton({ journey, rel, testId, label, loading, onClick }: ActionButtonProps) {
  if (!hasLink(journey, rel)) return null;
  return (
    <Button variant="subtle" size="compact-xs" onClick={onClick} loading={loading} data-testid={testId}>
      {label}
    </Button>
  );
}

type OpenEditor = 'progress' | 'details' | 'sources' | 'abandon' | null;

export interface JourneyActionsProps {
  journey: CapabilityJourney;
  realizations: CapabilityRealization[];
}

export function JourneyActions({ journey, realizations }: JourneyActionsProps) {
  const startMutation = useStartJourney();
  const completeMutation = useCompleteJourney();
  const abandonMutation = useAbandonJourney();
  const [open, setOpen] = useState<OpenEditor>(null);

  const close = () => setOpen(null);

  const abandon = async () => {
    await abandonMutation.mutateAsync(journey);
    close();
  };

  return (
    <>
      <Group gap="xs">
        <ActionButton
          journey={journey}
          rel="x-start"
          testId="start-journey-btn"
          label="Start"
          loading={startMutation.isPending}
          onClick={() => startMutation.mutateAsync(journey)}
        />
        <ActionButton
          journey={journey}
          rel="x-complete"
          testId="complete-journey-btn"
          label="Complete"
          loading={completeMutation.isPending}
          onClick={() => completeMutation.mutateAsync(journey)}
        />
        <ActionButton
          journey={journey}
          rel="x-abandon"
          testId="abandon-journey-btn"
          label="Abandon"
          onClick={() => setOpen('abandon')}
        />
        <ActionButton
          journey={journey}
          rel="x-progress"
          testId="update-progress-btn"
          label="Update progress"
          onClick={() => setOpen('progress')}
        />
        <ActionButton
          journey={journey}
          rel="edit"
          testId="edit-journey-btn"
          label="Edit details"
          onClick={() => setOpen('details')}
        />
        <ActionButton
          journey={journey}
          rel="x-change-sources"
          testId="change-sources-btn"
          label="Change sources"
          onClick={() => setOpen('sources')}
        />
      </Group>
      {open === 'progress' && <ProgressEditor journey={journey} onClose={close} />}
      {open === 'abandon' && (
        <ConfirmationDialog
          title="Abandon journey"
          message="Abandon this journey? It will be frozen and preserved as history."
          confirmText="Abandon"
          onConfirm={abandon}
          onCancel={close}
          isLoading={abandonMutation.isPending}
        />
      )}
      {open === 'details' && (
        <Modal opened onClose={close} title="Edit journey details" data-testid="edit-journey-modal">
          <EditJourneyDetailsForm journey={journey} onDone={close} />
        </Modal>
      )}
      {open === 'sources' && (
        <Modal opened onClose={close} title="Change source applications" data-testid="change-sources-modal">
          <ChangeSourcesForm journey={journey} realizations={realizations} onDone={close} />
        </Modal>
      )}
    </>
  );
}
