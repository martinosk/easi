import { expect, type Page, test } from '@playwright/test';
import { detailGroupIds } from '../helpers';

const SHOTS = process.env.EASI_E2E_SHOTS;

const DEFAULT_GROUPS = [
  'detail-group-description',
  'detail-group-ownership',
  'detail-group-composition',
  'detail-group-metadata',
  'detail-group-realisations',
  'detail-group-origins',
  'detail-group-fit',
];

async function openPage(page: Page, path: string): Promise<void> {
  await page.goto(path);
  await page.waitForSelector('[data-testid="main-region"] > *', { timeout: 30000 });
}

async function snapshot(page: Page, name: string): Promise<void> {
  if (SHOTS) await page.screenshot({ path: `${SHOTS}/${name}.png`, fullPage: true });
}

test.describe('application details arranged in groups (spec 224)', () => {
  test('collapses and moves a group in the drawer and the canvas pane shows the same arrangement', async ({ page }) => {
    test.slow();
    await openPage(page, '/business-domains');
    await page.getByTestId('l1-toggle-cap-account-management').click();
    await page.getByTestId('capability-card-cap-account-creation').click();
    await page.getByTestId('capability-drawer').getByTestId('app-chip-comp-phoenix').click();

    const drawer = page.getByTestId('application-drawer');
    await expect(drawer.getByRole('heading', { name: 'Phoenix' })).toBeVisible();
    await expect(drawer.getByTestId('detail-group-fit')).toBeVisible();
    expect(await detailGroupIds(drawer)).toEqual(DEFAULT_GROUPS);
    await expect(drawer.getByText('Experts', { exact: true })).toBeVisible();
    await snapshot(page, 'application-drawer-default');

    await drawer.getByRole('button', { name: 'Metadata', exact: true }).click();
    await expect(drawer.getByTestId('detail-group-metadata').getByRole('region')).toBeHidden();
    await drawer.getByRole('button', { name: 'Move Fit scores up' }).click();
    expect((await detailGroupIds(drawer)).slice(-2)).toEqual(['detail-group-fit', 'detail-group-origins']);
    await snapshot(page, 'application-drawer-rearranged');

    await page.keyboard.press('Escape');
    await expect(drawer).toBeHidden();
    await page.getByRole('button', { name: 'Architecture Canvas' }).click();
    await page.getByTestId('tree-item').filter({ hasText: 'Phoenix' }).first().click();

    const details = page.getByTestId('details-pane');
    await expect(details.getByRole('heading', { name: 'Phoenix' })).toBeVisible();
    await expect(details.getByTestId('detail-group-fit')).toBeVisible();
    const ids = await detailGroupIds(details);
    expect(ids.slice(0, 5)).toEqual(DEFAULT_GROUPS.slice(0, 5));
    expect(ids.slice(5, 7)).toEqual(['detail-group-fit', 'detail-group-origins']);
    await expect(details.getByRole('button', { name: 'Metadata', exact: true })).toHaveAttribute(
      'aria-expanded',
      'false',
    );
    await snapshot(page, 'application-canvas-rearranged');
  });
});
