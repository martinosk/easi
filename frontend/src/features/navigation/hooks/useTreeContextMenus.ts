import { useState, useRef } from 'react';
import type { View, Component, Capability } from '../../../api/types';
import type { ViewContextMenuState, ComponentContextMenuState, CapabilityContextMenuState, EditingState } from '../types';
import type { ContextMenuItem } from '../../../components/shared/ContextMenu';
import { useUpdateComponent, useDeleteComponent } from '../../components/hooks/useComponents';
import { useCreateView, useDeleteView, useRenameView, useSetDefaultView, useChangeViewVisibility } from '../../views/hooks/useViews';
import { getContextMenuPosition } from '../utils/treeUtils';

interface UseTreeContextMenusProps {
  components: Component[];
  onEditCapability?: (capability: Capability) => void;
  onEditComponent?: (componentId: string) => void;
}

export function useTreeContextMenus({ components, onEditCapability, onEditComponent }: UseTreeContextMenusProps) {
  const [viewContextMenu, setViewContextMenu] = useState<ViewContextMenuState | null>(null);
  const [componentContextMenu, setComponentContextMenu] = useState<ComponentContextMenuState | null>(null);
  const [capabilityContextMenu, setCapabilityContextMenu] = useState<CapabilityContextMenuState | null>(null);
  const [editingState, setEditingState] = useState<EditingState | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<{ type: 'view'; view: View } | { type: 'component'; component: Component } | null>(null);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [createViewName, setCreateViewName] = useState('');
  const [isDeleting, setIsDeleting] = useState(false);
  const [deleteCapability, setDeleteCapability] = useState<Capability | null>(null);
  const editInputRef = useRef<HTMLInputElement>(null);

  const updateComponentMutation = useUpdateComponent();
  const deleteComponentMutation = useDeleteComponent();
  const createViewMutation = useCreateView();
  const deleteViewMutation = useDeleteView();
  const renameViewMutation = useRenameView();
  const setDefaultViewMutation = useSetDefaultView();
  const changeVisibilityMutation = useChangeViewVisibility();

  const handleViewContextMenu = (e: React.MouseEvent, view: View) => {
    const pos = getContextMenuPosition(e);
    setViewContextMenu({ ...pos, view });
  };

  const handleComponentContextMenu = (e: React.MouseEvent, component: Component) => {
    const pos = getContextMenuPosition(e);
    setComponentContextMenu({ ...pos, component });
  };

  const handleCapabilityContextMenu = (e: React.MouseEvent, capability: Capability) => {
    const pos = getContextMenuPosition(e);
    setCapabilityContextMenu({ ...pos, capability });
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
          setEditingState({ viewId: view.id, name: view.name });
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
            setDeleteTarget({ type: 'view', view });
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
          setDeleteTarget({ type: 'component', component });
        },
        isDanger: true,
        ariaLabel: 'Delete application from model',
      });
    }

    return items;
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

  return {
    viewContextMenu,
    setViewContextMenu,
    componentContextMenu,
    setComponentContextMenu,
    capabilityContextMenu,
    setCapabilityContextMenu,
    editingState,
    setEditingState,
    deleteTarget,
    setDeleteTarget,
    showCreateDialog,
    setShowCreateDialog,
    createViewName,
    setCreateViewName,
    isDeleting,
    deleteCapability,
    setDeleteCapability,
    editInputRef,
    handleViewContextMenu,
    handleComponentContextMenu,
    handleCapabilityContextMenu,
    handleRenameSubmit,
    handleCreateView,
    handleDeleteConfirm,
    getViewContextMenuItems,
    getComponentContextMenuItems,
    getCapabilityContextMenuItems,
  };
}
