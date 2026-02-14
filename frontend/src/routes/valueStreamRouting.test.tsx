import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import { ROUTES } from './routes';

vi.mock('../features/value-streams/pages/ValueStreamsPage', () => ({
  ValueStreamsPage: () => <div data-testid="value-streams-page">List</div>,
}));

vi.mock('../features/value-streams/pages/ValueStreamDetailPage', () => ({
  ValueStreamDetailPage: () => <div data-testid="value-stream-detail-page">Detail</div>,
}));

const { ValueStreamsRouter } = await import('../features/value-streams/ValueStreamsRouter');

function renderWithRouteConfig(path: string, routePath: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route path={routePath} element={<ValueStreamsRouter />} />
        <Route path="*" element={<div data-testid="not-found">Not Found</div>} />
      </Routes>
    </MemoryRouter>,
  );
}

describe('Value stream route configuration', () => {
  const withWildcard = `${ROUTES.VALUE_STREAMS}/*`;
  const withoutWildcard = ROUTES.VALUE_STREAMS;

  describe('Correct config: path with /* wildcard', () => {
    it('should match list route /value-streams', () => {
      renderWithRouteConfig('/value-streams', withWildcard);

      expect(screen.getByTestId('value-streams-page')).toBeInTheDocument();
    });

    it('should match detail route /value-streams/:id', () => {
      renderWithRouteConfig('/value-streams/vs-abc', withWildcard);

      expect(screen.getByTestId('value-stream-detail-page')).toBeInTheDocument();
    });
  });

  describe('Broken config: path without /* wildcard', () => {
    it('should still match the exact /value-streams path', () => {
      renderWithRouteConfig('/value-streams', withoutWildcard);

      expect(screen.getByTestId('value-streams-page')).toBeInTheDocument();
    });

    it('should NOT match /value-streams/:id without wildcard', () => {
      renderWithRouteConfig('/value-streams/vs-abc', withoutWildcard);

      expect(screen.queryByTestId('value-stream-detail-page')).not.toBeInTheDocument();
      expect(screen.getByTestId('not-found')).toBeInTheDocument();
    });
  });
});
