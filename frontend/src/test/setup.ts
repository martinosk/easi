import '@testing-library/jest-dom/vitest';
import { act, cleanup } from '@testing-library/react';
import { afterAll, afterEach, beforeAll, beforeEach, expect } from 'vitest';
import { resetAssistantStatus } from './mocks/assistantStatus';
import { resetDb } from './mocks/db';
import { resetOnePagerCompleteness } from './mocks/onePagerCompleteness';
import { server } from './mocks/server';
import { resetSpec180Db } from './mocks/spec180/store';
import { resetSpec181Db } from './mocks/spec181/store';
import { resetSpec182Db } from './mocks/spec182/store';

const originalConsoleError = console.error;
let actWarnings: string[] = [];

console.error = (...args: unknown[]) => {
  const message = typeof args[0] === 'string' ? args[0] : '';
  if (message.includes('was not wrapped in act')) {
    actWarnings.push(message);
  }
  originalConsoleError.apply(console, args);
};

beforeAll(() => {
  server.listen({ onUnhandledRequest: 'error' });
  if (!HTMLDialogElement.prototype.showModal) {
    HTMLDialogElement.prototype.showModal = function () {
      this.open = true;
    };
  }

  if (!HTMLDialogElement.prototype.close) {
    HTMLDialogElement.prototype.close = function () {
      this.open = false;
    };
  }

  if (!HTMLDialogElement.prototype.show) {
    HTMLDialogElement.prototype.show = function () {
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

  if (typeof window.confirm === 'undefined') {
    window.confirm = () => true;
  }
});

beforeEach(() => {
  resetAssistantStatus();
  resetDb();
  resetOnePagerCompleteness();
  resetSpec180Db();
  resetSpec181Db();
  resetSpec182Db();
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
        capturedWarnings[0].substring(0, 200),
    );
  }
});

afterAll(() => {
  server.close();
});
