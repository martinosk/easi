import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { ChatMessage } from './ChatMessage';

describe('ChatMessage', () => {
  it('should render user message content', () => {
    render(<ChatMessage sender="user" content="Hello world" />);
    expect(screen.getByText('Hello world')).toBeInTheDocument();
  });

  it('should render assistant message content', () => {
    render(<ChatMessage sender="assistant" content="Hi there" />);
    expect(screen.getByText('Hi there')).toBeInTheDocument();
  });

  it('should apply user message styling', () => {
    const { container } = render(<ChatMessage sender="user" content="Test" />);
    expect(container.querySelector('.chat-message-user')).toBeInTheDocument();
  });

  it('should apply assistant message styling', () => {
    const { container } = render(<ChatMessage sender="assistant" content="Test" />);
    expect(container.querySelector('.chat-message-assistant')).toBeInTheDocument();
  });

  it('should render markdown in assistant messages', () => {
    render(<ChatMessage sender="assistant" content="**bold text**" />);
    expect(screen.getByText('bold text')).toBeInTheDocument();
    const strong = screen.getByText('bold text').closest('strong');
    expect(strong).toBeInTheDocument();
  });

  it('should show streaming cursor when isStreaming is true', () => {
    const { container } = render(<ChatMessage sender="assistant" content="Loading" isStreaming />);
    expect(container.querySelector('.chat-cursor')).toBeInTheDocument();
  });

  it('should not show streaming cursor when isStreaming is false', () => {
    const { container } = render(<ChatMessage sender="assistant" content="Done" />);
    expect(container.querySelector('.chat-cursor')).not.toBeInTheDocument();
  });
});
