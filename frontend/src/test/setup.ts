import '@testing-library/jest-dom/vitest';
import { afterEach, beforeAll, afterAll, beforeEach, expect } from 'vitest';
import { cleanup, act } from '@testing-library/react';
import { server } from './mocks/server';
import { resetDb } from './mocks/db';

const originalConsoleError = console.error;
let actWarnings: string[] = [];

console.error = (...args: unknown[]) => {
  const message = typeof args[0] === 'string' ? args[0] : '';
  if (message.includes('was not wrapped in act')) {
    actWarnings.push(message);
  }
  originalConsoleError.apply(console, args);
};

process.on('uncaughtException', (error: Error) => {
  if (error.message?.includes('EINVAL') || (error as NodeJS.ErrnoException).code === 'EINVAL') {
    return;
  }
  throw error;
});

beforeAll(() => {
  server.listen({ onUnhandledRequest: 'bypass' });
  if (!HTMLDialogElement.prototype.showModal) {
    HTMLDialogElement.prototype.showModal = function() {
      this.open = true;
    };
  }

  if (!HTMLDialogElement.prototype.close) {
    HTMLDialogElement.prototype.close = function() {
      this.open = false;
    };
  }

  if (!HTMLDialogElement.prototype.show) {
    HTMLDialogElement.prototype.show = function() {
      this.open = true;
    };
  }

  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: (query: string) => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: () => {},
      removeListener: () => {},
      addEventListener: () => {},
      removeEventListener: () => {},
      dispatchEvent: () => true,
    }),
  });

  global.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  };
});

beforeEach(() => {
  resetDb();
  actWarnings = [];
});

afterEach(async () => {
  await act(async () => {
    cleanup();
  });
  server.resetHandlers();

  if (actWarnings.length > 0) {
    const warningCount = actWarnings.length;
    const testName = expect.getState().currentTestName || 'Unknown test';
    const capturedWarnings = [...actWarnings];
    actWarnings = [];
    throw new Error(
      `Test "${testName}" caused ${warningCount} act warning(s). ` +
      'State updates must be wrapped in act(). First warning:\n' +
      capturedWarnings[0].substring(0, 200)
    );
  }
});

afterAll(() => {
  server.close();
});
