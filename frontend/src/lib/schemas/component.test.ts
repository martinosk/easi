import { describe, it, expect } from 'vitest';
import {
  componentNameSchema,
  componentDescriptionSchema,
  createComponentSchema,
  editComponentSchema,
} from './component';

describe('componentNameSchema', () => {
  it('should accept valid names', () => {
    const result = componentNameSchema.safeParse('User Service');
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data).toBe('User Service');
    }
  });

  it('should trim whitespace', () => {
    const result = componentNameSchema.safeParse('  Order Service  ');
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data).toBe('Order Service');
    }
  });

  it('should reject empty strings', () => {
    const result = componentNameSchema.safeParse('');
    expect(result.success).toBe(false);
  });

  it('should reject whitespace-only strings', () => {
    const result = componentNameSchema.safeParse('   ');
    expect(result.success).toBe(false);
  });

  it('should reject names exceeding 200 characters', () => {
    const longName = 'a'.repeat(201);
    const result = componentNameSchema.safeParse(longName);
    expect(result.success).toBe(false);
  });

  it('should accept names at exactly 200 characters', () => {
    const exactName = 'a'.repeat(200);
    const result = componentNameSchema.safeParse(exactName);
    expect(result.success).toBe(true);
  });
});

describe('componentDescriptionSchema', () => {
  it('should accept valid descriptions', () => {
    const result = componentDescriptionSchema.safeParse('A service that handles user data');
    expect(result.success).toBe(true);
  });

  it('should accept empty strings', () => {
    const result = componentDescriptionSchema.safeParse('');
    expect(result.success).toBe(true);
  });

  it('should trim whitespace', () => {
    const result = componentDescriptionSchema.safeParse('  Description  ');
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data).toBe('Description');
    }
  });

  it('should reject descriptions exceeding 1000 characters', () => {
    const longDesc = 'a'.repeat(1001);
    const result = componentDescriptionSchema.safeParse(longDesc);
    expect(result.success).toBe(false);
  });

  it('should accept descriptions at exactly 1000 characters', () => {
    const exactDesc = 'a'.repeat(1000);
    const result = componentDescriptionSchema.safeParse(exactDesc);
    expect(result.success).toBe(true);
  });
});

describe('createComponentSchema', () => {
  it('should accept valid data', () => {
    const result = createComponentSchema.safeParse({
      name: 'User Service',
      description: 'Handles user authentication',
    });
    expect(result.success).toBe(true);
  });

  it('should accept empty description', () => {
    const result = createComponentSchema.safeParse({
      name: 'User Service',
      description: '',
    });
    expect(result.success).toBe(true);
  });

  it('should reject missing name', () => {
    const result = createComponentSchema.safeParse({
      description: 'Some description',
    });
    expect(result.success).toBe(false);
  });

  it('should reject empty name', () => {
    const result = createComponentSchema.safeParse({
      name: '',
      description: 'Some description',
    });
    expect(result.success).toBe(false);
  });

  it('should trim all fields', () => {
    const result = createComponentSchema.safeParse({
      name: '  User Service  ',
      description: '  Description  ',
    });
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.name).toBe('User Service');
      expect(result.data.description).toBe('Description');
    }
  });
});

describe('editComponentSchema', () => {
  it('should accept valid data', () => {
    const result = editComponentSchema.safeParse({
      name: 'Updated Service',
      description: 'Updated description',
    });
    expect(result.success).toBe(true);
  });

  it('should reject empty name', () => {
    const result = editComponentSchema.safeParse({
      name: '',
      description: 'Some description',
    });
    expect(result.success).toBe(false);
  });

  it('should accept empty description', () => {
    const result = editComponentSchema.safeParse({
      name: 'Service',
      description: '',
    });
    expect(result.success).toBe(true);
  });
});
