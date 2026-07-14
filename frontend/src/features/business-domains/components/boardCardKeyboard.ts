import type { KeyboardEvent, MouseEvent } from 'react';

export function activationKeyHandler<T>(item: T, onActivate: (item: T, event: MouseEvent) => void) {
  return (event: KeyboardEvent) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      onActivate(item, event as unknown as MouseEvent);
    }
  };
}
