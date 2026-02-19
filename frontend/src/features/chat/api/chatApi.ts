import { httpClient } from '../../../api/core';
import type { Conversation, SendMessageRequest } from './types';

const BASE = '/api/v1/assistant/conversations';

export const chatApi = {
  async createConversation(): Promise<Conversation> {
    const response = await httpClient.post<Conversation>(BASE);
    return response.data;
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
