import type { ReactElement } from 'react';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { OnePagerIncompleteIndicator } from './OnePagerIncompleteIndicator';

function render(ui: ReactElement) {
  return renderWithProviders(ui, { withRouter: false });
}

describe('OnePagerIncompleteIndicator', () => {
  it('renders nothing when the subject is complete', () => {
    render(<OnePagerIncompleteIndicator id="sub-1" complete={true} />);

    expect(screen.queryByTestId('one-pager-incomplete-sub-1')).not.toBeInTheDocument();
  });

  it('renders nothing when the subject is absent from the completeness collection', () => {
    render(<OnePagerIncompleteIndicator id="sub-1" />);

    expect(screen.queryByTestId('one-pager-incomplete-sub-1')).not.toBeInTheDocument();
  });

  it('renders a warning indicator when the subject is incomplete', () => {
    render(<OnePagerIncompleteIndicator id="sub-1" complete={false} />);

    expect(screen.getByTestId('one-pager-incomplete-sub-1')).toBeInTheDocument();
  });

  it('shows the "One-pager incomplete" tooltip on hover', async () => {
    const user = userEvent.setup();
    render(<OnePagerIncompleteIndicator id="sub-1" complete={false} />);

    await user.hover(screen.getByTestId('one-pager-incomplete-sub-1'));

    expect(await screen.findByText('One-pager incomplete')).toBeInTheDocument();
  });
});
