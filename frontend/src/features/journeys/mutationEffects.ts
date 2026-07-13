import { journeyQueryKeys } from './queryKeys';

function journeyEffects() {
  return [journeyQueryKeys.all];
}

export const journeyMutationEffects = {
  capture: journeyEffects,
  start: journeyEffects,
  complete: journeyEffects,
  abandon: journeyEffects,
  updateDetails: journeyEffects,
  updateProgress: journeyEffects,
  changeSourceApplications: journeyEffects,
  addMilestone: journeyEffects,
  updateMilestone: journeyEffects,
  removeMilestone: journeyEffects,
};
