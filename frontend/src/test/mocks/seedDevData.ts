import { toBusinessDomainId, toCapabilityId, toComponentId } from '../../api/types';
import {
  buildBusinessDomain,
  buildCapabilityAt,
  buildCapabilityRealization,
  buildComponent,
  buildView,
} from '../helpers/entityBuilders';
import { seedDb } from './db';
import { buildStubJourney } from './spec182/builders';
import { seedSpec182Db } from './spec182/store';

interface DevCapability {
  id: string;
  name: string;
  level: 'L1' | 'L2' | 'L3' | 'L4';
  parentId: string | null;
}

const devCapabilities: DevCapability[] = [
  { id: 'cap-account-management', name: 'Customer Account Management', level: 'L1', parentId: null },
  { id: 'cap-account-creation', name: 'Customer Account Creation', level: 'L2', parentId: 'cap-account-management' },
  { id: 'cap-identity-verification', name: 'Identity Verification', level: 'L3', parentId: 'cap-account-creation' },
  { id: 'cap-account-recovery', name: 'Account Recovery', level: 'L3', parentId: 'cap-account-creation' },
  { id: 'cap-account-closure', name: 'Customer Account Closure', level: 'L2', parentId: 'cap-account-management' },
  { id: 'cap-fraud-prevention', name: 'Customer Fraud Prevention', level: 'L2', parentId: null },
  { id: 'cap-credential-issuance', name: 'Credential Issuance', level: 'L4', parentId: null },
];

const devJourneys = [
  buildStubJourney({
    id: 'journey-account-management',
    capabilityId: 'cap-account-creation',
    capabilityName: 'Customer Account Creation',
    status: 'in-flight',
    progress: 35,
    milestones: [
      {
        id: 'ms-api-live',
        label: 'Phoenix booking API live',
        targetPeriod: { year: 2025, quarter: 4 },
        status: 'done',
      },
      {
        id: 'ms-routes',
        label: 'Channel & Baltic routes migrated',
        targetPeriod: { year: 2026, quarter: 1 },
        status: 'done',
      },
      {
        id: 'ms-north-sea',
        label: 'North Sea corridor migrated',
        targetPeriod: { year: 2026, quarter: 4 },
        status: 'in-flight',
      },
      { id: 'ms-readonly', label: 'Seabook read-only', targetPeriod: { year: 2027, quarter: 1 }, status: 'planned' },
    ],
  }),
];

export function seedDevData(): void {
  seedSpec182Db({ journeys: devJourneys, canWrite: true });
  seedDb({
    businessDomains: [buildBusinessDomain({ id: toBusinessDomainId('bd-access'), name: 'Access Domain' })],
    capabilities: devCapabilities.map((cap) =>
      buildCapabilityAt(cap.id, cap.name, cap.level, cap.parentId ?? undefined),
    ),
    components: [
      buildComponent({
        id: toComponentId('comp-phoenix'),
        name: 'Phoenix',
        description: 'Booking platform for passenger and freight routes.',
      }),
    ],
    capabilityRealizations: [
      buildCapabilityRealization({
        capabilityId: toCapabilityId('cap-account-creation'),
        componentId: toComponentId('comp-phoenix'),
        componentName: 'Phoenix',
      }),
    ],
    views: [buildView({ name: 'Default View', isDefault: true })],
  });
}
