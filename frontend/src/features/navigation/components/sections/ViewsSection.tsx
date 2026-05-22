import { TextInput, UnstyledButton } from '@mantine/core';
import React, { useMemo } from 'react';
import type { View } from '../../../../api/types';
import { useActiveUsers } from '../../../users/hooks/useUsers';
import type { User } from '../../../users/types';
import type { EditingState } from '../../types';
import { TreeSection } from '../TreeSection';

const buildUserNameMap = (users: User[]): Map<string, string> => {
  const map = new Map<string, string>();
  users.forEach((user) => {
    if (user.name) {
      map.set(user.id, user.name);
    }
  });
  return map;
};

const getOwnerDisplayName = (view: View, userNameMap: Map<string, string>): string => {
  if (!view.isPrivate) return '';
  if (view.ownerUserId) {
    const name = userNameMap.get(view.ownerUserId);
    if (name) return name;
  }
  return view.ownerEmail?.split('@')[0] || 'unknown';
};

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

interface ViewEditInputProps {
  editingState: EditingState;
  setEditingState: (state: EditingState | null) => void;
  onRenameSubmit: () => void;
  editInputRef: React.RefObject<HTMLInputElement | null>;
}

const ViewEditInput: React.FC<ViewEditInputProps> = ({ editingState, setEditingState, onRenameSubmit, editInputRef }) => {
  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      onRenameSubmit();
    } else if (e.key === 'Escape') {
      setEditingState(null);
    }
  };

  return (
    <div className="tree-item-edit">
      <span className="tree-item-icon">👁️</span>
      <TextInput
        ref={editInputRef}
        className="tree-item-input"
        value={editingState.name}
        onChange={(e) => setEditingState({ ...editingState, name: e.currentTarget.value })}
        onBlur={onRenameSubmit}
        onKeyDown={handleKeyDown}
        size="xs"
        autoFocus
      />
    </div>
  );
};

interface ViewButtonProps {
  view: View;
  isActive: boolean;
  ownerDisplayName: string;
  onClick: () => void;
  onDoubleClick: () => void;
  onContextMenu: (e: React.MouseEvent) => void;
}

const ViewButton: React.FC<ViewButtonProps> = ({ view, isActive, ownerDisplayName, onClick, onDoubleClick, onContextMenu }) => (
  <UnstyledButton
    type="button"
    className={`tree-item ${isActive ? 'selected' : ''}`}
    onClick={onClick}
    onDoubleClick={onDoubleClick}
    onContextMenu={onContextMenu}
    title={view.isPrivate ? `Private view by ${ownerDisplayName}` : view.name}
  >
    <span className="tree-item-icon">{view.isPrivate ? '🔒' : '👁️'}</span>
    <span className="tree-item-label">
      {view.name}
      {view.isPrivate && <span className="owner-badge"> ({ownerDisplayName})</span>}
      {view.isDefault && <span className="default-badge"> ⭐</span>}
    </span>
  </UnstyledButton>
);

interface ViewItemProps {
  view: View;
  isActive: boolean;
  isEditing: boolean;
  ownerDisplayName: string;
  editingState: EditingState | null;
  setEditingState: (state: EditingState | null) => void;
  onRenameSubmit: () => void;
  editInputRef: React.RefObject<HTMLInputElement | null>;
  onViewClick: () => void;
  onViewContextMenu: (e: React.MouseEvent) => void;
}

const ViewItem: React.FC<ViewItemProps> = ({
  view,
  isActive,
  isEditing,
  ownerDisplayName,
  editingState,
  setEditingState,
  onRenameSubmit,
  editInputRef,
  onViewClick,
  onViewContextMenu,
}) => {
  const handleDoubleClick = () => {
    if (view._links?.edit) {
      setEditingState({ viewId: view.id, name: view.name });
    }
  };

  return (
    <div className={`tree-item-container ${isActive ? 'selected' : ''}`}>
      {isEditing && editingState ? (
        <ViewEditInput
          editingState={editingState}
          setEditingState={setEditingState}
          onRenameSubmit={onRenameSubmit}
          editInputRef={editInputRef}
        />
      ) : (
        <ViewButton
          view={view}
          isActive={isActive}
          ownerDisplayName={ownerDisplayName}
          onClick={onViewClick}
          onDoubleClick={handleDoubleClick}
          onContextMenu={onViewContextMenu}
        />
      )}
    </div>
  );
};

interface ViewListProps {
  views: View[];
  currentViewId: string | undefined;
  userNameMap: Map<string, string>;
  editingState: EditingState | null;
  setEditingState: (state: EditingState | null) => void;
  onRenameSubmit: () => void;
  editInputRef: React.RefObject<HTMLInputElement | null>;
  onViewSelect?: (viewId: string) => void;
  onViewContextMenu: (e: React.MouseEvent, view: View) => void;
}

const ViewList: React.FC<ViewListProps> = ({
  views,
  currentViewId,
  userNameMap,
  editingState,
  setEditingState,
  onRenameSubmit,
  editInputRef,
  onViewSelect,
  onViewContextMenu,
}) => {
  if (views.length === 0) {
    return <div className="tree-item-empty">No views</div>;
  }

  return (
    <>
      {views.map((view) => (
        <ViewItem
          key={view.id}
          view={view}
          isActive={currentViewId === view.id}
          isEditing={editingState?.viewId === view.id}
          ownerDisplayName={getOwnerDisplayName(view, userNameMap)}
          editingState={editingState}
          setEditingState={setEditingState}
          onRenameSubmit={onRenameSubmit}
          editInputRef={editInputRef}
          onViewClick={() => onViewSelect?.(view.id)}
          onViewContextMenu={(e) => onViewContextMenu(e, view)}
        />
      ))}
    </>
  );
};

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
  const userNameMap = useMemo(() => buildUserNameMap(users), [users]);

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
        <ViewList
          views={views}
          currentViewId={currentView?.id}
          userNameMap={userNameMap}
          editingState={editingState}
          setEditingState={setEditingState}
          onRenameSubmit={onRenameSubmit}
          editInputRef={editInputRef}
          onViewSelect={onViewSelect}
          onViewContextMenu={onViewContextMenu}
        />
      </div>
    </TreeSection>
  );
};
