import { expect, type Page, test } from '@playwright/test';

async function openHome(page: Page): Promise<void> {
  await page.goto('/');
  await expect(page.getByRole('heading', { name: 'Home', level: 2 })).toBeVisible({ timeout: 30000 });
}

test.describe('personal home (spec 228)', () => {
  test('lands on Home with scope, tiles and My Work', async ({ page }) => {
    await openHome(page);

    await expect(page.getByTestId('nav-home')).toHaveAttribute('aria-current', 'page');
    await expect(page.getByTestId('home-scope')).toContainText('Steward of Assessment in Customer Engagement');
    await expect(page.getByTestId('home-tile-capabilities')).toContainText('12');
    await expect(page.getByTestId('home-tile-time')).toContainText('Not assessed 15%');
    await expect(page.getByRole('link', { name: /Phoenix/ })).toContainText('Edit access until 21 Oct 2026');
  });

  test('opens a one-pager from a My Work card', async ({ page }) => {
    await openHome(page);

    await page.getByRole('link', { name: /Customer Account Creation/ }).click();

    await expect(page).toHaveURL(/\/one-pagers\/capability\/cap-account-creation$/);
  });

  test('opens the canvas at /canvas from the navigation', async ({ page }) => {
    await openHome(page);

    await page.getByTestId('nav-canvas').click();

    await expect(page).toHaveURL(/\/canvas$/);
    await expect(page.getByTestId('canvas-loaded')).toBeVisible({ timeout: 30000 });
    await expect(page.getByTestId('nav-canvas')).toHaveAttribute('aria-current', 'page');
  });

  test('opens a legacy view link on the canvas with that view', async ({ page }) => {
    await page.goto('/?view=view-integration-landscape');

    await expect(page.getByTestId('canvas-loaded')).toBeVisible({ timeout: 30000 });
    await expect(page).toHaveURL(/\/canvas$/);
    await expect(page.getByRole('button', { name: 'Switch to Integration Landscape' })).toBeVisible();
  });

  test('returns to Home from the logo', async ({ page }) => {
    await page.goto('/business-domains');
    await page.getByRole('link', { name: 'EASI home' }).click();

    await expect(page).toHaveURL(/\/$/);
    await expect(page.getByRole('heading', { name: 'Home', level: 2 })).toBeVisible();
  });
});
