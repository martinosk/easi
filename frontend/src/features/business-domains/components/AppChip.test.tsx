import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { toComponentId } from '../../../api/types';
import { buildCapabilityRealization, renderWithProviders } from '../../../test/helpers';
import { AppChip } from './AppChip';
import classes from './AppChip.module.css';

describe('AppChip', () => {
  it('renders the component name and calls onClick with the componentId, stopping propagation', async () => {
    const onClick = vi.fn();
    const onCardClick = vi.fn();
    const realization = buildCapabilityRealization({ componentId: toComponentId('comp-1'), componentName: 'Phoenix' });

    renderWithProviders(
      // biome-ignore lint/a11y/noStaticElementInteractions: test-only wrapper simulating an ancestor click handler
      // biome-ignore lint/a11y/useKeyWithClickEvents: test-only wrapper simulating an ancestor click handler
      <div onClick={onCardClick}>
        <AppChip realization={realization} onClick={onClick} />
      </div>,
      { withRouter: false },
    );

    await userEvent.click(screen.getByTestId('app-chip-comp-1'));

    expect(onClick).toHaveBeenCalledWith('comp-1');
    expect(onCardClick).not.toHaveBeenCalled();
  });

  it('shows an inherited tooltip title when the realization is inherited', () => {
    const realization = buildCapabilityRealization({
      componentName: 'Phoenix',
      origin: 'Inherited',
      sourceCapabilityName: 'Ferry Booking',
    });

    renderWithProviders(<AppChip realization={realization} onClick={vi.fn()} />, { withRouter: false });

    expect(screen.getByTitle('Phoenix (inherited from Ferry Booking)')).toBeInTheDocument();
  });

  it('falls back to the componentId when no componentName is provided', () => {
    const realization = buildCapabilityRealization({ componentId: toComponentId('comp-42'), componentName: undefined });

    renderWithProviders(<AppChip realization={realization} onClick={vi.fn()} />, { withRouter: false });

    expect(screen.getByTestId('app-chip-comp-42')).toHaveTextContent('comp-42');
  });

  it.each([
    ['Full', classes.full],
    ['Partial', classes.partial],
    ['Planned', classes.planned],
  ] as const)('tints a %s realisation with its status class', (realizationLevel, expectedClass) => {
    const realization = buildCapabilityRealization({ componentId: toComponentId('comp-1'), realizationLevel });

    renderWithProviders(<AppChip realization={realization} onClick={vi.fn()} />, { withRouter: false });

    const chip = screen.getByTestId('app-chip-comp-1');
    expect(chip.className).toContain(expectedClass);
    expect(chip.className).not.toContain(classes.inherited);
  });

  it('dims an inherited realization regardless of its realisation-level tint', () => {
    const realization = buildCapabilityRealization({
      componentId: toComponentId('comp-1'),
      realizationLevel: 'Full',
      origin: 'Inherited',
    });

    renderWithProviders(<AppChip realization={realization} onClick={vi.fn()} />, { withRouter: false });

    const chip = screen.getByTestId('app-chip-comp-1');
    expect(chip.className).toContain(classes.full);
    expect(chip.className).toContain(classes.inherited);
  });
});
