import { useState } from 'react';
import { useAuditHistory } from '../hooks/useAuditHistory';
import type { AuditEntry } from '../../../api/types';
import './AuditHistorySection.css';

interface AuditHistorySectionProps {
  aggregateId: string;
}

function formatDate(dateString: string): string {
  const date = new Date(dateString);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
}

function formatEventDataValue(value: unknown): string {
  if (value === null || value === undefined) {
    return '-';
  }
  if (typeof value === 'object') {
    return JSON.stringify(value);
  }
  return String(value);
}

function formatFieldName(key: string): string {
  return key
    .replace(/([a-z])([A-Z])/g, '$1 $2')
    .replace(/_/g, ' ')
    .replace(/^./, (str) => str.toUpperCase());
}

interface AuditEntryCardProps {
  entry: AuditEntry;
}

function AuditEntryCard({ entry }: AuditEntryCardProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const eventDataEntries = Object.entries(entry.eventData || {});

  return (
    <div className="audit-entry-card">
      <button
        type="button"
        className="audit-entry-header"
        onClick={() => setIsExpanded(!isExpanded)}
        aria-expanded={isExpanded}
      >
        <div className="audit-entry-summary">
          <span className="audit-event-name">{entry.displayName}</span>
          <span className="audit-meta">
            <span className="audit-actor">{entry.actorEmail}</span>
            <span className="audit-separator">•</span>
            <span className="audit-date">{formatDate(entry.occurredAt)}</span>
          </span>
        </div>
        {eventDataEntries.length > 0 && (
          <span className={`audit-expand-icon ${isExpanded ? 'expanded' : ''}`} aria-hidden="true">▸</span>
        )}
      </button>
      {isExpanded && eventDataEntries.length > 0 && (
        <div className="audit-entry-details">
          {eventDataEntries.map(([key, value]) => (
            <div key={key} className="audit-detail-row">
              <span className="audit-detail-label">{formatFieldName(key)}</span>
              <span className="audit-detail-value">{formatEventDataValue(value)}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export function AuditHistorySection({ aggregateId }: AuditHistorySectionProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const { data, isLoading, error } = useAuditHistory(aggregateId);

  const entries = data?.entries || [];
  const entryCount = entries.length;

  return (
    <div className="audit-history-section">
      <button
        type="button"
        className="audit-section-header"
        onClick={() => setIsExpanded(!isExpanded)}
        aria-expanded={isExpanded}
        aria-label={`History, ${entryCount} events`}
      >
        <span className="audit-section-title">
          History {entryCount > 0 && <span className="audit-count">({entryCount})</span>}
        </span>
        <span className={`audit-expand-icon ${isExpanded ? 'expanded' : ''}`} aria-hidden="true">▸</span>
      </button>

      {isExpanded && (
        <div className="audit-section-content">
          {isLoading ? (
            <div className="audit-loading">
              <div className="audit-spinner" />
              <span>Loading history...</span>
            </div>
          ) : error ? (
            <div className="audit-error">Failed to load history</div>
          ) : entries.length === 0 ? (
            <div className="audit-empty">No history available</div>
          ) : (
            <div className="audit-entries">
              {entries.map((entry) => (
                <AuditEntryCard key={entry.eventId} entry={entry} />
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
