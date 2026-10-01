import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { TimeGradeBadge } from './TimeGradeBadge';

describe('TimeGradeBadge', () => {
  it.each([
    ['Invest', 'I'],
    ['Tolerate', 'T'],
    ['Migrate', 'M'],
    ['Eliminate', 'E'],
  ] as const)('shows %s as the letter %s', (grade, letter) => {
    renderWithProviders(<TimeGradeBadge grade={grade} testId="badge" />, { withRouter: false });

    expect(screen.getByTestId('badge')).toHaveTextContent(letter);
  });

  it('exposes the full grade name as accessible label when labelled', () => {
    renderWithProviders(<TimeGradeBadge grade="Eliminate" labelled />, { withRouter: false });

    expect(screen.getByRole('img', { name: 'Eliminate' })).toHaveTextContent('E');
  });

  it('carries the given title', () => {
    renderWithProviders(<TimeGradeBadge grade="Migrate" title="Migrate — for this capability" testId="badge" />, {
      withRouter: false,
    });

    expect(screen.getByTestId('badge')).toHaveAttribute('title', 'Migrate — for this capability');
  });
});
