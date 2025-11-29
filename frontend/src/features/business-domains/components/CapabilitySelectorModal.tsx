import { useState, useEffect, useRef } from 'react';
import { useCapabilityTree, type CapabilityTreeNode } from '../hooks/useCapabilityTree';
import type { Capability, CapabilityId } from '../../../api/types';

interface CapabilitySelectorModalProps {
  isOpen: boolean;
  onClose: () => void;
  currentAssociations: Capability[];
  onSave: (selectedCapabilities: Capability[]) => Promise<void>;
}

interface CapabilityTreeProps {
  nodes: CapabilityTreeNode[];
  selectedIds: Set<CapabilityId>;
  disabledIds: Set<CapabilityId>;
  orphanedIds: Set<CapabilityId>;
  onToggle: (capability: Capability) => void;
  level: number;
}

function CapabilityTree({ nodes, selectedIds, disabledIds, orphanedIds, onToggle, level }: CapabilityTreeProps) {
  return (
    <div className="capability-tree" style={{ marginLeft: level > 0 ? '20px' : '0' }}>
      {nodes.map((node) => {
        const isL1 = node.capability.level === 'L1';
        const isSelected = selectedIds.has(node.capability.id);
        const isDisabled = disabledIds.has(node.capability.id);
        const isOrphaned = orphanedIds.has(node.capability.id);

        return (
          <div key={node.capability.id} className="capability-tree-item">
            <div className="capability-tree-row">
              {isL1 ? (
                <label className="capability-checkbox-label">
                  <input
                    type="checkbox"
                    checked={isSelected}
                    disabled={isDisabled}
                    onChange={() => onToggle(node.capability)}
                    data-testid={`capability-checkbox-${node.capability.id}`}
                  />
                  <span className={isOrphaned ? 'capability-orphaned' : ''}>
                    {node.capability.name}
                    {isOrphaned && <span className="orphan-warning" title="No child capabilities"> ⚠️</span>}
                  </span>
                </label>
              ) : (
                <span className="capability-child-name">{node.capability.name}</span>
              )}
            </div>
            {node.children.length > 0 && (
              <CapabilityTree
                nodes={node.children}
                selectedIds={selectedIds}
                disabledIds={disabledIds}
                orphanedIds={orphanedIds}
                onToggle={onToggle}
                level={level + 1}
              />
            )}
          </div>
        );
      })}
    </div>
  );
}

export function CapabilitySelectorModal({ isOpen, onClose, currentAssociations, onSave }: CapabilitySelectorModalProps) {
  const { tree, isLoading, error, orphanedL1Ids } = useCapabilityTree();
  const [selectedIds, setSelectedIds] = useState<Set<CapabilityId>>(new Set());
  const [selectedCapabilities, setSelectedCapabilities] = useState<Map<CapabilityId, Capability>>(new Map());
  const [isSaving, setIsSaving] = useState(false);
  const dialogRef = useRef<HTMLDialogElement>(null);

  const currentAssociationIds = new Set(currentAssociations.map((c) => c.id));

  useEffect(() => {
    const dialog = dialogRef.current;
    if (!dialog) return;

    if (isOpen) {
      dialog.showModal();
    } else {
      dialog.close();
    }
  }, [isOpen]);

  const handleToggle = (capability: Capability) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(capability.id)) {
        next.delete(capability.id);
        setSelectedCapabilities((prevCaps) => {
          const nextCaps = new Map(prevCaps);
          nextCaps.delete(capability.id);
          return nextCaps;
        });
      } else {
        next.add(capability.id);
        setSelectedCapabilities((prevCaps) => {
          const nextCaps = new Map(prevCaps);
          nextCaps.set(capability.id, capability);
          return nextCaps;
        });
      }
      return next;
    });
  };

  const handleSave = async () => {
    setIsSaving(true);
    try {
      await onSave(Array.from(selectedCapabilities.values()));
      setSelectedIds(new Set());
      setSelectedCapabilities(new Map());
      onClose();
    } catch (err) {
      console.error('Failed to save:', err);
    } finally {
      setIsSaving(false);
    }
  };

  const handleCancel = () => {
    setSelectedIds(new Set());
    setSelectedCapabilities(new Map());
    onClose();
  };

  return (
    <dialog ref={dialogRef} className="dialog capability-selector-modal" onClose={handleCancel} data-testid="capability-selector-modal">
      <div className="dialog-content">
        <h2 className="dialog-title">Add Capabilities</h2>

        {isLoading && <div className="loading-message">Loading capabilities...</div>}

        {error && (
          <div className="error-message" data-testid="capability-selector-error">
            {error.message}
          </div>
        )}

        {!isLoading && !error && (
          <div className="capability-selector-tree" data-testid="capability-selector-tree">
            <p className="capability-selector-hint">Select L1 capabilities to associate with this domain:</p>
            <CapabilityTree
              nodes={tree}
              selectedIds={selectedIds}
              disabledIds={currentAssociationIds}
              orphanedIds={orphanedL1Ids}
              onToggle={handleToggle}
              level={0}
            />
          </div>
        )}

        <div className="dialog-actions">
          <button
            type="button"
            className="btn btn-secondary"
            onClick={handleCancel}
            disabled={isSaving}
            data-testid="capability-selector-cancel"
          >
            Cancel
          </button>
          <button
            type="button"
            className="btn btn-primary"
            onClick={handleSave}
            disabled={isSaving || selectedIds.size === 0}
            data-testid="capability-selector-save"
          >
            {isSaving ? 'Saving...' : `Add ${selectedIds.size} ${selectedIds.size === 1 ? 'Capability' : 'Capabilities'}`}
          </button>
        </div>
      </div>
    </dialog>
  );
}
