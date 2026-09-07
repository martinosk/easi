import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import type { TimeAssessment, TimeSuggestion } from '../../architecture-direction/types';
import { RealizationAssessment, type RealizationAssessmentProps } from './RealizationAssessment';

const mutateAssessAsync = vi.fn();
const mutateRemoveAsync = vi.fn();

vi.mock('../../architecture-direction/hooks/useTimeAssessments', () => ({
  useAssessRealization: () => ({ mutateAsync: mutateAssessAsync, isPending: false }),
  useRemoveTimeAssessment: () => ({ mutateAsync: mutateRemoveAsync, isPending: false }),
}));

function buildAssessment(overrides: Partial<TimeAssessment> = {}): TimeAssessment {
  return {
    id: 'ta-1',
    capabilityId: 'cap-1',
    capabilityName: 'Booking management',
    componentId: 'comp-1',
    componentName: 'Seabook',
    grade: 'Migrate',
    rationale: '',
    assessedBy: 'user-1',
    assessedByName: 'Domain Architect',
    assessedAt: '2026-02-01T00:00:00Z',
    stale: false,
    suggestion: null,
    _links: {
      self: { href: '', method: 'GET' },
      edit: { href: '', method: 'PUT' },
      delete: { href: '', method: 'DELETE' },
    },
    ...overrides,
  };
}

function buildSuggestion(overrides: Partial<TimeSuggestion> = {}): TimeSuggestion {
  return { grade: 'Migrate', confidence: 'MEDIUM', technicalGap: 2, functionalGap: 0.5, ...overrides };
}

function renderAssessment(overrides: Partial<RealizationAssessmentProps> = {}) {
  renderWithProviders(
    <RealizationAssessment
      capabilityId="cap-1"
      componentId="comp-1"
      assessment={undefined}
      rollup={undefined}
      canAssess={false}
      suggestion={null}
      {...overrides}
    />,
    { withRouter: false },
  );
}

describe('RealizationAssessment', () => {
  beforeEach(() => {
    mutateAssessAsync.mockReset().mockResolvedValue(undefined);
    mutateRemoveAsync.mockReset().mockResolvedValue(undefined);
  });

  it('shows unassessed with no Assess control for a read-only caller', () => {
    renderAssessment();

    expect(screen.getByText('unassessed')).toBeInTheDocument();
    expect(screen.queryByTestId('assess-btn-comp-1')).not.toBeInTheDocument();
  });

  it('shows an Assess control for an unassessed realization when the caller can write', () => {
    renderAssessment({ canAssess: true });

    expect(screen.getByTestId('assess-btn-comp-1')).toBeInTheDocument();
  });

  it('shows the current grade, assessor, and date for an assessed realization', () => {
    renderAssessment({ assessment: buildAssessment() });

    expect(screen.getByTestId('assessment-comp-1')).toHaveTextContent('Migrate — for this capability');
    expect(screen.getByTestId('assessment-comp-1')).toHaveTextContent('Domain Architect');
  });

  it('falls back to assessedBy when assessedByName is absent', () => {
    renderAssessment({ assessment: buildAssessment({ assessedByName: undefined, assessedBy: 'user-raw-id' }) });

    expect(screen.getByTestId('assessment-comp-1')).toHaveTextContent('user-raw-id');
  });

  it.each([
    [true, 'stale'],
    [false, 'not stale'],
  ] as const)('shows the stale marker only when the assessment is %s', (stale, _description) => {
    renderAssessment({ assessment: buildAssessment({ stale }) });

    if (stale) {
      expect(screen.getByTestId('assessment-stale-comp-1')).toHaveTextContent('stale');
    } else {
      expect(screen.queryByTestId('assessment-stale-comp-1')).not.toBeInTheDocument();
    }
  });

  it('shows the landscape rollup line for an assessed realization', () => {
    renderAssessment({ assessment: buildAssessment(), rollup: { Invest: 1, Tolerate: 1, Migrate: 1, Eliminate: 1 } });

    expect(screen.getByTestId('assessment-rollup-comp-1')).toHaveTextContent('Across landscape: I×1 · T×1 · M×1 · E×1');
  });

  it.each([
    [{ self: { href: '', method: 'GET' as const } }, false, 'no write links'],
    [
      {
        self: { href: '', method: 'GET' as const },
        edit: { href: '', method: 'PUT' as const },
        delete: { href: '', method: 'DELETE' as const },
      },
      true,
      'edit and delete links',
    ],
  ] as const)('shows re-assess and remove controls only when the assessment carries %s', (_links, expectVisible, _description) => {
    renderAssessment({ assessment: buildAssessment({ _links }) });

    if (expectVisible) {
      expect(screen.getByTestId('reassess-btn-comp-1')).toBeInTheDocument();
      expect(screen.getByTestId('remove-assessment-btn-comp-1')).toBeInTheDocument();
    } else {
      expect(screen.queryByTestId('reassess-btn-comp-1')).not.toBeInTheDocument();
      expect(screen.queryByTestId('remove-assessment-btn-comp-1')).not.toBeInTheDocument();
    }
  });

  it('shows the suggestion and its confidence beside the grade choices, without pre-selecting one', async () => {
    renderAssessment({ canAssess: true, suggestion: buildSuggestion({ grade: 'Eliminate', confidence: 'MEDIUM' }) });

    await userEvent.click(screen.getByTestId('assess-btn-comp-1'));

    expect(screen.getByTestId('assessment-suggestion-comp-1')).toHaveTextContent('Suggested: Eliminate');
    expect(screen.getByTestId('assessment-suggestion-comp-1')).toHaveTextContent('medium confidence');
    expect(screen.getByRole('radio', { name: 'Eliminate' })).not.toBeChecked();
  });

  it('keeps the recorded grade selected when re-assessing a realisation the suggestion disagrees with', async () => {
    renderAssessment({
      assessment: buildAssessment({ grade: 'Tolerate' }),
      suggestion: buildSuggestion({ grade: 'Migrate' }),
    });

    await userEvent.click(screen.getByTestId('reassess-btn-comp-1'));

    expect(screen.getByRole('radio', { name: 'Tolerate' })).toBeChecked();
    expect(screen.getByTestId('assessment-suggestion-comp-1')).toHaveTextContent('Suggested: Migrate');
  });

  it('shows the suggestion alongside a recorded grade that disagrees with it', () => {
    renderAssessment({
      assessment: buildAssessment({ grade: 'Tolerate' }),
      suggestion: buildSuggestion({ grade: 'Migrate' }),
    });

    expect(screen.getByTestId('assessment-comp-1')).toHaveTextContent('Tolerate — for this capability');
    expect(screen.getByTestId('assessment-suggestion-comp-1')).toHaveTextContent('Suggested: Migrate');
  });

  it('shows no suggestion when the fit data yields none', async () => {
    renderAssessment({ canAssess: true, suggestion: buildSuggestion({ grade: null, confidence: 'LOW' }) });

    expect(screen.queryByTestId('assessment-suggestion-comp-1')).not.toBeInTheDocument();

    await userEvent.click(screen.getByTestId('assess-btn-comp-1'));

    expect(screen.queryByTestId('assessment-suggestion-comp-1')).not.toBeInTheDocument();
  });

  it('shows the suggestion on an unassessed realisation', () => {
    renderAssessment({ suggestion: buildSuggestion({ grade: 'Invest', confidence: 'HIGH' }) });

    expect(screen.getByText('unassessed')).toBeInTheDocument();
    expect(screen.getByTestId('assessment-suggestion-comp-1')).toHaveTextContent('Suggested: Invest');
    expect(screen.getByTestId('assessment-suggestion-comp-1')).toHaveTextContent('high confidence');
  });

  it('assesses the realisation with the selected grade when Save is clicked', async () => {
    renderAssessment({ canAssess: true });

    await userEvent.click(screen.getByTestId('assess-btn-comp-1'));
    await userEvent.click(screen.getByRole('radio', { name: 'Migrate' }));
    await userEvent.click(screen.getByRole('button', { name: 'Save' }));

    expect(mutateAssessAsync).toHaveBeenCalledWith({
      capabilityId: 'cap-1',
      componentId: 'comp-1',
      request: { grade: 'Migrate', rationale: undefined },
    });
  });

  it('removes the assessment and returns the row to unassessed when Remove is clicked', async () => {
    const assessment = buildAssessment();
    renderAssessment({ assessment });

    await userEvent.click(screen.getByTestId('remove-assessment-btn-comp-1'));

    expect(mutateRemoveAsync).toHaveBeenCalledWith({ assessment });
  });
});
