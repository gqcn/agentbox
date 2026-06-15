import type {
  AICapabilityTestResult,
  AICapabilityTierInfo,
  AICapabilityTestPayload,
  AICapabilityTierPayload,
  AIInvocationLogInfo,
  AgentInfo,
  AgentRuntimeServiceInfo,
  AgentServiceBridgeInfo,
  ApiResponse,
  AuthLogoutResponse,
  AuthSessionResponse,
  ChangeImageResponse,
  ChatMessagesResponse,
  ChatInteractionInfo,
  ChatInteractionResponsePayload,
  ChatInteractionStatusPayload,
  ChatRecoverResponse,
  ChatSessionInfo,
  CodingImageInfo,
  ContainerInfo,
  DockerHealth,
  LogsResponse,
  PromptTemplateInfo,
  ProviderInfo,
  ProviderModel,
  SettingInfo,
  GitCommitResponse,
  GitCommitMessageSuggestionResponse,
  GitDiffResponse,
  GitFileResponse,
  GitStatusResponse,
  GitDiffScope,
  WorkspaceCreateEntryPayload,
  WorkspaceFilePreview,
  WorkspaceFileSavePayload,
  WorkspaceFileInfo,
  WorkspacePathSuggestion,
  WorkspaceSkillListResponse,
  WorkspaceSkillUploadResponse,
  WorkspaceTreeNode,
  WorkspaceUploadResponse,
  UpdatePromptTemplatePayload,
} from "./types";
import {
  pluginApiBasePath,
  pluginApiSessionPath,
  pluginApiSessionsPath,
} from "./plugin-paths";

export class ApiError extends Error {
  status: number;
  code: number;
  errorCode?: string;
  messageKey?: string;
  messageParams?: Record<string, unknown>;

  constructor(
    message: string,
    status: number,
    code: number,
    metadata?: {
      errorCode?: string;
      messageKey?: string;
      messageParams?: Record<string, unknown>;
    },
  ) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.errorCode = metadata?.errorCode;
    this.messageKey = metadata?.messageKey;
    this.messageParams = metadata?.messageParams;
  }
}

export const agentBoxSettingNotFoundErrorCode =
  "JOHN_AI_AGENTBOX_SETTING_NOT_FOUND";

let unauthorizedHandler: (() => void) | undefined;

export function setUnauthorizedHandler(handler: (() => void) | undefined) {
  unauthorizedHandler = handler;
}

async function request<T>(
  path: string,
  options?: RequestInit,
  isForm = false,
): Promise<T> {
  const response = await fetch(path, {
    ...options,
    credentials: "include",
    headers: isForm
      ? options?.headers
      : {
          "Content-Type": "application/json",
          ...(options?.headers ?? {}),
        },
  });
  const body = await parseApiResponse<T>(response);
  if (!response.ok || body.code !== 0) {
    const error = new ApiError(
      body.message || `HTTP ${response.status}`,
      response.status,
      body.code,
      {
        errorCode: body.errorCode,
        messageKey: body.messageKey,
        messageParams: body.messageParams,
      },
    );
    if (response.status === 401 && path !== pluginApiSessionsPath) {
      unauthorizedHandler?.();
    }
    throw error;
  }
  return body.data;
}

async function parseApiResponse<T>(
  response: Response,
): Promise<ApiResponse<T>> {
  try {
    return (await response.json()) as ApiResponse<T>;
  } catch {
    return {
      code: response.ok ? 0 : response.status,
      message: response.ok ? "ok" : `HTTP ${response.status}`,
      data: null as T,
    };
  }
}

function query(params: Record<string, string | number | boolean | undefined>) {
  const search = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== "") {
      search.set(key, String(value));
    }
  });
  const value = search.toString();
  return value ? `?${value}` : "";
}

export const api = {
  login: (payload: { username: string; password: string }) =>
    request<AuthSessionResponse>(pluginApiSessionsPath, {
      method: "POST",
      body: JSON.stringify(payload),
    }),
  currentSession: () => request<AuthSessionResponse>(pluginApiSessionPath),
  logout: () =>
    request<AuthLogoutResponse>(pluginApiSessionPath, { method: "DELETE" }),
  dockerHealth: () => request<DockerHealth>(`${pluginApiBasePath}/health/docker`),
  getSetting: (key: string) =>
    request<SettingInfo>(`${pluginApiBasePath}/settings/${encodeURIComponent(key)}`),
  updateSetting: (key: string, value: string) =>
    request<SettingInfo>(`${pluginApiBasePath}/settings/${encodeURIComponent(key)}`, {
      method: "PUT",
      body: JSON.stringify({ value }),
    }),
  listProviders: async () =>
    (await request<ProviderInfo[] | null>(`${pluginApiBasePath}/providers`)) ?? [],
  createProvider: (payload: {
    name: string;
    homepageUrl: string;
    notes: string;
    apiKey: string;
    openaiBaseUrl: string;
    anthropicBaseUrl: string;
  }) =>
    request<ProviderInfo>(`${pluginApiBasePath}/providers`, {
      method: "POST",
      body: JSON.stringify(payload),
    }),
  updateProvider: (
    providerId: number,
    payload: {
      name: string;
      homepageUrl: string;
      notes: string;
      apiKey: string;
      openaiBaseUrl: string;
      anthropicBaseUrl: string;
    },
  ) =>
    request<ProviderInfo>(`${pluginApiBasePath}/providers/${providerId}`, {
      method: "PUT",
      body: JSON.stringify(payload),
    }),
  addProviderModel: (
    providerId: number,
    payload: { name: string; protocol: string },
  ) =>
    request<ProviderModel>(`${pluginApiBasePath}/providers/${providerId}/models`, {
      method: "POST",
      body: JSON.stringify(payload),
    }),
  deleteProviderModel: (providerId: number, modelId: number) =>
    request<{ deleted: boolean }>(
      `${pluginApiBasePath}/providers/${providerId}/models/${modelId}`,
      {
        method: "DELETE",
      },
    ),
  syncProviderModels: (providerId: number, protocol: string) =>
    request<{ count: number; models: ProviderModel[] }>(
      `${pluginApiBasePath}/providers/${providerId}/models/sync`,
      {
        method: "POST",
        body: JSON.stringify({ protocol }),
      },
    ),
  deleteProvider: (providerId: number) =>
    request<{ deleted: boolean }>(`${pluginApiBasePath}/providers/${providerId}`, {
      method: "DELETE",
    }),
  listAICapabilityTiers: async () =>
    (await request<AICapabilityTierInfo[] | null>(
      `${pluginApiBasePath}/ai/capability-tiers`,
    )) ?? [],
  updateAICapabilityTier: (code: string, payload: AICapabilityTierPayload) =>
    request<AICapabilityTierInfo>(
      `${pluginApiBasePath}/ai/capability-tiers/${encodeURIComponent(code)}`,
      {
        method: "PUT",
        body: JSON.stringify(payload),
      },
    ),
  testAICapabilityTier: (code: string, payload?: AICapabilityTestPayload) =>
    request<AICapabilityTestResult>(
      `${pluginApiBasePath}/ai/capability-tiers/${encodeURIComponent(code)}/test`,
      {
        method: "POST",
        body: JSON.stringify(payload ?? {}),
      },
    ),
  listAIInvocations: (
    params: {
      purpose?: string;
      tier?: string;
      status?: string;
      limit?: number;
    } = {},
  ) => request<AIInvocationLogInfo[]>(`${pluginApiBasePath}/ai/invocations${query(params)}`),
  listPromptTemplates: async () =>
    (await request<PromptTemplateInfo[] | null>(`${pluginApiBasePath}/prompt-templates`)) ?? [],
  getPromptTemplate: (code: string) =>
    request<PromptTemplateInfo>(
      `${pluginApiBasePath}/prompt-templates/${encodeURIComponent(code)}`,
    ),
  updatePromptTemplate: (code: string, payload: UpdatePromptTemplatePayload) =>
    request<PromptTemplateInfo>(
      `${pluginApiBasePath}/prompt-templates/${encodeURIComponent(code)}`,
      {
        method: "PUT",
        body: JSON.stringify(payload),
      },
    ),
  restorePromptTemplate: (code: string) =>
    request<PromptTemplateInfo>(
      `${pluginApiBasePath}/prompt-templates/${encodeURIComponent(code)}/restore`,
      {
        method: "POST",
      },
    ),
  listImages: async () =>
    (await request<CodingImageInfo[] | null>(`${pluginApiBasePath}/images`)) ?? [],
  createImage: (payload: {
    name: string;
    imageRef: string;
    agentType: string;
    defaultShell: string;
    notes: string;
    enabled: boolean;
  }) =>
    request<CodingImageInfo>(`${pluginApiBasePath}/images`, {
      method: "POST",
      body: JSON.stringify(payload),
    }),
  updateImage: (
    imageId: number,
    payload: {
      name: string;
      imageRef: string;
      agentType: string;
      defaultShell: string;
      notes: string;
      enabled: boolean;
    },
  ) =>
    request<CodingImageInfo>(`${pluginApiBasePath}/images/${imageId}`, {
      method: "PUT",
      body: JSON.stringify(payload),
    }),
  deleteImage: (imageId: number) =>
    request<{ deleted: boolean }>(`${pluginApiBasePath}/images/${imageId}`, {
      method: "DELETE",
    }),
  listAgents: async () =>
    (await request<AgentInfo[] | null>(`${pluginApiBasePath}/agents`)) ?? [],
  createAgent: (payload: {
    name: string;
    providerId: number;
    modelName: string;
    modelProtocol: string;
    imageId: number;
    agentType: string;
    iconKey: string;
    notes: string;
  }) =>
    request<AgentInfo>(`${pluginApiBasePath}/agents`, {
      method: "POST",
      body: JSON.stringify(payload),
    }),
  updateAgent: (
    id: string,
    payload: {
      name: string;
      providerId: number;
      modelName: string;
      modelProtocol: string;
      agentType: string;
      iconKey: string;
      notes: string;
    },
  ) =>
    request<AgentInfo>(`${pluginApiBasePath}/agents/${encodeURIComponent(id)}`, {
      method: "PUT",
      body: JSON.stringify(payload),
    }),
  startAgent: (id: string) =>
    request<AgentInfo>(`${pluginApiBasePath}/agents/${encodeURIComponent(id)}/start`, {
      method: "POST",
    }),
  stopAgent: (id: string) =>
    request<AgentInfo>(`${pluginApiBasePath}/agents/${encodeURIComponent(id)}/stop`, {
      method: "POST",
    }),
  deleteAgent: (id: string, deleteVolumes: boolean) =>
    request<{ deleted: boolean }>(`${pluginApiBasePath}/agents/${encodeURIComponent(id)}`, {
      method: "DELETE",
      body: JSON.stringify({ deleteVolumes }),
    }),
  changeAgentImage: (id: string, imageId: number) =>
    request<ChangeImageResponse>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/image`,
      {
        method: "PUT",
        body: JSON.stringify({ imageId }),
      },
    ),
  getAgentLogs: (id: string) =>
    request<LogsResponse>(`${pluginApiBasePath}/agents/${encodeURIComponent(id)}/logs`),
  listAgentServices: (id: string) =>
    request<AgentRuntimeServiceInfo[]>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/services`,
    ),
  getAgentService: (id: string, serviceId: string) =>
    request<AgentRuntimeServiceInfo>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/services/${encodeURIComponent(serviceId)}`,
    ),
  listAgentServiceBridges: (id: string) =>
    request<AgentServiceBridgeInfo[]>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/service-bridges`,
    ),
  createAgentServiceBridge: (
    id: string,
    payload: { serviceId: string; listenAddress: string; port: number },
  ) =>
    request<AgentServiceBridgeInfo>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/service-bridges`,
      {
        method: "POST",
        body: JSON.stringify(payload),
      },
    ),
  deleteAgentServiceBridge: (id: string, bridgeId: string) =>
    request<{ deleted: boolean }>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/service-bridges/${encodeURIComponent(bridgeId)}`,
      {
        method: "DELETE",
      },
    ),
  listAgentChatSessions: (id: string) =>
    request<ChatSessionInfo[]>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/chat/sessions`,
    ),
  createAgentChatSession: (id: string) =>
    request<ChatSessionInfo>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/chat/sessions`,
      {
        method: "POST",
      },
    ),
  getAgentChatSession: (id: string, sessionId: string) =>
    request<ChatSessionInfo>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/chat/sessions/${encodeURIComponent(sessionId)}`,
    ),
  updateAgentChatSession: (id: string, sessionId: string, title: string) =>
    request<ChatSessionInfo>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/chat/sessions/${encodeURIComponent(sessionId)}`,
      {
        method: "PUT",
        body: JSON.stringify({ title }),
      },
    ),
  deleteAgentChatSession: (id: string, sessionId: string) =>
    request<{ deleted: boolean }>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/chat/sessions/${encodeURIComponent(sessionId)}`,
      {
        method: "DELETE",
      },
    ),
  getAgentChatMessages: (id: string, sessionId: string) =>
    request<ChatMessagesResponse>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/chat/sessions/${encodeURIComponent(sessionId)}/messages`,
    ),
  recoverAgentChat: (id: string, sessionId: string) =>
    request<ChatRecoverResponse>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/chat/sessions/${encodeURIComponent(sessionId)}/recover`,
      {
        method: "POST",
        body: JSON.stringify({ startNew: true }),
      },
    ),
  listAgentChatInteractions: (
    id: string,
    sessionId: string,
    params: { status?: string; type?: string } = {},
  ) =>
    request<ChatInteractionInfo[]>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/chat/sessions/${encodeURIComponent(sessionId)}/interactions${query(params)}`,
    ),
  getAgentChatInteraction: (
    id: string,
    sessionId: string,
    interactionId: string,
  ) =>
    request<ChatInteractionInfo>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/chat/sessions/${encodeURIComponent(sessionId)}/interactions/${encodeURIComponent(interactionId)}`,
    ),
  updateAgentChatInteractionResponse: (
    id: string,
    sessionId: string,
    interactionId: string,
    payload: ChatInteractionResponsePayload,
  ) =>
    request<ChatInteractionInfo>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/chat/sessions/${encodeURIComponent(sessionId)}/interactions/${encodeURIComponent(interactionId)}/response`,
      {
        method: "PUT",
        body: JSON.stringify(payload),
      },
    ),
  updateAgentChatInteractionStatus: (
    id: string,
    sessionId: string,
    interactionId: string,
    payload: ChatInteractionStatusPayload,
  ) =>
    request<ChatInteractionInfo>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/chat/sessions/${encodeURIComponent(sessionId)}/interactions/${encodeURIComponent(interactionId)}/status`,
      {
        method: "PUT",
        body: JSON.stringify(payload),
      },
    ),
  workspacePathSuggestions: (id: string, value: string) =>
    request<WorkspacePathSuggestion[]>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/workspace/paths${query({ query: value })}`,
    ),
  workspaceTree: (id: string, pathValue: string, includeFiles = false) =>
    request<WorkspaceTreeNode[]>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/workspace/tree${query({ path: pathValue, includeFiles })}`,
    ),
  workspaceFile: (id: string, pathValue: string) =>
    request<WorkspaceFilePreview>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/workspace/file${query({ path: pathValue })}`,
    ),
  saveWorkspaceFile: (id: string, payload: WorkspaceFileSavePayload) =>
    request<WorkspaceFilePreview>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/workspace/file`,
      {
        method: "PUT",
        body: JSON.stringify(payload),
      },
    ),
  createWorkspaceFile: (id: string, payload: WorkspaceCreateEntryPayload) =>
    request<WorkspaceFilePreview>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/workspace/files`,
      {
        method: "POST",
        body: JSON.stringify(payload),
      },
    ),
  createWorkspaceDirectory: (
    id: string,
    payload: WorkspaceCreateEntryPayload,
  ) =>
    request<WorkspaceFileInfo>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/workspace/directories`,
      {
        method: "POST",
        body: JSON.stringify(payload),
      },
    ),
  workspaceDownloadUrl: (id: string, pathValue: string) =>
    `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/workspace/download${query({ path: pathValue })}`,
  workspaceResourceUrl: (
    id: string,
    pathValue: string,
    disposition?: "inline" | "attachment",
  ) =>
    `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/workspace/resources${query({ path: pathValue, disposition })}`,
  workspaceHtmlPreviewUrl: (id: string, pathValue: string) =>
    `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/workspace/html-previews${query({ path: pathValue })}`,
  uploadWorkspaceFiles: (id: string, pathValue: string, files: File[]) => {
    const body = new FormData();
    files.forEach((file) => body.append("file", file));
    return request<WorkspaceUploadResponse>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/workspace/upload${query({ path: pathValue })}`,
      { method: "POST", body },
      true,
    );
  },
  listSkills: (
    id: string,
    scope: string,
    pathValue: string,
    searchQuery: string,
  ) =>
    request<WorkspaceSkillListResponse>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/skills${query({ scope, path: pathValue, query: searchQuery })}`,
    ),
  uploadProjectSkills: (id: string, pathValue: string, files: File[]) => {
    const body = new FormData();
    files.forEach((file) => body.append("file", file));
    return request<WorkspaceSkillUploadResponse>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/skills/upload${query({ scope: "project", path: pathValue })}`,
      { method: "POST", body },
      true,
    );
  },
  gitStatus: (id: string, pathValue: string) =>
    request<GitStatusResponse>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/git/status${query({ path: pathValue })}`,
    ),
  gitFile: (id: string, pathValue: string, file: string) =>
    request<GitFileResponse>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/git/file${query({ path: pathValue, file })}`,
    ),
  gitDiff: (
    id: string,
    pathValue: string,
    file: string,
    scope: GitDiffScope = "unstaged",
  ) =>
    request<GitDiffResponse>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/git/diff${query({ path: pathValue, file, scope })}`,
    ),
  gitCommitMessageSuggestion: (id: string, pathValue: string) =>
    request<GitCommitMessageSuggestionResponse>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/git/commit-message-suggestions`,
      {
        method: "POST",
        body: JSON.stringify({ path: pathValue }),
      },
    ),
  stageGitFiles: (id: string, pathValue: string, files: string[]) =>
    request<GitStatusResponse>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/git/index`,
      {
        method: "PUT",
        body: JSON.stringify({ path: pathValue, files }),
      },
    ),
  stageAllGitFiles: (id: string, pathValue: string) =>
    request<GitStatusResponse>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/git/index`,
      {
        method: "PUT",
        body: JSON.stringify({ path: pathValue, files: [], all: true }),
      },
    ),
  unstageGitFiles: (id: string, pathValue: string, files: string[]) =>
    request<GitStatusResponse>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/git/index`,
      {
        method: "DELETE",
        body: JSON.stringify({ path: pathValue, files }),
      },
    ),
  unstageAllGitFiles: (id: string, pathValue: string) =>
    request<GitStatusResponse>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/git/index`,
      {
        method: "DELETE",
        body: JSON.stringify({ path: pathValue, files: [], all: true }),
      },
    ),
  discardGitFiles: (id: string, pathValue: string, files: string[]) =>
    request<GitStatusResponse>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/git/changes`,
      {
        method: "DELETE",
        body: JSON.stringify({ path: pathValue, files }),
      },
    ),
  gitCommit: (id: string, pathValue: string, message: string, push = false) =>
    request<GitCommitResponse>(
      `${pluginApiBasePath}/agents/${encodeURIComponent(id)}/git/commits`,
      {
        method: "POST",
        body: JSON.stringify({ path: pathValue, message, push }),
      },
    ),
  listContainers: () => request<ContainerInfo[]>(`${pluginApiBasePath}/containers`),
  createContainer: (name: string) =>
    request<ContainerInfo>(`${pluginApiBasePath}/containers`, {
      method: "POST",
      body: JSON.stringify({ name }),
    }),
  startContainer: (id: string) =>
    request<ContainerInfo>(`${pluginApiBasePath}/containers/${encodeURIComponent(id)}/start`, {
      method: "POST",
    }),
  stopContainer: (id: string) =>
    request<ContainerInfo>(`${pluginApiBasePath}/containers/${encodeURIComponent(id)}/stop`, {
      method: "POST",
    }),
  deleteContainer: (id: string) =>
    request<{ deleted: boolean }>(`${pluginApiBasePath}/containers/${encodeURIComponent(id)}`, {
      method: "DELETE",
    }),
  getLogs: (id: string) =>
    request<LogsResponse>(`${pluginApiBasePath}/containers/${encodeURIComponent(id)}/logs`),
};
