import { expect, type Locator, type Page } from '@playwright/test';

export async function createComponent(page: Page, name: string, description?: string): Promise<void> {
  await page.click('[data-testid="create-component-button"]');
  await expect(page.locator('[data-testid="create-component-dialog"]')).toBeVisible();
  await page.fill('[data-testid="component-name-input"]', name);
  if (description !== undefined) {
    await page.fill('[data-testid="component-description-input"]', description);
  }
  await page.click('[data-testid="create-component-submit"]');
  await expect(page.locator('[data-testid="create-component-dialog"]')).not.toBeVisible();
}

export function treeItem(page: Page, name: string): Locator {
  return page.locator('.tree-item', { hasText: name }).first();
}

export async function dragTreeItemToCanvas(
  page: Page,
  name: string,
  targetPosition: { x: number; y: number } = { x: 300, y: 200 },
): Promise<void> {
  const item = treeItem(page, name);
  await expect(item).toBeVisible();
  await item.dragTo(page.locator('[data-testid="canvas-loaded"]'), { targetPosition });
}

export async function saveView(page: Page): Promise<void> {
  const saveButton = page.getByRole('button', { name: /^Save view/ });
  await expect(saveButton).toBeEnabled();
  await saveButton.click();
  await expect(saveButton).toBeDisabled();
}
