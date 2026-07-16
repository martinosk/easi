import { screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { buildComponent } from '../../../test/helpers/entityBuilders';
import { ComponentDetailsContent } from './ComponentDetails';

function renderPanel(hasOnePagerLink: boolean) {
  const component = buildComponent({
    _links: {
      self: { href: '/api/v1/components/comp-1', method: 'GET' },
      ...(hasOnePagerLink ? { 'x-one-pager': { href: '/api/v1/one-pagers/application/comp-1', method: 'GET' } } : {}),
    },
  });
  return renderWithProviders(
    <ComponentDetailsContent component={component} realizations={[]} capabilities={[]} onEdit={vi.fn()} />,
  );
}

describe('ComponentDetails one-pager surface', () => {
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
