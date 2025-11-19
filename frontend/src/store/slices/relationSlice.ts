import type { StateCreator } from 'zustand';
import type { Relation } from '../../api/types';
import type { RelationId, RelationData, ComponentData } from '../types/storeTypes';
import apiClient from '../../api/client';
import { handleApiCall } from '../utils/apiHelpers';
import toast from 'react-hot-toast';
import { ApiError } from '../../api/types';

export interface RelationState {
  relations: Relation[];
}

export interface RelationActions {
  createRelation: (data: RelationData) => Promise<Relation>;
  updateRelation: (id: RelationId, data: Partial<ComponentData>) => Promise<Relation>;
  deleteRelation: (id: RelationId) => Promise<void>;
}

export const createRelationSlice: StateCreator<
  RelationState & RelationActions,
  [],
  [],
  RelationState & RelationActions
> = (set, get) => ({
  relations: [],

  createRelation: async (data: RelationData) => {
    const { relations } = get();

    const newRelation = await handleApiCall(
      () => apiClient.createRelation(data),
      'Failed to create relation'
    );

    set({ relations: [...relations, newRelation] });
    toast.success('Relation created');
    return newRelation;
  },

  updateRelation: async (id: RelationId, data: Partial<ComponentData>) => {
    const { relations } = get();

    const updatedRelation = await handleApiCall(
      () => apiClient.updateRelation(id, data),
      'Failed to update relation'
    );

    set({
      relations: relations.map((r) =>
        r.id === id ? updatedRelation : r
      ),
    });

    toast.success('Relation updated');
    return updatedRelation;
  },

  deleteRelation: async (id: RelationId) => {
    const { relations } = get();

    try {
      await apiClient.deleteRelation(id);

      set({
        relations: relations.filter((r) => r.id !== id),
      });

      toast.success('Relation deleted from model');
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.message
        : 'Failed to delete relation';

      toast.error(errorMessage);
      throw error;
    }
  },
});
