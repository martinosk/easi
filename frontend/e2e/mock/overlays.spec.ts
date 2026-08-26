import { expect, type Locator, type Page, test } from '@playwright/test';

const PAGES = [
  { name: 'canvas', path: '/' },
  { name: 'business domains', path: '/business-domains' },
  { name: 'value streams', path: '/value-streams' },
  { name: 'enterprise architecture', path: '/enterprise-architecture' },
];

const VIEWPORTS = [
  { name: 'wide', width: 1280, height: 800, overflowExpected: false },
  { name: 'narrow', width: 420, height: 800, overflowExpected: true },
];

const HEADER_OVERLAYS = [
  { trigger: 'nav-more', overlay: 'nav-more-menu', onlyWhenOverflowing: true },
  { trigger: 'user-menu-trigger', overlay: 'user-menu-dropdown', onlyWhenOverflowing: false },
];

async function openPage(page: Page, path: string): Promise<void> {
  await page.goto(path);
  await page.waitForSelector('[data-testid="main-region"] > *', { timeout: 30000 });
  await page.waitForLoadState('networkidle');
}

async function expectTopmost(overlay: Locator): Promise<void> {
  await expect(overlay).toBeVisible();
  const topmost = await overlay.evaluate((element) => {
    const rect = element.getBoundingClientRect();
    const hit = document.elementFromPoint(rect.left + rect.width / 2, rect.top + rect.height / 2);
    return hit !== null && element.contains(hit);
  });
  expect(topmost, 'overlay is painted behind page content').toBe(true);
}

for (const viewport of VIEWPORTS) {
  test.describe(`header overlays at ${viewport.name} width`, () => {
    test.use({ viewport: { width: viewport.width, height: viewport.height } });

    for (const target of PAGES) {
      test(`are topmost on the ${target.name} page`, async ({ page }) => {
        await openPage(page, target.path);

        const overflowing = (await page.getByTestId('nav-more').count()) > 0;
        expect(overflowing, 'overflow mode did not match the viewport').toBe(viewport.overflowExpected);

        for (const { trigger, overlay, onlyWhenOverflowing } of HEADER_OVERLAYS) {
          if (onlyWhenOverflowing && !overflowing) continue;
          await page.getByTestId(trigger).click();
          await expectTopmost(page.getByTestId(overlay));
          await page.keyboard.press('Escape');
          await expect(page.getByTestId(overlay)).toBeHidden();
        }
      });
    }
  });
}
