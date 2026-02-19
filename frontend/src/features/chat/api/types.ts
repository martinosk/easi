export interface Conversation {
  id: string;
  title: string;
  createdAt: string;
  _links: Record<string, { href: string; method: string }>;
}

export interface ConversationListResponse {
  data: Conversation[];
  _links: Record<string, { href: string; method: string }>;
  meta?: { total?: number };
}

export interface ConversationDetail extends Conversation {
  lastMessageAt: string;
  messages: MessageResponse[];
}

export interface MessageResponse {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  createdAt: string;
}

export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
}

export interface SendMessageRequest {
  content: string;
  allowWriteOperations?: boolean;
}

export interface ToolCall {
  id: string;
  name: string;
  status: 'running' | 'completed';
  resultPreview?: string;
}
