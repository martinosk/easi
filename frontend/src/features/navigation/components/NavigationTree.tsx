import React, { useState, useEffect, useRef, useMemo, useCallback } from 'react';
import { useAppStore } from '../../../store/appStore';
import { useCurrentView } from '../../../hooks/useCurrentView';
import { useMaturityColorScale } from '../../../hooks/useMaturityColorScale';
import { deriveMaturityValue } from '../../../constants/maturityColors';
import type { View, Component, Capability, ViewId, ComponentId, ViewCapability } from '../../../api/types';
import { ContextMenu } from '../../../components/shared/ContextMenu';
import type { ContextMenuItem } from '../../../components/shared/ContextMenu';
import { ConfirmationDialog } from '../../../components/shared/ConfirmationDialog';
import { DeleteCapabilityDialog } from '../../capabilities/components/DeleteCapabilityDialog';
import { useCapabilities } from '../../capabilities/hooks/useCapabilities';
import { useComponents, useUpdateComponent, useDeleteComponent } from '../../components/hooks/useComponents';
import { useViews, useCreateView, useDeleteView, useRenameView, useSetDefaultView, useChangeViewVisibility } from '../../views/hooks/useViews';
import { useActiveUsers } from '../../users/hooks/useUsers';

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


const LEVEL_NUMBER_MAP: Record<string, number> = {
  L1: 1, L2: 2, L3: 3, L4: 4,
};

const getLevelNumber = (level: string): number => LEVEL_NUMBER_MAP[level] ?? 1;

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

const hasCustomColor = (
  colorScheme: string | undefined,
  customColor: string | undefined | null
): boolean =>
  colorScheme === 'custom' &&
  customColor !== undefined &&
  customColor !== null &&
  customColor !== '';

interface ColorIndicatorProps {
  customColor: string | undefined;
}

const ColorIndicator: React.FC<ColorIndicatorProps> = ({ customColor }) => (
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
);

interface ExpandButtonProps {
  hasChildren: boolean;
  isExpanded: boolean;
  onClick: (e: React.MouseEvent) => void;
}

const ExpandButton: React.FC<ExpandButtonProps> = ({ hasChildren, isExpanded, onClick }) => {
  if (!hasChildren) {
    return <span className="capability-expand-placeholder" />;
  }
  return (
    <button className="capability-expand-btn" onClick={onClick}>
      {isExpanded ? '\u25BC' : '\u25B6'}
    </button>
  );
};

interface NavigationTreeProps {
  onComponentSelect?: (componentId: string) => void;
  onViewSelect?: (viewId: string) => void;
  onAddComponent?: () => void;
  onCapabilitySelect?: (capabilityId: string) => void;
  onAddCapability?: () => void;
  onEditCapability?: (capability: Capability) => void;
  onEditComponent?: (componentId: string) => void;
  canCreateView?: boolean;
}

interface ViewContextMenuState {
  x: number;
  y: number;
  view: View;
}

interface ComponentContextMenuState {
  x: number;
  y: number;
  component: Component;
}

interface CapabilityContextMenuState {
  x: number;
  y: number;
  capability: Capability;
}

interface EditingState {
  viewId?: ViewId;
  componentId?: ComponentId;
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
  canCreateView = true,
}) => {
  const { data: components = [] } = useComponents();
  const { currentView } = useCurrentView();
  const selectedNodeId = useAppStore((state) => state.selectedNodeId);
  const { data: capabilities = [] } = useCapabilities();
  const { data: views = [] } = useViews();
  const { data: users = [] } = useActiveUsers();
  const { getColorForValue, getSectionNameForValue } = useMaturityColorScale();

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

  const [isOpen, setIsOpen] = useState(() => getPersistedBoolean('navigationTreeOpen', true));
  const [isModelsExpanded, setIsModelsExpanded] = useState(() => getPersistedBoolean('navigationTreeModelsExpanded', true));
  const [isViewsExpanded, setIsViewsExpanded] = useState(() => getPersistedBoolean('navigationTreeViewsExpanded', true));
  const [isCapabilitiesExpanded, setIsCapabilitiesExpanded] = useState(() => getPersistedBoolean('navigationTreeCapabilitiesExpanded', true));
  const [expandedCapabilities, setExpandedCapabilities] = useState<Set<string>>(() => getPersistedSet('navigationTreeExpandedCapabilities'));

  const [selectedCapabilityId, setSelectedCapabilityId] = useState<string | null>(null);

  const updateComponentMutation = useUpdateComponent();
  const deleteComponentMutation = useDeleteComponent();
  const createViewMutation = useCreateView();
  const deleteViewMutation = useDeleteView();
  const renameViewMutation = useRenameView();
  const setDefaultViewMutation = useSetDefaultView();
  const changeVisibilityMutation = useChangeViewVisibility();

  const [viewContextMenu, setViewContextMenu] = useState<ViewContextMenuState | null>(null);
  const [componentContextMenu, setComponentContextMenu] = useState<ComponentContextMenuState | null>(null);
  const [capabilityContextMenu, setCapabilityContextMenu] = useState<CapabilityContextMenuState | null>(null);
  const [editingState, setEditingState] = useState<EditingState | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<
    | { type: 'view'; view: View }
    | { type: 'component'; component: Component }
    | null
  >(null);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [createViewName, setCreateViewName] = useState('');
  const [isDeleting, setIsDeleting] = useState(false);
  const editInputRef = useRef<HTMLInputElement>(null);

  const [deleteCapability, setDeleteCapability] = useState<Capability | null>(null);
  const [applicationSearch, setApplicationSearch] = useState('');

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

  const capabilityTree = useMemo(() => buildCapabilityTree(capabilities), [capabilities]);

  const filteredComponents = useMemo(() => {
    if (!applicationSearch.trim()) {
      return components;
    }
    const searchLower = applicationSearch.toLowerCase();
    return components.filter(
      (c) =>
        c.name.toLowerCase().includes(searchLower) ||
        (c.description && c.description.toLowerCase().includes(searchLower))
    );
  }, [components, applicationSearch]);

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
    const items: ContextMenuItem[] = [];
    const canEdit = menu.capability._links?.edit !== undefined;
    const canDelete = menu.capability._links?.delete !== undefined;

    if (canEdit && onEditCapability) {
      items.push({
        label: 'Edit',
        onClick: () => onEditCapability(menu.capability),
      });
    }

    if (canDelete) {
      items.push({
        label: 'Delete from Model',
        onClick: () => setDeleteCapability(menu.capability),
        isDanger: true,
        ariaLabel: 'Delete capability from model',
      });
    }

    return items;
  };

  const getCapabilityNodeData = (node: CapabilityTreeNode) => {
    const { capability } = node;
    const viewCapability = currentView?.capabilities.find((vc: ViewCapability) => vc.capabilityId === capability.id);
    const isOnCanvas = !!viewCapability;
    const customColor = viewCapability?.customColor;
    const colorScheme = currentView?.colorScheme ?? 'maturity';

    return {
      hasChildNodes: node.children.length > 0,
      isExpanded: expandedCapabilities.has(capability.id),
      levelNum: getLevelNumber(capability.level),
      isSelected: selectedCapabilityId === capability.id,
      isOnCanvas,
      showColorIndicator: hasCustomColor(currentView?.colorScheme, customColor),
      title: isOnCanvas ? (capability.description || capability.name) : `${capability.description || capability.name} (not in view)`,
      customColor,
      colorScheme,
    };
  };

  const buildCapabilityItemClassName = (levelNum: number, isSelected: boolean, isOnCanvas: boolean): string => {
    return [
      'capability-tree-item',
      `capability-level-${levelNum}`,
      isSelected && 'selected',
      !isOnCanvas && 'not-in-view',
    ].filter(Boolean).join(' ');
  };

  const renderCapabilityNode = (node: CapabilityTreeNode): React.ReactNode => {
    const { capability, children } = node;
    const nodeData = getCapabilityNodeData(node);

    const effectiveMaturityValue = capability.maturityValue ?? deriveMaturityValue(capability.maturityLevel);
    const maturityColor = nodeData.colorScheme === 'classic' ? '#f9c268' : getColorForValue(effectiveMaturityValue);
    const sectionName = capability.maturitySection?.name || getSectionNameForValue(effectiveMaturityValue);
    const maturityTooltip = `${sectionName} (${effectiveMaturityValue})`;

    const handleExpandClick = (e: React.MouseEvent) => {
      e.stopPropagation();
      toggleCapabilityExpanded(capability.id);
    };

    const handleDragStart = (e: React.DragEvent) => {
      e.dataTransfer.setData('capabilityId', capability.id);
      e.dataTransfer.effectAllowed = 'copy';
    };

    return (
      <div key={capability.id}>
        <div
          className={buildCapabilityItemClassName(nodeData.levelNum, nodeData.isSelected, nodeData.isOnCanvas)}
          draggable
          onDragStart={handleDragStart}
          onClick={() => handleCapabilityClick(capability.id)}
          onContextMenu={(e) => handleCapabilityContextMenu(e, capability)}
          title={nodeData.title}
        >
          <ExpandButton
            hasChildren={nodeData.hasChildNodes}
            isExpanded={nodeData.isExpanded}
            onClick={handleExpandClick}
          />
          <span className="capability-level-badge">{capability.level}:</span>
          <span className="capability-name">{capability.name}</span>
          <span
            className="capability-maturity-indicator"
            style={{ backgroundColor: maturityColor }}
            title={maturityTooltip}
          />
          {nodeData.showColorIndicator && <ColorIndicator customColor={nodeData.customColor} />}
        </div>
        {nodeData.hasChildNodes && nodeData.isExpanded && (
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
    setViewContextMenu({ ...pos, view });
  };

  const handleComponentContextMenu = (e: React.MouseEvent, component: Component) => {
    const pos = getContextMenuPosition(e);
    setComponentContextMenu({ ...pos, component });
  };

  const handleRenameSubmit = async () => {
    if (!editingState || !editingState.name.trim()) {
      setEditingState(null);
      return;
    }

    try {
      if (editingState.viewId) {
        await renameViewMutation.mutateAsync({
          viewId: editingState.viewId,
          request: { name: editingState.name }
        });
      } else if (editingState.componentId) {
        const component = components.find(c => c.id === editingState.componentId);
        if (component) {
          await updateComponentMutation.mutateAsync({
            component,
            request: {
              name: editingState.name,
              description: component.description,
            },
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
      await createViewMutation.mutateAsync({ name: createViewName, description: '' });
      setShowCreateDialog(false);
      setCreateViewName('');
    } catch (error) {
      console.error('Failed to create view:', error);
    }
  };

  const handleDeleteConfirm = async () => {
    if (!deleteTarget) return;

    setIsDeleting(true);
    try {
      if (deleteTarget.type === 'view') {
        await deleteViewMutation.mutateAsync(deleteTarget.view);
      } else if (deleteTarget.type === 'component') {
        await deleteComponentMutation.mutateAsync(deleteTarget.component);
      }
      setDeleteTarget(null);
    } catch (error) {
      console.error('Failed to delete:', error);
    } finally {
      setIsDeleting(false);
    }
  };

  const getViewContextMenuItems = (menu: ViewContextMenuState): ContextMenuItem[] => {
    const { view } = menu;
    const items: ContextMenuItem[] = [];
    const canEdit = view._links?.edit !== undefined;
    const canDelete = view._links?.delete !== undefined;
    const canChangeVisibility = view._links?.['x-change-visibility'] !== undefined;

    if (canEdit) {
      items.push({
        label: 'Rename View',
        onClick: () => {
          setEditingState({
            viewId: view.id,
            name: view.name,
          });
        },
      });
    }

    if (canChangeVisibility) {
      items.push({
        label: view.isPrivate ? 'Make Public' : 'Make Private',
        onClick: async () => {
          try {
            await changeVisibilityMutation.mutateAsync({
              viewId: view.id,
              isPrivate: !view.isPrivate,
            });
          } catch (error) {
            console.error('Failed to change visibility:', error);
          }
        },
      });
    }

    if (!view.isDefault) {
      items.push({
        label: 'Set as Default',
        onClick: async () => {
          try {
            await setDefaultViewMutation.mutateAsync(view.id);
          } catch (error) {
            console.error('Failed to set default view:', error);
          }
        },
      });

      if (canDelete) {
        items.push({
          label: 'Delete View',
          onClick: () => {
            setDeleteTarget({
              type: 'view',
              view,
            });
          },
          isDanger: true,
        });
      }
    }

    return items;
  };

  const getComponentContextMenuItems = (menu: ComponentContextMenuState): ContextMenuItem[] => {
    const { component } = menu;
    const items: ContextMenuItem[] = [];
    const canEdit = component._links?.edit !== undefined;
    const canDelete = component._links?.delete !== undefined;

    if (canEdit && onEditComponent) {
      items.push({
        label: 'Edit',
        onClick: () => onEditComponent(component.id),
      });
    }

    if (canDelete) {
      items.push({
        label: 'Delete from Model',
        onClick: () => {
          setDeleteTarget({
            type: 'component',
            component,
          });
        },
        isDanger: true,
        ariaLabel: 'Delete application from model',
      });
    }

    return items;
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
                {onAddComponent && (
                  <button
                    className="add-view-btn"
                    onClick={onAddComponent}
                    title="Create new application"
                    data-testid="create-component-button"
                  >
                    +
                  </button>
                )}
              </div>

              {isModelsExpanded && (
                <>
                  <div className="tree-search">
                    <input
                      type="text"
                      className="tree-search-input"
                      placeholder="Search applications..."
                      value={applicationSearch}
                      onChange={(e) => setApplicationSearch(e.target.value)}
                    />
                    {applicationSearch && (
                      <button
                        className="tree-search-clear"
                        onClick={() => setApplicationSearch('')}
                        aria-label="Clear search"
                      >
                        ×
                      </button>
                    )}
                  </div>
                  <div className="tree-items">
                  {filteredComponents.length === 0 ? (
                    <div className="tree-item-empty">
                      {components.length === 0 ? 'No applications' : 'No matches'}
                    </div>
                  ) : (
                    filteredComponents.map((component) => {
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
                      const showColorIndicator = hasCustomColor(currentView?.colorScheme, customColor);

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
                          {showColorIndicator && <ColorIndicator customColor={customColor} />}
                        </button>
                      );
                    })
                  )}
                  </div>
                </>
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
                {canCreateView && (
                  <button
                    className="add-view-btn"
                    onClick={() => setShowCreateDialog(true)}
                    title="Create new view"
                  >
                    +
                  </button>
                )}
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
                              onDoubleClick={() => {
                                if (view._links?.edit) {
                                  setEditingState({ viewId: view.id, name: view.name });
                                }
                              }}
                              onContextMenu={(e) => handleViewContextMenu(e, view)}
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
                {onAddCapability && (
                  <button
                    className="add-view-btn"
                    onClick={onAddCapability}
                    title="Create new capability"
                    data-testid="create-capability-button"
                  >
                    +
                  </button>
                )}
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
          itemName={deleteTarget.type === 'view' ? deleteTarget.view.name : deleteTarget.component.name}
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
