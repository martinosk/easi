import { describe, it, expect } from 'vitest';
import {
  createAcquiredEntitySchema,
  editAcquiredEntitySchema,
  createVendorSchema,
  editVendorSchema,
  createInternalTeamSchema,
  editInternalTeamSchema,
} from './originEntity';

describe('createAcquiredEntitySchema', () => {
  describe('name validation', () => {
    it('should accept valid name', () => {
      const result = createAcquiredEntitySchema.safeParse({
        name: 'TechCorp',
      });
      expect(result.success).toBe(true);
    });

    it('should trim whitespace from name', () => {
      const result = createAcquiredEntitySchema.safeParse({
        name: '  TechCorp  ',
      });
      expect(result.success).toBe(true);
      if (result.success) {
        expect(result.data.name).toBe('TechCorp');
      }
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
      const result = createAcquiredEntitySchema.safeParse({
        name: 'TechCorp',
        notes: '  Some notes  ',
      });
      expect(result.success).toBe(true);
      if (result.success) {
        expect(result.data.notes).toBe('Some notes');
      }
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

describe('editAcquiredEntitySchema', () => {
  it('should accept valid edit data', () => {
    const result = editAcquiredEntitySchema.safeParse({
      name: 'Updated TechCorp',
      acquisitionDate: '2021-06-01',
      integrationStatus: 'Completed',
      notes: 'Integration completed successfully.',
    });
    expect(result.success).toBe(true);
  });

  it('should reject empty name on edit', () => {
    const result = editAcquiredEntitySchema.safeParse({
      name: '',
    });
    expect(result.success).toBe(false);
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
      const result = createVendorSchema.safeParse({
        name: '  SAP  ',
      });
      expect(result.success).toBe(true);
      if (result.success) {
        expect(result.data.name).toBe('SAP');
      }
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
      const result = createVendorSchema.safeParse({
        name: 'SAP',
        implementationPartner: '  Accenture  ',
      });
      expect(result.success).toBe(true);
      if (result.success) {
        expect(result.data.implementationPartner).toBe('Accenture');
      }
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

describe('editVendorSchema', () => {
  it('should accept valid edit data', () => {
    const result = editVendorSchema.safeParse({
      name: 'Updated SAP',
      implementationPartner: 'Deloitte',
      notes: 'Updated implementation partner.',
    });
    expect(result.success).toBe(true);
  });

  it('should reject empty name on edit', () => {
    const result = editVendorSchema.safeParse({
      name: '',
    });
    expect(result.success).toBe(false);
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
      const result = createInternalTeamSchema.safeParse({
        name: '  Platform Engineering  ',
      });
      expect(result.success).toBe(true);
      if (result.success) {
        expect(result.data.name).toBe('Platform Engineering');
      }
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
      const result = createInternalTeamSchema.safeParse({
        name: 'Platform Engineering',
        department: '  Technology  ',
      });
      expect(result.success).toBe(true);
      if (result.success) {
        expect(result.data.department).toBe('Technology');
      }
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
      const result = createInternalTeamSchema.safeParse({
        name: 'Platform Engineering',
        contactPerson: '  John Doe  ',
      });
      expect(result.success).toBe(true);
      if (result.success) {
        expect(result.data.contactPerson).toBe('John Doe');
      }
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

describe('editInternalTeamSchema', () => {
  it('should accept valid edit data', () => {
    const result = editInternalTeamSchema.safeParse({
      name: 'Updated Platform Engineering',
      department: 'Engineering',
      contactPerson: 'Jane Smith',
      notes: 'Updated team information.',
    });
    expect(result.success).toBe(true);
  });

  it('should reject empty name on edit', () => {
    const result = editInternalTeamSchema.safeParse({
      name: '',
    });
    expect(result.success).toBe(false);
  });
});
