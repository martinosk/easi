import { expect, test } from '@playwright/test';
import { dragTreeItemToCanvas, openApp } from './helpers';

test.describe('Spec 165 — handle-click create related component', () => {
  test.beforeEach(async ({ page, request }) => {
    await openApp(page, request);
  });

  test('clicking a handle opens the picker and creates a related component', async ({ page }) => {
    await page.click('[data-testid="create-component-button"]');
    await expect(page.locator('[data-testid="create-component-dialog"]')).toBeVisible();
    await page.fill('[data-testid="component-name-input"]', 'Order Service');
    await page.click('[data-testid="create-component-submit"]');
    await expect(page.locator('[data-testid="create-component-dialog"]')).not.toBeVisible();

    await dragTreeItemToCanvas(page, 'Order Service');

    const sourceNode = page.locator('[data-component-id]').first();
    await expect(sourceNode).toBeVisible();

    const rightHandle = sourceNode.getByTestId('component-handle-right').first();
    await expect(rightHandle).toBeVisible();
    const box = await rightHandle.boundingBox();
    if (!box) throw new Error('right handle had no bounding box');
    const cx = box.x + box.width / 2;
    const cy = box.y + box.height / 2;
    await page.mouse.move(cx, cy);
    await page.mouse.down();
    await page.mouse.up();

    const pickerItem = page.getByRole('menuitem', { name: 'Component (triggers)' });
    await expect(pickerItem).toBeVisible();
    await pickerItem.click();

    await expect(page.locator('[data-testid="create-component-dialog"]')).toBeVisible();
    await page.fill('[data-testid="component-name-input"]', 'Payment Service');
    await page.click('[data-testid="create-component-submit"]');
    await expect(page.locator('[data-testid="create-component-dialog"]')).not.toBeVisible();

    await expect(page.getByTestId('component-node-header').filter({ hasText: 'Order Service' })).toBeVisible();
    await expect(page.getByTestId('component-node-header').filter({ hasText: 'Payment Service' })).toBeVisible();
  });

  test('clicking one handle then another does not create a relation (drag-only)', async ({ page }) => {
    await page.click('[data-testid="create-component-button"]');
    await page.fill('[data-testid="component-name-input"]', 'Alpha');
    await page.click('[data-testid="create-component-submit"]');
    await expect(page.locator('[data-testid="create-component-dialog"]')).not.toBeVisible();

    await page.click('[data-testid="create-component-button"]');
    await page.fill('[data-testid="component-name-input"]', 'Beta');
    await page.click('[data-testid="create-component-submit"]');
    await expect(page.locator('[data-testid="create-component-dialog"]')).not.toBeVisible();

    await dragTreeItemToCanvas(page, 'Alpha', { x: 250, y: 200 });
    await dragTreeItemToCanvas(page, 'Beta', { x: 600, y: 220 });

    const alpha = page.locator('[data-component-id]').filter({ hasText: 'Alpha' }).first();
    const beta = page.locator('[data-component-id]').filter({ hasText: 'Beta' }).first();
    await expect(alpha).toBeVisible();
    await expect(beta).toBeVisible();

    const clickHandle = async (node: ReturnType<typeof page.locator>) => {
      const handle = node.getByTestId('component-handle-right').first();
      const box = await handle.boundingBox();
      if (!box) throw new Error('handle had no bounding box');
      await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
      await page.mouse.down();
      await page.mouse.up();
    };

    await clickHandle(alpha);
    await page.keyboard.press('Escape');
    await clickHandle(beta);

    await expect(page.locator('[data-testid="relation-name-input"]')).toHaveCount(0);
  });
});
