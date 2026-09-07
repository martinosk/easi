import { realizationRoleQueryKeys, timeAssessmentQueryKeys } from './queryKeys';

function timeAssessmentEffects() {
  return [timeAssessmentQueryKeys.all];
}

export const timeAssessmentMutationEffects = {
  assess: timeAssessmentEffects,
  remove: timeAssessmentEffects,
};

function realizationRoleEffects() {
  return [realizationRoleQueryKeys.all];
}

export const realizationRoleMutationEffects = {
  assign: realizationRoleEffects,
  clear: realizationRoleEffects,
};
