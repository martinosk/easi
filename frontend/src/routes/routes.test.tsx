import { describe, it, expect } from 'vitest';
import { ROUTES } from './routes';

describe('Route configuration', () => {
  it('should define VALUE_STREAMS route at /value-streams', () => {
    expect(ROUTES.VALUE_STREAMS).toBe('/value-streams');
  });

  it('should define VALUE_STREAM_DETAIL route with :valueStreamId param', () => {
    expect(ROUTES.VALUE_STREAM_DETAIL).toBe('/value-streams/:valueStreamId');
  });

  it('should have VALUE_STREAM_DETAIL as a child path of VALUE_STREAMS', () => {
    expect(ROUTES.VALUE_STREAM_DETAIL.startsWith(ROUTES.VALUE_STREAMS)).toBe(true);
  });
});
