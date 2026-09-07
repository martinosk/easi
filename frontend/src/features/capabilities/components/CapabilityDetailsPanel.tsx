import type React from 'react';
import type { CapabilityId, ComponentId } from '../../../api/types';
import { DetailPanelFailure, DetailPanelLoading } from '../../../components/shared/DetailPanelStatus';
import { useCapabilities, useCapability } from '../hooks/useCapabilities';
import { CapabilityDetailsContent } from './CapabilityDetailsContent';

export interface CapabilityDetailsPanelProps {
  capabilityId: string;
  viewMembership?: React.ReactNode;
  domainContext?: React.ReactNode;
  onApplicationClick?: (componentId: ComponentId) => void;
}

export function CapabilityDetailsPanel({
  capabilityId,
  viewMembership,
  domainContext,
  onApplicationClick,
}: CapabilityDetailsPanelProps) {
  const id = capabilityId as CapabilityId;
  const listQuery = useCapabilities();
  const fromList = listQuery.data?.find((c) => c.id === id);
  const detailQuery = useCapability(listQuery.isSuccess && !fromList ? id : undefined);

  const capability = fromList ?? detailQuery.data;

  if (capability) {
    return (
      <CapabilityDetailsContent
        capability={capability}
        viewMembership={viewMembership}
        domainContext={domainContext}
        onApplicationClick={onApplicationClick}
      />
    );
  }
  if (listQuery.isPending || detailQuery.isPending) return <DetailPanelLoading />;
  return <DetailPanelFailure message="Failed to load capability" />;
}
