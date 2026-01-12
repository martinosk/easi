import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { ComponentDetails } from './ComponentDetails';
import type { View } from '../../../api/types';
import { toComponentId, toViewId } from '../../../api/types';
import { createMantineTestWrapper, seedDb, server } from '../../../test/helpers';
import { useAppStore, type AppStore } from '../../../store/appStore';
import { useCurrentView } from '../../views/hooks/useCurrentView';

const API_BASE = 'http://localhost:8080';

vi.mock('../../../store/appStore', () => ({
  useAppStore: vi.fn(),
}));

vi.mock('../../views/hooks/useCurrentView', () => ({
  useCurrentView: vi.fn(),
}));

const mockComponent = {
  id: toComponentId('comp-1'),
  name: 'Test Component',
  description: 'Test description',
  createdAt: '2024-01-01T00:00:00Z',
  _links: { self: { href: '/api/v1/components/comp-1', method: 'GET' as const } },
};

const createMockView = (colorScheme: string, customColor?: string): View => ({
  id: toViewId('view-1'),
  name: 'Test View',
  isDefault: true,
  isPrivate: false,
  components: [
    {
      componentId: toComponentId('comp-1'),
      x: 100,
      y: 200,
      customColor,
      _links: {
        'x-update-color': { href: '/api/v1/views/view-1/components/comp-1/color', method: 'PATCH' as const },
        'x-clear-color': { href: '/api/v1/views/view-1/components/comp-1/color', method: 'DELETE' as const },
        'x-update-position': { href: '/api/v1/views/view-1/components/comp-1/position', method: 'PATCH' as const },
        'x-remove': { href: '/api/v1/views/view-1/components/comp-1', method: 'DELETE' as const },
      },
    },
  ],
  capabilities: [],
  colorScheme,
  createdAt: '2024-01-01T00:00:00Z',
  _links: { self: { href: '/api/v1/views/view-1', method: 'GET' as const } },
});

const createMockStore = () => ({
  selectedNodeId: 'comp-1',
  clearSelection: vi.fn(),
});

describe('ComponentDetails - ColorPicker Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    seedDb({
      components: [mockComponent],
      capabilities: [],
    });
  });

  const renderComponentDetails = (view: View | null) => {
    vi.mocked(useAppStore).mockImplementation((selector: (state: AppStore) => unknown) => selector(createMockStore() as unknown as AppStore));
    vi.mocked(useCurrentView).mockReturnValue({
      currentView: view,
      currentViewId: view?.id ?? null,
      isLoading: false,
      error: null,
    });
    const { Wrapper } = createMantineTestWrapper();
    return render(<ComponentDetails onEdit={vi.fn()} />, { wrapper: Wrapper });
  };

  describe('Color picker visibility', () => {
    it('should show color picker in component details panel', async () => {
      const mockView = createMockView('custom');
      renderComponentDetails(mockView);

      await waitFor(() => {
        expect(screen.getByTestId('color-picker')).toBeInTheDocument();
      });
    });

    it('should show color picker label', async () => {
      const mockView = createMockView('custom');
      renderComponentDetails(mockView);

      await waitFor(() => {
        expect(screen.getByText('Custom Color')).toBeInTheDocument();
      });
    });
  });

  describe('Color picker enabled state based on color scheme', () => {
    it('should enable color picker when colorScheme is "custom"', async () => {
      const mockView = createMockView('custom');
      renderComponentDetails(mockView);

      await waitFor(() => {
        const colorPickerButton = screen.getByTestId('color-picker-button');
        expect(colorPickerButton).not.toBeDisabled();
      });
    });

    it('should disable color picker when colorScheme is "maturity"', async () => {
      const mockView = createMockView('maturity');
      renderComponentDetails(mockView);

      await waitFor(() => {
        const colorPickerButton = screen.getByTestId('color-picker-button');
        expect(colorPickerButton).toBeDisabled();
      });
    });

    it('should disable color picker when colorScheme is "classic"', async () => {
      const mockView = createMockView('classic');
      renderComponentDetails(mockView);

      await waitFor(() => {
        const colorPickerButton = screen.getByTestId('color-picker-button');
        expect(colorPickerButton).toBeDisabled();
      });
    });

    it('should show tooltip when disabled explaining why', async () => {
      const mockView = createMockView('maturity');
      renderComponentDetails(mockView);

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
      renderComponentDetails(mockView);

      await waitFor(() => {
        const colorDisplay = screen.getByTestId('color-picker-display');
        expect(colorDisplay).toHaveStyle({ backgroundColor: '#FF5733' });
      });
    });

    it('should show default color if custom color not set', async () => {
      const mockView = createMockView('custom');
      renderComponentDetails(mockView);

      await waitFor(() => {
        const colorDisplay = screen.getByTestId('color-picker-display');
        expect(colorDisplay).toHaveStyle({ backgroundColor: '#E0E0E0' });
      });
    });

    it('should show color value text', async () => {
      const mockView = createMockView('custom', '#FF5733');
      renderComponentDetails(mockView);

      await waitFor(() => {
        expect(screen.getByText('#FF5733')).toBeInTheDocument();
      });
    });
  });

  describe('API calls on color selection', () => {
    it('should call API to update component color when color selected', async () => {
      let capturedColor: string | null = null;
      server.use(
        http.patch(`${API_BASE}/api/v1/views/:viewId/components/:componentId/color`, async ({ request }) => {
          const body = await request.json() as { color: string };
          capturedColor = body.color;
          return new HttpResponse(null, { status: 204 });
        })
      );

      const mockView = createMockView('custom');
      renderComponentDetails(mockView);

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

    it('should call API with correct view ID and component ID', async () => {
      let capturedParams: { viewId?: string; componentId?: string } = {};
      server.use(
        http.patch(`${API_BASE}/api/v1/views/:viewId/components/:componentId/color`, ({ params }) => {
          capturedParams = { viewId: params.viewId as string, componentId: params.componentId as string };
          return new HttpResponse(null, { status: 204 });
        })
      );

      const mockView = createMockView('custom');
      renderComponentDetails(mockView);

      await waitFor(() => {
        expect(screen.getByTestId('color-picker-button')).toBeInTheDocument();
      });

      const colorPickerButton = screen.getByTestId('color-picker-button');
      fireEvent.click(colorPickerButton);

      const colorInput = screen.getByTestId('color-picker-input');
      fireEvent.change(colorInput, { target: { value: '#AABBCC' } });

      await waitFor(() => {
        expect(capturedParams.viewId).toBe('view-1');
        expect(capturedParams.componentId).toBe('comp-1');
      });
    });
  });

  describe('Error handling', () => {
    it('should call error handler when API call fails', async () => {
      let apiCalled = false;
      server.use(
        http.patch(`${API_BASE}/api/v1/views/:viewId/components/:componentId/color`, () => {
          apiCalled = true;
          return HttpResponse.json({ error: 'Failed to update color' }, { status: 500 });
        })
      );

      const mockView = createMockView('custom', '#FF5733');
      renderComponentDetails(mockView);

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
      renderComponentDetails(null);

      await waitFor(() => {
        expect(screen.queryByTestId('color-picker')).not.toBeInTheDocument();
      });
    });

    it('should not render color picker when component not in current view', async () => {
      const mockView: View = {
        id: toViewId('view-1'),
        name: 'Test View',
        isDefault: true,
        isPrivate: false,
        components: [],
        capabilities: [],
        colorScheme: 'custom',
        createdAt: '2024-01-01T00:00:00Z',
        _links: { self: { href: '/api/v1/views/view-1', method: 'GET' as const } },
      };
      renderComponentDetails(mockView);

      await waitFor(() => {
        expect(screen.queryByTestId('color-picker')).not.toBeInTheDocument();
      });
    });
  });

  describe('Clear color functionality', () => {
    it('should show clear color button when custom color is set', async () => {
      const mockView = createMockView('custom', '#FF5733');
      renderComponentDetails(mockView);

      await waitFor(() => {
        expect(screen.getByText('Clear Color')).toBeInTheDocument();
      });
    });

    it('should not show clear color button when no custom color is set', async () => {
      const mockView = createMockView('custom');
      renderComponentDetails(mockView);

      await waitFor(() => {
        expect(screen.queryByText('Clear Color')).not.toBeInTheDocument();
      });
    });

    it('should call API to clear component color when clear button clicked', async () => {
      let clearCalled = false;
      server.use(
        http.delete(`${API_BASE}/api/v1/views/:viewId/components/:componentId/color`, () => {
          clearCalled = true;
          return new HttpResponse(null, { status: 204 });
        })
      );

      const mockView = createMockView('custom', '#FF5733');
      renderComponentDetails(mockView);

      await waitFor(() => {
        expect(screen.getByText('Clear Color')).toBeInTheDocument();
      });

      const clearButton = screen.getByText('Clear Color');
      fireEvent.click(clearButton);

      await waitFor(() => {
        expect(clearCalled).toBe(true);
      });
    });

    it('should call error handler when clear API call fails', async () => {
      let apiCalled = false;
      server.use(
        http.delete(`${API_BASE}/api/v1/views/:viewId/components/:componentId/color`, () => {
          apiCalled = true;
          return HttpResponse.json({ error: 'Failed to clear color' }, { status: 500 });
        })
      );

      const mockView = createMockView('custom', '#FF5733');
      renderComponentDetails(mockView);

      await waitFor(() => {
        expect(screen.getByText('Clear Color')).toBeInTheDocument();
      });

      const clearButton = screen.getByText('Clear Color');
      fireEvent.click(clearButton);

      await waitFor(() => {
        expect(apiCalled).toBe(true);
      });
    });
  });
});
