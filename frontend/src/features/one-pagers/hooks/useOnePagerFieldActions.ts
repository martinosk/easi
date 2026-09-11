import { hasLink } from '../../../utils/hateoas';
import type { BuiltInField, CustomField, FieldRef, OnePagerConfiguration, OnePagerSubjectType } from '../types';
import {
  useChangeBuiltInFieldRequirement,
  useChangeFieldRequirement,
  useExcludeBuiltInField,
  useIncludeBuiltInField,
  useReorderFields,
} from './useOnePagerMutations';

export interface PresentationActions {
  onMoveUp: (index: number) => void;
  onMoveDown: (index: number) => void;
  onToggleRequired: (field: CustomField, required: boolean) => void;
  onExcludeBuiltIn: (field: BuiltInField) => void;
  onToggleBuiltInRequired: (field: BuiltInField, required: boolean) => void;
}

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
  onRequireConfirmationNeeded: (field: CustomField) => void,
  onRequireBuiltInConfirmationNeeded: (field: BuiltInField) => void,
) {
  const reorder = useReorderFields(subjectType);
  const includeBuiltIn = useIncludeBuiltInField(subjectType);
  const excludeBuiltIn = useExcludeBuiltInField(subjectType);
  const changeRequirement = useChangeFieldRequirement(subjectType);
  const changeBuiltInRequirement = useChangeBuiltInFieldRequirement(subjectType);

  const version = configuration?.version;

  const move = (index: number, direction: -1 | 1) => {
    if (!configuration || version === undefined) return;
    const order = swapAdjacent(configuration.displayOrder, index, direction);
    if (order) reorder.mutate({ configuration, request: { order, version } });
  };

  const fieldActions: PresentationActions = {
    onMoveUp: (index) => move(index, -1),
    onMoveDown: (index) => move(index, 1),
    onToggleRequired: (field, required) => {
      if (version === undefined) return;
      if (required && hasLink(configuration, 'x-impact-preview')) {
        onRequireConfirmationNeeded(field);
        return;
      }
      changeRequirement.mutate({ field, request: { required, version } });
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
  };

  const includeField = (field: BuiltInField) => {
    if (version === undefined) return;
    includeBuiltIn.mutate({ field, request: { version } });
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
    confirmRequireField,
    isConfirmingRequired: changeRequirement.isPending,
    confirmRequireBuiltIn,
    isConfirmingBuiltInRequired: changeBuiltInRequirement.isPending,
  };
}
