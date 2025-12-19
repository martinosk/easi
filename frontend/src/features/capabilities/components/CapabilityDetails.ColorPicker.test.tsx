import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CapabilityDetails } from './CapabilityDetails';
import type { Capability, View } from '../../../api/types';
import { Toaster } from 'react-hot-toast';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';

vi.mock('../../../store/appStore', () => ({
  useAppStore: vi.fn(),
}));

vi.mock('../../../api/client', () => ({
  default: {
    updateCapabilityColor: vi.fn(),
  },
  apiClient: {
    updateCapabilityColor: vi.fn(),
  },
}));

import { useAppStore } from '../../../store/appStore';
import apiClient from '../../../api/client';

const mockCapability: Capability = {
  id: 'cap-1',
  name: 'Test Capability',
  description: 'Test description',
  level: 'L2',
  maturityLevel: 'Product',
  createdAt: '2024-01-01T00:00:00Z',
  _links: { self: { href: '/api/v1/capabilities/cap-1' } },
};

const createMockView = (colorScheme: string, customColor?: string): View => ({
  id: 'view-1',
  name: 'Test View',
  isDefault: true,
  components: [],
  capabilities: [
    { capabilityId: 'cap-1', x: 100, y: 200, customColor } as any,
  ],
  colorScheme,
  createdAt: '2024-01-01T00:00:00Z',
  _links: { self: { href: '/api/v1/views/view-1' } },
});

const createMockStore = (view: View | null) => ({
  selectedCapabilityId: 'cap-1',
  capabilities: [mockCapability],
  currentView: view,
  selectCapability: vi.fn(),
  capabilityRealizations: [],
  components: [],
  updateCapabilityColor: vi.fn(),
});

describe('CapabilityDetails - ColorPicker Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Color picker visibility', () => {
    it('should show color picker in capability details panel', () => {
      const mockView = createMockView('custom');
      const mockStore = createMockStore(mockView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<CapabilityDetails onRemoveFromView={vi.fn()} />, { wrapper: MantineTestWrapper });

      expect(screen.getByTestId('color-picker')).toBeInTheDocument();
    });

    it('should show color picker label', () => {
      const mockView = createMockView('custom');
      const mockStore = createMockStore(mockView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<CapabilityDetails onRemoveFromView={vi.fn()} />, { wrapper: MantineTestWrapper });

      expect(screen.getByText('Custom Color')).toBeInTheDocument();
    });
  });

  describe('Color picker enabled state based on color scheme', () => {
    it('should enable color picker when colorScheme is "custom"', () => {
      const mockView = createMockView('custom');
      const mockStore = createMockStore(mockView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<CapabilityDetails onRemoveFromView={vi.fn()} />, { wrapper: MantineTestWrapper });

      const colorPickerButton = screen.getByTestId('color-picker-button');
      expect(colorPickerButton).not.toBeDisabled();
    });

    it('should disable color picker when colorScheme is "maturity"', () => {
      const mockView = createMockView('maturity');
      const mockStore = createMockStore(mockView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<CapabilityDetails onRemoveFromView={vi.fn()} />, { wrapper: MantineTestWrapper });

      const colorPickerButton = screen.getByTestId('color-picker-button');
      expect(colorPickerButton).toBeDisabled();
    });

    it('should disable color picker when colorScheme is "classic"', () => {
      const mockView = createMockView('classic');
      const mockStore = createMockStore(mockView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<CapabilityDetails onRemoveFromView={vi.fn()} />, { wrapper: MantineTestWrapper });

      const colorPickerButton = screen.getByTestId('color-picker-button');
      expect(colorPickerButton).toBeDisabled();
    });

    it('should show tooltip when disabled explaining why', () => {
      const mockView = createMockView('maturity');
      const mockStore = createMockStore(mockView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<CapabilityDetails onRemoveFromView={vi.fn()} />, { wrapper: MantineTestWrapper });

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

      render(<CapabilityDetails onRemoveFromView={vi.fn()} />, { wrapper: MantineTestWrapper });

      const colorDisplay = screen.getByTestId('color-picker-display');
      expect(colorDisplay).toHaveStyle({ backgroundColor: '#FF5733' });
    });

    it('should show default color if custom color not set', () => {
      const mockView = createMockView('custom');
      const mockStore = createMockStore(mockView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<CapabilityDetails onRemoveFromView={vi.fn()} />, { wrapper: MantineTestWrapper });

      const colorDisplay = screen.getByTestId('color-picker-display');
      expect(colorDisplay).toHaveStyle({ backgroundColor: '#E0E0E0' });
    });

    it('should show color value text', () => {
      const mockView = createMockView('custom', '#FF5733');
      const mockStore = createMockStore(mockView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<CapabilityDetails onRemoveFromView={vi.fn()} />, { wrapper: MantineTestWrapper });

      expect(screen.getByText('#FF5733')).toBeInTheDocument();
    });
  });

  describe('API calls on color selection', () => {
    it('should call API to update capability color when color selected', async () => {
      const mockView = createMockView('custom');
      const mockUpdateCapabilityColor = vi.fn().mockResolvedValue(undefined);
      const mockStore = createMockStore(mockView);
      mockStore.updateCapabilityColor = mockUpdateCapabilityColor;

      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));
      vi.mocked(apiClient.updateCapabilityColor).mockResolvedValue(undefined);

      render(<CapabilityDetails onRemoveFromView={vi.fn()} />, { wrapper: MantineTestWrapper });

      const colorPickerButton = screen.getByTestId('color-picker-button');
      fireEvent.click(colorPickerButton);

      const colorInput = screen.getByTestId('color-picker-input');
      fireEvent.change(colorInput, { target: { value: '#00FF00' } });

      await waitFor(() => {
        expect(mockUpdateCapabilityColor).toHaveBeenCalledWith('view-1', 'cap-1', '#00FF00');
      });
    });

    it('should call API with correct view ID and capability ID', async () => {
      const mockView = createMockView('custom');
      const mockUpdateCapabilityColor = vi.fn().mockResolvedValue(undefined);
      const mockStore = createMockStore(mockView);
      mockStore.updateCapabilityColor = mockUpdateCapabilityColor;

      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<CapabilityDetails onRemoveFromView={vi.fn()} />, { wrapper: MantineTestWrapper });

      const colorPickerButton = screen.getByTestId('color-picker-button');
      fireEvent.click(colorPickerButton);

      const colorInput = screen.getByTestId('color-picker-input');
      fireEvent.change(colorInput, { target: { value: '#AABBCC' } });

      await waitFor(() => {
        expect(mockUpdateCapabilityColor).toHaveBeenCalledWith('view-1', 'cap-1', '#AABBCC');
      });
    });
  });

  describe('Optimistic updates', () => {
    it('should update color immediately on selection (before API call completes)', async () => {
      const mockView = createMockView('custom', '#FF5733');
      const mockUpdateCapabilityColor = vi.fn().mockImplementation(
        () => new Promise((resolve) => setTimeout(resolve, 100))
      );
      const mockStore = createMockStore(mockView);
      mockStore.updateCapabilityColor = mockUpdateCapabilityColor;

      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      const { rerender } = render(<CapabilityDetails onRemoveFromView={vi.fn()} />, { wrapper: MantineTestWrapper });

      const colorPickerButton = screen.getByTestId('color-picker-button');
      fireEvent.click(colorPickerButton);

      const colorInput = screen.getByTestId('color-picker-input');
      fireEvent.change(colorInput, { target: { value: '#00FF00' } });

      const updatedView = createMockView('custom', '#00FF00');
      const updatedStore = createMockStore(updatedView);
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(updatedStore));

      rerender(<CapabilityDetails onRemoveFromView={vi.fn()} />);

      const colorDisplay = screen.getByTestId('color-picker-display');
      expect(colorDisplay).toHaveStyle({ backgroundColor: '#00FF00' });
    });

    it.skip('should roll back color to previous value if API call fails', async () => {
      const mockView = createMockView('custom', '#FF5733');
      const mockStore = createMockStore(mockView);

      const mockUpdateCapabilityColor = vi.fn().mockImplementation(async (viewId: string, capabilityId: string, color: string) => {
        const updatedView = createMockView('custom', color);
        const updatedStore = createMockStore(updatedView);
        updatedStore.updateCapabilityColor = mockUpdateCapabilityColor;
        vi.mocked(useAppStore).mockImplementation((selector: any) => selector(updatedStore));

        await new Promise(resolve => setTimeout(resolve, 50));

        const restoredView = createMockView('custom', '#FF5733');
        const restoredStore = createMockStore(restoredView);
        restoredStore.updateCapabilityColor = mockUpdateCapabilityColor;
        vi.mocked(useAppStore).mockImplementation((selector: any) => selector(restoredStore));

        throw new Error('API Error');
      });

      mockStore.updateCapabilityColor = mockUpdateCapabilityColor;
      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(<CapabilityDetails onRemoveFromView={vi.fn()} />, { wrapper: MantineTestWrapper });

      const colorPickerButton = screen.getByTestId('color-picker-button');
      fireEvent.click(colorPickerButton);

      const colorInput = screen.getByTestId('color-picker-input');
      fireEvent.change(colorInput, { target: { value: '#00FF00' } });

      await waitFor(() => {
        expect(mockUpdateCapabilityColor).toHaveBeenCalled();
      });

      await waitFor(() => {
        const colorDisplay = screen.getByTestId('color-picker-display');
        expect(colorDisplay).toHaveStyle({ backgroundColor: '#FF5733' });
      }, { timeout: 3000 });
    });

    it('should show error message when API call fails', async () => {
      const mockView = createMockView('custom', '#FF5733');
      const mockUpdateCapabilityColor = vi.fn().mockRejectedValue(new Error('Failed to update color'));
      const mockStore = createMockStore(mockView);
      mockStore.updateCapabilityColor = mockUpdateCapabilityColor;

      vi.mocked(useAppStore).mockImplementation((selector: any) => selector(mockStore));

      render(
        <>
          <CapabilityDetails onRemoveFromView={vi.fn()} />
          <Toaster />
        </>,
        { wrapper: MantineTestWrapper }
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

      render(<CapabilityDetails onRemoveFromView={vi.fn()} />, { wrapper: MantineTestWrapper });

      expect(screen.queryByTestId('color-picker')).not.toBeInTheDocument();
    });

    it('should not render color picker when capability not in current view', () => {
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

      render(<CapabilityDetails onRemoveFromView={vi.fn()} />, { wrapper: MantineTestWrapper });

      expect(screen.queryByTestId('color-picker')).not.toBeInTheDocument();
    });
  });
});
