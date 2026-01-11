import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { CapabilityDetails } from './CapabilityDetails';
import type { Capability, View, CapabilityId, ViewId } from '../../../api/types';
import { createMantineTestWrapper, seedDb, server } from '../../../test/helpers';
import { useAppStore } from '../../../store/appStore';
import { useCurrentView } from '../../views/hooks/useCurrentView';

const API_BASE = 'http://localhost:8080';

vi.mock('../../../store/appStore', () => ({
  useAppStore: vi.fn(),
}));

vi.mock('../../views/hooks/useCurrentView', () => ({
  useCurrentView: vi.fn(),
}));

const mockCapability: Capability = {
  id: 'cap-1' as CapabilityId,
  name: 'Test Capability',
  description: 'Test description',
  level: 'L2',
  maturityLevel: 'Product',
  createdAt: '2024-01-01T00:00:00Z',
  _links: { self: { href: '/api/v1/capabilities/cap-1', method: 'GET' as const } },
};

const createMockView = (colorScheme: string, customColor?: string): View => ({
  id: 'view-1' as ViewId,
  name: 'Test View',
  isDefault: true,
  components: [],
  capabilities: [
    { capabilityId: 'cap-1' as CapabilityId, x: 100, y: 200, customColor },
  ],
  colorScheme,
  createdAt: '2024-01-01T00:00:00Z',
  _links: { self: { href: '/api/v1/views/view-1', method: 'GET' as const } },
});

const createMockStore = () => ({
  selectedCapabilityId: 'cap-1',
  selectCapability: vi.fn(),
});

describe('CapabilityDetails - ColorPicker Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    seedDb({
      capabilities: [mockCapability],
      components: [],
    });
  });

  const renderCapabilityDetails = (view: View | null) => {
    vi.mocked(useAppStore).mockImplementation((selector: (state: unknown) => unknown) => selector(createMockStore()));
    vi.mocked(useCurrentView).mockReturnValue({
      currentView: view,
      currentViewId: view?.id ?? null,
      isLoading: false,
      error: null,
    });
    const { Wrapper } = createMantineTestWrapper();
    return render(<CapabilityDetails onRemoveFromView={vi.fn()} />, { wrapper: Wrapper });
  };

  describe('Color picker visibility', () => {
    it('should show color picker in capability details panel', async () => {
      const mockView = createMockView('custom');
      renderCapabilityDetails(mockView);

      await waitFor(() => {
        expect(screen.getByTestId('color-picker')).toBeInTheDocument();
      });
    });

    it('should show color picker label', async () => {
      const mockView = createMockView('custom');
      renderCapabilityDetails(mockView);

      await waitFor(() => {
        expect(screen.getByText('Custom Color')).toBeInTheDocument();
      });
    });
  });

  describe('Color picker enabled state based on color scheme', () => {
    it('should enable color picker when colorScheme is "custom"', async () => {
      const mockView = createMockView('custom');
      renderCapabilityDetails(mockView);

      await waitFor(() => {
        const colorPickerButton = screen.getByTestId('color-picker-button');
        expect(colorPickerButton).not.toBeDisabled();
      });
    });

    it('should disable color picker when colorScheme is "maturity"', async () => {
      const mockView = createMockView('maturity');
      renderCapabilityDetails(mockView);

      await waitFor(() => {
        const colorPickerButton = screen.getByTestId('color-picker-button');
        expect(colorPickerButton).toBeDisabled();
      });
    });

    it('should disable color picker when colorScheme is "classic"', async () => {
      const mockView = createMockView('classic');
      renderCapabilityDetails(mockView);

      await waitFor(() => {
        const colorPickerButton = screen.getByTestId('color-picker-button');
        expect(colorPickerButton).toBeDisabled();
      });
    });

    it('should show tooltip when disabled explaining why', async () => {
      const mockView = createMockView('maturity');
      renderCapabilityDetails(mockView);

      await waitFor(() => {
        const colorPickerButton = screen.getByTestId('color-picker-button');
        fireEvent.mouseOver(colorPickerButton);
      });

      expect(screen.getByText('Switch to custom color scheme to assign colors')).toBeInTheDocument();
    });
  });

  describe('Displaying current color', () => {
    it('should show current custom color if set', async () => {
      const mockView = createMockView('custom', '#FF5733');
      renderCapabilityDetails(mockView);

      await waitFor(() => {
        const colorDisplay = screen.getByTestId('color-picker-display');
        expect(colorDisplay).toHaveStyle({ backgroundColor: '#FF5733' });
      });
    });

    it('should show default color if custom color not set', async () => {
      const mockView = createMockView('custom');
      renderCapabilityDetails(mockView);

      await waitFor(() => {
        const colorDisplay = screen.getByTestId('color-picker-display');
        expect(colorDisplay).toHaveStyle({ backgroundColor: '#E0E0E0' });
      });
    });

    it('should show color value text', async () => {
      const mockView = createMockView('custom', '#FF5733');
      renderCapabilityDetails(mockView);

      await waitFor(() => {
        expect(screen.getByText('#FF5733')).toBeInTheDocument();
      });
    });
  });

  describe('API calls on color selection', () => {
    it('should call API to update capability color when color selected', async () => {
      let capturedColor: string | null = null;
      server.use(
        http.patch(`${API_BASE}/api/v1/views/:viewId/capabilities/:capabilityId/color`, async ({ request }) => {
          const body = await request.json() as { color: string };
          capturedColor = body.color;
          return new HttpResponse(null, { status: 204 });
        })
      );

      const mockView = createMockView('custom');
      renderCapabilityDetails(mockView);

      await waitFor(() => {
        expect(screen.getByTestId('color-picker-button')).toBeInTheDocument();
      });

      const colorPickerButton = screen.getByTestId('color-picker-button');
      fireEvent.click(colorPickerButton);

      const colorInput = screen.getByTestId('color-picker-input');
      fireEvent.change(colorInput, { target: { value: '#00FF00' } });

      await waitFor(() => {
        expect(capturedColor).toBe('#00FF00');
      });
    });

    it('should call API with correct view ID and capability ID', async () => {
      let capturedParams: { viewId?: string; capabilityId?: string } = {};
      server.use(
        http.patch(`${API_BASE}/api/v1/views/:viewId/capabilities/:capabilityId/color`, ({ params }) => {
          capturedParams = { viewId: params.viewId as string, capabilityId: params.capabilityId as string };
          return new HttpResponse(null, { status: 204 });
        })
      );

      const mockView = createMockView('custom');
      renderCapabilityDetails(mockView);

      await waitFor(() => {
        expect(screen.getByTestId('color-picker-button')).toBeInTheDocument();
      });

      const colorPickerButton = screen.getByTestId('color-picker-button');
      fireEvent.click(colorPickerButton);

      const colorInput = screen.getByTestId('color-picker-input');
      fireEvent.change(colorInput, { target: { value: '#AABBCC' } });

      await waitFor(() => {
        expect(capturedParams.viewId).toBe('view-1');
        expect(capturedParams.capabilityId).toBe('cap-1');
      });
    });
  });

  describe('Error handling', () => {
    it('should call error handler when API call fails', async () => {
      let apiCalled = false;
      server.use(
        http.patch(`${API_BASE}/api/v1/views/:viewId/capabilities/:capabilityId/color`, () => {
          apiCalled = true;
          return HttpResponse.json({ error: 'Failed to update color' }, { status: 500 });
        })
      );

      const mockView = createMockView('custom', '#FF5733');
      renderCapabilityDetails(mockView);

      await waitFor(() => {
        expect(screen.getByTestId('color-picker-button')).toBeInTheDocument();
      });

      const colorPickerButton = screen.getByTestId('color-picker-button');
      fireEvent.click(colorPickerButton);

      const colorInput = screen.getByTestId('color-picker-input');
      fireEvent.change(colorInput, { target: { value: '#00FF00' } });

      await waitFor(() => {
        expect(apiCalled).toBe(true);
      });
    });
  });

  describe('Color picker in different contexts', () => {
    it('should not render color picker when no view is selected', async () => {
      renderCapabilityDetails(createMockStore(null));

      await waitFor(() => {
        expect(screen.queryByTestId('color-picker')).not.toBeInTheDocument();
      });
    });

    it('should not render color picker when capability not in current view', async () => {
      const mockView: View = {
        id: 'view-1' as ViewId,
        name: 'Test View',
        isDefault: true,
        components: [],
        capabilities: [],
        colorScheme: 'custom',
        createdAt: '2024-01-01T00:00:00Z',
        _links: { self: { href: '/api/v1/views/view-1', method: 'GET' as const } },
      };
      renderCapabilityDetails(mockView);

      await waitFor(() => {
        expect(screen.queryByTestId('color-picker')).not.toBeInTheDocument();
      });
    });
  });
});
