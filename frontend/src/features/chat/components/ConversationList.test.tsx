import type { ComponentProps, ReactElement } from 'react';
import { fireEvent, screen } from '@testing-library/react';
import { renderWithProviders } from '../../../test/helpers';
const render = (ui: ReactElement) => renderWithProviders(ui, { withRouter: false });
import { describe, expect, it, vi } from 'vitest';
import type { Conversation } from '../api/types';
import { ConversationList } from './ConversationList';

function buildConversation(overrides: Partial<Conversation> = {}): Conversation {
  return {
    id: 'conv-1',
    title: 'Test conversation',
    createdAt: '2026-01-15T10:00:00Z',
    _links: {},
    ...overrides,
  };
}

type ListProps = ComponentProps<typeof ConversationList>;

function renderList(overrides: Partial<ListProps> = {}) {
  const props: ListProps = {
    conversations: [],
    activeConversationId: null,
    onSelect: vi.fn(),
    onDelete: vi.fn(),
    onNewConversation: vi.fn(),
    ...overrides,
  };
  render(<ConversationList {...props} />);
  return props;
}

describe('ConversationList', () => {
  it('should render conversation titles', () => {
    renderList({
      conversations: [
        buildConversation({ id: 'conv-1', title: 'Architecture review' }),
        buildConversation({ id: 'conv-2', title: 'Portfolio analysis' }),
      ],
    });

    expect(screen.getByText('Architecture review')).toBeInTheDocument();
    expect(screen.getByText('Portfolio analysis')).toBeInTheDocument();
  });

  it('should call onSelect when a conversation is clicked', () => {
    const { onSelect } = renderList({
      conversations: [buildConversation({ id: 'conv-1', title: 'Architecture review' })],
    });

    fireEvent.click(screen.getByText('Architecture review'));
    expect(onSelect).toHaveBeenCalledWith('conv-1');
  });

  it('should highlight the active conversation', () => {
    renderList({
      conversations: [
        buildConversation({ id: 'conv-1', title: 'Active chat' }),
        buildConversation({ id: 'conv-2', title: 'Other chat' }),
      ],
      activeConversationId: 'conv-1',
    });

    const [activeItem, otherItem] = screen.getAllByTestId('conversation-item');
    expect(activeItem).toHaveTextContent('Active chat');
    expect(activeItem).toHaveAttribute('data-active');
    expect(otherItem).not.toHaveAttribute('data-active');
  });

  it('should call onDelete when delete button is clicked', () => {
    const { onDelete } = renderList({
      conversations: [buildConversation({ id: 'conv-1', title: 'Delete me' })],
    });

    fireEvent.click(screen.getByLabelText('Delete conversation'));
    expect(onDelete).toHaveBeenCalledWith('conv-1');
  });

  it('should not trigger onSelect when delete button is clicked', () => {
    const { onSelect } = renderList({
      conversations: [buildConversation({ id: 'conv-1', title: 'Test' })],
    });

    fireEvent.click(screen.getByLabelText('Delete conversation'));
    expect(onSelect).not.toHaveBeenCalled();
  });

  it('should render new conversation button', () => {
    renderList();

    expect(screen.getByLabelText('New conversation')).toBeInTheDocument();
  });

  it('should call onNewConversation when new button is clicked', () => {
    const { onNewConversation } = renderList();

    fireEvent.click(screen.getByLabelText('New conversation'));
    expect(onNewConversation).toHaveBeenCalled();
  });

  it('should show empty state when no conversations', () => {
    renderList();

    expect(screen.getByText('No conversations yet')).toBeInTheDocument();
  });
});
