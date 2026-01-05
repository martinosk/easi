import React from 'react';
import { useAppStore } from '../../../store/appStore';
import { useViews } from '../hooks/useViews';
import { useCurrentView } from '../../../hooks/useCurrentView';
import type { ViewId } from '../../../api/types';

export const ViewSelector: React.FC = () => {
  const { data: views } = useViews();
  const { currentView } = useCurrentView();
  const setCurrentViewId = useAppStore((state) => state.setCurrentViewId);

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
                ? `Private view by ${view.ownerEmail || 'unknown'}`
                : view.description || view.name
            }
          >
            {view.isPrivate && <span className="private-indicator">🔒</span>}
            <span className="view-tab-name">{view.name}</span>
            {view.isDefault && <span className="default-indicator">⭐</span>}
          </button>
        ))}
      </div>
    </div>
  );
};
