import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { CompletenessBadge } from './CompletenessBadge';

function render(completeness: 'complete' | 'incomplete' | 'not-applicable') {
  return renderWithProviders(<CompletenessBadge completeness={completeness} />, { withRouter: false });
}

describe('CompletenessBadge', () => {
  it('renders a distinct label for complete', () => {
    render('complete');
    expect(screen.getByText('Complete')).toBeInTheDocument();
  });

  it('renders a distinct label for incomplete', () => {
    render('incomplete');
    expect(screen.getByText('Incomplete')).toBeInTheDocument();
  });

  it('renders a distinct label for not-applicable', () => {
    render('not-applicable');
    expect(screen.getByText('Not Applicable')).toBeInTheDocument();
  });

  it('uses a visibly distinct color for complete vs incomplete vs not-applicable', () => {
    const { unmount: unmountComplete } = render('complete');
    const completeStyle = screen.getByTestId('completeness-badge').getAttribute('style');
    unmountComplete();

    const { unmount: unmountIncomplete } = render('incomplete');
    const incompleteStyle = screen.getByTestId('completeness-badge').getAttribute('style');
    unmountIncomplete();

    render('not-applicable');
    const notApplicableStyle = screen.getByTestId('completeness-badge').getAttribute('style');

    expect(completeStyle).toContain('teal');
    expect(incompleteStyle).toContain('red');
    expect(notApplicableStyle).toContain('gray');
    expect(new Set([completeStyle, incompleteStyle, notApplicableStyle]).size).toBe(3);
  });
});
