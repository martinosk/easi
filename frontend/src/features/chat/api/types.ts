export interface Conversation {
  id: string;
  title: string;
  createdAt: string;
  _links: Record<string, { href: string; method: string }>;
}

export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
}

export interface SendMessageRequest {
  content: string;
}
