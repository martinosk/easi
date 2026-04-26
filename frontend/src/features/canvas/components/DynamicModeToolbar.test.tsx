import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';
import { DynamicModeToolbar } from './DynamicModeToolbar';

function renderToolbar(props: Partial<React.ComponentProps<typeof DynamicModeToolbar>> = {}) {
  return render(
    <MantineTestWrapper>
      <DynamicModeToolbar
        enabled={false}
        dirty={false}
        isSaving={false}
        saveLabel="Save view (0)"
        onEnable={vi.fn()}
        onSave={vi.fn()}
        onDiscard={vi.fn()}
        {...props}
      />
    </MantineTestWrapper>,
  );
}

describe('DynamicModeToolbar', () => {
  it('shows the Dynamic mode toggle when disabled', () => {
    renderToolbar({ enabled: false });
    expect(screen.getByRole('button', { name: /dynamic mode/i })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /save view/i })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /cancel/i })).not.toBeInTheDocument();
  });

  it('clicking the toggle calls onEnable', async () => {
    const onEnable = vi.fn();
    renderToolbar({ enabled: false, onEnable });

    await userEvent.click(screen.getByRole('button', { name: /dynamic mode/i }));
    expect(onEnable).toHaveBeenCalled();
  });

  it('shows Save and Cancel buttons when enabled', () => {
    renderToolbar({ enabled: true, saveLabel: 'Save view (3)' });
    expect(screen.getByRole('button', { name: /save view \(3\)/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
  });

  it('clicking Save calls onSave', async () => {
    const onSave = vi.fn();
    renderToolbar({ enabled: true, onSave });

    await userEvent.click(screen.getByRole('button', { name: /save view/i }));
    expect(onSave).toHaveBeenCalled();
  });

  it('clicking Cancel when no changes calls onDiscard immediately', async () => {
    const onDiscard = vi.fn();
    renderToolbar({ enabled: true, dirty: false, onDiscard });

    await userEvent.click(screen.getByRole('button', { name: /cancel/i }));
    expect(onDiscard).toHaveBeenCalled();
  });

  it('clicking Cancel when dirty shows a confirm dialog before calling onDiscard', async () => {
    const onDiscard = vi.fn();
    renderToolbar({ enabled: true, dirty: true, onDiscard });

    await userEvent.click(screen.getByRole('button', { name: /^cancel$/i }));
    expect(onDiscard).not.toHaveBeenCalled();
    expect(screen.getByRole('dialog', { name: /discard changes/i })).toBeInTheDocument();

    const confirmButton = screen.getByRole('button', { name: /discard changes/i });
    await userEvent.click(confirmButton);
    expect(onDiscard).toHaveBeenCalled();
  });

  it('the confirm dialog can be dismissed without discarding', async () => {
    const onDiscard = vi.fn();
    renderToolbar({ enabled: true, dirty: true, onDiscard });

    await userEvent.click(screen.getByRole('button', { name: /^cancel$/i }));
    const dialogCancel = screen.getByRole('button', { name: /keep editing/i });
    await userEvent.click(dialogCancel);

    expect(onDiscard).not.toHaveBeenCalled();
  });

  it('disables Save and Cancel while saving', () => {
    renderToolbar({ enabled: true, isSaving: true });
    expect(screen.getByRole('button', { name: /save view/i })).toBeDisabled();
    expect(screen.getByRole('button', { name: /cancel/i })).toBeDisabled();
  });
});
