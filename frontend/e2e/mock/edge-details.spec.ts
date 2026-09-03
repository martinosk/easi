import { expect, type Page, test } from '@playwright/test';

async function openCanvas(page: Page): Promise<void> {
  await page.goto('/');
  await page.waitForSelector('[data-testid="main-region"] > *', { timeout: 30000 });
  await page.locator('.react-flow__edge').first().waitFor({ timeout: 30000 });
}

test.describe('canvas edge details edited in place (spec 222)', () => {
  test('renames a relation in place and the edge label follows', async ({ page }) => {
    test.slow();
    await openCanvas(page);
    await page.locator('.react-flow__edge[data-id="rel-phoenix-seabook"]').click({ force: true });

    const details = page.getByTestId('details-pane');
    await expect(details.getByRole('heading', { name: 'Sends bookings' })).toBeVisible();
    await expect(details.getByRole('button', { name: 'Edit', exact: true })).toHaveCount(0);

    await details.getByRole('button', { name: 'Edit name' }).click();
    await details.getByTestId('relation-name-input').fill('Sends confirmed bookings');
    await details.getByTestId('relation-name-input').press('Enter');

    await expect(details.getByRole('heading', { name: 'Sends confirmed bookings' })).toBeVisible();
    await expect(page.locator('.react-flow__edge[data-id="rel-phoenix-seabook"]')).toContainText(
      'Sends confirmed bookings',
    );
  });

  test('changes a realization level in place', async ({ page }) => {
    test.slow();
    await openCanvas(page);
    await page.locator('.react-flow__edge[data-id="realization-real-phoenix-account"]').click({ force: true });

    const details = page.getByTestId('details-pane');
    await expect(details.getByText('Customer Account Creation')).toBeVisible();
    await expect(details.getByRole('button', { name: 'Edit', exact: true })).toHaveCount(0);

    await details.getByRole('button', { name: 'Edit realization level' }).click();
    await details.getByTestId('realization-level-input').click();
    await page.getByRole('option', { name: 'Planned' }).click();
    await details.getByTestId('realization-level-save').click();

    await expect(details.getByTestId('realization-level-value')).toHaveText('Planned');
  });
});
