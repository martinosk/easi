import React, { useMemo, useCallback } from 'react';
import { useAppStore } from '../../../store/appStore';
import { useViews } from '../hooks/useViews';
import { useCurrentView } from '../../../hooks/useCurrentView';
import { useActiveUsers } from '../../users/hooks/useUsers';
import type { ViewId, View } from '../../../api/types';

export const ViewSelector: React.FC = () => {
  const { data: views } = useViews();
  const { currentView } = useCurrentView();
  const { data: users = [] } = useActiveUsers();
  const setCurrentViewId = useAppStore((state) => state.setCurrentViewId);

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
    if (currentView?.id !== viewId) {
      setCurrentViewId(viewId as ViewId);
    }
  };

  if (!views || views.length === 0) {
    return null;
  }

  return (
    <div className="view-selector">
      <div className="view-tabs">
        {views.map((view) => (
          <button
            key={view.id}
            className={`view-tab ${currentView?.id === view.id ? 'active' : ''}`}
            onClick={() => handleViewClick(view.id)}
            title={
              view.isPrivate
                ? `Private view by ${getOwnerDisplayName(view)}`
                : view.description || view.name
            }
          >
            {view.isPrivate && <span className="private-indicator">🔒</span>}
            <span className="view-tab-name">
              {view.name}
              {view.isPrivate && ` (${getOwnerDisplayName(view)})`}
            </span>
            {view.isDefault && <span className="default-indicator">⭐</span>}
          </button>
        ))}
      </div>
    </div>
  );
};
