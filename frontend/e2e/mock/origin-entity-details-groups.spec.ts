import { expect, type Page, test } from '@playwright/test';
import { detailGroupIds } from '../helpers';

const SHOTS = process.env.EASI_E2E_SHOTS;

async function openPage(page: Page, path: string): Promise<void> {
  await page.goto(path);
  await page.waitForSelector('[data-testid="main-region"] > *', { timeout: 30000 });
}

async function snapshot(page: Page, name: string): Promise<void> {
  if (SHOTS) await page.screenshot({ path: `${SHOTS}/${name}.png`, fullPage: true });
}

test.describe('origin entity details arranged in groups (spec 225)', () => {
  test('collapses and moves a group on a vendor and an acquired entity shows the same arrangement', async ({
    page,
  }) => {
    test.slow();
    await openPage(page, '/');
    await page.getByRole('button', { name: /Vendors/ }).click();
    await page.getByTestId('tree-item').filter({ hasText: 'SAP' }).first().click();

    const details = page.getByTestId('details-pane');
    await expect(details.getByRole('heading', { name: 'SAP' })).toBeVisible();
    await expect(details.getByTestId('detail-group-applications')).toBeVisible();
    expect((await detailGroupIds(details)).slice(0, 3)).toEqual([
      'detail-group-description',
      'detail-group-metadata',
      'detail-group-applications',
    ]);
    await snapshot(page, 'origin-canvas-default');

    await details.getByRole('button', { name: 'Metadata', exact: true }).click();
    await expect(details.getByTestId('detail-group-metadata').getByRole('region')).toBeHidden();
    await details.getByRole('button', { name: 'Move Applications up' }).click();
    expect((await detailGroupIds(details)).slice(0, 3)).toEqual([
      'detail-group-description',
      'detail-group-applications',
      'detail-group-metadata',
    ]);

    await page.getByRole('button', { name: /Acquired Entities/ }).click();
    await page.getByTestId('tree-item').filter({ hasText: 'Nordic Cargo' }).first().click();
    await expect(details.getByRole('heading', { name: 'Nordic Cargo' })).toBeVisible();
    expect((await detailGroupIds(details)).slice(0, 3)).toEqual([
      'detail-group-description',
      'detail-group-applications',
      'detail-group-metadata',
    ]);
    await expect(details.getByRole('button', { name: 'Metadata', exact: true })).toHaveAttribute(
      'aria-expanded',
      'false',
    );
    await snapshot(page, 'origin-canvas-rearranged');
  });
});
