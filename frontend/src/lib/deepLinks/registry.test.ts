import { afterEach, describe, expect, it, vi } from 'vitest';
import { deepLinkParams, getParamValue, readDeepLink } from './registry';

function setLocation(pathname: string, search: string) {
  Object.defineProperty(window, 'location', {
    value: { pathname, search },
    writable: true,
    configurable: true,
  });
}

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

  describe('readDeepLink', () => {
    afterEach(() => {
      vi.unstubAllEnvs();
    });

    it('reads the view parameter on the canvas', () => {
      setLocation('/canvas', '?view=v1');

      expect(readDeepLink(deepLinkParams.VIEW)).toBe('v1');
    });

    it.each(['/', '/business-domains', '/canvas-archive'])('ignores the view parameter on %s', (pathname) => {
      setLocation(pathname, '?view=v1');

      expect(readDeepLink(deepLinkParams.VIEW)).toBeNull();
    });

    it('reads a parameter on a page below its registered route', () => {
      setLocation('/business-domains/domain-1', '?capability=cap-1');

      expect(readDeepLink(deepLinkParams.CAPABILITY)).toBe('cap-1');
    });

    it('ignores the domain parameter on the canvas', () => {
      setLocation('/canvas', '?domain=domain-1');

      expect(readDeepLink(deepLinkParams.DOMAIN)).toBeNull();
    });

    it('matches registered routes below the application base path', () => {
      vi.stubEnv('BASE_URL', '/easi/');
      setLocation('/easi/canvas', '?view=v1');

      expect(readDeepLink(deepLinkParams.VIEW)).toBe('v1');
    });

    it('returns null when the parameter is absent on its route', () => {
      setLocation('/canvas', '');

      expect(readDeepLink(deepLinkParams.VIEW)).toBeNull();
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
