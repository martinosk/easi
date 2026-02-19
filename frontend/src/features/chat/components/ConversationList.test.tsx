import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ConversationList } from './ConversationList';
import type { Conversation } from '../api/types';

function buildConversation(overrides: Partial<Conversation> = {}): Conversation {
  return {
    id: 'conv-1',
    title: 'Test conversation',
    createdAt: '2026-01-15T10:00:00Z',
    _links: {},
    ...overrides,
  };
}

describe('ConversationList', () => {
  it('should render conversation titles', () => {
    const conversations = [
      buildConversation({ id: 'conv-1', title: 'Architecture review' }),
      buildConversation({ id: 'conv-2', title: 'Portfolio analysis' }),
    ];

    render(
      <ConversationList
        conversations={conversations}
        activeConversationId={null}
        onSelect={vi.fn()}
        onDelete={vi.fn()}
        onNewConversation={vi.fn()}
      />
    );

    expect(screen.getByText('Architecture review')).toBeInTheDocument();
    expect(screen.getByText('Portfolio analysis')).toBeInTheDocument();
  });

  it('should call onSelect when a conversation is clicked', () => {
    const onSelect = vi.fn();
    const conversations = [
      buildConversation({ id: 'conv-1', title: 'Architecture review' }),
    ];

    render(
      <ConversationList
        conversations={conversations}
        activeConversationId={null}
        onSelect={onSelect}
        onDelete={vi.fn()}
        onNewConversation={vi.fn()}
      />
    );

    fireEvent.click(screen.getByText('Architecture review'));
    expect(onSelect).toHaveBeenCalledWith('conv-1');
  });

  it('should highlight the active conversation', () => {
    const conversations = [
      buildConversation({ id: 'conv-1', title: 'Active chat' }),
      buildConversation({ id: 'conv-2', title: 'Other chat' }),
    ];

    render(
      <ConversationList
        conversations={conversations}
        activeConversationId="conv-1"
        onSelect={vi.fn()}
        onDelete={vi.fn()}
        onNewConversation={vi.fn()}
      />
    );

    const activeItem = screen.getByText('Active chat').closest('.conversation-item');
    expect(activeItem).toHaveClass('conversation-item-active');
  });

  it('should call onDelete when delete button is clicked', () => {
    const onDelete = vi.fn();
    const conversations = [
      buildConversation({ id: 'conv-1', title: 'Delete me' }),
    ];

    render(
      <ConversationList
        conversations={conversations}
        activeConversationId={null}
        onSelect={vi.fn()}
        onDelete={onDelete}
        onNewConversation={vi.fn()}
      />
    );

    fireEvent.click(screen.getByLabelText('Delete conversation'));
    expect(onDelete).toHaveBeenCalledWith('conv-1');
  });

  it('should not trigger onSelect when delete button is clicked', () => {
    const onSelect = vi.fn();
    const conversations = [
      buildConversation({ id: 'conv-1', title: 'Test' }),
    ];

    render(
      <ConversationList
        conversations={conversations}
        activeConversationId={null}
        onSelect={onSelect}
        onDelete={vi.fn()}
        onNewConversation={vi.fn()}
      />
    );

    fireEvent.click(screen.getByLabelText('Delete conversation'));
    expect(onSelect).not.toHaveBeenCalled();
  });

  it('should render new conversation button', () => {
    render(
      <ConversationList
        conversations={[]}
        activeConversationId={null}
        onSelect={vi.fn()}
        onDelete={vi.fn()}
        onNewConversation={vi.fn()}
      />
    );

    expect(screen.getByLabelText('New conversation')).toBeInTheDocument();
  });

  it('should call onNewConversation when new button is clicked', () => {
    const onNew = vi.fn();

    render(
      <ConversationList
        conversations={[]}
        activeConversationId={null}
        onSelect={vi.fn()}
        onDelete={vi.fn()}
        onNewConversation={onNew}
      />
    );

    fireEvent.click(screen.getByLabelText('New conversation'));
    expect(onNew).toHaveBeenCalled();
  });

  it('should show empty state when no conversations', () => {
    render(
      <ConversationList
        conversations={[]}
        activeConversationId={null}
        onSelect={vi.fn()}
        onDelete={vi.fn()}
        onNewConversation={vi.fn()}
      />
    );

    expect(screen.getByText('No conversations yet')).toBeInTheDocument();
  });
});
