import { HttpResponse, http } from 'msw';

let configured = false;

export function seedAssistantStatus(status: { configured: boolean }): void {
  configured = status.configured;
}

export function resetAssistantStatus(): void {
  configured = false;
}

export const assistantStatusHandlers = [
  http.get('*/api/v1/assistant/status', () => {
    return HttpResponse.json({
      configured,
      _links: {
        self: { href: '/api/v1/assistant/status', method: 'GET' },
        ...(configured ? { 'x-conversations': { href: '/api/v1/assistant/conversations', method: 'GET' } } : {}),
      },
    });
  }),
];
