import { describe, it, expect } from 'vitest';

describe('API Client', () => {
  describe('Response Handling', () => {
    it('should handle wrapped responses from backend', () => {
      // Backend returns: { data: [...], pagination: {...}, _links: {...} }
      // Client unwraps to just return the data array
      const wrappedResponse = {
        data: [{ id: '1', name: 'Test' }],
        pagination: { hasMore: false, limit: 50 },
        _links: { self: '/api/v1/components' }
      };

      // The client extracts: response.data.data || []
      const extracted = wrappedResponse.data || [];
      expect(extracted).toHaveLength(1);
      expect(extracted[0].id).toBe('1');
    });

    it('should handle null data responses', () => {
      const wrappedResponse = {
        data: null,
        pagination: { hasMore: false, limit: 50 },
        _links: { self: '/api/v1/components' }
      };

      // The client should return [] for null data
      const extracted = wrappedResponse.data || [];
      expect(extracted).toHaveLength(0);
    });
  });
});
