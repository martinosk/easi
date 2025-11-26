import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ColorPicker } from './ColorPicker';

describe('ColorPicker', () => {
  const mockOnChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('Rendering', () => {
    it('should render with initial color value', () => {
      render(
        <ColorPicker
          color="#FF5733"
          onChange={mockOnChange}
          disabled={false}
        />
      );

      const colorDisplay = screen.getByTestId('color-picker-display');
      expect(colorDisplay).toHaveStyle({ backgroundColor: '#FF5733' });
    });

    it('should render with default color when no color provided', () => {
      render(
        <ColorPicker
          color={null}
          onChange={mockOnChange}
          disabled={false}
        />
      );

      const colorDisplay = screen.getByTestId('color-picker-display');
      expect(colorDisplay).toHaveStyle({ backgroundColor: '#E0E0E0' });
    });

    it('should display color value as text', () => {
      render(
        <ColorPicker
          color="#FF5733"
          onChange={mockOnChange}
          disabled={false}
        />
      );

      expect(screen.getByText('#FF5733')).toBeInTheDocument();
    });
  });

  describe('Interaction', () => {
    it('should call onChange when color is selected', () => {
      render(
        <ColorPicker
          color="#FF5733"
          onChange={mockOnChange}
          disabled={false}
        />
      );

      const colorButton = screen.getByTestId('color-picker-button');
      fireEvent.click(colorButton);

      const colorInput = screen.getByTestId('color-picker-input');
      fireEvent.change(colorInput, { target: { value: '#00FF00' } });

      expect(mockOnChange).toHaveBeenCalledWith('#00FF00');
    });

    it('should not call onChange when disabled', () => {
      render(
        <ColorPicker
          color="#FF5733"
          onChange={mockOnChange}
          disabled={true}
        />
      );

      const colorButton = screen.getByTestId('color-picker-button');
      fireEvent.click(colorButton);

      expect(mockOnChange).not.toHaveBeenCalled();
    });

    it('should open color picker popover when clicked', () => {
      render(
        <ColorPicker
          color="#FF5733"
          onChange={mockOnChange}
          disabled={false}
        />
      );

      const colorButton = screen.getByTestId('color-picker-button');
      fireEvent.click(colorButton);

      const colorPickerPopover = screen.getByTestId('color-picker-popover');
      expect(colorPickerPopover).toBeVisible();
    });

    it('should close color picker when color is selected', () => {
      render(
        <ColorPicker
          color="#FF5733"
          onChange={mockOnChange}
          disabled={false}
        />
      );

      const colorButton = screen.getByTestId('color-picker-button');
      fireEvent.click(colorButton);

      const colorInput = screen.getByTestId('color-picker-input');
      fireEvent.change(colorInput, { target: { value: '#00FF00' } });

      const colorPickerPopover = screen.queryByTestId('color-picker-popover');
      expect(colorPickerPopover).not.toBeVisible();
    });
  });

  describe('Disabled state', () => {
    it('should show disabled state visually when disabled', () => {
      render(
        <ColorPicker
          color="#FF5733"
          onChange={mockOnChange}
          disabled={true}
        />
      );

      const colorButton = screen.getByTestId('color-picker-button');
      expect(colorButton).toBeDisabled();
    });

    it('should not open picker when disabled and clicked', () => {
      render(
        <ColorPicker
          color="#FF5733"
          onChange={mockOnChange}
          disabled={true}
        />
      );

      const colorButton = screen.getByTestId('color-picker-button');
      fireEvent.click(colorButton);

      const colorPickerPopover = screen.queryByTestId('color-picker-popover');
      expect(colorPickerPopover).not.toBeInTheDocument();
    });

    it('should show tooltip when disabled', () => {
      const tooltipText = 'Switch to custom color scheme to assign colors';
      render(
        <ColorPicker
          color="#FF5733"
          onChange={mockOnChange}
          disabled={true}
          disabledTooltip={tooltipText}
        />
      );

      const colorButton = screen.getByTestId('color-picker-button');
      fireEvent.mouseOver(colorButton);

      expect(screen.getByText(tooltipText)).toBeInTheDocument();
    });

    it('should not show tooltip when enabled', () => {
      render(
        <ColorPicker
          color="#FF5733"
          onChange={mockOnChange}
          disabled={false}
          disabledTooltip="Switch to custom color scheme to assign colors"
        />
      );

      const colorButton = screen.getByTestId('color-picker-button');
      fireEvent.mouseOver(colorButton);

      expect(screen.queryByText('Switch to custom color scheme to assign colors')).not.toBeInTheDocument();
    });
  });

  describe('Color validation', () => {
    it('should only accept valid hex colors', () => {
      render(
        <ColorPicker
          color="#FF5733"
          onChange={mockOnChange}
          disabled={false}
        />
      );

      const colorButton = screen.getByTestId('color-picker-button');
      fireEvent.click(colorButton);

      const colorInput = screen.getByTestId('color-picker-input');
      fireEvent.change(colorInput, { target: { value: 'invalid-color' } });

      expect(mockOnChange).not.toHaveBeenCalled();
    });

    it('should accept hex colors with # prefix', () => {
      render(
        <ColorPicker
          color="#FF5733"
          onChange={mockOnChange}
          disabled={false}
        />
      );

      const colorButton = screen.getByTestId('color-picker-button');
      fireEvent.click(colorButton);

      const colorInput = screen.getByTestId('color-picker-input');
      fireEvent.change(colorInput, { target: { value: '#AABBCC' } });

      expect(mockOnChange).toHaveBeenCalledWith('#AABBCC');
    });

    it('should normalize hex colors to uppercase', () => {
      render(
        <ColorPicker
          color="#FF5733"
          onChange={mockOnChange}
          disabled={false}
        />
      );

      const colorButton = screen.getByTestId('color-picker-button');
      fireEvent.click(colorButton);

      const colorInput = screen.getByTestId('color-picker-input');
      fireEvent.change(colorInput, { target: { value: '#aabbcc' } });

      expect(mockOnChange).toHaveBeenCalledWith('#AABBCC');
    });
  });
});
