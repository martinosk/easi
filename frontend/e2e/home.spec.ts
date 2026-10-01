import { type APIRequestContext, expect, test } from '@playwright/test';
import { dismissReleaseNotes } from './helpers';

const API_URL = 'http://localhost:8081';
const CONCERN = 'assessment';
const BYPASS_USER_TENANT = { 'X-Tenant-ID': 'acme' };

interface Created {
  id: string;
  name: string;
}

async function currentUserId(request: APIRequestContext): Promise<string> {
  const response = await request.get(`${API_URL}/api/v1/auth/sessions/current`);
  const session = (await response.json()) as { user: { id: string } };
  return session.user.id;
}

async function create(request: APIRequestContext, path: string, data: Record<string, string>): Promise<Created> {
  const response = await request.post(`${API_URL}${path}`, { data, headers: BYPASS_USER_TENANT });
  expect(response.ok()).toBe(true);
  const created = (await response.json()) as { id: string };
  return { id: created.id, name: data.name };
}

test.describe('personal home (spec 228)', () => {
  test.beforeEach(async ({ page, request }) => {
    await page.setExtraHTTPHeaders(BYPASS_USER_TENANT);
    await dismissReleaseNotes(page, request);
  });

  test.describe('a steward', () => {
    let domain: Created;

    test.beforeEach(async ({ request }) => {
      domain = await create(request, '/api/v1/business-domains', { name: `E2E Home Domain ${Date.now()}` });
      const response = await request.put(`${API_URL}/api/v1/stewardships/${domain.id}/${CONCERN}`, {
        data: { stewardId: await currentUserId(request) },
        headers: BYPASS_USER_TENANT,
      });
      expect(response.ok()).toBe(true);
    });

    test.afterEach(async ({ request }) => {
      await request.delete(`${API_URL}/api/v1/stewardships/${domain.id}/${CONCERN}`, { headers: BYPASS_USER_TENANT });
      await request.delete(`${API_URL}/api/v1/business-domains/${domain.id}`, { headers: BYPASS_USER_TENANT });
    });

    test('opens Home at / and sees their domain', async ({ page }) => {
      await page.goto('/');

      await expect(page.getByRole('heading', { name: 'Home', level: 2 })).toBeVisible({ timeout: 30000 });
      await expect(page.getByTestId('nav-home')).toHaveAttribute('aria-current', 'page');
      await expect(page.getByTestId('home-scope')).toContainText(`Steward of Assessment in ${domain.name}`);
      await expect(page.getByTestId('home-tile-domains')).toContainText(domain.name);
      await expect(page.getByTestId('home-tile-capabilities')).toBeVisible();
      await expect(page.getByTestId('home-tile-applications')).toBeVisible();
    });
  });

  test('opens the canvas at /canvas from the navigation', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByRole('heading', { name: 'Home', level: 2 })).toBeVisible({ timeout: 30000 });

    await page.getByTestId('nav-canvas').click();

    await expect(page).toHaveURL(/\/canvas$/);
    await expect(page.getByTestId('canvas-loaded')).toBeVisible({ timeout: 30000 });
  });

  test.describe('a legacy view link', () => {
    let view: Created;

    test.beforeEach(async ({ request }) => {
      view = await create(request, '/api/v1/views', { name: `E2E Linked View ${Date.now()}`, description: 'Spec 228' });
    });

    test.afterEach(async ({ request }) => {
      await request.delete(`${API_URL}/api/v1/views/${view.id}`, { headers: BYPASS_USER_TENANT });
    });

    test('opens the canvas with that view', async ({ page }) => {
      await page.goto(`/?view=${view.id}`);

      await expect(page.getByTestId('canvas-loaded')).toBeVisible({ timeout: 30000 });
      await expect(page).toHaveURL(/\/canvas$/);
      await expect(page.getByRole('button', { name: `Switch to ${view.name}` })).toBeVisible();
    });
  });
});
