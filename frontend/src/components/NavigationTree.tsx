import React, { useState, useEffect, useRef } from 'react';
import { useAppStore } from '../store/appStore';
import apiClient from '../api/client';
import type { View } from '../api/types';

interface NavigationTreeProps {
  onComponentSelect?: (componentId: string) => void;
  onViewSelect?: (viewId: string) => void;
  onAddComponent?: () => void;
}

interface ContextMenuState {
  x: number;
  y: number;
  viewId: string;
  viewName: string;
  isDefault: boolean;
}

interface EditingState {
  viewId: string;
  viewName: string;
}

export const NavigationTree: React.FC<NavigationTreeProps> = ({
  onComponentSelect,
  onViewSelect,
  onAddComponent,
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

  const views = useAppStore((state) => state.views);
  const loadViews = useAppStore((state) => state.loadViews);
  const [contextMenu, setContextMenu] = useState<ContextMenuState | null>(null);
  const [editingView, setEditingView] = useState<EditingState | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<{ viewId: string; viewName: string } | null>(null);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [createViewName, setCreateViewName] = useState('');
  const contextMenuRef = useRef<HTMLDivElement>(null);
  const editInputRef = useRef<HTMLInputElement>(null);

  // Load views when component mounts or when currentView changes
  useEffect(() => {
    loadViews();
  }, [loadViews, currentView?.id]);

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

  // Close context menu on click outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (contextMenuRef.current && !contextMenuRef.current.contains(event.target as Node)) {
        setContextMenu(null);
      }
    };

    if (contextMenu) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }
  }, [contextMenu]);

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

  const handleViewContextMenu = (e: React.MouseEvent, view: View) => {
    e.preventDefault();
    e.stopPropagation();
    setContextMenu({
      x: e.clientX,
      y: e.clientY,
      viewId: view.id,
      viewName: view.name,
      isDefault: view.isDefault,
    });
  };

  const handleRenameClick = () => {
    if (contextMenu) {
      setEditingView({
        viewId: contextMenu.viewId,
        viewName: contextMenu.viewName,
      });
      setContextMenu(null);
    }
  };

  const handleDeleteClick = () => {
    if (contextMenu) {
      setDeleteTarget({
        viewId: contextMenu.viewId,
        viewName: contextMenu.viewName,
      });
      setContextMenu(null);
    }
  };

  const handleSetDefaultClick = async () => {
    if (contextMenu) {
      try {
        await apiClient.setDefaultView(contextMenu.viewId);
        await loadViews();
        setContextMenu(null);
      } catch (error) {
        console.error('Failed to set default view:', error);
        alert('Failed to set default view');
      }
    }
  };

  const handleRenameSubmit = async (viewId: string, newName: string) => {
    if (!newName.trim()) {
      setEditingView(null);
      return;
    }

    try {
      await apiClient.renameView(viewId, { name: newName });
      await loadViews();
      setEditingView(null);
    } catch (error) {
      console.error('Failed to rename view:', error);
      alert('Failed to rename view');
      setEditingView(null);
    }
  };

  const handleCreateView = async () => {
    if (!createViewName.trim()) return;

    try {
      await apiClient.createView({ name: createViewName, description: '' });
      await loadViews();
      setShowCreateDialog(false);
      setCreateViewName('');
    } catch (error) {
      console.error('Failed to create view:', error);
      alert('Failed to create view');
    }
  };

  const handleDeleteConfirm = async () => {
    if (!deleteTarget) return;

    try {
      await apiClient.deleteView(deleteTarget.viewId);
      await loadViews();
      setDeleteTarget(null);
    } catch (error) {
      console.error('Failed to delete view:', error);
      alert('Failed to delete view. Cannot delete the default view.');
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
              <div className="category-header-wrapper">
                <button
                  className="category-header"
                  onClick={() => setIsModelsExpanded(!isModelsExpanded)}
                >
                  <span className="category-icon">{isModelsExpanded ? '▼' : '▶'}</span>
                  <span className="category-label">Models</span>
                  <span className="category-count">{components.length}</span>
                </button>
                <button
                  className="add-view-btn"
                  onClick={onAddComponent}
                  title="Create new component"
                  data-testid="create-component-button"
                >
                  +
                </button>
              </div>

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
                          draggable={!isInCurrentView}
                          onDragStart={(e) => {
                            if (!isInCurrentView) {
                              e.dataTransfer.setData('componentId', component.id);
                              e.dataTransfer.effectAllowed = 'copy';
                            }
                          }}
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
              <div className="category-header-wrapper">
                <button
                  className="category-header"
                  onClick={() => setIsViewsExpanded(!isViewsExpanded)}
                >
                  <span className="category-icon">{isViewsExpanded ? '▼' : '▶'}</span>
                  <span className="category-label">Views</span>
                  <span className="category-count">{views.length}</span>
                </button>
                <button
                  className="add-view-btn"
                  onClick={() => setShowCreateDialog(true)}
                  title="Create new view"
                >
                  +
                </button>
              </div>

              {isViewsExpanded && (
                <div className="tree-items">
                  {views.length === 0 ? (
                    <div className="tree-item-empty">No views</div>
                  ) : (
                    views.map((view) => {
                      const isActive = currentView?.id === view.id;
                      const isEditing = editingView?.viewId === view.id;

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
                                value={editingView.viewName}
                                onChange={(e) => setEditingView({ ...editingView, viewName: e.target.value })}
                                onBlur={() => handleRenameSubmit(view.id, editingView.viewName)}
                                onKeyDown={(e) => {
                                  if (e.key === 'Enter') {
                                    handleRenameSubmit(view.id, editingView.viewName);
                                  } else if (e.key === 'Escape') {
                                    setEditingView(null);
                                  }
                                }}
                                autoFocus
                              />
                            </div>
                          ) : (
                            <button
                              className={`tree-item ${isActive ? 'selected' : ''}`}
                              onClick={() => handleViewClick(view.id)}
                              onDoubleClick={() => setEditingView({ viewId: view.id, viewName: view.name })}
                              onContextMenu={(e) => handleViewContextMenu(e, view)}
                              title={view.name}
                            >
                              <span className="tree-item-icon">👁️</span>
                              <span className="tree-item-label">
                                {view.name}
                                {view.isDefault && <span className="default-badge"> ⭐</span>}
                              </span>
                            </button>
                          )}
                        </div>
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

      {/* Context Menu */}
      {contextMenu && (
        <div
          ref={contextMenuRef}
          className="context-menu"
          style={{ top: contextMenu.y, left: contextMenu.x }}
        >
          <button className="context-menu-item" onClick={handleRenameClick}>
            Rename View
          </button>
          {!contextMenu.isDefault && (
            <button className="context-menu-item" onClick={handleSetDefaultClick}>
              Set as Default
            </button>
          )}
          {!contextMenu.isDefault && (
            <button className="context-menu-item danger" onClick={handleDeleteClick}>
              Delete View
            </button>
          )}
        </div>
      )}

      {/* Create View Dialog */}
      {showCreateDialog && (
        <div className="dialog-overlay" onClick={() => setShowCreateDialog(false)}>
          <div className="dialog" onClick={(e) => e.stopPropagation()}>
            <h3>Create New View</h3>
            <input
              type="text"
              placeholder="View name"
              value={createViewName}
              onChange={(e) => setCreateViewName(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') handleCreateView();
                if (e.key === 'Escape') setShowCreateDialog(false);
              }}
              autoFocus
              className="dialog-input"
            />
            <div className="dialog-actions">
              <button onClick={() => setShowCreateDialog(false)} className="btn-secondary">
                Cancel
              </button>
              <button onClick={handleCreateView} className="btn-primary">
                Create
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Dialog */}
      {deleteTarget && (
        <div className="dialog-overlay" onClick={() => setDeleteTarget(null)}>
          <div className="dialog" onClick={(e) => e.stopPropagation()}>
            <h3>Delete View</h3>
            <p>Are you sure you want to delete "{deleteTarget.viewName}"?</p>
            <p className="dialog-warning">This action cannot be undone.</p>
            <div className="dialog-actions">
              <button onClick={() => setDeleteTarget(null)} className="btn-secondary">
                Cancel
              </button>
              <button onClick={handleDeleteConfirm} className="btn-danger">
                Delete
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
};
