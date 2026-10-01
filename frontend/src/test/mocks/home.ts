import { HttpResponse, http } from 'msw';
import type { HATEOASLinks } from '../../api/types';
import type { HomeResponse, MyWorkItem } from '../../features/home/types';

function onePagerLinks(subjectType: MyWorkItem['subjectType'], id: string): HATEOASLinks {
  return { 'x-one-pager': { href: `/api/v1/one-pagers/${subjectType}/${id}`, method: 'GET' } };
}

export function buildHomeFixture(): HomeResponse {
  return {
    scope: {
      kind: 'personal',
      stewardships: [
        { domain: { id: 'd1', name: 'Customer Engagement' }, concern: 'assessment', concernLabel: 'Assessment' },
      ],
      architectedDomains: [{ id: 'd2', name: 'Finance' }],
      architectedDomainCount: 1,
      eaOwnedCapabilities: 1,
      ownedApplications: 2,
      editGrants: 1,
    },
    portfolio: {
      capabilities: { total: 12, byStatus: { active: 10, planned: 1, deprecated: 1 } },
      applications: { total: 9 },
      domains: { total: 1, names: ['Customer Engagement'] },
      time: { total: 40, shares: { invest: 50, tolerate: 20, migrate: 10, eliminate: 5, notAssessed: 15 } },
    },
    myWork: {
      total: 3,
      items: [
        {
          subjectType: 'capability',
          id: 'cap-account-creation',
          name: 'Customer Account Creation',
          level: 'L2',
          relation: 'ea-owner',
          dominantGrade: 'invest',
          _links: onePagerLinks('capability', 'cap-account-creation'),
        },
        {
          subjectType: 'application',
          id: 'comp-seabook',
          name: 'Seabook',
          relation: 'nominated',
          _links: onePagerLinks('application', 'comp-seabook'),
        },
        {
          subjectType: 'application',
          id: 'comp-phoenix',
          name: 'Phoenix',
          relation: 'edit-grant',
          grantExpiresOn: '2026-10-21',
          dominantGrade: 'eliminate',
          _links: onePagerLinks('application', 'comp-phoenix'),
        },
      ],
    },
    _links: { self: { href: '/api/v1/home', method: 'GET' } },
  };
}

let homeResponse: HomeResponse = buildHomeFixture();

export function seedHome(response: HomeResponse): void {
  homeResponse = response;
}

export function resetHome(): void {
  homeResponse = buildHomeFixture();
}

export const homeHandlers = [http.get('*/api/v1/home', () => HttpResponse.json(homeResponse))];
