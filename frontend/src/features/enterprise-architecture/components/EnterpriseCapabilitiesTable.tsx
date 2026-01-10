import React, { useState, useCallback } from 'react';
import toast from 'react-hot-toast';
import type { EnterpriseCapability, EnterpriseCapabilityId } from '../types';
import type { Capability } from '../../../api/types';

interface EnterpriseCapabilitiesTableProps {
  capabilities: EnterpriseCapability[];
  onSelect: (capability: EnterpriseCapability) => void;
  onDelete: (capability: EnterpriseCapability) => void;
  selectedId?: string;
  isDockPanelOpen?: boolean;
  onLinkCapability?: (enterpriseCapabilityId: EnterpriseCapabilityId, domainCapability: Capability) => void;
}

export const EnterpriseCapabilitiesTable = React.memo<EnterpriseCapabilitiesTableProps>(({
  capabilities,
  onSelect,
  onDelete,
  selectedId,
  isDockPanelOpen = false,
  onLinkCapability,
}) => {
  const hasAnyDeletable = capabilities.some(cap => cap._links?.delete);
  const [dragOverRowId, setDragOverRowId] = useState<string | null>(null);

  const handleKeyDown = (e: React.KeyboardEvent, capability: EnterpriseCapability) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      onSelect(capability);
    }
  };

  const canAcceptLink = useCallback((capability: EnterpriseCapability) => {
    return isDockPanelOpen && capability._links?.['x-create-link'] !== undefined;
  }, [isDockPanelOpen]);

  const handleDragOver = useCallback((e: React.DragEvent, capability: EnterpriseCapability) => {
    if (!canAcceptLink(capability)) return;
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
    setDragOverRowId(capability.id);
  }, [canAcceptLink]);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    if (!isDockPanelOpen) return;
    const relatedTarget = e.relatedTarget as HTMLElement;
    if (!relatedTarget || !e.currentTarget.contains(relatedTarget)) {
      setDragOverRowId(null);
    }
  }, [isDockPanelOpen]);

  const handleDrop = useCallback((e: React.DragEvent, enterpriseCapability: EnterpriseCapability) => {
    if (!canAcceptLink(enterpriseCapability) || !onLinkCapability) return;
    e.preventDefault();
    setDragOverRowId(null);

    try {
      const data = e.dataTransfer.getData('application/json');
      if (!data) return;
      const domainCapability = JSON.parse(data) as Capability;
      onLinkCapability(enterpriseCapability.id, domainCapability);
    } catch {
      toast.error('Invalid drag data');
    }
  }, [canAcceptLink, onLinkCapability]);

  const getRowClassName = (capability: EnterpriseCapability) => {
    const classes: string[] = [];
    if (selectedId === capability.id) classes.push('selected');
    if (dragOverRowId === capability.id) classes.push('drag-over');
    return classes.join(' ');
  };

  return (
    <div className="capabilities-table-container">
      <table className="capabilities-table" data-testid="enterprise-capabilities-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Category</th>
            <th>Linked Capabilities</th>
            <th>Domains</th>
            {hasAnyDeletable && <th>Actions</th>}
          </tr>
        </thead>
        <tbody>
          {capabilities.map((capability) => (
            <tr
              key={capability.id}
              className={getRowClassName(capability)}
              onClick={() => onSelect(capability)}
              onKeyDown={(e) => handleKeyDown(e, capability)}
              onDragOver={(e) => handleDragOver(e, capability)}
              onDragLeave={handleDragLeave}
              onDrop={(e) => handleDrop(e, capability)}
              tabIndex={0}
              role="button"
              aria-label={`Select enterprise capability ${capability.name}`}
              data-testid={`capability-row-${capability.id}`}
            >
              <td className="capability-name">{capability.name}</td>
              <td className="capability-category">{capability.category || '-'}</td>
              <td className="capability-links">{capability.linkCount}</td>
              <td className="capability-domains">{capability.domainCount}</td>
              {hasAnyDeletable && (
                <td className="capability-actions">
                  {capability._links?.delete && (
                    <button
                      type="button"
                      className="btn btn-icon btn-danger"
                      onClick={(e) => {
                        e.stopPropagation();
                        onDelete(capability);
                      }}
                      title="Delete capability"
                      data-testid={`delete-capability-${capability.id}`}
                    >
                      <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" width="16" height="16">
                        <path d="M19 7L18.1327 19.1425C18.0579 20.1891 17.187 21 16.1378 21H7.86224C6.81296 21 5.94208 20.1891 5.86732 19.1425L5 7M10 11V17M14 11V17M15 7V4C15 3.44772 14.5523 3 14 3H10C9.44772 3 9 3.44772 9 4V7M4 7H20" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
                      </svg>
                    </button>
                  )}
                </td>
              )}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
});

EnterpriseCapabilitiesTable.displayName = 'EnterpriseCapabilitiesTable';
