import { test, expect } from '@playwright/test';

/**
 * Core E2E Workflows
 *
 * These tests cover the essential user workflows in an isolated environment.
 * Each test runs against a clean database spun up via Docker Compose.
 */

test.describe('Core Application Workflows', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to app - isolated backend will have empty database
    await page.goto('/');

    // Wait for canvas to be ready
    await page.waitForSelector('[data-testid="canvas-loaded"]', {
      state: 'visible',
      timeout: 10000
    });

    // Give React Flow time to initialize
    await page.waitForTimeout(500);
  });

  test('should create and display a component', async ({ page }) => {
    // Open create component dialog
    await page.click('[data-testid="create-component-button"]');

    // Wait for dialog
    await expect(page.locator('[data-testid="create-component-dialog"]')).toBeVisible();

    // Fill in component details
    await page.fill('[data-testid="component-name-input"]', 'User Service');
    await page.fill('[data-testid="component-description-input"]', 'Handles user authentication');

    // Submit
    await page.click('[data-testid="create-component-submit"]');

    // Wait for dialog to close
    await page.waitForTimeout(500);
    await expect(page.locator('[data-testid="create-component-dialog"]')).not.toBeVisible();

    // Verify component appears on canvas
    const componentHeader = page.locator('.component-node-header').filter({ hasText: 'User Service' });
    await expect(componentHeader).toBeVisible();

    // Verify we have exactly 1 component
    const componentNodes = page.locator('[data-component-id]');
    await expect(componentNodes).toHaveCount(1);
  });

  test('should validate component name is required', async ({ page }) => {
    // Open create component dialog
    await page.click('[data-testid="create-component-button"]');
    await expect(page.locator('[data-testid="create-component-dialog"]')).toBeVisible();

    // Try to submit without name
    await page.fill('[data-testid="component-name-input"]', '');
    await page.fill('[data-testid="component-description-input"]', 'Some description');

    // Submit button should be disabled
    const submitButton = page.locator('[data-testid="create-component-submit"]');
    await expect(submitButton).toBeDisabled();

    // Fill name and verify button becomes enabled
    await page.fill('[data-testid="component-name-input"]', 'Valid Name');
    await expect(submitButton).toBeEnabled();
  });

  test('should persist component after page reload', async ({ page }) => {
    // Create a component
    await page.click('[data-testid="create-component-button"]');
    await page.fill('[data-testid="component-name-input"]', 'Payment Service');
    await page.fill('[data-testid="component-description-input"]', 'Processes payments');
    await page.click('[data-testid="create-component-submit"]');
    await page.waitForTimeout(500);

    // Verify it exists
    await expect(page.locator('.component-node-header').filter({ hasText: 'Payment Service' })).toBeVisible();

    // Reload page
    await page.reload();
    await page.waitForSelector('[data-testid="canvas-loaded"]', { state: 'visible', timeout: 10000 });
    await page.waitForTimeout(500);

    // Verify component still exists
    await expect(page.locator('.component-node-header').filter({ hasText: 'Payment Service' })).toBeVisible();

    // Should still be exactly 1 component
    const componentNodes = page.locator('[data-component-id]');
    await expect(componentNodes).toHaveCount(1);
  });

  test('should create multiple components and display all', async ({ page }) => {
    // Create first component
    await page.click('[data-testid="create-component-button"]');
    await page.fill('[data-testid="component-name-input"]', 'Order Service');
    await page.fill('[data-testid="component-description-input"]', 'Manages orders');
    await page.click('[data-testid="create-component-submit"]');
    await page.waitForTimeout(500);

    // Create second component
    await page.click('[data-testid="create-component-button"]');
    await page.fill('[data-testid="component-name-input"]', 'Inventory Service');
    await page.fill('[data-testid="component-description-input"]', 'Tracks inventory');
    await page.click('[data-testid="create-component-submit"]');
    await page.waitForTimeout(500);

    // Verify both components exist
    await expect(page.locator('.component-node-header').filter({ hasText: 'Order Service' })).toBeVisible();
    await expect(page.locator('.component-node-header').filter({ hasText: 'Inventory Service' })).toBeVisible();

    // Verify we have exactly 2 components
    const componentNodes = page.locator('[data-component-id]');
    await expect(componentNodes).toHaveCount(2);
  });
});
