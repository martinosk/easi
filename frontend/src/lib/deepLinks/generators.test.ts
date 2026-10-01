import { afterEach, describe, expect, it } from 'vitest';
import { generateDomainShareUrl, generateOnePagerShareUrl, generateViewPath, generateViewShareUrl } from './generators';

function setLocationOrigin(origin: string) {
  Object.defineProperty(window, 'location', {
    value: { origin },
    writable: true,
    configurable: true,
  });
}

describe('deepLinks generators', () => {
  const originalLocation = window.location;

  afterEach(() => {
    Object.defineProperty(window, 'location', {
      value: originalLocation,
      writable: true,
      configurable: true,
    });
  });

  describe('generateViewShareUrl', () => {
    it('should generate a canvas URL with view parameter', () => {
      setLocationOrigin('https://app.example.com');

      const url = generateViewShareUrl('view-123');

      expect(url).toBe('https://app.example.com/canvas?view=view-123');
    });

    it('should URL-encode special characters in view ID', () => {
      setLocationOrigin('https://app.example.com');

      const url = generateViewShareUrl('view with spaces');

      expect(url).toBe('https://app.example.com/canvas?view=view+with+spaces');
    });
  });

  describe('generateViewPath', () => {
    it('should generate a router path on the canvas with the view parameter', () => {
      expect(generateViewPath('view-123')).toBe('/canvas?view=view-123');
    });

    it('should URL-encode special characters in view ID', () => {
      expect(generateViewPath('a&b=c')).toBe('/canvas?view=a%26b%3Dc');
    });
  });

  describe('generateDomainShareUrl', () => {
    it('should generate URL with domain parameter and correct path', () => {
      setLocationOrigin('https://app.example.com');

      const url = generateDomainShareUrl('domain-456');

      expect(url).toBe('https://app.example.com/business-domains?domain=domain-456');
    });
  });

  describe('generateOnePagerShareUrl', () => {
    it('should generate a URL with the subject type and subject id in the path', () => {
      setLocationOrigin('https://app.example.com');

      const url = generateOnePagerShareUrl('vendor', 'vendor-1');

      expect(url).toBe('https://app.example.com/one-pagers/vendor/vendor-1');
    });
  });
});
