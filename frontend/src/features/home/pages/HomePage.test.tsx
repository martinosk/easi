import { act, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { HttpResponse, http } from 'msw';
import { describe, expect, it } from 'vitest';
import { renderWithProviders } from '../../../test/helpers';
import { buildHomeFixture, seedHome } from '../../../test/mocks/home';
import { server } from '../../../test/mocks/server';
import { homeQueryKeys } from '../queryKeys';
import type { HomeResponse, MyWorkItem } from '../types';
import { HomePage } from './HomePage';

function seed(overrides: Partial<HomeResponse>): void {
  seedHome({ ...buildHomeFixture(), ...overrides });
}

function renderHome() {
  return renderWithProviders(<HomePage />);
}

async function waitForHome(): Promise<void> {
  await screen.findByRole('heading', { name: 'Home' });
}

function capabilityItem(index: number): MyWorkItem {
  return {
    subjectType: 'capability',
    id: `c${index}`,
    name: `Capability ${String(index).padStart(2, '0')}`,
    level: 'L1',
    relation: 'ea-owner',
  };
}

describe('HomePage', () => {
  describe('loading and errors', () => {
    it('shows a loader while the home loads', () => {
      renderHome();

      expect(screen.getByText('Loading your home...')).toBeInTheDocument();
    });

    it('shows an error alert with a retry that reloads the home', async () => {
      server.use(
        http.get('*/api/v1/home', () => HttpResponse.json({ message: 'boom' }, { status: 500 }), { once: true }),
      );
      const user = userEvent.setup();
      renderHome();

      const alert = await screen.findByRole('alert');
      expect(alert).toHaveTextContent('Could not load your home');

      await user.click(within(alert).getByRole('button', { name: 'Retry' }));

      expect(await screen.findByRole('heading', { name: 'Home' })).toBeInTheDocument();
    });

    it('keeps the rendered home when a background refetch fails', async () => {
      const { queryClient } = renderHome();
      await waitForHome();
      server.use(http.get('*/api/v1/home', () => HttpResponse.json({ message: 'boom' }, { status: 500 })));

      await act(async () => {
        await queryClient.refetchQueries({ queryKey: homeQueryKeys.all });
        await new Promise((resolve) => setTimeout(resolve, 0));
      });

      expect(queryClient.getQueryState(homeQueryKeys.detail())?.status).toBe('error');
      expect(screen.getByTestId('home-scope')).toHaveTextContent('Steward of Assessment in Customer Engagement');
      expect(screen.getAllByTestId('my-work-card')).toHaveLength(3);
      expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });
  });

  describe('scope', () => {
    it('reads the personal scope of the caller', async () => {
      renderHome();
      await waitForHome();

      const scope = screen.getByTestId('home-scope');
      expect(scope).toHaveTextContent('Steward of Assessment in Customer Engagement');
      expect(scope).toHaveTextContent('Domain architect of Finance');
      expect(scope).toHaveTextContent('EA owner of 1 capability');
      expect(scope).toHaveTextContent('Owner of 2 applications');
      expect(scope).toHaveTextContent('Edit access to 1 capability or application');
    });

    it('reads counts when domain names are absent', async () => {
      const fixture = buildHomeFixture();
      seed({
        scope: {
          ...fixture.scope,
          stewardships: [{ concern: 'assessment', concernLabel: 'Assessment' }],
          architectedDomains: undefined,
          architectedDomainCount: 2,
        },
      });
      renderHome();
      await waitForHome();

      const scope = screen.getByTestId('home-scope');
      expect(scope).toHaveTextContent('Steward of Assessment');
      expect(scope).not.toHaveTextContent('Customer Engagement');
      expect(scope).toHaveTextContent('Domain architect of 2 domains');
    });

    it('collapses stewardships of one concern into a count when domain names are absent', async () => {
      const fixture = buildHomeFixture();
      seed({
        scope: {
          ...fixture.scope,
          stewardships: [
            { concern: 'assessment', concernLabel: 'Assessment' },
            { concern: 'ownership', concernLabel: 'Ownership' },
            { concern: 'assessment', concernLabel: 'Assessment' },
          ],
          architectedDomains: undefined,
          architectedDomainCount: 0,
        },
      });
      renderHome();
      await waitForHome();

      const statements = within(screen.getByTestId('home-scope'))
        .getAllByRole('listitem')
        .map((item) => item.textContent);
      expect(statements.filter((text) => text?.startsWith('Steward of'))).toEqual([
        'Steward of Assessment in 2 domains',
        'Steward of Ownership',
      ]);
    });

    it('explains the tenant scope', async () => {
      seed({
        scope: {
          kind: 'tenant',
          stewardships: [],
          architectedDomainCount: 0,
          eaOwnedCapabilities: 0,
          ownedApplications: 0,
          editGrants: 0,
        },
        myWork: { total: 0, items: [] },
      });
      renderHome();
      await waitForHome();

      expect(screen.getByTestId('home-scope')).toHaveTextContent(
        'No personal scope is set, so the whole portfolio is shown.',
      );
      expect(screen.getByText('Nothing is assigned to you personally.')).toBeInTheDocument();
    });

    it('guides a caller with an empty scope and shows no tiles or My Work', async () => {
      seed({
        scope: {
          kind: 'empty',
          stewardships: [],
          architectedDomainCount: 0,
          eaOwnedCapabilities: 0,
          ownedApplications: 0,
          editGrants: 0,
        },
        portfolio: undefined,
        myWork: undefined,
      });
      renderHome();
      await waitForHome();

      expect(
        screen.getByText(
          'Home fills when you are made a steward, own an application, or are granted edit access to a capability or application.',
        ),
      ).toBeInTheDocument();
      expect(screen.queryByTestId('home-tiles')).not.toBeInTheDocument();
      expect(screen.queryByRole('heading', { name: 'My Work' })).not.toBeInTheDocument();
    });
  });

  describe('portfolio tiles', () => {
    it('shows capabilities per status, applications and domains', async () => {
      renderHome();
      await waitForHome();

      const capabilities = screen.getByTestId('home-tile-capabilities');
      expect(capabilities).toHaveTextContent('12');
      expect(capabilities).toHaveTextContent('Active 10');
      expect(capabilities).toHaveTextContent('Planned 1');
      expect(capabilities).toHaveTextContent('Deprecated 1');
      expect(screen.getByTestId('home-tile-applications')).toHaveTextContent('9');
      const domains = screen.getByTestId('home-tile-domains');
      expect(domains).toHaveTextContent('1');
      expect(domains).toHaveTextContent('Customer Engagement');
    });

    it('names up to five domains and counts the rest', async () => {
      const fixture = buildHomeFixture();
      seed({
        portfolio: {
          ...fixture.portfolio,
          domains: { total: 7, names: ['Alpha', 'Beta', 'Gamma', 'Delta', 'Epsilon'] },
        },
      });
      renderHome();
      await waitForHome();

      const domains = screen.getByTestId('home-tile-domains');
      expect(domains).toHaveTextContent('Alpha, Beta, Gamma, Delta, Epsilon and 2 more');
    });

    it('shows only the domain count when names are absent', async () => {
      const fixture = buildHomeFixture();
      seed({ portfolio: { ...fixture.portfolio, domains: { total: 3 } } });
      renderHome();
      await waitForHome();

      const domains = screen.getByTestId('home-tile-domains');
      expect(domains).toHaveTextContent('3');
      expect(domains).not.toHaveTextContent('more');
    });

    it('shows the TIME distribution with a percentage legend', async () => {
      renderHome();
      await waitForHome();

      const time = screen.getByTestId('home-tile-time');
      expect(time).toHaveTextContent('40 realisations');
      expect(within(time).getByText('Invest 50%')).toBeInTheDocument();
      expect(within(time).getByText('Tolerate 20%')).toBeInTheDocument();
      expect(within(time).getByText('Migrate 10%')).toBeInTheDocument();
      expect(within(time).getByText('Eliminate 5%')).toBeInTheDocument();
      expect(within(time).getByText('Not assessed 15%')).toBeInTheDocument();
    });

    it('states that there are no realisations to assess when the TIME total is 0', async () => {
      const fixture = buildHomeFixture();
      seed({ portfolio: { ...fixture.portfolio, time: { total: 0 } } });
      renderHome();
      await waitForHome();

      const time = screen.getByTestId('home-tile-time');
      expect(time).toHaveTextContent('No realisations to assess');
      expect(within(time).queryByText(/Invest/)).not.toBeInTheDocument();
    });

    it('omits every tile absent from the response', async () => {
      seed({ portfolio: { applications: { total: 9 } } });
      renderHome();
      await waitForHome();

      expect(screen.getByTestId('home-tile-applications')).toBeInTheDocument();
      expect(screen.queryByTestId('home-tile-capabilities')).not.toBeInTheDocument();
      expect(screen.queryByTestId('home-tile-domains')).not.toBeInTheDocument();
      expect(screen.queryByTestId('home-tile-time')).not.toBeInTheDocument();
    });
  });

  describe('My Work', () => {
    it('links capability and application cards to their one-pagers', async () => {
      renderHome();
      await waitForHome();

      expect(screen.getByRole('link', { name: /Customer Account Creation/ })).toHaveAttribute(
        'href',
        '/one-pagers/capability/cap-account-creation',
      );
      expect(screen.getByRole('link', { name: /Phoenix/ })).toHaveAttribute(
        'href',
        '/one-pagers/application/comp-phoenix',
      );
    });

    it('does not make a card clickable without its one-pager link', async () => {
      const fixture = buildHomeFixture();
      const [accountCreation, ...rest] = fixture.myWork?.items ?? [];
      seed({ myWork: { total: 3, items: [{ ...accountCreation, _links: {} }, ...rest] } });
      renderHome();
      await waitForHome();

      expect(screen.getByText('Customer Account Creation')).toBeInTheDocument();
      expect(screen.queryByRole('link', { name: /Customer Account Creation/ })).not.toBeInTheDocument();
    });

    it('shows level, relation, grant expiry and dominant grade', async () => {
      renderHome();
      await waitForHome();

      const accountCreation = screen.getByRole('link', { name: /Customer Account Creation/ });
      expect(accountCreation).toHaveTextContent('L2');
      expect(accountCreation).toHaveTextContent('EA owner');
      expect(within(accountCreation).getByRole('img', { name: 'Invest' })).toHaveTextContent('I');

      const phoenix = screen.getByRole('link', { name: /Phoenix/ });
      expect(phoenix).toHaveTextContent('Edit access until 21 Oct 2026');
      expect(within(phoenix).getByRole('img', { name: 'Eliminate' })).toHaveTextContent('E');

      const seabook = screen.getByRole('link', { name: /Seabook/ });
      expect(seabook).toHaveTextContent('Nominated owner');
      expect(within(seabook).queryByRole('img')).not.toBeInTheDocument();
    });

    it('shows the first 12 cards and reveals the rest on Show all', async () => {
      const items = Array.from({ length: 20 }, (_, i) => capabilityItem(i + 1));
      seed({ myWork: { total: 20, items } });
      const user = userEvent.setup();
      renderHome();
      await waitForHome();

      expect(screen.getAllByTestId('my-work-card')).toHaveLength(12);

      await user.click(screen.getByRole('button', { name: 'Show all 20' }));

      expect(screen.getAllByTestId('my-work-card')).toHaveLength(20);
      expect(screen.queryByRole('button', { name: /Show all/ })).not.toBeInTheDocument();
    });

    it('offers no Show all control for 12 cards or fewer', async () => {
      renderHome();
      await waitForHome();

      expect(screen.getAllByTestId('my-work-card')).toHaveLength(3);
      expect(screen.queryByRole('button', { name: /Show all/ })).not.toBeInTheDocument();
    });

    it('omits My Work when the response carries none', async () => {
      seed({ myWork: undefined });
      renderHome();
      await waitForHome();

      expect(screen.queryByRole('heading', { name: 'My Work' })).not.toBeInTheDocument();
    });
  });
});
