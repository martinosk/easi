import React, { lazy, Suspense, useState } from 'react';
import { EdgeTypeSelector, ColorSchemeSelector } from '../../features/views';
import { ImportButton } from '../../features/importing';
import { useBusinessDomains } from '../../features/business-domains';

const ImportDialog = lazy(() =>
  import('../../features/importing').then(module => ({ default: module.ImportDialog }))
);

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
      {isImportDialogOpen && (
        <Suspense fallback={null}>
          <ImportDialog
            isOpen={isImportDialogOpen}
            onClose={() => setIsImportDialogOpen(false)}
            businessDomains={domains}
          />
        </Suspense>
      )}
    </>
  );
};
