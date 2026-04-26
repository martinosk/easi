import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { MantineTestWrapper } from '../../../test/helpers/mantineTestWrapper';
import { DynamicModeToolbar } from './DynamicModeToolbar';

function renderToolbar(props: Partial<React.ComponentProps<typeof DynamicModeToolbar>> = {}) {
  return render(
    <MantineTestWrapper>
      <DynamicModeToolbar
        dirty={false}
        isSaving={false}
        saveLabel="Save view"
        onSave={vi.fn()}
        onDiscard={vi.fn()}
        {...props}
      />
    </MantineTestWrapper>,
  );
}

describe('DynamicModeToolbar', () => {
  it('always renders Save and Cancel buttons', () => {
    renderToolbar();
    expect(screen.getByRole('button', { name: /save view/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
  });

  it('Save and Cancel are disabled when there are no draft changes', () => {
    renderToolbar({ dirty: false });
    expect(screen.getByRole('button', { name: /save view/i })).toBeDisabled();
    expect(screen.getByRole('button', { name: /cancel/i })).toBeDisabled();
  });

  it('Save and Cancel become enabled when dirty', () => {
    renderToolbar({ dirty: true });
    expect(screen.getByRole('button', { name: /save view/i })).toBeEnabled();
    expect(screen.getByRole('button', { name: /cancel/i })).toBeEnabled();
  });

  it('clicking Save calls onSave', async () => {
    const onSave = vi.fn();
    renderToolbar({ dirty: true, onSave });

    await userEvent.click(screen.getByRole('button', { name: /save view/i }));
    expect(onSave).toHaveBeenCalled();
  });

  it('clicking Cancel when dirty shows a confirm dialog before calling onDiscard', async () => {
    const onDiscard = vi.fn();
    renderToolbar({ dirty: true, onDiscard });

    await userEvent.click(screen.getByRole('button', { name: /^cancel$/i }));
    expect(onDiscard).not.toHaveBeenCalled();
    expect(screen.getByRole('dialog', { name: /discard changes/i })).toBeInTheDocument();

    const confirmButton = screen.getByRole('button', { name: /discard changes/i });
    await userEvent.click(confirmButton);
    expect(onDiscard).toHaveBeenCalled();
  });

  it('the confirm dialog can be dismissed without discarding', async () => {
    const onDiscard = vi.fn();
    renderToolbar({ dirty: true, onDiscard });

    await userEvent.click(screen.getByRole('button', { name: /^cancel$/i }));
    const dialogCancel = screen.getByRole('button', { name: /keep editing/i });
    await userEvent.click(dialogCancel);

    expect(onDiscard).not.toHaveBeenCalled();
  });

  it('disables Save and Cancel while saving', () => {
    renderToolbar({ dirty: true, isSaving: true });
    expect(screen.getByRole('button', { name: /save view/i })).toBeDisabled();
    expect(screen.getByRole('button', { name: /cancel/i })).toBeDisabled();
  });
});
