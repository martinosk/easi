import React, { useState, useEffect, useRef } from 'react';
import { ImportUploadStep } from './ImportUploadStep';
import { ImportPreviewStep } from './ImportPreviewStep';
import { ImportProgressStep } from './ImportProgressStep';
import { ImportResultsStep } from './ImportResultsStep';
import { useImportSession } from '../hooks/useImportSession';
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

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;

    if (isOpen) {
      dialog.showModal();
    } else {
      dialog.close();
    }
  }, [isOpen]);

  useEffect(() => {
    if (!session) {
      setCurrentStep('upload');
      return;
    }

    switch (session.status) {
      case 'pending':
        if (session.preview) {
          setCurrentStep('preview');
        }
        break;
      case 'importing':
        setCurrentStep('progress');
        break;
      case 'completed':
      case 'failed':
        setCurrentStep('results');
        break;
    }
  }, [session]);

  const handleUpload = async (file: File, businessDomainId?: string) => {
    await createSession({
      file,
      sourceFormat: 'archimate-openexchange',
      businessDomainId,
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

  const renderStep = () => {
    switch (currentStep) {
      case 'upload':
        return (
          <ImportUploadStep
            businessDomains={businessDomains}
            isLoading={isLoading}
            error={error}
            onUpload={handleUpload}
            onCancel={handleCancel}
          />
        );

      case 'preview':
        if (!session?.preview) return null;
        return (
          <ImportPreviewStep
            preview={session.preview}
            onConfirm={handleConfirm}
            onCancel={handleCancel}
            isLoading={isLoading}
          />
        );

      case 'progress':
        if (!session?.progress) return null;
        return <ImportProgressStep progress={session.progress} />;

      case 'results':
        if (!session?.result) return null;
        return <ImportResultsStep result={session.result} onClose={handleClose} />;

      default:
        return null;
    }
  };

  return (
    <dialog ref={dialogRef} className="dialog import-dialog" data-testid="import-dialog">
      <div className="dialog-content">
        <h2 className="dialog-title">Import from ArchiMate</h2>
        {renderStep()}
      </div>
    </dialog>
  );
};
