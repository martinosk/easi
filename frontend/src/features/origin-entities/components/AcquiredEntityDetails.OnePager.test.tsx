import { screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { buildAcquiredEntity } from '../../../test/helpers/entityBuilders';
import { AcquiredEntityDetails } from './AcquiredEntityDetails';

function renderPanel(hasOnePagerLink: boolean) {
  const entity = buildAcquiredEntity({
    _links: {
      self: { href: '/api/v1/acquired-entities/ae-1', method: 'GET' },
      ...(hasOnePagerLink ? { 'x-one-pager': { href: '/api/v1/one-pagers/acquired-entity/ae-1', method: 'GET' } } : {}),
    },
  });
  return renderWithProviders(
    <AcquiredEntityDetails entity={entity} relationships={[]} canRemoveFromView={false} onEdit={vi.fn()} onRemoveFromView={vi.fn()} />,
  );
}

describe('AcquiredEntityDetails one-pager surface', () => {
  it('shows the One-Pager action and no inline edit form', async () => {
    renderPanel(true);

    expect(await screen.findByRole('button', { name: 'One-Pager' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Save one-pager' })).not.toBeInTheDocument();
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
  });

  it('shows no One-Pager action when the subject lacks the link', () => {
    renderPanel(false);

    expect(screen.queryByRole('button', { name: 'One-Pager' })).not.toBeInTheDocument();
  });
});
