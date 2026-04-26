import React, { useCallback, useMemo, useState } from 'react';
import type { View, ViewId } from '../../../api/types';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import { useAppStore } from '../../../store/appStore';
import { selectDirtyForView } from '../../../store/slices/dynamicModeSlice';
import { useActiveUsers } from '../../users/hooks/useUsers';
import { useCurrentView } from '../hooks/useCurrentView';
import { useViews } from '../hooks/useViews';

interface OwnerLookup {
  getOwnerDisplayName: (view: View) => string;
}

function useOwnerLookup(): OwnerLookup {
  const { data: users = [] } = useActiveUsers();
  const userNameMap = useMemo(() => {
    const map = new Map<string, string>();
    users.forEach((user) => {
      if (user.name) map.set(user.id, user.name);
    });
    return map;
  }, [users]);

  const getOwnerDisplayName = useCallback(
    (view: View): string => {
      if (!view.isPrivate) return '';
      if (view.ownerUserId) {
        const name = userNameMap.get(view.ownerUserId);
        if (name) return name;
      }
      return view.ownerEmail?.split('@')[0] || 'unknown';
    },
    [userNameMap],
  );

  return { getOwnerDisplayName };
}

function pickNextActiveView(openIds: ViewId[], closingId: ViewId): ViewId | null {
  const idx = openIds.indexOf(closingId);
  if (idx === -1) return openIds[0] ?? null;
  return openIds[idx + 1] ?? openIds[idx - 1] ?? null;
}

function useOpenViews(views: View[] | undefined, openViewIds: ViewId[]): View[] {
  return useMemo(() => {
    if (!views) return [];
    const byId = new Map(views.map((v) => [v.id, v]));
    return openViewIds.map((id) => byId.get(id)).filter((v): v is View => Boolean(v));
  }, [views, openViewIds]);
}

interface ViewTabProps {
  view: View;
  isActive: boolean;
  isOnlyTab: boolean;
  ownerLookup: OwnerLookup;
  onSelect: (id: ViewId) => void;
  onRequestClose: (view: View) => void;
}

function ViewTab({ view, isActive, isOnlyTab, ownerLookup, onSelect, onRequestClose }: ViewTabProps) {
  const isDirty = useAppStore((s) => selectDirtyForView(s, view.id));
  const ownerName = ownerLookup.getOwnerDisplayName(view);
  const title = view.isPrivate ? `Private view by ${ownerName}` : view.description || view.name;
  const handleClose = (e: React.MouseEvent) => {
    e.stopPropagation();
    onRequestClose(view);
  };

  return (
    <div className={`view-tab ${isActive ? 'active' : ''}`}>
      <button
        type="button"
        className="view-tab-body"
        onClick={() => onSelect(view.id)}
        title={title}
        aria-label={`Switch to ${view.name}`}
      >
        {view.isPrivate && <span className="private-indicator">🔒</span>}
        <span className="view-tab-name">
          {view.name}
          {view.isPrivate && ` (${ownerName})`}
        </span>
        {view.isDefault && <span className="default-indicator">⭐</span>}
        {isDirty && <span className="view-tab-dirty" role="img" aria-label="Unsaved changes">●</span>}
      </button>
      <button
        type="button"
        className="view-tab-close"
        onClick={handleClose}
        disabled={isOnlyTab}
        aria-label={`Close ${view.name}`}
      >
        ×
      </button>
    </div>
  );
}

interface CloseHandlers {
  performClose: (view: View) => void;
  requestClose: (view: View) => void;
}

function useCloseHandlers(currentView: View | null, openViewIds: ViewId[], setPendingClose: (v: View | null) => void): CloseHandlers {
  const setCurrentViewId = useAppStore((s) => s.setCurrentViewId);
  const closeView = useAppStore((s) => s.closeView);
  const discardDraftForView = useAppStore((s) => s.discardDraftForView);

  const performClose = useCallback(
    (view: View) => {
      const next = pickNextActiveView(openViewIds, view.id);
      closeView(view.id);
      discardDraftForView(view.id);
      if (currentView?.id === view.id && next) setCurrentViewId(next);
    },
    [openViewIds, closeView, discardDraftForView, currentView, setCurrentViewId],
  );

  const requestClose = useCallback(
    (view: View) => {
      const dirty = selectDirtyForView(useAppStore.getState(), view.id);
      if (dirty) {
        setPendingClose(view);
        return;
      }
      performClose(view);
    },
    [performClose, setPendingClose],
  );

  return { performClose, requestClose };
}

export const ViewSelector: React.FC = () => {
  const { data: views } = useViews();
  const { currentView } = useCurrentView();
  const ownerLookup = useOwnerLookup();
  const openViewIds = useAppStore((s) => s.openViewIds);
  const setCurrentViewId = useAppStore((s) => s.setCurrentViewId);
  const [pendingClose, setPendingClose] = useState<View | null>(null);
  const openViews = useOpenViews(views, openViewIds);
  const { performClose, requestClose } = useCloseHandlers(currentView, openViewIds, setPendingClose);

  const handleSelect = useCallback(
    (id: ViewId) => {
      if (currentView?.id !== id) setCurrentViewId(id);
    },
    [currentView, setCurrentViewId],
  );

  const handleConfirm = () => {
    if (pendingClose) performClose(pendingClose);
    setPendingClose(null);
  };

  if (!views || openViews.length === 0) return null;

  return (
    <div className="view-selector">
      <div className="view-tabs">
        {openViews.map((view) => (
          <ViewTab
            key={view.id}
            view={view}
            isActive={currentView?.id === view.id}
            isOnlyTab={openViews.length === 1}
            ownerLookup={ownerLookup}
            onSelect={handleSelect}
            onRequestClose={requestClose}
          />
        ))}
      </div>
      {pendingClose && (
        <ConfirmationDialog
          title="Discard changes?"
          message="Closing this view will discard your unsaved changes."
          confirmText="Discard & close"
          cancelText="Keep editing"
          onConfirm={handleConfirm}
          onCancel={() => setPendingClose(null)}
        />
      )}
    </div>
  );
};
