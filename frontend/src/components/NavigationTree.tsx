import React, { useState, useEffect } from 'react';
import { useAppStore } from '../store/appStore';
import apiClient from '../api/client';
import type { View } from '../api/types';

interface NavigationTreeProps {
  onComponentSelect?: (componentId: string) => void;
  onViewSelect?: (viewId: string) => void;
}

export const NavigationTree: React.FC<NavigationTreeProps> = ({
  onComponentSelect,
  onViewSelect,
}) => {
  const components = useAppStore((state) => state.components);
  const currentView = useAppStore((state) => state.currentView);
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);

  const [isOpen, setIsOpen] = useState(() => {
    const saved = localStorage.getItem('navigationTreeOpen');
    return saved !== null ? JSON.parse(saved) : true;
  });

  const [isModelsExpanded, setIsModelsExpanded] = useState(() => {
    const saved = localStorage.getItem('navigationTreeModelsExpanded');
    return saved !== null ? JSON.parse(saved) : true;
  });

  const [isViewsExpanded, setIsViewsExpanded] = useState(() => {
    const saved = localStorage.getItem('navigationTreeViewsExpanded');
    return saved !== null ? JSON.parse(saved) : true;
  });

  const [views, setViews] = useState<View[]>([]);

  // Load views
  useEffect(() => {
    const loadViews = async () => {
      try {
        const loadedViews = await apiClient.getViews();
        setViews(loadedViews);
      } catch (error) {
        console.error('Failed to load views:', error);
      }
    };
    loadViews();
  }, []);

  // Persist menu state
  useEffect(() => {
    localStorage.setItem('navigationTreeOpen', JSON.stringify(isOpen));
  }, [isOpen]);

  useEffect(() => {
    localStorage.setItem('navigationTreeModelsExpanded', JSON.stringify(isModelsExpanded));
  }, [isModelsExpanded]);

  useEffect(() => {
    localStorage.setItem('navigationTreeViewsExpanded', JSON.stringify(isViewsExpanded));
  }, [isViewsExpanded]);

  const handleComponentClick = (componentId: string) => {
    if (onComponentSelect) {
      onComponentSelect(componentId);
    }
  };

  const handleViewClick = (viewId: string) => {
    if (onViewSelect) {
      onViewSelect(viewId);
    }
  };

  return (
    <>
      <div className={`navigation-tree ${isOpen ? 'open' : 'closed'}`}>
        {isOpen && (
          <div className="navigation-tree-content">
            <div className="navigation-tree-header">
              <h3>Explorer</h3>
              <button
                className="tree-toggle-btn"
                onClick={() => setIsOpen(false)}
                aria-label="Close navigation"
              >
                ‹
              </button>
            </div>

            {/* Models Section */}
            <div className="tree-category">
              <button
                className="category-header"
                onClick={() => setIsModelsExpanded(!isModelsExpanded)}
              >
                <span className="category-icon">{isModelsExpanded ? '▼' : '▶'}</span>
                <span className="category-label">Models</span>
                <span className="category-count">{components.length}</span>
              </button>

              {isModelsExpanded && (
                <div className="tree-items">
                  {components.length === 0 ? (
                    <div className="tree-item-empty">No components</div>
                  ) : (
                    components.map((component) => {
                      const isInCurrentView = currentView?.components.some(
                        vc => vc.componentId === component.id
                      );
                      const isSelected = selectedNodeId === component.id;

                      return (
                        <button
                          key={component.id}
                          className={`tree-item ${isSelected ? 'selected' : ''} ${!isInCurrentView ? 'not-in-view' : ''}`}
                          onClick={() => handleComponentClick(component.id)}
                          title={isInCurrentView ? component.name : `${component.name} (not in current view)`}
                        >
                          <span className="tree-item-icon">📦</span>
                          <span className="tree-item-label">{component.name}</span>
                        </button>
                      );
                    })
                  )}
                </div>
              )}
            </div>

            {/* Views Section */}
            <div className="tree-category">
              <button
                className="category-header"
                onClick={() => setIsViewsExpanded(!isViewsExpanded)}
              >
                <span className="category-icon">{isViewsExpanded ? '▼' : '▶'}</span>
                <span className="category-label">Views</span>
                <span className="category-count">{views.length}</span>
              </button>

              {isViewsExpanded && (
                <div className="tree-items">
                  {views.length === 0 ? (
                    <div className="tree-item-empty">No views</div>
                  ) : (
                    views.map((view) => {
                      const isActive = currentView?.id === view.id;

                      return (
                        <button
                          key={view.id}
                          className={`tree-item ${isActive ? 'selected' : ''}`}
                          onClick={() => handleViewClick(view.id)}
                          title={view.name}
                        >
                          <span className="tree-item-icon">👁️</span>
                          <span className="tree-item-label">{view.name}</span>
                        </button>
                      );
                    })
                  )}
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {!isOpen && (
        <button
          className="tree-toggle-btn-collapsed"
          onClick={() => setIsOpen(true)}
          aria-label="Open navigation"
        >
          ›
        </button>
      )}
    </>
  );
};
