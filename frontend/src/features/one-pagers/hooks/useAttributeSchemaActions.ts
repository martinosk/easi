import type { SubjectAttribute, SubjectAttributeOption, SubjectAttributeSchema } from '../../../api/types';
import type { RenameCustomFieldFormData } from '../../../lib/schemas/onePagerConfiguration';
import type { OnePagerSubjectType } from '../types';
import {
  useAddAttributeOption,
  useReactivateAttribute,
  useRenameAttribute,
  useRetireAttribute,
  useRetireAttributeOption,
  useSetAttributeBounds,
} from './useAttributeSchemaMutations';

export interface AttributeSchemaActions {
  onRename: (attribute: SubjectAttribute) => void;
  onRetire: (attribute: SubjectAttribute) => void;
  onAddOption: (attribute: SubjectAttribute, label: string) => void;
  onRetireOption: (option: SubjectAttributeOption) => void;
  onSetBounds: (attribute: SubjectAttribute, min: number | undefined, max: number | undefined) => void;
}

export function useAttributeSchemaActions(
  subjectType: OnePagerSubjectType,
  schema: SubjectAttributeSchema | undefined,
  onRename: (attribute: SubjectAttribute) => void,
) {
  const rename = useRenameAttribute(subjectType);
  const retire = useRetireAttribute(subjectType);
  const reactivate = useReactivateAttribute(subjectType);
  const addOption = useAddAttributeOption(subjectType);
  const retireOption = useRetireAttributeOption(subjectType);
  const setBounds = useSetAttributeBounds(subjectType);

  const version = schema?.version;

  const actions: AttributeSchemaActions = {
    onRename,
    onRetire: (attribute) => {
      if (version === undefined) return;
      retire.mutate({ attribute, request: { version } });
    },
    onAddOption: (attribute, label) => {
      if (version === undefined) return;
      addOption.mutate({ attribute, request: { label, version } });
    },
    onRetireOption: (option) => {
      if (version === undefined) return;
      retireOption.mutate({ option, request: { version } });
    },
    onSetBounds: (attribute, min, max) => {
      if (version === undefined) return;
      setBounds.mutate({ attribute, request: { min, max, version } });
    },
  };

  const reactivateAttribute = (attribute: SubjectAttribute) => {
    if (version === undefined) return;
    reactivate.mutate({ attribute, request: { version } });
  };

  const saveRename = (attribute: SubjectAttribute, data: RenameCustomFieldFormData, onSaved: () => void) => {
    if (version === undefined) return;
    rename.mutate(
      { attribute, request: { name: data.name, helpText: data.helpText, type: attribute.type, version } },
      { onSuccess: onSaved },
    );
  };

  return { actions, reactivateAttribute, saveRename, isRenaming: rename.isPending };
}
