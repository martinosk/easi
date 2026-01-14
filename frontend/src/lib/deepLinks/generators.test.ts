import { describe, it, expect, afterEach } from 'vitest';
import { generateViewShareUrl, generateDomainShareUrl } from './generators';

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
    it('should generate URL with view parameter', () => {
      setLocationOrigin('https://app.example.com');

      const url = generateViewShareUrl('view-123');

      expect(url).toBe('https://app.example.com/?view=view-123');
    });

    it('should URL-encode special characters in view ID', () => {
      setLocationOrigin('https://app.example.com');

      const url = generateViewShareUrl('view with spaces');

      expect(url).toBe('https://app.example.com/?view=view+with+spaces');
    });
  });

  describe('generateDomainShareUrl', () => {
    it('should generate URL with domain parameter and correct path', () => {
      setLocationOrigin('https://app.example.com');

      const url = generateDomainShareUrl('domain-456');

      expect(url).toBe('https://app.example.com/business-domains?domain=domain-456');
    });
  });
});
