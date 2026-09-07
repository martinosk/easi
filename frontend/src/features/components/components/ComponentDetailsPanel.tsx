import type React from 'react';
import { toComponentId } from '../../../api/types';
import { DetailPanelFailure, DetailPanelLoading } from '../../../components/shared/DetailPanelStatus';
import { useCapabilities, useCapabilitiesByComponent } from '../../capabilities/hooks/useCapabilities';
import { useComponent, useComponents } from '../hooks/useComponents';
import { ComponentDetailsContent } from './ComponentDetailsContent';

interface ComponentDetailsPanelProps {
  componentId: string;
  viewMembership?: React.ReactNode;
}

export function ComponentDetailsPanel({ componentId, viewMembership }: ComponentDetailsPanelProps) {
  const id = toComponentId(componentId);
  const listQuery = useComponents();
  const fromList = listQuery.data?.find((c) => c.id === id);
  const detailQuery = useComponent(listQuery.isSuccess && !fromList ? id : undefined);
  const { data: capabilities = [] } = useCapabilities();
  const { data: realizations = [] } = useCapabilitiesByComponent(id);

  const component = fromList ?? detailQuery.data;

  if (component) {
    return (
      <ComponentDetailsContent
        component={component}
        realizations={realizations}
        capabilities={capabilities}
        viewMembership={viewMembership}
      />
    );
  }
  if (listQuery.isPending || detailQuery.isPending) return <DetailPanelLoading />;
  return <DetailPanelFailure message="Failed to load application" />;
}
