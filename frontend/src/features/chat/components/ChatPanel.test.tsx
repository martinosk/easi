import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, act, waitFor } from '@testing-library/react';
import { ChatPanel } from './ChatPanel';
import { createTestQueryClient, TestProviders } from '../../../test/helpers/renderWithProviders';
import type { ReactNode } from 'react';
import type { QueryClient } from '@tanstack/react-query';

vi.mock('../api/chatApi', () => ({
  chatApi: {
    createConversation: vi.fn(),
    sendMessageStream: vi.fn(),
    listConversations: vi.fn(),
    getConversation: vi.fn(),
    deleteConversation: vi.fn(),
  },
}));

import { chatApi } from '../api/chatApi';

function createWrapper(queryClient: QueryClient) {
  return function Wrapper({ children }: { children: ReactNode }) {
    return (
      <TestProviders withRouter={false} queryClient={queryClient}>
        {children}
      </TestProviders>
    );
  };
}

function renderPanel(isOpen: boolean, onClose = vi.fn(), queryClient?: QueryClient) {
  const qc = queryClient ?? createTestQueryClient();
  vi.mocked(chatApi.listConversations).mockResolvedValue({ data: [], _links: {} });
  const Wrapper = createWrapper(qc);
  return render(
    <Wrapper>
      <ChatPanel isOpen={isOpen} onClose={onClose} />
    </Wrapper>
  );
}

function mockConversationAndStream(convId: string) {
  vi.mocked(chatApi.createConversation).mockResolvedValue({
    id: convId,
    title: 'New conversation',
    createdAt: new Date().toISOString(),
    _links: {},
  });

  const encoder = new TextEncoder();
  const stream = new ReadableStream({
    start(controller) {
      controller.enqueue(encoder.encode('event: token\ndata: {"content":"Hello"}\n\nevent: done\ndata: {"messageId":"msg-1","tokensUsed":5}\n\n'));
      controller.close();
    },
  });
  vi.mocked(chatApi.sendMessageStream).mockResolvedValue(
    new Response(stream, { status: 200, headers: { 'Content-Type': 'text/event-stream' } })
  );
}

async function typeAndSendMessage(text: string) {
  const textarea = screen.getByPlaceholderText('Ask about your architecture...');
  await act(async () => {
    fireEvent.change(textarea, { target: { value: text } });
    fireEvent.keyDown(textarea, { key: 'Enter' });
  });
}

describe('ChatPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should not render when isOpen is false', () => {
    const { container } = renderPanel(false);
    expect(container.querySelector('.chat-panel')).not.toBeInTheDocument();
  });

  it('should render when isOpen is true', () => {
    const { container } = renderPanel(true);
    expect(container.querySelector('.chat-panel')).toBeInTheDocument();
  });

  it('should render header with title', () => {
    renderPanel(true);
    expect(screen.getByText('Architecture Assistant')).toBeInTheDocument();
  });

  it('should render close button', () => {
    renderPanel(true);
    expect(screen.getByLabelText('Close chat')).toBeInTheDocument();
  });

  it('should call onClose when close button is clicked', () => {
    const onClose = vi.fn();
    renderPanel(true, onClose);
    fireEvent.click(screen.getByLabelText('Close chat'));
    expect(onClose).toHaveBeenCalled();
  });

  it('should call onClose when Escape is pressed', () => {
    const onClose = vi.fn();
    renderPanel(true, onClose);
    fireEvent.keyDown(document, { key: 'Escape' });
    expect(onClose).toHaveBeenCalled();
  });

  it('should show prompt suggestions in empty state', () => {
    renderPanel(true);
    expect(screen.getByText('What applications are in the Finance domain?')).toBeInTheDocument();
    expect(screen.getByText('Show me a portfolio summary')).toBeInTheDocument();
  });

  it('should render chat input', () => {
    renderPanel(true);
    expect(screen.getByPlaceholderText('Ask about your architecture...')).toBeInTheDocument();
  });

  it('should create conversation on first send and send message', async () => {
    mockConversationAndStream('conv-1');
    renderPanel(true);
    await typeAndSendMessage('Hello');

    await waitFor(() => {
      expect(chatApi.createConversation).toHaveBeenCalled();
    });
  });

  it('should render YOLO checkbox', () => {
    renderPanel(true);
    expect(screen.getByLabelText('YOLO (allow changes)')).toBeInTheDocument();
  });

  it('should load conversation messages when selecting a previous conversation', async () => {
    const qc = createTestQueryClient();
    vi.mocked(chatApi.listConversations).mockResolvedValue({
      data: [{ id: 'conv-old', title: 'Old chat', createdAt: new Date().toISOString(), _links: {} }],
      _links: {},
    });
    vi.mocked(chatApi.getConversation).mockResolvedValue({
      id: 'conv-old',
      title: 'Old chat',
      createdAt: new Date().toISOString(),
      lastMessageAt: new Date().toISOString(),
      _links: {},
      messages: [
        { id: 'msg-1', role: 'user', content: 'What apps exist?', createdAt: new Date().toISOString() },
        { id: 'msg-2', role: 'assistant', content: 'There are 3 apps.', createdAt: new Date().toISOString() },
      ],
    });

    const Wrapper = createWrapper(qc);
    render(<Wrapper><ChatPanel isOpen={true} onClose={vi.fn()} /></Wrapper>);

    await waitFor(() => {
      expect(screen.queryByText('No conversations yet')).not.toBeInTheDocument();
    });

    fireEvent.click(screen.getByLabelText('Conversation history'));
    await waitFor(() => { expect(screen.getByText('Old chat')).toBeInTheDocument(); });

    await act(async () => { fireEvent.click(screen.getByText('Old chat')); });

    await waitFor(() => {
      expect(chatApi.getConversation).toHaveBeenCalledWith('conv-old');
      expect(screen.getByText('What apps exist?')).toBeInTheDocument();
      expect(screen.getByText('There are 3 apps.')).toBeInTheDocument();
    });
  });

  it('should clear messages when starting a new conversation', async () => {
    const qc = createTestQueryClient();
    vi.mocked(chatApi.listConversations).mockResolvedValue({
      data: [{ id: 'conv-old', title: 'Old chat', createdAt: new Date().toISOString(), _links: {} }],
      _links: {},
    });
    vi.mocked(chatApi.getConversation).mockResolvedValue({
      id: 'conv-old',
      title: 'Old chat',
      createdAt: new Date().toISOString(),
      lastMessageAt: new Date().toISOString(),
      _links: {},
      messages: [
        { id: 'msg-1', role: 'user', content: 'Previous question', createdAt: new Date().toISOString() },
      ],
    });

    const Wrapper = createWrapper(qc);
    render(<Wrapper><ChatPanel isOpen={true} onClose={vi.fn()} /></Wrapper>);

    fireEvent.click(screen.getByLabelText('Conversation history'));
    await waitFor(() => { expect(screen.getByText('Old chat')).toBeInTheDocument(); });
    await act(async () => { fireEvent.click(screen.getByText('Old chat')); });
    await waitFor(() => { expect(screen.getByText('Previous question')).toBeInTheDocument(); });

    fireEvent.click(screen.getByLabelText('Conversation history'));
    fireEvent.click(screen.getByLabelText('New conversation'));

    await waitFor(() => {
      expect(screen.queryByText('Previous question')).not.toBeInTheDocument();
      expect(screen.getByText('How can I help with your architecture?')).toBeInTheDocument();
    });
  });

  it('should pass yoloEnabled as allowWriteOperations when sending message', async () => {
    mockConversationAndStream('conv-yolo');
    renderPanel(true);

    fireEvent.click(screen.getByLabelText('YOLO (allow changes)'));
    await typeAndSendMessage('Create app');

    await waitFor(() => {
      expect(chatApi.sendMessageStream).toHaveBeenCalledWith('conv-yolo', {
        content: 'Create app',
        allowWriteOperations: true,
      });
    });
  });
});
