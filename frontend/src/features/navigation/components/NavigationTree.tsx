import React, { useState, useEffect, useRef, useMemo, useCallback } from 'react';
import { useAppStore } from '../../../store/appStore';
import apiClient from '../../../api/client';
import type { View, Component, Capability } from '../../../api/types';
import { ContextMenu } from '../../../components/shared/ContextMenu';
import type { ContextMenuItem } from '../../../components/shared/ContextMenu';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import { DeleteCapabilityDialog } from '../../capabilities/components/DeleteCapabilityDialog';

interface CapabilityTreeNode {
  capability: Capability;
  children: CapabilityTreeNode[];
}

const getPersistedBoolean = (key: string, defaultValue: boolean): boolean => {
  const saved = localStorage.getItem(key);
  return saved !== null ? JSON.parse(saved) : defaultValue;
};

const getPersistedSet = (key: string): Set<string> => {
  const saved = localStorage.getItem(key);
  return saved ? new Set(JSON.parse(saved)) : new Set();
};

const getMaturityClass = (colorScheme: string, maturityLevel?: string): string => {
  if (colorScheme === 'archimate' || colorScheme === 'archimate-classic') {
    return 'maturity-archimate';
  }

  switch (maturityLevel?.toLowerCase()) {
    case 'genesis': return 'maturity-genesis';
    case 'custom build': return 'maturity-custom-build';
    case 'product': return 'maturity-product';
    case 'commodity': return 'maturity-commodity';
    default: return 'maturity-genesis';
  }
};

const getLevelNumber = (level: string): number => {
  switch (level) {
    case 'L1': return 1;
    case 'L2': return 2;
    case 'L3': return 3;
    case 'L4': return 4;
    default: return 1;
  }
};

const getContextMenuPosition = (e: React.MouseEvent) => {
  e.preventDefault();
  e.stopPropagation();
  return { x: e.clientX, y: e.clientY };
};

const buildCapabilityTree = (capabilities: Capability[]): CapabilityTreeNode[] => {
  const capabilityMap = new Map<string, CapabilityTreeNode>();

  capabilities.forEach((cap) => {
    capabilityMap.set(cap.id, { capability: cap, children: [] });
  });

  const roots: CapabilityTreeNode[] = [];

  capabilities.forEach((cap) => {
    const node = capabilityMap.get(cap.id)!;
    if (cap.parentId && capabilityMap.has(cap.parentId)) {
      capabilityMap.get(cap.parentId)!.children.push(node);
    } else {
      roots.push(node);
    }
  });

  roots.sort((a, b) => a.capability.name.localeCompare(b.capability.name));

  return roots;
};

interface NavigationTreeProps {
  onComponentSelect?: (componentId: string) => void;
  onViewSelect?: (viewId: string) => void;
  onAddComponent?: () => void;
  onCapabilitySelect?: (capabilityId: string) => void;
  onAddCapability?: () => void;
  onEditCapability?: (capability: Capability) => void;
  onEditComponent?: (componentId: string) => void;
}

interface ViewContextMenuState {
  x: number;
  y: number;
  viewId: string;
  viewName: string;
  isDefault: boolean;
}

interface ComponentContextMenuState {
  x: number;
  y: number;
  componentId: string;
  componentName: string;
}

interface CapabilityContextMenuState {
  x: number;
  y: number;
  capability: Capability;
}

interface EditingState {
  viewId?: string;
  componentId?: string;
  name: string;
}

export const NavigationTree: React.FC<NavigationTreeProps> = ({
  onComponentSelect,
  onViewSelect,
  onAddComponent,
  onCapabilitySelect,
  onAddCapability,
  onEditCapability,
  onEditComponent,
}) => {
  const components = useAppStore((state) => state.components);
  const currentView = useAppStore((state) => state.currentView);
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const capabilities = useAppStore((state) => state.capabilities);
  const loadCapabilities = useAppStore((state) => state.loadCapabilities);
  const canvasCapabilities = useAppStore((state) => state.canvasCapabilities);

  const [isOpen, setIsOpen] = useState(() => getPersistedBoolean('navigationTreeOpen', true));
  const [isModelsExpanded, setIsModelsExpanded] = useState(() => getPersistedBoolean('navigationTreeModelsExpanded', true));
  const [isViewsExpanded, setIsViewsExpanded] = useState(() => getPersistedBoolean('navigationTreeViewsExpanded', true));
  const [isCapabilitiesExpanded, setIsCapabilitiesExpanded] = useState(() => getPersistedBoolean('navigationTreeCapabilitiesExpanded', true));
  const [expandedCapabilities, setExpandedCapabilities] = useState<Set<string>>(() => getPersistedSet('navigationTreeExpandedCapabilities'));

  const [selectedCapabilityId, setSelectedCapabilityId] = useState<string | null>(null);

  const views = useAppStore((state) => state.views);
  const loadViews = useAppStore((state) => state.loadViews);
  const updateComponent = useAppStore((state) => state.updateComponent);
  const deleteComponent = useAppStore((state) => state.deleteComponent);

  const [viewContextMenu, setViewContextMenu] = useState<ViewContextMenuState | null>(null);
  const [componentContextMenu, setComponentContextMenu] = useState<ComponentContextMenuState | null>(null);
  const [capabilityContextMenu, setCapabilityContextMenu] = useState<CapabilityContextMenuState | null>(null);
  const [editingState, setEditingState] = useState<EditingState | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<{
    type: 'view' | 'component';
    id: string;
    name: string;
  } | null>(null);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [createViewName, setCreateViewName] = useState('');
  const [isDeleting, setIsDeleting] = useState(false);
  const editInputRef = useRef<HTMLInputElement>(null);

  const [deleteCapability, setDeleteCapability] = useState<Capability | null>(null);

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

  useEffect(() => {
    localStorage.setItem('navigationTreeCapabilitiesExpanded', JSON.stringify(isCapabilitiesExpanded));
  }, [isCapabilitiesExpanded]);

  useEffect(() => {
    localStorage.setItem('navigationTreeExpandedCapabilities', JSON.stringify([...expandedCapabilities]));
  }, [expandedCapabilities]);

  useEffect(() => {
    loadCapabilities();
  }, [loadCapabilities]);

  const capabilityTree = useMemo(() => buildCapabilityTree(capabilities), [capabilities]);

  const toggleCapabilityExpanded = useCallback((capabilityId: string) => {
    setExpandedCapabilities((prev) => {
      const next = new Set(prev);
      if (next.has(capabilityId)) {
        next.delete(capabilityId);
      } else {
        next.add(capabilityId);
      }
      return next;
    });
  }, []);

  const handleCapabilityClick = (capabilityId: string) => {
    setSelectedCapabilityId(capabilityId);
    if (onCapabilitySelect) {
      onCapabilitySelect(capabilityId);
    }
  };

  const handleCapabilityContextMenu = (e: React.MouseEvent, capability: Capability) => {
    const pos = getContextMenuPosition(e);
    setCapabilityContextMenu({ ...pos, capability });
  };

  const getCapabilityContextMenuItems = (menu: CapabilityContextMenuState): ContextMenuItem[] => {
    return [
      {
        label: 'Edit',
        onClick: () => {
          if (onEditCapability) {
            onEditCapability(menu.capability);
          }
        },
      },
      {
        label: 'Delete from Model',
        onClick: () => setDeleteCapability(menu.capability),
        isDanger: true,
        ariaLabel: 'Delete capability from model',
      },
    ];
  };

  const renderCapabilityNode = (node: CapabilityTreeNode): React.ReactNode => {
    const { capability, children } = node;
    const hasChildren = children.length > 0;
    const isExpanded = expandedCapabilities.has(capability.id);
    const levelNum = getLevelNumber(capability.level);
    const isSelected = selectedCapabilityId === capability.id;
    const isOnCanvas = canvasCapabilities.some((cc) => cc.capabilityId === capability.id);
    const colorScheme = currentView?.colorScheme || 'maturity';

    const viewCapability = currentView?.capabilities.find(vc => vc.capabilityId === capability.id);
    const customColor = viewCapability?.customColor;
    const shouldShowColorIndicator =
      currentView?.colorScheme === 'custom' &&
      customColor !== undefined &&
      customColor !== null &&
      customColor !== '';

    const baseTitle = capability.description || capability.name;
    const title = isOnCanvas ? baseTitle : `${baseTitle} (not in view)`;

    return (
      <div key={capability.id}>
        <div
          className={`capability-tree-item capability-level-${levelNum} ${isSelected ? 'selected' : ''} ${!isOnCanvas ? 'not-in-view' : ''}`}
          draggable
          onDragStart={(e) => {
            e.dataTransfer.setData('capabilityId', capability.id);
            e.dataTransfer.effectAllowed = 'copy';
          }}
          onClick={() => handleCapabilityClick(capability.id)}
          onContextMenu={(e) => handleCapabilityContextMenu(e, capability)}
          title={title}
        >
          {hasChildren ? (
            <button
              className="capability-expand-btn"
              onClick={(e) => {
                e.stopPropagation();
                toggleCapabilityExpanded(capability.id);
              }}
            >
              {isExpanded ? '▼' : '▶'}
            </button>
          ) : (
            <span className="capability-expand-placeholder" />
          )}
          <span className="capability-level-badge">{capability.level}:</span>
          <span className="capability-name">{capability.name}</span>
          <span className={`capability-maturity-indicator ${getMaturityClass(colorScheme, capability.maturityLevel)}`} title={capability.maturityLevel || 'Initial'} />
          {shouldShowColorIndicator && (
            <div
              data-testid="custom-color-indicator"
              style={{
                width: '10px',
                height: '10px',
                borderRadius: '2px',
                backgroundColor: customColor,
                display: 'inline-block',
                marginLeft: '8px',
                border: '1px solid rgba(0,0,0,0.1)',
              }}
            />
          )}
        </div>
        {hasChildren && isExpanded && (
          <div className="capability-children">
            {children.map(renderCapabilityNode)}
          </div>
        )}
      </div>
    );
  };

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
    const pos = getContextMenuPosition(e);
    setViewContextMenu({ ...pos, viewId: view.id, viewName: view.name, isDefault: view.isDefault });
  };

  const handleComponentContextMenu = (e: React.MouseEvent, component: Component) => {
    const pos = getContextMenuPosition(e);
    setComponentContextMenu({ ...pos, componentId: component.id, componentName: component.name });
  };

  const handleRenameSubmit = async () => {
    if (!editingState || !editingState.name.trim()) {
      setEditingState(null);
      return;
    }

    try {
      if (editingState.viewId) {
        await apiClient.renameView(editingState.viewId, { name: editingState.name });
        await loadViews();
      } else if (editingState.componentId) {
        const component = components.find(c => c.id === editingState.componentId);
        if (component) {
          await updateComponent(editingState.componentId, {
            name: editingState.name,
            description: component.description,
          });
        }
      }
      setEditingState(null);
    } catch (error) {
      console.error('Failed to rename:', error);
      setEditingState(null);
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

    setIsDeleting(true);
    try {
      if (deleteTarget.type === 'view') {
        await apiClient.deleteView(deleteTarget.id);
        await loadViews();
      } else if (deleteTarget.type === 'component') {
        await deleteComponent(deleteTarget.id);
      }
      setDeleteTarget(null);
    } catch (error) {
      console.error('Failed to delete:', error);
    } finally {
      setIsDeleting(false);
    }
  };

  const getViewContextMenuItems = (menu: ViewContextMenuState): ContextMenuItem[] => {
    const items: ContextMenuItem[] = [
      {
        label: 'Rename View',
        onClick: () => {
          setEditingState({
            viewId: menu.viewId,
            name: menu.viewName,
          });
        },
      },
    ];

    if (!menu.isDefault) {
      items.push({
        label: 'Set as Default',
        onClick: async () => {
          try {
            await apiClient.setDefaultView(menu.viewId);
            await loadViews();
          } catch (error) {
            console.error('Failed to set default view:', error);
          }
        },
      });

      items.push({
        label: 'Delete View',
        onClick: () => {
          setDeleteTarget({
            type: 'view',
            id: menu.viewId,
            name: menu.viewName,
          });
        },
        isDanger: true,
      });
    }

    return items;
  };

  const getComponentContextMenuItems = (menu: ComponentContextMenuState): ContextMenuItem[] => {
    return [
      {
        label: 'Edit',
        onClick: () => {
          if (onEditComponent) {
            onEditComponent(menu.componentId);
          }
        },
      },
      {
        label: 'Delete from Model',
        onClick: () => {
          setDeleteTarget({
            type: 'component',
            id: menu.componentId,
            name: menu.componentName,
          });
        },
        isDanger: true,
        ariaLabel: 'Delete application from model',
      },
    ];
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
                  <span className="category-label">Applications</span>
                  <span className="category-count">{components.length}</span>
                </button>
                <button
                  className="add-view-btn"
                  onClick={onAddComponent}
                  title="Create new application"
                  data-testid="create-component-button"
                >
                  +
                </button>
              </div>

              {isModelsExpanded && (
                <div className="tree-items">
                  {components.length === 0 ? (
                    <div className="tree-item-empty">No applications</div>
                  ) : (
                    components.map((component) => {
                      const isInCurrentView = currentView?.components.some(
                        vc => vc.componentId === component.id
                      );
                      const isSelected = selectedNodeId === component.id;
                      const isEditing = editingState?.componentId === component.id;

                      if (isEditing) {
                        return (
                          <div key={component.id} className="tree-item-edit">
                            <span className="tree-item-icon">📦</span>
                            <input
                              ref={editInputRef}
                              type="text"
                              className="tree-item-input"
                              value={editingState.name}
                              onChange={(e) => setEditingState({ ...editingState, name: e.target.value })}
                              onBlur={handleRenameSubmit}
                              onKeyDown={(e) => {
                                if (e.key === 'Enter') {
                                  handleRenameSubmit();
                                } else if (e.key === 'Escape') {
                                  setEditingState(null);
                                }
                              }}
                              autoFocus
                            />
                          </div>
                        );
                      }

                      const viewComponent = currentView?.components.find(vc => vc.componentId === component.id);
                      const customColor = viewComponent?.customColor;
                      const shouldShowColorIndicator =
                        currentView?.colorScheme === 'custom' &&
                        customColor !== undefined &&
                        customColor !== null &&
                        customColor !== '';

                      return (
                        <button
                          key={component.id}
                          className={`tree-item ${isSelected ? 'selected' : ''} ${!isInCurrentView ? 'not-in-view' : ''}`}
                          onClick={() => handleComponentClick(component.id)}
                          onContextMenu={(e) => handleComponentContextMenu(e, component)}
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
                          {shouldShowColorIndicator && (
                            <div
                              data-testid="custom-color-indicator"
                              style={{
                                width: '10px',
                                height: '10px',
                                borderRadius: '2px',
                                backgroundColor: customColor,
                                display: 'inline-block',
                                marginLeft: '8px',
                                border: '1px solid rgba(0,0,0,0.1)',
                              }}
                            />
                          )}
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
                                onBlur={handleRenameSubmit}
                                onKeyDown={(e) => {
                                  if (e.key === 'Enter') {
                                    handleRenameSubmit();
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
                              onDoubleClick={() => setEditingState({ viewId: view.id, name: view.name })}
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

            {/* Capabilities Section */}
            <div className="tree-category">
              <div className="category-header-wrapper">
                <button
                  className="category-header"
                  onClick={() => setIsCapabilitiesExpanded(!isCapabilitiesExpanded)}
                >
                  <span className="category-icon">{isCapabilitiesExpanded ? '▼' : '▶'}</span>
                  <span className="category-label">Capabilities</span>
                  <span className="category-count">{capabilities.length}</span>
                </button>
                <button
                  className="add-view-btn"
                  onClick={onAddCapability}
                  title="Create new capability"
                  data-testid="create-capability-button"
                >
                  +
                </button>
              </div>

              {isCapabilitiesExpanded && (
                <div className="tree-items">
                  {capabilityTree.length === 0 ? (
                    <div className="tree-item-empty">No capabilities</div>
                  ) : (
                    capabilityTree.map(renderCapabilityNode)
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

      {/* Context Menus */}
      {viewContextMenu && (
        <ContextMenu
          x={viewContextMenu.x}
          y={viewContextMenu.y}
          items={getViewContextMenuItems(viewContextMenu)}
          onClose={() => setViewContextMenu(null)}
        />
      )}

      {componentContextMenu && (
        <ContextMenu
          x={componentContextMenu.x}
          y={componentContextMenu.y}
          items={getComponentContextMenuItems(componentContextMenu)}
          onClose={() => setComponentContextMenu(null)}
        />
      )}

      {capabilityContextMenu && (
        <ContextMenu
          x={capabilityContextMenu.x}
          y={capabilityContextMenu.y}
          items={getCapabilityContextMenuItems(capabilityContextMenu)}
          onClose={() => setCapabilityContextMenu(null)}
        />
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
        <ConfirmationDialog
          title={deleteTarget.type === 'view' ? 'Delete View' : 'Delete Application'}
          message={
            deleteTarget.type === 'view'
              ? `Are you sure you want to delete this view?`
              : `This will delete the application from the entire model, remove it from ALL views, and delete ALL relations involving this application.`
          }
          itemName={deleteTarget.name}
          confirmText="Delete"
          cancelText="Cancel"
          onConfirm={handleDeleteConfirm}
          onCancel={() => setDeleteTarget(null)}
          isLoading={isDeleting}
        />
      )}

      <DeleteCapabilityDialog
        isOpen={deleteCapability !== null}
        onClose={() => setDeleteCapability(null)}
        capability={deleteCapability}
      />
    </>
  );
};
