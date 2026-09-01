import { HttpResponse, http } from 'msw';
import type { ComponentStatistics } from '../../api/types';

const zeroStatistics: ComponentStatistics = {
  unknown: 0,
  nominated: 0,
  owned: 0,
  managed: 0,
  hosting: { 'on-premises': 0, cloud: 0, saas: 0, 'third-party-hosted': 0, unknown: 0 },
  total: 0,
};

let statistics: ComponentStatistics = zeroStatistics;

export function seedComponentStatistics(value: Partial<ComponentStatistics>): void {
  statistics = { ...zeroStatistics, ...value };
}

export function resetComponentStatistics(): void {
  statistics = zeroStatistics;
}

export const componentStatisticsHandlers = [
  http.get('*/api/v1/components/ownership-statistics', () =>
    HttpResponse.json({
      ...statistics,
      _links: { self: { href: '/api/v1/components/ownership-statistics', method: 'GET' } },
    }),
  ),
];
