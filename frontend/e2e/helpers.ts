import { type APIRequestContext, expect, type Locator, type Page } from '@playwright/test';

const API_URL = 'http://localhost:8081';

interface ListedResource {
  id: string;
  _links?: { delete?: { href: string } };
}

async function deleteAll(request: APIRequestContext, collectionPath: string): Promise<void> {
  const response = await request.get(`${API_URL}${collectionPath}`);
  const body = (await response.json()) as { data: ListedResource[] | null };
  for (const resource of body.data ?? []) {
    const href = resource._links?.delete?.href;
    if (href) {
      await request.delete(`${API_URL}${href}`);
    }
  }
}

export async function resetBackendData(request: APIRequestContext): Promise<void> {
  await deleteAll(request, '/api/v1/relations');
  await deleteAll(request, '/api/v1/components');
}

export async function openApp(page: Page, request: APIRequestContext): Promise<void> {
  await resetBackendData(request);
  const response = await request.get(`${API_URL}/api/v1/version`);
  const { version } = (await response.json()) as { version: string };
  await page.addInitScript((dismissedVersion) => {
    localStorage.setItem('releaseNotesPreferences', JSON.stringify({ dismissedVersion, dismissMode: 'forever' }));
  }, version);
  await page.goto('/');
  await page.waitForSelector('[data-testid="canvas-loaded"]', { state: 'visible', timeout: 30000 });
  await page.waitForTimeout(500);
}

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
  return page.getByTestId('tree-item').filter({ hasText: name }).first();
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
