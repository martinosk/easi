import React from 'react';
import { CreateComponentDialog, EditComponentDialog } from '../../features/components';
import { CreateRelationDialog, EditRelationDialog } from '../../features/relations';
import { CreateCapabilityDialog, EditCapabilityDialog } from '../../features/capabilities';
import { ReleaseNotesBrowser } from '../../contexts/releases/components/ReleaseNotesBrowser';
import { useDialog } from '../../contexts/dialogs';

export const DialogManager: React.FC = () => {
  const createComponent = useDialog('create-component');
  const editComponent = useDialog('edit-component');
  const createRelation = useDialog('create-relation');
  const editRelation = useDialog('edit-relation');
  const createCapability = useDialog('create-capability');
  const editCapability = useDialog('edit-capability');
  const releaseNotesBrowser = useDialog('release-notes-browser');

  return (
    <>
      <CreateComponentDialog
        isOpen={createComponent.isOpen}
        onClose={createComponent.close}
      />

      <CreateRelationDialog
        isOpen={createRelation.isOpen}
        onClose={createRelation.close}
        sourceComponentId={createRelation.data?.sourceComponentId}
        targetComponentId={createRelation.data?.targetComponentId}
      />

      {editComponent.data && (
        <EditComponentDialog
          isOpen={editComponent.isOpen}
          onClose={editComponent.close}
          component={editComponent.data.component}
        />
      )}

      {editRelation.data && (
        <EditRelationDialog
          isOpen={editRelation.isOpen}
          onClose={editRelation.close}
          relation={editRelation.data.relation}
        />
      )}

      <CreateCapabilityDialog
        isOpen={createCapability.isOpen}
        onClose={createCapability.close}
      />

      {editCapability.data && (
        <EditCapabilityDialog
          isOpen={editCapability.isOpen}
          onClose={editCapability.close}
          capability={editCapability.data.capability}
        />
      )}

      <ReleaseNotesBrowser
        isOpen={releaseNotesBrowser.isOpen}
        onClose={releaseNotesBrowser.close}
      />
    </>
  );
};
