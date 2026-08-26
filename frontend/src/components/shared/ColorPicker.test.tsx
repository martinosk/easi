import { fireEvent, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../test/helpers';
import { ColorPicker } from './ColorPicker';

const DEBOUNCE_MS = 300;
const ARROW_RIGHT = { key: 'ArrowRight', keyCode: 39 };

function renderPicker(props: Partial<React.ComponentProps<typeof ColorPicker>> = {}) {
  const onChange = vi.fn();
  renderWithProviders(<ColorPicker color="#FF5733" onChange={onChange} disabled={false} {...props} />, {
    withRouter: false,
  });
  return { onChange };
}

const swatchButton = () => screen.getByTestId('color-picker-button');
const popover = () => screen.queryByTestId('color-picker-popover');
const hexInput = () => screen.getByTestId('color-picker-input');

async function openPicker() {
  fireEvent.click(swatchButton());
  await screen.findByTestId('color-picker-popover');
}

const expectClosed = () => waitFor(() => expect(popover()).not.toBeInTheDocument());

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

function nudgeSaturation() {
  fireEvent.keyDown(screen.getAllByRole('slider', { hidden: true })[0], ARROW_RIGHT);
}

describe('ColorPicker', () => {
  describe('display', () => {
    it('shows the current colour on the swatch and as hex text', () => {
      renderPicker();
      expect(screen.getByTestId('color-picker-display')).toHaveStyle({ backgroundColor: '#FF5733' });
      expect(screen.getByText('#FF5733')).toBeInTheDocument();
    });

    it('falls back to the neutral default when no colour is set', () => {
      renderPicker({ color: null });
      expect(screen.getByText('#E0E0E0')).toBeInTheDocument();
    });

    it('follows the colour prop when it changes', () => {
      const onChange = vi.fn();
      const { rerender } = renderWithProviders(<ColorPicker color="#FF5733" onChange={onChange} disabled={false} />, {
        withRouter: false,
      });
      rerender(<ColorPicker color="#00AA00" onChange={onChange} disabled={false} />);
      expect(screen.getByText('#00AA00')).toBeInTheDocument();
    });
  });

  describe('opening and closing', () => {
    it('is closed until the swatch is clicked', async () => {
      renderPicker();
      expect(popover()).not.toBeInTheDocument();
      await openPicker();
      expect(hexInput()).toHaveValue('#FF5733');
    });

    it('toggles closed when the swatch is clicked again', async () => {
      renderPicker();
      await openPicker();
      fireEvent.click(swatchButton());
      await expectClosed();
    });

    it('closes on Escape without committing', async () => {
      const { onChange } = renderPicker();
      await openPicker();
      fireEvent.keyDown(hexInput(), { key: 'Escape' });
      await expectClosed();
      expect(onChange).not.toHaveBeenCalled();
    });

    it('closes on click outside without committing when nothing changed', async () => {
      const { onChange } = renderPicker();
      await openPicker();
      fireEvent.mouseDown(document.body);
      await expectClosed();
      expect(onChange).not.toHaveBeenCalled();
    });
  });

  describe('committing', () => {
    it('commits and closes when a valid 6-digit hex is typed', async () => {
      const { onChange } = renderPicker();
      await openPicker();
      fireEvent.change(hexInput(), { target: { value: '#00ff00' } });
      expect(onChange).toHaveBeenCalledWith('#00FF00');
      await expectClosed();
    });

    it('does not commit a partial hex while typing', async () => {
      const { onChange } = renderPicker();
      await openPicker();
      fireEvent.change(hexInput(), { target: { value: '#00F' } });
      expect(onChange).not.toHaveBeenCalled();
      expect(popover()).toBeInTheDocument();
    });

    it('debounces picker changes before committing and stays open', async () => {
      const { onChange } = renderPicker();
      await openPicker();
      nudgeSaturation();
      expect(onChange).not.toHaveBeenCalled();
      await waitFor(() => expect(onChange).toHaveBeenCalledTimes(1));
      expect(onChange).toHaveBeenCalledWith(expect.stringMatching(/^#[0-9A-F]{6}$/));
      expect(popover()).toBeInTheDocument();
    });

    it('commits the pending colour once, immediately, when closed by click outside', async () => {
      const { onChange } = renderPicker();
      await openPicker();
      nudgeSaturation();
      fireEvent.mouseDown(document.body);
      expect(onChange).toHaveBeenCalledTimes(1);
      await expectClosed();
      await sleep(DEBOUNCE_MS);
      expect(onChange).toHaveBeenCalledTimes(1);
    });
  });

  describe('disabled', () => {
    it('disables the swatch and does not open on click', () => {
      renderPicker({ disabled: true });
      expect(swatchButton()).toBeDisabled();
      fireEvent.click(swatchButton());
      expect(popover()).not.toBeInTheDocument();
    });

    it('shows the disabled reason as a tooltip on hover', async () => {
      renderPicker({ disabled: true, disabledTooltip: 'Switch to custom color scheme' });
      fireEvent.mouseEnter(screen.getByTestId('color-picker-target'));
      expect(await screen.findByText('Switch to custom color scheme')).toBeInTheDocument();
    });
  });
});
