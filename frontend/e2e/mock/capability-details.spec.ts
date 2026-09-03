import { expect, type Page, test } from '@playwright/test';

async function openPage(page: Page, path: string): Promise<void> {
  await page.goto(path);
  await page.waitForSelector('[data-testid="main-region"] > *', { timeout: 30000 });
}

test.describe('capability details edited in place (spec 220)', () => {
  test('renames the capability from the canvas details pane and the tree follows', async ({ page }) => {
    test.slow();
    await openPage(page, '/');
    await page.getByTestId('capability-tree-item-cap-fraud-prevention').click();

    const details = page.getByTestId('details-pane');
    await expect(details.getByRole('heading', { name: 'Customer Fraud Prevention' })).toBeVisible();
    await expect(details.getByRole('button', { name: 'Edit', exact: true })).toHaveCount(0);

    await details.getByRole('button', { name: 'Edit name' }).click();
    await details.getByTestId('capability-name-input').fill('Fraud Prevention');
    await details.getByTestId('capability-name-input').press('Enter');

    await expect(details.getByRole('heading', { name: 'Fraud Prevention' })).toBeVisible();
    await expect(page.getByTestId('capability-tree-item-cap-fraud-prevention')).toContainText('Fraud Prevention');
  });

  test('changes the status from a domain board drawer and the canvas pane reflects it', async ({ page }) => {
    test.slow();
    await openPage(page, '/business-domains');
    await page.getByTestId('l1-toggle-cap-account-management').click();
    await page.getByTestId('capability-card-cap-account-creation').click();

    const drawer = page.getByTestId('capability-drawer');
    await expect(drawer.getByRole('heading', { name: 'Customer Account Creation' })).toBeVisible();
    await expect(drawer.getByRole('button', { name: 'Edit name' })).toBeVisible();

    await drawer.getByRole('button', { name: 'Set a status' }).click();
    await drawer.getByTestId('capability-status-input').click();
    await page.getByRole('option', { name: 'Inactive' }).click();
    await drawer.getByTestId('capability-status-save').click();

    await expect(drawer.getByTestId('capability-status-value')).toHaveText('Inactive');

    await page.keyboard.press('Escape');
    await expect(drawer).toBeHidden();
    await page.getByRole('button', { name: 'Architecture Canvas' }).click();
    await page
      .getByTestId('capability-tree-item-cap-account-management')
      .getByRole('button', { name: 'Expand' })
      .click();
    await page.getByTestId('capability-tree-item-cap-account-creation').click();
    await expect(page.getByTestId('details-pane').getByTestId('capability-status-value')).toHaveText('Inactive');
  });
});
