CREATE TABLE IF NOT EXISTS john_ai_agentbox_users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    status TEXT NOT NULL DEFAULT 'active',
    last_login_at TIMESTAMPTZ NULL,
    deleted_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS john_ai_agentbox_idx_users_username_active
    ON john_ai_agentbox_users (LOWER(username))
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS john_ai_agentbox_idx_users_status
    ON john_ai_agentbox_users (status);

CREATE TABLE IF NOT EXISTS john_ai_agentbox_user_sessions (
    token_hash TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES john_ai_agentbox_users(id) ON DELETE CASCADE,
    user_agent TEXT NOT NULL DEFAULT '',
    ip_address TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS john_ai_agentbox_idx_user_sessions_user
    ON john_ai_agentbox_user_sessions (user_id);

CREATE INDEX IF NOT EXISTS john_ai_agentbox_idx_user_sessions_expires
    ON john_ai_agentbox_user_sessions (expires_at);

CREATE TABLE IF NOT EXISTS john_ai_agentbox_settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS john_ai_agentbox_user_settings (
    user_id TEXT NOT NULL REFERENCES john_ai_agentbox_users(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    value TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, key)
);

CREATE TABLE IF NOT EXISTS john_ai_agentbox_ai_providers (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    homepage_url TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    api_key TEXT NOT NULL DEFAULT '',
    openai_base_url TEXT NOT NULL DEFAULT '',
    anthropic_base_url TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS john_ai_agentbox_provider_models (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    provider_id BIGINT NOT NULL REFERENCES john_ai_agentbox_ai_providers(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    protocol TEXT NOT NULL,
    source TEXT NOT NULL,
    last_synced_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider_id, name, protocol)
);

CREATE INDEX IF NOT EXISTS john_ai_agentbox_idx_provider_models_provider
    ON john_ai_agentbox_provider_models (provider_id);

CREATE TABLE IF NOT EXISTS john_ai_agentbox_coding_images (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL,
    image_ref TEXT NOT NULL UNIQUE,
    agent_type TEXT NOT NULL,
    default_shell TEXT NOT NULL DEFAULT '/bin/bash',
    notes TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS john_ai_agentbox_coding_agents (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES john_ai_agentbox_users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    provider_id BIGINT NOT NULL REFERENCES john_ai_agentbox_ai_providers(id),
    model_name TEXT NOT NULL,
    model_protocol TEXT NOT NULL,
    image_id BIGINT NOT NULL REFERENCES john_ai_agentbox_coding_images(id),
    agent_type TEXT NOT NULL,
    icon_key TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    deleted_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS john_ai_agentbox_idx_agents_provider
    ON john_ai_agentbox_coding_agents (provider_id);

CREATE INDEX IF NOT EXISTS john_ai_agentbox_idx_agents_image
    ON john_ai_agentbox_coding_agents (image_id);

CREATE INDEX IF NOT EXISTS john_ai_agentbox_idx_agents_user_deleted
    ON john_ai_agentbox_coding_agents (user_id, deleted_at);

CREATE INDEX IF NOT EXISTS john_ai_agentbox_idx_agents_user_updated
    ON john_ai_agentbox_coding_agents (user_id, deleted_at, updated_at DESC);

CREATE TABLE IF NOT EXISTS john_ai_agentbox_agent_runtimes (
    agent_id TEXT PRIMARY KEY REFERENCES john_ai_agentbox_coding_agents(id) ON DELETE CASCADE,
    container_id TEXT NOT NULL DEFAULT '',
    docker_id TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT '',
    config_mount_path TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS john_ai_agentbox_agent_chat_sessions (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL REFERENCES john_ai_agentbox_coding_agents(id) ON DELETE CASCADE,
    title TEXT NOT NULL DEFAULT '新对话',
    status TEXT NOT NULL DEFAULT 'idle',
    tool_type TEXT NOT NULL DEFAULT '',
    tool_session_id TEXT NOT NULL DEFAULT '',
    runtime_state TEXT NOT NULL DEFAULT 'idle',
    last_error TEXT NOT NULL DEFAULT '',
    message_count BIGINT NOT NULL DEFAULT 0,
    last_message_preview TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_active_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS john_ai_agentbox_idx_chat_sessions_agent
    ON john_ai_agentbox_agent_chat_sessions (agent_id);

CREATE INDEX IF NOT EXISTS john_ai_agentbox_idx_chat_sessions_agent_last_active
    ON john_ai_agentbox_agent_chat_sessions (agent_id, last_active_at DESC);

CREATE TABLE IF NOT EXISTS john_ai_agentbox_agent_chat_messages (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES john_ai_agentbox_agent_chat_sessions(id) ON DELETE CASCADE,
    sequence BIGINT NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'complete',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (session_id, sequence)
);

CREATE INDEX IF NOT EXISTS john_ai_agentbox_idx_chat_messages_session_sequence
    ON john_ai_agentbox_agent_chat_messages (session_id, sequence);

CREATE TABLE IF NOT EXISTS john_ai_agentbox_agent_chat_interactions (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL REFERENCES john_ai_agentbox_coding_agents(id) ON DELETE CASCADE,
    session_id TEXT NOT NULL REFERENCES john_ai_agentbox_agent_chat_sessions(id) ON DELETE CASCADE,
    assistant_message_id BIGINT NULL REFERENCES john_ai_agentbox_agent_chat_messages(id) ON DELETE SET NULL,
    tool_type TEXT NOT NULL DEFAULT '',
    tool_interaction_id TEXT NOT NULL DEFAULT '',
    interaction_type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    title TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    risk_level TEXT NOT NULL DEFAULT 'low',
    payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    response_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    response_mode TEXT NOT NULL DEFAULT '',
    response_scope TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ NULL,
    resolved_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS john_ai_agentbox_idx_chat_interactions_agent_session
    ON john_ai_agentbox_agent_chat_interactions (agent_id, session_id, created_at DESC);

CREATE INDEX IF NOT EXISTS john_ai_agentbox_idx_chat_interactions_session_status
    ON john_ai_agentbox_agent_chat_interactions (session_id, status, created_at ASC);

CREATE INDEX IF NOT EXISTS john_ai_agentbox_idx_chat_interactions_pending
    ON john_ai_agentbox_agent_chat_interactions (session_id, created_at ASC)
    WHERE status = 'pending';

CREATE TABLE IF NOT EXISTS john_ai_agentbox_agent_terminal_sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL REFERENCES john_ai_agentbox_users(id) ON DELETE CASCADE,
    agent_id TEXT NOT NULL REFERENCES john_ai_agentbox_coding_agents(id) ON DELETE CASCADE,
    terminal_id TEXT NOT NULL,
    backend_type TEXT NOT NULL,
    backend_session_name TEXT NOT NULL,
    working_dir TEXT NOT NULL DEFAULT '',
    shell TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
    last_error TEXT NOT NULL DEFAULT '',
    closed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, agent_id, terminal_id)
);

CREATE INDEX IF NOT EXISTS john_ai_agentbox_idx_terminal_sessions_agent_status
    ON john_ai_agentbox_agent_terminal_sessions (user_id, agent_id, status);

CREATE UNIQUE INDEX IF NOT EXISTS john_ai_agentbox_idx_terminal_sessions_backend_name
    ON john_ai_agentbox_agent_terminal_sessions (backend_type, backend_session_name);

CREATE TABLE IF NOT EXISTS john_ai_agentbox_ai_capability_tiers (
    code TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS john_ai_agentbox_ai_capability_bindings (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tier_code TEXT NOT NULL REFERENCES john_ai_agentbox_ai_capability_tiers(code),
    provider_id BIGINT NOT NULL REFERENCES john_ai_agentbox_ai_providers(id),
    provider_model_id BIGINT NOT NULL REFERENCES john_ai_agentbox_provider_models(id),
    priority INTEGER NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tier_code, priority)
);

CREATE INDEX IF NOT EXISTS john_ai_agentbox_idx_ai_bindings_tier_priority
    ON john_ai_agentbox_ai_capability_bindings (tier_code, priority);

CREATE INDEX IF NOT EXISTS john_ai_agentbox_idx_ai_bindings_provider
    ON john_ai_agentbox_ai_capability_bindings (provider_id);

CREATE INDEX IF NOT EXISTS john_ai_agentbox_idx_ai_bindings_model
    ON john_ai_agentbox_ai_capability_bindings (provider_model_id);

CREATE TABLE IF NOT EXISTS john_ai_agentbox_ai_invocation_logs (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    purpose TEXT NOT NULL,
    tier_code TEXT NOT NULL,
    provider_id BIGINT NULL REFERENCES john_ai_agentbox_ai_providers(id) ON DELETE SET NULL,
    provider_model_id BIGINT NULL REFERENCES john_ai_agentbox_provider_models(id) ON DELETE SET NULL,
    model_name TEXT NOT NULL DEFAULT '',
    protocol TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    latency_ms BIGINT NOT NULL DEFAULT 0,
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS john_ai_agentbox_idx_ai_invocation_logs_created_at
    ON john_ai_agentbox_ai_invocation_logs (created_at DESC);

CREATE INDEX IF NOT EXISTS john_ai_agentbox_idx_ai_invocation_logs_filters
    ON john_ai_agentbox_ai_invocation_logs (purpose, tier_code, status);

CREATE TABLE IF NOT EXISTS john_ai_agentbox_system_prompt_overrides (
    code TEXT PRIMARY KEY,
    content TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE john_ai_agentbox_users IS 'AgentBox user accounts';
COMMENT ON COLUMN john_ai_agentbox_users.id IS 'User ID';
COMMENT ON COLUMN john_ai_agentbox_users.username IS 'Login username';
COMMENT ON COLUMN john_ai_agentbox_users.password_hash IS 'BCrypt password hash';
COMMENT ON COLUMN john_ai_agentbox_users.role IS 'AgentBox role';
COMMENT ON COLUMN john_ai_agentbox_users.status IS 'User status';
COMMENT ON COLUMN john_ai_agentbox_users.last_login_at IS 'Last successful login time';
COMMENT ON COLUMN john_ai_agentbox_users.deleted_at IS 'Soft deletion time';
COMMENT ON COLUMN john_ai_agentbox_users.created_at IS 'Creation time';
COMMENT ON COLUMN john_ai_agentbox_users.updated_at IS 'Update time';

COMMENT ON TABLE john_ai_agentbox_user_sessions IS 'AgentBox user login sessions';
COMMENT ON COLUMN john_ai_agentbox_user_sessions.token_hash IS 'Opaque session token hash';
COMMENT ON COLUMN john_ai_agentbox_user_sessions.user_id IS 'Owner user ID';
COMMENT ON COLUMN john_ai_agentbox_user_sessions.user_agent IS 'Client user agent';
COMMENT ON COLUMN john_ai_agentbox_user_sessions.ip_address IS 'Client IP address';
COMMENT ON COLUMN john_ai_agentbox_user_sessions.expires_at IS 'Session expiration time';
COMMENT ON COLUMN john_ai_agentbox_user_sessions.revoked_at IS 'Session revocation time';
COMMENT ON COLUMN john_ai_agentbox_user_sessions.created_at IS 'Creation time';
COMMENT ON COLUMN john_ai_agentbox_user_sessions.updated_at IS 'Update time';

COMMENT ON TABLE john_ai_agentbox_settings IS 'Global AgentBox settings';
COMMENT ON COLUMN john_ai_agentbox_settings.key IS 'Setting key';
COMMENT ON COLUMN john_ai_agentbox_settings.value IS 'Setting value';
COMMENT ON COLUMN john_ai_agentbox_settings.created_at IS 'Creation time';
COMMENT ON COLUMN john_ai_agentbox_settings.updated_at IS 'Update time';

COMMENT ON TABLE john_ai_agentbox_user_settings IS 'Per-user AgentBox settings';
COMMENT ON COLUMN john_ai_agentbox_user_settings.user_id IS 'Owner user ID';
COMMENT ON COLUMN john_ai_agentbox_user_settings.key IS 'Setting key';
COMMENT ON COLUMN john_ai_agentbox_user_settings.value IS 'Setting value';
COMMENT ON COLUMN john_ai_agentbox_user_settings.created_at IS 'Creation time';
COMMENT ON COLUMN john_ai_agentbox_user_settings.updated_at IS 'Update time';

COMMENT ON TABLE john_ai_agentbox_ai_providers IS 'AI provider credentials and endpoints';
COMMENT ON COLUMN john_ai_agentbox_ai_providers.id IS 'Provider ID';
COMMENT ON COLUMN john_ai_agentbox_ai_providers.name IS 'Provider display name';
COMMENT ON COLUMN john_ai_agentbox_ai_providers.homepage_url IS 'Provider homepage URL';
COMMENT ON COLUMN john_ai_agentbox_ai_providers.notes IS 'Provider notes';
COMMENT ON COLUMN john_ai_agentbox_ai_providers.api_key IS 'Provider API key';
COMMENT ON COLUMN john_ai_agentbox_ai_providers.openai_base_url IS 'OpenAI-compatible base URL';
COMMENT ON COLUMN john_ai_agentbox_ai_providers.anthropic_base_url IS 'Anthropic-compatible base URL';
COMMENT ON COLUMN john_ai_agentbox_ai_providers.created_at IS 'Creation time';
COMMENT ON COLUMN john_ai_agentbox_ai_providers.updated_at IS 'Update time';

COMMENT ON TABLE john_ai_agentbox_provider_models IS 'Models exposed by AI providers';
COMMENT ON COLUMN john_ai_agentbox_provider_models.id IS 'Provider model ID';
COMMENT ON COLUMN john_ai_agentbox_provider_models.provider_id IS 'Provider ID';
COMMENT ON COLUMN john_ai_agentbox_provider_models.name IS 'Model name';
COMMENT ON COLUMN john_ai_agentbox_provider_models.protocol IS 'Model protocol';
COMMENT ON COLUMN john_ai_agentbox_provider_models.source IS 'Model source';
COMMENT ON COLUMN john_ai_agentbox_provider_models.last_synced_at IS 'Last remote synchronization time';
COMMENT ON COLUMN john_ai_agentbox_provider_models.created_at IS 'Creation time';
COMMENT ON COLUMN john_ai_agentbox_provider_models.updated_at IS 'Update time';

COMMENT ON TABLE john_ai_agentbox_coding_images IS 'Container images available to coding agents';
COMMENT ON COLUMN john_ai_agentbox_coding_images.id IS 'Coding image ID';
COMMENT ON COLUMN john_ai_agentbox_coding_images.name IS 'Image display name';
COMMENT ON COLUMN john_ai_agentbox_coding_images.image_ref IS 'Container image reference';
COMMENT ON COLUMN john_ai_agentbox_coding_images.agent_type IS 'Agent runtime type';
COMMENT ON COLUMN john_ai_agentbox_coding_images.default_shell IS 'Default shell path';
COMMENT ON COLUMN john_ai_agentbox_coding_images.notes IS 'Image notes';
COMMENT ON COLUMN john_ai_agentbox_coding_images.enabled IS 'Whether the image is enabled';
COMMENT ON COLUMN john_ai_agentbox_coding_images.is_default IS 'Whether the image is a default option';
COMMENT ON COLUMN john_ai_agentbox_coding_images.created_at IS 'Creation time';
COMMENT ON COLUMN john_ai_agentbox_coding_images.updated_at IS 'Update time';

COMMENT ON TABLE john_ai_agentbox_coding_agents IS 'User-owned coding agents';
COMMENT ON COLUMN john_ai_agentbox_coding_agents.id IS 'Agent ID';
COMMENT ON COLUMN john_ai_agentbox_coding_agents.user_id IS 'Owner user ID';
COMMENT ON COLUMN john_ai_agentbox_coding_agents.name IS 'Agent display name';
COMMENT ON COLUMN john_ai_agentbox_coding_agents.provider_id IS 'Provider ID';
COMMENT ON COLUMN john_ai_agentbox_coding_agents.model_name IS 'Selected model name';
COMMENT ON COLUMN john_ai_agentbox_coding_agents.model_protocol IS 'Selected model protocol';
COMMENT ON COLUMN john_ai_agentbox_coding_agents.image_id IS 'Coding image ID';
COMMENT ON COLUMN john_ai_agentbox_coding_agents.agent_type IS 'Agent runtime type';
COMMENT ON COLUMN john_ai_agentbox_coding_agents.icon_key IS 'Agent icon key';
COMMENT ON COLUMN john_ai_agentbox_coding_agents.notes IS 'Agent notes';
COMMENT ON COLUMN john_ai_agentbox_coding_agents.deleted_at IS 'Soft deletion time';
COMMENT ON COLUMN john_ai_agentbox_coding_agents.created_at IS 'Creation time';
COMMENT ON COLUMN john_ai_agentbox_coding_agents.updated_at IS 'Update time';

COMMENT ON TABLE john_ai_agentbox_agent_runtimes IS 'Runtime state for coding agents';
COMMENT ON COLUMN john_ai_agentbox_agent_runtimes.agent_id IS 'Agent ID';
COMMENT ON COLUMN john_ai_agentbox_agent_runtimes.container_id IS 'AgentBox logical container ID';
COMMENT ON COLUMN john_ai_agentbox_agent_runtimes.docker_id IS 'Docker container ID';
COMMENT ON COLUMN john_ai_agentbox_agent_runtimes.status IS 'Runtime status';
COMMENT ON COLUMN john_ai_agentbox_agent_runtimes.config_mount_path IS 'Runtime configuration mount path';
COMMENT ON COLUMN john_ai_agentbox_agent_runtimes.created_at IS 'Creation time';
COMMENT ON COLUMN john_ai_agentbox_agent_runtimes.updated_at IS 'Update time';

COMMENT ON TABLE john_ai_agentbox_agent_chat_sessions IS 'Agent chat sessions';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_sessions.id IS 'Chat session ID';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_sessions.agent_id IS 'Agent ID';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_sessions.title IS 'Session title';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_sessions.status IS 'Session status';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_sessions.tool_type IS 'Connected tool type';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_sessions.tool_session_id IS 'Connected tool session ID';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_sessions.runtime_state IS 'Runtime state';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_sessions.last_error IS 'Last runtime error';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_sessions.message_count IS 'Message count';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_sessions.last_message_preview IS 'Latest message preview';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_sessions.created_at IS 'Creation time';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_sessions.updated_at IS 'Update time';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_sessions.last_active_at IS 'Last activity time';

COMMENT ON TABLE john_ai_agentbox_agent_chat_messages IS 'Messages in agent chat sessions';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_messages.id IS 'Chat message ID';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_messages.session_id IS 'Chat session ID';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_messages.sequence IS 'Message sequence in the session';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_messages.role IS 'Message role';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_messages.content IS 'Message content';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_messages.status IS 'Message status';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_messages.metadata IS 'Message metadata JSON';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_messages.created_at IS 'Creation time';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_messages.updated_at IS 'Update time';

COMMENT ON TABLE john_ai_agentbox_agent_chat_interactions IS 'Interactive requests produced during chat sessions';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_interactions.id IS 'Interaction ID';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_interactions.agent_id IS 'Agent ID';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_interactions.session_id IS 'Chat session ID';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_interactions.assistant_message_id IS 'Related assistant message ID';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_interactions.tool_type IS 'Tool type';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_interactions.tool_interaction_id IS 'External tool interaction ID';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_interactions.interaction_type IS 'Interaction type';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_interactions.status IS 'Interaction status';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_interactions.title IS 'Interaction title';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_interactions.body IS 'Interaction body';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_interactions.risk_level IS 'Risk level';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_interactions.payload_json IS 'Interaction payload JSON';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_interactions.response_json IS 'Interaction response JSON';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_interactions.response_mode IS 'Response mode';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_interactions.response_scope IS 'Response scope';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_interactions.expires_at IS 'Expiration time';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_interactions.resolved_at IS 'Resolution time';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_interactions.created_at IS 'Creation time';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_interactions.updated_at IS 'Update time';
COMMENT ON COLUMN john_ai_agentbox_agent_chat_interactions.deleted_at IS 'Soft deletion time';

COMMENT ON TABLE john_ai_agentbox_agent_terminal_sessions IS 'Agent terminal sessions';
COMMENT ON COLUMN john_ai_agentbox_agent_terminal_sessions.id IS 'Terminal session ID';
COMMENT ON COLUMN john_ai_agentbox_agent_terminal_sessions.user_id IS 'Owner user ID';
COMMENT ON COLUMN john_ai_agentbox_agent_terminal_sessions.agent_id IS 'Agent ID';
COMMENT ON COLUMN john_ai_agentbox_agent_terminal_sessions.terminal_id IS 'Frontend terminal ID';
COMMENT ON COLUMN john_ai_agentbox_agent_terminal_sessions.backend_type IS 'Terminal backend type';
COMMENT ON COLUMN john_ai_agentbox_agent_terminal_sessions.backend_session_name IS 'Terminal backend session name';
COMMENT ON COLUMN john_ai_agentbox_agent_terminal_sessions.working_dir IS 'Working directory';
COMMENT ON COLUMN john_ai_agentbox_agent_terminal_sessions.shell IS 'Shell path';
COMMENT ON COLUMN john_ai_agentbox_agent_terminal_sessions.status IS 'Terminal session status';
COMMENT ON COLUMN john_ai_agentbox_agent_terminal_sessions.last_error IS 'Last terminal error';
COMMENT ON COLUMN john_ai_agentbox_agent_terminal_sessions.closed_at IS 'Close time';
COMMENT ON COLUMN john_ai_agentbox_agent_terminal_sessions.created_at IS 'Creation time';
COMMENT ON COLUMN john_ai_agentbox_agent_terminal_sessions.updated_at IS 'Update time';

COMMENT ON TABLE john_ai_agentbox_ai_capability_tiers IS 'AI capability tiers';
COMMENT ON COLUMN john_ai_agentbox_ai_capability_tiers.code IS 'Capability tier code';
COMMENT ON COLUMN john_ai_agentbox_ai_capability_tiers.display_name IS 'Capability tier display name';
COMMENT ON COLUMN john_ai_agentbox_ai_capability_tiers.description IS 'Capability tier description';
COMMENT ON COLUMN john_ai_agentbox_ai_capability_tiers.enabled IS 'Whether the tier is enabled';
COMMENT ON COLUMN john_ai_agentbox_ai_capability_tiers.created_at IS 'Creation time';
COMMENT ON COLUMN john_ai_agentbox_ai_capability_tiers.updated_at IS 'Update time';

COMMENT ON TABLE john_ai_agentbox_ai_capability_bindings IS 'AI capability tier provider bindings';
COMMENT ON COLUMN john_ai_agentbox_ai_capability_bindings.id IS 'Capability binding ID';
COMMENT ON COLUMN john_ai_agentbox_ai_capability_bindings.tier_code IS 'Capability tier code';
COMMENT ON COLUMN john_ai_agentbox_ai_capability_bindings.provider_id IS 'Provider ID';
COMMENT ON COLUMN john_ai_agentbox_ai_capability_bindings.provider_model_id IS 'Provider model ID';
COMMENT ON COLUMN john_ai_agentbox_ai_capability_bindings.priority IS 'Binding priority';
COMMENT ON COLUMN john_ai_agentbox_ai_capability_bindings.enabled IS 'Whether the binding is enabled';
COMMENT ON COLUMN john_ai_agentbox_ai_capability_bindings.created_at IS 'Creation time';
COMMENT ON COLUMN john_ai_agentbox_ai_capability_bindings.updated_at IS 'Update time';

COMMENT ON TABLE john_ai_agentbox_ai_invocation_logs IS 'AI invocation audit logs';
COMMENT ON COLUMN john_ai_agentbox_ai_invocation_logs.id IS 'AI invocation log ID';
COMMENT ON COLUMN john_ai_agentbox_ai_invocation_logs.purpose IS 'Invocation purpose';
COMMENT ON COLUMN john_ai_agentbox_ai_invocation_logs.tier_code IS 'Capability tier code';
COMMENT ON COLUMN john_ai_agentbox_ai_invocation_logs.provider_id IS 'Provider ID';
COMMENT ON COLUMN john_ai_agentbox_ai_invocation_logs.provider_model_id IS 'Provider model ID';
COMMENT ON COLUMN john_ai_agentbox_ai_invocation_logs.model_name IS 'Model name used for the invocation';
COMMENT ON COLUMN john_ai_agentbox_ai_invocation_logs.protocol IS 'Model protocol';
COMMENT ON COLUMN john_ai_agentbox_ai_invocation_logs.status IS 'Invocation status';
COMMENT ON COLUMN john_ai_agentbox_ai_invocation_logs.latency_ms IS 'Invocation latency in milliseconds';
COMMENT ON COLUMN john_ai_agentbox_ai_invocation_logs.error_message IS 'Invocation error message';
COMMENT ON COLUMN john_ai_agentbox_ai_invocation_logs.created_at IS 'Creation time';

COMMENT ON TABLE john_ai_agentbox_system_prompt_overrides IS 'System prompt override content';
COMMENT ON COLUMN john_ai_agentbox_system_prompt_overrides.code IS 'Prompt override code';
COMMENT ON COLUMN john_ai_agentbox_system_prompt_overrides.content IS 'Prompt override content';
COMMENT ON COLUMN john_ai_agentbox_system_prompt_overrides.created_at IS 'Creation time';
COMMENT ON COLUMN john_ai_agentbox_system_prompt_overrides.updated_at IS 'Update time';

INSERT INTO john_ai_agentbox_users (id, username, password_hash, role, status)
VALUES (
    'usr-admin',
    'admin',
    '$2a$10$VY0pDziBDz3uLcIOCXiCQ.7TZ4YHHhs072uOwDuUwq/KqmRHLf2EK',
    'admin',
    'active'
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO john_ai_agentbox_coding_images (name, image_ref, agent_type, default_shell, notes, enabled, is_default)
VALUES
    ('Claude Code', 'loads/claude:latest', 'claude_code', '/bin/bash', 'Default Claude Code coding image. Must include agent user 1000:1000, HOME /home/agent, passwordless sudo, and writable workspace/shared roots.', TRUE, TRUE),
    ('Codex', 'loads/codex:latest', 'codex', '/bin/bash', 'Default Codex coding image. Must include agent user 1000:1000, HOME /home/agent, passwordless sudo, and writable workspace/shared roots.', TRUE, TRUE),
    ('Coding Base', 'loads/aicoding:base-26.04', 'custom', '/bin/bash', 'Default base coding image. Must include agent user 1000:1000, HOME /home/agent, passwordless sudo, and writable workspace/shared roots.', TRUE, TRUE)
ON CONFLICT (image_ref) DO NOTHING;

INSERT INTO john_ai_agentbox_ai_capability_tiers (code, display_name, description, enabled)
VALUES
    ('basic', '基础', '基础 AI 能力，用于 Git commit message、标题摘要和简单文本生成。', TRUE),
    ('standard', '标准', '标准 AI 能力，用于常规代码生成、代码解释和代码优化。', TRUE),
    ('advanced', '高级', '高级 AI 能力，用于复杂代码生成、多文件推理和架构级优化。', TRUE)
ON CONFLICT (code) DO NOTHING;
