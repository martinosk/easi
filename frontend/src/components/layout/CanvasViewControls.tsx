import { Divider, Group } from '@mantine/core';
import { lazy, Suspense, useState } from 'react';
import { useBusinessDomains } from '../../features/business-domains';
import { ImportButton } from '../../features/importing';
import { ColorSchemeSelector, EdgeTypeSelector } from '../../features/views';

const ImportDialog = lazy(() =>
  import('../../features/importing/components/ImportDialog').then((module) => ({ default: module.ImportDialog })),
);

export function CanvasViewControls() {
  const [isImportDialogOpen, setIsImportDialogOpen] = useState(false);
  const { domains } = useBusinessDomains();

  return (
    <>
      <Group gap="xs" wrap="nowrap" data-testid="canvas-view-controls">
        <EdgeTypeSelector />
        <ColorSchemeSelector />
        <Divider orientation="vertical" />
        <ImportButton onClick={() => setIsImportDialogOpen(true)} />
      </Group>
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
}
