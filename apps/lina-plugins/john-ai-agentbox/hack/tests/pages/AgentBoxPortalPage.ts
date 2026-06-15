import { expect } from '../../../../../../hack/tests/fixtures/auth';
import { agentBoxPublicPath } from '../support/plugin';

type AgentBoxLocator = any;
type AgentBoxPage = {
  getByTestId: (testId: string) => any;
  goto: (url: string) => Promise<unknown>;
  waitForLoadState: (state: 'networkidle') => Promise<unknown>;
  waitForURL: (url: RegExp) => Promise<unknown>;
};

/**
 * AgentBoxPortalPage models the plugin-owned portal login flow.
 * It keeps assertions on the plugin's root portal and independent login page
 * inside the plugin test boundary so host workbench selectors do not leak in.
 */
export class AgentBoxPortalPage {
  readonly page: AgentBoxPage;

  constructor(page: AgentBoxPage) {
    this.page = page;
  }

  get portalGate(): AgentBoxLocator {
    return this.page.getByTestId('portal-auth-gate');
  }

  get portalLoginLink(): AgentBoxLocator {
    return this.page.getByTestId('portal-login-link');
  }

  get loginPage(): AgentBoxLocator {
    return this.page.getByTestId('login-page');
  }

  get loginUsername(): AgentBoxLocator {
    return this.page.getByTestId('login-username');
  }

  get loginPassword(): AgentBoxLocator {
    return this.page.getByTestId('login-password');
  }

  get loginSubmit(): AgentBoxLocator {
    return this.page.getByTestId('login-submit');
  }

  get appShell(): AgentBoxLocator {
    return this.page.getByTestId('agentbox-app-shell');
  }

  get logoutButton(): AgentBoxLocator {
    return this.page.getByTestId('agentbox-logout-button');
  }

  async gotoRoot() {
    await this.page.goto(agentBoxPublicPath('/'));
    await this.page.waitForLoadState('networkidle');
    await expect(this.portalGate).toBeVisible();
  }

  async gotoLogin() {
    await this.page.goto(agentBoxPublicPath('/login'));
    await this.page.waitForLoadState('networkidle');
    await expect(this.loginPage).toBeVisible();
  }

  async openLoginFromPortal() {
    await this.portalLoginLink.click();
    await this.page.waitForURL(/\/login$/);
    await expect(this.loginPage).toBeVisible();
  }

  async login(username: string, password: string) {
    await this.loginUsername.fill(username);
    await this.loginPassword.fill(password);
    await this.loginSubmit.click();
  }

  async loginAndWaitForApp(username: string, password: string) {
    await this.login(username, password);
    await expect(this.appShell).toBeVisible({ timeout: 15000 });
  }

  async logoutAndWaitForPortal() {
    await this.logoutButton.click();
    await expect(this.portalGate).toBeVisible({ timeout: 15000 });
  }
}
