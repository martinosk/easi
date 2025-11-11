import React from 'react';
import { useAppStore } from '../store/appStore';

export const ViewSelector: React.FC = () => {
  const views = useAppStore((state) => state.views);
  const currentView = useAppStore((state) => state.currentView);
  const switchView = useAppStore((state) => state.switchView);
  const loadViews = useAppStore((state) => state.loadViews);

  // Reload views when component mounts or when current view changes
  React.useEffect(() => {
    loadViews();
  }, [loadViews, currentView?.id]);

  const handleViewClick = async (viewId: string) => {
    if (currentView?.id !== viewId) {
      await switchView(viewId);
    }
  };

  if (views.length === 0) {
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
            title={view.description || view.name}
          >
            <span className="view-tab-name">{view.name}</span>
            {view.isDefault && <span className="default-indicator">⭐</span>}
          </button>
        ))}
      </div>
    </div>
  );
};
