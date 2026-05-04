import type { View } from '../../../api/types';

export function filterViewsByCreator(views: View[], selectedCreatorIds: string[]): View[] {
  if (selectedCreatorIds.length === 0) {
    return views;
  }
  return views.filter((view) => view.ownerUserId !== undefined && selectedCreatorIds.includes(view.ownerUserId));
}
