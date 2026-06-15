import { createHash } from 'node:crypto';

import { config } from '../../../../../../hack/tests/fixtures/config';
import {
  execPgSQLStatements,
  pgEscapeLiteral,
} from '../../../../../../hack/tests/support/postgres';

export const agentBoxCurrentSessionToken = 'e2e-agentbox-current-session-token';
export const agentBoxCurrentUserID = 'usr-e2e-agentbox-current';
export const agentBoxOtherUserID = 'usr-e2e-agentbox-other';
export const agentBoxCurrentAgentID = 'agt-e2e-agentbox-current';
export const agentBoxOtherAgentID = 'agt-e2e-agentbox-other';
export const agentBoxOtherChatSessionID = 'chat-e2e-agentbox-other';
export const agentBoxOtherTerminalID = 'term-e2e-agentbox-secret';

type CookieCapableContext = {
  addCookies: (cookies: readonly AgentBoxCookie[]) => Promise<unknown>;
};

type AgentBoxCookiePage = {
  context: () => CookieCapableContext;
};

type AgentBoxCookie = {
  domain?: string;
  expires?: number;
  httpOnly?: boolean;
  name: string;
  path?: string;
  sameSite?: 'Strict' | 'Lax' | 'None';
  secure?: boolean;
  url?: string;
  value: string;
};

const providerName = 'E2E AgentBox Provider';
const modelName = 'e2e-model';

/**
 * Seeds minimal plugin-owned data for AgentBox API isolation assertions.
 * The rows use stable E2E-only business keys and are cleaned before and after
 * each test so cross-user access failures can be asserted independently.
 */
export function seedAgentBoxIsolationData() {
  cleanupAgentBoxIsolationData();
  const tokenHash = createHash('sha256')
    .update(agentBoxCurrentSessionToken)
    .digest('hex');
  const expiresAt = new Date(Date.now() + 60 * 60 * 1000).toISOString();

  execPgSQLStatements([
    `INSERT INTO john_ai_agentbox_users (id, username, password_hash, role, status) VALUES ('${pgEscapeLiteral(agentBoxCurrentUserID)}', 'e2e-agentbox-current', 'unused', 'user', 'active') ON CONFLICT (id) DO NOTHING;`,
    `INSERT INTO john_ai_agentbox_users (id, username, password_hash, role, status) VALUES ('${pgEscapeLiteral(agentBoxOtherUserID)}', 'e2e-agentbox-other', 'unused', 'user', 'active') ON CONFLICT (id) DO NOTHING;`,
    `INSERT INTO john_ai_agentbox_user_sessions (token_hash, user_id, user_agent, ip_address, expires_at) VALUES ('${tokenHash}', '${pgEscapeLiteral(agentBoxCurrentUserID)}', 'playwright', '127.0.0.1', '${expiresAt}') ON CONFLICT (token_hash) DO NOTHING;`,
    `WITH provider_row AS (
      INSERT INTO john_ai_agentbox_ai_providers (name, homepage_url, notes)
      VALUES ('${pgEscapeLiteral(providerName)}', 'https://example.test', 'E2E isolation provider')
      RETURNING id
    ), image_row AS (
      SELECT id FROM john_ai_agentbox_coding_images ORDER BY id LIMIT 1
    )
    INSERT INTO john_ai_agentbox_coding_agents (id, user_id, name, provider_id, model_name, model_protocol, image_id, agent_type, icon_key, notes)
    SELECT '${pgEscapeLiteral(agentBoxCurrentAgentID)}', '${pgEscapeLiteral(agentBoxCurrentUserID)}', 'Current E2E Agent', provider_row.id, '${pgEscapeLiteral(modelName)}', 'openai', image_row.id, 'codex', 'terminal', 'current user agent'
    FROM provider_row, image_row;`,
    `INSERT INTO john_ai_agentbox_coding_agents (id, user_id, name, provider_id, model_name, model_protocol, image_id, agent_type, icon_key, notes)
    SELECT '${pgEscapeLiteral(agentBoxOtherAgentID)}', '${pgEscapeLiteral(agentBoxOtherUserID)}', 'Other E2E Agent', p.id, '${pgEscapeLiteral(modelName)}', 'openai', i.id, 'codex', 'terminal', 'other user agent'
    FROM john_ai_agentbox_ai_providers p
    CROSS JOIN LATERAL (SELECT id FROM john_ai_agentbox_coding_images ORDER BY id LIMIT 1) i
    WHERE p.name = '${pgEscapeLiteral(providerName)}'
    ORDER BY p.id DESC
    LIMIT 1;`,
    `INSERT INTO john_ai_agentbox_agent_chat_sessions (id, agent_id, title, status, tool_type, runtime_state)
    VALUES ('${pgEscapeLiteral(agentBoxOtherChatSessionID)}', '${pgEscapeLiteral(agentBoxOtherAgentID)}', 'Other User Secret Chat', 'idle', 'codex', 'idle');`,
    `INSERT INTO john_ai_agentbox_agent_terminal_sessions (id, user_id, agent_id, terminal_id, backend_type, backend_session_name, working_dir, shell, status)
    VALUES ('term-row-e2e-agentbox-secret', '${pgEscapeLiteral(agentBoxOtherUserID)}', '${pgEscapeLiteral(agentBoxOtherAgentID)}', '${pgEscapeLiteral(agentBoxOtherTerminalID)}', 'tmux', 'abx-other-secret-backend', '/home/agent/workspace/secret', '/bin/zsh', 'active');`,
  ]);
}

export function cleanupAgentBoxIsolationData() {
  execPgSQLStatements([
    `DELETE FROM john_ai_agentbox_users WHERE id IN ('${pgEscapeLiteral(agentBoxCurrentUserID)}', '${pgEscapeLiteral(agentBoxOtherUserID)}');`,
    `DELETE FROM john_ai_agentbox_ai_providers WHERE name = '${pgEscapeLiteral(providerName)}';`,
  ]);
}

export async function setAgentBoxSessionCookie(page: AgentBoxCookiePage) {
  await page.context().addCookies([
    {
      name: 'agent_box_session',
      value: agentBoxCurrentSessionToken,
      url: config.publicBaseURL || config.baseURL,
      httpOnly: true,
      sameSite: 'Lax',
    },
  ]);
}
