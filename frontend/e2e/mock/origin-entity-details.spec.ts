import { expect, type Page, test } from '@playwright/test';

async function openPage(page: Page, path: string): Promise<void> {
  await page.goto(path);
  await page.waitForSelector('[data-testid="main-region"] > *', { timeout: 30000 });
}

test.describe('origin entity details edited in place (spec 221)', () => {
  test('renames a vendor from the canvas details pane and the tree follows', async ({ page }) => {
    test.slow();
    await openPage(page, '/');
    await page.getByRole('button', { name: /Vendors/ }).click();
    await page.getByTestId('tree-item').filter({ hasText: 'SAP' }).first().click();

    const details = page.getByTestId('details-pane');
    await expect(details.getByRole('heading', { name: 'SAP' })).toBeVisible();
    await expect(details.getByRole('button', { name: 'Edit', exact: true })).toHaveCount(0);

    await details.getByRole('button', { name: 'Edit name' }).click();
    await details.getByTestId('origin-entity-name-input').fill('SAP SE');
    await details.getByTestId('origin-entity-name-input').press('Enter');

    await expect(details.getByRole('heading', { name: 'SAP SE' })).toBeVisible();
    await expect(page.getByTestId('tree-item').filter({ hasText: 'SAP SE' })).toHaveCount(1);
  });

  test("changes an acquired entity's integration status in place", async ({ page }) => {
    test.slow();
    await openPage(page, '/');
    await page.getByRole('button', { name: /Acquired Entities/ }).click();
    await page.getByTestId('tree-item').filter({ hasText: 'Nordic Cargo' }).first().click();

    const details = page.getByTestId('details-pane');
    await expect(details.getByRole('heading', { name: 'Nordic Cargo' })).toBeVisible();
    await expect(details.getByText('Phoenix')).toBeVisible();

    await details.getByRole('button', { name: 'Edit integration status' }).click();
    await details.getByTestId('origin-entity-integration-status-input').click();
    await page.getByRole('option', { name: 'Completed' }).click();
    await details.getByTestId('origin-entity-integration-status-save').click();

    await expect(details.getByTestId('origin-entity-integration-status-value')).toHaveText('Completed');
    await expect(details.getByRole('heading', { name: 'Nordic Cargo' })).toBeVisible();
  });
});
