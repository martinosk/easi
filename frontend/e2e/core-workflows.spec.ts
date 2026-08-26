import { expect, test } from '@playwright/test';
import { createComponent, dragTreeItemToCanvas, openApp, saveView, treeItem } from './helpers';

test.describe('Core Application Workflows', () => {
  test.beforeEach(async ({ page, request }) => {
    await openApp(page, request);
  });

  test('creates a component into the model without auto-adding it to the canvas', async ({ page }) => {
    await createComponent(page, 'User Service', 'Handles user authentication');

    await expect(treeItem(page, 'User Service')).toBeVisible();
    await expect(page.locator('[data-component-id]')).toHaveCount(0);
  });

  test('should validate component name is required', async ({ page }) => {
    await page.click('[data-testid="create-component-button"]');
    await expect(page.locator('[data-testid="create-component-dialog"]')).toBeVisible();

    await page.fill('[data-testid="component-name-input"]', '');
    await page.fill('[data-testid="component-description-input"]', 'Some description');

    const submitButton = page.locator('[data-testid="create-component-submit"]');
    await expect(submitButton).toBeDisabled();

    await page.fill('[data-testid="component-name-input"]', 'Valid Name');
    await expect(submitButton).toBeEnabled();
  });

  test('adds a component to the view via the tree and persists it after reload', async ({ page }) => {
    await createComponent(page, 'Payment Service', 'Processes payments');

    await dragTreeItemToCanvas(page, 'Payment Service');
    await expect(page.getByTestId('component-node-header').filter({ hasText: 'Payment Service' })).toBeVisible();

    await saveView(page);

    await page.reload();
    await page.waitForSelector('[data-testid="canvas-loaded"]', {
      state: 'visible',
      timeout: 10000,
    });
    await page.waitForTimeout(500);

    await expect(page.getByTestId('component-node-header').filter({ hasText: 'Payment Service' })).toBeVisible();
    await expect(page.locator('[data-component-id]')).toHaveCount(1);
  });

  test('creates multiple components and shows them all in the explorer', async ({ page }) => {
    await createComponent(page, 'Order Service', 'Manages orders');
    await createComponent(page, 'Inventory Service', 'Tracks inventory');

    await expect(treeItem(page, 'Order Service')).toBeVisible();
    await expect(treeItem(page, 'Inventory Service')).toBeVisible();
    await expect(page.locator('[data-component-id]')).toHaveCount(0);
  });
});
