import { expect, type Locator, type Page } from '../../../../../../hack/tests/fixtures/auth';
import { agentBoxPublicPath } from '../support/plugin';

/**
 * AgentBoxPortalPage models the plugin-owned portal login flow.
 * It keeps assertions on the plugin's root portal and independent login page
 * inside the plugin test boundary so host workbench selectors do not leak in.
 */
export class AgentBoxPortalPage {
  readonly page: Page;

  constructor(page: Page) {
    this.page = page;
  }

  get portalGate(): Locator {
    return this.page.getByTestId('portal-auth-gate');
  }

  get portalLoginLink(): Locator {
    return this.page.getByTestId('portal-login-link');
  }

  get loginPage(): Locator {
    return this.page.getByTestId('login-page');
  }

  get loginUsername(): Locator {
    return this.page.getByTestId('login-username');
  }

  get loginPassword(): Locator {
    return this.page.getByTestId('login-password');
  }

  get loginSubmit(): Locator {
    return this.page.getByTestId('login-submit');
  }

  get appShell(): Locator {
    return this.page.getByTestId('agentbox-app-shell');
  }

  get logoutButton(): Locator {
    return this.page.getByTestId('agentbox-logout-button');
  }

  get createAgentButton(): Locator {
    return this.page.getByRole('button', { name: '新增智能体' });
  }

  get providersViewButton(): Locator {
    return this.page.getByRole('button', { name: '供应商' });
  }

  get createProviderButton(): Locator {
    return this.page.getByRole('button', { name: '新增供应商' });
  }

  get agentDialog(): Locator {
    return this.page.getByRole('dialog', { name: '新增智能体' });
  }

  get providerDialog(): Locator {
    return this.page.getByRole('dialog', { name: '新增供应商' });
  }

  get agentProviderSelect(): Locator {
    return this.agentDialog.getByLabel('供应商');
  }

  get providerAnthropicBaseUrlInput(): Locator {
    return this.providerDialog.getByLabel('Anthropic 接入地址');
  }

  async gotoRoot() {
    await this.page.goto(agentBoxPublicPath('/'));
    await this.page.waitForLoadState('networkidle');
    await expect(this.portalGate).toBeVisible();
  }

  async gotoApp() {
    await this.page.goto(agentBoxPublicPath('/'));
    await this.page.waitForLoadState('networkidle');
    await expect(this.appShell).toBeVisible({ timeout: 15000 });
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

  async openCreateAgentDialog() {
    await this.createAgentButton.click();
    await expect(this.agentDialog).toBeVisible();
  }

  async openCreateProviderDialog() {
    await this.providersViewButton.click();
    await this.createProviderButton.click();
    await expect(this.providerDialog).toBeVisible();
  }

  async logoutAndWaitForPortal() {
    await this.logoutButton.click();
    await expect(this.portalGate).toBeVisible({ timeout: 15000 });
  }
}
