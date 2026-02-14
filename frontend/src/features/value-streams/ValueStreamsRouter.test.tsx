import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { ValueStreamsRouter } from './ValueStreamsRouter';

vi.mock('./pages/ValueStreamsPage', () => ({
  ValueStreamsPage: () => <div data-testid="value-streams-page">ValueStreamsPage</div>,
}));

vi.mock('./pages/ValueStreamDetailPage', () => ({
  ValueStreamDetailPage: () => <div data-testid="value-stream-detail-page">ValueStreamDetailPage</div>,
}));

function renderAtPath(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route path="/value-streams/*" element={<ValueStreamsRouter />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe('ValueStreamsRouter', () => {
  it('should render ValueStreamsPage at /value-streams', () => {
    renderAtPath('/value-streams');

    expect(screen.getByTestId('value-streams-page')).toBeInTheDocument();
  });

  it('should render ValueStreamsPage at /value-streams/', () => {
    renderAtPath('/value-streams/');

    expect(screen.getByTestId('value-streams-page')).toBeInTheDocument();
  });

  it('should render ValueStreamDetailPage at /value-streams/:valueStreamId', () => {
    renderAtPath('/value-streams/vs-123');

    expect(screen.getByTestId('value-stream-detail-page')).toBeInTheDocument();
  });

  it('should not render list page when navigating to detail', () => {
    renderAtPath('/value-streams/vs-123');

    expect(screen.queryByTestId('value-streams-page')).not.toBeInTheDocument();
  });

  it('should not render detail page when navigating to list', () => {
    renderAtPath('/value-streams');

    expect(screen.queryByTestId('value-stream-detail-page')).not.toBeInTheDocument();
  });
});
