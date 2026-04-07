import React, { useState, useEffect, useLayoutEffect, useRef, useCallback, useMemo } from 'react';
import { ImportUploadStep } from './ImportUploadStep';
import { ImportPreviewStep } from './ImportPreviewStep';
import { ImportProgressStep } from './ImportProgressStep';
import { ImportResultsStep } from './ImportResultsStep';
import { useImportSession } from '../hooks/useImportSession';
import { useEAOwnerCandidates } from '../../users/hooks/useUsers';
import type { BusinessDomain } from '../../../api/types';

interface ImportDialogProps {
  isOpen: boolean;
  onClose: () => void;
  businessDomains?: BusinessDomain[];
}

type ImportStep = 'upload' | 'preview' | 'progress' | 'results';

export const ImportDialog: React.FC<ImportDialogProps> = ({
  isOpen,
  onClose,
  businessDomains = [],
}) => {
  const [currentStep, setCurrentStep] = useState<ImportStep>('upload');
  const dialogRef = useRef<HTMLDialogElement>(null);

  const {
    session,
    isLoading,
    error,
    createSession,
    confirmSession,
    cancelSession,
    reset,
  } = useImportSession();

  const { data: eaOwnerCandidates = [] } = useEAOwnerCandidates();

  const eaOwnerName = useMemo(() => {
    if (!session?.capabilityEAOwner) return undefined;
    const user = eaOwnerCandidates.find((u) => u.id === session.capabilityEAOwner);
    return user?.name || user?.email;
  }, [session, eaOwnerCandidates]);

  const handleBackdropClick = useCallback((e: React.MouseEvent<HTMLDialogElement>) => {
    if (e.target === dialogRef.current) {
      reset();
      onClose();
    }
  }, [reset, onClose]);

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;

    if (isOpen && !dialog.open) {
      dialog.showModal();
    } else if (!isOpen && dialog.open) {
      dialog.close();
    }
  }, [isOpen]);

  useLayoutEffect(() => {
    if (!session) {
      if (currentStep !== 'upload') queueMicrotask(() => setCurrentStep('upload'));
      return;
    }

    switch (session.status) {
      case 'pending':
        if (session.preview) {
          queueMicrotask(() => setCurrentStep('preview'));
        }
        break;
      case 'importing':
        queueMicrotask(() => setCurrentStep('progress'));
        break;
      case 'completed':
      case 'failed':
        queueMicrotask(() => setCurrentStep('results'));
        break;
    }
  }, [session, currentStep]);

  const handleUpload = async (file: File, businessDomainId?: string, capabilityEAOwner?: string) => {
    await createSession({
      file,
      sourceFormat: 'archimate-openexchange',
      businessDomainId,
      capabilityEAOwner,
    });
  };

  const handleConfirm = async () => {
    await confirmSession();
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

  const renderUploadStep = () => (
    <ImportUploadStep
      businessDomains={businessDomains}
      eaOwnerCandidates={eaOwnerCandidates}
      isLoading={isLoading}
      error={error}
      onUpload={handleUpload}
      onCancel={handleCancel}
    />
  );

  const renderPreviewStep = () =>
    session?.preview && (
      <ImportPreviewStep
        preview={session.preview}
        eaOwnerName={eaOwnerName}
        onConfirm={handleConfirm}
        onCancel={handleCancel}
        isLoading={isLoading}
      />
    );

  const renderProgressStep = () =>
    session?.progress && <ImportProgressStep progress={session.progress} />;

  const renderResultsStep = () =>
    session?.result && <ImportResultsStep result={session.result} onClose={handleClose} />;

  const stepRenderers: Record<ImportStep, () => React.ReactNode> = {
    upload: renderUploadStep,
    preview: renderPreviewStep,
    progress: renderProgressStep,
    results: renderResultsStep,
  };

  const renderStep = () => stepRenderers[currentStep]?.() ?? null;

  return (
    <dialog
      ref={dialogRef}
      className="dialog import-dialog"
      data-testid="import-dialog"
      onClick={handleBackdropClick}
    >
      <div className="dialog-content">
        <h2 className="dialog-title">Import from ArchiMate</h2>
        {renderStep()}
      </div>
    </dialog>
  );
};
