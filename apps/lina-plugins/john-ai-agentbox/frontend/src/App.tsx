import {
  lazy,
  Suspense,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type {
  ButtonHTMLAttributes,
  ComponentType,
  CSSProperties,
  Dispatch,
  FormEvent,
  ReactNode,
  SetStateAction,
} from "react";
import { toast } from "sonner";
import {
  Activity,
  BrainCircuit,
  Boxes,
  Check,
  FileText,
  Fullscreen,
  Image,
  Info,
  KeyRound,
  LogOut,
  MessageSquare,
  Moon,
  MoreHorizontal,
  Network,
  PanelLeftClose as PanelLeftCloseIcon,
  PanelLeftOpen as PanelLeftOpenIcon,
  Pencil,
  Play,
  Plus,
  RefreshCw,
  Repeat2,
  ScrollText,
  Settings,
  Shrink,
  Square,
  Sun,
  Terminal,
  Trash2,
  UserCircle,
  X,
} from "lucide-react";
import {
  ApiError,
  agentBoxSettingNotFoundErrorCode,
  api,
  setUnauthorizedHandler,
} from "./api";
import AICapabilitiesPage from "./AICapabilitiesPage";
import ChatPanel from "./ChatPanel";
import FilesPanel from "./FilesPanel";
import type { FilesLocateRequest } from "./FilesPanel";
import GitPanel, { hasGitChanges } from "./GitPanel";
import type { GitLocateRequest } from "./GitPanel";
const ShellPanel = lazy(() => import("./ShellPanel"));
import PromptsPage from "./PromptsPage";
import ServicesPanel from "./ServicesPanel";
import SkillsPanel from "./SkillsPanel";
import WorkbenchSettingsDialog from "./WorkbenchSettingsDialog";
import WorkspacePathPicker from "./WorkspacePathPicker";
import type {
  AgentInfo,
  CodingImageInfo,
  ProviderInfo,
  ProviderModel,
  UserInfo,
} from "./types";
import type { ColumnDef } from "@/components/ui";
import {
  agentDetailPanelIds,
  defaultWorkbenchSettings,
  decodeWorkbenchSettings,
  encodeWorkbenchSettings,
  loadLocalWorkbenchSettings,
  loadWorkbenchSettings,
  saveWorkbenchSettings,
  workbenchSettingsServerKey,
} from "@/lib/workbench-settings";
import type {
  AgentDetailPanelId,
  WorkbenchSettings,
} from "@/lib/workbench-settings";
import {
  Badge,
  Button,
  CheckboxField,
  ConfirmDialog,
  ContextMenu,
  ContextMenuContent,
  ContextMenuGroup,
  ContextMenuItem,
  ContextMenuLabel,
  ContextMenuSeparator,
  ContextMenuTrigger,
  DataTable,
  Dialog,
  Dropdown,
  DropdownItem,
  Field,
  Form,
  IconButton,
  Input,
  ListButton,
  Pagination,
  SearchField,
  Select,
  SelectOption,
  TabButton,
  Textarea,
  Toaster,
  Alert,
  Spinner,
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui";
import {
  agentIconCategories,
  agentIconOptions,
  recommendedAgentIconOptions,
  resolveAgentIcon,
  resolveAgentIconKey,
} from "@/lib/agent-icons";
import type { AgentIconCategoryId, AgentIconOption } from "@/lib/agent-icons";
import {
  cn,
  normalizeWorkspacePath,
  sharedRootPath,
  workspaceRootPath,
} from "@/lib/utils";
import { useMediaQuery } from "@/hooks/useMediaQuery";

type PanelMode = AgentDetailPanelId;
type ViewMode =
  | "agents"
  | "agentDetail"
  | "providers"
  | "aiCapabilities"
  | "prompts"
  | "images";
type ThemeMode = "light" | "dark";
type AuthStatus = "checking" | "authenticated" | "anonymous";
type EntityMode = "create" | "edit";
type DeleteKind = "agent" | "provider" | "image" | "providerModel";
type AgentUiStatus =
  | "idle"
  | "starting"
  | "running"
  | "working"
  | "waiting_input"
  | "stopping"
  | "stopped"
  | "deleting"
  | "deleted"
  | string;
type AgentUiStatusMap = Record<string, AgentUiStatus>;
type GitDirtyStateMap = Record<string, boolean>;
type WorkspaceLocateIntent =
  | { kind: "files-to-git"; request: GitLocateRequest }
  | { kind: "git-to-files"; request: FilesLocateRequest };

type AgentWorkbenchState = {
  panelMode: PanelMode;
  workspacePath: string;
  activeChatSessionId?: string;
  chatHistoryOpen?: boolean;
};

type PersistedWorkbenchState = {
  viewMode: ViewMode;
  selectedAgentId: string;
  agentWorkbenchState: Record<string, AgentWorkbenchState>;
};

type ProviderForm = {
  id: number;
  name: string;
  homepageUrl: string;
  notes: string;
  apiKey: string;
  openaiBaseUrl: string;
  anthropicBaseUrl: string;
};

type ModelForm = {
  providerId: number;
  name: string;
  protocol: "anthropic" | "openai";
};

type ModelProtocol = ModelForm["protocol"];

type ImageForm = {
  id: number;
  name: string;
  imageRef: string;
  agentType: "claude_code" | "codex" | "custom";
  defaultShell: string;
  notes: string;
  enabled: boolean;
};

type AgentForm = {
  id: string;
  name: string;
  providerId: number;
  modelName: string;
  modelProtocol: "anthropic" | "openai";
  imageId: number;
  agentType: "claude_code" | "codex" | "custom";
  iconKey: string;
  notes: string;
};

type DeleteDialog = {
  open: boolean;
  kind: DeleteKind;
  title: string;
  description: string;
  targetName: string;
  deleteVolumes: boolean;
  agentId: string;
  providerId: number;
  imageId: number;
  modelProviderId: number;
  modelId: number;
};

type PagerState = {
  query: string;
  page: number;
  pageSize: number;
};

const pageSizeOptions = [5, 10, 20];
const defaultWorkspacePath = workspaceRootPath;
const sharedWorkspacePath = sharedRootPath;
const defaultWorkbenchState: AgentWorkbenchState = {
  panelMode: "chat",
  workspacePath: defaultWorkspacePath,
  activeChatSessionId: "",
  chatHistoryOpen: false,
};
const persistedWorkbenchStorageKey = "john-ai-agentbox-workbench";
const panelModes: readonly PanelMode[] = agentDetailPanelIds;
const viewModes: readonly ViewMode[] = [
  "agents",
  "agentDetail",
  "providers",
  "aiCapabilities",
  "prompts",
  "images",
];
const loginPath = "login";
const loginBrowserPath = "/login";
const portalRootPath = "";

function resolvePortalPath() {
  return window.location.pathname === loginBrowserPath ||
    window.location.hash === "#/login"
    ? loginPath
    : portalRootPath;
}

const agentDetailPanelDefinitions: Record<
  PanelMode,
  { label: string; Icon: ComponentType<{ className?: string }> }
> = {
  chat: { label: "对话", Icon: MessageSquare },
  shell: { label: "终端", Icon: Terminal },
  services: { label: "服务", Icon: Network },
  skills: { label: "技能", Icon: Boxes },
  files: { label: "文件", Icon: FileText },
  git: { label: "变更", Icon: SourceControlCodicon },
};

export default function App() {
  const [currentPath, setCurrentPath] = useState(resolvePortalPath);
  const [persistedWorkbench] = useState<PersistedWorkbenchState>(
    loadPersistedWorkbenchState,
  );
  const [providers, setProviders] = useState<ProviderInfo[]>([]);
  const [images, setImages] = useState<CodingImageInfo[]>([]);
  const [agents, setAgents] = useState<AgentInfo[]>([]);
  const [agentsLoaded, setAgentsLoaded] = useState(false);
  const [selectedAgentId, setSelectedAgentId] = useState(
    persistedWorkbench.selectedAgentId,
  );
  const [agentWorkbenchState, setAgentWorkbenchState] = useState<
    Record<string, AgentWorkbenchState>
  >(persistedWorkbench.agentWorkbenchState);
  const isLoginPath = currentPath === loginPath;
  const [agentUiStatuses, setAgentUiStatuses] = useState<AgentUiStatusMap>({});
  const [gitDirtyStates, setGitDirtyStates] = useState<GitDirtyStateMap>({});
  const [workspaceLocateIntent, setWorkspaceLocateIntent] =
    useState<WorkspaceLocateIntent | null>(null);
  const [viewMode, setViewMode] = useState<ViewMode>(
    persistedWorkbench.viewMode,
  );
  const [loading, setLoading] = useState(false);
  const [themeMode, setThemeMode] = useState<ThemeMode>("light");
  const [authStatus, setAuthStatus] = useState<AuthStatus>("checking");
  const [currentUser, setCurrentUser] = useState<UserInfo | null>(null);
  const [workbenchSettings, setWorkbenchSettings] = useState<WorkbenchSettings>(
    loadWorkbenchSettings,
  );
  const [workbenchSettingsHydrated, setWorkbenchSettingsHydrated] =
    useState(false);
  const [workbenchSettingsOpen, setWorkbenchSettingsOpen] = useState(false);
  const loadedWorkbenchSettingsRef = useRef<string | null>(null);
  const [agentPager, setAgentPager] = useState<PagerState>({
    query: "",
    page: 1,
    pageSize: 10,
  });
  const [providerPager, setProviderPager] = useState<PagerState>({
    query: "",
    page: 1,
    pageSize: 10,
  });
  const [imagePager, setImagePager] = useState<PagerState>({
    query: "",
    page: 1,
    pageSize: 10,
  });
  const [providerDialog, setProviderDialog] = useState({
    open: false,
    mode: "create" as EntityMode,
  });
  const [modelDialogOpen, setModelDialogOpen] = useState(false);
  const [imageDialog, setImageDialog] = useState({
    open: false,
    mode: "create" as EntityMode,
  });
  const [agentDialog, setAgentDialog] = useState({
    open: false,
    mode: "create" as EntityMode,
  });
  const [changeImageDialog, setChangeImageDialog] = useState({
    open: false,
    imageId: 0,
  });
  const [stopDialogOpen, setStopDialogOpen] = useState(false);
  const [stopTarget, setStopTarget] = useState<AgentInfo | null>(null);
  const [deleteDialog, setDeleteDialog] =
    useState<DeleteDialog>(blankDeleteDialog());
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [agentDetailOpen, setAgentDetailOpen] = useState(false);
  const [agentDetailFullscreen, setAgentDetailFullscreen] = useState(false);
  const agentDetailRef = useRef<HTMLElement | null>(null);
  const [providerForm, setProviderForm] =
    useState<ProviderForm>(blankProviderForm());
  const [modelForm, setModelForm] = useState<ModelForm>(blankModelForm());
  const [imageForm, setImageForm] = useState<ImageForm>(blankImageForm());
  const [agentForm, setAgentForm] = useState<AgentForm>(blankAgentForm());

  const selectedAgent = useMemo(() => {
    const agent =
      agents.find((item) => item.id === selectedAgentId) ?? agents[0];
    return agent ? withAgentUiStatus(agent, agentUiStatuses) : undefined;
  }, [agents, selectedAgentId, agentUiStatuses]);
  const displayAgents = useMemo(
    () => agents.map((item) => withAgentUiStatus(item, agentUiStatuses)),
    [agents, agentUiStatuses],
  );
  const selectedAgentWorkbench = selectedAgent
    ? (agentWorkbenchState[selectedAgent.id] ?? defaultWorkbenchState)
    : defaultWorkbenchState;
  const visiblePanelModes = useMemo(
    () =>
      workbenchSettings.agentDetailPanels
        .filter((item) => item.visible)
        .map((item) => item.id),
    [workbenchSettings.agentDetailPanels],
  );
  const requestedPanelMode = selectedAgentWorkbench.panelMode;
  const panelMode = visiblePanelModes.includes(requestedPanelMode)
    ? requestedPanelMode
    : (visiblePanelModes[0] ?? defaultWorkbenchState.panelMode);
  const workspacePath = selectedAgentWorkbench.workspacePath;
  const gitPanelVisible = visiblePanelModes.includes("git");
  const filesPanelVisible = visiblePanelModes.includes("files");
  const selectedProviderForForm = providers.find(
    (item) => item.id === Number(agentForm.providerId),
  );
  const providerModelsForForm = selectedProviderForForm?.models ?? [];
  const selectedGitDirty = Boolean(
    gitPanelVisible &&
    selectedAgent?.id &&
    selectedAgent.runtimeStatus === "running" &&
    workspacePath &&
    gitDirtyStates[gitDirtyStateKey(selectedAgent.id, workspacePath)],
  );

  const updateGitDirtyState = useCallback(
    (state: {
      agentId: string;
      workspacePath: string;
      hasChanges: boolean;
    }) => {
      setGitDirtyStates((current) => {
        const key = gitDirtyStateKey(state.agentId, state.workspacePath);
        if (!key) {
          return current;
        }
        if (state.hasChanges) {
          return current[key] ? current : { ...current, [key]: true };
        }
        if (!(key in current)) {
          return current;
        }
        const next = { ...current };
        delete next[key];
        return next;
      });
    },
    [],
  );

  const filteredAgents = useMemo(
    () =>
      filterItems(displayAgents, agentPager.query, (item) => [
        item.name,
        item.providerName,
        item.modelName,
        item.imageName,
        item.imageRef,
        item.agentType,
        item.runtimeStatus,
        item.activityStatus,
        item.notes,
      ]),
    [displayAgents, agentPager.query],
  );
  const filteredProviders = useMemo(
    () =>
      filterItems(providers, providerPager.query, (item) => [
        item.name,
        item.homepageUrl,
        item.openaiBaseUrl,
        item.anthropicBaseUrl,
        item.notes,
        ...(item.models ?? []).map((model) => model.name),
      ]),
    [providers, providerPager.query],
  );
  const filteredImages = useMemo(
    () =>
      filterItems(images, imagePager.query, (item) => [
        item.name,
        item.imageRef,
        item.agentType,
        item.defaultShell,
        item.notes,
      ]),
    [images, imagePager.query],
  );

  const pagedAgents = paginate(filteredAgents, agentPager);
  const pagedProviders = paginate(filteredProviders, providerPager);
  const pagedImages = paginate(filteredImages, imagePager);
  const agentPageInfo = pageInfo(filteredAgents.length, agentPager);
  const providerPageInfo = pageInfo(filteredProviders.length, providerPager);
  const imagePageInfo = pageInfo(filteredImages.length, imagePager);

  useEffect(() => {
    function syncCurrentPath() {
      setCurrentPath(resolvePortalPath());
    }
    window.addEventListener("popstate", syncCurrentPath);
    window.addEventListener("hashchange", syncCurrentPath);
    return () => {
      window.removeEventListener("popstate", syncCurrentPath);
      window.removeEventListener("hashchange", syncCurrentPath);
    };
  }, []);

  useEffect(() => {
    const saved = localStorage.getItem("john-ai-agentbox-theme");
    if (saved === "light" || saved === "dark") {
      setThemeMode(saved);
    }
  }, []);

  useEffect(() => {
    let cancelled = false;
    api
      .currentSession()
      .then((session) => {
        if (cancelled) {
          return;
        }
        setCurrentUser(session.user);
        setAuthStatus("authenticated");
      })
      .catch(() => {
        if (cancelled) {
          return;
        }
        setCurrentUser(null);
        setAuthStatus("anonymous");
      });
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    setUnauthorizedHandler(() => handleSessionExpired());
    return () => setUnauthorizedHandler(undefined);
  }, []);

  useEffect(() => {
    if (authStatus !== "authenticated") {
      return;
    }
    void refreshAll();
    const timer = window.setInterval(() => {
      setAgents((current) => {
        if (current.some((item) => item.runtimeStatus === "running")) {
          void refreshAgents();
        }
        return current;
      });
    }, 5000);
    return () => window.clearInterval(timer);
  }, [authStatus]);

  useEffect(() => {
    if (authStatus !== "authenticated") {
      return;
    }
    void hydrateWorkbenchSettings();
  }, [authStatus]);

  useEffect(() => {
    if (!agentsLoaded) {
      return;
    }
    if (!selectedAgentId && agents.length > 0) {
      setSelectedAgentId(agents[0].id);
    } else if (
      selectedAgentId &&
      !agents.some((item) => item.id === selectedAgentId)
    ) {
      const nextAgentId = agents[0]?.id ?? "";
      setSelectedAgentId(nextAgentId);
      if (!nextAgentId && viewMode === "agentDetail") {
        setViewMode("agents");
      }
    } else if (agents.length === 0 && viewMode === "agentDetail") {
      setViewMode("agents");
    }
  }, [agents, agentsLoaded, selectedAgentId, viewMode]);

  useEffect(() => {
    if (viewMode !== "agentDetail") {
      if (document.fullscreenElement === agentDetailRef.current) {
        void exitDocumentFullscreen().catch(() => undefined);
      }
      setAgentDetailFullscreen(false);
    }
  }, [viewMode]);

  useEffect(() => {
    function syncAgentDetailFullscreen() {
      setAgentDetailFullscreen(
        document.fullscreenElement === agentDetailRef.current,
      );
    }
    document.addEventListener("fullscreenchange", syncAgentDetailFullscreen);
    return () =>
      document.removeEventListener(
        "fullscreenchange",
        syncAgentDetailFullscreen,
      );
  }, []);

  useEffect(() => {
    savePersistedWorkbenchState({
      viewMode,
      selectedAgentId,
      agentWorkbenchState,
    });
  }, [viewMode, selectedAgentId, agentWorkbenchState]);

  useEffect(() => {
    if (!selectedAgent?.id || requestedPanelMode === panelMode) {
      return;
    }
    setAgentWorkbenchState((current) => ({
      ...current,
      [selectedAgent.id]: {
        ...defaultWorkbenchState,
        ...current[selectedAgent.id],
        panelMode,
      },
    }));
  }, [selectedAgent?.id, requestedPanelMode, panelMode]);

  useEffect(() => {
    if (authStatus !== "authenticated" || !workbenchSettingsHydrated) {
      return;
    }
    const encoded = encodeWorkbenchSettings(workbenchSettings);
    saveWorkbenchSettings(workbenchSettings);
    if (loadedWorkbenchSettingsRef.current === encoded) {
      return;
    }
    api
      .updateSetting(workbenchSettingsServerKey, encoded)
      .then(() => {
        loadedWorkbenchSettingsRef.current = encoded;
      })
      .catch((error: Error) => {
        toast.error(`同步工作台设置失败：${error.message}`);
      });
  }, [authStatus, workbenchSettings, workbenchSettingsHydrated]);

  useEffect(() => {
    if (
      !selectedAgent?.id ||
      !workspacePath ||
      selectedAgent.runtimeStatus === "running"
    ) {
      return;
    }
    updateGitDirtyState({
      agentId: selectedAgent.id,
      workspacePath,
      hasChanges: false,
    });
  }, [selectedAgent?.id, selectedAgent?.runtimeStatus, workspacePath]);

  useEffect(() => {
    if (
      viewMode !== "agentDetail" ||
      panelMode === "git" ||
      !gitPanelVisible ||
      !selectedAgent?.id ||
      selectedAgent.runtimeStatus !== "running" ||
      !workspacePath
    ) {
      return;
    }
    let cancelled = false;

    async function refreshGitDirtyBadge() {
      if (!selectedAgent?.id) {
        return;
      }
      try {
        const status = await api.gitStatus(selectedAgent.id, workspacePath);
        if (!cancelled) {
          updateGitDirtyState({
            agentId: selectedAgent.id,
            workspacePath,
            hasChanges: hasGitChanges(status),
          });
        }
      } catch {
        if (!cancelled) {
          updateGitDirtyState({
            agentId: selectedAgent.id,
            workspacePath,
            hasChanges: false,
          });
        }
      }
    }

    void refreshGitDirtyBadge();
    const timer = window.setInterval(() => {
      void refreshGitDirtyBadge();
    }, 15000);

    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, [
    viewMode,
    panelMode,
    gitPanelVisible,
    selectedAgent?.id,
    selectedAgent?.runtimeStatus,
    workspacePath,
  ]);

  useEffect(() => {
    setAgentPager((current) => ({ ...current, page: 1 }));
  }, [agentPager.query]);
  useEffect(() => {
    setProviderPager((current) => ({ ...current, page: 1 }));
  }, [providerPager.query]);
  useEffect(() => {
    setImagePager((current) => ({ ...current, page: 1 }));
  }, [imagePager.query]);

  async function refreshAll() {
    if (authStatus !== "authenticated") {
      return;
    }
    setLoading(true);
    try {
      const [providerItems, imageItems, agentItems] = await Promise.all([
        api.listProviders(),
        api.listImages(),
        api.listAgents(),
      ]);
      setProviders(providerItems);
      setImages(imageItems);
      setAgents(agentItems);
      reconcileAgentUiStatuses(agentItems);
      setAgentsLoaded(true);
    } catch (error) {
      toast.error(actionErrorMessage(error));
    } finally {
      setLoading(false);
    }
  }

  async function refreshAgents() {
    if (authStatus !== "authenticated") {
      return;
    }
    try {
      const agentItems = await api.listAgents();
      setAgents(agentItems);
      reconcileAgentUiStatuses(agentItems);
      setAgentsLoaded(true);
    } catch (error) {
      toast.error(actionErrorMessage(error));
    }
  }

  async function runAction(
    action: () => Promise<unknown>,
    refresh = true,
    successMessage = "操作已完成",
  ) {
    setLoading(true);
    try {
      await action();
      if (refresh) {
        await refreshAll();
      }
      toast.success(successMessage);
    } catch (error) {
      toast.error(actionErrorMessage(error));
    } finally {
      setLoading(false);
    }
  }

  function actionErrorMessage(error: unknown) {
    const message = (error as Error).message || "操作失败";
    if (message.toLowerCase().includes("tmux")) {
      return `${message}。请重建包含 tmux 的 coding 镜像。`;
    }
    return message;
  }

  async function runAgentAction(
    agentId: string,
    status: AgentUiStatus,
    action: () => Promise<unknown>,
    successMessage: string,
  ) {
    setAgentUiStatus(agentId, status);
    try {
      await runAction(action, true, successMessage);
    } finally {
      setAgentUiStatus(agentId, undefined);
    }
  }

  async function hydrateWorkbenchSettings() {
    const localSettings = loadLocalWorkbenchSettings();
    try {
      const persisted = await api.getSetting(workbenchSettingsServerKey);
      const settings = decodeWorkbenchSettings(persisted.value);
      loadedWorkbenchSettingsRef.current = encodeWorkbenchSettings(settings);
      setWorkbenchSettings(settings);
      saveWorkbenchSettings(settings);
    } catch (error) {
      const fallback = localSettings ?? defaultWorkbenchSettings;
      const encodedFallback = encodeWorkbenchSettings(fallback);
      loadedWorkbenchSettingsRef.current = encodedFallback;
      setWorkbenchSettings(fallback);
      saveWorkbenchSettings(fallback);
      if (!isWorkbenchSettingNotFound(error)) {
        toast.error(`加载工作台设置失败：${(error as Error).message}`);
        setWorkbenchSettingsHydrated(true);
        return;
      }
      try {
        await api.updateSetting(workbenchSettingsServerKey, encodedFallback);
      } catch (error) {
        toast.error(`初始化工作台设置失败：${(error as Error).message}`);
      }
    } finally {
      setWorkbenchSettingsHydrated(true);
    }
  }

  function isWorkbenchSettingNotFound(error: unknown) {
    return (
      error instanceof ApiError &&
      (error.errorCode === agentBoxSettingNotFoundErrorCode ||
        error.status === 404)
    );
  }

  function setTheme(nextTheme: ThemeMode) {
    setThemeMode(nextTheme);
    localStorage.setItem("john-ai-agentbox-theme", nextTheme);
  }

  function resetAuthenticatedData() {
    setCurrentUser(null);
    setProviders([]);
    setImages([]);
    setAgents([]);
    setAgentsLoaded(false);
    setSelectedAgentId("");
    setAgentUiStatuses({});
    setGitDirtyStates({});
    setWorkbenchSettingsHydrated(false);
    loadedWorkbenchSettingsRef.current = null;
    setViewMode("agents");
  }

  function handleSessionExpired() {
    resetAuthenticatedData();
    setAuthStatus("anonymous");
  }

  async function handleLogin(username: string, password: string) {
    const session = await api.login({ username, password });
    setCurrentUser(session.user);
    setAuthStatus("authenticated");
    if (resolvePortalPath() === loginPath || window.location.pathname !== "/") {
      window.history.replaceState(null, "", "/");
    }
    setCurrentPath(portalRootPath);
  }

  async function handleLogout() {
    try {
      await api.logout();
    } catch {
      // Logout should return the UI to the anonymous boundary even if the
      // session was already expired server-side.
    }
    handleSessionExpired();
  }

  function openProviderDialog(mode: EntityMode, provider?: ProviderInfo) {
    setProviderForm(provider ? providerToForm(provider) : blankProviderForm());
    setProviderDialog({ open: true, mode });
  }

  function openModelDialog(provider?: ProviderInfo) {
    setModelForm(blankModelForm(provider?.id ?? providers[0]?.id ?? 0));
    setModelDialogOpen(true);
  }

  function openImageDialog(mode: EntityMode, image?: CodingImageInfo) {
    setImageForm(image ? imageToForm(image) : blankImageForm());
    setImageDialog({ open: true, mode });
  }

  function openAgentDialog(mode: EntityMode, agent?: AgentInfo) {
    setAgentForm(
      agent ? agentToForm(agent) : defaultAgentForm(providers, images),
    );
    setAgentDialog({ open: true, mode });
  }

  async function submitProvider() {
    const payload = {
      name: providerForm.name,
      homepageUrl: providerForm.homepageUrl,
      notes: providerForm.notes,
      apiKey: providerForm.apiKey,
      openaiBaseUrl: providerForm.openaiBaseUrl,
      anthropicBaseUrl: providerForm.anthropicBaseUrl,
    };
    if (providerDialog.mode === "edit" && providerForm.id) {
      await runAction(
        () => api.updateProvider(providerForm.id, payload),
        true,
        "供应商已更新",
      );
    } else {
      await runAction(() => api.createProvider(payload), true, "供应商已创建");
    }
    setProviderDialog((current) => ({ ...current, open: false }));
  }

  async function submitModel() {
    await runAction(
      () =>
        api.addProviderModel(modelForm.providerId, {
          name: modelForm.name,
          protocol: modelForm.protocol,
        }),
      true,
      "模型已新增",
    );
    setModelDialogOpen(false);
  }

  async function syncModels(
    protocol: "anthropic" | "openai",
    providerId = modelForm.providerId,
  ) {
    await runAction(
      () => api.syncProviderModels(providerId, protocol),
      true,
      "模型同步已完成",
    );
  }

  async function syncProviderModels(provider: ProviderInfo) {
    const protocols = protocolsForSync(provider);
    await runAction(
      async () => {
        for (const nextProtocol of protocols) {
          await api.syncProviderModels(provider.id, nextProtocol);
        }
      },
      true,
      "模型同步已完成",
    );
  }

  async function submitImage() {
    const payload = {
      name: imageForm.name,
      imageRef: imageForm.imageRef,
      agentType: imageForm.agentType,
      defaultShell: imageForm.defaultShell,
      notes: imageForm.notes,
      enabled: imageForm.enabled,
    };
    if (imageDialog.mode === "edit" && imageForm.id) {
      await runAction(
        () => api.updateImage(imageForm.id, payload),
        true,
        "镜像已更新",
      );
    } else {
      await runAction(() => api.createImage(payload), true, "镜像已创建");
    }
    setImageDialog((current) => ({ ...current, open: false }));
  }

  async function submitAgent() {
    if (agentDialog.mode === "edit" && agentForm.id) {
      await runAction(
        () =>
          api.updateAgent(agentForm.id, {
            name: agentForm.name,
            providerId: agentForm.providerId,
            modelName: agentForm.modelName,
            modelProtocol: agentForm.modelProtocol,
            agentType: agentForm.agentType,
            iconKey: agentForm.iconKey,
            notes: agentForm.notes,
          }),
        true,
        "智能体已更新",
      );
    } else {
      await runAction(
        () =>
          api.createAgent({
            name: agentForm.name,
            providerId: agentForm.providerId,
            modelName: agentForm.modelName,
            modelProtocol: agentForm.modelProtocol,
            imageId: agentForm.imageId,
            agentType: agentForm.agentType,
            iconKey: agentForm.iconKey,
            notes: agentForm.notes,
          }),
        true,
        "智能体已创建",
      );
    }
    setAgentDialog((current) => ({ ...current, open: false }));
  }

  function askDeleteAgent(agent: AgentInfo) {
    setDeleteDialog({
      open: true,
      kind: "agent",
      title: "删除智能体",
      description:
        "删除智能体会停止并移除运行时。勾选后会同步删除该智能体的 Docker 数据卷。",
      targetName: agent.name,
      deleteVolumes: true,
      agentId: agent.id,
      providerId: 0,
      imageId: 0,
      modelProviderId: 0,
      modelId: 0,
    });
  }

  function askDeleteProvider(provider: ProviderInfo) {
    setDeleteDialog({
      open: true,
      kind: "provider",
      title: "删除供应商",
      description:
        "删除后会移除供应商配置和模型列表；已被智能体使用的供应商无法删除。",
      targetName: provider.name,
      deleteVolumes: false,
      agentId: "",
      providerId: provider.id,
      imageId: 0,
      modelProviderId: 0,
      modelId: 0,
    });
  }

  function askDeleteImage(image: CodingImageInfo) {
    setDeleteDialog({
      open: true,
      kind: "image",
      title: "删除镜像",
      description: image.isDefault
        ? "默认镜像不能删除。"
        : "删除后该镜像配置会从注册表移除；已被智能体使用的镜像无法删除。",
      targetName: image.name,
      deleteVolumes: false,
      agentId: "",
      providerId: 0,
      imageId: image.id,
      modelProviderId: 0,
      modelId: 0,
    });
  }

  function askDeleteModel(provider: ProviderInfo, model: ProviderModel) {
    setDeleteDialog({
      open: true,
      kind: "providerModel",
      title: "删除模型",
      description:
        "删除后该模型会从供应商模型列表移除；正在被智能体使用的模型无法删除。",
      targetName: `${provider.name} / ${model.name}`,
      deleteVolumes: false,
      agentId: "",
      providerId: 0,
      imageId: 0,
      modelProviderId: provider.id,
      modelId: model.id,
    });
  }

  function askStopAgent(agent: AgentInfo) {
    setStopTarget(agent);
    setStopDialogOpen(true);
  }

  async function confirmDelete() {
    const dialog = { ...deleteDialog };
    setDeleteDialog(blankDeleteDialog());
    if (dialog.kind === "agent") {
      await runAgentAction(
        dialog.agentId,
        "deleting",
        () => api.deleteAgent(dialog.agentId, dialog.deleteVolumes),
        "智能体已删除",
      );
    } else if (dialog.kind === "provider") {
      await runAction(
        () => api.deleteProvider(dialog.providerId),
        true,
        "供应商已删除",
      );
    } else if (dialog.kind === "image") {
      await runAction(
        () => api.deleteImage(dialog.imageId),
        true,
        "镜像已删除",
      );
    } else {
      await runAction(
        () => api.deleteProviderModel(dialog.modelProviderId, dialog.modelId),
        true,
        "模型已删除",
      );
    }
  }

  function askChangeImage(agent = selectedAgent) {
    if (!agent) {
      return;
    }
    setChangeImageDialog({
      open: true,
      imageId: agent.imageId || images[0]?.id || 0,
    });
  }

  async function confirmChangeImage() {
    if (!selectedAgent || !changeImageDialog.imageId) {
      return;
    }
    const imageId = changeImageDialog.imageId;
    setChangeImageDialog({ open: false, imageId: 0 });
    await runAction(
      () => api.changeAgentImage(selectedAgent.id, imageId),
      true,
      "镜像已更换",
    );
  }

  function selectAgentPanel(panel: PanelMode) {
    if (!selectedAgent?.id || !visiblePanelModes.includes(panel)) {
      return;
    }
    setAgentWorkbenchState((current) => ({
      ...current,
      [selectedAgent.id]: {
        ...defaultWorkbenchState,
        ...current[selectedAgent.id],
        panelMode: panel,
      },
    }));
  }

  function selectAgent(
    agent: AgentInfo,
    nextViewMode: ViewMode = "agentDetail",
  ) {
    setSelectedAgentId(agent.id);
    setViewMode(nextViewMode);
  }

  function openAgentDetailDialog(agent: AgentInfo) {
    selectAgent(agent);
    setAgentDetailOpen(true);
  }

  function startAgent(agent: AgentInfo) {
    selectAgent(agent);
    void runAgentAction(
      agent.id,
      "starting",
      () => api.startAgent(agent.id),
      "智能体已启动",
    );
  }

  function stopAgent(agent: AgentInfo) {
    selectAgent(agent);
    askStopAgent(agent);
  }

  function editAgent(agent: AgentInfo) {
    selectAgent(agent);
    openAgentDialog("edit", agent);
  }

  function changeAgentImage(agent: AgentInfo) {
    selectAgent(agent);
    askChangeImage(agent);
  }

  function deleteAgent(agent: AgentInfo) {
    selectAgent(agent);
    askDeleteAgent(agent);
  }

  function showWorkspacePathInGit(path: string) {
    if (!selectedAgent?.id || !gitPanelVisible) {
      return;
    }
    selectAgentPanel("git");
    setWorkspaceLocateIntent({
      kind: "files-to-git",
      request: { id: Date.now(), workspacePath: path },
    });
  }

  function showGitPathInFiles(
    path: string,
    deleted?: boolean,
    type?: "file" | "directory",
  ) {
    if (!selectedAgent?.id || !filesPanelVisible) {
      return;
    }
    selectAgentPanel("files");
    setWorkspaceLocateIntent({
      kind: "git-to-files",
      request: { id: Date.now(), path, deleted, type },
    });
  }

  function clearWorkspaceLocateIntent(id: number) {
    setWorkspaceLocateIntent((current) => {
      if (!current || current.request.id !== id) {
        return current;
      }
      return null;
    });
  }

  function setSelectedAgentWorkspacePath(path: string) {
    if (!selectedAgent?.id) {
      return;
    }
    setAgentWorkbenchState((current) => ({
      ...current,
      [selectedAgent.id]: {
        ...defaultWorkbenchState,
        ...current[selectedAgent.id],
        workspacePath: path,
      },
    }));
  }

  function updateSelectedAgentChatState(
    chatState: Pick<
      AgentWorkbenchState,
      "activeChatSessionId" | "chatHistoryOpen"
    >,
  ) {
    if (!selectedAgent?.id) {
      return;
    }
    setAgentWorkbenchState((current) => ({
      ...current,
      [selectedAgent.id]: {
        ...defaultWorkbenchState,
        ...current[selectedAgent.id],
        ...chatState,
      },
    }));
  }

  const setSelectedAgentWorking = useCallback(
    (state?: "working" | "waiting_input") => {
      if (!selectedAgent?.id) {
        return;
      }
      setAgentUiStatuses((current) => {
        const currentStatus = current[selectedAgent.id];
        const nextStatus = state;
        if (
          currentStatus === nextStatus ||
          (currentStatus && !canChatOverrideAgentStatus(currentStatus))
        ) {
          return current;
        }
        if (nextStatus) {
          return { ...current, [selectedAgent.id]: nextStatus };
        }
        const next = { ...current };
        delete next[selectedAgent.id];
        return next;
      });
    },
    [selectedAgent?.id],
  );

  function setAgentUiStatus(agentId: string, status?: AgentUiStatus) {
    if (!agentId) {
      return;
    }
    setAgentUiStatuses((current) => {
      if (!status) {
        if (!(agentId in current)) {
          return current;
        }
        const next = { ...current };
        delete next[agentId];
        return next;
      }
      if (current[agentId] === status) {
        return current;
      }
      return { ...current, [agentId]: status };
    });
  }

  async function toggleAgentDetailWindowFullscreen() {
    const target = agentDetailRef.current;
    if (!target) {
      return;
    }
    try {
      if (document.fullscreenElement === target) {
        await exitDocumentFullscreen();
        return;
      }
      if (document.fullscreenElement) {
        await exitDocumentFullscreen();
      }
      await target.requestFullscreen();
    } catch (error) {
      toast.error(`切换全屏失败：${(error as Error).message}`);
    }
  }

  function reconcileAgentUiStatuses(agentItems: AgentInfo[]) {
    const existingIds = new Set(agentItems.map((item) => item.id));
    setAgentUiStatuses((current) => {
      let changed = false;
      const next: AgentUiStatusMap = {};
      Object.entries(current).forEach(([agentId, status]) => {
        if (!existingIds.has(agentId)) {
          changed = true;
          return;
        }
        next[agentId] = status;
      });
      return changed ? next : current;
    });
  }

  const viewEyebrow =
    viewMode === "agentDetail"
      ? "智能体详情"
      : viewMode === "providers"
        ? "供应商"
        : viewMode === "aiCapabilities"
          ? "AI 能力"
          : viewMode === "prompts"
            ? "提示词"
            : viewMode === "images"
              ? "镜像"
              : "智能体";
  const viewTitle =
    viewMode === "agentDetail"
      ? selectedAgent?.name || "智能体详情"
      : viewMode === "providers"
        ? "供应商管理"
        : viewMode === "aiCapabilities"
          ? "AI 能力"
          : viewMode === "prompts"
            ? "提示词管理"
            : viewMode === "images"
              ? "镜像管理"
              : "智能体管理";
  const appLayoutStyle = {
    "--agent-layout-sidebar-width": sidebarCollapsed ? "56px" : "320px",
  } as CSSProperties;

  if (authStatus === "checking") {
    return (
      <main className={themeMode === "dark" ? "dark" : ""}>
        <Toaster theme={themeMode} />
        <AuthShell>
          <Spinner label="正在验证登录状态" />
        </AuthShell>
      </main>
    );
  }

  if (authStatus === "anonymous") {
    return (
      <main className={themeMode === "dark" ? "dark" : ""}>
        <Toaster theme={themeMode} />
        {isLoginPath ? (
          <LoginPage
            themeMode={themeMode}
            onLogin={handleLogin}
            onThemeToggle={() =>
              setTheme(themeMode === "dark" ? "light" : "dark")
            }
          />
        ) : (
          <PortalLoginGate
            themeMode={themeMode}
            onLoginPathChange={setCurrentPath}
            onThemeToggle={() =>
              setTheme(themeMode === "dark" ? "light" : "dark")
            }
          />
        )}
      </main>
    );
  }

  return (
    <main className={themeMode === "dark" ? "dark" : ""}>
      <TooltipProvider>
        <Toaster theme={themeMode} />
        <div
          className={cn(
            "app-shell bg-background text-foreground transition-colors",
            viewMode === "agentDetail" && "app-shell-workbench",
          )}
          data-testid="agentbox-app-shell"
        >
          <div
            className={cn(
              "app-layout grid",
              viewMode === "agentDetail" && "app-layout-workbench",
            )}
            style={appLayoutStyle}
          >
            <aside
              className={cn(
                "app-sidebar flex min-h-0 flex-col gap-4 border-r border-border bg-sidebar py-5 max-[980px]:border-b max-[980px]:border-r-0",
                sidebarCollapsed ? "items-center px-1.5" : "px-5",
              )}
            >
              <div
                className={cn(
                  "flex items-center gap-3",
                  sidebarCollapsed ? "flex-col" : "justify-between",
                )}
              >
                <div className="flex min-w-0 items-center gap-3">
                  <div className="grid h-10 w-10 shrink-0 place-items-center rounded-[6px] border border-primary/30 bg-primary text-primary-foreground">
                    <Boxes className="h-5 w-5" />
                  </div>
                  {!sidebarCollapsed ? (
                    <div className="min-w-0">
                      <h1 className="truncate text-lg font-semibold">
                        Agent Box
                      </h1>
                      <p className="truncate text-xs text-muted-foreground">
                        智能体运行时管理
                      </p>
                    </div>
                  ) : null}
                </div>
                {sidebarCollapsed ? null : (
                  <Button
                    aria-label="收起左侧菜单"
                    className="h-9 shrink-0 px-3 text-xs"
                    title="收起左侧菜单"
                    type="button"
                    variant="soft"
                    onClick={() => setSidebarCollapsed(true)}
                  >
                    <PanelLeftCloseIcon className="h-4 w-4" />
                    收起
                  </Button>
                )}
              </div>

              {sidebarCollapsed ? (
                <CollapsedIconButton
                  className="h-10 w-10 border border-primary/30 bg-primary/10 text-primary hover:bg-primary/15"
                  label="展开左侧菜单"
                  onClick={() => setSidebarCollapsed(false)}
                >
                  <PanelLeftOpenIcon className="h-4 w-4" />
                </CollapsedIconButton>
              ) : null}

              <nav
                className={cn(
                  "gap-2",
                  sidebarCollapsed
                    ? "grid grid-cols-1 justify-items-center"
                    : "grid grid-cols-2",
                )}
              >
                <NavButton
                  active={viewMode === "agents" || viewMode === "agentDetail"}
                  collapsed={sidebarCollapsed}
                  label="智能体"
                  onClick={() => setViewMode("agents")}
                >
                  <Activity className="h-4 w-4" />{" "}
                  {!sidebarCollapsed ? "智能体" : null}
                </NavButton>
                <NavButton
                  active={viewMode === "providers"}
                  collapsed={sidebarCollapsed}
                  label="供应商"
                  onClick={() => setViewMode("providers")}
                >
                  <KeyRound className="h-4 w-4" />{" "}
                  {!sidebarCollapsed ? "供应商" : null}
                </NavButton>
                <NavButton
                  active={viewMode === "aiCapabilities"}
                  collapsed={sidebarCollapsed}
                  label="AI 能力"
                  onClick={() => setViewMode("aiCapabilities")}
                >
                  <BrainCircuit className="h-4 w-4" />{" "}
                  {!sidebarCollapsed ? "AI 能力" : null}
                </NavButton>
                <NavButton
                  active={viewMode === "prompts"}
                  collapsed={sidebarCollapsed}
                  label="提示词"
                  onClick={() => setViewMode("prompts")}
                >
                  <ScrollText className="h-4 w-4" />{" "}
                  {!sidebarCollapsed ? "提示词" : null}
                </NavButton>
                <NavButton
                  active={viewMode === "images"}
                  collapsed={sidebarCollapsed}
                  label="镜像"
                  onClick={() => setViewMode("images")}
                >
                  <Image className="h-4 w-4" />{" "}
                  {!sidebarCollapsed ? "镜像" : null}
                </NavButton>
                <NavButton
                  active={workbenchSettingsOpen}
                  collapsed={sidebarCollapsed}
                  label="设置"
                  testId="workbench-settings-nav-button"
                  onClick={() => setWorkbenchSettingsOpen(true)}
                >
                  <Settings className="h-4 w-4" />{" "}
                  {!sidebarCollapsed ? "设置" : null}
                </NavButton>
              </nav>

              <div
                className={cn(
                  "rounded-[6px] border border-border bg-card",
                  sidebarCollapsed
                    ? "grid place-items-center p-1.5"
                    : "flex items-center gap-2 p-2.5",
                )}
              >
                {!sidebarCollapsed ? (
                  <>
                    <div className="grid h-9 w-9 shrink-0 place-items-center rounded-[6px] bg-muted text-muted-foreground">
                      <UserCircle className="h-4 w-4" />
                    </div>
                    <div className="min-w-0 flex-1">
                      <div className="truncate text-sm font-medium">
                        {currentUser?.username ?? "admin"}
                      </div>
                      <div className="truncate text-xs text-muted-foreground">
                        {currentUser?.role ?? "admin"}
                      </div>
                    </div>
                    <IconButton
                      data-testid="agentbox-logout-button"
                      title="退出登录"
                      onClick={() => void handleLogout()}
                    >
                      <LogOut className="h-4 w-4" />
                    </IconButton>
                  </>
                ) : (
                  <CollapsedIconButton
                    data-testid="agentbox-logout-button"
                    label="退出登录"
                    onClick={() => void handleLogout()}
                  >
                    <LogOut className="h-4 w-4" />
                  </CollapsedIconButton>
                )}
              </div>

              {sidebarCollapsed ? (
                <div className="flex flex-col items-center gap-2">
                  <CollapsedIconButton
                    label="切换浅色/深色样式"
                    onClick={() =>
                      setTheme(themeMode === "dark" ? "light" : "dark")
                    }
                  >
                    {themeMode === "dark" ? (
                      <Sun className="h-4 w-4" />
                    ) : (
                      <Moon className="h-4 w-4" />
                    )}
                  </CollapsedIconButton>
                </div>
              ) : null}

              {!sidebarCollapsed ? (
                <>
                  <div className="flex items-center justify-between gap-3">
                    <div className="text-xs font-medium uppercase text-muted-foreground">
                      智能体列表
                    </div>
                    <div className="flex gap-2">
                      <IconButton
                        title="切换浅色/深色样式"
                        onClick={() =>
                          setTheme(themeMode === "dark" ? "light" : "dark")
                        }
                      >
                        {themeMode === "dark" ? (
                          <Sun className="h-4 w-4" />
                        ) : (
                          <Moon className="h-4 w-4" />
                        )}
                      </IconButton>
                    </div>
                  </div>

                  <section className="flex min-h-0 flex-1 flex-col gap-2 overflow-auto">
                    {displayAgents.length === 0 ? (
                      <div className="rounded-[6px] border border-dashed border-border p-4 text-center text-sm text-muted-foreground">
                        暂无智能体
                      </div>
                    ) : null}
                    {displayAgents.map((item) => (
                      <AgentContextMenu
                        key={item.id}
                        agent={item}
                        trigger={
                          <AgentSidebarButton
                            active={selectedAgent?.id === item.id}
                            agent={item}
                            onClick={() => selectAgent(item)}
                          />
                        }
                        onChangeImage={changeAgentImage}
                        onDelete={deleteAgent}
                        onEdit={editAgent}
                        onOpenDetail={openAgentDetailDialog}
                        onStart={startAgent}
                        onStop={stopAgent}
                      />
                    ))}
                  </section>
                </>
              ) : (
                <section className="flex min-h-0 flex-1 flex-col items-center gap-2 overflow-auto">
                  {displayAgents.length === 0 ? (
                    <div className="rounded-[6px] border border-dashed border-border p-2 text-center text-xs text-muted-foreground">
                      无
                    </div>
                  ) : null}
                  {displayAgents.map((item) => (
                    <AgentContextMenu
                      key={item.id}
                      agent={item}
                      trigger={
                        <CollapsedAgentButton
                          active={selectedAgent?.id === item.id}
                          agent={item}
                          onClick={() => selectAgent(item)}
                        />
                      }
                      onChangeImage={changeAgentImage}
                      onDelete={deleteAgent}
                      onEdit={editAgent}
                      onOpenDetail={openAgentDetailDialog}
                      onStart={startAgent}
                      onStop={stopAgent}
                    />
                  ))}
                </section>
              )}
            </aside>

            <section
              className={cn(
                "app-content grid min-h-0 min-w-0 max-[980px]:min-h-[860px]",
                viewMode === "agentDetail"
                  ? "grid-rows-[minmax(0,1fr)] gap-0 p-0 max-[980px]:min-h-0"
                  : "grid-rows-[auto_auto_minmax(0,1fr)] gap-4 p-6 max-[640px]:p-4",
              )}
            >
              {viewMode !== "agentDetail" ? (
                <header className="flex min-w-0 items-start justify-between gap-4 max-[760px]:flex-col">
                  <div className="min-w-0">
                    <p className="mb-1 text-xs font-semibold text-muted-foreground">
                      {viewEyebrow}
                    </p>
                    <h2 className="truncate text-2xl font-semibold leading-tight">
                      {viewTitle}
                    </h2>
                  </div>
                  <div className="flex flex-wrap justify-end gap-2">
                    {viewMode === "agents" ? (
                      <Button
                        disabled={loading}
                        type="button"
                        variant="primary"
                        onClick={() => openAgentDialog("create")}
                      >
                        <Plus className="h-4 w-4" /> 新增智能体
                      </Button>
                    ) : null}
                    {viewMode === "providers" ? (
                      <>
                        <Button
                          disabled={loading}
                          type="button"
                          variant="primary"
                          onClick={() => openProviderDialog("create")}
                        >
                          <Plus className="h-4 w-4" /> 新增供应商
                        </Button>
                        <Button
                          disabled={loading || providers.length === 0}
                          type="button"
                          variant="soft"
                          onClick={() => openModelDialog()}
                        >
                          <Plus className="h-4 w-4" /> 新增模型
                        </Button>
                      </>
                    ) : null}
                    {viewMode === "aiCapabilities" ? (
                      <Button
                        disabled={loading}
                        type="button"
                        variant="soft"
                        onClick={() => void refreshAll()}
                      >
                        <RefreshCw className="h-4 w-4" /> 刷新供应商
                      </Button>
                    ) : null}
                    {viewMode === "images" ? (
                      <Button
                        disabled={loading}
                        type="button"
                        variant="primary"
                        onClick={() => openImageDialog("create")}
                      >
                        <Plus className="h-4 w-4" /> 新增镜像
                      </Button>
                    ) : null}
                    <Button
                      disabled={loading}
                      type="button"
                      variant="soft"
                      onClick={() => void refreshAll()}
                    >
                      <RefreshCw className="h-4 w-4" /> 刷新
                    </Button>
                  </div>
                </header>
              ) : null}

              {viewMode !== "agentDetail" ? <div /> : null}

              {viewMode === "agents" ? (
                <ManagementCard>
                  <ManagementToolbar
                    pageSize={agentPager.pageSize}
                    placeholder="搜索智能体、模型、供应商、镜像"
                    query={agentPager.query}
                    total={filteredAgents.length}
                    onPageSize={(pageSize) =>
                      setAgentPager((current) => ({ ...current, pageSize }))
                    }
                    onQuery={(query) =>
                      setAgentPager((current) => ({ ...current, query }))
                    }
                  />
                  <AgentTable
                    agents={pagedAgents}
                    selectedAgent={selectedAgent}
                    onChangeImage={(agent) => {
                      setSelectedAgentId(agent.id);
                      askChangeImage(agent);
                    }}
                    onDelete={askDeleteAgent}
                    onEdit={(agent) => openAgentDialog("edit", agent)}
                    onSelect={(agent) => setSelectedAgentId(agent.id)}
                    onStart={(agent) =>
                      void runAgentAction(
                        agent.id,
                        "starting",
                        () => api.startAgent(agent.id),
                        "智能体已启动",
                      )
                    }
                    onStop={askStopAgent}
                  />
                  <Pagination
                    label={`显示 ${agentPageInfo.start}-${agentPageInfo.end} / ${agentPageInfo.total}`}
                    page={agentPageInfo.current}
                    pages={agentPageInfo.pages}
                    onNext={() =>
                      setAgentPager((current) => ({
                        ...current,
                        page: Math.min(agentPageInfo.pages, current.page + 1),
                      }))
                    }
                    onPrev={() =>
                      setAgentPager((current) => ({
                        ...current,
                        page: Math.max(1, current.page - 1),
                      }))
                    }
                  />
                </ManagementCard>
              ) : null}

              {viewMode === "agentDetail" ? (
                <section
                  ref={agentDetailRef}
                  className={cn(
                    "h-full min-h-0 min-w-0 max-w-full overflow-hidden bg-background max-[980px]:max-w-[100dvw]",
                    agentDetailFullscreen && "h-screen w-screen bg-background",
                  )}
                  data-fullscreen={agentDetailFullscreen ? "window" : "false"}
                  data-testid="agent-detail-workbench"
                >
                  <section
                    className={cn(
                      "agent-detail-shell grid h-full min-h-0 min-w-0 max-w-full overflow-hidden bg-card max-[980px]:max-w-[100dvw]",
                      "grid-rows-[auto_minmax(0,1fr)]",
                    )}
                  >
                    <div
                      className="flex min-h-11 items-center gap-2 overflow-visible border-b border-border bg-muted px-2 max-[760px]:flex-wrap"
                      data-testid="agent-detail-toolbar"
                    >
                      <div
                        className="flex min-w-0 shrink-0 items-center gap-1 overflow-auto"
                        data-testid="agent-detail-panel-tabs"
                        role="tablist"
                      >
                        {visiblePanelModes.map((key) => {
                          const { Icon, label } =
                            agentDetailPanelDefinitions[key];
                          return (
                            <TabButton
                              active={panelMode === key}
                              key={String(key)}
                              role="tab"
                              onClick={() => selectAgentPanel(key)}
                            >
                              <span className="relative inline-flex min-w-0 items-center gap-1.5">
                                <Icon className="h-4 w-4" />
                                <span>{label}</span>
                                {key === "git" && selectedGitDirty ? (
                                  <span
                                    aria-hidden="true"
                                    className="dot git-dirty absolute -right-2 -top-1"
                                    data-testid="git-tab-dirty-badge"
                                  />
                                ) : null}
                              </span>
                            </TabButton>
                          );
                        })}
                      </div>
                      <div
                        className="flex min-w-0 flex-1 justify-center max-[1180px]:order-3 max-[1180px]:w-full max-[1180px]:basis-full"
                        data-testid="workspace-path-slot"
                      >
                        <WorkspacePathPicker
                          agent={selectedAgent}
                          value={workspacePath}
                          variant="inline"
                          onChange={setSelectedAgentWorkspacePath}
                        />
                      </div>
                      <div
                        className="flex shrink-0 items-center gap-1 max-[760px]:ml-auto"
                        data-testid="agent-detail-right-menu"
                      >
                        {selectedAgent ? (
                          <div
                            className="flex min-w-0 shrink-0 items-center gap-1"
                            data-testid="agent-detail-metadata"
                          >
                            <Button
                              aria-label={`编辑模型 ${selectedAgent.modelName || "未配置"}`}
                              className="h-8 shrink-0 gap-1.5 whitespace-nowrap border-transparent px-2 text-xs hover:border-primary/60 hover:bg-primary/15 hover:text-foreground hover:shadow-sm hover:ring-1 hover:ring-primary/20"
                              data-testid="agent-detail-model-chip"
                              size="sm"
                              title="编辑模型配置"
                              type="button"
                              variant="ghost"
                              onClick={() => editAgent(selectedAgent)}
                            >
                              <BrainCircuit data-icon="inline-start" />
                              <span className="text-muted-foreground">
                                模型
                              </span>
                              <span data-testid="agent-detail-model-name">
                                {selectedAgent.modelName || "-"}
                              </span>
                            </Button>
                            <Button
                              aria-label={`更换镜像 ${selectedAgent.imageName || selectedAgent.imageRef || "未配置"}`}
                              className="h-8 max-w-36 gap-1.5 border-transparent px-2 text-xs hover:border-primary/60 hover:bg-primary/15 hover:text-foreground hover:shadow-sm hover:ring-1 hover:ring-primary/20"
                              data-testid="agent-detail-image-chip"
                              size="sm"
                              title="更换镜像"
                              type="button"
                              variant="ghost"
                              onClick={() => changeAgentImage(selectedAgent)}
                            >
                              <Image data-icon="inline-start" />
                              <span className="text-muted-foreground">
                                镜像
                              </span>
                              <span className="min-w-0 truncate">
                                {selectedAgent.imageName ||
                                  selectedAgent.imageRef ||
                                  "-"}
                              </span>
                            </Button>
                          </div>
                        ) : null}
                        <div
                          className="flex shrink-0 items-center gap-1"
                          data-testid="agent-detail-actions"
                        >
                          <TabButton
                            aria-label={
                              agentDetailFullscreen
                                ? "退出 Agent 详情全屏"
                                : "Agent 详情全屏"
                            }
                            className="h-7"
                            title={agentDetailFullscreen ? "退出全屏" : "全屏"}
                            onClick={() =>
                              void toggleAgentDetailWindowFullscreen()
                            }
                          >
                            {agentDetailFullscreen ? (
                              <Shrink className="h-4 w-4" />
                            ) : (
                              <Fullscreen className="h-4 w-4" />
                            )}
                          </TabButton>
                        </div>
                      </div>
                    </div>
                    <div className="h-full min-h-0 min-w-0 max-w-full overflow-hidden max-[980px]:max-w-[100dvw]">
                      <ChatPanel
                        active={
                          viewMode === "agentDetail" && panelMode === "chat"
                        }
                        agent={selectedAgent}
                        chatHistoryOpen={Boolean(
                          selectedAgentWorkbench.chatHistoryOpen,
                        )}
                        connected={viewMode === "agentDetail"}
                        preferredSessionId={
                          selectedAgentWorkbench.activeChatSessionId
                        }
                        workspacePath={workspacePath}
                        onAgentWorkingChange={setSelectedAgentWorking}
                        onChatStateChange={updateSelectedAgentChatState}
                      />
                      <Suspense fallback={null}>
                        <ShellPanel
                          active={
                            viewMode === "agentDetail" && panelMode === "shell"
                          }
                          agent={selectedAgent}
                          agents={displayAgents}
                          settings={workbenchSettings}
                          workspacePath={workspacePath}
                        />
                      </Suspense>
                      <ServicesPanel
                        active={
                          viewMode === "agentDetail" && panelMode === "services"
                        }
                        agent={selectedAgent}
                      />
                      <SkillsPanel
                        active={
                          viewMode === "agentDetail" && panelMode === "skills"
                        }
                        agent={selectedAgent}
                        workspacePath={workspacePath}
                      />
                      <FilesPanel
                        active={
                          viewMode === "agentDetail" && panelMode === "files"
                        }
                        agent={selectedAgent}
                        agents={displayAgents}
                        locateRequest={
                          workspaceLocateIntent?.kind === "git-to-files"
                            ? workspaceLocateIntent.request
                            : null
                        }
                        settings={workbenchSettings}
                        workspacePath={workspacePath}
                        canShowInGit={gitPanelVisible}
                        onLocateRequestHandled={clearWorkspaceLocateIntent}
                        onShowInGit={showWorkspacePathInGit}
                      />
                      <GitPanel
                        active={
                          viewMode === "agentDetail" && panelMode === "git"
                        }
                        agent={selectedAgent}
                        locateRequest={
                          workspaceLocateIntent?.kind === "files-to-git"
                            ? workspaceLocateIntent.request
                            : null
                        }
                        settings={workbenchSettings}
                        workspacePath={workspacePath}
                        canLocateInFiles={filesPanelVisible}
                        onDirtyStateChange={updateGitDirtyState}
                        onLocateInFiles={showGitPathInFiles}
                        onLocateRequestHandled={clearWorkspaceLocateIntent}
                      />
                    </div>
                  </section>
                </section>
              ) : null}

              {viewMode === "providers" ? (
                <ManagementCard>
                  <ManagementToolbar
                    pageSize={providerPager.pageSize}
                    placeholder="搜索供应商、端点、模型"
                    query={providerPager.query}
                    total={filteredProviders.length}
                    onPageSize={(pageSize) =>
                      setProviderPager((current) => ({ ...current, pageSize }))
                    }
                    onQuery={(query) =>
                      setProviderPager((current) => ({ ...current, query }))
                    }
                  />
                  <ProviderTable
                    providers={pagedProviders}
                    onAddModel={openModelDialog}
                    onDelete={askDeleteProvider}
                    onDeleteModel={askDeleteModel}
                    onEdit={(provider) => openProviderDialog("edit", provider)}
                    onSync={(provider) => void syncProviderModels(provider)}
                  />
                  <Pagination
                    label={`显示 ${providerPageInfo.start}-${providerPageInfo.end} / ${providerPageInfo.total}`}
                    page={providerPageInfo.current}
                    pages={providerPageInfo.pages}
                    onNext={() =>
                      setProviderPager((current) => ({
                        ...current,
                        page: Math.min(
                          providerPageInfo.pages,
                          current.page + 1,
                        ),
                      }))
                    }
                    onPrev={() =>
                      setProviderPager((current) => ({
                        ...current,
                        page: Math.max(1, current.page - 1),
                      }))
                    }
                  />
                </ManagementCard>
              ) : null}

              {viewMode === "aiCapabilities" ? (
                <AICapabilitiesPage providers={providers} />
              ) : null}

              {viewMode === "prompts" ? <PromptsPage /> : null}

              {viewMode === "images" ? (
                <ManagementCard>
                  <ManagementToolbar
                    pageSize={imagePager.pageSize}
                    placeholder="搜索镜像名称、地址、类型、终端"
                    query={imagePager.query}
                    total={filteredImages.length}
                    onPageSize={(pageSize) =>
                      setImagePager((current) => ({ ...current, pageSize }))
                    }
                    onQuery={(query) =>
                      setImagePager((current) => ({ ...current, query }))
                    }
                  />
                  <ImageTable
                    images={pagedImages}
                    onDelete={askDeleteImage}
                    onEdit={(image) => openImageDialog("edit", image)}
                  />
                  <Pagination
                    label={`显示 ${imagePageInfo.start}-${imagePageInfo.end} / ${imagePageInfo.total}`}
                    page={imagePageInfo.current}
                    pages={imagePageInfo.pages}
                    onNext={() =>
                      setImagePager((current) => ({
                        ...current,
                        page: Math.min(imagePageInfo.pages, current.page + 1),
                      }))
                    }
                    onPrev={() =>
                      setImagePager((current) => ({
                        ...current,
                        page: Math.max(1, current.page - 1),
                      }))
                    }
                  />
                </ManagementCard>
              ) : null}
            </section>
          </div>

          <AgentDialog
            dialog={agentDialog}
            form={agentForm}
            images={images}
            providerModels={providerModelsForForm}
            providers={providers}
            setForm={setAgentForm}
            onClose={() =>
              setAgentDialog((current) => ({ ...current, open: false }))
            }
            onSubmit={() => void submitAgent()}
          />
          <ProviderDialog
            dialog={providerDialog}
            form={providerForm}
            setForm={setProviderForm}
            onClose={() =>
              setProviderDialog((current) => ({ ...current, open: false }))
            }
            onSubmit={() => void submitProvider()}
          />
          <ModelDialog
            form={modelForm}
            open={modelDialogOpen}
            providers={providers}
            setForm={setModelForm}
            onClose={() => setModelDialogOpen(false)}
            onSubmit={() => void submitModel()}
            onSync={() => void syncModels(modelForm.protocol)}
          />
          <ImageDialog
            dialog={imageDialog}
            form={imageForm}
            setForm={setImageForm}
            onClose={() =>
              setImageDialog((current) => ({ ...current, open: false }))
            }
            onSubmit={() => void submitImage()}
          />
          <ConfirmDialog
            danger
            confirmText={
              deleteDialog.kind === "agent"
                ? "删除"
                : deleteDialog.kind === "image" &&
                    images.find((item) => item.id === deleteDialog.imageId)
                      ?.isDefault
                  ? "不可删除"
                  : "删除"
            }
            description={
              <div className="grid gap-3">
                <p>{deleteDialog.description}</p>
                <p className="rounded-[6px] border border-border bg-muted/50 px-3 py-2 text-foreground">
                  {deleteDialog.targetName}
                </p>
                {deleteDialog.kind === "agent" ? (
                  <CheckboxField
                    checked={deleteDialog.deleteVolumes}
                    onChange={(event) =>
                      setDeleteDialog((current) => ({
                        ...current,
                        deleteVolumes: event.target.checked,
                      }))
                    }
                  >
                    同时删除该智能体的 Docker 数据卷
                  </CheckboxField>
                ) : null}
              </div>
            }
            disabled={
              deleteDialog.kind === "image" &&
              images.find((item) => item.id === deleteDialog.imageId)?.isDefault
            }
            open={deleteDialog.open}
            title={deleteDialog.title}
            onClose={() => setDeleteDialog(blankDeleteDialog())}
            onConfirm={() => void confirmDelete()}
          />
          <ConfirmDialog
            danger
            confirmText="停止"
            description={
              <div className="grid gap-3">
                <p>停止智能体会中断正在进行的任务，确认停止吗？</p>
                <p className="rounded-[6px] border border-border bg-muted/50 px-3 py-2 text-foreground">
                  {stopTarget?.name}
                </p>
                <p className="text-xs text-muted-foreground">
                  当前状态：{agentStatusText(stopTarget, agentUiStatuses)}
                </p>
              </div>
            }
            open={stopDialogOpen}
            title="停止智能体"
            onClose={() => {
              setStopDialogOpen(false);
              setStopTarget(null);
            }}
            onConfirm={() => {
              if (stopTarget) {
                void runAgentAction(
                  stopTarget.id,
                  "stopping",
                  () => api.stopAgent(stopTarget.id),
                  "智能体已停止",
                );
              }
              setStopDialogOpen(false);
              setStopTarget(null);
            }}
          />
          <Dialog
            footer={
              <>
                <Button
                  type="button"
                  variant="soft"
                  onClick={() =>
                    setChangeImageDialog({ open: false, imageId: 0 })
                  }
                >
                  取消
                </Button>
                <Button
                  type="button"
                  variant="danger"
                  onClick={() => void confirmChangeImage()}
                >
                  更换镜像
                </Button>
              </>
            }
            open={changeImageDialog.open}
            title="更换镜像"
            onClose={() => setChangeImageDialog({ open: false, imageId: 0 })}
          >
            <div className="grid gap-4">
              <p className="text-sm leading-6 text-muted-foreground">
                更换镜像会重置 /etc、/opt、/usr 和 /var，并保留 /home、/root、
                {defaultWorkspacePath} 与 {sharedWorkspacePath}。
              </p>
              <Field label="目标镜像">
                <Select
                  value={changeImageDialog.imageId}
                  onChange={(event) =>
                    setChangeImageDialog((current) => ({
                      ...current,
                      imageId: Number(event.target.value),
                    }))
                  }
                >
                  {images
                    .filter((item) => item.enabled)
                    .map((image) => (
                      <SelectOption key={image.id} value={image.id}>
                        {image.name} / {image.imageRef}
                      </SelectOption>
                    ))}
                </Select>
              </Field>
            </div>
          </Dialog>
          <Dialog
            open={agentDetailOpen}
            title={
              selectedAgent?.name ? `${selectedAgent.name} 详情` : "智能体详情"
            }
            onClose={() => setAgentDetailOpen(false)}
          >
            <div className="grid gap-4">
              {selectedAgent ? (
                <>
                  <Field label="ID">
                    <div className="text-sm text-foreground">
                      {selectedAgent.id}
                    </div>
                  </Field>
                  <Field label="名称">
                    <div className="text-sm text-foreground">
                      {selectedAgent.name}
                    </div>
                  </Field>
                  <Field label="状态">
                    <div className="text-sm text-foreground">
                      {activityStatusLabel(selectedAgent)}
                    </div>
                  </Field>
                  <Field label="运行时状态">
                    <div className="text-sm text-foreground">
                      {statusLabel(selectedAgent.runtimeStatus)}
                    </div>
                  </Field>
                  <Field label="供应商">
                    <div className="text-sm text-foreground">
                      {selectedAgent.providerName || "-"}
                    </div>
                  </Field>
                  <Field label="模型">
                    <div className="text-sm text-foreground">
                      {selectedAgent.modelName || "-"}
                    </div>
                  </Field>
                  <Field label="镜像">
                    <div className="text-sm text-foreground">
                      {selectedAgent.imageName || selectedAgent.imageRef || "-"}
                    </div>
                  </Field>
                  <Field label="类型">
                    <div className="text-sm text-foreground">
                      {agentTypeLabel(selectedAgent.agentType)}
                    </div>
                  </Field>
                  <Field label="备注">
                    <div className="text-sm leading-6 text-foreground">
                      {selectedAgent.notes || "-"}
                    </div>
                  </Field>
                </>
              ) : (
                <div className="text-sm text-muted-foreground">
                  未选择智能体
                </div>
              )}
            </div>
          </Dialog>
        </div>
        <WorkbenchSettingsDialog
          open={workbenchSettingsOpen}
          settings={workbenchSettings}
          onClose={() => setWorkbenchSettingsOpen(false)}
          onReset={() => setWorkbenchSettings(defaultWorkbenchSettings)}
          onSettingsChange={setWorkbenchSettings}
        />
      </TooltipProvider>
    </main>
  );
}

function AuthShell({ children }: { children: ReactNode }) {
  return (
    <div className="grid min-h-dvh place-items-center bg-background px-4 py-10 text-foreground">
      <div className="w-full max-w-sm">{children}</div>
    </div>
  );
}

function PortalLoginGate({
  themeMode,
  onLoginPathChange,
  onThemeToggle,
}: {
  themeMode: ThemeMode;
  onLoginPathChange: Dispatch<SetStateAction<string>>;
  onThemeToggle: () => void;
}) {
  function openLogin() {
    window.history.pushState(null, "", loginBrowserPath);
    onLoginPathChange(loginPath);
  }

  return (
    <AuthShell>
      <div
        className="grid gap-4 rounded-[8px] border border-border bg-card p-5 shadow-sm"
        data-testid="portal-auth-gate"
      >
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <div className="mb-2 grid h-10 w-10 place-items-center rounded-[6px] border border-primary/30 bg-primary text-primary-foreground">
              <Boxes className="h-5 w-5" />
            </div>
            <h1 className="text-xl font-semibold">Agent Box</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              登录后进入智能体工作台
            </p>
          </div>
          <IconButton title="切换浅色/深色样式" onClick={onThemeToggle}>
            {themeMode === "dark" ? (
              <Sun className="h-4 w-4" />
            ) : (
              <Moon className="h-4 w-4" />
            )}
          </IconButton>
        </div>
        <Button
          className="w-full"
          data-testid="portal-login-link"
          type="button"
          variant="primary"
          onClick={openLogin}
        >
          前往登录
        </Button>
      </div>
    </AuthShell>
  );
}

function LoginPage({
  themeMode,
  onLogin,
  onThemeToggle,
}: {
  themeMode: ThemeMode;
  onLogin: (username: string, password: string) => Promise<void>;
  onThemeToggle: () => void;
}) {
  const [username, setUsername] = useState("admin");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);

  async function submitLogin(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setSubmitting(true);
    try {
      await onLogin(username, password);
    } catch (error) {
      setError((error as Error).message || "登录失败");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <AuthShell>
      <div
        className="grid gap-4 rounded-[8px] border border-border bg-card p-5 shadow-sm"
        data-testid="login-page"
      >
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <div className="mb-2 grid h-10 w-10 place-items-center rounded-[6px] border border-primary/30 bg-primary text-primary-foreground">
              <KeyRound className="h-5 w-5" />
            </div>
            <h1 className="text-xl font-semibold">Agent Box</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              登录后进入智能体工作台
            </p>
          </div>
          <IconButton title="切换浅色/深色样式" onClick={onThemeToggle}>
            {themeMode === "dark" ? (
              <Sun className="h-4 w-4" />
            ) : (
              <Moon className="h-4 w-4" />
            )}
          </IconButton>
        </div>

        {error ? <Alert tone="danger">{error}</Alert> : null}

        <Form className="gap-3" onSubmit={(event) => void submitLogin(event)}>
          <Field label="用户名">
            <Input
              autoComplete="username"
              data-testid="login-username"
              disabled={submitting}
              value={username}
              onChange={(event) => setUsername(event.target.value)}
            />
          </Field>
          <Field label="密码">
            <Input
              autoComplete="current-password"
              data-testid="login-password"
              disabled={submitting}
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
            />
          </Field>
          <Button
            className="w-full"
            data-testid="login-submit"
            disabled={submitting}
            type="submit"
            variant="primary"
          >
            {submitting ? <Spinner label="正在登录" /> : "登录"}
          </Button>
        </Form>
      </div>
    </AuthShell>
  );
}

function NavButton({
  active,
  collapsed,
  children,
  label,
  testId,
  onClick,
}: {
  active: boolean;
  collapsed?: boolean;
  children: ReactNode;
  label: string;
  testId?: string;
  onClick: () => void;
}) {
  const className = cn(
    "inline-flex h-9 min-w-0 items-center gap-1 rounded-[6px] border border-border px-2 text-xs text-muted-foreground hover:bg-accent [&_svg]:shrink-0",
    active && "border-primary/30 bg-card text-foreground hover:bg-primary/10",
    collapsed ? "w-9 shrink-0 justify-center p-0" : "justify-center",
  );
  if (collapsed) {
    return (
      <Tooltip>
        <TooltipTrigger
          render={
            <Button
              aria-label={label}
              className={className}
              data-testid={testId}
              type="button"
              variant="ghost"
              onClick={onClick}
            />
          }
        >
          {children}
        </TooltipTrigger>
        <TooltipContent side="right">{label}</TooltipContent>
      </Tooltip>
    );
  }
  return (
    <Button
      className={className}
      data-testid={testId}
      type="button"
      variant="ghost"
      onClick={onClick}
    >
      {children}
    </Button>
  );
}

function CollapsedIconButton({
  children,
  className,
  label,
  onClick,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & {
  label: string;
}) {
  return (
    <Tooltip>
      <TooltipTrigger
        render={
          <IconButton
            {...props}
            aria-label={label}
            className={className}
            onClick={onClick}
          />
        }
      >
        {children}
      </TooltipTrigger>
      <TooltipContent side="right">{label}</TooltipContent>
    </Tooltip>
  );
}

function AgentContextMenu({
  agent,
  trigger,
  onOpenDetail,
  onStart,
  onStop,
  onEdit,
  onChangeImage,
  onDelete,
}: {
  agent: AgentInfo;
  trigger: ReactNode;
  onOpenDetail: (agent: AgentInfo) => void;
  onStart: (agent: AgentInfo) => void;
  onStop: (agent: AgentInfo) => void;
  onEdit: (agent: AgentInfo) => void;
  onChangeImage: (agent: AgentInfo) => void;
  onDelete: (agent: AgentInfo) => void;
}) {
  return (
    <ContextMenu>
      <ContextMenuTrigger className="contents">{trigger}</ContextMenuTrigger>
      <ContextMenuContent
        className="w-56"
        data-testid={`agent-context-menu-${agent.id}`}
      >
        <ContextMenuGroup>
          <ContextMenuLabel>{agent.name}</ContextMenuLabel>
        </ContextMenuGroup>
        <ContextMenuSeparator />
        <ContextMenuGroup>
          <ContextMenuItem
            disabled={!canStartAgent(agent)}
            onClick={() => onStart(agent)}
          >
            <Play />
            启动
          </ContextMenuItem>
          <ContextMenuItem
            disabled={!canStopAgent(agent)}
            onClick={() => onStop(agent)}
          >
            <Square />
            停止
          </ContextMenuItem>
        </ContextMenuGroup>
        <ContextMenuSeparator />
        <ContextMenuGroup>
          <ContextMenuItem onClick={() => onOpenDetail(agent)}>
            <Info />
            查看详情
          </ContextMenuItem>
          <ContextMenuItem onClick={() => onEdit(agent)}>
            <Pencil />
            编辑配置
          </ContextMenuItem>
          <ContextMenuItem onClick={() => onChangeImage(agent)}>
            <Repeat2 />
            更换镜像
          </ContextMenuItem>
          <ContextMenuItem
            variant="destructive"
            onClick={() => onDelete(agent)}
          >
            <Trash2 />
            删除
          </ContextMenuItem>
        </ContextMenuGroup>
      </ContextMenuContent>
    </ContextMenu>
  );
}

function AgentSidebarButton({
  active,
  agent,
  onClick,
}: {
  active: boolean;
  agent: AgentInfo;
  onClick: () => void;
}) {
  const iconKey = resolveAgentIconKey(agent.iconKey, agent.agentType);
  const AgentIcon = resolveAgentIcon(agent.iconKey, agent.agentType);
  return (
    <ListButton
      active={active}
      className={cn(
        "grid min-h-[50px] w-full grid-cols-[10px_32px_minmax(0,1fr)_auto] items-center gap-2 rounded-[6px] border border-border bg-card p-2 text-left hover:bg-accent",
        active && "border-primary/30 bg-primary/10 hover:bg-primary/15",
      )}
      data-agent-icon-key={iconKey}
      data-testid={`sidebar-agent-${agent.id}`}
      onClick={onClick}
    >
      <span className={cn("dot", activityDotClass(agent))} />
      <span
        aria-hidden="true"
        className={cn(
          "grid size-8 shrink-0 place-items-center rounded-[6px] border border-border bg-muted/50 text-muted-foreground",
          active && "border-primary/30 bg-primary/10 text-primary",
        )}
      >
        <AgentIcon className="size-5" />
      </span>
      <span className="min-w-0">
        <strong className="block truncate text-sm">{agent.name}</strong>
      </span>
      <StatusBadge agent={agent} />
    </ListButton>
  );
}

function CollapsedAgentButton({
  active,
  agent,
  onClick,
}: {
  active: boolean;
  agent: AgentInfo;
  onClick: () => void;
}) {
  const iconKey = resolveAgentIconKey(agent.iconKey, agent.agentType);
  const AgentIcon = resolveAgentIcon(agent.iconKey, agent.agentType);
  return (
    <Tooltip>
      <TooltipTrigger
        render={
          <IconButton
            aria-label={agent.name}
            className={cn(
              "relative size-9 shrink-0 text-muted-foreground hover:bg-accent hover:text-foreground",
              active &&
                "border-primary/30 bg-primary/10 text-primary hover:bg-primary/15 hover:text-primary",
            )}
            data-agent-icon-key={iconKey}
            data-testid={`collapsed-agent-${agent.id}`}
            onClick={onClick}
          />
        }
      >
        <AgentIcon className="size-5" />
        <span
          className={cn(
            "dot absolute bottom-1 right-1 border border-card",
            activityDotClass(agent),
          )}
        />
      </TooltipTrigger>
      <TooltipContent side="right">
        <span className="flex flex-col gap-0.5">
          <span className="font-medium">{agent.name}</span>
          <span className="text-[11px] opacity-80">
            {activityStatusLabel(agent)}
          </span>
        </span>
      </TooltipContent>
    </Tooltip>
  );
}

function ManagementCard({ children }: { children: ReactNode }) {
  return (
    <section className="grid min-h-0 grid-rows-[auto_minmax(0,1fr)_auto] overflow-hidden rounded-[8px] border border-border bg-card">
      {children}
    </section>
  );
}

function ManagementToolbar({
  query,
  pageSize,
  total,
  placeholder,
  onQuery,
  onPageSize,
}: {
  query: string;
  pageSize: number;
  total: number;
  placeholder: string;
  onQuery: (query: string) => void;
  onPageSize: (pageSize: number) => void;
}) {
  return (
    <div className="flex min-h-14 items-center justify-between gap-3 border-b border-border px-3 max-[720px]:flex-col max-[720px]:items-stretch">
      <SearchField
        className="w-full max-w-xl"
        placeholder={placeholder}
        value={query}
        onChange={(event) => onQuery(event.target.value)}
      />
      <div className="inline-flex items-center gap-2 text-xs text-muted-foreground">
        <span>共 {total} 项</span>
        <Select
          className="w-28"
          value={pageSize}
          onChange={(event) => onPageSize(Number(event.target.value))}
        >
          {pageSizeOptions.map((size) => (
            <SelectOption key={size} value={size}>
              每页 {size}
            </SelectOption>
          ))}
        </Select>
      </div>
    </div>
  );
}

function AgentTable(props: {
  agents: AgentInfo[];
  selectedAgent?: AgentInfo;
  onSelect: (agent: AgentInfo) => void;
  onStart: (agent: AgentInfo) => void;
  onStop: (agent: AgentInfo) => void;
  onEdit: (agent: AgentInfo) => void;
  onChangeImage: (agent: AgentInfo) => void;
  onDelete: (agent: AgentInfo) => void;
}) {
  const columns = useMemo<ColumnDef<AgentInfo>[]>(
    () => [
      {
        accessorKey: "name",
        header: "名称",
        cell: ({ row }) => (
          <CellTitle
            title={row.original.name}
            subtitle={row.original.notes || row.original.id}
          />
        ),
      },
      {
        id: "status",
        header: "状态",
        enableSorting: false,
        cell: ({ row }) => <StatusBadge agent={row.original} />,
      },
      { accessorKey: "providerName", header: "供应商" },
      { accessorKey: "modelName", header: "模型" },
      {
        id: "image",
        header: "镜像",
        cell: ({ row }) => row.original.imageName || row.original.imageRef,
      },
      {
        id: "agentType",
        header: "类型",
        cell: ({ row }) => agentTypeLabel(row.original.agentType),
      },
      {
        id: "actions",
        header: "操作",
        enableSorting: false,
        cell: ({ row }) => (
          <RowActions testId="agent-row-actions">
            <IconButton
              disabled={!canStartAgent(row.original)}
              title="启动"
              onClick={(e) => {
                e.stopPropagation();
                props.onStart(row.original);
              }}
            >
              <Play className="h-4 w-4" />
            </IconButton>
            <IconButton
              disabled={!canStopAgent(row.original)}
              title="停止"
              onClick={(e) => {
                e.stopPropagation();
                props.onStop(row.original);
              }}
            >
              <Square className="h-4 w-4" />
            </IconButton>
            <Dropdown
              trigger={
                <IconButton title="更多操作">
                  <MoreHorizontal className="h-4 w-4" />
                </IconButton>
              }
            >
              <DropdownItem onClick={() => props.onEdit(row.original)}>
                <Pencil className="h-4 w-4" />
                修改
              </DropdownItem>
              <DropdownItem onClick={() => props.onChangeImage(row.original)}>
                <Repeat2 className="h-4 w-4" />
                更换镜像
              </DropdownItem>
              <DropdownItem danger onClick={() => props.onDelete(row.original)}>
                <Trash2 className="h-4 w-4" />
                删除
              </DropdownItem>
            </Dropdown>
          </RowActions>
        ),
      },
    ],
    [
      props.onStart,
      props.onStop,
      props.onEdit,
      props.onChangeImage,
      props.onDelete,
    ],
  );

  return (
    <DataTable<AgentInfo>
      columns={columns}
      data={props.agents}
      getRowId={(row) => row.id}
      onRowClick={(row) => props.onSelect(row.original)}
      selectedRowId={props.selectedAgent?.id}
      emptyText="暂无匹配的智能体"
    />
  );
}

function ProviderTable({
  providers,
  onEdit,
  onAddModel,
  onSync,
  onDelete,
  onDeleteModel,
}: {
  providers: ProviderInfo[];
  onEdit: (provider: ProviderInfo) => void;
  onAddModel: (provider: ProviderInfo) => void;
  onSync: (provider: ProviderInfo) => void;
  onDelete: (provider: ProviderInfo) => void;
  onDeleteModel: (provider: ProviderInfo, model: ProviderModel) => void;
}) {
  const columns = useMemo<ColumnDef<ProviderInfo>[]>(
    () => [
      {
        accessorKey: "name",
        header: "名称",
        cell: ({ row }) => (
          <CellTitle
            title={row.original.name}
            subtitle={row.original.homepageUrl || "未配置主页"}
          />
        ),
      },
      {
        id: "endpoint",
        header: "端点",
        cell: ({ row }) => (
          <CellTitle
            title={
              row.original.anthropicBaseUrl || row.original.openaiBaseUrl || "-"
            }
            subtitle={
              row.original.openaiBaseUrl && row.original.anthropicBaseUrl
                ? row.original.openaiBaseUrl
                : ""
            }
          />
        ),
      },
      {
        id: "apiKey",
        header: "密钥",
        cell: ({ row }) =>
          row.original.apiKeyConfigured ? row.original.apiKeyMasked : "未配置",
      },
      {
        id: "models",
        header: "模型",
        enableSorting: false,
        cell: ({ row }) => (
          <div className="flex max-w-md flex-wrap gap-1.5">
            {(row.original.models?.length ?? 0) === 0 ? (
              <span className="text-muted-foreground">暂无模型</span>
            ) : null}
            {row.original.models?.map((model) => (
              <span
                key={model.id}
                className="inline-flex max-w-full items-center gap-1 rounded-full border border-border bg-muted/50 px-2 py-1 text-xs"
              >
                {model.name}
                <small className="text-muted-foreground">
                  {protocolLabel(model.protocol)}
                </small>
                <IconButton
                  className="h-5 w-5 border-0 bg-transparent"
                  title="删除模型"
                  onClick={() => onDeleteModel(row.original, model)}
                >
                  <X className="h-3 w-3" />
                </IconButton>
              </span>
            ))}
          </div>
        ),
      },
      {
        accessorKey: "notes",
        header: "备注",
        cell: ({ row }) => row.original.notes || "-",
      },
      {
        id: "actions",
        header: "操作",
        enableSorting: false,
        cell: ({ row }) => (
          <RowActions testId="provider-row-actions">
            <IconButton title="修改供应商" onClick={() => onEdit(row.original)}>
              <Pencil className="h-4 w-4" />
            </IconButton>
            <Dropdown
              trigger={
                <IconButton title="更多操作">
                  <MoreHorizontal className="h-4 w-4" />
                </IconButton>
              }
            >
              <DropdownItem onClick={() => onAddModel(row.original)}>
                <Plus className="h-4 w-4" />
                新增模型
              </DropdownItem>
              <DropdownItem onClick={() => onSync(row.original)}>
                <RefreshCw className="h-4 w-4" />
                同步模型
              </DropdownItem>
              <DropdownItem danger onClick={() => onDelete(row.original)}>
                <Trash2 className="h-4 w-4" />
                删除
              </DropdownItem>
            </Dropdown>
          </RowActions>
        ),
      },
    ],
    [onEdit, onAddModel, onSync, onDelete, onDeleteModel],
  );

  return (
    <DataTable<ProviderInfo>
      columns={columns}
      data={providers}
      getRowId={(row) => String(row.id)}
      emptyText="暂无匹配的供应商"
    />
  );
}

function ImageTable({
  images,
  onEdit,
  onDelete,
}: {
  images: CodingImageInfo[];
  onEdit: (image: CodingImageInfo) => void;
  onDelete: (image: CodingImageInfo) => void;
}) {
  const isMobile = useMediaQuery("(max-width: 768px)");

  const columns = useMemo<ColumnDef<CodingImageInfo>[]>(
    () => [
      {
        accessorKey: "name",
        header: "名称",
        size: 200,
        minSize: 150,
        cell: ({ row }) => (
          <CellTitle
            title={row.original.name}
            subtitle={
              row.original.isDefault ? "默认镜像" : `ID ${row.original.id}`
            }
          />
        ),
      },
      {
        accessorKey: "imageRef",
        header: "镜像地址",
        size: 250,
        minSize: 180,
      },
      {
        id: "agentType",
        header: "类型",
        size: 100,
        cell: ({ row }) => agentTypeLabel(row.original.agentType),
      },
      {
        accessorKey: "defaultShell",
        header: "默认终端",
        size: 100,
      },
      {
        id: "enabled",
        header: "状态",
        size: 80,
        cell: ({ row }) => (
          <Badge tone={row.original.enabled ? "success" : "neutral"}>
            {row.original.enabled ? "启用" : "停用"}
          </Badge>
        ),
      },
      {
        accessorKey: "notes",
        header: "备注",
        size: 200,
        minSize: 120,
        cell: ({ row }) => {
          const notes = row.original.notes;
          if (!notes) return "-";
          return (
            <Tooltip>
              <TooltipTrigger
                render={
                  <span className="block max-w-[200px] cursor-default truncate" />
                }
              >
                {notes}
              </TooltipTrigger>
              <TooltipContent
                className="max-w-xs whitespace-pre-wrap"
                side="top"
              >
                {notes}
              </TooltipContent>
            </Tooltip>
          );
        },
      },
      {
        id: "actions",
        header: "操作",
        size: 100,
        enableSorting: false,
        cell: ({ row }) => (
          <RowActions testId="image-row-actions">
            <IconButton title="修改镜像" onClick={() => onEdit(row.original)}>
              <Pencil className="h-4 w-4" />
            </IconButton>
            <IconButton
              className="text-destructive"
              disabled={row.original.isDefault}
              title="删除镜像"
              onClick={() => onDelete(row.original)}
            >
              <Trash2 className="h-4 w-4" />
            </IconButton>
          </RowActions>
        ),
      },
    ],
    [onEdit, onDelete],
  );

  const columnVisibility = useMemo(
    () => ({
      defaultShell: !isMobile,
    }),
    [isMobile],
  );

  return (
    <DataTable<CodingImageInfo>
      columns={columns}
      data={images}
      getRowId={(row) => String(row.id)}
      emptyText="暂无匹配的镜像"
      columnVisibility={columnVisibility}
    />
  );
}

function CellTitle({ title, subtitle }: { title: string; subtitle?: string }) {
  return (
    <div className="grid min-w-0 gap-1">
      <strong className="truncate text-sm">{title}</strong>
      {subtitle ? (
        <span className="truncate text-xs text-muted-foreground">
          {subtitle}
        </span>
      ) : null}
    </div>
  );
}

function RowActions({
  children,
  testId,
}: {
  children: ReactNode;
  testId?: string;
}) {
  return (
    <div
      className="flex w-full items-center justify-center gap-1"
      data-testid={testId}
    >
      {children}
    </div>
  );
}

function Metric({
  icon,
  label,
  value,
}: {
  icon: ReactNode;
  label: string;
  value: string;
}) {
  return (
    <div className="grid min-w-0 grid-cols-[44px_minmax(0,1fr)] grid-rows-2 gap-x-3 rounded-[10px] bg-card p-5 shadow-sm">
      <div className="row-span-2 grid h-10 w-10 place-items-center rounded-[10px] bg-chart-3/10 text-chart-3">
        {icon}
      </div>
      <span className="text-xs text-muted-foreground">{label}</span>
      <strong className="truncate text-sm font-semibold" title={value}>
        {value}
      </strong>
    </div>
  );
}

function AgentDialog(props: {
  dialog: { open: boolean; mode: EntityMode };
  form: AgentForm;
  providers: ProviderInfo[];
  providerModels: ProviderModel[];
  images: CodingImageInfo[];
  setForm: Dispatch<SetStateAction<AgentForm>>;
  onSubmit: () => void;
  onClose: () => void;
}) {
  const selectedImage = props.images.find(
    (image) => image.id === Number(props.form.imageId),
  );
  const selectedProvider = props.providers.find(
    (provider) => provider.id === Number(props.form.providerId),
  );
  const displayAgentType = selectedImage?.agentType ?? props.form.agentType;
  const fixedModelProtocol = singleProviderProtocol(selectedProvider);
  const selectableModels = props.providerModels.filter(
    (model) => model.protocol === props.form.modelProtocol,
  );

  function setImage(imageId: number) {
    const image = props.images.find((item) => item.id === imageId);
    props.setForm((current) => ({
      ...current,
      imageId,
      agentType: image?.agentType ?? current.agentType,
    }));
  }

  function setProvider(providerId: number) {
    const provider = props.providers.find((item) => item.id === providerId);
    props.setForm((current) =>
      normalizeAgentFormForProvider({ ...current, providerId }, provider),
    );
  }

  function setModelProtocol(modelProtocol: ModelProtocol) {
    props.setForm((current) => ({
      ...current,
      modelProtocol,
      modelName: modelNameForProtocol(
        props.providerModels,
        modelProtocol,
        current.modelName,
      ),
    }));
  }

  return (
    <Dialog
      footer={<FormActions formId="agent-form" onClose={props.onClose} />}
      open={props.dialog.open}
      title={props.dialog.mode === "edit" ? "修改智能体" : "新增智能体"}
      onClose={props.onClose}
    >
      <Form
        id="agent-form"
        onSubmit={(event) => {
          event.preventDefault();
          props.onSubmit();
        }}
      >
        <Field label="名称">
          <Input
            required
            placeholder="例如 Claude 工作台"
            value={props.form.name}
            onChange={(event) =>
              props.setForm((current) => ({
                ...current,
                name: event.target.value,
              }))
            }
          />
        </Field>
        <Field label="供应商">
          <Select
            disabled={props.providers.length === 0}
            required
            value={props.form.providerId || ""}
            onChange={(event) => setProvider(Number(event.target.value))}
          >
            <SelectOption value="" disabled>
              {props.providers.length === 0 ? "暂无供应商" : "选择供应商"}
            </SelectOption>
            {props.providers.map((provider) => (
              <SelectOption key={provider.id} value={provider.id}>
                {provider.name}
              </SelectOption>
            ))}
          </Select>
        </Field>
        {fixedModelProtocol ? (
          <Field label="模型协议">
            <div className="min-h-9 rounded-lg border border-input bg-muted/30 px-3 py-2 text-sm text-foreground">
              {protocolLabel(fixedModelProtocol)}
            </div>
          </Field>
        ) : (
          <Field label="模型协议">
            <Select
              value={props.form.modelProtocol}
              onChange={(event) =>
                setModelProtocol(event.target.value as ModelProtocol)
              }
            >
              <SelectOption value="anthropic">Anthropic</SelectOption>
              <SelectOption value="openai">OpenAI</SelectOption>
            </Select>
          </Field>
        )}
        <Field label="模型">
          {selectableModels.length > 0 ? (
            <Select
              required
              value={props.form.modelName}
              onChange={(event) =>
                props.setForm((current) => ({
                  ...current,
                  modelName: event.target.value,
                }))
              }
            >
              {selectableModels.map((model) => (
                <SelectOption key={model.id} value={model.name}>
                  {model.name} / {protocolLabel(model.protocol)}
                </SelectOption>
              ))}
            </Select>
          ) : (
            <Input
              required
              placeholder="模型名称"
              value={props.form.modelName}
              onChange={(event) =>
                props.setForm((current) => ({
                  ...current,
                  modelName: event.target.value,
                }))
              }
            />
          )}
        </Field>
        {props.dialog.mode === "create" ? (
          <Field label="镜像">
            <Select
              required
              value={props.form.imageId}
              onChange={(event) => setImage(Number(event.target.value))}
            >
              {props.images
                .filter((item) => item.enabled)
                .map((image) => (
                  <SelectOption key={image.id} value={image.id}>
                    {image.name} / {image.imageRef}
                  </SelectOption>
                ))}
            </Select>
          </Field>
        ) : null}
        <Field label="智能体类型">
          <div className="text-sm text-foreground">
            {agentTypeLabel(displayAgentType)}
          </div>
        </Field>
        <Field label="入口图标">
          <AgentIconPicker
            agentType={displayAgentType}
            value={props.form.iconKey}
            onChange={(iconKey) =>
              props.setForm((current) => ({ ...current, iconKey }))
            }
          />
        </Field>
        <Field label="备注">
          <Textarea
            placeholder="用途、负责人或运行说明"
            value={props.form.notes}
            onChange={(event) =>
              props.setForm((current) => ({
                ...current,
                notes: event.target.value,
              }))
            }
          />
        </Field>
      </Form>
    </Dialog>
  );
}

function AgentIconPicker({
  agentType,
  value,
  onChange,
}: {
  agentType: string;
  value: string;
  onChange: (iconKey: string) => void;
}) {
  const DefaultIcon = resolveAgentIcon("", agentType);
  const [activeCategory, setActiveCategory] = useState<
    AgentIconCategoryId | "all"
  >("all");
  const selectedKey = value || "default";

  const filteredIconOptions = useMemo(() => {
    return agentIconOptions.filter((option) => {
      return activeCategory === "all" || option.category === activeCategory;
    });
  }, [activeCategory]);

  const renderIconButton = (
    option: AgentIconOption,
    testId: string,
    compact = false,
  ) => {
    const active = value === option.key;
    const Icon = option.Icon;
    return (
      <Button
        aria-label={option.label}
        aria-pressed={active}
        className={cn(
          active && "border-primary/30 bg-primary/10 text-primary",
          active && compact && "hover:bg-primary/10",
        )}
        data-agent-icon-key={option.key}
        data-testid={testId}
        key={testId}
        size={compact ? "icon-sm" : "icon"}
        title={option.label}
        type="button"
        variant={active ? "outline" : "ghost"}
        onClick={() => onChange(option.key)}
      >
        <Icon />
      </Button>
    );
  };

  return (
    <div
      className="flex flex-col gap-2 rounded-md border border-border bg-muted/20 p-2"
      role="group"
      aria-label="入口图标"
    >
      <div className="flex items-start gap-2">
        <div className="flex flex-col items-center gap-1">
          <Button
            aria-label="默认入口图标"
            aria-pressed={selectedKey === "default"}
            className={cn(
              selectedKey === "default" &&
                "border-primary/30 bg-primary/10 text-primary",
            )}
            data-agent-icon-key="default"
            data-testid="agent-icon-option-default"
            size="icon"
            title="默认"
            type="button"
            variant={selectedKey === "default" ? "outline" : "ghost"}
            onClick={() => onChange("")}
          >
            <DefaultIcon />
          </Button>
          <span className="text-xs text-muted-foreground">默认</span>
        </div>
        <div className="flex min-w-0 flex-1 flex-col gap-1">
          <div className="text-xs font-medium text-muted-foreground">推荐</div>
          <div className="flex flex-wrap gap-1">
            {recommendedAgentIconOptions.map((option) =>
              renderIconButton(
                option,
                `agent-icon-recommended-${option.key}`,
                true,
              ),
            )}
          </div>
        </div>
      </div>
      <div className="flex flex-wrap gap-1" aria-label="图标分类">
        <Button
          aria-pressed={activeCategory === "all"}
          data-testid="agent-icon-category-all"
          size="xs"
          type="button"
          variant={activeCategory === "all" ? "outline" : "ghost"}
          onClick={() => setActiveCategory("all")}
        >
          全部
        </Button>
        {agentIconCategories.map((category) => (
          <Button
            aria-pressed={activeCategory === category.id}
            data-testid={`agent-icon-category-${category.id}`}
            key={category.id}
            size="xs"
            type="button"
            variant={activeCategory === category.id ? "outline" : "ghost"}
            onClick={() => setActiveCategory(category.id)}
          >
            {category.label}
          </Button>
        ))}
      </div>
      <div className="grid max-h-40 grid-cols-[repeat(auto-fill,minmax(2rem,1fr))] gap-1 overflow-y-auto rounded-md border border-border bg-background p-1">
        {filteredIconOptions.map((option) =>
          renderIconButton(option, `agent-icon-option-${option.key}`),
        )}
      </div>
    </div>
  );
}

function ProviderDialog(props: {
  dialog: { open: boolean; mode: EntityMode };
  form: ProviderForm;
  setForm: Dispatch<SetStateAction<ProviderForm>>;
  onSubmit: () => void;
  onClose: () => void;
}) {
  return (
    <Dialog
      footer={<FormActions formId="provider-form" onClose={props.onClose} />}
      open={props.dialog.open}
      title={props.dialog.mode === "edit" ? "修改供应商" : "新增供应商"}
      onClose={props.onClose}
    >
      <Form
        id="provider-form"
        onSubmit={(event) => {
          event.preventDefault();
          props.onSubmit();
        }}
      >
        <Field label="名称">
          <Input
            required
            placeholder="例如 Anthropic"
            value={props.form.name}
            onChange={(event) =>
              props.setForm((current) => ({
                ...current,
                name: event.target.value,
              }))
            }
          />
        </Field>
        <Field label="主页">
          <Input
            placeholder="https://example.com"
            value={props.form.homepageUrl}
            onChange={(event) =>
              props.setForm((current) => ({
                ...current,
                homepageUrl: event.target.value,
              }))
            }
          />
        </Field>
        <Field label="API 密钥">
          <Input
            type="password"
            placeholder={
              props.dialog.mode === "edit"
                ? "留空则保持原密钥"
                : "输入 API 密钥"
            }
            value={props.form.apiKey}
            onChange={(event) =>
              props.setForm((current) => ({
                ...current,
                apiKey: event.target.value,
              }))
            }
          />
        </Field>
        <Field label="OpenAI 接入地址">
          <Input
            placeholder="https://api.openai.com/v1"
            value={props.form.openaiBaseUrl}
            onChange={(event) =>
              props.setForm((current) => ({
                ...current,
                openaiBaseUrl: event.target.value,
              }))
            }
          />
        </Field>
        <Field label="Anthropic 接入地址">
          <Input
            placeholder="https://api.anthropic.com/v1"
            value={props.form.anthropicBaseUrl}
            onChange={(event) =>
              props.setForm((current) => ({
                ...current,
                anthropicBaseUrl: event.target.value,
              }))
            }
          />
        </Field>
        <Field label="备注">
          <Textarea
            placeholder="供应商说明"
            value={props.form.notes}
            onChange={(event) =>
              props.setForm((current) => ({
                ...current,
                notes: event.target.value,
              }))
            }
          />
        </Field>
      </Form>
    </Dialog>
  );
}

function ModelDialog(props: {
  open: boolean;
  form: ModelForm;
  providers: ProviderInfo[];
  setForm: Dispatch<SetStateAction<ModelForm>>;
  onSubmit: () => void;
  onSync: () => void;
  onClose: () => void;
}) {
  return (
    <Dialog
      footer={
        <>
          <Button type="button" variant="soft" onClick={props.onSync}>
            同步模型
          </Button>
          <FormActions formId="model-form" onClose={props.onClose} />
        </>
      }
      open={props.open}
      title="新增模型"
      onClose={props.onClose}
    >
      <Form
        id="model-form"
        onSubmit={(event) => {
          event.preventDefault();
          props.onSubmit();
        }}
      >
        <Field label="供应商">
          <Select
            required
            value={props.form.providerId}
            onChange={(event) =>
              props.setForm((current) => ({
                ...current,
                providerId: Number(event.target.value),
              }))
            }
          >
            {props.providers.map((provider) => (
              <SelectOption key={provider.id} value={provider.id}>
                {provider.name}
              </SelectOption>
            ))}
          </Select>
        </Field>
        <Field label="协议">
          <Select
            value={props.form.protocol}
            onChange={(event) =>
              props.setForm((current) => ({
                ...current,
                protocol: event.target.value as ModelForm["protocol"],
              }))
            }
          >
            <SelectOption value="anthropic">Anthropic</SelectOption>
            <SelectOption value="openai">OpenAI</SelectOption>
          </Select>
        </Field>
        <Field label="模型名称">
          <Input
            required
            placeholder="例如 claude-sonnet-4-5"
            value={props.form.name}
            onChange={(event) =>
              props.setForm((current) => ({
                ...current,
                name: event.target.value,
              }))
            }
          />
        </Field>
      </Form>
    </Dialog>
  );
}

function ImageDialog(props: {
  dialog: { open: boolean; mode: EntityMode };
  form: ImageForm;
  setForm: Dispatch<SetStateAction<ImageForm>>;
  onSubmit: () => void;
  onClose: () => void;
}) {
  return (
    <Dialog
      footer={<FormActions formId="image-form" onClose={props.onClose} />}
      open={props.dialog.open}
      title={props.dialog.mode === "edit" ? "修改镜像" : "新增镜像"}
      onClose={props.onClose}
    >
      <Form
        id="image-form"
        onSubmit={(event) => {
          event.preventDefault();
          props.onSubmit();
        }}
      >
        <Field label="名称">
          <Input
            required
            placeholder="例如 Claude Code 基础镜像"
            value={props.form.name}
            onChange={(event) =>
              props.setForm((current) => ({
                ...current,
                name: event.target.value,
              }))
            }
          />
        </Field>
        <Field label="镜像地址">
          <Input
            required
            placeholder="loads/cc:latest"
            value={props.form.imageRef}
            onChange={(event) =>
              props.setForm((current) => ({
                ...current,
                imageRef: event.target.value,
              }))
            }
          />
        </Field>
        <Field label="智能体类型">
          <Select
            value={props.form.agentType}
            onChange={(event) =>
              props.setForm((current) => ({
                ...current,
                agentType: event.target.value as ImageForm["agentType"],
              }))
            }
          >
            <SelectOption value="claude_code">Claude Code</SelectOption>
            <SelectOption value="codex">Codex</SelectOption>
            <SelectOption value="custom">自定义</SelectOption>
          </Select>
        </Field>
        <Field label="默认终端">
          <Input
            required
            placeholder="/bin/bash"
            value={props.form.defaultShell}
            onChange={(event) =>
              props.setForm((current) => ({
                ...current,
                defaultShell: event.target.value,
              }))
            }
          />
        </Field>
        <Field label="备注">
          <Textarea
            placeholder="镜像说明"
            value={props.form.notes}
            onChange={(event) =>
              props.setForm((current) => ({
                ...current,
                notes: event.target.value,
              }))
            }
          />
        </Field>
        <CheckboxField
          checked={props.form.enabled}
          onChange={(event) =>
            props.setForm((current) => ({
              ...current,
              enabled: event.target.checked,
            }))
          }
        >
          启用镜像
        </CheckboxField>
      </Form>
    </Dialog>
  );
}

function FormActions({
  formId,
  onClose,
}: {
  formId: string;
  onClose: () => void;
}) {
  return (
    <>
      <Button type="button" variant="soft" onClick={onClose}>
        取消
      </Button>
      <Button form={formId} type="submit" variant="primary">
        <Check className="h-4 w-4" />
        保存
      </Button>
    </>
  );
}

function StatusBadge({ agent }: { agent?: AgentInfo }) {
  const status = activityStatus(agent);
  const tone =
    status === "working" ||
    status === "waiting_input" ||
    status === "starting" ||
    status === "stopping" ||
    status === "deleting"
      ? "warning"
      : status === "running"
        ? "success"
        : "neutral";
  return <Badge tone={tone}>{activityStatusLabel(agent)}</Badge>;
}

function activityStatus(agent: AgentInfo | undefined) {
  if (!agent) {
    return "";
  }
  if (agent.activityStatus) {
    return agent.activityStatus;
  }
  if (agent.runtimeStatus === "running") {
    return "running";
  }
  if (agent.runtimeStatus === "stopped") {
    return "stopped";
  }
  if (!agent.runtimeStatus || agent.runtimeStatus === "created") {
    return "idle";
  }
  return agent.runtimeStatus;
}

function activityStatusLabel(agent: AgentInfo | undefined) {
  const value = activityStatus(agent);
  if (value === "waiting_input") return "待确认";
  if (value === "working") return "工作中";
  if (value === "running") return "运行中";
  return statusLabel(value);
}

function agentStatusText(
  agent: AgentInfo | null | undefined,
  statuses: AgentUiStatusMap,
) {
  if (!agent) {
    return statusLabel("");
  }
  return statusLabel(statuses[agent.id] || activityStatus(agent));
}

function activityDotClass(agent: AgentInfo | undefined) {
  const status = activityStatus(agent);
  if (
    status === "working" ||
    status === "waiting_input" ||
    status === "starting" ||
    status === "stopping" ||
    status === "deleting"
  )
    return "working";
  return status === "running" ? "online" : "offline";
}

function statusLabel(value: string | undefined) {
  const labels: Record<string, string> = {
    idle: "空闲",
    running: "运行中",
    stopped: "已停止",
    stopping: "停止中",
    deleting: "删除中",
    deleted: "已删除",
    exited: "已退出",
    paused: "已暂停",
    missing: "容器离线",
    unavailable: "Docker 不可用",
    dead: "异常退出",
    created: "未启动",
    starting: "启动中",
    error: "异常",
    waiting_input: "待确认",
    working: "工作中",
  };
  return labels[value || ""] ?? (value || "空闲");
}

function withAgentUiStatus(
  agent: AgentInfo,
  statuses: AgentUiStatusMap,
): AgentInfo {
  const status = statuses[agent.id];
  if (!status) {
    return agent;
  }
  return {
    ...agent,
    activityStatus: status,
  };
}

function canStartAgent(agent: AgentInfo | undefined) {
  if (!agent) {
    return false;
  }
  const status = activityStatus(agent);
  return (
    status !== "running" &&
    status !== "working" &&
    status !== "waiting_input" &&
    status !== "starting" &&
    status !== "stopping" &&
    status !== "deleting"
  );
}

function canStopAgent(agent: AgentInfo | undefined) {
  if (!agent) {
    return false;
  }
  const status = activityStatus(agent);
  return (
    status === "running" || status === "working" || status === "waiting_input"
  );
}

function gitDirtyStateKey(agentId: string, workspacePath: string) {
  return agentId && workspacePath ? `${agentId}\u0000${workspacePath}` : "";
}

function SourceControlCodicon({ className }: { className?: string }) {
  return (
    <span
      aria-hidden="true"
      className={cn("codicon codicon-source-control", className)}
    />
  );
}

function canChatOverrideAgentStatus(status: AgentUiStatus) {
  return status === "working" || status === "waiting_input";
}

function isTransientAgentUiStatus(status: AgentUiStatus) {
  return (
    status === "starting" || status === "stopping" || status === "deleting"
  );
}

function agentTypeLabel(value: string | undefined) {
  const labels: Record<string, string> = {
    claude_code: "Claude Code",
    codex: "Codex",
    custom: "自定义",
  };
  return labels[value || ""] ?? (value || "-");
}

function protocolLabel(value: string | undefined) {
  return value === "openai" ? "OpenAI" : "Anthropic";
}

function filterItems<T>(
  items: T[],
  query: string,
  fields: (item: T) => Array<string | number | boolean | undefined>,
) {
  const normalizedQuery = query.trim().toLowerCase();
  if (!normalizedQuery) return items;
  return items.filter((item) =>
    fields(item)
      .filter((value) => value !== undefined)
      .some((value) => String(value).toLowerCase().includes(normalizedQuery)),
  );
}

function paginate<T>(items: T[], pager: PagerState) {
  const current = Math.min(
    pager.page,
    totalPages(items.length, pager.pageSize),
  );
  const start = (current - 1) * pager.pageSize;
  return items.slice(start, start + pager.pageSize);
}

function pageInfo(total: number, pager: PagerState) {
  const pages = totalPages(total, pager.pageSize);
  const current = Math.min(pager.page, pages);
  const start = total === 0 ? 0 : (current - 1) * pager.pageSize + 1;
  const end = Math.min(total, current * pager.pageSize);
  return { total, pages, current, start, end };
}

function totalPages(total: number, pageSize: number) {
  return Math.max(1, Math.ceil(total / pageSize));
}

function loadPersistedWorkbenchState(): PersistedWorkbenchState {
  if (typeof window === "undefined") {
    return defaultPersistedWorkbenchState();
  }

  try {
    const rawState = window.localStorage.getItem(persistedWorkbenchStorageKey);
    if (!rawState) {
      return defaultPersistedWorkbenchState();
    }
    const parsed = JSON.parse(rawState) as Record<string, unknown>;
    return {
      viewMode: isViewMode(parsed.viewMode) ? parsed.viewMode : "agents",
      selectedAgentId:
        typeof parsed.selectedAgentId === "string"
          ? parsed.selectedAgentId
          : "",
      agentWorkbenchState: parsePersistedAgentWorkbenchState(
        parsed.agentWorkbenchState,
      ),
    };
  } catch {
    return defaultPersistedWorkbenchState();
  }
}

function savePersistedWorkbenchState(state: PersistedWorkbenchState) {
  if (typeof window === "undefined") {
    return;
  }

  try {
    window.localStorage.setItem(
      persistedWorkbenchStorageKey,
      JSON.stringify(state),
    );
  } catch {
    return;
  }
}

function defaultPersistedWorkbenchState(): PersistedWorkbenchState {
  return {
    viewMode: "agents",
    selectedAgentId: "",
    agentWorkbenchState: {},
  };
}

function parsePersistedAgentWorkbenchState(
  value: unknown,
): Record<string, AgentWorkbenchState> {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return {};
  }

  return Object.entries(value as Record<string, unknown>).reduce<
    Record<string, AgentWorkbenchState>
  >((result, [agentId, state]) => {
    if (
      !agentId ||
      !state ||
      typeof state !== "object" ||
      Array.isArray(state)
    ) {
      return result;
    }
    const candidate = state as Record<string, unknown>;
    result[agentId] = {
      activeChatSessionId:
        typeof candidate.activeChatSessionId === "string"
          ? candidate.activeChatSessionId
          : defaultWorkbenchState.activeChatSessionId,
      chatHistoryOpen:
        typeof candidate.chatHistoryOpen === "boolean"
          ? candidate.chatHistoryOpen
          : defaultWorkbenchState.chatHistoryOpen,
      panelMode: isPanelMode(candidate.panelMode)
        ? candidate.panelMode
        : defaultWorkbenchState.panelMode,
      workspacePath:
        typeof candidate.workspacePath === "string" && candidate.workspacePath
          ? normalizeWorkspacePath(candidate.workspacePath)
          : defaultWorkbenchState.workspacePath,
    };
    return result;
  }, {});
}

function isPanelMode(value: unknown): value is PanelMode {
  return typeof value === "string" && panelModes.includes(value as PanelMode);
}

function isViewMode(value: unknown): value is ViewMode {
  return typeof value === "string" && viewModes.includes(value as ViewMode);
}

function blankProviderForm(): ProviderForm {
  return {
    id: 0,
    name: "",
    homepageUrl: "",
    notes: "",
    apiKey: "",
    openaiBaseUrl: "",
    anthropicBaseUrl: "",
  };
}

function providerToForm(provider: ProviderInfo): ProviderForm {
  return {
    id: provider.id,
    name: provider.name,
    homepageUrl: provider.homepageUrl,
    notes: provider.notes,
    apiKey: "",
    openaiBaseUrl: provider.openaiBaseUrl,
    anthropicBaseUrl: provider.anthropicBaseUrl,
  };
}

function blankModelForm(providerId = 0): ModelForm {
  return { providerId, name: "claude-sonnet-4-5", protocol: "anthropic" };
}

function blankImageForm(): ImageForm {
  return {
    id: 0,
    name: "",
    imageRef: "",
    agentType: "claude_code",
    defaultShell: "/bin/bash",
    notes: "",
    enabled: true,
  };
}

function imageToForm(image: CodingImageInfo): ImageForm {
  return {
    id: image.id,
    name: image.name,
    imageRef: image.imageRef,
    agentType: image.agentType,
    defaultShell: image.defaultShell,
    notes: image.notes,
    enabled: image.enabled,
  };
}

function blankAgentForm(): AgentForm {
  return {
    id: "",
    name: "",
    providerId: 0,
    modelName: "",
    modelProtocol: "anthropic",
    imageId: 0,
    agentType: "claude_code",
    iconKey: "",
    notes: "",
  };
}

function defaultAgentForm(
  providers: ProviderInfo[],
  images: CodingImageInfo[],
): AgentForm {
  const provider = providers[0];
  const model = provider?.models?.[0];
  const image =
    images.find((item) => item.enabled && item.agentType === "claude_code") ??
    images[0];
  return normalizeAgentFormForProvider(
    {
      id: "",
      name: "",
      providerId: provider?.id ?? 0,
      modelName: model?.name ?? "",
      modelProtocol: model?.protocol ?? "anthropic",
      imageId: image?.id ?? 0,
      agentType: image?.agentType ?? "claude_code",
      iconKey: "",
      notes: "",
    },
    provider,
  );
}

function agentToForm(agent: AgentInfo): AgentForm {
  return {
    id: agent.id,
    name: agent.name,
    providerId: agent.providerId,
    modelName: agent.modelName,
    modelProtocol: agent.modelProtocol,
    imageId: agent.imageId,
    agentType: agent.agentType,
    iconKey: agent.iconKey ?? "",
    notes: agent.notes,
  };
}

function blankDeleteDialog(): DeleteDialog {
  return {
    open: false,
    kind: "agent",
    title: "",
    description: "",
    targetName: "",
    deleteVolumes: true,
    agentId: "",
    providerId: 0,
    imageId: 0,
    modelProviderId: 0,
    modelId: 0,
  };
}

function normalizeAgentFormForProvider(
  form: AgentForm,
  provider?: ProviderInfo,
): AgentForm {
  const fixedProtocol = singleProviderProtocol(provider);
  const modelProtocol = fixedProtocol ?? form.modelProtocol;
  return {
    ...form,
    modelProtocol,
    modelName: modelNameForProtocol(
      provider?.models ?? [],
      modelProtocol,
      form.modelName,
    ),
  };
}

function modelNameForProtocol(
  models: ProviderModel[],
  protocol: ModelProtocol,
  currentModelName: string,
) {
  if (models.length === 0) {
    return currentModelName;
  }
  const matchingModels = models.filter((model) => model.protocol === protocol);
  if (matchingModels.some((model) => model.name === currentModelName)) {
    return currentModelName;
  }
  return matchingModels[0]?.name ?? "";
}

function singleProviderProtocol(
  provider?: ProviderInfo,
): ModelProtocol | undefined {
  const protocols = supportedProviderProtocols(provider);
  return protocols.length === 1 ? protocols[0] : undefined;
}

function supportedProviderProtocols(provider?: ProviderInfo): ModelProtocol[] {
  if (!provider) {
    return [];
  }
  const protocols = new Set<ModelProtocol>();
  if (provider.anthropicBaseUrl.trim()) {
    protocols.add("anthropic");
  }
  if (provider.openaiBaseUrl.trim()) {
    protocols.add("openai");
  }
  if (protocols.size === 0) {
    provider.models?.forEach((model) => protocols.add(model.protocol));
  }
  return (["anthropic", "openai"] as const).filter((protocol) =>
    protocols.has(protocol),
  );
}

function protocolsForSync(provider: ProviderInfo): ModelProtocol[] {
  const protocols: ModelProtocol[] = [];
  if (provider.openaiBaseUrl.trim()) {
    protocols.push("openai");
  }
  if (provider.anthropicBaseUrl.trim()) {
    protocols.push("anthropic");
  }
  return protocols.length > 0
    ? protocols
    : supportedProviderProtocols(provider);
}

async function exitDocumentFullscreen() {
  if (document.fullscreenElement) {
    await document.exitFullscreen();
  }
}
