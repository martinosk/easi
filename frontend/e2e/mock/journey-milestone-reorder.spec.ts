import { expect, type Page, test } from '@playwright/test';

const L1_TOGGLE = 'l1-toggle-cap-account-management';
const CAPABILITY_CARD = 'capability-card-cap-account-creation';

async function openAccountManagementDrawer(page: Page): Promise<void> {
  await page.goto('/business-domains');
  await page.waitForSelector('[data-testid="main-region"] > *', { timeout: 30000 });
  await page.getByTestId(L1_TOGGLE).click();
  await page.getByTestId(CAPABILITY_CARD).click();
  await expect(page.getByTestId('milestone-row-ms-readonly')).toBeVisible();
}

async function milestoneOrder(page: Page): Promise<string[]> {
  return page
    .getByTestId(/^milestone-row-/)
    .evaluateAll((rows) => rows.map((row) => row.getAttribute('data-testid') ?? ''));
}

test.describe('journey milestone reorder (spec 196)', () => {
  test('moves a milestone with the keyboard and renumbers the list', async ({ page }) => {
    await openAccountManagementDrawer(page);
    expect(await milestoneOrder(page)).toEqual([
      'milestone-row-ms-api-live',
      'milestone-row-ms-routes',
      'milestone-row-ms-north-sea',
      'milestone-row-ms-readonly',
    ]);

    await page.getByTestId('milestone-handle-ms-readonly').focus();
    await page.keyboard.press('ArrowUp');

    await expect(page.getByTestId('milestone-seq-ms-readonly')).toHaveText('3');
    expect(await milestoneOrder(page)).toEqual([
      'milestone-row-ms-api-live',
      'milestone-row-ms-routes',
      'milestone-row-ms-readonly',
      'milestone-row-ms-north-sea',
    ]);
    await expect(page.getByTestId('milestone-when-ms-readonly')).toHaveText('Q1 2027');
    await expect(page.getByTestId('milestone-schedule-conflict-ms-north-sea')).toHaveAttribute(
      'aria-label',
      'Targeted for Q4 2026 but listed after a milestone targeted for Q1 2027',
    );
    await expect(page.getByTestId(/^milestone-schedule-conflict-/)).toHaveCount(1);
  });

  test('drags a milestone onto another row', async ({ page }) => {
    await openAccountManagementDrawer(page);

    await page.getByTestId('milestone-row-ms-api-live').dragTo(page.getByTestId('milestone-row-ms-north-sea'));

    await expect(page.getByTestId('milestone-seq-ms-api-live')).toHaveText('3');
    expect(await milestoneOrder(page)).toEqual([
      'milestone-row-ms-routes',
      'milestone-row-ms-north-sea',
      'milestone-row-ms-api-live',
      'milestone-row-ms-readonly',
    ]);
  });
});
