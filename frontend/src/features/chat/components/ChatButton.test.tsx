import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ChatButton } from './ChatButton';

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

  it('should show active state when isActive is true', () => {
    render(<ChatButton assistantAvailable={true} onClick={vi.fn()} isActive={true} />);
    expect(screen.getByTestId('nav-chat')).toHaveClass('app-header-action-btn-active');
  });

  it('should not show active state when isActive is false', () => {
    render(<ChatButton assistantAvailable={true} onClick={vi.fn()} isActive={false} />);
    expect(screen.getByTestId('nav-chat')).not.toHaveClass('app-header-action-btn-active');
  });
});
