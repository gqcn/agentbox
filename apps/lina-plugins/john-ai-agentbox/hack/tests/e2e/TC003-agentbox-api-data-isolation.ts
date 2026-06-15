import { test, expect } from '../../../../../../hack/tests/fixtures/auth';

import { ensureSourcePluginEnabled } from '../../../../../../hack/tests/fixtures/plugin';
import { pluginApiPath } from '../../../../../../hack/tests/fixtures/config';
import {
  agentBoxCurrentSessionToken,
  agentBoxOtherAgentID,
  agentBoxOtherChatSessionID,
  agentBoxOtherTerminalID,
  cleanupAgentBoxIsolationData,
  seedAgentBoxIsolationData,
  setAgentBoxSessionCookie,
} from '../support/isolation';
import {
  agentBoxPluginID,
  captureAgentBoxScreenshot,
  gotoAgentBoxLogin,
} from '../support/plugin';

test.describe('TC003 AgentBox API and resource isolation', () => {
  test.beforeEach(async ({ page }) => {
    await ensureSourcePluginEnabled(page, agentBoxPluginID);
    seedAgentBoxIsolationData();
  });

  test.afterEach(async ({ page }) => {
    await page.context().clearCookies();
    cleanupAgentBoxIsolationData();
  });

  test('TC003a: protected API rejects missing agent_box_session', async ({ page }) => {
    await page.context().clearCookies();
    await gotoAgentBoxLogin(page);
    await captureAgentBoxScreenshot(page, 'tc003a-login-boundary-for-unauthenticated-api');

    const response = await page.request.get(pluginApiPath(agentBoxPluginID, '/agents'));
    await expectBusinessError(response, 'JOHN_AI_AGENTBOX_AUTH_REQUIRED');

    const proxyResponse = await page.request.get(
      pluginApiPath(agentBoxPluginID, '/proxy/e2e-proxy-key/private.html'),
    );
    await expectBusinessError(proxyResponse, 'JOHN_AI_AGENTBOX_AUTH_REQUIRED');
  });

  test('TC003b: current user cannot read another user Agent or Chat session', async ({ page }) => {
    await setAgentBoxSessionCookie(page);

    const agentResponse = await page.request.get(
      pluginApiPath(agentBoxPluginID, `/agents/${agentBoxOtherAgentID}`),
    );
    await expectBusinessError(agentResponse);
    await expectResponseDoesNotLeak(agentResponse, [
      agentBoxOtherAgentID,
      'Other E2E Agent',
    ]);

    const startResponse = await page.request.post(
      pluginApiPath(agentBoxPluginID, `/agents/${agentBoxOtherAgentID}/start`),
    );
    await expectBusinessError(startResponse);
    await expectResponseDoesNotLeak(startResponse, [
      agentBoxOtherAgentID,
      'Other E2E Agent',
    ]);

    const stopResponse = await page.request.post(
      pluginApiPath(agentBoxPluginID, `/agents/${agentBoxOtherAgentID}/stop`),
    );
    await expectBusinessError(stopResponse);
    await expectResponseDoesNotLeak(stopResponse, [
      agentBoxOtherAgentID,
      'Other E2E Agent',
    ]);

    const logsResponse = await page.request.get(
      pluginApiPath(agentBoxPluginID, `/agents/${agentBoxOtherAgentID}/logs`),
    );
    await expectBusinessError(logsResponse);
    await expectResponseDoesNotLeak(logsResponse, [
      agentBoxOtherAgentID,
      'Other E2E Agent',
    ]);

    const chatResponse = await page.request.get(
      pluginApiPath(
        agentBoxPluginID,
        `/agents/${agentBoxOtherAgentID}/chat/sessions/${agentBoxOtherChatSessionID}/messages`,
      ),
    );
    await expectBusinessError(chatResponse);
    await expectResponseDoesNotLeak(chatResponse, [
      agentBoxOtherAgentID,
      agentBoxOtherChatSessionID,
      'Other User Secret Chat',
    ]);

    const chatSocketResponse = await page.request.get(
      pluginApiPath(
        agentBoxPluginID,
        `/ws/agents/${agentBoxOtherAgentID}/chat/sessions/${agentBoxOtherChatSessionID}`,
      ),
    );
    await expectBusinessError(chatSocketResponse);
    await expectResponseDoesNotLeak(chatSocketResponse, [
      agentBoxOtherAgentID,
      agentBoxOtherChatSessionID,
      'Other User Secret Chat',
    ]);
    await captureAgentBoxScreenshot(page, 'tc003b-cross-user-api-denied');
  });

  test('TC003c: current user cannot access another user workspace or service proxy route', async ({ page }) => {
    await setAgentBoxSessionCookie(page);

    const workspaceResponse = await page.request.get(
      pluginApiPath(
        agentBoxPluginID,
        `/agents/${agentBoxOtherAgentID}/workspace/tree?path=/home/agent/workspace&includeFiles=true`,
      ),
    );
    await expectBusinessError(workspaceResponse);
    await expectResponseDoesNotLeak(workspaceResponse, [
      agentBoxOtherAgentID,
      '/home/agent/workspace',
      'Other E2E Agent',
    ]);

    const downloadResponse = await page.request.get(
      pluginApiPath(
        agentBoxPluginID,
        `/agents/${agentBoxOtherAgentID}/workspace/download?path=/home/agent/workspace/private.txt`,
      ),
    );
    await expectBusinessError(downloadResponse);
    await expectResponseDoesNotLeak(downloadResponse, [
      agentBoxOtherAgentID,
      '/home/agent/workspace/private.txt',
      'private.txt',
      'Other E2E Agent',
    ]);

    const resourceResponse = await page.request.get(
      pluginApiPath(
        agentBoxPluginID,
        `/agents/${agentBoxOtherAgentID}/workspace/resources?path=/home/agent/workspace/private.txt&disposition=inline`,
      ),
    );
    await expectBusinessError(resourceResponse);
    await expectResponseDoesNotLeak(resourceResponse, [
      agentBoxOtherAgentID,
      '/home/agent/workspace/private.txt',
      'private.txt',
      'Other E2E Agent',
    ]);

    const htmlPreviewResponse = await page.request.get(
      pluginApiPath(
        agentBoxPluginID,
        `/agents/${agentBoxOtherAgentID}/workspace/html-previews?path=/home/agent/workspace/private.html`,
      ),
    );
    await expectBusinessError(htmlPreviewResponse);
    await expectResponseDoesNotLeak(htmlPreviewResponse, [
      agentBoxOtherAgentID,
      '/home/agent/workspace/private.html',
      'private.html',
      'Other E2E Agent',
    ]);

    const skillsResponse = await page.request.get(
      pluginApiPath(
        agentBoxPluginID,
        `/agents/${agentBoxOtherAgentID}/skills?scope=project&path=/home/agent/workspace&query=secret`,
      ),
    );
    await expectBusinessError(skillsResponse);
    await expectResponseDoesNotLeak(skillsResponse, [
      agentBoxOtherAgentID,
      '/home/agent/workspace',
      'secret',
      'Other E2E Agent',
    ]);

    const gitResponse = await page.request.get(
      pluginApiPath(
        agentBoxPluginID,
        `/agents/${agentBoxOtherAgentID}/git/file?path=/home/agent/workspace&file=private.ts`,
      ),
    );
    await expectBusinessError(gitResponse);
    await expectResponseDoesNotLeak(gitResponse, [
      agentBoxOtherAgentID,
      '/home/agent/workspace',
      'private.ts',
      'Other E2E Agent',
    ]);

    const shellSocketResponse = await page.request.get(
      pluginApiPath(
        agentBoxPluginID,
        `/ws/agents/${agentBoxOtherAgentID}/shell?terminalId=${agentBoxOtherTerminalID}&cwd=/home/agent/workspace`,
      ),
    );
    await expectBusinessError(shellSocketResponse);
    await expectResponseDoesNotLeak(shellSocketResponse, [
      agentBoxOtherAgentID,
      agentBoxOtherTerminalID,
      '/home/agent/workspace',
      'Other E2E Agent',
    ]);

    const terminalResponse = await page.request.get(
      pluginApiPath(
        agentBoxPluginID,
        `/agents/${agentBoxOtherAgentID}/terminal/sessions/${agentBoxOtherTerminalID}`,
      ),
    );
    await expectBusinessError(terminalResponse);
    await expectResponseDoesNotLeak(terminalResponse, [
      agentBoxOtherAgentID,
      agentBoxOtherTerminalID,
      'abx-other-secret-backend',
      '/home/agent/workspace/secret',
      '/bin/zsh',
      'Other E2E Agent',
    ]);

    const serviceResponse = await page.request.get(
      pluginApiPath(
        agentBoxPluginID,
        `/agents/${agentBoxOtherAgentID}/services/svc-secret-runtime`,
      ),
    );
    await expectBusinessError(serviceResponse);
    await expectResponseDoesNotLeak(serviceResponse, [
      agentBoxOtherAgentID,
      'svc-secret-runtime',
      'Other E2E Agent',
    ]);

    const tunnelResponse = await page.request.get(
      pluginApiPath(
        agentBoxPluginID,
        `/ws/agents/${agentBoxOtherAgentID}/services/svc-secret-runtime/tcp?key=tunnel-secret`,
      ),
    );
    await expectBusinessError(tunnelResponse);
    await expectResponseDoesNotLeak(tunnelResponse, [
      agentBoxOtherAgentID,
      'svc-secret-runtime',
      'tunnel-secret',
      'Other E2E Agent',
    ]);

    const bridgeResponse = await page.request.post(
      pluginApiPath(agentBoxPluginID, `/agents/${agentBoxOtherAgentID}/service-bridges`),
      {
        data: {
          serviceId: 'svc-secret-runtime',
          listenAddress: '127.0.0.1',
          port: 3000,
        },
      },
    );
    await expectBusinessError(bridgeResponse);
    await expectResponseDoesNotLeak(bridgeResponse, [
      agentBoxOtherAgentID,
      'svc-secret-runtime',
      'Other E2E Agent',
    ]);
    await captureAgentBoxScreenshot(page, 'tc003c-cross-user-proxy-denied');
  });
});

async function expectBusinessError(
  response: { json: () => Promise<any>; status: () => number },
  expectedErrorCode?: string,
) {
  expect(response.status()).toBe(200);
  const payload = await response.json();
  expect(payload.code).not.toBe(0);
  if (expectedErrorCode) {
    expect(payload.errorCode).toBe(expectedErrorCode);
  }
}

async function expectResponseDoesNotLeak(
  response: { text: () => Promise<string> },
  forbiddenValues: string[],
) {
  const body = await response.text();
  for (const value of forbiddenValues) {
    expect(body).not.toContain(value);
  }
  expect(body).not.toContain(agentBoxCurrentSessionToken);
}
