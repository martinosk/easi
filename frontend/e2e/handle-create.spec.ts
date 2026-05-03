import { expect, test } from '@playwright/test';

test.describe('Spec 165 — handle-click create related component', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.waitForSelector('[data-testid="canvas-loaded"]', { state: 'visible', timeout: 10000 });
    await page.waitForTimeout(500);
  });

  test('clicking a handle opens the picker and creates a related component', async ({ page }) => {
    await page.click('[data-testid="create-component-button"]');
    await expect(page.locator('[data-testid="create-component-dialog"]')).toBeVisible();
    await page.fill('[data-testid="component-name-input"]', 'Order Service');
    await page.click('[data-testid="create-component-submit"]');
    await expect(page.locator('[data-testid="create-component-dialog"]')).not.toBeVisible();

    const sourceNode = page.locator('[data-component-id]').first();
    await expect(sourceNode).toBeVisible();

    const rightHandle = sourceNode.locator('.component-handle-right').first();
    await expect(rightHandle).toBeVisible();
    const box = await rightHandle.boundingBox();
    if (!box) throw new Error('right handle had no bounding box');
    const cx = box.x + box.width / 2;
    const cy = box.y + box.height / 2;
    await page.mouse.move(cx, cy);
    await page.mouse.down();
    await page.mouse.up();

    const pickerItem = page.getByRole('menuitem', { name: 'Component (Triggers)' });
    await expect(pickerItem).toBeVisible();
    await pickerItem.click();

    await expect(page.locator('[data-testid="create-component-dialog"]')).toBeVisible();
    await page.fill('[data-testid="component-name-input"]', 'Payment Service');
    await page.click('[data-testid="create-component-submit"]');
    await expect(page.locator('[data-testid="create-component-dialog"]')).not.toBeVisible();

    await expect(page.locator('.component-node-header').filter({ hasText: 'Order Service' })).toBeVisible();
    await expect(page.locator('.component-node-header').filter({ hasText: 'Payment Service' })).toBeVisible();
  });
});
