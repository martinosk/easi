import { useCallback, useState } from 'react';
import type { ValueStreamStage, StageCapabilityMapping } from '../../../api/types';
import { StageColumn } from './StageColumn';
import { AddStageButton } from './AddStageButton';
import { useRemoveStageCapability } from '../hooks/useValueStreamStages';
import './StageFlowDiagram.css';

interface StageFlowDiagramProps {
  stages: ValueStreamStage[];
  stageCapabilities: StageCapabilityMapping[];
  canWrite: boolean;
  onAddStage: () => void;
  onEditStage: (stage: ValueStreamStage) => void;
  onDeleteStage: (stage: ValueStreamStage) => void;
  onReorder: (orderedStageIds: string[]) => void;
  onAddCapability?: (stageId: string, capabilityId: string) => void;
}

export function StageFlowDiagram({
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

  const capsByStage = new Map<string, StageCapabilityMapping[]>();
  for (const cap of stageCapabilities) {
    const list = capsByStage.get(cap.stageId) || [];
    list.push(cap);
    capsByStage.set(cap.stageId, list);
  }

  const handleDragStart = useCallback((e: React.DragEvent, stageId: string) => {
    setDraggedStageId(stageId);
    e.dataTransfer.effectAllowed = 'move';
  }, []);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
  }, []);

  const handleDrop = useCallback((e: React.DragEvent, targetStageId: string) => {
    e.preventDefault();

    const json = e.dataTransfer.getData('application/json');
    if (json && onAddCapability) {
      try {
        const capability = JSON.parse(json);
        if (capability?.id) {
          onAddCapability(targetStageId, capability.id);
          setDraggedStageId(null);
          return;
        }
      } catch { /* not a capability drop, continue with stage reorder */ }
    }

    if (!draggedStageId || draggedStageId === targetStageId) return;

    const ordered = sortedStages.map(s => s.id as string);
    const fromIndex = ordered.indexOf(draggedStageId);
    const toIndex = ordered.indexOf(targetStageId);
    if (fromIndex < 0 || toIndex < 0) return;

    ordered.splice(fromIndex, 1);
    ordered.splice(toIndex, 0, draggedStageId);
    onReorder(ordered);
    setDraggedStageId(null);
  }, [draggedStageId, sortedStages, onReorder, onAddCapability]);

  const handleRemoveCapability = useCallback((mapping: StageCapabilityMapping) => {
    removeCapMutation.mutate(mapping);
  }, [removeCapMutation]);

  if (stages.length === 0) {
    return (
      <div className="stage-flow-empty" data-testid="empty-stages">
        <svg viewBox="0 0 24 24" fill="none" width="48" height="48">
          <path d="M22 12H18L15 21L9 3L6 12H2" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
        </svg>
        <h3>No stages yet</h3>
        <p>Add stages to model the flow of this value stream.</p>
        {canWrite && <AddStageButton onClick={onAddStage} />}
      </div>
    );
  }

  return (
    <div className="stage-flow" data-testid="stage-flow-diagram">
      <div className="stage-flow-scroll">
        {sortedStages.map((stage, i) => (
          <div key={stage.id} className="stage-flow-item">
            {i > 0 && <div className="stage-connector" />}
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
            <AddStageButton onClick={onAddStage} />
          </div>
        )}
      </div>
    </div>
  );
}
