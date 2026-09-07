import { describe, expect, it } from 'vitest';
import type { ZodTypeAny } from 'zod';
import {
  createAcquiredEntitySchema,
  createInternalTeamSchema,
  createVendorSchema,
  integrationStatusSchema,
  internalTeamContactPersonSchema,
  internalTeamDepartmentSchema,
  originEntityNotesSchema,
  vendorImplementationPartnerSchema,
} from './originEntity';

const expectTrimsField = (
  schema: ZodTypeAny,
  input: Record<string, unknown>,
  field: string,
  expected: string,
): void => {
  const result = schema.safeParse(input);
  expect(result.success).toBe(true);
  if (result.success) {
    expect((result.data as Record<string, unknown>)[field]).toBe(expected);
  }
};

describe('createAcquiredEntitySchema', () => {
  describe('name validation', () => {
    it('should accept valid name', () => {
      const result = createAcquiredEntitySchema.safeParse({
        name: 'TechCorp',
      });
      expect(result.success).toBe(true);
    });

    it('should trim whitespace from name', () => {
      expectTrimsField(createAcquiredEntitySchema, { name: '  TechCorp  ' }, 'name', 'TechCorp');
    });

    it('should reject empty name', () => {
      const result = createAcquiredEntitySchema.safeParse({
        name: '',
      });
      expect(result.success).toBe(false);
    });

    it('should reject whitespace-only name', () => {
      const result = createAcquiredEntitySchema.safeParse({
        name: '   ',
      });
      expect(result.success).toBe(false);
    });

    it('should reject name exceeding 100 characters', () => {
      const result = createAcquiredEntitySchema.safeParse({
        name: 'a'.repeat(101),
      });
      expect(result.success).toBe(false);
    });

    it('should accept name at exactly 100 characters', () => {
      const result = createAcquiredEntitySchema.safeParse({
        name: 'a'.repeat(100),
      });
      expect(result.success).toBe(true);
    });
  });

  describe('acquisitionDate validation', () => {
    it('should accept valid date string', () => {
      const result = createAcquiredEntitySchema.safeParse({
        name: 'TechCorp',
        acquisitionDate: '2021-03-15',
      });
      expect(result.success).toBe(true);
    });

    it('should accept empty acquisition date', () => {
      const result = createAcquiredEntitySchema.safeParse({
        name: 'TechCorp',
        acquisitionDate: undefined,
      });
      expect(result.success).toBe(true);
    });
  });

  describe('integrationStatus validation', () => {
    it('should accept NotStarted status', () => {
      const result = createAcquiredEntitySchema.safeParse({
        name: 'TechCorp',
        integrationStatus: 'NotStarted',
      });
      expect(result.success).toBe(true);
    });

    it('should accept InProgress status', () => {
      const result = createAcquiredEntitySchema.safeParse({
        name: 'TechCorp',
        integrationStatus: 'InProgress',
      });
      expect(result.success).toBe(true);
    });

    it('should accept Completed status', () => {
      const result = createAcquiredEntitySchema.safeParse({
        name: 'TechCorp',
        integrationStatus: 'Completed',
      });
      expect(result.success).toBe(true);
    });

    it('should accept OnHold status', () => {
      const result = createAcquiredEntitySchema.safeParse({
        name: 'TechCorp',
        integrationStatus: 'OnHold',
      });
      expect(result.success).toBe(true);
    });

    it('should reject invalid status', () => {
      const result = createAcquiredEntitySchema.safeParse({
        name: 'TechCorp',
        integrationStatus: 'InvalidStatus',
      });
      expect(result.success).toBe(false);
    });

    it('should accept undefined status', () => {
      const result = createAcquiredEntitySchema.safeParse({
        name: 'TechCorp',
      });
      expect(result.success).toBe(true);
    });
  });

  describe('notes validation', () => {
    it('should accept valid notes', () => {
      const result = createAcquiredEntitySchema.safeParse({
        name: 'TechCorp',
        notes: 'Cloud infrastructure company acquired for platform expansion.',
      });
      expect(result.success).toBe(true);
    });

    it('should trim whitespace from notes', () => {
      expectTrimsField(
        createAcquiredEntitySchema,
        { name: 'TechCorp', notes: '  Some notes  ' },
        'notes',
        'Some notes',
      );
    });

    it('should reject notes exceeding 500 characters', () => {
      const result = createAcquiredEntitySchema.safeParse({
        name: 'TechCorp',
        notes: 'a'.repeat(501),
      });
      expect(result.success).toBe(false);
    });

    it('should accept notes at exactly 500 characters', () => {
      const result = createAcquiredEntitySchema.safeParse({
        name: 'TechCorp',
        notes: 'a'.repeat(500),
      });
      expect(result.success).toBe(true);
    });
  });
});

describe('createVendorSchema', () => {
  describe('name validation', () => {
    it('should accept valid name', () => {
      const result = createVendorSchema.safeParse({
        name: 'SAP',
      });
      expect(result.success).toBe(true);
    });

    it('should trim whitespace from name', () => {
      expectTrimsField(createVendorSchema, { name: '  SAP  ' }, 'name', 'SAP');
    });

    it('should reject empty name', () => {
      const result = createVendorSchema.safeParse({
        name: '',
      });
      expect(result.success).toBe(false);
    });

    it('should reject whitespace-only name', () => {
      const result = createVendorSchema.safeParse({
        name: '   ',
      });
      expect(result.success).toBe(false);
    });

    it('should reject name exceeding 100 characters', () => {
      const result = createVendorSchema.safeParse({
        name: 'a'.repeat(101),
      });
      expect(result.success).toBe(false);
    });
  });

  describe('implementationPartner validation', () => {
    it('should accept valid implementation partner', () => {
      const result = createVendorSchema.safeParse({
        name: 'SAP',
        implementationPartner: 'Accenture',
      });
      expect(result.success).toBe(true);
    });

    it('should trim whitespace from implementation partner', () => {
      expectTrimsField(
        createVendorSchema,
        { name: 'SAP', implementationPartner: '  Accenture  ' },
        'implementationPartner',
        'Accenture',
      );
    });

    it('should reject implementation partner exceeding 100 characters', () => {
      const result = createVendorSchema.safeParse({
        name: 'SAP',
        implementationPartner: 'a'.repeat(101),
      });
      expect(result.success).toBe(false);
    });

    it('should accept empty implementation partner', () => {
      const result = createVendorSchema.safeParse({
        name: 'SAP',
      });
      expect(result.success).toBe(true);
    });
  });

  describe('notes validation', () => {
    it('should accept valid notes', () => {
      const result = createVendorSchema.safeParse({
        name: 'SAP',
        notes: 'Enterprise ERP vendor.',
      });
      expect(result.success).toBe(true);
    });

    it('should reject notes exceeding 500 characters', () => {
      const result = createVendorSchema.safeParse({
        name: 'SAP',
        notes: 'a'.repeat(501),
      });
      expect(result.success).toBe(false);
    });
  });
});

describe('createInternalTeamSchema', () => {
  describe('name validation', () => {
    it('should accept valid name', () => {
      const result = createInternalTeamSchema.safeParse({
        name: 'Platform Engineering',
      });
      expect(result.success).toBe(true);
    });

    it('should trim whitespace from name', () => {
      expectTrimsField(createInternalTeamSchema, { name: '  Platform Engineering  ' }, 'name', 'Platform Engineering');
    });

    it('should reject empty name', () => {
      const result = createInternalTeamSchema.safeParse({
        name: '',
      });
      expect(result.success).toBe(false);
    });

    it('should reject whitespace-only name', () => {
      const result = createInternalTeamSchema.safeParse({
        name: '   ',
      });
      expect(result.success).toBe(false);
    });

    it('should reject name exceeding 100 characters', () => {
      const result = createInternalTeamSchema.safeParse({
        name: 'a'.repeat(101),
      });
      expect(result.success).toBe(false);
    });
  });

  describe('department validation', () => {
    it('should accept valid department', () => {
      const result = createInternalTeamSchema.safeParse({
        name: 'Platform Engineering',
        department: 'Technology',
      });
      expect(result.success).toBe(true);
    });

    it('should trim whitespace from department', () => {
      expectTrimsField(
        createInternalTeamSchema,
        { name: 'Platform Engineering', department: '  Technology  ' },
        'department',
        'Technology',
      );
    });

    it('should reject department exceeding 100 characters', () => {
      const result = createInternalTeamSchema.safeParse({
        name: 'Platform Engineering',
        department: 'a'.repeat(101),
      });
      expect(result.success).toBe(false);
    });

    it('should accept empty department', () => {
      const result = createInternalTeamSchema.safeParse({
        name: 'Platform Engineering',
      });
      expect(result.success).toBe(true);
    });
  });

  describe('contactPerson validation', () => {
    it('should accept valid contact person', () => {
      const result = createInternalTeamSchema.safeParse({
        name: 'Platform Engineering',
        contactPerson: 'John Doe',
      });
      expect(result.success).toBe(true);
    });

    it('should trim whitespace from contact person', () => {
      expectTrimsField(
        createInternalTeamSchema,
        { name: 'Platform Engineering', contactPerson: '  John Doe  ' },
        'contactPerson',
        'John Doe',
      );
    });

    it('should reject contact person exceeding 100 characters', () => {
      const result = createInternalTeamSchema.safeParse({
        name: 'Platform Engineering',
        contactPerson: 'a'.repeat(101),
      });
      expect(result.success).toBe(false);
    });

    it('should accept empty contact person', () => {
      const result = createInternalTeamSchema.safeParse({
        name: 'Platform Engineering',
      });
      expect(result.success).toBe(true);
    });
  });

  describe('notes validation', () => {
    it('should accept valid notes', () => {
      const result = createInternalTeamSchema.safeParse({
        name: 'Platform Engineering',
        notes: 'Responsible for core platform services.',
      });
      expect(result.success).toBe(true);
    });

    it('should reject notes exceeding 500 characters', () => {
      const result = createInternalTeamSchema.safeParse({
        name: 'Platform Engineering',
        notes: 'a'.repeat(501),
      });
      expect(result.success).toBe(false);
    });
  });
});

describe('in-place field schemas', () => {
  it('accepts an empty note so notes can be cleared, and trims', () => {
    expect(originEntityNotesSchema.safeParse('').success).toBe(true);
    expect(originEntityNotesSchema.parse('  keep  ')).toBe('keep');
  });

  it('rejects notes over 500 characters', () => {
    expect(originEntityNotesSchema.safeParse('x'.repeat(501)).success).toBe(false);
  });

  it('accepts only the API integration statuses', () => {
    expect(integrationStatusSchema.safeParse('COMPLETED').success).toBe(true);
    expect(integrationStatusSchema.safeParse('Completed').success).toBe(false);
  });

  it('bounds the optional text fields at 100 characters', () => {
    for (const schema of [
      vendorImplementationPartnerSchema,
      internalTeamDepartmentSchema,
      internalTeamContactPersonSchema,
    ]) {
      expect(schema.safeParse('').success).toBe(true);
      expect(schema.safeParse('x'.repeat(101)).success).toBe(false);
    }
  });
});
