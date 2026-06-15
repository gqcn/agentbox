export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
  errorCode?: string;
  messageKey?: string;
  messageParams?: Record<string, unknown>;
}

export interface UserInfo {
  id: string;
  username: string;
  role: "admin" | "user" | string;
  status: "active" | "disabled" | string;
  lastLoginAt?: number;
  createdAt: number;
  updatedAt: number;
}

export interface AuthSessionResponse {
  user: UserInfo;
}

export interface AuthLogoutResponse {
  loggedOut: boolean;
}

export interface MountInfo {
  type: string;
  source: string;
  destination: string;
  name?: string;
}

export interface ContainerInfo {
  id: string;
  name: string;
  dockerId: string;
  image: string;
  state: string;
  status: string;
  createdAt: number;
  mounts?: MountInfo[];
  labels?: Record<string, string>;
  workspace?: string;
}

export interface ProviderModel {
  id: number;
  providerId: number;
  name: string;
  protocol: "openai" | "anthropic";
  source: "manual" | "api";
  lastSyncedAt?: number;
  createdAt: number;
  updatedAt: number;
}

export interface ProviderInfo {
  id: number;
  name: string;
  homepageUrl: string;
  notes: string;
  apiKeyMasked: string;
  apiKeyConfigured: boolean;
  openaiBaseUrl: string;
  anthropicBaseUrl: string;
  createdAt: number;
  updatedAt: number;
  models?: ProviderModel[];
}

export type AICapabilityTierCode = "basic" | "standard" | "advanced";
export type AIInvocationStatus = "success" | "error";

export interface AICapabilityBindingInfo {
  id: number;
  tierCode: AICapabilityTierCode;
  providerId: number;
  providerName: string;
  providerModelId: number;
  modelName: string;
  protocol: "openai" | "anthropic";
  priority: number;
  enabled: boolean;
  createdAt: number;
  updatedAt: number;
}

export interface AIInvocationLogInfo {
  id: number;
  purpose: string;
  tierCode: AICapabilityTierCode;
  providerId?: number;
  providerName?: string;
  providerModelId?: number;
  modelName?: string;
  protocol?: "openai" | "anthropic";
  status: AIInvocationStatus;
  latencyMs: number;
  errorMessage?: string;
  createdAt: number;
}

export interface AICapabilityTierInfo {
  code: AICapabilityTierCode;
  displayName: string;
  description: string;
  enabled: boolean;
  configured: boolean;
  available: boolean;
  binding?: AICapabilityBindingInfo;
  lastTest?: AIInvocationLogInfo;
  createdAt: number;
  updatedAt: number;
}

export interface AICapabilityTierPayload {
  enabled: boolean;
  providerId: number;
  providerModelId: number;
  protocol?: "openai" | "anthropic";
}

export interface AICapabilityTestPayload {
  providerId: number;
  providerModelId: number;
  protocol?: "openai" | "anthropic";
}

export interface AICapabilityTestResult {
  status: AIInvocationStatus;
  tierCode: AICapabilityTierCode;
  providerId?: number;
  providerName?: string;
  providerModelId?: number;
  modelName?: string;
  protocol?: "openai" | "anthropic";
  latencyMs: number;
  errorMessage?: string;
  testedAt: number;
}

export type SystemPromptCode = "git_commit_message" | string;

export interface PromptTemplateVariableInfo {
  name: string;
  description: string;
  required: boolean;
  sampleValue: string;
}

export interface PromptTemplateInfo {
  code: SystemPromptCode;
  displayName: string;
  description: string;
  purpose: string;
  tierCode: AICapabilityTierCode;
  defaultContent: string;
  content: string;
  variables: PromptTemplateVariableInfo[];
  createdAt: number;
  updatedAt: number;
}

export interface UpdatePromptTemplatePayload {
  content: string;
}

export interface CodingImageInfo {
  id: number;
  name: string;
  imageRef: string;
  agentType: "claude_code" | "codex" | "custom";
  defaultShell: string;
  notes: string;
  enabled: boolean;
  isDefault: boolean;
  createdAt: number;
  updatedAt: number;
}

export interface AgentInfo {
  id: string;
  name: string;
  providerId: number;
  providerName: string;
  modelName: string;
  modelProtocol: "openai" | "anthropic";
  imageId: number;
  imageName: string;
  imageRef: string;
  agentType: "claude_code" | "codex" | "custom";
  iconKey: string;
  notes: string;
  runtimeStatus: string;
  activityStatus: string;
  containerId?: string;
  dockerId?: string;
  deletedAt?: number;
  createdAt: number;
  updatedAt: number;
}

export type AgentServiceProtocol = "http" | "https" | "tcp" | "unknown";
export type AgentServiceAccessStatus =
  | "direct"
  | "bridge_required"
  | "bridged"
  | "unavailable";
export type AgentServiceBridgeStatus = "active" | "closed" | "error";

export interface AgentServiceListenAddress {
  address: string;
  port: number;
  network: string;
  accessStatus: AgentServiceAccessStatus;
  bridgeId?: string;
  localHost?: string;
  localPort?: number;
  unavailableReason?: string;
}

export interface AgentRuntimeServiceInfo {
  id: string;
  agentId: string;
  port: number;
  protocol: AgentServiceProtocol;
  accessStatus: AgentServiceAccessStatus;
  listenAddresses: AgentServiceListenAddress[];
  processName?: string;
  pid?: string;
  proxyUrl?: string;
  tunnelUrl?: string;
  tunnelCommand?: string;
  bridgeId?: string;
  localHost?: string;
  localPort?: number;
  unavailableReason?: string;
  lastCheckedAt: number;
}

export interface AgentServiceBridgeInfo {
  id: string;
  agentId: string;
  serviceId: string;
  listenAddress: string;
  port: number;
  bridgePort: number;
  localHost?: string;
  localPort?: number;
  status: AgentServiceBridgeStatus;
  errorMessage?: string;
  createdAt: number;
  closedAt?: number;
}

export interface ChangeImageResponse {
  agent: AgentInfo;
  lostPaths: string[];
  preservedPaths: string[];
}

export interface LogsResponse {
  logs: string;
}

export interface DockerHealth {
  ok: boolean;
  apiVersion?: string;
  osType?: string;
  error?: string;
}

export interface SettingInfo {
  key: string;
  value: string;
  createdAt: number;
  updatedAt: number;
}

export type ShellStatus =
  | "connecting"
  | "recovering"
  | "connected"
  | "rebuilding"
  | "closed"
  | "error"
  | "detached"
  | string;

export interface ShellMessage {
  type:
    | "input"
    | "resize"
    | "output"
    | "replay"
    | "status"
    | "error"
    | "close"
    | "rebuild";
  data?: string;
  status?: ShellStatus;
  message?: string;
  replay?: boolean;
  cols?: number;
  rows?: number;
}

export type ChatSessionStatus =
  | "idle"
  | "running"
  | "waiting_input"
  | "exited"
  | "recovering"
  | "error";
export type ChatRuntimeState =
  | "idle"
  | "running"
  | "waiting_input"
  | "exited"
  | "recovering"
  | "error";
export type ChatMessageRole =
  | "user"
  | "assistant"
  | "system"
  | "error"
  | "terminal";
export type ChatMessageStatus = "streaming" | "complete" | "error";

export interface ChatSessionInfo {
  id: string;
  agentId: string;
  title: string;
  status: ChatSessionStatus;
  toolType: string;
  toolSessionId?: string;
  runtimeState: ChatRuntimeState;
  lastError?: string;
  messageCount: number;
  lastMessagePreview: string;
  createdAt: number;
  updatedAt: number;
  lastActiveAt: number;
}

export interface ChatMessageInfo {
  id: number;
  sessionId: string;
  sequence: number;
  role: ChatMessageRole;
  content: string;
  status: ChatMessageStatus;
  metadata?: string;
  createdAt: number;
  updatedAt: number;
}

export type ChatExecutionEventKind =
  | "status"
  | "tool"
  | "command"
  | "file"
  | "test"
  | "thinking";
export type ChatExecutionEventStatus = "running" | "complete" | "error";

export interface ChatExecutionEvent {
  id: string;
  kind: ChatExecutionEventKind;
  title: string;
  detail?: string;
  status: ChatExecutionEventStatus;
  createdAt: string;
}

export interface ChatMessagesResponse {
  session: ChatSessionInfo;
  messages: ChatMessageInfo[];
}

export type ChatInteractionType =
  | "permission"
  | "question"
  | "choice"
  | "text"
  | "auth"
  | "plan"
  | "custom";
export type ChatInteractionStatus =
  | "pending"
  | "resolved"
  | "rejected"
  | "cancelled"
  | "expired"
  | "error";
export type ChatInteractionRiskLevel = "low" | "medium" | "high" | "critical";
export type ChatInteractionResponseScope =
  | "once"
  | "session"
  | "agent"
  | "provider"
  | "";
export type ChatInteractionResponseMode =
  | "allow"
  | "answer"
  | "reject"
  | "cancel"
  | "allow_once"
  | "allow_session"
  | string;

export interface ChatInteractionInfo {
  id: string;
  agentId: string;
  sessionId: string;
  assistantMessageId?: number;
  toolType: string;
  toolInteractionId?: string;
  type: ChatInteractionType;
  status: ChatInteractionStatus;
  title: string;
  body: string;
  riskLevel: ChatInteractionRiskLevel;
  payload: string;
  response: string;
  responseMode?: string;
  responseScope?: ChatInteractionResponseScope;
  expiresAt?: number;
  resolvedAt?: number;
  createdAt: number;
  updatedAt: number;
}

export interface ChatInteractionResponsePayload {
  response: string;
  responseMode: ChatInteractionResponseMode;
  responseScope?: ChatInteractionResponseScope;
}

export interface ChatInteractionStatusPayload {
  status: ChatInteractionStatus;
}

export interface ChatRecoverResponse {
  session: ChatSessionInfo;
  message?: ChatMessageInfo;
}

export interface ChatEvent {
  type:
    | "status"
    | "user_message"
    | "assistant_delta"
    | "execution_event"
    | "message"
    | "message_complete"
    | "resume_required"
    | "recoverable_error"
    | "terminal_output"
    | "interaction_requested"
    | "interaction_updated"
    | "interaction_resolved"
    | "interaction_cancelled"
    | "error";
  session?: ChatSessionInfo;
  message?: ChatMessageInfo;
  interaction?: ChatInteractionInfo;
  messageId?: number;
  content?: string;
  event?: ChatExecutionEvent;
  terminalId?: string;
  status?: string;
  error?: string;
  notice?: string;
}

export interface WorkspacePathSuggestion {
  name: string;
  path: string;
}

export type WorkspaceNodeType = "directory" | "file";
export type WorkspacePreviewType = "text" | "image" | "unsupported";

export interface WorkspaceTreeNode {
  name: string;
  path: string;
  type: WorkspaceNodeType;
  size?: number;
  modifiedAt?: number;
  expandable: boolean;
  children?: WorkspaceTreeNode[];
}

export interface WorkspaceFileInfo {
  name: string;
  path: string;
  type: WorkspaceNodeType;
  size: number;
  modifiedAt?: number;
  contentType?: string;
}

export interface WorkspaceFilePreview {
  file: WorkspaceFileInfo;
  previewType: WorkspacePreviewType;
  content?: string;
  encoding?: string;
  contentHash?: string;
  tooLarge: boolean;
  downloadUrl?: string;
}

export interface WorkspaceFileSavePayload {
  path: string;
  content: string;
  encoding?: string;
  baseHash?: string;
}

export interface WorkspaceCreateEntryPayload {
  parentPath: string;
  name: string;
}

export interface WorkspaceUploadResponse {
  files: WorkspaceFileInfo[];
}

export type WorkspaceSkillScope = "global" | "project";

export interface WorkspaceSkillInfo {
  name: string;
  description?: string;
  scope: WorkspaceSkillScope;
  path: string;
  source: string;
  hasManifest: boolean;
}

export interface WorkspaceSkillListResponse {
  scope: WorkspaceSkillScope;
  path?: string;
  items: WorkspaceSkillInfo[];
}

export interface WorkspaceSkillUploadResponse {
  skills: WorkspaceSkillInfo[];
}

export type GitRepositoryState = "ok" | "clean" | "not_repo" | "error";

export interface GitChange {
  path: string;
  oldPath?: string;
  status: string;
  indexState?: string;
  workState?: string;
  changeScope?: GitChangeScope;
}

export interface GitTreeNode {
  name: string;
  path: string;
  oldPath?: string;
  type: WorkspaceNodeType;
  status?: string;
  changeScope?: GitChangeScope;
  children?: GitTreeNode[];
}

export interface GitStatusResponse {
  state: GitRepositoryState;
  path: string;
  root?: string;
  message?: string;
  changes?: GitChange[];
  stagedChanges?: GitChange[];
  changeTree?: GitTreeNode[];
  stagedTree?: GitTreeNode[];
}

export interface GitFileResponse {
  file: WorkspaceFilePreview;
  status?: string;
  message?: string;
}

export interface GitDiffResponse {
  path: string;
  status?: string;
  scope?: GitDiffScope;
  diff: string;
  originalContent: string;
  modifiedContent: string;
  originalPath: string;
  modifiedPath: string;
  language: string;
  message?: string;
}

export interface GitCommitMessageSuggestionResponse {
  message: string;
  diffScope: GitDiffScope;
  tierCode: AICapabilityTierCode;
  providerId: number;
  providerName: string;
  providerModelId: number;
  modelName: string;
  protocol: "openai" | "anthropic";
  truncated: boolean;
  generatedAt: number;
}

export interface GitCommitResponse {
  commitHash: string;
  pushed: boolean;
  status: GitStatusResponse;
}

export type GitChangeScope = "staged" | "unstaged";
export type GitDiffScope = "staged" | "unstaged";
