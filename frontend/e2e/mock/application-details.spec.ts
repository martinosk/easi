import { expect, type Page, test } from '@playwright/test';

async function openPage(page: Page, path: string): Promise<void> {
  await page.goto(path);
  await page.waitForSelector('[data-testid="main-region"] > *', { timeout: 30000 });
}

test.describe('application details edited in place (spec 219)', () => {
  test('renames the application from the canvas details pane and the tree follows', async ({ page }) => {
    await openPage(page, '/');
    await page.getByTestId('tree-item').filter({ hasText: 'Phoenix' }).first().click();

    const details = page.getByTestId('details-pane');
    await expect(details.getByRole('heading', { name: 'Phoenix' })).toBeVisible();
    await expect(details.getByRole('button', { name: 'Edit', exact: true })).toHaveCount(0);

    await details.getByRole('button', { name: 'Edit name' }).click();
    await details.getByTestId('component-name-input').fill('Phoenix Platform');
    await details.getByTestId('component-name-input').press('Enter');

    await expect(details.getByRole('heading', { name: 'Phoenix Platform' })).toBeVisible();
    await expect(page.getByTestId('tree-item').filter({ hasText: 'Phoenix Platform' })).toHaveCount(1);
  });

  test('edits the description from a domain board chip and shows experts in the drawer', async ({ page }) => {
    await openPage(page, '/business-domains');
    await page.getByTestId('l1-toggle-cap-account-management').click();
    await page.getByTestId('capability-card-cap-account-creation').click();
    await page.getByTestId('capability-drawer').getByTestId('app-chip-comp-phoenix').click();

    const drawer = page.getByTestId('application-drawer');
    await expect(drawer.getByRole('heading', { name: 'Phoenix' })).toBeVisible();
    await expect(drawer.getByText('Experts', { exact: true })).toBeVisible();
    await expect(drawer.getByRole('button', { name: 'Edit', exact: true })).toHaveCount(0);

    await drawer.getByRole('button', { name: 'Edit description' }).click();
    await drawer.getByTestId('component-description-input').fill('Booking platform, now also freight.');
    await drawer.getByRole('button', { name: 'Save' }).click();

    await expect(drawer.getByTestId('component-description-value')).toHaveText('Booking platform, now also freight.');
    await expect(drawer.getByRole('heading', { name: 'Phoenix' })).toBeVisible();
  });
});
