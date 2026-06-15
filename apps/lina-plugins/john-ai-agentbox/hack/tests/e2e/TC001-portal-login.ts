import { test, expect } from '../../../../../../hack/tests/fixtures/auth';

import { ensureSourcePluginEnabled } from '../../../../../../hack/tests/fixtures/plugin';
import {
  execPgSQLStatements,
  pgEscapeLiteral,
} from '../../../../../../hack/tests/support/postgres';
import { AgentBoxPortalPage } from '../pages/AgentBoxPortalPage';
import {
  agentBoxPublicPath,
  agentBoxPluginID,
  captureAgentBoxScreenshot,
} from '../support/plugin';

const agentBoxAdminUserID = 'usr-admin';
const settingNotFoundErrorCode = 'JOHN_AI_AGENTBOX_SETTING_NOT_FOUND';

test.describe('TC001 AgentBox portal and login', () => {
  test.beforeEach(async ({ page }) => {
    await ensureSourcePluginEnabled(page, agentBoxPluginID);
  });

  test('TC001a: root portal shows AgentBox gate', async ({ page }) => {
    const portal = new AgentBoxPortalPage(page);
    await expectPortalEntryServesIndex('/');
    await portal.gotoRoot();

    await expect(portal.portalGate).toBeVisible();
    await expect(portal.portalLoginLink).toBeVisible();
    await expect(portal.loginPage).toHaveCount(0);
    await captureAgentBoxScreenshot(page, 'tc001a-root-auth-gate');
  });

  test('TC001b: plugin login route shows the standalone login page', async ({ page }) => {
    const portal = new AgentBoxPortalPage(page);
    await expectPortalEntryServesIndex('/login');
    await portal.gotoLogin();

    await expect(portal.loginPage).toBeVisible();
    await expect(portal.loginUsername).toBeVisible();
    await expect(portal.loginPassword).toBeVisible();
    await expect(portal.loginSubmit).toBeVisible();
    await captureAgentBoxScreenshot(page, 'tc001b-login-page');
  });

  test('TC001c: login link from portal opens the plugin login route', async ({ page }) => {
    const portal = new AgentBoxPortalPage(page);
    await portal.gotoRoot();
    await portal.openLoginFromPortal();

    await expect(portal.loginPage).toBeVisible();
    await expect(page).toHaveURL(/\/login$/);
    await captureAgentBoxScreenshot(page, 'tc001c-login-from-portal');
  });

  test('TC001d: valid AgentBox login opens the functional app shell', async ({ page }) => {
    resetAdminWorkbenchSetting();
    const portal = new AgentBoxPortalPage(page);
    const workbenchSettingPath = `/x/${agentBoxPluginID}/api/v1/settings/workbench`;
    const settingsGetResponse = page.waitForResponse(
      (response) =>
        response.url().includes(workbenchSettingPath) &&
        response.request().method() === 'GET',
    );

    await portal.gotoLogin();
    try {
      await portal.loginAndWaitForApp('admin', 'admin123');

      const getPayload = await (await settingsGetResponse).json();
      expect(getPayload.errorCode).toBe(settingNotFoundErrorCode);
      await expect(portal.appShell).toBeVisible();
      await expect(page).not.toHaveURL(/\/login$/);
      await page.waitForTimeout(1000);
      await expect(page.getByText(/(加载|初始化|同步)工作台设置失败/u)).toHaveCount(0);
      await captureAgentBoxScreenshot(page, 'tc001d-app-shell');
    } finally {
      resetAdminWorkbenchSetting();
    }
  });
});

function resetAdminWorkbenchSetting() {
  execPgSQLStatements([
    `DELETE FROM john_ai_agentbox_user_settings WHERE user_id = '${pgEscapeLiteral(
      agentBoxAdminUserID,
    )}' AND key = 'workbench';`,
  ]);
}

async function expectPortalEntryServesIndex(route: '/' | '/login') {
  const response = await fetch(agentBoxPublicPath(route), {
    redirect: 'manual',
  });
  expect(response.status).toBe(200);
  expect(response.headers.get('location')).toBeNull();
  expect(response.headers.get('content-type')).toContain('text/html');
  const body = await response.text();
  expect(body).toContain('<title>Agent Box</title>');
  expect(body).toContain('<div id="root"></div>');
  expect(body).toContain('/x-assets/john-ai-agentbox/0.1.0/');
}
