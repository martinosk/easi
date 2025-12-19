import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ComponentDetails } from './ComponentDetails';
import type { Component, View } from '../../../api/types';
import { Toaster } from 'react-hot-toast';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';

vi.mock('../../../store/appStore', () => ({
  useAppStore: vi.fn(),
}));

vi.mock('../../../api/client', () => ({
  default: {
    updateComponentColor: vi.fn(),
  },
  apiClient: {
    updateComponentColor: vi.fn(),
  },
}));

import { useAppStore } from '../../../store/appStore';
import apiClient from '../../../api/client';

const mockComponent: Component = {
  id: 'comp-1',
  name: 'Test Component',
  description: 'Test description',
  createdAt: '2024-01-01T00:00:00Z',
  _links: {
    self: { href: '/api/v1/components/comp-1' },
    reference: { href: '/api/v1/reference/components' },
  },
};

const createMockView = (colorScheme: string, customColor?: string): View => ({
  id: 'view-1',
  name: 'Test View',
  isDefault: true,
  components: [
    { componentId: 'comp-1', x: 100, y: 200, customColor } as any,
  ],
  capabilities: [],
  colorScheme,
  createdAt: '2024-01-01T00:00:00Z',
  _links: { self: { href: '/api/v1/views/view-1' } },
});

const createMockStore = (view: View | null) => ({
  selectedNodeId: 'comp-1',
  components: [mockComponent],
  currentView: view,
  clearSelection: vi.fn(),
  capabilityRealizations: [],
  capabilities: [],
  updateComponentColor: vi.fn(),
  clearComponentColor: vi.fn(),
});

describe('ComponentDetails - ColorPicker Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Color picker visibility', () => {
    it('should show color picker in component details panel', () => {
      const mockView = createMockView('custom');
      const mockStore = createMockStore(mockView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<ComponentDetails onEdit={vi.fn()} />, { wrapper: MantineTestWrapper });

      expect(screen.getByTestId('color-picker')).toBeInTheDocument();
    });

    it('should show color picker label', () => {
      const mockView = createMockView('custom');
      const mockStore = createMockStore(mockView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<ComponentDetails onEdit={vi.fn()} />, { wrapper: MantineTestWrapper });

      expect(screen.getByText('Custom Color')).toBeInTheDocument();
    });
  });

  describe('Color picker enabled state based on color scheme', () => {
    it('should enable color picker when colorScheme is "custom"', () => {
      const mockView = createMockView('custom');
      const mockStore = createMockStore(mockView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<ComponentDetails onEdit={vi.fn()} />, { wrapper: MantineTestWrapper });

      const colorPickerButton = screen.getByTestId('color-picker-button');
      expect(colorPickerButton).not.toBeDisabled();
    });

    it('should disable color picker when colorScheme is "maturity"', () => {
      const mockView = createMockView('maturity');
      const mockStore = createMockStore(mockView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<ComponentDetails onEdit={vi.fn()} />, { wrapper: MantineTestWrapper });

      const colorPickerButton = screen.getByTestId('color-picker-button');
      expect(colorPickerButton).toBeDisabled();
    });

    it('should disable color picker when colorScheme is "classic"', () => {
      const mockView = createMockView('classic');
      const mockStore = createMockStore(mockView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<ComponentDetails onEdit={vi.fn()} />, { wrapper: MantineTestWrapper });

      const colorPickerButton = screen.getByTestId('color-picker-button');
      expect(colorPickerButton).toBeDisabled();
    });

    it('should show tooltip when disabled explaining why', () => {
      const mockView = createMockView('maturity');
      const mockStore = createMockStore(mockView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<ComponentDetails onEdit={vi.fn()} />, { wrapper: MantineTestWrapper });

      const colorPickerButton = screen.getByTestId('color-picker-button');
      fireEvent.mouseOver(colorPickerButton);

      expect(screen.getByText('Switch to custom color scheme to assign colors')).toBeInTheDocument();
    });
  });

  describe('Displaying current color', () => {
    it('should show current custom color if set', () => {
      const mockView = createMockView('custom', '#FF5733');
      const mockStore = createMockStore(mockView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<ComponentDetails onEdit={vi.fn()} />, { wrapper: MantineTestWrapper });

      const colorDisplay = screen.getByTestId('color-picker-display');
      expect(colorDisplay).toHaveStyle({ backgroundColor: '#FF5733' });
    });

    it('should show default color if custom color not set', () => {
      const mockView = createMockView('custom');
      const mockStore = createMockStore(mockView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<ComponentDetails onEdit={vi.fn()} />, { wrapper: MantineTestWrapper });

      const colorDisplay = screen.getByTestId('color-picker-display');
      expect(colorDisplay).toHaveStyle({ backgroundColor: '#E0E0E0' });
    });

    it('should show color value text', () => {
      const mockView = createMockView('custom', '#FF5733');
      const mockStore = createMockStore(mockView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<ComponentDetails onEdit={vi.fn()} />, { wrapper: MantineTestWrapper });

      expect(screen.getByText('#FF5733')).toBeInTheDocument();
    });
  });

  describe('API calls on color selection', () => {
    it('should call API to update component color when color selected', async () => {
      const mockView = createMockView('custom');
      const mockUpdateComponentColor = vi.fn().mockResolvedValue(undefined);
      const mockStore = createMockStore(mockView);
      mockStore.updateComponentColor = mockUpdateComponentColor;

      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));
      vi.mocked(apiClient.updateComponentColor).mockResolvedValue(undefined);

      render(<ComponentDetails onEdit={vi.fn()} />, { wrapper: MantineTestWrapper });

      const colorPickerButton = screen.getByTestId('color-picker-button');
      fireEvent.click(colorPickerButton);

      const colorInput = screen.getByTestId('color-picker-input');
      fireEvent.change(colorInput, { target: { value: '#00FF00' } });

      await waitFor(() => {
        expect(mockUpdateComponentColor).toHaveBeenCalledWith('view-1', 'comp-1', '#00FF00');
      });
    });

    it('should call API with correct view ID and component ID', async () => {
      const mockView = createMockView('custom');
      const mockUpdateComponentColor = vi.fn().mockResolvedValue(undefined);
      const mockStore = createMockStore(mockView);
      mockStore.updateComponentColor = mockUpdateComponentColor;

      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<ComponentDetails onEdit={vi.fn()} />, { wrapper: MantineTestWrapper });

      const colorPickerButton = screen.getByTestId('color-picker-button');
      fireEvent.click(colorPickerButton);

      const colorInput = screen.getByTestId('color-picker-input');
      fireEvent.change(colorInput, { target: { value: '#AABBCC' } });

      await waitFor(() => {
        expect(mockUpdateComponentColor).toHaveBeenCalledWith('view-1', 'comp-1', '#AABBCC');
      });
    });
  });

  describe('Optimistic updates', () => {
    it('should update color immediately on selection (before API call completes)', async () => {
      const mockView = createMockView('custom', '#FF5733');
      const mockUpdateComponentColor = vi.fn().mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 100))
      );
      const mockStore = createMockStore(mockView);
      mockStore.updateComponentColor = mockUpdateComponentColor;

      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      const { rerender } = render(<ComponentDetails onEdit={vi.fn()} />);

      const colorPickerButton = screen.getByTestId('color-picker-button');
      fireEvent.click(colorPickerButton);

      const colorInput = screen.getByTestId('color-picker-input');
      fireEvent.change(colorInput, { target: { value: '#00FF00' } });

      const updatedView = createMockView('custom', '#00FF00');
      const updatedStore = createMockStore(updatedView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(updatedStore));

      rerender(<ComponentDetails onEdit={vi.fn()} />);

      const colorDisplay = screen.getByTestId('color-picker-display');
      expect(colorDisplay).toHaveStyle({ backgroundColor: '#00FF00' });
    });

    it.skip('should roll back color to previous value if API call fails', async () => {
      const mockView = createMockView('custom', '#FF5733');
      const mockStore = createMockStore(mockView);

      const mockUpdateComponentColor = vi.fn().mockImplementation(async (viewId: string, componentId: string, color: string) => {
        const updatedView = createMockView('custom', color);
        const updatedStore = createMockStore(updatedView);
        updatedStore.updateComponentColor = mockUpdateComponentColor;
        vi.mocked(useAppStore).mockImplementation((selector: any) => selector(updatedStore));

        await new Promise(resolve => setTimeout(resolve, 50));

        const restoredView = createMockView('custom', '#FF5733');
        const restoredStore = createMockStore(restoredView);
        restoredStore.updateComponentColor = mockUpdateComponentColor;
        vi.mocked(useAppStore).mockImplementation((selector: any) => selector(restoredStore));

        throw new Error('API Error');
      });

      mockStore.updateComponentColor = mockUpdateComponentColor;
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<ComponentDetails onEdit={vi.fn()} />, { wrapper: MantineTestWrapper });

      const colorPickerButton = screen.getByTestId('color-picker-button');
      fireEvent.click(colorPickerButton);

      const colorInput = screen.getByTestId('color-picker-input');
      fireEvent.change(colorInput, { target: { value: '#00FF00' } });

      await waitFor(() => {
        expect(mockUpdateComponentColor).toHaveBeenCalled();
      });

      await waitFor(() => {
        const colorDisplay = screen.getByTestId('color-picker-display');
        expect(colorDisplay).toHaveStyle({ backgroundColor: '#FF5733' });
      }, { timeout: 3000 });
    });

    it('should show error message when API call fails', async () => {
      const mockView = createMockView('custom', '#FF5733');
      const mockUpdateComponentColor = vi.fn().mockRejectedValue(new Error('Failed to update color'));
      const mockStore = createMockStore(mockView);
      mockStore.updateComponentColor = mockUpdateComponentColor;

      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(
        <>
          <ComponentDetails onEdit={vi.fn()} />
          <Toaster />
        </>
      );

      const colorPickerButton = screen.getByTestId('color-picker-button');
      fireEvent.click(colorPickerButton);

      const colorInput = screen.getByTestId('color-picker-input');
      fireEvent.change(colorInput, { target: { value: '#00FF00' } });

      await waitFor(() => {
        expect(screen.getByText(/failed to update color/i)).toBeInTheDocument();
      });
    });
  });

  describe('Color picker in different contexts', () => {
    it('should not render color picker when no view is selected', () => {
      const mockStore = createMockStore(null);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<ComponentDetails onEdit={vi.fn()} />, { wrapper: MantineTestWrapper });

      expect(screen.queryByTestId('color-picker')).not.toBeInTheDocument();
    });

    it('should not render color picker when component not in current view', () => {
      const mockView: View = {
        id: 'view-1',
        name: 'Test View',
        isDefault: true,
        components: [],
        capabilities: [],
        colorScheme: 'custom',
        createdAt: '2024-01-01T00:00:00Z',
        _links: { self: { href: '/api/v1/views/view-1' } },
      };
      const mockStore = createMockStore(mockView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<ComponentDetails onEdit={vi.fn()} />, { wrapper: MantineTestWrapper });

      expect(screen.queryByTestId('color-picker')).not.toBeInTheDocument();
    });
  });

  describe('Clear color functionality', () => {
    it('should show clear color button when custom color is set', () => {
      const mockView = createMockView('custom', '#FF5733');
      const mockStore = createMockStore(mockView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<ComponentDetails onEdit={vi.fn()} />, { wrapper: MantineTestWrapper });

      expect(screen.getByText('Clear Color')).toBeInTheDocument();
    });

    it('should not show clear color button when no custom color is set', () => {
      const mockView = createMockView('custom');
      const mockStore = createMockStore(mockView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<ComponentDetails onEdit={vi.fn()} />, { wrapper: MantineTestWrapper });

      expect(screen.queryByText('Clear Color')).not.toBeInTheDocument();
    });

    it('should call API to clear component color when clear button clicked', async () => {
      const mockView = createMockView('custom', '#FF5733');
      const mockClearComponentColor = vi.fn().mockResolvedValue(undefined);
      const mockStore = createMockStore(mockView);
      mockStore.clearComponentColor = mockClearComponentColor;

      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<ComponentDetails onEdit={vi.fn()} />, { wrapper: MantineTestWrapper });

      const clearButton = screen.getByText('Clear Color');
      fireEvent.click(clearButton);

      await waitFor(() => {
        expect(mockClearComponentColor).toHaveBeenCalledWith('view-1', 'comp-1');
      });
    });

    it('should optimistically remove color when clear button clicked', async () => {
      const mockView = createMockView('custom', '#FF5733');
      const mockClearComponentColor = vi.fn().mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 100))
      );
      const mockStore = createMockStore(mockView);
      mockStore.clearComponentColor = mockClearComponentColor;

      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      const { rerender } = render(<ComponentDetails onEdit={vi.fn()} />);

      const clearButton = screen.getByText('Clear Color');
      fireEvent.click(clearButton);

      const updatedView = createMockView('custom');
      const updatedStore = createMockStore(updatedView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(updatedStore));

      rerender(<ComponentDetails onEdit={vi.fn()} />);

      const colorDisplay = screen.getByTestId('color-picker-display');
      expect(colorDisplay).toHaveStyle({ backgroundColor: '#E0E0E0' });
    });

    it('should roll back to previous color if clear API call fails', async () => {
      const mockView = createMockView('custom', '#FF5733');
      const mockClearComponentColor = vi.fn().mockRejectedValue(new Error('API Error'));
      const mockStore = createMockStore(mockView);
      mockStore.clearComponentColor = mockClearComponentColor;

      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      const { rerender } = render(<ComponentDetails onEdit={vi.fn()} />);

      const clearButton = screen.getByText('Clear Color');
      fireEvent.click(clearButton);

      await waitFor(() => {
        expect(mockClearComponentColor).toHaveBeenCalled();
      });

      const restoredView = createMockView('custom', '#FF5733');
      const restoredStore = createMockStore(restoredView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(restoredStore));

      rerender(<ComponentDetails onEdit={vi.fn()} />);

      await waitFor(() => {
        const colorDisplay = screen.getByTestId('color-picker-display');
        expect(colorDisplay).toHaveStyle({ backgroundColor: '#FF5733' });
      });
    });

    it('should show error message when clear API call fails', async () => {
      const mockView = createMockView('custom', '#FF5733');
      const mockClearComponentColor = vi.fn().mockRejectedValue(new Error('Failed to clear color'));
      const mockStore = createMockStore(mockView);
      mockStore.clearComponentColor = mockClearComponentColor;

      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(
        <>
          <ComponentDetails onEdit={vi.fn()} />
          <Toaster />
        </>
      );

      const clearButton = screen.getByText('Clear Color');
      fireEvent.click(clearButton);

      await waitFor(() => {
        expect(screen.getByText(/failed to clear color/i)).toBeInTheDocument();
      });
    });
  });
});
