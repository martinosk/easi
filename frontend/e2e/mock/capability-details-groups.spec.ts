import { expect, type Locator, type Page, test } from '@playwright/test';

const SHOTS = process.env.EASI_E2E_SHOTS;

async function openPage(page: Page, path: string): Promise<void> {
  await page.goto(path);
  await page.waitForSelector('[data-testid="main-region"] > *', { timeout: 30000 });
}

function groupIds(surface: Locator): Promise<(string | undefined)[]> {
  return surface
    .locator('[data-testid^="detail-group-"]')
    .evaluateAll((items) => items.map((item) => (item as HTMLElement).dataset.testid));
}

async function snapshot(page: Page, name: string): Promise<void> {
  if (SHOTS) await page.screenshot({ path: `${SHOTS}/${name}.png`, fullPage: true });
}

test.describe('capability details arranged in groups (spec 223)', () => {
  test('collapses and moves a group in the drawer and the canvas pane shows the same arrangement', async ({ page }) => {
    test.slow();
    await openPage(page, '/business-domains');
    await page.getByTestId('l1-toggle-cap-account-management').click();
    await page.getByTestId('capability-card-cap-account-creation').click();

    const drawer = page.getByTestId('capability-drawer');
    await expect(drawer.getByRole('heading', { name: 'Customer Account Creation' })).toBeVisible();
    await expect(drawer.getByTestId('detail-group-realisations')).toBeVisible();
    expect(await groupIds(drawer)).toEqual([
      'detail-group-description',
      'detail-group-transition',
      'detail-group-fitness',
      'detail-group-metadata',
      'detail-group-realisations',
    ]);
    await expect(drawer.getByRole('button', { name: 'Plan journey' })).toBeVisible();
    await snapshot(page, 'drawer-default');

    await drawer.getByRole('button', { name: 'Metadata', exact: true }).click();
    await expect(drawer.getByRole('button', { name: 'Metadata', exact: true })).toHaveAttribute('aria-expanded', 'false');
    await expect(drawer.getByTestId('detail-group-metadata').getByRole('region')).toBeHidden();
    await drawer.getByRole('button', { name: 'Move Fitness up' }).click();
    expect(await groupIds(drawer)).toEqual([
      'detail-group-description',
      'detail-group-fitness',
      'detail-group-transition',
      'detail-group-metadata',
      'detail-group-realisations',
    ]);
    await snapshot(page, 'drawer-rearranged');

    await page.keyboard.press('Escape');
    await expect(drawer).toBeHidden();
    await page.getByRole('button', { name: 'Architecture Canvas' }).click();
    await page
      .getByTestId('capability-tree-item-cap-account-management')
      .getByRole('button', { name: 'Expand' })
      .click();
    await page.getByTestId('capability-tree-item-cap-account-creation').click();

    const details = page.getByTestId('details-pane');
    await expect(details.getByTestId('detail-group-realisations')).toBeVisible();
    expect((await groupIds(details)).slice(0, 4)).toEqual([
      'detail-group-description',
      'detail-group-fitness',
      'detail-group-metadata',
      'detail-group-realisations',
    ]);
    await expect(details.getByTestId('detail-group-transition')).toHaveCount(0);
    await expect(details.getByRole('button', { name: 'Metadata', exact: true })).toHaveAttribute('aria-expanded', 'false');
    await snapshot(page, 'canvas-rearranged');
  });
});
