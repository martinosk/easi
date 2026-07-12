import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { toComponentId } from '../../../api/types';
import { buildCapabilityRealization, renderWithProviders } from '../../../test/helpers';
import type { AssessedRealization } from '../hooks/domainBoardViewModel';
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

  function renderChip(overrides: Partial<AssessedRealization> = {}) {
    const realization: AssessedRealization = {
      ...buildCapabilityRealization({ componentId: toComponentId('comp-1') }),
      ...overrides,
    };
    renderWithProviders(<AppChip realization={realization} onClick={vi.fn()} />, { withRouter: false });
    return screen.getByTestId('app-chip-comp-1');
  }

  it.each([
    ['Full', classes.full],
    ['Partial', classes.partial],
    ['Planned', classes.planned],
  ] as const)('tints a %s realisation with its status class', (realizationLevel, expectedClass) => {
    const chip = renderChip({ realizationLevel });

    expect(chip.className).toContain(expectedClass);
    expect(chip.className).not.toContain(classes.inherited);
  });

  it('dims an inherited realization regardless of its realisation-level tint', () => {
    const chip = renderChip({ realizationLevel: 'Full', origin: 'Inherited' });

    expect(chip.className).toContain(classes.full);
    expect(chip.className).toContain(classes.inherited);
  });

  it('shows a single-letter grade badge on a Direct chip with a current assessment', () => {
    const realization: AssessedRealization = {
      ...buildCapabilityRealization({ componentId: toComponentId('comp-1'), origin: 'Direct' }),
      timeGrade: 'Migrate',
    };

    renderWithProviders(<AppChip realization={realization} onClick={vi.fn()} />, { withRouter: false });

    expect(screen.getByTestId('app-chip-grade-comp-1')).toHaveTextContent('M');
  });

  it('shows no grade badge on an unassessed Direct chip', () => {
    const realization = buildCapabilityRealization({ componentId: toComponentId('comp-1'), origin: 'Direct' });

    renderWithProviders(<AppChip realization={realization} onClick={vi.fn()} />, { withRouter: false });

    expect(screen.queryByTestId('app-chip-grade-comp-1')).not.toBeInTheDocument();
  });

  it('shows no grade badge on an Inherited chip even when a grade is present', () => {
    const realization: AssessedRealization = {
      ...buildCapabilityRealization({ componentId: toComponentId('comp-1'), origin: 'Inherited' }),
      timeGrade: 'Migrate',
    };

    renderWithProviders(<AppChip realization={realization} onClick={vi.fn()} />, { withRouter: false });

    expect(screen.queryByTestId('app-chip-grade-comp-1')).not.toBeInTheDocument();
  });

  it('tints a standard-role Direct chip with the standard class instead of the level class', () => {
    const chip = renderChip({ origin: 'Direct', realizationLevel: 'Partial', role: 'standard' });

    expect(chip.className).toContain(classes.roleStandard);
    expect(chip.className).not.toContain(classes.partial);
  });

  it('tints a legacy-role Direct chip with the legacy class instead of the level class', () => {
    const chip = renderChip({ origin: 'Direct', realizationLevel: 'Full', role: 'legacy' });

    expect(chip.className).toContain(classes.roleLegacy);
    expect(chip.className).not.toContain(classes.full);
  });

  it('falls back to the level tint for an unclassified Direct chip', () => {
    const chip = renderChip({ origin: 'Direct', realizationLevel: 'Full' });

    expect(chip.className).toContain(classes.full);
    expect(chip.className).not.toContain(classes.roleStandard);
    expect(chip.className).not.toContain(classes.roleLegacy);
  });

  it('never applies role tinting to an Inherited chip even when a role is present', () => {
    const realization: AssessedRealization = {
      ...buildCapabilityRealization({ componentId: toComponentId('comp-1'), origin: 'Inherited', realizationLevel: 'Full' }),
      role: 'standard',
    };

    renderWithProviders(<AppChip realization={realization} onClick={vi.fn()} />, { withRouter: false });

    const chip = screen.getByTestId('app-chip-comp-1');
    expect(chip.className).not.toContain(classes.roleStandard);
    expect(chip.className).toContain(classes.full);
  });
});
