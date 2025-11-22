import type { StateCreator } from 'zustand';
import type {
  Capability,
  CapabilityDependency,
  CapabilityRealization,
  CreateCapabilityRequest,
  UpdateCapabilityRequest,
  UpdateCapabilityMetadataRequest,
  AddCapabilityExpertRequest,
  CreateCapabilityDependencyRequest,
  LinkSystemToCapabilityRequest,
  UpdateRealizationRequest,
} from '../../api/types';
import type { CapabilityId, DependencyId, RealizationId } from '../types/storeTypes';
import apiClient from '../../api/client';
import { handleApiCall } from '../utils/apiHelpers';
import toast from 'react-hot-toast';
import { ApiError } from '../../api/types';

export interface CapabilityState {
  capabilities: Capability[];
  capabilityDependencies: CapabilityDependency[];
  capabilityRealizations: CapabilityRealization[];
}

export interface CapabilityActions {
  loadCapabilities: () => Promise<void>;
  loadCapabilityDependencies: () => Promise<void>;
  createCapability: (data: CreateCapabilityRequest) => Promise<Capability>;
  updateCapability: (id: CapabilityId, data: UpdateCapabilityRequest) => Promise<Capability>;
  updateCapabilityMetadata: (id: CapabilityId, data: UpdateCapabilityMetadataRequest) => Promise<Capability>;
  addCapabilityExpert: (id: CapabilityId, data: AddCapabilityExpertRequest) => Promise<void>;
  addCapabilityTag: (id: CapabilityId, tag: string) => Promise<void>;
  createCapabilityDependency: (data: CreateCapabilityDependencyRequest) => Promise<CapabilityDependency>;
  deleteCapabilityDependency: (id: DependencyId) => Promise<void>;
  linkSystemToCapability: (capabilityId: CapabilityId, data: LinkSystemToCapabilityRequest) => Promise<CapabilityRealization>;
  updateRealization: (id: RealizationId, data: UpdateRealizationRequest) => Promise<CapabilityRealization>;
  deleteRealization: (id: RealizationId) => Promise<void>;
  loadRealizationsByCapability: (capabilityId: CapabilityId) => Promise<CapabilityRealization[]>;
  loadRealizationsByComponent: (componentId: string) => Promise<CapabilityRealization[]>;
}

export const createCapabilitySlice: StateCreator<
  CapabilityState & CapabilityActions,
  [],
  [],
  CapabilityState & CapabilityActions
> = (set, get) => ({
  capabilities: [],
  capabilityDependencies: [],
  capabilityRealizations: [],

  loadCapabilities: async () => {
    const capabilities = await handleApiCall(
      () => apiClient.getCapabilities(),
      'Failed to load capabilities'
    );
    set({ capabilities });
  },

  loadCapabilityDependencies: async () => {
    const capabilityDependencies = await handleApiCall(
      () => apiClient.getCapabilityDependencies(),
      'Failed to load capability dependencies'
    );
    set({ capabilityDependencies });
  },

  createCapability: async (data: CreateCapabilityRequest) => {
    const { capabilities } = get();

    const newCapability = await handleApiCall(
      () => apiClient.createCapability(data),
      'Failed to create capability'
    );

    set({ capabilities: [...capabilities, newCapability] });
    toast.success(`Capability "${data.name}" created`);
    return newCapability;
  },

  updateCapability: async (id: CapabilityId, data: UpdateCapabilityRequest) => {
    const { capabilities } = get();

    const updatedCapability = await handleApiCall(
      () => apiClient.updateCapability(id, data),
      'Failed to update capability'
    );

    set({
      capabilities: capabilities.map((c) =>
        c.id === id ? updatedCapability : c
      ),
    });

    toast.success(`Capability "${data.name}" updated`);
    return updatedCapability;
  },

  updateCapabilityMetadata: async (id: CapabilityId, data: UpdateCapabilityMetadataRequest) => {
    const { capabilities } = get();

    const updatedCapability = await handleApiCall(
      () => apiClient.updateCapabilityMetadata(id, data),
      'Failed to update capability metadata'
    );

    set({
      capabilities: capabilities.map((c) =>
        c.id === id ? updatedCapability : c
      ),
    });

    toast.success('Capability metadata updated');
    return updatedCapability;
  },

  addCapabilityExpert: async (id: CapabilityId, data: AddCapabilityExpertRequest) => {
    await handleApiCall(
      () => apiClient.addCapabilityExpert(id, data),
      'Failed to add expert'
    );

    const updatedCapability = await apiClient.getCapabilityById(id);
    const { capabilities } = get();
    set({
      capabilities: capabilities.map((c) =>
        c.id === id ? updatedCapability : c
      ),
    });

    toast.success(`Expert "${data.expertName}" added`);
  },

  addCapabilityTag: async (id: CapabilityId, tag: string) => {
    await handleApiCall(
      () => apiClient.addCapabilityTag(id, { tag }),
      'Failed to add tag'
    );

    const updatedCapability = await apiClient.getCapabilityById(id);
    const { capabilities } = get();
    set({
      capabilities: capabilities.map((c) =>
        c.id === id ? updatedCapability : c
      ),
    });

    toast.success(`Tag "${tag}" added`);
  },

  createCapabilityDependency: async (data: CreateCapabilityDependencyRequest) => {
    const { capabilityDependencies } = get();

    const newDependency = await handleApiCall(
      () => apiClient.createCapabilityDependency(data),
      'Failed to create dependency'
    );

    set({ capabilityDependencies: [...capabilityDependencies, newDependency] });
    toast.success('Dependency created');
    return newDependency;
  },

  deleteCapabilityDependency: async (id: DependencyId) => {
    const { capabilityDependencies } = get();

    try {
      await apiClient.deleteCapabilityDependency(id);
      set({
        capabilityDependencies: capabilityDependencies.filter((d) => d.id !== id),
      });
      toast.success('Dependency deleted');
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.message
        : 'Failed to delete dependency';
      toast.error(errorMessage);
      throw error;
    }
  },

  linkSystemToCapability: async (capabilityId: CapabilityId, data: LinkSystemToCapabilityRequest) => {
    const { capabilityRealizations } = get();

    const newRealization = await handleApiCall(
      () => apiClient.linkSystemToCapability(capabilityId, data),
      'Failed to link system to capability'
    );

    set({ capabilityRealizations: [...capabilityRealizations, newRealization] });
    toast.success('System linked to capability');
    return newRealization;
  },

  updateRealization: async (id: RealizationId, data: UpdateRealizationRequest) => {
    const { capabilityRealizations } = get();

    const updatedRealization = await handleApiCall(
      () => apiClient.updateRealization(id, data),
      'Failed to update realization'
    );

    set({
      capabilityRealizations: capabilityRealizations.map((r) =>
        r.id === id ? updatedRealization : r
      ),
    });

    toast.success('Realization updated');
    return updatedRealization;
  },

  deleteRealization: async (id: RealizationId) => {
    const { capabilityRealizations } = get();

    try {
      await apiClient.deleteRealization(id);
      set({
        capabilityRealizations: capabilityRealizations.filter((r) => r.id !== id),
      });
      toast.success('Realization deleted');
    } catch (error) {
      const errorMessage = error instanceof ApiError
        ? error.message
        : 'Failed to delete realization';
      toast.error(errorMessage);
      throw error;
    }
  },

  loadRealizationsByCapability: async (capabilityId: CapabilityId) => {
    const realizations = await handleApiCall(
      () => apiClient.getSystemsByCapability(capabilityId),
      'Failed to load realizations'
    );

    const { capabilityRealizations } = get();
    const existingIds = new Set(realizations.map((r) => r.id));
    const filtered = capabilityRealizations.filter((r) => !existingIds.has(r.id));
    set({ capabilityRealizations: [...filtered, ...realizations] });

    return realizations;
  },

  loadRealizationsByComponent: async (componentId: string) => {
    const realizations = await handleApiCall(
      () => apiClient.getCapabilitiesByComponent(componentId),
      'Failed to load realizations'
    );

    const { capabilityRealizations } = get();
    const existingIds = new Set(realizations.map((r) => r.id));
    const filtered = capabilityRealizations.filter((r) => !existingIds.has(r.id));
    set({ capabilityRealizations: [...filtered, ...realizations] });

    return realizations;
  },
});
