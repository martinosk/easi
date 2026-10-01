import { screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../test/helpers';
import CanvasView from './CanvasView';

const initialization = vi.hoisted(() => ({
  state: { isLoading: false, error: null as Error | null, isInitialized: true, currentViewId: null },
}));

vi.mock('../../hooks/useAppInitialization', () => ({
  useAppInitialization: () => initialization.state,
}));

vi.mock('./CanvasContainer', () => ({
  default: () => <div data-testid="canvas-container" />,
}));

function inlineStyle(testId: string): string {
  return screen.getByTestId(testId).getAttribute('style') ?? '';
}

describe('CanvasView', () => {
  beforeEach(() => {
    initialization.state = { isLoading: false, error: null, isInitialized: true, currentViewId: null };
  });

  it('shows the canvas once initialised', () => {
    renderWithProviders(<CanvasView />);

    expect(screen.getByTestId('canvas-container')).toBeInTheDocument();
  });

  it('fits the loading screen into the main region instead of the viewport', () => {
    initialization.state = { ...initialization.state, isLoading: true };
    renderWithProviders(<CanvasView />);

    expect(inlineStyle('loading-screen')).toContain('flex-grow: 1');
    expect(inlineStyle('loading-screen')).not.toContain('100vh');
  });

  it('fits the error screen into the main region instead of the viewport', () => {
    initialization.state = { ...initialization.state, error: new Error('Failed to load views') };
    renderWithProviders(<CanvasView />);

    expect(screen.getByText('Failed to load views')).toBeInTheDocument();
    expect(inlineStyle('error-screen')).toContain('flex-grow: 1');
    expect(inlineStyle('error-screen')).not.toContain('100vh');
  });
});
