import { httpClient } from '../../../api/core/httpClient';
import type { Conversation, ConversationDetail, ConversationListResponse, SendMessageRequest } from './types';

const BASE = '/api/v1/assistant/conversations';

export const chatApi = {
  async createConversation(): Promise<Conversation> {
    const response = await httpClient.post<Conversation>(BASE);
    return response.data;
  },

  async listConversations(limit?: number, offset?: number): Promise<ConversationListResponse> {
    const params = new URLSearchParams();
    if (limit) params.set('limit', String(limit));
    if (offset) params.set('offset', String(offset));
    const query = params.toString();
    const url = query ? `${BASE}?${query}` : BASE;
    const response = await httpClient.get<ConversationListResponse>(url);
    return response.data;
  },

  async getConversation(id: string): Promise<ConversationDetail> {
    const response = await httpClient.get<ConversationDetail>(`${BASE}/${id}`);
    return response.data;
  },

  async deleteConversation(id: string): Promise<void> {
    await httpClient.delete(`${BASE}/${id}`);
  },

  sendMessageStream(conversationId: string, request: SendMessageRequest): Promise<Response> {
    const baseURL = import.meta.env.VITE_API_URL ?? '';
    return fetch(`${baseURL}${BASE}/${conversationId}/messages`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(request),
    });
  },
};
