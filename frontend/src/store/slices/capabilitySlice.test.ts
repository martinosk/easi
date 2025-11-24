import { describe, it, expect, vi, beforeEach } from 'vitest';
import { create, type StoreApi, type UseBoundStore } from 'zustand';
import type {
  Capability,
  CapabilityDependency,
  CapabilityRealization,
} from '../../api/types';
import { ApiError } from '../../api/types';
import apiClient from '../../api/client';
import { createCapabilitySlice, type CapabilityState, type CapabilityActions } from './capabilitySlice';

vi.mock('../../api/client');

vi.mock('react-hot-toast', () => ({
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

import toast from 'react-hot-toast';
const mockToast = vi.mocked(toast);

type CapabilityStore = CapabilityState & CapabilityActions;

function buildCapability(overrides: Partial<Capability> = {}): Capability {
  const id = overrides.id ?? 'cap-1';
  return {
    id,
    name: 'Test Capability',
    level: 'L1',
    createdAt: '2024-01-01T00:00:00Z',
    _links: { self: { href: `/api/v1/capabilities/${id}` } },
    ...overrides,
  };
}

function buildDependency(overrides: Partial<CapabilityDependency> = {}): CapabilityDependency {
  const id = overrides.id ?? 'dep-1';
  return {
    id,
    sourceCapabilityId: 'cap-1',
    targetCapabilityId: 'cap-2',
    dependencyType: 'Requires',
    createdAt: '2024-01-01T00:00:00Z',
    _links: { self: { href: `/api/v1/capability-dependencies/${id}` } },
    ...overrides,
  };
}

function buildRealization(overrides: Partial<CapabilityRealization> = {}): CapabilityRealization {
  const id = overrides.id ?? 'real-1';
  return {
    id,
    capabilityId: 'cap-1',
    componentId: 'comp-1',
    realizationLevel: 'Full',
    origin: 'Direct',
    linkedAt: '2024-01-01T00:00:00Z',
    _links: { self: { href: `/api/v1/capability-realizations/${id}` } },
    ...overrides,
  };
}

describe('CapabilitySlice Tests', () => {
  let useStore: UseBoundStore<StoreApi<CapabilityStore>>;

  beforeEach(() => {
    vi.clearAllMocks();
    useStore = create<CapabilityStore>()(createCapabilitySlice);
  });

  describe('Capability Management', () => {
    describe('loadCapabilities', () => {
      it('should load capabilities and update state', async () => {
        const mockCapabilities: Capability[] = [
          {
            id: 'cap-1',
            name: 'Customer Management',
            level: 'L1',
            createdAt: '2024-01-01T00:00:00Z',
            _links: { self: { href: '/api/v1/capabilities/cap-1' } },
          },
          {
            id: 'cap-2',
            name: 'Order Processing',
            level: 'L2',
            parentId: 'cap-1',
            createdAt: '2024-01-01T00:00:00Z',
            _links: { self: { href: '/api/v1/capabilities/cap-2' } },
          },
        ];

        vi.mocked(apiClient.getCapabilities).mockResolvedValueOnce(mockCapabilities);

        await useStore.getState().loadCapabilities();

        expect(apiClient.getCapabilities).toHaveBeenCalled();
        expect(useStore.getState().capabilities).toEqual(mockCapabilities);
      });

      it('should show error toast when loading fails', async () => {
        const networkError = new Error('Network error');
        vi.mocked(apiClient.getCapabilities).mockRejectedValueOnce(networkError);

        await expect(useStore.getState().loadCapabilities()).rejects.toThrow();

        expect(mockToast.error).toHaveBeenCalledWith('Failed to load capabilities');
      });

      it('should show API error message when loading fails', async () => {
        const apiError = new ApiError('Unauthorized', 401);
        vi.mocked(apiClient.getCapabilities).mockRejectedValueOnce(apiError);

        await expect(useStore.getState().loadCapabilities()).rejects.toThrow(apiError);

        expect(mockToast.error).toHaveBeenCalledWith('Unauthorized');
      });
    });

    describe('createCapability', () => {
      it('should create capability and add to state', async () => {
        const request = {
          name: 'New Capability',
          description: 'A new capability',
          level: 'L1' as const,
        };

        const mockCapability: Capability = {
          id: 'cap-new',
          name: 'New Capability',
          description: 'A new capability',
          level: 'L1',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capabilities/cap-new' } },
        };

        vi.mocked(apiClient.createCapability).mockResolvedValueOnce(mockCapability);

        const result = await useStore.getState().createCapability(request);

        expect(apiClient.createCapability).toHaveBeenCalledWith(request);
        expect(result).toEqual(mockCapability);
        expect(useStore.getState().capabilities).toContainEqual(mockCapability);
        expect(mockToast.success).toHaveBeenCalledWith('Capability "New Capability" created');
      });

      it('should create child capability with parentId', async () => {
        const existingCapability: Capability = {
          id: 'cap-1',
          name: 'Parent',
          level: 'L1',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capabilities/cap-1' } },
        };

        useStore.setState({ capabilities: [existingCapability] });

        const request = {
          name: 'Child Capability',
          parentId: 'cap-1',
          level: 'L2' as const,
        };

        const mockCapability: Capability = {
          id: 'cap-child',
          name: 'Child Capability',
          level: 'L2',
          parentId: 'cap-1',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capabilities/cap-child' } },
        };

        vi.mocked(apiClient.createCapability).mockResolvedValueOnce(mockCapability);

        const result = await useStore.getState().createCapability(request);

        expect(result.parentId).toBe('cap-1');
        expect(useStore.getState().capabilities).toHaveLength(2);
      });

      it('should handle validation error for empty name', async () => {
        const validationError = new ApiError('Capability name is required', 400);
        vi.mocked(apiClient.createCapability).mockRejectedValueOnce(validationError);

        await expect(
          useStore.getState().createCapability({ name: '', level: 'L1' })
        ).rejects.toThrow(validationError);

        expect(mockToast.error).toHaveBeenCalledWith('Capability name is required');
      });

      it('should handle conflict error for duplicate name', async () => {
        const conflictError = new ApiError('Capability with this name already exists', 409);
        vi.mocked(apiClient.createCapability).mockRejectedValueOnce(conflictError);

        await expect(
          useStore.getState().createCapability({ name: 'Duplicate', level: 'L1' })
        ).rejects.toThrow(conflictError);

        expect(mockToast.error).toHaveBeenCalledWith('Capability with this name already exists');
      });
    });

    describe('updateCapability', () => {
      it('should update capability and reflect in state', async () => {
        const existingCapability: Capability = {
          id: 'cap-1',
          name: 'Old Name',
          level: 'L1',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capabilities/cap-1' } },
        };

        useStore.setState({ capabilities: [existingCapability] });

        const request = {
          name: 'Updated Name',
          description: 'Updated description',
        };

        const updatedCapability: Capability = {
          ...existingCapability,
          name: 'Updated Name',
          description: 'Updated description',
        };

        vi.mocked(apiClient.updateCapability).mockResolvedValueOnce(updatedCapability);

        const result = await useStore.getState().updateCapability('cap-1', request);

        expect(apiClient.updateCapability).toHaveBeenCalledWith('cap-1', request);
        expect(result).toEqual(updatedCapability);
        expect(useStore.getState().capabilities[0].name).toBe('Updated Name');
        expect(mockToast.success).toHaveBeenCalledWith('Capability "Updated Name" updated');
      });

      it('should handle not found error', async () => {
        const notFoundError = new ApiError('Capability not found', 404);
        vi.mocked(apiClient.updateCapability).mockRejectedValueOnce(notFoundError);

        await expect(
          useStore.getState().updateCapability('non-existent', { name: 'Test' })
        ).rejects.toThrow(notFoundError);

        expect(mockToast.error).toHaveBeenCalledWith('Capability not found');
      });
    });

    describe('updateCapabilityMetadata', () => {
      it('should update capability metadata', async () => {
        const existingCapability: Capability = {
          id: 'cap-1',
          name: 'Capability',
          level: 'L1',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capabilities/cap-1' } },
        };

        useStore.setState({ capabilities: [existingCapability] });

        const request = {
          strategyPillar: 'Growth',
          pillarWeight: 0.8,
          maturityLevel: 'Optimized',
          status: 'Active',
        };

        const updatedCapability: Capability = {
          ...existingCapability,
          ...request,
        };

        vi.mocked(apiClient.updateCapabilityMetadata).mockResolvedValueOnce(updatedCapability);

        const result = await useStore.getState().updateCapabilityMetadata('cap-1', request);

        expect(apiClient.updateCapabilityMetadata).toHaveBeenCalledWith('cap-1', request);
        expect(result.strategyPillar).toBe('Growth');
        expect(mockToast.success).toHaveBeenCalledWith('Capability metadata updated');
      });
    });

    describe('addCapabilityExpert', () => {
      it('should add expert and refresh capability', async () => {
        const existingCapability: Capability = {
          id: 'cap-1',
          name: 'Capability',
          level: 'L1',
          experts: [],
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capabilities/cap-1' } },
        };

        useStore.setState({ capabilities: [existingCapability] });

        const request = {
          expertName: 'John Doe',
          expertRole: 'Architect',
          contactInfo: 'john@example.com',
        };

        const updatedCapability: Capability = {
          ...existingCapability,
          experts: [
            {
              name: 'John Doe',
              role: 'Architect',
              contact: 'john@example.com',
              addedAt: '2024-01-01T00:00:00Z',
            },
          ],
        };

        vi.mocked(apiClient.addCapabilityExpert).mockResolvedValueOnce(undefined);
        vi.mocked(apiClient.getCapabilityById).mockResolvedValueOnce(updatedCapability);

        await useStore.getState().addCapabilityExpert('cap-1', request);

        expect(apiClient.addCapabilityExpert).toHaveBeenCalledWith('cap-1', request);
        expect(apiClient.getCapabilityById).toHaveBeenCalledWith('cap-1');
        expect(useStore.getState().capabilities[0].experts).toHaveLength(1);
        expect(mockToast.success).toHaveBeenCalledWith('Expert "John Doe" added');
      });
    });

    describe('addCapabilityTag', () => {
      it('should add tag and refresh capability', async () => {
        const existingCapability: Capability = {
          id: 'cap-1',
          name: 'Capability',
          level: 'L1',
          tags: [],
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capabilities/cap-1' } },
        };

        useStore.setState({ capabilities: [existingCapability] });

        const updatedCapability: Capability = {
          ...existingCapability,
          tags: ['core'],
        };

        vi.mocked(apiClient.addCapabilityTag).mockResolvedValueOnce(undefined);
        vi.mocked(apiClient.getCapabilityById).mockResolvedValueOnce(updatedCapability);

        await useStore.getState().addCapabilityTag('cap-1', 'core');

        expect(apiClient.addCapabilityTag).toHaveBeenCalledWith('cap-1', { tag: 'core' });
        expect(useStore.getState().capabilities[0].tags).toContain('core');
        expect(mockToast.success).toHaveBeenCalledWith('Tag "core" added');
      });
    });
  });

  describe('Dependency Management', () => {
    describe('loadCapabilityDependencies', () => {
      it('should load dependencies and update state', async () => {
        const mockDependencies: CapabilityDependency[] = [
          {
            id: 'dep-1',
            sourceCapabilityId: 'cap-1',
            targetCapabilityId: 'cap-2',
            dependencyType: 'Requires',
            createdAt: '2024-01-01T00:00:00Z',
            _links: { self: { href: '/api/v1/capability-dependencies/dep-1' } },
          },
        ];

        vi.mocked(apiClient.getCapabilityDependencies).mockResolvedValueOnce(mockDependencies);

        await useStore.getState().loadCapabilityDependencies();

        expect(apiClient.getCapabilityDependencies).toHaveBeenCalled();
        expect(useStore.getState().capabilityDependencies).toEqual(mockDependencies);
      });
    });

    describe('createCapabilityDependency', () => {
      it('should create dependency and add to state', async () => {
        const request = {
          sourceCapabilityId: 'cap-1',
          targetCapabilityId: 'cap-2',
          dependencyType: 'Requires' as const,
        };

        const mockDependency: CapabilityDependency = {
          id: 'dep-new',
          sourceCapabilityId: 'cap-1',
          targetCapabilityId: 'cap-2',
          dependencyType: 'Requires',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capability-dependencies/dep-new' } },
        };

        vi.mocked(apiClient.createCapabilityDependency).mockResolvedValueOnce(mockDependency);

        const result = await useStore.getState().createCapabilityDependency(request);

        expect(apiClient.createCapabilityDependency).toHaveBeenCalledWith(request);
        expect(result).toEqual(mockDependency);
        expect(useStore.getState().capabilityDependencies).toContainEqual(mockDependency);
        expect(mockToast.success).toHaveBeenCalledWith('Dependency created');
      });

      it('should handle self-dependency error', async () => {
        const validationError = new ApiError('Source and target capabilities must be different', 400);
        vi.mocked(apiClient.createCapabilityDependency).mockRejectedValueOnce(validationError);

        await expect(
          useStore.getState().createCapabilityDependency({
            sourceCapabilityId: 'cap-1',
            targetCapabilityId: 'cap-1',
            dependencyType: 'Requires',
          })
        ).rejects.toThrow(validationError);

        expect(mockToast.error).toHaveBeenCalledWith('Source and target capabilities must be different');
      });
    });

    describe('deleteCapabilityDependency', () => {
      it('should delete dependency and remove from state', async () => {
        const existingDependency: CapabilityDependency = {
          id: 'dep-1',
          sourceCapabilityId: 'cap-1',
          targetCapabilityId: 'cap-2',
          dependencyType: 'Requires',
          createdAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capability-dependencies/dep-1' } },
        };

        useStore.setState({ capabilityDependencies: [existingDependency] });

        vi.mocked(apiClient.deleteCapabilityDependency).mockResolvedValueOnce(undefined);

        await useStore.getState().deleteCapabilityDependency('dep-1');

        expect(apiClient.deleteCapabilityDependency).toHaveBeenCalledWith('dep-1');
        expect(useStore.getState().capabilityDependencies).toHaveLength(0);
        expect(mockToast.success).toHaveBeenCalledWith('Dependency deleted');
      });

      it('should handle delete error', async () => {
        const deleteError = new ApiError('Dependency not found', 404);
        vi.mocked(apiClient.deleteCapabilityDependency).mockRejectedValueOnce(deleteError);

        await expect(
          useStore.getState().deleteCapabilityDependency('non-existent')
        ).rejects.toThrow(deleteError);

        expect(mockToast.error).toHaveBeenCalledWith('Dependency not found');
      });
    });
  });

  describe('Realization Management', () => {
    describe('linkSystemToCapability', () => {
      it('should link system and fetch all realizations including inherited', async () => {
        const request = {
          componentId: 'comp-1',
          realizationLevel: 'Full' as const,
          notes: 'Primary system',
        };

        const directRealization: CapabilityRealization = {
          id: 'real-new',
          capabilityId: 'cap-l3',
          componentId: 'comp-1',
          realizationLevel: 'Full',
          notes: 'Primary system',
          origin: 'Direct',
          linkedAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capability-realizations/real-new' } },
        };

        const inheritedL2: CapabilityRealization = {
          id: 'real-inherited-l2',
          capabilityId: 'cap-l2',
          componentId: 'comp-1',
          realizationLevel: 'Full',
          origin: 'Inherited',
          sourceRealizationId: 'real-new',
          linkedAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capability-realizations/real-inherited-l2' } },
        };

        const inheritedL1: CapabilityRealization = {
          id: 'real-inherited-l1',
          capabilityId: 'cap-l1',
          componentId: 'comp-1',
          realizationLevel: 'Full',
          origin: 'Inherited',
          sourceRealizationId: 'real-new',
          linkedAt: '2024-01-01T00:00:00Z',
          _links: { self: { href: '/api/v1/capability-realizations/real-inherited-l1' } },
        };

        vi.mocked(apiClient.linkSystemToCapability).mockResolvedValueOnce(directRealization);
        vi.mocked(apiClient.getCapabilitiesByComponent).mockResolvedValueOnce([
          directRealization,
          inheritedL2,
          inheritedL1,
        ]);

        const result = await useStore.getState().linkSystemToCapability('cap-l3', request);

        expect(apiClient.linkSystemToCapability).toHaveBeenCalledWith('cap-l3', request);
        expect(apiClient.getCapabilitiesByComponent).toHaveBeenCalledWith('comp-1');
        expect(result).toEqual(directRealization);

        const state = useStore.getState();
        expect(state.capabilityRealizations).toHaveLength(3);
        expect(state.capabilityRealizations).toContainEqual(directRealization);
        expect(state.capabilityRealizations).toContainEqual(inheritedL2);
        expect(state.capabilityRealizations).toContainEqual(inheritedL1);
        expect(mockToast.success).toHaveBeenCalledWith('System linked to capability');
      });

      it('should handle duplicate link error', async () => {
        const conflictError = new ApiError('System is already linked to this capability', 409);
        vi.mocked(apiClient.linkSystemToCapability).mockRejectedValueOnce(conflictError);

        await expect(
          useStore.getState().linkSystemToCapability('cap-1', {
            componentId: 'comp-1',
            realizationLevel: 'Full',
          })
        ).rejects.toThrow(conflictError);

        expect(mockToast.error).toHaveBeenCalledWith('System is already linked to this capability');
      });
    });

    describe('updateRealization', () => {
      it('should update realization and reflect in state', async () => {
        const existingRealization = buildRealization();
        useStore.setState({ capabilityRealizations: [existingRealization] });

        const request = {
          realizationLevel: 'Partial' as const,
          notes: 'Updated notes',
        };

        const updatedRealization: CapabilityRealization = {
          ...existingRealization,
          realizationLevel: 'Partial',
          notes: 'Updated notes',
        };

        vi.mocked(apiClient.updateRealization).mockResolvedValueOnce(updatedRealization);

        const result = await useStore.getState().updateRealization('real-1', request);

        expect(apiClient.updateRealization).toHaveBeenCalledWith('real-1', request);
        expect(result.realizationLevel).toBe('Partial');
        expect(useStore.getState().capabilityRealizations[0].notes).toBe('Updated notes');
        expect(mockToast.success).toHaveBeenCalledWith('Realization updated');
      });
    });

    describe('deleteRealization', () => {
      it('should delete realization and remove from state', async () => {
        const existingRealization = buildRealization();
        useStore.setState({ capabilityRealizations: [existingRealization] });

        vi.mocked(apiClient.deleteRealization).mockResolvedValueOnce(undefined);

        await useStore.getState().deleteRealization('real-1');

        expect(apiClient.deleteRealization).toHaveBeenCalledWith('real-1');
        expect(useStore.getState().capabilityRealizations).toHaveLength(0);
        expect(mockToast.success).toHaveBeenCalledWith('Realization deleted');
      });

      it('should cascade delete inherited realizations from state', async () => {
        const directRealization = buildRealization({ id: 'real-direct', capabilityId: 'cap-l3' });
        const inheritedL2 = buildRealization({
          id: 'real-inherited-l2',
          capabilityId: 'cap-l2',
          origin: 'Inherited',
          sourceRealizationId: 'real-direct',
        });
        const inheritedL1 = buildRealization({
          id: 'real-inherited-l1',
          capabilityId: 'cap-l1',
          origin: 'Inherited',
          sourceRealizationId: 'real-direct',
        });
        const unrelatedRealization = buildRealization({
          id: 'real-unrelated',
          capabilityId: 'cap-other',
          componentId: 'comp-2',
          realizationLevel: 'Partial',
        });

        useStore.setState({
          capabilityRealizations: [directRealization, inheritedL2, inheritedL1, unrelatedRealization],
        });

        vi.mocked(apiClient.deleteRealization).mockResolvedValueOnce(undefined);

        await useStore.getState().deleteRealization('real-direct');

        expect(apiClient.deleteRealization).toHaveBeenCalledWith('real-direct');

        const state = useStore.getState();
        expect(state.capabilityRealizations).toHaveLength(1);
        expect(state.capabilityRealizations[0].id).toBe('real-unrelated');
        expect(state.capabilityRealizations).not.toContainEqual(directRealization);
        expect(state.capabilityRealizations).not.toContainEqual(inheritedL2);
        expect(state.capabilityRealizations).not.toContainEqual(inheritedL1);
        expect(mockToast.success).toHaveBeenCalledWith('Realization deleted');
      });

      it('should handle delete error', async () => {
        const deleteError = new ApiError('Realization not found', 404);
        vi.mocked(apiClient.deleteRealization).mockRejectedValueOnce(deleteError);

        await expect(
          useStore.getState().deleteRealization('non-existent')
        ).rejects.toThrow(deleteError);

        expect(mockToast.error).toHaveBeenCalledWith('Realization not found');
      });
    });

    describe('loadRealizationsByCapability', () => {
      it('should load realizations for a capability', async () => {
        const mockRealizations = [buildRealization()];

        vi.mocked(apiClient.getSystemsByCapability).mockResolvedValueOnce(mockRealizations);

        const result = await useStore.getState().loadRealizationsByCapability('cap-1');

        expect(apiClient.getSystemsByCapability).toHaveBeenCalledWith('cap-1');
        expect(result).toEqual(mockRealizations);
        expect(useStore.getState().capabilityRealizations).toContainEqual(mockRealizations[0]);
      });

      it('should merge new realizations with existing ones', async () => {
        const existingRealization = buildRealization({
          id: 'real-existing',
          capabilityId: 'cap-2',
          componentId: 'comp-2',
          realizationLevel: 'Partial',
        });
        useStore.setState({ capabilityRealizations: [existingRealization] });

        const newRealizations = [buildRealization()];

        vi.mocked(apiClient.getSystemsByCapability).mockResolvedValueOnce(newRealizations);

        await useStore.getState().loadRealizationsByCapability('cap-1');

        const state = useStore.getState();
        expect(state.capabilityRealizations).toHaveLength(2);
        expect(state.capabilityRealizations).toContainEqual(existingRealization);
        expect(state.capabilityRealizations).toContainEqual(newRealizations[0]);
      });

      it('should replace existing realizations with same id', async () => {
        const existingRealization = buildRealization({ realizationLevel: 'Partial' });
        useStore.setState({ capabilityRealizations: [existingRealization] });

        const updatedRealizations = [
          buildRealization({ realizationLevel: 'Full', notes: 'Updated' }),
        ];

        vi.mocked(apiClient.getSystemsByCapability).mockResolvedValueOnce(updatedRealizations);

        await useStore.getState().loadRealizationsByCapability('cap-1');

        const state = useStore.getState();
        expect(state.capabilityRealizations).toHaveLength(1);
        expect(state.capabilityRealizations[0].realizationLevel).toBe('Full');
      });
    });

    describe('loadRealizationsByComponent', () => {
      it('should load realizations for a component', async () => {
        const mockRealizations = [
          buildRealization(),
          buildRealization({ id: 'real-2', capabilityId: 'cap-2', realizationLevel: 'Partial' }),
        ];

        vi.mocked(apiClient.getCapabilitiesByComponent).mockResolvedValueOnce(mockRealizations);

        const result = await useStore.getState().loadRealizationsByComponent('comp-1');

        expect(apiClient.getCapabilitiesByComponent).toHaveBeenCalledWith('comp-1');
        expect(result).toEqual(mockRealizations);
        expect(useStore.getState().capabilityRealizations).toHaveLength(2);
      });
    });
  });

  describe('Error Handling', () => {
    it('should handle 500 server error when creating capability', async () => {
      const serverError = new ApiError('Internal server error', 500);
      vi.mocked(apiClient.createCapability).mockRejectedValueOnce(serverError);

      await expect(
        useStore.getState().createCapability({ name: 'Test', level: 'L1' })
      ).rejects.toThrow(serverError);

      expect(mockToast.error).toHaveBeenCalledWith('Internal server error');
    });

    it('should handle network error when loading dependencies', async () => {
      const networkError = new Error('Network error');
      vi.mocked(apiClient.getCapabilityDependencies).mockRejectedValueOnce(networkError);

      await expect(
        useStore.getState().loadCapabilityDependencies()
      ).rejects.toThrow();

      expect(mockToast.error).toHaveBeenCalledWith('Failed to load capability dependencies');
    });

    it('should handle unauthorized error', async () => {
      const authError = new ApiError('Unauthorized', 401);
      vi.mocked(apiClient.getCapabilities).mockRejectedValueOnce(authError);

      await expect(useStore.getState().loadCapabilities()).rejects.toThrow(authError);

      expect(mockToast.error).toHaveBeenCalledWith('Unauthorized');
    });

    it('should handle forbidden error when creating dependency', async () => {
      const forbiddenError = new ApiError('Forbidden', 403);
      vi.mocked(apiClient.createCapabilityDependency).mockRejectedValueOnce(forbiddenError);

      await expect(
        useStore.getState().createCapabilityDependency({
          sourceCapabilityId: 'cap-1',
          targetCapabilityId: 'cap-2',
          dependencyType: 'Requires',
        })
      ).rejects.toThrow(forbiddenError);

      expect(mockToast.error).toHaveBeenCalledWith('Forbidden');
    });
  });
});
