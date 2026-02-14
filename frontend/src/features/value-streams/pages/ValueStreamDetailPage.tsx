import { useParams, useNavigate } from 'react-router-dom';
import { useMemo } from 'react';
import { useValueStreamDetail } from '../hooks/useValueStreamStages';
import { useStageOperations } from '../hooks/useStageOperations';
import { useUserStore } from '../../../store/userStore';
import { hasLink } from '../../../utils/hateoas';
import { StageFlowDiagram } from '../components/StageFlowDiagram';
import { StageFormOverlay } from '../components/StageFormOverlay';
import { CapabilitySidebar } from '../components/CapabilitySidebar';
import { SummaryBar } from '../components/SummaryBar';
import type { ValueStreamDetail } from '../../../api/types';
import { toValueStreamId } from '../../../api/types';
import './ValueStreamDetailPage.css';

function LoadingState() {
  return (
    <div className="vsd-page">
      <div className="vsd-loading">Loading value stream...</div>
    </div>
  );
}

function ErrorState({ message }: { message?: string }) {
  return (
    <div className="vsd-page">
      <div className="vsd-error">{message || 'Value stream not found'}</div>
    </div>
  );
}

interface DetailContentProps {
  detail: ValueStreamDetail;
  canWrite: boolean;
}

function DetailContent({ detail, canWrite }: DetailContentProps) {
  const navigate = useNavigate();
  const ops = useStageOperations(detail);

  const mappedCapabilityIds = useMemo(
    () => new Set((detail.stageCapabilities ?? []).map(c => c.capabilityId as string)),
    [detail.stageCapabilities],
  );

  const uniqueCapCount = new Set(detail.stageCapabilities.map(c => c.capabilityId)).size;
  const canAddStage = canWrite && hasLink(detail, 'x-add-stage');

  return (
    <div className="vsd-page" data-testid="value-stream-detail-page">
      <div className="vsd-header">
        <button type="button" className="vsd-back-btn" onClick={() => navigate('/value-streams')}>
          <svg viewBox="0 0 24 24" fill="none" width="16" height="16">
            <path d="M19 12H5M12 19l-7-7 7-7" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
          Back to Value Streams
        </button>
        <div className="vsd-header-content">
          <div>
            <h1 className="vsd-title">{detail.name}</h1>
            {detail.description && <p className="vsd-description">{detail.description}</p>}
          </div>
        </div>
        <SummaryBar stageCount={detail.stages.length} capabilityCount={uniqueCapCount} />
      </div>

      {ops.isFormOpen && (
        <StageFormOverlay
          isEditing={ops.editingStage !== null}
          formData={ops.formData}
          onFormDataChange={ops.setFormData}
          onSubmit={ops.submitForm}
          onCancel={ops.closeForm}
        />
      )}

      <div className="vsd-content">
        <div className="vsd-main">
          <StageFlowDiagram
            stages={detail.stages}
            stageCapabilities={detail.stageCapabilities}
            canWrite={canAddStage}
            onAddStage={ops.openAddForm}
            onEditStage={ops.openEditForm}
            onDeleteStage={ops.deleteStage}
            onReorder={ops.reorderStages}
            onAddCapability={canWrite ? ops.addCapability : undefined}
          />
        </div>
        {canWrite && (
          <CapabilitySidebar
            mappedCapabilityIds={mappedCapabilityIds}
          />
        )}
      </div>
    </div>
  );
}

export function ValueStreamDetailPage() {
  const { valueStreamId } = useParams<{ valueStreamId: string }>();
  const id = valueStreamId ? toValueStreamId(valueStreamId) : undefined;
  const { data: detail, isLoading, error } = useValueStreamDetail(id);
  const hasPermission = useUserStore((state) => state.hasPermission);
  const canWrite = hasPermission('valuestreams:write');

  if (isLoading) return <LoadingState />;
  if (error) return <ErrorState message={`Failed to load: ${error.message}`} />;
  if (!detail) return <ErrorState />;

  return <DetailContent detail={detail} canWrite={canWrite} />;
}
