import { screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { buildInternalTeam } from '../../../test/helpers/entityBuilders';
import { InternalTeamDetails } from './InternalTeamDetails';

function renderPanel(hasOnePagerLink: boolean) {
  const team = buildInternalTeam({
    _links: {
      self: { href: '/api/v1/internal-teams/team-1', method: 'GET' },
      ...(hasOnePagerLink ? { 'x-one-pager': { href: '/api/v1/one-pagers/internal-team/team-1', method: 'GET' } } : {}),
    },
  });
  return renderWithProviders(
    <InternalTeamDetails team={team} relationships={[]} canRemoveFromView={false} onEdit={vi.fn()} onRemoveFromView={vi.fn()} />,
  );
}

describe('InternalTeamDetails one-pager surface', () => {
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
