import { describe, expect, it } from 'vitest';
import { pluralSubjectTypeLabel } from './subjectTypes';

describe('pluralSubjectTypeLabel', () => {
  it('pluralizes application as Applications', () => {
    expect(pluralSubjectTypeLabel('application')).toBe('Applications');
  });

  it('pluralizes vendor as Vendors', () => {
    expect(pluralSubjectTypeLabel('vendor')).toBe('Vendors');
  });

  it('pluralizes capability as Capabilities', () => {
    expect(pluralSubjectTypeLabel('capability')).toBe('Capabilities');
  });

  it('pluralizes enterprise-capability as Enterprise Capabilities', () => {
    expect(pluralSubjectTypeLabel('enterprise-capability')).toBe('Enterprise Capabilities');
  });

  it('pluralizes acquired-entity as Acquired Entities', () => {
    expect(pluralSubjectTypeLabel('acquired-entity')).toBe('Acquired Entities');
  });

  it('pluralizes internal-team as Internal Teams', () => {
    expect(pluralSubjectTypeLabel('internal-team')).toBe('Internal Teams');
  });
});
