import React, { useMemo, useCallback } from 'react';
import type { View } from '../../../../api/types';
import { TreeSection } from '../TreeSection';
import type { EditingState } from '../../types';
import { useActiveUsers } from '../../../users/hooks/useUsers';

interface ViewsSectionProps {
  views: View[];
  currentView: View | null;
  isExpanded: boolean;
  onToggle: () => void;
  canCreateView: boolean;
  onCreateView: () => void;
  onViewSelect?: (viewId: string) => void;
  onViewContextMenu: (e: React.MouseEvent, view: View) => void;
  editingState: EditingState | null;
  setEditingState: (state: EditingState | null) => void;
  onRenameSubmit: () => void;
  editInputRef: React.RefObject<HTMLInputElement | null>;
}

export const ViewsSection: React.FC<ViewsSectionProps> = ({
  views,
  currentView,
  isExpanded,
  onToggle,
  canCreateView,
  onCreateView,
  onViewSelect,
  onViewContextMenu,
  editingState,
  setEditingState,
  onRenameSubmit,
  editInputRef,
}) => {
  const { data: users = [] } = useActiveUsers();

  const userNameMap = useMemo(() => {
    const map = new Map<string, string>();
    users.forEach((user) => {
      if (user.name) {
        map.set(user.id, user.name);
      }
    });
    return map;
  }, [users]);

  const getOwnerDisplayName = useCallback((view: View): string => {
    if (!view.isPrivate) return '';
    if (view.ownerUserId) {
      const name = userNameMap.get(view.ownerUserId);
      if (name) return name;
    }
    return view.ownerEmail?.split('@')[0] || 'unknown';
  }, [userNameMap]);

  const handleViewClick = (viewId: string) => {
    if (onViewSelect) {
      onViewSelect(viewId);
    }
  };

  return (
    <TreeSection
      label="Views"
      count={views.length}
      isExpanded={isExpanded}
      onToggle={onToggle}
      onAdd={canCreateView ? onCreateView : undefined}
      addTitle="Create new view"
    >
      <div className="tree-items">
        {views.length === 0 ? (
          <div className="tree-item-empty">No views</div>
        ) : (
          views.map((view) => {
            const isActive = currentView?.id === view.id;
            const isEditing = editingState?.viewId === view.id;

            return (
              <div
                key={view.id}
                className={`tree-item-container ${isActive ? 'selected' : ''}`}
              >
                {isEditing ? (
                  <div className="tree-item-edit">
                    <span className="tree-item-icon">👁️</span>
                    <input
                      ref={editInputRef}
                      type="text"
                      className="tree-item-input"
                      value={editingState.name}
                      onChange={(e) => setEditingState({ ...editingState, name: e.target.value })}
                      onBlur={onRenameSubmit}
                      onKeyDown={(e) => {
                        if (e.key === 'Enter') {
                          onRenameSubmit();
                        } else if (e.key === 'Escape') {
                          setEditingState(null);
                        }
                      }}
                      autoFocus
                    />
                  </div>
                ) : (
                  <button
                    className={`tree-item ${isActive ? 'selected' : ''}`}
                    onClick={() => handleViewClick(view.id)}
                    onDoubleClick={() => {
                      if (view._links?.edit) {
                        setEditingState({ viewId: view.id, name: view.name });
                      }
                    }}
                    onContextMenu={(e) => onViewContextMenu(e, view)}
                    title={view.isPrivate ? `Private view by ${getOwnerDisplayName(view)}` : view.name}
                  >
                    <span className="tree-item-icon">{view.isPrivate ? '🔒' : '👁️'}</span>
                    <span className="tree-item-label">
                      {view.name}
                      {view.isPrivate && <span className="owner-badge"> ({getOwnerDisplayName(view)})</span>}
                      {view.isDefault && <span className="default-badge"> ⭐</span>}
                    </span>
                  </button>
                )}
              </div>
            );
          })
        )}
      </div>
    </TreeSection>
  );
};
