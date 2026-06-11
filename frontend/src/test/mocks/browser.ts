import { setupWorker } from 'msw/browser';
import { devHandlers } from './devHandlers';
import { handlers } from './handlers';

export const worker = setupWorker(...handlers, ...devHandlers);
