import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import type { ChatMessage } from '../api/types';
import { ChatPanel } from './ChatPanel';

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

interface RenderPanelOptions {
  onClose?: () => void;
  writeAvailable?: boolean;
}

function renderPanel(isOpen: boolean, { onClose = vi.fn(), writeAvailable = true }: RenderPanelOptions = {}) {
  renderWithProviders(<ChatPanel isOpen={isOpen} onClose={onClose} writeAvailable={writeAvailable} />, {
    withRouter: false,
  });
}

async function renderOpenPanel(options: RenderPanelOptions = {}) {
  renderPanel(true, options);
  return screen.findByRole('dialog');
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
      controller.enqueue(
        encoder.encode(
          'event: token\ndata: {"content":"Hello"}\n\nevent: done\ndata: {"messageId":"msg-1","tokensUsed":5}\n\n',
        ),
      );
      controller.close();
    },
  });
  vi.mocked(chatApi.sendMessageStream).mockResolvedValue(
    new Response(stream, { status: 200, headers: { 'Content-Type': 'text/event-stream' } }),
  );
}

function mockStoredConversation(messages: ChatMessage[]) {
  const now = new Date().toISOString();
  vi.mocked(chatApi.listConversations).mockResolvedValue({
    data: [{ id: 'conv-old', title: 'Old chat', createdAt: now, _links: {} }],
    _links: {},
  });
  vi.mocked(chatApi.getConversation).mockResolvedValue({
    id: 'conv-old',
    title: 'Old chat',
    createdAt: now,
    lastMessageAt: now,
    _links: {},
    messages: messages.map((m) => ({ ...m, createdAt: now })),
  });
}

async function openStoredConversation() {
  fireEvent.click(screen.getByLabelText('Conversation history'));
  await waitFor(() => {
    expect(screen.getByText('Old chat')).toBeInTheDocument();
  });
  await act(async () => {
    fireEvent.click(screen.getByText('Old chat'));
  });
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
    vi.mocked(chatApi.listConversations).mockResolvedValue({ data: [], _links: {} });
  });

  it('should not render a dialog when isOpen is false', () => {
    renderPanel(false);
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('should render a dialog when isOpen is true', async () => {
    const dialog = await renderOpenPanel();
    expect(dialog).toBeInTheDocument();
  });

  it('should name the dialog for assistive technology', async () => {
    const dialog = await renderOpenPanel();
    expect(dialog).toHaveAccessibleName('Chat panel');
  });

  it('should leave the page behind interactive by rendering no overlay', async () => {
    await renderOpenPanel();
    expect(document.querySelector('.mantine-Drawer-overlay')).toBeNull();
  });

  it('should render header with title', async () => {
    await renderOpenPanel();
    expect(screen.getByText('Architecture Assistant')).toBeInTheDocument();
  });

  it('should render close button', async () => {
    await renderOpenPanel();
    expect(screen.getByLabelText('Close chat')).toBeInTheDocument();
  });

  it('should call onClose when close button is clicked', async () => {
    const onClose = vi.fn();
    await renderOpenPanel({ onClose });
    fireEvent.click(screen.getByLabelText('Close chat'));
    expect(onClose).toHaveBeenCalled();
  });

  it('should call onClose when Escape is pressed', async () => {
    const onClose = vi.fn();
    await renderOpenPanel({ onClose });
    fireEvent.keyDown(document.body, { key: 'Escape' });
    expect(onClose).toHaveBeenCalled();
  });

  it('should show prompt suggestions in empty state', async () => {
    await renderOpenPanel();
    expect(screen.getByText('What applications are in the Finance domain?')).toBeInTheDocument();
    expect(screen.getByText('Show me a portfolio summary')).toBeInTheDocument();
  });

  it('should render chat input', async () => {
    await renderOpenPanel();
    expect(screen.getByPlaceholderText('Ask about your architecture...')).toBeInTheDocument();
  });

  it('should create conversation on first send and send message', async () => {
    mockConversationAndStream('conv-1');
    await renderOpenPanel();
    await typeAndSendMessage('Hello');

    await waitFor(() => {
      expect(chatApi.createConversation).toHaveBeenCalled();
    });
  });

  it('should render YOLO checkbox', async () => {
    await renderOpenPanel();
    expect(screen.getByLabelText('YOLO (allow changes)')).toBeInTheDocument();
  });

  it('should load conversation messages when selecting a previous conversation', async () => {
    mockStoredConversation([
      { id: 'msg-1', role: 'user', content: 'What apps exist?' },
      { id: 'msg-2', role: 'assistant', content: 'There are 3 apps.' },
    ]);
    await renderOpenPanel();

    await openStoredConversation();

    await waitFor(() => {
      expect(chatApi.getConversation).toHaveBeenCalledWith('conv-old');
      expect(screen.getByText('What apps exist?')).toBeInTheDocument();
      expect(screen.getByText('There are 3 apps.')).toBeInTheDocument();
    });
  });

  it('should clear messages when starting a new conversation', async () => {
    mockStoredConversation([{ id: 'msg-1', role: 'user', content: 'Previous question' }]);
    await renderOpenPanel();

    await openStoredConversation();
    await waitFor(() => {
      expect(screen.getByText('Previous question')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByLabelText('Conversation history'));
    fireEvent.click(screen.getByLabelText('New conversation'));

    await waitFor(() => {
      expect(screen.queryByText('Previous question')).not.toBeInTheDocument();
      expect(screen.getByText('How can I help with your architecture?')).toBeInTheDocument();
    });
  });

  it('should pass yoloEnabled as allowWriteOperations when sending message', async () => {
    mockConversationAndStream('conv-yolo');
    await renderOpenPanel();

    fireEvent.click(screen.getByLabelText('YOLO (allow changes)'));
    await typeAndSendMessage('Create app');

    await waitFor(() => {
      expect(chatApi.sendMessageStream).toHaveBeenCalledWith('conv-yolo', {
        content: 'Create app',
        allowWriteOperations: true,
      });
    });
  });

  it('should not render YOLO checkbox when write is unavailable', async () => {
    await renderOpenPanel({ writeAvailable: false });
    expect(screen.queryByLabelText('YOLO (allow changes)')).not.toBeInTheDocument();
  });

  it('should send allowWriteOperations=false when write is unavailable', async () => {
    mockConversationAndStream('conv-readonly');
    await renderOpenPanel({ writeAvailable: false });

    await typeAndSendMessage('What apps exist?');

    await waitFor(() => {
      expect(chatApi.sendMessageStream).toHaveBeenCalledWith('conv-readonly', {
        content: 'What apps exist?',
        allowWriteOperations: false,
      });
    });
  });
});
