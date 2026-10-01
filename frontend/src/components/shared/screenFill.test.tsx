import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { renderWithProviders } from '../../test/helpers';
import { ErrorScreen } from './ErrorScreen';
import { LoadingScreen } from './LoadingScreen';

function inlineStyle(testId: string): string {
  return screen.getByTestId(testId).getAttribute('style') ?? '';
}

describe('screens used as a full page', () => {
  it('lets the error screen fill the viewport by default', () => {
    renderWithProviders(<ErrorScreen error="boom" onRetry={() => {}} />, { withRouter: false });

    expect(inlineStyle('error-screen')).toContain('min-height: 100vh');
  });

  it('lets the loading screen fill the viewport by default', () => {
    renderWithProviders(<LoadingScreen />, { withRouter: false });

    expect(inlineStyle('loading-screen')).toContain('min-height: 100vh');
  });
});
