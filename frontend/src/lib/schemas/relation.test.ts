import { describe, expect, it } from 'vitest';
import { createRelationSchema, realizationNotesSchema, relationTypeSchema } from './relation';

describe('relationTypeSchema', () => {
  it('should accept Triggers', () => {
    const result = relationTypeSchema.safeParse('Triggers');
    expect(result.success).toBe(true);
  });

  it('should accept Serves', () => {
    const result = relationTypeSchema.safeParse('Serves');
    expect(result.success).toBe(true);
  });

  it('should reject invalid types', () => {
    const result = relationTypeSchema.safeParse('InvalidType');
    expect(result.success).toBe(false);
  });

  it('should reject empty string', () => {
    const result = relationTypeSchema.safeParse('');
    expect(result.success).toBe(false);
  });
});

describe('createRelationSchema', () => {
  const validData = {
    sourceComponentId: 'source-123',
    targetComponentId: 'target-456',
    relationType: 'Triggers' as const,
    name: 'API Call',
    description: 'Calls the target API',
  };

  it('should accept valid data', () => {
    const result = createRelationSchema.safeParse(validData);
    expect(result.success).toBe(true);
  });

  it('should accept empty name and description', () => {
    const result = createRelationSchema.safeParse({
      ...validData,
      name: '',
      description: '',
    });
    expect(result.success).toBe(true);
  });

  it('should reject same source and target', () => {
    const result = createRelationSchema.safeParse({
      ...validData,
      sourceComponentId: 'same-id',
      targetComponentId: 'same-id',
    });
    expect(result.success).toBe(false);
    if (!result.success) {
      const errorMessages = result.error.issues.map((e) => e.message);
      expect(errorMessages).toContain('Source and target components must be different');
    }
  });

  it('should reject empty source component', () => {
    const result = createRelationSchema.safeParse({
      ...validData,
      sourceComponentId: '',
    });
    expect(result.success).toBe(false);
  });

  it('should reject empty target component', () => {
    const result = createRelationSchema.safeParse({
      ...validData,
      targetComponentId: '',
    });
    expect(result.success).toBe(false);
  });

  it('should reject invalid relation type', () => {
    const result = createRelationSchema.safeParse({
      ...validData,
      relationType: 'Invalid',
    });
    expect(result.success).toBe(false);
  });

  it('should trim name and description', () => {
    const result = createRelationSchema.safeParse({
      ...validData,
      name: '  API Call  ',
      description: '  Description  ',
    });
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.name).toBe('API Call');
      expect(result.data.description).toBe('Description');
    }
  });

  it('should reject name exceeding 200 characters', () => {
    const result = createRelationSchema.safeParse({
      ...validData,
      name: 'a'.repeat(201),
    });
    expect(result.success).toBe(false);
  });

  it('should reject description exceeding 1000 characters', () => {
    const result = createRelationSchema.safeParse({
      ...validData,
      description: 'a'.repeat(1001),
    });
    expect(result.success).toBe(false);
  });
});

describe('realizationNotesSchema', () => {
  it('accepts empty notes so they can be cleared, and trims', () => {
    expect(realizationNotesSchema.safeParse('').success).toBe(true);
    expect(realizationNotesSchema.parse('  keep  ')).toBe('keep');
  });

  it('rejects notes over 1000 characters', () => {
    expect(realizationNotesSchema.safeParse('x'.repeat(1001)).success).toBe(false);
  });
});
