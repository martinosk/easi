import React from 'react';
import { useDialog } from '../../contexts/dialogs';
import { ReleaseNotesBrowser } from '../../contexts/releases/components/ReleaseNotesBrowser';
import { CreateCapabilityDialog } from '../../features/capabilities';
import { CreateComponentDialog } from '../../features/components';
import { CreateRelationDialog } from '../../features/relations';

export const DialogManager: React.FC = () => {
  const createComponent = useDialog('create-component');
  const createRelation = useDialog('create-relation');
  const createCapability = useDialog('create-capability');
  const releaseNotesBrowser = useDialog('release-notes-browser');

  return (
    <>
      <CreateComponentDialog isOpen={createComponent.isOpen} onClose={createComponent.close} />

      <CreateRelationDialog
        isOpen={createRelation.isOpen}
        onClose={createRelation.close}
        sourceComponentId={createRelation.data?.sourceComponentId}
        targetComponentId={createRelation.data?.targetComponentId}
      />

      <CreateCapabilityDialog isOpen={createCapability.isOpen} onClose={createCapability.close} />

      <ReleaseNotesBrowser isOpen={releaseNotesBrowser.isOpen} onClose={releaseNotesBrowser.close} />
    </>
  );
};
