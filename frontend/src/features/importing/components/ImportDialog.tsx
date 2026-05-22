import { Modal, Stepper } from '@mantine/core';
import { useLayoutEffect, useMemo, useState } from 'react';
import type { BusinessDomain } from '../../../api/types';
import { useEAOwnerCandidates } from '../../users/hooks/useUsers';
import { useImportSession } from '../hooks/useImportSession';
import { ImportPreviewStep } from './ImportPreviewStep';
import { ImportProgressStep } from './ImportProgressStep';
import { ImportResultsStep } from './ImportResultsStep';
import { ImportUploadStep } from './ImportUploadStep';

interface ImportDialogProps {
  isOpen: boolean;
  onClose: () => void;
  businessDomains?: BusinessDomain[];
}

type ImportStep = 'upload' | 'preview' | 'progress' | 'results';

const STEP_INDEX: Record<ImportStep, number> = {
  upload: 0,
  preview: 1,
  progress: 2,
  results: 3,
};

function useStepFromSession(session: ReturnType<typeof useImportSession>['session'], currentStep: ImportStep) {
  const [step, setStep] = useState<ImportStep>(currentStep);

  useLayoutEffect(() => {
    if (!session) {
      if (step !== 'upload') queueMicrotask(() => setStep('upload'));
      return;
    }

    switch (session.status) {
      case 'pending':
        if (session.preview) queueMicrotask(() => setStep('preview'));
        break;
      case 'importing':
        queueMicrotask(() => setStep('progress'));
        break;
      case 'completed':
      case 'failed':
        queueMicrotask(() => setStep('results'));
        break;
    }
  }, [session, step]);

  return step;
}

export function ImportDialog({ isOpen, onClose, businessDomains = [] }: ImportDialogProps) {
  const { session, isLoading, error, createSession, confirmSession, cancelSession, reset } = useImportSession();
  const { data: eaOwnerCandidates = [] } = useEAOwnerCandidates();

  const currentStep = useStepFromSession(session, 'upload');

  const eaOwnerName = useMemo(() => {
    if (!session?.capabilityEAOwner) return undefined;
    const user = eaOwnerCandidates.find((u) => u.id === session.capabilityEAOwner);
    return user?.name || user?.email;
  }, [session, eaOwnerCandidates]);

  const handleUpload = async (file: File, businessDomainId?: string, capabilityEAOwner?: string) => {
    await createSession({ file, sourceFormat: 'archimate-openexchange', businessDomainId, capabilityEAOwner });
  };

  const handleCancel = async () => {
    if (session && session.status === 'pending') {
      await cancelSession();
    }
    reset();
    onClose();
  };

  const handleClose = () => {
    reset();
    onClose();
  };

  return (
    <Modal
      opened={isOpen}
      onClose={handleCancel}
      title="Import from ArchiMate"
      size="lg"
      centered
      data-testid="import-dialog"
    >
      <Stepper active={STEP_INDEX[currentStep]} mb="lg">
        <Stepper.Step label="Upload" />
        <Stepper.Step label="Preview" />
        <Stepper.Step label="Import" />
        <Stepper.Step label="Results" />
      </Stepper>

      <StepContent
        step={currentStep}
        session={session}
        isLoading={isLoading}
        error={error}
        businessDomains={businessDomains}
        eaOwnerCandidates={eaOwnerCandidates}
        eaOwnerName={eaOwnerName}
        onUpload={handleUpload}
        onConfirm={confirmSession}
        onCancel={handleCancel}
        onClose={handleClose}
      />
    </Modal>
  );
}

type Session = ReturnType<typeof useImportSession>['session'];

interface StepContentProps {
  step: ImportStep;
  session: Session;
  isLoading: boolean;
  error: string | null;
  businessDomains: BusinessDomain[];
  eaOwnerCandidates: ReturnType<typeof useEAOwnerCandidates>['data'];
  eaOwnerName: string | undefined;
  onUpload: (file: File, businessDomainId?: string, capabilityEAOwner?: string) => Promise<void>;
  onConfirm: () => Promise<void>;
  onCancel: () => Promise<void>;
  onClose: () => void;
}

function PreviewContent({ session, ...rest }: {
  session: Session;
  isLoading: boolean;
  eaOwnerName: string | undefined;
  onConfirm: () => Promise<void>;
  onCancel: () => Promise<void>;
}) {
  if (!session?.preview) return null;
  return <ImportPreviewStep preview={session.preview} {...rest} />;
}

function ProgressContent({ session }: { session: Session }) {
  if (!session?.progress) return null;
  return <ImportProgressStep progress={session.progress} />;
}

function ResultsContent({ session, onClose }: { session: Session; onClose: () => void }) {
  if (!session?.result) return null;
  return <ImportResultsStep result={session.result} onClose={onClose} />;
}

function StepContent(props: StepContentProps) {
  const { step, session, isLoading, error, businessDomains, eaOwnerCandidates, eaOwnerName, onUpload, onConfirm, onCancel, onClose } = props;

  switch (step) {
    case 'upload':
      return (
        <ImportUploadStep
          businessDomains={businessDomains}
          eaOwnerCandidates={eaOwnerCandidates ?? []}
          isLoading={isLoading}
          error={error}
          onUpload={onUpload}
          onCancel={onCancel}
        />
      );
    case 'preview':
      return (
        <PreviewContent
          session={session}
          isLoading={isLoading}
          eaOwnerName={eaOwnerName}
          onConfirm={onConfirm}
          onCancel={onCancel}
        />
      );
    case 'progress':
      return <ProgressContent session={session} />;
    case 'results':
      return <ResultsContent session={session} onClose={onClose} />;
  }
}
