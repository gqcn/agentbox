import { readFile } from 'node:fs/promises';
import path from 'node:path';

import { test, expect, type Page, type Route } from '../../../../../../hack/tests/fixtures/auth';
import { AgentBoxPortalPage } from '../pages/AgentBoxPortalPage';
import {
  agentBoxPluginID,
  captureAgentBoxScreenshot,
} from '../support/plugin';

const apiPrefix = `/x/${agentBoxPluginID}/api/v1`;
const assetPrefix = `/x-assets/${agentBoxPluginID}/0.1.0`;
const repoRoot = path.resolve(process.cwd(), '../..');
const pluginDistDir = path.join(
  repoRoot,
  'apps/lina-plugins/john-ai-agentbox/frontend/dist',
);
const now = 1_735_689_600;

test.describe('TC004 AgentBox agent form empty provider', () => {
  test.beforeEach(async ({ page }) => {
    await installBuiltAssetRoute(page);
    await installEmptyProviderApiMock(page);
  });

  test('TC004a: create agent provider select does not render numeric sentinel', async ({ page }) => {
    const portal = new AgentBoxPortalPage(page);
    await portal.gotoApp();

    await portal.openCreateAgentDialog();

    await expect(portal.agentProviderSelect).toContainText('暂无供应商');
    await expect(portal.agentProviderSelect).not.toContainText('0');
    await expect(portal.agentProviderSelect).toBeDisabled();
    await captureAgentBoxScreenshot(page, 'tc004a-empty-provider-agent-form');
  });

  test('TC004b: create provider form leaves Anthropic base URL empty', async ({ page }) => {
    const portal = new AgentBoxPortalPage(page);
    await portal.gotoApp();

    await portal.openCreateProviderDialog();

    await expect(portal.providerAnthropicBaseUrlInput).toHaveValue('');
    await captureAgentBoxScreenshot(page, 'tc004b-empty-anthropic-provider-url');
  });
});

async function installEmptyProviderApiMock(page: Page) {
  await page.route(`**${apiPrefix}/**`, async (route) => {
    const request = route.request();
    const url = new URL(request.url());
    const path = url.pathname;

    if (path === `${apiPrefix}/auth/session`) {
      await fulfillJson(route, {
        user: {
          id: 'usr-agentbox-empty-provider',
          username: 'agentbox-admin',
          role: 'admin',
          status: 'active',
          createdAt: now,
          updatedAt: now,
        },
      });
      return;
    }

    if (path === `${apiPrefix}/providers`) {
      await fulfillJson(route, []);
      return;
    }

    if (path === `${apiPrefix}/images`) {
      await fulfillJson(route, [
        {
          id: 1,
          name: 'E2E AgentBox Image',
          imageRef: 'agentbox/e2e:latest',
          agentType: 'claude_code',
          defaultShell: '/bin/zsh',
          notes: 'E2E image',
          enabled: true,
          isDefault: true,
          createdAt: now,
          updatedAt: now,
        },
      ]);
      return;
    }

    if (path === `${apiPrefix}/agents`) {
      await fulfillJson(route, []);
      return;
    }

    if (path === `${apiPrefix}/settings/workbench`) {
      await fulfillJson(route, {
        key: 'workbench',
        value: '{}',
        createdAt: now,
        updatedAt: now,
      });
      return;
    }

    await route.fallback();
  });
}

async function installBuiltAssetRoute(page: Page) {
  await page.route('**/*', async (route) => {
    const url = new URL(route.request().url());
    if (url.pathname === '/' || url.pathname === '/login') {
      await fulfillStaticFile(route, path.join(pluginDistDir, 'index.html'));
      return;
    }
    if (url.pathname.startsWith(`${assetPrefix}/`)) {
      const relativeAssetPath = url.pathname.slice(assetPrefix.length + 1);
      await fulfillStaticFile(route, path.join(pluginDistDir, relativeAssetPath));
      return;
    }
    await route.fallback();
  });
}

async function fulfillStaticFile(route: Route, filePath: string) {
  await route.fulfill({
    status: 200,
    contentType: contentTypeFor(filePath),
    body: await readFile(filePath),
  });
}

async function fulfillJson(route: Route, data: unknown) {
  await route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({
      code: 0,
      message: 'ok',
      data,
    }),
  });
}

function contentTypeFor(filePath: string) {
  if (filePath.endsWith('.html')) return 'text/html';
  if (filePath.endsWith('.js')) return 'text/javascript';
  if (filePath.endsWith('.css')) return 'text/css';
  if (filePath.endsWith('.woff2')) return 'font/woff2';
  if (filePath.endsWith('.ttf')) return 'font/ttf';
  return 'application/octet-stream';
}
