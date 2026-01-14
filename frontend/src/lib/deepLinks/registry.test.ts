import { describe, it, expect, afterEach } from 'vitest';
import { getParamValue } from './registry';

function setLocationSearch(search: string) {
  Object.defineProperty(window, 'location', {
    value: { search },
    writable: true,
    configurable: true,
  });
}

describe('deepLinks registry', () => {
  const originalLocation = window.location;

  afterEach(() => {
    Object.defineProperty(window, 'location', {
      value: originalLocation,
      writable: true,
      configurable: true,
    });
  });

  describe('getParamValue', () => {
    it('should return param value when present', () => {
      setLocationSearch('?view=test-view-id');

      expect(getParamValue('view')).toBe('test-view-id');
    });

    it('should return null when param is not present', () => {
      setLocationSearch('');

      expect(getParamValue('view')).toBeNull();
    });

    it('should handle URL-encoded values (single decode only)', () => {
      setLocationSearch('?returnUrl=https%3A%2F%2Fexample.com');

      expect(getParamValue('returnUrl')).toBe('https://example.com');
    });

    it('should NOT double-decode values (security)', () => {
      setLocationSearch('?returnUrl=https%253A%252F%252Fevil.com');

      expect(getParamValue('returnUrl')).toBe('https%3A%2F%2Fevil.com');
    });
  });
});
