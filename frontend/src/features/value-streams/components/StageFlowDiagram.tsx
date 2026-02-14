import { useCallback, useState } from 'react';
import type { ValueStreamStage, StageCapabilityMapping } from '../../../api/types';
import { StageColumn } from './StageColumn';
import { AddStageButton } from './AddStageButton';
import { useRemoveStageCapability } from '../hooks/useValueStreamStages';
import './StageFlowDiagram.css';

interface StageFlowDiagramProps {
  valueStreamId: string;
  stages: ValueStreamStage[];
  stageCapabilities: StageCapabilityMapping[];
  canWrite: boolean;
  onAddStage: (position?: number) => void;
  onEditStage: (stage: ValueStreamStage) => void;
  onDeleteStage: (stage: ValueStreamStage) => void;
  onReorder: (orderedStageIds: string[]) => void;
  onAddCapability?: (stageId: string, capabilityId: string) => void;
}

function tryHandleCapabilityDrop(
  e: React.DragEvent,
  targetStageId: string,
  onAddCapability?: (stageId: string, capabilityId: string) => void,
): boolean {
  const json = e.dataTransfer.getData('application/json');
  if (!json || !onAddCapability) return false;
  try {
    const capability = JSON.parse(json);
    if (capability?.id) {
      onAddCapability(targetStageId, capability.id);
      return true;
    }
  } catch { /* not a capability drop */ }
  return false;
}

function reorderStageIds(sortedStages: ValueStreamStage[], draggedId: string, targetId: string): string[] | null {
  const ordered: string[] = sortedStages.map(s => s.id);
  const fromIndex = ordered.indexOf(draggedId);
  const toIndex = ordered.indexOf(targetId);
  if (fromIndex < 0 || toIndex < 0) return null;
  ordered.splice(fromIndex, 1);
  ordered.splice(toIndex, 0, draggedId);
  return ordered;
}

function applyStageReorder(
  draggedStageId: string | null,
  targetStageId: string,
  sortedStages: ValueStreamStage[],
  onReorder: (orderedStageIds: string[]) => void,
): void {
  if (!draggedStageId || draggedStageId === targetStageId) return;
  const ordered = reorderStageIds(sortedStages, draggedStageId, targetStageId);
  if (ordered) onReorder(ordered);
}

function groupCapabilitiesByStage(stageCapabilities: StageCapabilityMapping[]) {
  const capsByStage = new Map<string, StageCapabilityMapping[]>();
  for (const cap of stageCapabilities) {
    const list = capsByStage.get(cap.stageId) || [];
    list.push(cap);
    capsByStage.set(cap.stageId, list);
  }
  return capsByStage;
}

interface StageConnectorProps {
  canWrite: boolean;
  position: number;
  onInsert: (position: number) => void;
  index: number;
}

function StageConnector({ canWrite, position, onInsert, index }: StageConnectorProps) {
  return (
    <div className="stage-connector-group">
      <div className="stage-connector" />
      {canWrite && (
        <button
          type="button"
          className="stage-insert-btn"
          data-testid={`insert-stage-btn-${index}`}
          onClick={() => onInsert(position)}
          title="Insert stage here"
        >
          <svg viewBox="0 0 24 24" fill="none" width="14" height="14">
            <path d="M12 5V19M5 12H19" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
          </svg>
        </button>
      )}
    </div>
  );
}

function EmptyStages({ canWrite, onAddStage }: { canWrite: boolean; onAddStage: (position?: number) => void }) {
  return (
    <div className="stage-flow-empty" data-testid="empty-stages">
      <svg viewBox="0 0 24 24" fill="none" width="48" height="48">
        <path d="M22 12H18L15 21L9 3L6 12H2" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
      </svg>
      <h3>No stages yet</h3>
      <p>Add stages to model the flow of this value stream.</p>
      {canWrite && <AddStageButton onClick={() => onAddStage()} />}
    </div>
  );
}

export function StageFlowDiagram({
  valueStreamId,
  stages,
  stageCapabilities,
  canWrite,
  onAddStage,
  onEditStage,
  onDeleteStage,
  onReorder,
  onAddCapability,
}: StageFlowDiagramProps) {
  const removeCapMutation = useRemoveStageCapability();
  const [draggedStageId, setDraggedStageId] = useState<string | null>(null);

  const sortedStages = [...stages].sort((a, b) => a.position - b.position);
  const capsByStage = groupCapabilitiesByStage(stageCapabilities);

  const handleDragStart = useCallback((e: React.DragEvent, stageId: string) => {
    setDraggedStageId(stageId);
    e.dataTransfer.effectAllowed = 'move';
  }, []);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = e.dataTransfer.effectAllowed === 'copy' ? 'copy' : 'move';
  }, []);

  const handleDrop = useCallback((e: React.DragEvent, targetStageId: string) => {
    e.preventDefault();
    if (!tryHandleCapabilityDrop(e, targetStageId, onAddCapability)) {
      applyStageReorder(draggedStageId, targetStageId, sortedStages, onReorder);
    }
    setDraggedStageId(null);
  }, [draggedStageId, sortedStages, onReorder, onAddCapability]);

  const handleRemoveCapability = useCallback(async (mapping: StageCapabilityMapping) => {
    await removeCapMutation.mutateAsync({ mapping, valueStreamId });
  }, [removeCapMutation, valueStreamId]);

  if (stages.length === 0) {
    return <EmptyStages canWrite={canWrite} onAddStage={onAddStage} />;
  }

  return (
    <div className="stage-flow" data-testid="stage-flow-diagram">
      <div className="stage-flow-scroll">
        {sortedStages.map((stage, i) => (
          <div key={stage.id} className="stage-flow-item">
            {i > 0 && <StageConnector canWrite={canWrite} position={stage.position} onInsert={onAddStage} index={i} />}
            <StageColumn
              stage={stage}
              capabilities={capsByStage.get(stage.id) || []}
              canWrite={canWrite}
              onEdit={onEditStage}
              onDelete={onDeleteStage}
              onRemoveCapability={handleRemoveCapability}
              onDragStart={handleDragStart}
              onDragOver={handleDragOver}
              onDrop={handleDrop}
            />
          </div>
        ))}
        {canWrite && (
          <div className="stage-flow-item">
            <div className="stage-connector" />
            <AddStageButton onClick={() => onAddStage()} />
          </div>
        )}
      </div>
    </div>
  );
}
