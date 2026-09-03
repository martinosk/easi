export type DialogId = 'create-component' | 'create-relation' | 'create-capability' | 'release-notes-browser';

export interface DialogDataMap {
  'create-component': undefined;
  'create-relation': { sourceComponentId?: string; targetComponentId?: string };
  'create-capability': undefined;
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
