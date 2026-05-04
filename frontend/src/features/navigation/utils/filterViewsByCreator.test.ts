import { describe, expect, it } from 'vitest';
import { toViewId } from '../../../api/types';
import { buildView } from '../../../test/helpers';
import { filterViewsByCreator } from './filterViewsByCreator';

describe('filterViewsByCreator', () => {
  const USER_ALICE = 'user-alice';
  const USER_BOB = 'user-bob';
  const USER_CAROL = 'user-carol';

  describe('when filter is active (selectedCreatorIds is non-empty)', () => {
    it('should return only views whose ownerUserId matches the selected creator', () => {
      const view1 = buildView({ id: toViewId('view-1'), name: 'Alice View', ownerUserId: USER_ALICE });
      const view2 = buildView({ id: toViewId('view-2'), name: 'Bob View', ownerUserId: USER_BOB });
      const view3 = buildView({ id: toViewId('view-3'), name: 'Alice View 2', ownerUserId: USER_ALICE });

      const result = filterViewsByCreator([view1, view2, view3], [USER_ALICE]);

      expect(result).toEqual([view1, view3]);
    });
  });

  describe('when filter is cleared (selectedCreatorIds is empty)', () => {
    it('should return all views unchanged', () => {
      const view1 = buildView({ id: toViewId('view-1'), name: 'Alice View', ownerUserId: USER_ALICE });
      const view2 = buildView({ id: toViewId('view-2'), name: 'Bob View', ownerUserId: USER_BOB });
      const view3 = buildView({ id: toViewId('view-3'), name: 'No Owner' });

      const result = filterViewsByCreator([view1, view2, view3], []);

      expect(result).toEqual([view1, view2, view3]);
    });
  });

  describe('when filter is set to a user with no views', () => {
    it('should return an empty array', () => {
      const view1 = buildView({ id: toViewId('view-1'), name: 'Alice View', ownerUserId: USER_ALICE });
      const view2 = buildView({ id: toViewId('view-2'), name: 'Bob View', ownerUserId: USER_BOB });

      const result = filterViewsByCreator([view1, view2], [USER_CAROL]);

      expect(result).toEqual([]);
    });
  });

  describe('when switching between users', () => {
    it('should return the correct views for the selected user', () => {
      const view1 = buildView({ id: toViewId('view-1'), name: 'Alice View', ownerUserId: USER_ALICE });
      const view2 = buildView({ id: toViewId('view-2'), name: 'Bob View', ownerUserId: USER_BOB });
      const view3 = buildView({ id: toViewId('view-3'), name: 'Carol View', ownerUserId: USER_CAROL });

      const views = [view1, view2, view3];

      const resultAlice = filterViewsByCreator(views, [USER_ALICE]);
      expect(resultAlice).toEqual([view1]);

      const resultBob = filterViewsByCreator(views, [USER_BOB]);
      expect(resultBob).toEqual([view2]);
    });
  });

  describe('when multiple users are selected', () => {
    it('should return the union of views for all selected creators', () => {
      const view1 = buildView({ id: toViewId('view-1'), name: 'Alice View', ownerUserId: USER_ALICE });
      const view2 = buildView({ id: toViewId('view-2'), name: 'Bob View', ownerUserId: USER_BOB });
      const view3 = buildView({ id: toViewId('view-3'), name: 'Carol View', ownerUserId: USER_CAROL });

      const views = [view1, view2, view3];

      const resultAliceAndCarol = filterViewsByCreator(views, [USER_ALICE, USER_CAROL]);
      expect(resultAliceAndCarol).toEqual([view1, view3]);
    });
  });

  describe('views without ownerUserId', () => {
    it('should exclude views without ownerUserId when filter is active', () => {
      const view1 = buildView({ id: toViewId('view-1'), name: 'Owned', ownerUserId: USER_ALICE });
      const view2 = buildView({ id: toViewId('view-2'), name: 'Unowned', ownerUserId: undefined });

      const result = filterViewsByCreator([view1, view2], [USER_ALICE]);

      expect(result).toEqual([view1]);
    });
  });
});
