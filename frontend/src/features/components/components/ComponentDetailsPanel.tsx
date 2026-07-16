import { useState } from 'react';
import type { ComponentId } from '../../../api/types';
import { DetailPanelFailure, DetailPanelLoading } from '../../../components/shared/DetailPanelStatus';
import { useCapabilities, useCapabilitiesByComponent } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../hooks/useComponents';
import { ComponentDetailsContent } from './ComponentDetails';
import { EditComponentDialog } from './EditComponentDialog';

interface ComponentDetailsPanelProps {
  componentId: string;
}

export function ComponentDetailsPanel({ componentId }: ComponentDetailsPanelProps) {
  const componentsQuery = useComponents();
  const { data: capabilities = [] } = useCapabilities();
  const { data: realizations = [] } = useCapabilitiesByComponent(componentId as ComponentId);
  const [editOpen, setEditOpen] = useState(false);
  const [addExpertOpen, setAddExpertOpen] = useState(false);

  const component = (componentsQuery.data ?? []).find((c) => c.id === componentId);

  if (componentsQuery.isLoading) return <DetailPanelLoading />;
  if (!component) return <DetailPanelFailure message="Failed to load application" />;

  return (
    <>
      <ComponentDetailsContent
        component={component}
        realizations={realizations}
        capabilities={capabilities}
        onEdit={() => setEditOpen(true)}
        onAddExpert={() => setAddExpertOpen(true)}
        isAddExpertOpen={addExpertOpen}
        onCloseAddExpert={() => setAddExpertOpen(false)}
      />
      <EditComponentDialog isOpen={editOpen} onClose={() => setEditOpen(false)} component={component} />
    </>
  );
}
