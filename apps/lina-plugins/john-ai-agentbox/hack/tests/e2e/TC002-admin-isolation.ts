import { test, expect } from '../../../../../../hack/tests/fixtures/auth';

import { ensureSourcePluginEnabled } from '../../../../../../hack/tests/fixtures/plugin';
import { AgentBoxPortalPage } from '../pages/AgentBoxPortalPage';
import {
  agentBoxPluginID,
  agentBoxPublicPath,
  captureAgentBoxScreenshot,
} from '../support/plugin';
import { config } from '../../../../../../hack/tests/fixtures/config';
import { LoginPage } from '../../../../../../hack/tests/pages/LoginPage';
import { MainLayout } from '../../../../../../hack/tests/pages/MainLayout';

test.describe('TC002 AgentBox and admin isolation', () => {
  test.beforeEach(async ({ page }) => {
    await ensureSourcePluginEnabled(page, agentBoxPluginID);
  });

  test('TC002a: AgentBox portal does not authenticate /admin', async ({ page }) => {
    const portal = new AgentBoxPortalPage(page);
    await portal.gotoLogin();
    await portal.loginAndWaitForApp('admin', 'admin123');

    await page.goto('/admin');
    await page.waitForLoadState('networkidle');
    await expect(page.url()).toContain('/auth/login');
    await captureAgentBoxScreenshot(page, 'tc002a-admin-auth-boundary');
  });

  test('TC002b: /admin login state does not replace AgentBox session boundary', async ({ page }) => {
    const adminLogin = new LoginPage(page);
    await adminLogin.goto();
    await adminLogin.loginAndWaitForRedirect(config.adminUser, config.adminPass);
    await expect(page).toHaveURL(/\/admin\//);

    await page.goto(agentBoxPublicPath('/'));
    await page.waitForLoadState('networkidle');
    await expect(page.getByTestId('portal-auth-gate')).toBeVisible();
    await captureAgentBoxScreenshot(page, 'tc002b-agentbox-boundary-after-admin-login');
  });

  test('TC002c: logout actions remain scoped to each surface', async ({ page }) => {
    const portal = new AgentBoxPortalPage(page);
    await portal.gotoLogin();
    await portal.loginAndWaitForApp('admin', 'admin123');
    await portal.logoutAndWaitForPortal();

    const adminLogin = new LoginPage(page);
    await adminLogin.goto();
    await adminLogin.loginAndWaitForRedirect(config.adminUser, config.adminPass);
    const adminLayout = new MainLayout(page);
    await adminLayout.logout();

    await expect(page.url()).toContain('/auth/login');
    await page.goto(agentBoxPublicPath('/'));
    await page.waitForLoadState('networkidle');
    await expect(page.getByTestId('portal-auth-gate')).toBeVisible();
    await captureAgentBoxScreenshot(page, 'tc002c-scoped-logout-boundary');
  });
});
