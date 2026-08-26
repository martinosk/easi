import { fireEvent, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { renderWithProviders } from '../../test/helpers';
import { HelpTooltip } from './HelpTooltip';

const render = (ui: React.ReactElement) => renderWithProviders(ui, { withRouter: false });

describe('HelpTooltip', () => {
  it('renders the label next to the help icon', () => {
    render(<HelpTooltip label="Maturity" content="How mature the capability is" />);
    expect(screen.getByText('Maturity')).toBeInTheDocument();
    expect(screen.getByRole('img', { name: 'Help' })).toBeInTheDocument();
  });

  it('hides the label in icon-only mode', () => {
    render(<HelpTooltip label="Maturity" content="Explanation" iconOnly />);
    expect(screen.queryByText('Maturity')).not.toBeInTheDocument();
    expect(screen.getByRole('img', { name: 'Help' })).toBeInTheDocument();
  });

  it('shows the content when the icon is hovered', async () => {
    render(<HelpTooltip content="How mature the capability is" />);
    fireEvent.mouseEnter(screen.getByRole('img', { name: 'Help' }));
    expect(await screen.findByText('How mature the capability is')).toBeInTheDocument();
  });
});
