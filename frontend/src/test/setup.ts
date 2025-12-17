import '@testing-library/jest-dom/vitest';
import { afterEach, beforeAll } from 'vitest';
import { cleanup } from '@testing-library/react';

beforeAll(() => {
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

  const originalGetComputedStyle = window.getComputedStyle;
  window.getComputedStyle = function(element: Element) {
    const styles = originalGetComputedStyle.call(this, element);
    const htmlElement = element as HTMLElement;

    if (htmlElement.style && htmlElement.style.border) {
      return new Proxy(styles, {
        get(target, prop) {
          if (prop === 'border') {
            return htmlElement.style.border;
          }
          return target[prop as any];
        }
      });
    }

    return styles;
  };
});

afterEach(() => {
  cleanup();
});
