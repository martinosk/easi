import { fireEvent, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { CapabilityId, CapabilityRealization, ComponentId, RealizationId } from '../../../api/types';
import { renderWithProviders } from '../../../test/helpers/renderWithProviders';
import { ApplicationChip } from './ApplicationChip';

function render(ui: React.ReactElement) {
  return renderWithProviders(ui, { withRouter: false });
}

describe('ApplicationChip', () => {
  const createRealization = (overrides: Partial<CapabilityRealization> = {}): CapabilityRealization => ({
    id: 'real-1' as RealizationId,
    capabilityId: 'cap-1' as CapabilityId,
    componentId: 'comp-1' as ComponentId,
    componentName: 'SAP Finance',
    realizationLevel: 'Full',
    origin: 'Direct',
    linkedAt: '2024-01-01',
    _links: { self: { href: '/api/v1/realizations/real-1', method: 'GET' } },
    ...overrides,
  });

  describe('rendering', () => {
    it('displays component name', () => {
      const realization = createRealization();
      render(<ApplicationChip realization={realization} onClick={vi.fn()} />);

      expect(screen.getByText('SAP Finance')).toBeInTheDocument();
    });

    it('falls back to componentId when componentName is not available', () => {
      const realization = createRealization({ componentName: undefined });
      render(<ApplicationChip realization={realization} onClick={vi.fn()} />);

      expect(screen.getByText('comp-1')).toBeInTheDocument();
    });

    it('shows inherited indicator for inherited realizations', () => {
      const realization = createRealization({ origin: 'Inherited' });
      render(<ApplicationChip realization={realization} onClick={vi.fn()} />);

      expect(screen.getByText('↓')).toBeInTheDocument();
    });

    it('does not show inherited indicator for direct realizations', () => {
      const realization = createRealization({ origin: 'Direct' });
      render(<ApplicationChip realization={realization} onClick={vi.fn()} />);

      expect(screen.queryByText('↓')).not.toBeInTheDocument();
    });
  });

  describe('click handling', () => {
    it('calls onClick with componentId when clicked', () => {
      const onClick = vi.fn();
      const realization = createRealization();
      render(<ApplicationChip realization={realization} onClick={onClick} />);

      fireEvent.click(screen.getByRole('button'));

      expect(onClick).toHaveBeenCalledWith('comp-1');
    });

    it('stops event propagation to prevent parent handlers from firing', () => {
      const onClick = vi.fn();
      const parentClick = vi.fn();
      const realization = createRealization();

      render(
        <div onClick={parentClick}>
          <ApplicationChip realization={realization} onClick={onClick} />
        </div>,
      );

      fireEvent.click(screen.getByRole('button'));

      expect(onClick).toHaveBeenCalledTimes(1);
      expect(parentClick).not.toHaveBeenCalled();
    });
  });

  describe('tooltip', () => {
    it('shows component name in tooltip for direct realizations', () => {
      const realization = createRealization({ origin: 'Direct' });
      render(<ApplicationChip realization={realization} onClick={vi.fn()} />);

      expect(screen.getByRole('button')).toHaveAttribute('title', 'SAP Finance');
    });

    it('shows inheritance info in tooltip for inherited realizations', () => {
      const realization = createRealization({
        origin: 'Inherited',
        sourceCapabilityName: 'Parent Capability',
      });
      render(<ApplicationChip realization={realization} onClick={vi.fn()} />);

      expect(screen.getByRole('button')).toHaveAttribute('title', 'SAP Finance (inherited from Parent Capability)');
    });
  });
});
