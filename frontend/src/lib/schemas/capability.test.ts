import { describe, it, expect } from 'vitest';
import {
  capabilityNameSchema,
  capabilityDescriptionSchema,
  createCapabilitySchema,
  editCapabilitySchema,
  addTagSchema,
  addExpertSchema,
} from './capability';

describe('capabilityNameSchema', () => {
  it('should accept valid names', () => {
    const result = capabilityNameSchema.safeParse('Valid Name');
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data).toBe('Valid Name');
    }
  });

  it('should trim whitespace', () => {
    const result = capabilityNameSchema.safeParse('  Trimmed Name  ');
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data).toBe('Trimmed Name');
    }
  });

  it('should reject empty strings', () => {
    const result = capabilityNameSchema.safeParse('');
    expect(result.success).toBe(false);
  });

  it('should reject whitespace-only strings', () => {
    const result = capabilityNameSchema.safeParse('   ');
    expect(result.success).toBe(false);
  });

  it('should reject names exceeding 200 characters', () => {
    const longName = 'a'.repeat(201);
    const result = capabilityNameSchema.safeParse(longName);
    expect(result.success).toBe(false);
  });

  it('should accept names at exactly 200 characters', () => {
    const exactName = 'a'.repeat(200);
    const result = capabilityNameSchema.safeParse(exactName);
    expect(result.success).toBe(true);
  });
});

describe('capabilityDescriptionSchema', () => {
  it('should accept valid descriptions', () => {
    const result = capabilityDescriptionSchema.safeParse('Valid description');
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data).toBe('Valid description');
    }
  });

  it('should accept empty strings', () => {
    const result = capabilityDescriptionSchema.safeParse('');
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data).toBe('');
    }
  });

  it('should trim whitespace', () => {
    const result = capabilityDescriptionSchema.safeParse('  Trimmed  ');
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data).toBe('Trimmed');
    }
  });

  it('should reject descriptions exceeding 1000 characters', () => {
    const longDesc = 'a'.repeat(1001);
    const result = capabilityDescriptionSchema.safeParse(longDesc);
    expect(result.success).toBe(false);
  });

  it('should accept descriptions at exactly 1000 characters', () => {
    const exactDesc = 'a'.repeat(1000);
    const result = capabilityDescriptionSchema.safeParse(exactDesc);
    expect(result.success).toBe(true);
  });
});

describe('createCapabilitySchema', () => {
  const schema = createCapabilitySchema({ min: 0, max: 24 });
  const validData = {
    name: 'Test Capability',
    description: 'Test description',
    status: 'Active',
    maturityValue: 12,
  };

  it('should accept valid data', () => {
    const result = schema.safeParse(validData);
    expect(result.success).toBe(true);
  });

  it('should reject missing name', () => {
    const { name, ...withoutName } = validData;
    const result = schema.safeParse(withoutName);
    expect(result.success).toBe(false);
  });

  it('should reject invalid maturity values', () => {
    const result = schema.safeParse({
      ...validData,
      maturityValue: 25,
    });
    expect(result.success).toBe(false);
  });

  it('should reject negative maturity values', () => {
    const result = schema.safeParse({
      ...validData,
      maturityValue: -1,
    });
    expect(result.success).toBe(false);
  });

  it('should accept maturity value at boundaries', () => {
    const resultMin = schema.safeParse({
      ...validData,
      maturityValue: 0,
    });
    expect(resultMin.success).toBe(true);

    const resultMax = schema.safeParse({
      ...validData,
      maturityValue: 24,
    });
    expect(resultMax.success).toBe(true);
  });
});

describe('editCapabilitySchema', () => {
  const schema = editCapabilitySchema();
  const validData = {
    name: 'Test Capability',
    description: 'Test description',
    status: 'Active',
    maturityValue: 12,
    ownershipModel: 'TribeOwned',
    primaryOwner: 'John Doe',
    eaOwner: 'user-123',
  };

  it('should accept valid data', () => {
    const result = schema.safeParse(validData);
    expect(result.success).toBe(true);
  });

  it('should accept empty ownership fields', () => {
    const result = schema.safeParse({
      ...validData,
      ownershipModel: '',
      primaryOwner: '',
      eaOwner: '',
    });
    expect(result.success).toBe(true);
  });

  it('should trim primaryOwner', () => {
    const result = schema.safeParse({
      ...validData,
      primaryOwner: '  John Doe  ',
    });
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.primaryOwner).toBe('John Doe');
    }
  });
});

describe('addTagSchema', () => {
  it('should accept valid tags', () => {
    const result = addTagSchema.safeParse({ tag: 'my-tag' });
    expect(result.success).toBe(true);
  });

  it('should trim whitespace', () => {
    const result = addTagSchema.safeParse({ tag: '  my-tag  ' });
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.tag).toBe('my-tag');
    }
  });

  it('should reject empty tags', () => {
    const result = addTagSchema.safeParse({ tag: '' });
    expect(result.success).toBe(false);
  });

  it('should reject whitespace-only tags', () => {
    const result = addTagSchema.safeParse({ tag: '   ' });
    expect(result.success).toBe(false);
  });

  it('should reject tags exceeding 100 characters', () => {
    const result = addTagSchema.safeParse({ tag: 'a'.repeat(101) });
    expect(result.success).toBe(false);
  });
});

describe('addExpertSchema', () => {
  const validData = {
    name: 'John Doe',
    role: 'Senior Engineer',
    contact: 'john@example.com',
  };

  it('should accept valid data', () => {
    const result = addExpertSchema.safeParse(validData);
    expect(result.success).toBe(true);
  });

  it('should reject empty name', () => {
    const result = addExpertSchema.safeParse({ ...validData, name: '' });
    expect(result.success).toBe(false);
  });

  it('should reject empty role', () => {
    const result = addExpertSchema.safeParse({ ...validData, role: '' });
    expect(result.success).toBe(false);
  });

  it('should reject empty contact', () => {
    const result = addExpertSchema.safeParse({ ...validData, contact: '' });
    expect(result.success).toBe(false);
  });

  it('should reject whitespace-only fields', () => {
    const result = addExpertSchema.safeParse({
      name: '   ',
      role: '   ',
      contact: '   ',
    });
    expect(result.success).toBe(false);
  });

  it('should trim all fields', () => {
    const result = addExpertSchema.safeParse({
      name: '  John  ',
      role: '  Engineer  ',
      contact: '  john@test.com  ',
    });
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.name).toBe('John');
      expect(result.data.role).toBe('Engineer');
      expect(result.data.contact).toBe('john@test.com');
    }
  });

  it('should reject names exceeding 200 characters', () => {
    const result = addExpertSchema.safeParse({
      ...validData,
      name: 'a'.repeat(201),
    });
    expect(result.success).toBe(false);
  });

  it('should reject roles exceeding 200 characters', () => {
    const result = addExpertSchema.safeParse({
      ...validData,
      role: 'a'.repeat(201),
    });
    expect(result.success).toBe(false);
  });

  it('should reject contact exceeding 500 characters', () => {
    const result = addExpertSchema.safeParse({
      ...validData,
      contact: 'a'.repeat(501),
    });
    expect(result.success).toBe(false);
  });
});
