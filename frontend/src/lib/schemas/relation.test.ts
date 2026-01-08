import { describe, it, expect } from 'vitest';
import {
  relationTypeSchema,
  createRelationSchema,
  editRelationSchema,
  editRealizationSchema,
} from './relation';

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

describe('editRelationSchema', () => {
  it('should accept valid data', () => {
    const result = editRelationSchema.safeParse({
      name: 'Updated Name',
      description: 'Updated description',
    });
    expect(result.success).toBe(true);
  });

  it('should accept empty name and description', () => {
    const result = editRelationSchema.safeParse({
      name: '',
      description: '',
    });
    expect(result.success).toBe(true);
  });

  it('should trim fields', () => {
    const result = editRelationSchema.safeParse({
      name: '  Name  ',
      description: '  Description  ',
    });
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.name).toBe('Name');
      expect(result.data.description).toBe('Description');
    }
  });
});

describe('editRealizationSchema', () => {
  it('should accept Full level', () => {
    const result = editRealizationSchema.safeParse({
      realizationLevel: 'Full',
      notes: 'Some notes',
    });
    expect(result.success).toBe(true);
  });

  it('should accept Partial level', () => {
    const result = editRealizationSchema.safeParse({
      realizationLevel: 'Partial',
      notes: '',
    });
    expect(result.success).toBe(true);
  });

  it('should accept Planned level', () => {
    const result = editRealizationSchema.safeParse({
      realizationLevel: 'Planned',
      notes: '',
    });
    expect(result.success).toBe(true);
  });

  it('should reject invalid level', () => {
    const result = editRealizationSchema.safeParse({
      realizationLevel: 'Invalid',
      notes: '',
    });
    expect(result.success).toBe(false);
  });

  it('should trim notes', () => {
    const result = editRealizationSchema.safeParse({
      realizationLevel: 'Full',
      notes: '  Some notes  ',
    });
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.notes).toBe('Some notes');
    }
  });

  it('should reject notes exceeding 1000 characters', () => {
    const result = editRealizationSchema.safeParse({
      realizationLevel: 'Full',
      notes: 'a'.repeat(1001),
    });
    expect(result.success).toBe(false);
  });
});
