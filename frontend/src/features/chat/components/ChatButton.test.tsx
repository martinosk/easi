import type { ReactElement } from 'react';
import { fireEvent, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { ChatButton } from './ChatButton';

const render = (ui: ReactElement) => renderWithProviders(ui, { withRouter: false });

describe('ChatButton', () => {
  it('should render when assistantAvailable is true', () => {
    render(<ChatButton assistantAvailable={true} onClick={vi.fn()} />);
    expect(screen.getByTestId('nav-chat')).toBeInTheDocument();
  });

  it('should not render when assistantAvailable is false', () => {
    render(<ChatButton assistantAvailable={false} onClick={vi.fn()} />);
    expect(screen.queryByTestId('nav-chat')).not.toBeInTheDocument();
  });

  it('should call onClick when clicked', () => {
    const onClick = vi.fn();
    render(<ChatButton assistantAvailable={true} onClick={onClick} />);
    fireEvent.click(screen.getByTestId('nav-chat'));
    expect(onClick).toHaveBeenCalled();
  });

  it('should expose active state when isActive is true', () => {
    render(<ChatButton assistantAvailable={true} onClick={vi.fn()} isActive={true} />);
    expect(screen.getByTestId('nav-chat')).toHaveAttribute('aria-pressed', 'true');
  });

  it('should not expose active state when isActive is false', () => {
    render(<ChatButton assistantAvailable={true} onClick={vi.fn()} isActive={false} />);
    expect(screen.getByTestId('nav-chat')).toHaveAttribute('aria-pressed', 'false');
  });

  it('should show a tooltip naming the assistant', async () => {
    const user = userEvent.setup();
    render(<ChatButton assistantAvailable={true} onClick={vi.fn()} />);

    await user.hover(screen.getByTestId('nav-chat'));

    expect(await screen.findByRole('tooltip')).toHaveTextContent('Architecture Assistant');
  });
});
