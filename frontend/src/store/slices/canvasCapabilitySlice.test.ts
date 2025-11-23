import { describe, it, expect, vi, beforeEach } from 'vitest';
import { create } from 'zustand';
import type { View } from '../../api/types';
import { ApiError } from '../../api/types';
import apiClient from '../../api/client';
import {
  createCanvasCapabilitySlice,
  type CanvasCapabilityState,
  type CanvasCapabilityActions,
} from './canvasCapabilitySlice';

vi.mock('../../api/client');

type TestStore = CanvasCapabilityState &
  CanvasCapabilityActions & { currentView: View | null; setCurrentView: (view: View | null) => void };

const createTestView = (overrides: Partial<View> = {}): View => ({
  id: 'view-1',
  name: 'Test View',
  isDefault: false,
  components: [],
  capabilities: [],
  createdAt: '2024-01-01T00:00:00Z',
  _links: { self: { href: '/api/v1/views/view-1' } },
  ...overrides,
});

const createTestStore = () =>
  create<TestStore>((set, get, store) => ({
    currentView: null,
    setCurrentView: (view: View | null) => set({ currentView: view }),
    ...createCanvasCapabilitySlice(set, get, store),
  }));

describe('CanvasCapabilitySlice Tests', () => {
  let useStore: ReturnType<typeof createTestStore>;

  beforeEach(() => {
    vi.clearAllMocks();
    useStore = createTestStore();
  });

  describe('addCapabilityToCanvas', () => {
    it('should add capability to canvas with optimistic update', async () => {
      const testView = createTestView();
      useStore.getState().setCurrentView(testView);

      vi.mocked(apiClient.addCapabilityToView).mockResolvedValueOnce(undefined);

      await useStore.getState().addCapabilityToCanvas('cap-1', 100, 200);

      expect(useStore.getState().canvasCapabilities).toEqual([
        { capabilityId: 'cap-1', x: 100, y: 200 },
      ]);
      expect(apiClient.addCapabilityToView).toHaveBeenCalledWith('view-1', {
        capabilityId: 'cap-1',
        x: 100,
        y: 200,
      });
    });

    it('should not add capability when no current view', async () => {

      await useStore.getState().addCapabilityToCanvas('cap-1', 100, 200);

      expect(useStore.getState().canvasCapabilities).toEqual([]);
      expect(apiClient.addCapabilityToView).not.toHaveBeenCalled();
    });

    it('should not add duplicate capability to canvas', async () => {
      const testView = createTestView();
      useStore.getState().setCurrentView(testView);
      useStore.setState({ canvasCapabilities: [{ capabilityId: 'cap-1', x: 50, y: 50 }] });

      await useStore.getState().addCapabilityToCanvas('cap-1', 100, 200);

      expect(useStore.getState().canvasCapabilities).toHaveLength(1);
      expect(useStore.getState().canvasCapabilities[0]).toEqual({
        capabilityId: 'cap-1',
        x: 50,
        y: 50,
      });
      expect(apiClient.addCapabilityToView).not.toHaveBeenCalled();
    });

    it('should rollback optimistic update on API error', async () => {
      const testView = createTestView();
      useStore.getState().setCurrentView(testView);

      const apiError = new ApiError('Failed to add capability', 500);
      vi.mocked(apiClient.addCapabilityToView).mockRejectedValueOnce(apiError);

      await expect(
        useStore.getState().addCapabilityToCanvas('cap-1', 100, 200)
      ).rejects.toThrow(apiError);

      expect(useStore.getState().canvasCapabilities).toEqual([]);
    });

    it('should preserve existing capabilities on rollback', async () => {
      const testView = createTestView();
      const existingCapability = { capabilityId: 'cap-existing', x: 50, y: 50 };
      useStore.getState().setCurrentView(testView);
      useStore.setState({ canvasCapabilities: [existingCapability] });

      const apiError = new ApiError('Server error', 500);
      vi.mocked(apiClient.addCapabilityToView).mockRejectedValueOnce(apiError);

      await expect(
        useStore.getState().addCapabilityToCanvas('cap-new', 100, 200)
      ).rejects.toThrow();

      expect(useStore.getState().canvasCapabilities).toEqual([existingCapability]);
    });

    it('should handle 400 validation error from API', async () => {
      const testView = createTestView();
      useStore.getState().setCurrentView(testView);

      const validationError = new ApiError('Invalid capability ID', 400);
      vi.mocked(apiClient.addCapabilityToView).mockRejectedValueOnce(validationError);

      await expect(
        useStore.getState().addCapabilityToCanvas('invalid-id', 100, 200)
      ).rejects.toThrow(validationError);

      expect(useStore.getState().canvasCapabilities).toEqual([]);
    });

    it('should handle 404 capability not found error', async () => {
      const testView = createTestView();
      useStore.getState().setCurrentView(testView);

      const notFoundError = new ApiError('Capability not found', 404);
      vi.mocked(apiClient.addCapabilityToView).mockRejectedValueOnce(notFoundError);

      await expect(
        useStore.getState().addCapabilityToCanvas('non-existent', 100, 200)
      ).rejects.toThrow(notFoundError);

      expect(useStore.getState().canvasCapabilities).toEqual([]);
    });
  });

  describe('removeCapabilityFromCanvas', () => {
    it('should remove capability from canvas with optimistic update', async () => {
      const testView = createTestView();
      useStore.getState().setCurrentView(testView);
      useStore.setState({
        canvasCapabilities: [
          { capabilityId: 'cap-1', x: 100, y: 200 },
          { capabilityId: 'cap-2', x: 300, y: 400 },
        ],
      });

      vi.mocked(apiClient.removeCapabilityFromView).mockResolvedValueOnce(undefined);

      await useStore.getState().removeCapabilityFromCanvas('cap-1');

      expect(useStore.getState().canvasCapabilities).toEqual([
        { capabilityId: 'cap-2', x: 300, y: 400 },
      ]);
      expect(apiClient.removeCapabilityFromView).toHaveBeenCalledWith('view-1', 'cap-1');
    });

    it('should not call API when no current view', async () => {
      useStore.setState({
        canvasCapabilities: [{ capabilityId: 'cap-1', x: 100, y: 200 }],
      });

      await useStore.getState().removeCapabilityFromCanvas('cap-1');

      expect(apiClient.removeCapabilityFromView).not.toHaveBeenCalled();
    });

    it('should clear selection when removed capability was selected', async () => {
      const testView = createTestView();
      useStore.getState().setCurrentView(testView);
      useStore.setState({
        canvasCapabilities: [{ capabilityId: 'cap-1', x: 100, y: 200 }],
        selectedCapabilityId: 'cap-1',
      });

      vi.mocked(apiClient.removeCapabilityFromView).mockResolvedValueOnce(undefined);

      await useStore.getState().removeCapabilityFromCanvas('cap-1');

      expect(useStore.getState().selectedCapabilityId).toBeNull();
    });

    it('should preserve selection when different capability removed', async () => {
      const testView = createTestView();
      useStore.getState().setCurrentView(testView);
      useStore.setState({
        canvasCapabilities: [
          { capabilityId: 'cap-1', x: 100, y: 200 },
          { capabilityId: 'cap-2', x: 300, y: 400 },
        ],
        selectedCapabilityId: 'cap-2',
      });

      vi.mocked(apiClient.removeCapabilityFromView).mockResolvedValueOnce(undefined);

      await useStore.getState().removeCapabilityFromCanvas('cap-1');

      expect(useStore.getState().selectedCapabilityId).toBe('cap-2');
    });

    it('should rollback optimistic update on API error', async () => {
      const testView = createTestView();
      const capability = { capabilityId: 'cap-1', x: 100, y: 200 };
      useStore.getState().setCurrentView(testView);
      useStore.setState({ canvasCapabilities: [capability] });

      const apiError = new ApiError('Failed to remove capability', 500);
      vi.mocked(apiClient.removeCapabilityFromView).mockRejectedValueOnce(apiError);

      await expect(
        useStore.getState().removeCapabilityFromCanvas('cap-1')
      ).rejects.toThrow(apiError);

      expect(useStore.getState().canvasCapabilities).toContainEqual(capability);
    });

    it('should handle 404 error when capability not in view', async () => {
      const testView = createTestView();
      const capability = { capabilityId: 'cap-1', x: 100, y: 200 };
      useStore.getState().setCurrentView(testView);
      useStore.setState({ canvasCapabilities: [capability] });

      const notFoundError = new ApiError('Capability not found in view', 404);
      vi.mocked(apiClient.removeCapabilityFromView).mockRejectedValueOnce(notFoundError);

      await expect(
        useStore.getState().removeCapabilityFromCanvas('cap-1')
      ).rejects.toThrow(notFoundError);

      expect(useStore.getState().canvasCapabilities).toContainEqual(capability);
    });
  });

  describe('updateCapabilityPosition', () => {
    it('should update position in state and call API', () => {
      const testView = createTestView();
      useStore.getState().setCurrentView(testView);
      useStore.setState({
        canvasCapabilities: [{ capabilityId: 'cap-1', x: 100, y: 200 }],
      });

      vi.mocked(apiClient.updateCapabilityPositionInView).mockResolvedValueOnce(undefined);

      useStore.getState().updateCapabilityPosition('cap-1', 150, 250);

      expect(useStore.getState().canvasCapabilities[0]).toEqual({
        capabilityId: 'cap-1',
        x: 150,
        y: 250,
      });
      expect(apiClient.updateCapabilityPositionInView).toHaveBeenCalledWith(
        'view-1',
        'cap-1',
        150,
        250
      );
    });

    it('should not call API when no current view', () => {
      useStore.setState({
        canvasCapabilities: [{ capabilityId: 'cap-1', x: 100, y: 200 }],
      });

      useStore.getState().updateCapabilityPosition('cap-1', 150, 250);

      expect(useStore.getState().canvasCapabilities[0]).toEqual({
        capabilityId: 'cap-1',
        x: 150,
        y: 250,
      });
      expect(apiClient.updateCapabilityPositionInView).not.toHaveBeenCalled();
    });

    it('should preserve other capabilities when updating one', () => {
      const testView = createTestView();
      useStore.getState().setCurrentView(testView);
      useStore.setState({
        canvasCapabilities: [
          { capabilityId: 'cap-1', x: 100, y: 200 },
          { capabilityId: 'cap-2', x: 300, y: 400 },
        ],
      });

      vi.mocked(apiClient.updateCapabilityPositionInView).mockResolvedValueOnce(undefined);

      useStore.getState().updateCapabilityPosition('cap-1', 150, 250);

      expect(useStore.getState().canvasCapabilities).toEqual([
        { capabilityId: 'cap-1', x: 150, y: 250 },
        { capabilityId: 'cap-2', x: 300, y: 400 },
      ]);
    });

    it('should silently handle API errors without rollback', async () => {
      const testView = createTestView();
      useStore.getState().setCurrentView(testView);
      useStore.setState({
        canvasCapabilities: [{ capabilityId: 'cap-1', x: 100, y: 200 }],
      });

      vi.mocked(apiClient.updateCapabilityPositionInView).mockRejectedValueOnce(
        new ApiError('Server error', 500)
      );

      useStore.getState().updateCapabilityPosition('cap-1', 150, 250);

      expect(useStore.getState().canvasCapabilities[0]).toEqual({
        capabilityId: 'cap-1',
        x: 150,
        y: 250,
      });
    });

    it('should not update position of non-existent capability', () => {
      const testView = createTestView();
      useStore.getState().setCurrentView(testView);
      useStore.setState({
        canvasCapabilities: [{ capabilityId: 'cap-1', x: 100, y: 200 }],
      });

      vi.mocked(apiClient.updateCapabilityPositionInView).mockResolvedValueOnce(undefined);

      useStore.getState().updateCapabilityPosition('non-existent', 150, 250);

      expect(useStore.getState().canvasCapabilities).toEqual([
        { capabilityId: 'cap-1', x: 100, y: 200 },
      ]);
    });
  });

  describe('selectCapability', () => {
    it('should select a capability by id', () => {
      useStore.setState({
        canvasCapabilities: [{ capabilityId: 'cap-1', x: 100, y: 200 }],
      });

      useStore.getState().selectCapability('cap-1');

      expect(useStore.getState().selectedCapabilityId).toBe('cap-1');
    });

    it('should clear selection when null is passed', () => {
      useStore.setState({
        selectedCapabilityId: 'cap-1',
      });

      useStore.getState().selectCapability(null);

      expect(useStore.getState().selectedCapabilityId).toBeNull();
    });

    it('should change selection to a different capability', () => {
      useStore.setState({
        canvasCapabilities: [
          { capabilityId: 'cap-1', x: 100, y: 200 },
          { capabilityId: 'cap-2', x: 300, y: 400 },
        ],
        selectedCapabilityId: 'cap-1',
      });

      useStore.getState().selectCapability('cap-2');

      expect(useStore.getState().selectedCapabilityId).toBe('cap-2');
    });
  });

  describe('clearCanvasCapabilities', () => {
    it('should clear all canvas capabilities and selection', () => {
      useStore.setState({
        canvasCapabilities: [
          { capabilityId: 'cap-1', x: 100, y: 200 },
          { capabilityId: 'cap-2', x: 300, y: 400 },
        ],
        selectedCapabilityId: 'cap-1',
      });

      useStore.getState().clearCanvasCapabilities();

      expect(useStore.getState().canvasCapabilities).toEqual([]);
      expect(useStore.getState().selectedCapabilityId).toBeNull();
    });

    it('should work when already empty', () => {
      useStore.setState({
        canvasCapabilities: [],
        selectedCapabilityId: null,
      });

      useStore.getState().clearCanvasCapabilities();

      expect(useStore.getState().canvasCapabilities).toEqual([]);
      expect(useStore.getState().selectedCapabilityId).toBeNull();
    });
  });

  describe('syncCanvasCapabilitiesFromView', () => {
    it('should sync capabilities from view data', () => {
      const viewWithCapabilities = createTestView({
        capabilities: [
          { capabilityId: 'cap-1', x: 100, y: 200 },
          { capabilityId: 'cap-2', x: 300, y: 400 },
        ],
      });

      useStore.getState().syncCanvasCapabilitiesFromView(viewWithCapabilities);

      expect(useStore.getState().canvasCapabilities).toEqual([
        { capabilityId: 'cap-1', x: 100, y: 200 },
        { capabilityId: 'cap-2', x: 300, y: 400 },
      ]);
    });

    it('should clear selection when syncing', () => {
      useStore.setState({ selectedCapabilityId: 'cap-1' });
      const viewWithCapabilities = createTestView({
        capabilities: [{ capabilityId: 'cap-1', x: 100, y: 200 }],
      });

      useStore.getState().syncCanvasCapabilitiesFromView(viewWithCapabilities);

      expect(useStore.getState().selectedCapabilityId).toBeNull();
    });

    it('should handle view with no capabilities', () => {
      useStore.setState({
        canvasCapabilities: [{ capabilityId: 'cap-1', x: 100, y: 200 }],
      });
      const emptyView = createTestView({ capabilities: [] });

      useStore.getState().syncCanvasCapabilitiesFromView(emptyView);

      expect(useStore.getState().canvasCapabilities).toEqual([]);
    });

    it('should handle view with undefined capabilities', () => {
      useStore.setState({
        canvasCapabilities: [{ capabilityId: 'cap-1', x: 100, y: 200 }],
      });
      const viewWithUndefinedCapabilities = {
        ...createTestView(),
        capabilities: undefined as any,
      };

      useStore.getState().syncCanvasCapabilitiesFromView(viewWithUndefinedCapabilities);

      expect(useStore.getState().canvasCapabilities).toEqual([]);
    });

    it('should replace existing capabilities with new ones from view', () => {
      useStore.setState({
        canvasCapabilities: [
          { capabilityId: 'old-cap', x: 50, y: 50 },
        ],
      });
      const newView = createTestView({
        capabilities: [
          { capabilityId: 'new-cap', x: 100, y: 200 },
        ],
      });

      useStore.getState().syncCanvasCapabilitiesFromView(newView);

      expect(useStore.getState().canvasCapabilities).toEqual([
        { capabilityId: 'new-cap', x: 100, y: 200 },
      ]);
    });
  });

  describe('Integration Scenarios', () => {
    it('should handle full drag-and-drop workflow: add, move, select', async () => {
      const testView = createTestView();
      useStore.getState().setCurrentView(testView);

      vi.mocked(apiClient.addCapabilityToView).mockResolvedValueOnce(undefined);
      vi.mocked(apiClient.updateCapabilityPositionInView).mockResolvedValueOnce(undefined);

      await useStore.getState().addCapabilityToCanvas('cap-1', 100, 200);
      useStore.getState().updateCapabilityPosition('cap-1', 150, 250);
      useStore.getState().selectCapability('cap-1');

      expect(useStore.getState().canvasCapabilities).toEqual([
        { capabilityId: 'cap-1', x: 150, y: 250 },
      ]);
      expect(useStore.getState().selectedCapabilityId).toBe('cap-1');
    });

    it('should handle remove and clear workflow', async () => {
      const testView = createTestView();
      useStore.getState().setCurrentView(testView);
      useStore.setState({
        canvasCapabilities: [
          { capabilityId: 'cap-1', x: 100, y: 200 },
          { capabilityId: 'cap-2', x: 300, y: 400 },
        ],
        selectedCapabilityId: 'cap-1',
      });

      vi.mocked(apiClient.removeCapabilityFromView).mockResolvedValueOnce(undefined);

      await useStore.getState().removeCapabilityFromCanvas('cap-1');

      expect(useStore.getState().canvasCapabilities).toHaveLength(1);
      expect(useStore.getState().selectedCapabilityId).toBeNull();

      useStore.getState().clearCanvasCapabilities();

      expect(useStore.getState().canvasCapabilities).toEqual([]);
    });

    it('should handle view switch scenario', async () => {
      const view1 = createTestView({
        id: 'view-1',
        capabilities: [{ capabilityId: 'cap-1', x: 100, y: 200 }],
      });
      const view2 = createTestView({
        id: 'view-2',
        capabilities: [{ capabilityId: 'cap-2', x: 300, y: 400 }],
      });

      useStore.getState().syncCanvasCapabilitiesFromView(view1);
      expect(useStore.getState().canvasCapabilities[0].capabilityId).toBe('cap-1');

      useStore.getState().syncCanvasCapabilitiesFromView(view2);
      expect(useStore.getState().canvasCapabilities[0].capabilityId).toBe('cap-2');
    });
  });
});
