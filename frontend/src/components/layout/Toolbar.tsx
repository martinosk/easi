import React, { useState } from 'react';
import { EdgeTypeSelector, ColorSchemeSelector } from '../../features/views';
import { ImportButton, ImportDialog } from '../../features/importing';
import { useBusinessDomains } from '../../features/business-domains';

export const Toolbar: React.FC = () => {
  const [isImportDialogOpen, setIsImportDialogOpen] = useState(false);
  const { domains } = useBusinessDomains();

  return (
    <>
      <div className="toolbar">
        <div className="toolbar-left">
          <EdgeTypeSelector />
          <ColorSchemeSelector />
          <ImportButton onClick={() => setIsImportDialogOpen(true)} />
        </div>
      </div>
      <ImportDialog
        isOpen={isImportDialogOpen}
        onClose={() => setIsImportDialogOpen(false)}
        businessDomains={domains}
      />
    </>
  );
};
