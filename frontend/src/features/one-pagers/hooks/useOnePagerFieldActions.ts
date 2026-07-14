import type { RenameCustomFieldFormData } from '../../../lib/schemas/onePagerConfiguration';
import { hasLink } from '../../../utils/hateoas';
import type { FieldRowActions } from '../components/FieldRow';
import type { BuiltInField, CustomField, FieldRef, OnePagerConfiguration, OnePagerSubjectType } from '../types';
import {
  useAddSelectionOption,
  useChangeBuiltInFieldRequirement,
  useChangeFieldRequirement,
  useExcludeBuiltInField,
  useIncludeBuiltInField,
  useReactivateCustomField,
  useRenameCustomField,
  useReorderFields,
  useRetireCustomField,
  useRetireSelectionOption,
  useSetNumberFieldBounds,
} from './useOnePagerMutations';

function swapAdjacent(order: FieldRef[], index: number, direction: -1 | 1): FieldRef[] | null {
  const target = index + direction;
  if (target < 0 || target >= order.length) return null;
  const next = [...order];
  [next[index], next[target]] = [next[target], next[index]];
  return next;
}

export function useOnePagerFieldActions(
  subjectType: OnePagerSubjectType,
  configuration: OnePagerConfiguration | undefined,
  onRename: (field: CustomField) => void,
  onRequireConfirmationNeeded: (field: CustomField) => void,
  onRequireBuiltInConfirmationNeeded: (field: BuiltInField) => void,
) {
  const reorder = useReorderFields(subjectType);
  const includeBuiltIn = useIncludeBuiltInField(subjectType);
  const excludeBuiltIn = useExcludeBuiltInField(subjectType);
  const rename = useRenameCustomField(subjectType);
  const changeRequirement = useChangeFieldRequirement(subjectType);
  const changeBuiltInRequirement = useChangeBuiltInFieldRequirement(subjectType);
  const retireCustom = useRetireCustomField(subjectType);
  const reactivateCustom = useReactivateCustomField(subjectType);
  const addOption = useAddSelectionOption(subjectType);
  const retireOption = useRetireSelectionOption(subjectType);
  const setBounds = useSetNumberFieldBounds(subjectType);

  const version = configuration?.version;

  const move = (index: number, direction: -1 | 1) => {
    if (!configuration || version === undefined) return;
    const order = swapAdjacent(configuration.displayOrder, index, direction);
    if (order) reorder.mutate({ configuration, request: { order, version } });
  };

  const fieldActions: FieldRowActions = {
    onMoveUp: (index) => move(index, -1),
    onMoveDown: (index) => move(index, 1),
    onRename,
    onToggleRequired: (field, required) => {
      if (version === undefined) return;
      if (required && hasLink(configuration, 'x-impact-preview')) {
        onRequireConfirmationNeeded(field);
        return;
      }
      changeRequirement.mutate({ field, request: { required, version } });
    },
    onRetireCustom: (field) => {
      if (version === undefined) return;
      retireCustom.mutate({ field, request: { version } });
    },
    onExcludeBuiltIn: (field) => {
      if (version === undefined) return;
      excludeBuiltIn.mutate({ field, request: { version } });
    },
    onToggleBuiltInRequired: (field, required) => {
      if (version === undefined) return;
      if (required && hasLink(configuration, 'x-impact-preview')) {
        onRequireBuiltInConfirmationNeeded(field);
        return;
      }
      changeBuiltInRequirement.mutate({ field, request: { required, version } });
    },
    onAddOption: (field, label) => {
      if (version === undefined) return;
      addOption.mutate({ field, request: { label, version } });
    },
    onRetireOption: (option) => {
      if (version === undefined) return;
      retireOption.mutate({ option, request: { version } });
    },
    onSetBounds: (field, min, max) => {
      if (version === undefined) return;
      setBounds.mutate({ field, request: { min, max, version } });
    },
  };

  const includeField = (field: BuiltInField) => {
    if (version === undefined) return;
    includeBuiltIn.mutate({ field, request: { version } });
  };

  const reactivateField = (field: CustomField) => {
    if (version === undefined) return;
    reactivateCustom.mutate({ field, request: { version } });
  };

  const saveRename = (field: CustomField, data: RenameCustomFieldFormData, onSaved: () => void) => {
    if (version === undefined) return;
    rename.mutate(
      { field, request: { name: data.name, helpText: data.helpText, fieldType: field.type, version } },
      { onSuccess: onSaved },
    );
  };

  const confirmRequireField = (field: CustomField, onDone: () => void) => {
    if (version === undefined) return;
    changeRequirement.mutate({ field, request: { required: true, version } }, { onSuccess: onDone });
  };

  const confirmRequireBuiltIn = (field: BuiltInField, onDone: () => void) => {
    if (version === undefined) return;
    changeBuiltInRequirement.mutate({ field, request: { required: true, version } }, { onSuccess: onDone });
  };

  return {
    fieldActions,
    includeField,
    reactivateField,
    saveRename,
    isRenaming: rename.isPending,
    confirmRequireField,
    isConfirmingRequired: changeRequirement.isPending,
    confirmRequireBuiltIn,
    isConfirmingBuiltInRequired: changeBuiltInRequirement.isPending,
  };
}
