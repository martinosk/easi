import type { Component, Relation, Capability } from '../../api/types';

export type DialogId =
  | 'create-component'
  | 'edit-component'
  | 'create-relation'
  | 'edit-relation'
  | 'create-capability'
  | 'edit-capability'
  | 'release-notes-browser';

export interface DialogDataMap {
  'create-component': undefined;
  'edit-component': { component: Component };
  'create-relation': { sourceComponentId?: string; targetComponentId?: string };
  'edit-relation': { relation: Relation };
  'create-capability': undefined;
  'edit-capability': { capability: Capability };
  'release-notes-browser': undefined;
}

export type DialogState<T extends DialogId = DialogId> = {
  id: T;
  data: DialogDataMap[T];
};

export interface DialogContextValue {
  openDialogs: Map<DialogId, DialogDataMap[DialogId]>;
  openDialog: <T extends DialogId>(id: T, data?: DialogDataMap[T]) => void;
  closeDialog: (id: DialogId) => void;
  isOpen: (id: DialogId) => boolean;
  getData: <T extends DialogId>(id: T) => DialogDataMap[T] | undefined;
}
