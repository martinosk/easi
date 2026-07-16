import { describe, expect, it } from 'vitest';
import {
  addSelectionOptionSchema,
  defineCustomFieldSchema,
  numberFieldBoundsSchema,
  onePagerFieldNameSchema,
  onePagerHelpTextSchema,
  onePagerOptionLabelSchema,
  renameCustomFieldSchema,
} from './onePagerConfiguration';

describe('onePagerFieldNameSchema', () => {
  it('trims whitespace', () => {
    const result = onePagerFieldNameSchema.safeParse('  Contract link  ');
    expect(result.success).toBe(true);
    if (result.success) expect(result.data).toBe('Contract link');
  });

  it('rejects empty names', () => {
    expect(onePagerFieldNameSchema.safeParse('   ').success).toBe(false);
  });

  it('rejects names over 100 characters', () => {
    expect(onePagerFieldNameSchema.safeParse('a'.repeat(101)).success).toBe(false);
  });

  it('accepts names at exactly 100 characters', () => {
    expect(onePagerFieldNameSchema.safeParse('a'.repeat(100)).success).toBe(true);
  });
});

describe('onePagerHelpTextSchema', () => {
  it('allows empty help text', () => {
    const result = onePagerHelpTextSchema.safeParse('');
    expect(result.success).toBe(true);
    if (result.success) expect(result.data).toBe('');
  });

  it('rejects help text over 500 characters', () => {
    expect(onePagerHelpTextSchema.safeParse('a'.repeat(501)).success).toBe(false);
  });
});

describe('onePagerOptionLabelSchema', () => {
  it('rejects empty option labels', () => {
    expect(onePagerOptionLabelSchema.safeParse('  ').success).toBe(false);
  });

  it('rejects option labels over 100 characters', () => {
    expect(onePagerOptionLabelSchema.safeParse('a'.repeat(101)).success).toBe(false);
  });
});

describe('defineCustomFieldSchema', () => {
  const base = {
    name: 'Business summary',
    fieldType: 'text' as const,
    required: false,
    helpText: '',
    options: [],
    min: '' as const,
    max: '' as const,
  };

  it('accepts a valid non-selection field with no options', () => {
    expect(defineCustomFieldSchema.safeParse(base).success).toBe(true);
  });

  it.each(['text', 'number', 'date', 'link', 'selection', 'contact-person'])('accepts field type %s', (fieldType) => {
    const options = fieldType === 'selection' ? ['Option A'] : [];
    const result = defineCustomFieldSchema.safeParse({ ...base, fieldType, options });
    expect(result.success).toBe(true);
  });

  it('rejects a selection field with no options', () => {
    const result = defineCustomFieldSchema.safeParse({ ...base, fieldType: 'selection', options: [] });
    expect(result.success).toBe(false);
  });

  it('rejects a selection field with duplicate option labels (case-insensitive)', () => {
    const result = defineCustomFieldSchema.safeParse({
      ...base,
      fieldType: 'selection',
      options: ['Cloud', 'cloud'],
    });
    expect(result.success).toBe(false);
  });

  it('accepts a selection field with distinct options', () => {
    const result = defineCustomFieldSchema.safeParse({
      ...base,
      fieldType: 'selection',
      options: ['On-prem', 'Cloud'],
    });
    expect(result.success).toBe(true);
  });

  it('rejects an empty name', () => {
    const result = defineCustomFieldSchema.safeParse({ ...base, name: '' });
    expect(result.success).toBe(false);
  });

  it('rejects an invalid field type', () => {
    const result = defineCustomFieldSchema.safeParse({ ...base, fieldType: 'checkbox' });
    expect(result.success).toBe(false);
  });

  it('rejects a payload missing the min or max keys', () => {
    const { min, max, ...withoutBounds } = base;
    expect(defineCustomFieldSchema.safeParse(withoutBounds).success).toBe(false);
  });

  it('accepts a number field with both bounds', () => {
    const result = defineCustomFieldSchema.safeParse({ ...base, fieldType: 'number', min: 0, max: 5 });
    expect(result.success).toBe(true);
  });

  it('accepts a number field with only a minimum', () => {
    const result = defineCustomFieldSchema.safeParse({ ...base, fieldType: 'number', min: 0, max: '' });
    expect(result.success).toBe(true);
  });

  it('accepts a number field with neither bound', () => {
    const result = defineCustomFieldSchema.safeParse({ ...base, fieldType: 'number', min: '', max: '' });
    expect(result.success).toBe(true);
  });

  it('rejects a number field where minimum exceeds maximum', () => {
    const result = defineCustomFieldSchema.safeParse({ ...base, fieldType: 'number', min: 10, max: 5 });
    expect(result.success).toBe(false);
  });
});

describe('numberFieldBoundsSchema', () => {
  it('accepts both bounds set', () => {
    expect(numberFieldBoundsSchema.safeParse({ min: 0, max: 5 }).success).toBe(true);
  });

  it('accepts neither bound set', () => {
    expect(numberFieldBoundsSchema.safeParse({ min: '', max: '' }).success).toBe(true);
  });

  it('rejects minimum greater than maximum', () => {
    const result = numberFieldBoundsSchema.safeParse({ min: 10, max: 5 });
    expect(result.success).toBe(false);
  });
});

describe('renameCustomFieldSchema', () => {
  it('accepts a valid rename payload', () => {
    const result = renameCustomFieldSchema.safeParse({ name: 'Contract link', helpText: 'Updated guidance' });
    expect(result.success).toBe(true);
  });

  it('rejects an empty name', () => {
    const result = renameCustomFieldSchema.safeParse({ name: '', helpText: '' });
    expect(result.success).toBe(false);
  });
});

describe('addSelectionOptionSchema', () => {
  it('accepts a valid option label', () => {
    expect(addSelectionOptionSchema.safeParse({ label: 'Hybrid' }).success).toBe(true);
  });

  it('rejects an empty option label', () => {
    expect(addSelectionOptionSchema.safeParse({ label: '' }).success).toBe(false);
  });
});
