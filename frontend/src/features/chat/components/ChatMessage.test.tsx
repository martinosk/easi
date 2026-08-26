import type { ReactElement } from 'react';
import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { ChatMessage } from './ChatMessage';

const render = (ui: ReactElement) => renderWithProviders(ui, { withRouter: false });

describe('ChatMessage', () => {
  it('should render user message content', () => {
    render(<ChatMessage sender="user" content="Hello world" />);
    expect(screen.getByText('Hello world')).toBeInTheDocument();
  });

  it('should render assistant message content', () => {
    render(<ChatMessage sender="assistant" content="Hi there" />);
    expect(screen.getByText('Hi there')).toBeInTheDocument();
  });

  it('should mark user messages with the user sender', () => {
    render(<ChatMessage sender="user" content="Test" />);
    expect(screen.getByTestId('chat-message')).toHaveAttribute('data-sender', 'user');
  });

  it('should mark assistant messages with the assistant sender', () => {
    render(<ChatMessage sender="assistant" content="Test" />);
    expect(screen.getByTestId('chat-message')).toHaveAttribute('data-sender', 'assistant');
  });

  it('should render markdown in assistant messages', () => {
    render(<ChatMessage sender="assistant" content="**bold text**" />);
    expect(screen.getByText('bold text')).toBeInTheDocument();
    const strong = screen.getByText('bold text').closest('strong');
    expect(strong).toBeInTheDocument();
  });

  it('should show streaming cursor when isStreaming is true', () => {
    render(<ChatMessage sender="assistant" content="Loading" isStreaming />);
    expect(screen.getByTestId('chat-cursor')).toBeInTheDocument();
  });

  it('should not show streaming cursor when isStreaming is false', () => {
    render(<ChatMessage sender="assistant" content="Done" />);
    expect(screen.queryByTestId('chat-cursor')).not.toBeInTheDocument();
  });
});
