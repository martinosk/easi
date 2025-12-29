import { useState, useCallback, useMemo } from 'react';
import type { BusinessDomain, CapabilityId, StrategyImportance, BusinessDomainId } from '../../../api/types';
import { useStrategyImportanceByDomainAndCapability } from '../hooks/useStrategyImportance';
import { useStrategyPillarsConfig } from '../../../hooks/useStrategyPillarsSettings';
import { SetImportanceDialog } from './SetImportanceDialog';

interface StrategicImportanceSectionProps {
  domain: BusinessDomain;
  capabilityId: CapabilityId;
  capabilityName: string;
}

function renderImportanceStars(importance: number): string {
  return '★'.repeat(importance) + '☆'.repeat(5 - importance);
}

interface ImportanceRatingCardProps {
  rating: StrategyImportance;
  pillarName: string;
  onEdit: (rating: StrategyImportance) => void;
}

function ImportanceRatingCard({ rating, pillarName, onEdit }: ImportanceRatingCardProps) {
  return (
    <div
      style={{
        padding: '0.75rem',
        backgroundColor: 'var(--color-gray-50)',
        borderRadius: '6px',
        border: '1px solid var(--color-gray-200)',
      }}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <div style={{ fontWeight: 500, fontSize: '0.875rem' }}>{pillarName}</div>
          <div style={{ color: 'var(--color-amber-500)', fontSize: '0.875rem' }}>
            {renderImportanceStars(rating.importance)}{' '}
            <span style={{ color: 'var(--color-gray-600)' }}>{rating.importanceLabel}</span>
          </div>
        </div>
        <button
          onClick={() => onEdit(rating)}
          style={{
            padding: '4px 8px',
            border: '1px solid var(--color-gray-300)',
            borderRadius: '4px',
            backgroundColor: 'white',
            cursor: 'pointer',
            fontSize: '12px',
          }}
          data-testid={`edit-importance-${rating.pillarId}`}
        >
          Edit
        </button>
      </div>
      {rating.rationale && (
        <div style={{ color: 'var(--color-gray-600)', fontSize: '0.75rem', marginTop: '0.5rem', fontStyle: 'italic' }}>
          {rating.rationale}
        </div>
      )}
    </div>
  );
}

export function StrategicImportanceSection({ domain, capabilityId, capabilityName }: StrategicImportanceSectionProps) {
  const [importanceDialogOpen, setImportanceDialogOpen] = useState(false);
  const [selectedImportance, setSelectedImportance] = useState<StrategyImportance | undefined>();

  const { data: importanceRatings = [], isLoading } = useStrategyImportanceByDomainAndCapability(
    domain.id as BusinessDomainId,
    capabilityId
  );
  const { data: pillarsConfig } = useStrategyPillarsConfig();

  const pillarNameMap = useMemo(() => {
    const map = new Map<string, string>();
    pillarsConfig?.data.forEach((p) => map.set(p.id, p.name));
    return map;
  }, [pillarsConfig]);

  const getPillarName = useCallback(
    (rating: StrategyImportance) => rating.pillarName || pillarNameMap.get(rating.pillarId) || '',
    [pillarNameMap]
  );

  const handleAddImportance = useCallback(() => {
    setSelectedImportance(undefined);
    setImportanceDialogOpen(true);
  }, []);

  const handleEditImportance = useCallback((importance: StrategyImportance) => {
    setSelectedImportance(importance);
    setImportanceDialogOpen(true);
  }, []);

  const handleCloseDialog = useCallback(() => {
    setImportanceDialogOpen(false);
    setSelectedImportance(undefined);
  }, []);

  return (
    <>
      <div className="detail-section" style={{ marginTop: '1.5rem', borderTop: '1px solid var(--color-gray-200)', paddingTop: '1rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
          <span className="detail-label" style={{ fontWeight: 600 }}>
            Strategic Importance ({domain.name})
          </span>
        </div>

        {isLoading ? (
          <div style={{ color: 'var(--color-gray-500)', fontSize: '0.875rem' }}>Loading...</div>
        ) : importanceRatings.length === 0 ? (
          <div style={{ color: 'var(--color-gray-500)', fontSize: '0.875rem', marginBottom: '0.5rem' }}>
            No ratings yet
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', marginBottom: '0.75rem' }}>
            {importanceRatings.map((rating) => (
              <ImportanceRatingCard
                key={rating.id}
                rating={rating}
                pillarName={getPillarName(rating)}
                onEdit={handleEditImportance}
              />
            ))}
          </div>
        )}

        <button
          onClick={handleAddImportance}
          style={{
            padding: '6px 12px',
            border: '1px solid var(--color-blue-500)',
            borderRadius: '4px',
            backgroundColor: 'transparent',
            color: 'var(--color-blue-500)',
            cursor: 'pointer',
            fontSize: '13px',
            width: '100%',
          }}
          data-testid="add-importance-btn"
        >
          + Rate Another Pillar
        </button>
      </div>

      <SetImportanceDialog
        isOpen={importanceDialogOpen}
        onClose={handleCloseDialog}
        domainId={domain.id}
        domainName={domain.name}
        capabilityId={capabilityId}
        capabilityName={capabilityName}
        existingImportance={selectedImportance}
        existingPillarIds={importanceRatings.map((r) => r.pillarId)}
      />
    </>
  );
}
