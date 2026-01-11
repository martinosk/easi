import React, { useState, useMemo } from 'react';
import { useAppStore } from '../../../store/appStore';
import { EditRealizationDialog } from './EditRealizationDialog';
import { DetailField } from '../../../components/shared/DetailField';
import { useCapabilities, useRealizationsForComponents } from '../../capabilities/hooks/useCapabilities';
import { useComponents } from '../../components/hooks/useComponents';
import { useCurrentView } from '../../views/hooks/useCurrentView';
import type { CapabilityRealization, Capability, Component } from '../../../api/types';

const REALIZATION_PREFIX = 'realization-';

const LEVEL_DISPLAY_MAP: Record<string, string> = {
  Full: 'Full (100%)',
  Partial: 'Partial',
  Planned: 'Planned',
};

const getLevelDisplay = (level: string): string => LEVEL_DISPLAY_MAP[level] ?? level;

const isRealizationEdge = (edgeId: string | null): boolean =>
  edgeId !== null && edgeId.startsWith(REALIZATION_PREFIX);

const extractRealizationId = (edgeId: string): string =>
  edgeId.replace(REALIZATION_PREFIX, '');

interface RealizationData {
  realization: CapabilityRealization;
  capability: Capability | undefined;
  component: Component | undefined;
  formattedDate: string;
  isInherited: boolean;
}

const getRealizationData = (
  selectedEdgeId: string | null,
  capabilityRealizations: CapabilityRealization[],
  capabilities: Capability[],
  components: Component[]
): RealizationData | null => {
  if (!isRealizationEdge(selectedEdgeId)) {
    return null;
  }

  const realizationId = extractRealizationId(selectedEdgeId!);
  const realization = capabilityRealizations.find((r) => r.id === realizationId);

  if (!realization) {
    return null;
  }

  const capability = capabilities.find((c) => c.id === realization.capabilityId);
  const component = components.find((c) => c.id === realization.componentId);
  const formattedDate = new Date(realization.linkedAt).toLocaleString();
  const isInherited = realization.origin === 'Inherited';

  return { realization, capability, component, formattedDate, isInherited };
};

export const RealizationDetails: React.FC = () => {
  const selectedEdgeId = useAppStore((state) => state.selectedEdgeId);
  const { currentView } = useCurrentView();
  const { data: components = [] } = useComponents();
  const { data: capabilities = [] } = useCapabilities();
  const [showEditDialog, setShowEditDialog] = useState(false);

  const componentIdsInView = useMemo(() =>
    currentView?.components.map((vc) => vc.componentId) || [],
    [currentView?.components]
  );
  const { data: capabilityRealizations = [] } = useRealizationsForComponents(componentIdsInView);

  const data = getRealizationData(selectedEdgeId, capabilityRealizations, capabilities, components);

  if (!data) {
    return null;
  }

  const { realization, capability, component, formattedDate, isInherited } = data;
  const canEdit = !isInherited && realization._links?.edit !== undefined;

  return (
    <div className="detail-panel">
      <div className="detail-header">
        <h3 className="detail-title">Realization Details</h3>
      </div>

      <div className="detail-content">
        {canEdit && (
          <div className="detail-actions">
            <button
              className="btn btn-secondary btn-small"
              onClick={() => setShowEditDialog(true)}
            >
              Edit
            </button>
          </div>
        )}

        <DetailField label="Capability">{capability?.name || 'Unknown'}</DetailField>
        <DetailField label="Application">{component?.name || 'Unknown'}</DetailField>
        <DetailField label="Realization Level">
          <span className={`level-badge level-${realization.realizationLevel.toLowerCase()}`}>
            {getLevelDisplay(realization.realizationLevel)}
          </span>
        </DetailField>
        <DetailField label="Origin">
          <span className={`origin-badge ${isInherited ? 'origin-inherited' : 'origin-direct'}`}>
            {realization.origin}
          </span>
        </DetailField>
        {realization.notes && (
          <DetailField label="Notes">{realization.notes}</DetailField>
        )}
        <DetailField label="Linked">
          <span className="detail-date">{formattedDate}</span>
        </DetailField>
        <DetailField label="ID">
          <span className="detail-id">{realization.id}</span>
        </DetailField>

        {isInherited && (
          <div className="detail-info">
            This is an inherited realization. It was automatically created when
            an application was linked to a child capability. To edit or delete,
            modify the original direct realization.
          </div>
        )}
      </div>

      <EditRealizationDialog
        isOpen={showEditDialog}
        onClose={() => setShowEditDialog(false)}
        realization={realization}
      />
    </div>
  );
};
