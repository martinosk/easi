import { HttpResponse, http } from 'msw';

let configured = false;
let canWrite = false;

export function seedAssistantStatus(status: { configured: boolean; canWrite?: boolean }): void {
  configured = status.configured;
  canWrite = status.canWrite ?? false;
}

export function resetAssistantStatus(): void {
  configured = false;
  canWrite = false;
}

export const assistantStatusHandlers = [
  http.get('*/api/v1/assistant/status', () => {
    return HttpResponse.json({
      configured,
      _links: {
        self: { href: '/api/v1/assistant/status', method: 'GET' },
        ...(configured ? { 'x-conversations': { href: '/api/v1/assistant/conversations', method: 'GET' } } : {}),
        ...(configured && canWrite
          ? { 'x-conversations-write': { href: '/api/v1/assistant/conversations', method: 'GET' } }
          : {}),
      },
    });
  }),
];
