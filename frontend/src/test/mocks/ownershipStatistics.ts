import { HttpResponse, http } from 'msw';
import type { OwnershipStatistics } from '../../api/types';

const zeroStatistics: OwnershipStatistics = { unknown: 0, nominated: 0, owned: 0, managed: 0, total: 0 };

let statistics: OwnershipStatistics = zeroStatistics;

export function seedOwnershipStatistics(value: OwnershipStatistics): void {
  statistics = value;
}

export function resetOwnershipStatistics(): void {
  statistics = zeroStatistics;
}

export const ownershipStatisticsHandlers = [
  http.get('*/api/v1/components/ownership-statistics', () =>
    HttpResponse.json({
      ...statistics,
      _links: { self: { href: '/api/v1/components/ownership-statistics', method: 'GET' } },
    }),
  ),
];
