import { mkdirSync } from 'node:fs';
import path from 'node:path';

import { config } from '../../../../../../hack/tests/fixtures/config';

type AgentBoxPage = {
  goto: (url: string) => Promise<unknown>;
  screenshot?: (options: {
    fullPage?: boolean;
    path: string;
  }) => Promise<unknown>;
  waitForLoadState: (state: 'networkidle') => Promise<unknown>;
};

/**
 * AgentBox plugin test helpers keep plugin-specific browser paths in one place.
 * They intentionally avoid reusing host workbench routes or selectors.
 */
export const agentBoxPluginID = 'john-ai-agentbox';

export async function gotoAgentBoxRoot(page: AgentBoxPage) {
  await page.goto(agentBoxPublicPath('/'));
  await page.waitForLoadState('networkidle');
}

export async function gotoAgentBoxLogin(page: AgentBoxPage) {
  await page.goto(agentBoxPublicPath('/login'));
  await page.waitForLoadState('networkidle');
}

export function agentBoxPublicPath(route: string) {
  const normalizedRoute = route.startsWith('/') ? route : `/${route}`;
  return `${config.publicBaseURL.replace(/\/$/, '')}${normalizedRoute}`;
}

export async function captureAgentBoxScreenshot(
  page: AgentBoxPage,
  description: string,
) {
  if (!page.screenshot) {
    return;
  }
  const now = new Date();
  const datePart = [
    now.getFullYear(),
    String(now.getMonth() + 1).padStart(2, '0'),
    String(now.getDate()).padStart(2, '0'),
  ].join('');
  const timePart = [
    String(now.getHours()).padStart(2, '0'),
    String(now.getMinutes()).padStart(2, '0'),
    String(now.getSeconds()).padStart(2, '0'),
  ].join('');
  const safeDescription = description
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');
  const repoRoot = path.resolve(process.cwd(), '../..');
  const screenshotDir = path.join(repoRoot, 'temp', datePart);
  mkdirSync(screenshotDir, { recursive: true });
  await page.screenshot({
    path: path.join(screenshotDir, `${timePart}-${safeDescription}.png`),
    fullPage: false,
  });
}
