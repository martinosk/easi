import { useState, useRef, useCallback } from 'react';
import type { View, Component, Capability, AcquiredEntity, Vendor, InternalTeam, HATEOASLinks } from '../../../api/types';
import type { ViewContextMenuState, ComponentContextMenuState, CapabilityContextMenuState, EditingState } from '../types';
import type { DeleteTarget } from '../components/DeleteConfirmation';
import type { ContextMenuItem } from '../../../components/shared/ContextMenu';
import type { ArtifactType } from '../../edit-grants/types';
import { useUpdateComponent, useDeleteComponent } from '../../components/hooks/useComponents';
import { useCreateView, useDeleteView, useRenameView, useSetDefaultView, useChangeViewVisibility } from '../../views/hooks/useViews';
import { useDeleteAcquiredEntity } from '../../origin-entities/hooks/useAcquiredEntities';
import { useDeleteVendor } from '../../origin-entities/hooks/useVendors';
import { useDeleteInternalTeam } from '../../origin-entities/hooks/useInternalTeams';
import { getContextMenuPosition } from '../utils/treeUtils';
import { copyToClipboard, generateViewShareUrl } from '../../../utils/clipboard';
import { hasLink } from '../../../utils/hateoas';

export interface InviteTarget {
  id: string;
  artifactType: ArtifactType;
}

export interface OriginEntityContextMenuState {
  x: number;
  y: number;
  entity: AcquiredEntity | Vendor | InternalTeam;
  entityType: 'acquired' | 'vendor' | 'team';
}

interface EntityMenuConfig {
  links?: HATEOASLinks;
  onEdit?: () => void;
  onDelete: () => void;
  onInviteToEdit?: () => void;
  deleteLabel: string;
  deleteAriaLabel: string;
}

function buildEntityMenuItems(config: EntityMenuConfig): ContextMenuItem[] {
  const hasInvite = config.links?.['x-edit-grants'] !== undefined && config.onInviteToEdit !== undefined;
  const hasEdit = config.links?.edit !== undefined && config.onEdit !== undefined;
  const hasDelete = config.links?.delete !== undefined;

  return filterNullItems([
    createConditionalMenuItem(hasInvite, { label: 'Invite to Edit...', onClick: config.onInviteToEdit! }),
    createConditionalMenuItem(hasEdit, { label: 'Edit', onClick: config.onEdit! }),
    createConditionalMenuItem(hasDelete, {
      label: config.deleteLabel,
      onClick: config.onDelete,
      isDanger: true,
      ariaLabel: config.deleteAriaLabel,
    }),
  ]);
}

function createConditionalMenuItem(
  condition: boolean,
  item: ContextMenuItem
): ContextMenuItem | null {
  return condition ? item : null;
}

function filterNullItems(items: (ContextMenuItem | null)[]): ContextMenuItem[] {
  return items.filter((item): item is ContextMenuItem => item !== null);
}

interface UseTreeContextMenusProps {
  components: Component[];
  onEditCapability?: (capability: Capability) => void;
  onEditComponent?: (componentId: string) => void;
  onEditAcquiredEntity?: (entity: AcquiredEntity) => void;
  onEditVendor?: (vendor: Vendor) => void;
  onEditInternalTeam?: (team: InternalTeam) => void;
}

export function useTreeContextMenus({
  components,
  onEditCapability,
  onEditComponent,
  onEditAcquiredEntity,
  onEditVendor,
  onEditInternalTeam,
}: UseTreeContextMenusProps) {
  const [viewContextMenu, setViewContextMenu] = useState<ViewContextMenuState | null>(null);
  const [componentContextMenu, setComponentContextMenu] = useState<ComponentContextMenuState | null>(null);
  const [capabilityContextMenu, setCapabilityContextMenu] = useState<CapabilityContextMenuState | null>(null);
  const [originEntityContextMenu, setOriginEntityContextMenu] = useState<OriginEntityContextMenuState | null>(null);
  const [editingState, setEditingState] = useState<EditingState | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<DeleteTarget | null>(null);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [createViewName, setCreateViewName] = useState('');
  const [isDeleting, setIsDeleting] = useState(false);
  const [deleteCapability, setDeleteCapability] = useState<Capability | null>(null);
  const [inviteTarget, setInviteTarget] = useState<InviteTarget | null>(null);
  const editInputRef = useRef<HTMLInputElement>(null);

  const updateComponentMutation = useUpdateComponent();
  const deleteComponentMutation = useDeleteComponent();
  const createViewMutation = useCreateView();
  const deleteViewMutation = useDeleteView();
  const renameViewMutation = useRenameView();
  const setDefaultViewMutation = useSetDefaultView();
  const changeVisibilityMutation = useChangeViewVisibility();
  const deleteAcquiredEntityMutation = useDeleteAcquiredEntity();
  const deleteVendorMutation = useDeleteVendor();
  const deleteInternalTeamMutation = useDeleteInternalTeam();

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

  const handleAcquiredEntityContextMenu = (e: React.MouseEvent, entity: AcquiredEntity) => {
    e.preventDefault();
    const pos = getContextMenuPosition(e);
    setOriginEntityContextMenu({ ...pos, entity, entityType: 'acquired' });
  };

  const handleVendorContextMenu = (e: React.MouseEvent, vendor: Vendor) => {
    e.preventDefault();
    const pos = getContextMenuPosition(e);
    setOriginEntityContextMenu({ ...pos, entity: vendor, entityType: 'vendor' });
  };

  const handleInternalTeamContextMenu = (e: React.MouseEvent, team: InternalTeam) => {
    e.preventDefault();
    const pos = getContextMenuPosition(e);
    setOriginEntityContextMenu({ ...pos, entity: team, entityType: 'team' });
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

  const executeDelete = useCallback(async (target: DeleteTarget): Promise<void> => {
    switch (target.type) {
      case 'view':
        await deleteViewMutation.mutateAsync(target.view);
        break;
      case 'component':
        await deleteComponentMutation.mutateAsync(target.component);
        break;
      case 'acquired':
        await deleteAcquiredEntityMutation.mutateAsync({ id: target.entity.id, name: target.entity.name });
        break;
      case 'vendor':
        await deleteVendorMutation.mutateAsync({ id: target.entity.id, name: target.entity.name });
        break;
      case 'team':
        await deleteInternalTeamMutation.mutateAsync({ id: target.entity.id, name: target.entity.name });
        break;
    }
  }, [deleteViewMutation, deleteComponentMutation, deleteAcquiredEntityMutation, deleteVendorMutation, deleteInternalTeamMutation]);

  const handleDeleteConfirm = async () => {
    if (!deleteTarget) return;

    setIsDeleting(true);
    try {
      await executeDelete(deleteTarget);
      setDeleteTarget(null);
    } catch (error) {
      console.error('Failed to delete:', error);
    } finally {
      setIsDeleting(false);
    }
  };

  const getViewContextMenuItems = (menu: ViewContextMenuState): ContextMenuItem[] => {
    const { view } = menu;
    const canEdit = view._links?.edit !== undefined;
    const canDelete = view._links?.delete !== undefined && !view.isDefault;
    const canChangeVisibility = view._links?.['x-change-visibility'] !== undefined;
    const canInvite = hasLink(view, 'x-edit-grants');
    const visibilityLabel = view.isPrivate ? 'Make Public' : 'Make Private';

    const shareItem: ContextMenuItem = {
      label: 'Share (copy URL)...',
      onClick: () => copyToClipboard(generateViewShareUrl(view.id)),
    };

    return filterNullItems([
      createConditionalMenuItem(canInvite, {
        label: 'Invite to Edit...',
        onClick: () => setInviteTarget({ id: view.id, artifactType: 'view' }),
      }),
      shareItem,
      createConditionalMenuItem(canEdit, {
        label: 'Rename View',
        onClick: () => setEditingState({ viewId: view.id, name: view.name }),
      }),
      createConditionalMenuItem(canChangeVisibility, {
        label: visibilityLabel,
        onClick: () => changeVisibilityMutation.mutate({ viewId: view.id, isPrivate: !view.isPrivate }),
      }),
      createConditionalMenuItem(!view.isDefault, {
        label: 'Set as Default',
        onClick: () => setDefaultViewMutation.mutate(view.id),
      }),
      createConditionalMenuItem(canDelete, {
        label: 'Delete View',
        onClick: () => setDeleteTarget({ type: 'view', view }),
        isDanger: true,
      }),
    ]);
  };

  const getComponentContextMenuItems = (menu: ComponentContextMenuState): ContextMenuItem[] => {
    const { component } = menu;
    return buildEntityMenuItems({
      links: component._links,
      onEdit: onEditComponent ? () => onEditComponent(component.id) : undefined,
      onInviteToEdit: () => setInviteTarget({ id: component.id, artifactType: 'component' }),
      onDelete: () => setDeleteTarget({ type: 'component', component }),
      deleteLabel: 'Delete from Model',
      deleteAriaLabel: 'Delete application from model',
    });
  };

  const getCapabilityContextMenuItems = (menu: CapabilityContextMenuState): ContextMenuItem[] => {
    return buildEntityMenuItems({
      links: menu.capability._links,
      onEdit: onEditCapability ? () => onEditCapability(menu.capability) : undefined,
      onInviteToEdit: () => setInviteTarget({ id: menu.capability.id, artifactType: 'capability' }),
      onDelete: () => setDeleteCapability(menu.capability),
      deleteLabel: 'Delete from Model',
      deleteAriaLabel: 'Delete capability from model',
    });
  };

  const getOriginEntityContextMenuItems = (menu: OriginEntityContextMenuState): ContextMenuItem[] => {
    const entityTypeLabels: Record<OriginEntityContextMenuState['entityType'], string> = {
      acquired: 'acquired entity',
      vendor: 'vendor',
      team: 'internal team',
    };

    const originEntityArtifactTypes: Record<OriginEntityContextMenuState['entityType'], ArtifactType> = {
      acquired: 'acquired_entity',
      vendor: 'vendor',
      team: 'internal_team',
    };

    const editHandlers: Record<OriginEntityContextMenuState['entityType'], (() => void) | undefined> = {
      acquired: onEditAcquiredEntity ? () => onEditAcquiredEntity(menu.entity as AcquiredEntity) : undefined,
      vendor: onEditVendor ? () => onEditVendor(menu.entity as Vendor) : undefined,
      team: onEditInternalTeam ? () => onEditInternalTeam(menu.entity as InternalTeam) : undefined,
    };

    const deleteTargetFactories: Record<OriginEntityContextMenuState['entityType'], () => DeleteTarget> = {
      acquired: () => ({ type: 'acquired', entity: menu.entity as AcquiredEntity }),
      vendor: () => ({ type: 'vendor', entity: menu.entity as Vendor }),
      team: () => ({ type: 'team', entity: menu.entity as InternalTeam }),
    };

    return buildEntityMenuItems({
      links: menu.entity._links,
      onEdit: editHandlers[menu.entityType],
      onInviteToEdit: () => setInviteTarget({ id: menu.entity.id, artifactType: originEntityArtifactTypes[menu.entityType] }),
      onDelete: () => setDeleteTarget(deleteTargetFactories[menu.entityType]()),
      deleteLabel: 'Delete from Model',
      deleteAriaLabel: `Delete ${entityTypeLabels[menu.entityType]} from model`,
    });
  };

  return {
    viewContextMenu,
    setViewContextMenu,
    componentContextMenu,
    setComponentContextMenu,
    capabilityContextMenu,
    setCapabilityContextMenu,
    originEntityContextMenu,
    setOriginEntityContextMenu,
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
    inviteTarget,
    setInviteTarget,
    editInputRef,
    handleViewContextMenu,
    handleComponentContextMenu,
    handleCapabilityContextMenu,
    handleAcquiredEntityContextMenu,
    handleVendorContextMenu,
    handleInternalTeamContextMenu,
    handleRenameSubmit,
    handleCreateView,
    handleDeleteConfirm,
    getViewContextMenuItems,
    getComponentContextMenuItems,
    getCapabilityContextMenuItems,
    getOriginEntityContextMenuItems,
  };
}
