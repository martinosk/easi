import type { ValueStreamStage, StageCapabilityMapping } from '../../../api/types';
import { hasLink } from '../../../utils/hateoas';
import { CapabilityChip } from './CapabilityChip';

interface StageColumnProps {
  stage: ValueStreamStage;
  capabilities: StageCapabilityMapping[];
  canWrite: boolean;
  onEdit: (stage: ValueStreamStage) => void;
  onDelete: (stage: ValueStreamStage) => void;
  onRemoveCapability: (mapping: StageCapabilityMapping) => void;
  onDragStart: (e: React.DragEvent, stageId: string) => void;
  onDragOver: (e: React.DragEvent) => void;
  onDrop: (e: React.DragEvent, stageId: string) => void;
}

export function StageColumn({
  stage,
  capabilities,
  canWrite,
  onEdit,
  onDelete,
  onRemoveCapability,
  onDragStart,
  onDragOver,
  onDrop,
}: StageColumnProps) {
  return (
    <div
      className="stage-column"
      data-testid={`stage-${stage.id}`}
      draggable={canWrite}
      onDragStart={(e) => onDragStart(e, stage.id)}
      onDragOver={onDragOver}
      onDrop={(e) => onDrop(e, stage.id)}
    >
      <div className="stage-header">
        <div className="stage-position">{stage.position}</div>
        <h3 className="stage-name">{stage.name}</h3>
        {canWrite && (
          <div className="stage-actions">
            {hasLink(stage, 'edit') && (
              <button type="button" className="stage-action-btn" onClick={() => onEdit(stage)} title="Edit stage">
                <svg viewBox="0 0 24 24" fill="none" width="14" height="14">
                  <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                  <path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </button>
            )}
            {hasLink(stage, 'delete') && (
              <button type="button" className="stage-action-btn stage-action-danger" onClick={() => onDelete(stage)} title="Delete stage">
                <svg viewBox="0 0 24 24" fill="none" width="14" height="14">
                  <path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                </svg>
              </button>
            )}
          </div>
        )}
      </div>
      {stage.description && <p className="stage-description">{stage.description}</p>}
      <div className="stage-capabilities">
        {capabilities.length > 0 ? (
          capabilities.map((cap) => (
            <CapabilityChip
              key={`${cap.stageId}-${cap.capabilityId}`}
              mapping={cap}
              canRemove={canWrite}
              onRemove={onRemoveCapability}
            />
          ))
        ) : (
          <span className="stage-no-caps">No capabilities mapped</span>
        )}
      </div>
    </div>
  );
}
