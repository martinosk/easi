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
import type { CapabilityId, DependencyId, RealizationId, ComponentId } from '../types/storeTypes';
import apiClient from '../../api/client';
import { handleApiCall } from '../utils/apiHelpers';
import toast from 'react-hot-toast';
const updateCapabilityInList = (
  capabilities: Capability[],
  id: CapabilityId,
  updated: Capability
): Capability[] => capabilities.map((c) => (c.id === id ? updated : c));

const mergeRealizations = (
  existing: CapabilityRealization[],
  incoming: CapabilityRealization[]
): CapabilityRealization[] => {
  const incomingIds = new Set(incoming.map((r) => r.id));
  const filtered = existing.filter((r) => !incomingIds.has(r.id));
  return [...filtered, ...incoming];
};

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
  deleteCapability: (id: CapabilityId) => Promise<void>;
  changeCapabilityParent: (id: CapabilityId, parentId: CapabilityId | null) => Promise<void>;
  createCapabilityDependency: (data: CreateCapabilityDependencyRequest) => Promise<CapabilityDependency>;
  deleteCapabilityDependency: (id: DependencyId) => Promise<void>;
  linkSystemToCapability: (capabilityId: CapabilityId, data: LinkSystemToCapabilityRequest) => Promise<CapabilityRealization>;
  updateRealization: (id: RealizationId, data: UpdateRealizationRequest) => Promise<CapabilityRealization>;
  deleteRealization: (id: RealizationId) => Promise<void>;
  loadRealizationsByCapability: (capabilityId: CapabilityId) => Promise<CapabilityRealization[]>;
  loadRealizationsByComponent: (componentId: ComponentId) => Promise<CapabilityRealization[]>;
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
    const updatedCapability = await handleApiCall(
      () => apiClient.updateCapability(id, data),
      'Failed to update capability'
    );
    set({ capabilities: updateCapabilityInList(get().capabilities, id, updatedCapability) });
    toast.success(`Capability "${data.name}" updated`);
    return updatedCapability;
  },

  updateCapabilityMetadata: async (id: CapabilityId, data: UpdateCapabilityMetadataRequest) => {
    const updatedCapability = await handleApiCall(
      () => apiClient.updateCapabilityMetadata(id, data),
      'Failed to update capability metadata'
    );
    set({ capabilities: updateCapabilityInList(get().capabilities, id, updatedCapability) });
    toast.success('Capability metadata updated');
    return updatedCapability;
  },

  addCapabilityExpert: async (id: CapabilityId, data: AddCapabilityExpertRequest) => {
    await handleApiCall(() => apiClient.addCapabilityExpert(id, data), 'Failed to add expert');
    const updatedCapability = await apiClient.getCapabilityById(id);
    set({ capabilities: updateCapabilityInList(get().capabilities, id, updatedCapability) });
    toast.success(`Expert "${data.expertName}" added`);
  },

  addCapabilityTag: async (id: CapabilityId, tag: string) => {
    await handleApiCall(() => apiClient.addCapabilityTag(id, { tag }), 'Failed to add tag');
    const updatedCapability = await apiClient.getCapabilityById(id);
    set({ capabilities: updateCapabilityInList(get().capabilities, id, updatedCapability) });
    toast.success(`Tag "${tag}" added`);
  },

  deleteCapability: async (id: CapabilityId) => {
    await handleApiCall(() => apiClient.deleteCapability(id), 'Failed to delete capability');
    set({ capabilities: get().capabilities.filter((c) => c.id !== id) });
    toast.success('Capability deleted');
  },

  changeCapabilityParent: async (id: CapabilityId, parentId: CapabilityId | null) => {
    await handleApiCall(
      () => apiClient.changeCapabilityParent(id, parentId),
      'Failed to change capability parent'
    );
    const capabilities = await apiClient.getCapabilities();
    set({ capabilities });
    toast.success(parentId ? 'Parent relationship created' : 'Parent relationship removed');
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
    await handleApiCall(() => apiClient.deleteCapabilityDependency(id), 'Failed to delete dependency');
    set({ capabilityDependencies: get().capabilityDependencies.filter((d) => d.id !== id) });
    toast.success('Dependency deleted');
  },

  linkSystemToCapability: async (capabilityId: CapabilityId, data: LinkSystemToCapabilityRequest) => {
    const newRealization = await handleApiCall(
      () => apiClient.linkSystemToCapability(capabilityId, data),
      'Failed to link system to capability'
    );

    const allRealizations = await apiClient.getCapabilitiesByComponent(data.componentId);
    set({ capabilityRealizations: mergeRealizations(get().capabilityRealizations, allRealizations) });
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
    await handleApiCall(() => apiClient.deleteRealization(id), 'Failed to delete realization');
    set({
      capabilityRealizations: get().capabilityRealizations.filter(
        (r) => r.id !== id && r.sourceRealizationId !== id
      ),
    });
    toast.success('Realization deleted');
  },

  loadRealizationsByCapability: async (capabilityId: CapabilityId) => {
    const realizations = await handleApiCall(
      () => apiClient.getSystemsByCapability(capabilityId),
      'Failed to load realizations'
    );
    set({ capabilityRealizations: mergeRealizations(get().capabilityRealizations, realizations) });
    return realizations;
  },

  loadRealizationsByComponent: async (componentId: ComponentId) => {
    const realizations = await handleApiCall(
      () => apiClient.getCapabilitiesByComponent(componentId),
      'Failed to load realizations'
    );
    set({ capabilityRealizations: mergeRealizations(get().capabilityRealizations, realizations) });
    return realizations;
  },
});
