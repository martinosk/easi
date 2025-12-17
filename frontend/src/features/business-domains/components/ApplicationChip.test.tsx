import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { ApplicationChip } from './ApplicationChip';
import type { CapabilityRealization, ComponentId, CapabilityId, RealizationId } from '../../../api/types';

describe('ApplicationChip', () => {
  const createRealization = (
    overrides: Partial<CapabilityRealization> = {}
  ): CapabilityRealization => ({
    id: 'real-1' as RealizationId,
    capabilityId: 'cap-1' as CapabilityId,
    componentId: 'comp-1' as ComponentId,
    componentName: 'SAP Finance',
    realizationLevel: 'Full',
    origin: 'Direct',
    linkedAt: '2024-01-01',
    _links: { self: '/api/v1/realizations/real-1' },
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
        </div>
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

      expect(screen.getByRole('button')).toHaveAttribute(
        'title',
        'SAP Finance (inherited from Parent Capability)'
      );
    });
  });

  describe('styling based on realization level', () => {
    it('renders button with expected styles for Full realization', () => {
      const realization = createRealization({ realizationLevel: 'Full' });
      render(<ApplicationChip realization={realization} onClick={vi.fn()} />);

      const button = screen.getByRole('button');
      expect(button).toBeInTheDocument();
    });

    it('renders button with expected styles for Partial realization', () => {
      const realization = createRealization({ realizationLevel: 'Partial' });
      render(<ApplicationChip realization={realization} onClick={vi.fn()} />);

      const button = screen.getByRole('button');
      expect(button).toBeInTheDocument();
    });

    it('renders button with expected styles for Planned realization', () => {
      const realization = createRealization({ realizationLevel: 'Planned' });
      render(<ApplicationChip realization={realization} onClick={vi.fn()} />);

      const button = screen.getByRole('button');
      expect(button).toBeInTheDocument();
    });
  });

  describe('styling based on origin', () => {
    it('applies darker background for Direct realizations', () => {
      const realization = createRealization({ origin: 'Direct' });
      render(<ApplicationChip realization={realization} onClick={vi.fn()} />);

      const button = screen.getByRole('button');
      expect(button).toHaveStyle({ backgroundColor: '#e2e8f0' });
    });

    it('applies lighter background for Inherited realizations', () => {
      const realization = createRealization({ origin: 'Inherited' });
      render(<ApplicationChip realization={realization} onClick={vi.fn()} />);

      const button = screen.getByRole('button');
      expect(button).toHaveStyle({ backgroundColor: '#f1f5f9' });
    });
  });
});
