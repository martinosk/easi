import React, { useState } from 'react';
import { useRealizationDetails } from '../hooks/useRealizationDetails';
import { EditRealizationDialog } from './EditRealizationDialog';
import { RealizationDetailsContent } from './RealizationDetailsContent';

export const RealizationDetails: React.FC = () => {
  const data = useRealizationDetails();
  const [showEditDialog, setShowEditDialog] = useState(false);

  if (!data) return null;

  return (
    <>
      <RealizationDetailsContent data={data} onEditClick={() => setShowEditDialog(true)} />
      <EditRealizationDialog
        isOpen={showEditDialog}
        onClose={() => setShowEditDialog(false)}
        realization={data.realization}
      />
    </>
  );
};
