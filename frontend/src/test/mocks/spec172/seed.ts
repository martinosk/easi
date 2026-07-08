import { buildCapabilityAt, buildView } from '../../helpers/entityBuilders';
import { seedDb } from '../db';
import type { StubCapability } from './composition';
import type { StubDirection, StubEnterpriseCapability } from './store';
import { seedSpec172Db } from './store';

const devEnterpriseCapabilities: StubEnterpriseCapability[] = [
  {
    id: 'ec-customer-identity',
    name: 'Customer Identity',
    description: 'Single source of truth for recognising and authenticating customers across channels.',
    category: 'Customer Domain',
    active: true,
    createdAt: '2026-01-04T09:00:00Z',
  },
  {
    id: 'ec-take-payment',
    name: 'Take Payment',
    description: 'Accept and settle customer payments across products and regions.',
    category: 'Customer Domain',
    active: true,
    createdAt: '2026-01-04T09:00:00Z',
  },
  {
    id: 'ec-identity-platform',
    name: 'Identity Platform',
    description: 'Cross-cutting identity and access services.',
    category: 'Access Domain',
    active: true,
    createdAt: '2026-02-10T09:00:00Z',
  },
  {
    id: 'ec-order-management',
    name: 'Order Management',
    description: 'Order capture, orchestration and fulfilment.',
    category: 'Commerce Domain',
    active: true,
    createdAt: '2026-01-20T09:00:00Z',
  },
];

const devDirections: StubDirection[] = [
  {
    id: 'dir-customer-identity',
    enterpriseCapabilityId: 'ec-customer-identity',
    type: 'consolidate',
    status: 'proposed',
    horizon: 'next',
    narrative:
      'Fold the scattered account-creation and verification capabilities into one identity backbone owned by the Customer domain.',
    sourceCapabilityIds: ['cap-account-management', 'cap-credential-issuance'],
    createdAt: '2026-03-01T10:00:00Z',
  },
  {
    id: 'dir-take-payment',
    enterpriseCapabilityId: 'ec-take-payment',
    type: 'consolidate',
    status: 'proposed',
    horizon: 'now',
    narrative: 'Consolidate fraud and recovery handling under Take Payment.',
    sourceCapabilityIds: ['cap-account-recovery', 'cap-fraud-prevention'],
    createdAt: '2026-03-02T10:00:00Z',
  },
  {
    id: 'dir-order-management',
    enterpriseCapabilityId: 'ec-order-management',
    type: 'consolidate',
    status: 'agreed',
    horizon: 'now',
    narrative: 'Identity backbone agreed by the architecture group; source set is now frozen.',
    sourceCapabilityIds: ['cap-account-closure'],
    createdAt: '2026-02-14T10:00:00Z',
  },
];

const devCapabilities: StubCapability[] = [
  {
    id: 'cap-account-management',
    name: 'Customer Account Management',
    level: 'L1',
    parentId: null,
    businessDomainId: 'bd-customer',
    businessDomainName: 'Customer Domain',
  },
  {
    id: 'cap-account-creation',
    name: 'Customer Account Creation',
    level: 'L2',
    parentId: 'cap-account-management',
    businessDomainId: 'bd-customer',
    businessDomainName: 'Customer Domain',
  },
  {
    id: 'cap-identity-verification',
    name: 'Identity Verification',
    level: 'L3',
    parentId: 'cap-account-creation',
    businessDomainId: 'bd-customer',
    businessDomainName: 'Customer Domain',
  },
  {
    id: 'cap-account-recovery',
    name: 'Account Recovery',
    level: 'L3',
    parentId: 'cap-account-creation',
    businessDomainId: 'bd-customer',
    businessDomainName: 'Customer Domain',
  },
  {
    id: 'cap-account-closure',
    name: 'Customer Account Closure',
    level: 'L2',
    parentId: 'cap-account-management',
    businessDomainId: 'bd-customer',
    businessDomainName: 'Customer Domain',
  },
  {
    id: 'cap-fraud-prevention',
    name: 'Customer Fraud Prevention',
    level: 'L2',
    parentId: null,
    businessDomainId: 'bd-customer',
    businessDomainName: 'Customer Domain',
  },
  {
    id: 'cap-credential-issuance',
    name: 'Credential Issuance',
    level: 'L4',
    parentId: null,
    businessDomainId: 'bd-access',
    businessDomainName: 'Access Domain',
  },
];

export function seedDevData(): void {
  seedSpec172Db({
    enterpriseCapabilities: devEnterpriseCapabilities,
    directions: devDirections,
    capabilities: devCapabilities,
  });
  seedDb({
    capabilities: devCapabilities.map((cap) =>
      buildCapabilityAt(cap.id, cap.name, cap.level, cap.parentId ?? undefined),
    ),
    views: [buildView({ name: 'Default View', isDefault: true })],
  });
}
