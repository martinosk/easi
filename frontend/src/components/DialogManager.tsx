import React from 'react';
import { CreateComponentDialog } from './CreateComponentDialog';
import { CreateRelationDialog } from './CreateRelationDialog';
import { CreateCapabilityDialog } from './CreateCapabilityDialog';
import { EditComponentDialog } from './EditComponentDialog';
import { EditRelationDialog } from './EditRelationDialog';
import { EditCapabilityDialog } from './EditCapabilityDialog';
import type { Component, Relation, Capability } from '../api/types';

interface DialogManagerProps {
  componentDialog: {
    isOpen: boolean;
    onClose: () => void;
  };
  relationDialog: {
    isOpen: boolean;
    onClose: () => void;
    sourceComponentId?: string;
    targetComponentId?: string;
  };
  editComponentDialog: {
    isOpen: boolean;
    onClose: () => void;
    component: Component | null;
  };
  editRelationDialog: {
    isOpen: boolean;
    onClose: () => void;
    relation: Relation | null;
  };
  capabilityDialog: {
    isOpen: boolean;
    onClose: () => void;
  };
  editCapabilityDialog: {
    isOpen: boolean;
    onClose: () => void;
    capability: Capability | null;
  };
}

export const DialogManager: React.FC<DialogManagerProps> = ({
  componentDialog,
  relationDialog,
  editComponentDialog,
  editRelationDialog,
  capabilityDialog,
  editCapabilityDialog,
}) => {
  return (
    <>
      <CreateComponentDialog
        isOpen={componentDialog.isOpen}
        onClose={componentDialog.onClose}
      />

      <CreateRelationDialog
        isOpen={relationDialog.isOpen}
        onClose={relationDialog.onClose}
        sourceComponentId={relationDialog.sourceComponentId}
        targetComponentId={relationDialog.targetComponentId}
      />

      <EditComponentDialog
        isOpen={editComponentDialog.isOpen}
        onClose={editComponentDialog.onClose}
        component={editComponentDialog.component}
      />

      <EditRelationDialog
        isOpen={editRelationDialog.isOpen}
        onClose={editRelationDialog.onClose}
        relation={editRelationDialog.relation}
      />

      <CreateCapabilityDialog
        isOpen={capabilityDialog.isOpen}
        onClose={capabilityDialog.onClose}
      />

      <EditCapabilityDialog
        isOpen={editCapabilityDialog.isOpen}
        onClose={editCapabilityDialog.onClose}
        capability={editCapabilityDialog.capability}
      />
    </>
  );
};
