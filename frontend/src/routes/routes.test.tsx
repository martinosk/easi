import { describe, expect, it } from 'vitest';
import { ROUTES } from './routePaths';

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

  it('should define ONE_PAGERS route at /one-pagers', () => {
    expect(ROUTES.ONE_PAGERS).toBe('/one-pagers');
  });

  it('should define ONE_PAGER_DETAIL route with :subjectType and :subjectId params', () => {
    expect(ROUTES.ONE_PAGER_DETAIL).toBe('/one-pagers/:subjectType/:subjectId');
  });

  it('should have ONE_PAGER_DETAIL as a child path of ONE_PAGERS', () => {
    expect(ROUTES.ONE_PAGER_DETAIL.startsWith(ROUTES.ONE_PAGERS)).toBe(true);
  });

  it('should define ONE_PAGER_QUALITY route at /one-pager-quality', () => {
    expect(ROUTES.ONE_PAGER_QUALITY).toBe('/one-pager-quality');
  });
});
