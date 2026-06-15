import {
  lazy,
  memo,
  Suspense,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  isValidElement,
} from "react";
import type { BeforeMount } from "@monaco-editor/react";
import type {
  ComponentType,
  CSSProperties,
  MouseEvent as ReactMouseEvent,
} from "react";
import ReactMarkdown from "react-markdown";
import type {
  Components as MarkdownComponents,
  Options as MarkdownOptions,
} from "react-markdown";
import type { ImperativePanelHandle } from "react-resizable-panels";
import {
  Braces,
  ChevronDown,
  ChevronRight,
  ChevronsUp,
  Clipboard,
  Columns2,
  Download,
  Eye,
  ExternalLink,
  File,
  FileCode2,
  FileCog,
  FileImage,
  FileJson,
  FilePlus2,
  FileText,
  FileX,
  Folder,
  FolderOpen,
  FolderPlus,
  HardDrive,
  Hash,
  GitBranch,
  PanelLeftClose,
  PanelLeftOpen,
  RefreshCw,
  RotateCcw,
  Save,
  Terminal as TerminalIcon,
  Upload,
} from "lucide-react";
import rehypeRaw from "rehype-raw";
import rehypeSanitize, { defaultSchema } from "rehype-sanitize";
import remarkGfm from "remark-gfm";
import { toast } from "sonner";
import { api } from "./api";
import type {
  AgentInfo,
  GitChange,
  GitStatusResponse,
  WorkspaceFilePreview,
  WorkspaceTreeNode,
} from "./types";
import {
  copyWorkspaceText,
  WorkspaceContextMenu,
} from "./WorkspaceContextMenu";
import type {
  WorkspaceContextMenuEntry,
  WorkspaceContextMenuPosition,
} from "./WorkspaceContextMenu";
import {
  Alert,
  Badge,
  Button,
  Dialog,
  EmptyState,
  Field,
  FileUploadButton,
  IconButton,
  Input,
  ResizablePanel,
  ResizablePanelGroup,
  ResizeHandle,
  Spinner,
  TreeButton,
} from "@/components/ui";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { languageFromPath } from "@/lib/editor-language";
import { loadMonacoEditor } from "@/lib/monaco-loader";
import type { WorkbenchSettings } from "@/lib/workbench-settings";
import {
  cn,
  formatBytes,
  sharedRootPath,
  workspaceRootPath,
} from "@/lib/utils";

const MonacoEditor = lazy(loadMonacoEditor);
const TerminalPanelSurface = lazy(() =>
  import("./ShellPanel").then((module) => ({
    default: module.TerminalPanelSurface,
  })),
);
const workbenchEditorTheme = "workbench-dark-plus";
const collapsedWorkbenchSidebarWidth = 44;
const defaultCollapsedPanelSize = 6;
const filesTerminalPanelOpenStorageKey =
  "john-ai-agentbox-files-terminal-panel-open";
const workbenchEditorFontFamily = "Menlo, Monaco";
let workbenchEditorThemeConfigured = false;

type MarkdownViewMode = "source" | "split" | "preview";

type MarkdownHrefResolution =
  | { kind: "blocked" }
  | { kind: "external"; href: string }
  | { kind: "fragment"; href: string }
  | { kind: "workspace"; href: string; path: string };

type MarkdownPreviewPaneProps = {
  agentId: string;
  content: string;
  editorFontSize: number;
  filePath: string;
  onOpenFileResource: (path: string) => void;
  onOpenWorkspacePath: (path: string) => void;
};

type MarkdownPreviewStyle = CSSProperties & {
  "--markdown-font-size": string;
  "--markdown-line-height": string;
};

type MermaidRenderState =
  | { status: "loading" }
  | { status: "rendered"; svg: string }
  | { status: "error"; message: string };

const markdownViewModes: Array<{
  value: MarkdownViewMode;
  label: string;
  icon: ComponentType<{ className?: string }>;
}> = [
  { value: "source", label: "源码", icon: FileCode2 },
  { value: "split", label: "分屏", icon: Columns2 },
  { value: "preview", label: "预览", icon: Eye },
];

const markdownRemarkPlugins: MarkdownOptions["remarkPlugins"] = [remarkGfm];
const mermaidRenderDebounceMs = 250;
const mermaidMaxSourceLength = 20_000;
const markdownSanitizeSchema = {
  ...defaultSchema,
  attributes: {
    ...defaultSchema.attributes,
    "*": [...(defaultSchema.attributes?.["*"] ?? []), "style"],
  },
  tagNames: [...new Set([...(defaultSchema.tagNames ?? []), "mark", "u"])],
};
const markdownRehypePlugins: MarkdownOptions["rehypePlugins"] = [
  rehypeRaw,
  [rehypeSanitize, markdownSanitizeSchema],
];
let mermaidInitializedFontSize: number | null = null;
let mermaidRenderSequence = 0;

const configureWorkbenchMonaco: BeforeMount = (monaco) => {
  if (workbenchEditorThemeConfigured) {
    return;
  }

  monaco.editor.defineTheme(workbenchEditorTheme, {
    base: "vs-dark",
    inherit: true,
    colors: {
      "editor.background": "#1e1e1e",
      "editor.foreground": "#d4d4d4",
      "editor.lineHighlightBackground": "#2a2d2e66",
      "editor.selectionBackground": "#264f78",
      "editorCursor.foreground": "#aeafad",
      "editorLineNumber.activeForeground": "#c6c6c6",
      "editorLineNumber.foreground": "#858585",
      "editorWhitespace.foreground": "#404040",
    },
    rules: [
      { token: "keyword", foreground: "4fc1ff" },
      { token: "keyword.table", foreground: "d4d4d4" },
      {
        token: "keyword.table.header",
        foreground: "d4d4d4",
        fontStyle: "bold",
      },
      { token: "strong", foreground: "4fc1ff", fontStyle: "bold" },
      { token: "emphasis", foreground: "ce9178", fontStyle: "italic" },
      { token: "variable", foreground: "ce9178" },
      { token: "string", foreground: "ce9178" },
      { token: "variable.source", foreground: "d4d4d4" },
      { token: "string.link", foreground: "3794ff", fontStyle: "underline" },
      { token: "string.target", foreground: "d7ba7d" },
      { token: "string.escape", foreground: "d7ba7d" },
      { token: "meta.separator", foreground: "6a9955" },
      { token: "comment", foreground: "6a9955" },
      { token: "tag", foreground: "569cd6" },
      { token: "attribute.name.html", foreground: "9cdcfe" },
      { token: "delimiter.html", foreground: "808080" },
      { token: "string.html", foreground: "ce9178" },
    ],
  });
  workbenchEditorThemeConfigured = true;
};

type Props = {
  active: boolean;
  agent?: AgentInfo;
  agents?: AgentInfo[];
  settings: WorkbenchSettings;
  workspacePath: string;
  canShowInGit?: boolean;
  locateRequest?: FilesLocateRequest | null;
  onLocateRequestHandled?: (id: number) => void;
  onShowInGit?: (path: string) => void;
};

type PendingAction =
  | {
      kind: "open";
      markdownMode?: MarkdownViewMode;
      node: WorkspaceTreeNode;
    }
  | { kind: "createFile"; parentPath: string; name: string }
  | { kind: "createDirectory"; parentPath: string; name: string }
  | { kind: "refreshRoot" }
  | { kind: "refreshDirectory"; path: string }
  | { kind: "reloadCurrent" };

type SelectedWorkspaceNode = Pick<WorkspaceTreeNode, "path" | "type">;

type GitStatusKind =
  | "modified"
  | "untracked"
  | "deleted"
  | "added"
  | "renamed"
  | "copied"
  | "type_changed"
  | "unmerged"
  | "updated"
  | "unknown";

type GitVisualState = {
  status: GitStatusKind;
  staged: boolean;
  unstaged: boolean;
};

type TreeContext = {
  expandedPaths: Set<string>;
  loadingPaths: Set<string>;
  gitStates: Map<string, GitVisualState>;
  selectedPath: string;
};

type FileTreeProps = {
  context: TreeContext;
  nodes: WorkspaceTreeNode[];
  level?: number;
  onContextMenu: (event: ReactMouseEvent, node: WorkspaceTreeNode) => void;
  onPreview: (node: WorkspaceTreeNode) => void;
  onSelectDirectory: (node: WorkspaceTreeNode) => void;
  onToggleDirectory: (node: WorkspaceTreeNode) => void;
};

export type FilesLocateRequest = {
  id: number;
  path: string;
  deleted?: boolean;
  type?: "file" | "directory";
};

type FilesContextMenuTarget =
  | { kind: "node"; node: WorkspaceTreeNode }
  | { kind: "blank" };

type CreateEntryKind = "file" | "directory";

type CreateEntryDialogState = {
  kind: CreateEntryKind;
  parentPath: string;
};

export default function FilesPanel({
  active,
  agent,
  agents = [],
  canShowInGit = true,
  settings,
  workspacePath,
  locateRequest,
  onLocateRequestHandled,
  onShowInGit,
}: Props) {
  const initialTerminalPanelContextKey =
    agent?.id && workspacePath
      ? filesTerminalPanelContextKey(agent.id, workspacePath)
      : "";
  const [tree, setTree] = useState<WorkspaceTreeNode[]>([]);
  const [expandedPaths, setExpandedPaths] = useState<Set<string>>(
    () => new Set(),
  );
  const [loadingPaths, setLoadingPaths] = useState<Set<string>>(
    () => new Set(),
  );
  const [preview, setPreview] = useState<WorkspaceFilePreview | null>(null);
  const [draft, setDraft] = useState("");
  const [selectedPath, setSelectedPath] = useState("");
  const [selectedNode, setSelectedNode] =
    useState<SelectedWorkspaceNode | null>(null);
  const [pendingAction, setPendingAction] = useState<PendingAction | null>(
    null,
  );
  const [confirmDiscardOpen, setConfirmDiscardOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [gitStatus, setGitStatus] = useState<GitStatusResponse | null>(null);
  const [contextMenu, setContextMenu] = useState<{
    position: WorkspaceContextMenuPosition;
    target: FilesContextMenuTarget;
  } | null>(null);
  const [contextUploadTarget, setContextUploadTarget] = useState<string | null>(
    null,
  );
  const [createDialog, setCreateDialog] =
    useState<CreateEntryDialogState | null>(null);
  const [createName, setCreateName] = useState("");
  const [createError, setCreateError] = useState("");
  const [creating, setCreating] = useState(false);
  const [treePaneCollapsed, setTreePaneCollapsed] = useState(false);
  const [terminalPanelOpen, setTerminalPanelOpen] = useState(() =>
    loadFilesTerminalPanelOpen(initialTerminalPanelContextKey),
  );
  const [markdownMode, setMarkdownMode] = useState<MarkdownViewMode>("source");
  const [sidebarCollapsedSize, setSidebarCollapsedSize] = useState(
    defaultCollapsedPanelSize,
  );
  const compactLayout = useMediaQuery("(max-width: 760px)");
  const contextUploadButtonRef = useRef<HTMLLabelElement | null>(null);
  const createNameInputRef = useRef<HTMLInputElement | null>(null);
  const panelGroupShellRef = useRef<HTMLDivElement | null>(null);
  const treePanelRef = useRef<ImperativePanelHandle | null>(null);
  const requestSeqRef = useRef(0);
  const pendingOpenMarkdownModeRef = useRef<MarkdownViewMode | null>(null);
  const sidebarDefaultSize = compactLayout ? 30 : settings.filesSidebarSize;
  const sidebarCurrentSize = treePaneCollapsed
    ? sidebarCollapsedSize
    : sidebarDefaultSize;
  const loadedWorkspaceContextKeyRef = useRef("");

  const agentReady = Boolean(agent?.id && agent.runtimeStatus === "running");
  const workspaceContextKey = `${agent?.id ?? ""}|${agent?.runtimeStatus ?? ""}|${workspacePath}`;
  const terminalPanelContextKey =
    agent?.id && workspacePath
      ? filesTerminalPanelContextKey(agent.id, workspacePath)
      : "";
  const editable =
    agentReady && preview?.previewType === "text" && !preview.tooLarge;
  const dirty = Boolean(
    editable && preview && draft !== (preview.content || ""),
  );
  const gitStates = useMemo(() => buildGitStateIndex(gitStatus), [gitStatus]);
  const treeContext = useMemo<TreeContext>(
    () => ({
      expandedPaths,
      gitStates,
      loadingPaths,
      selectedPath,
    }),
    [expandedPaths, gitStates, loadingPaths, selectedPath],
  );
  const stableOpenFileResource = useLatestCallback(openFileResource);
  const stableOpenWorkspacePathInDetail = useLatestCallback(
    openWorkspacePathInDetail,
  );

  useEffect(() => {
    if (!active || locateRequest) {
      return;
    }
    if (!agentReady || !workspacePath) {
      if (loadedWorkspaceContextKeyRef.current !== workspaceContextKey) {
        resetPanelState();
        loadedWorkspaceContextKeyRef.current = workspaceContextKey;
      }
      return;
    }
    if (loadedWorkspaceContextKeyRef.current === workspaceContextKey) {
      return;
    }
    void loadRoot({ protectDirty: true });
  }, [active, agentReady, locateRequest, workspaceContextKey, workspacePath]);

  useEffect(() => {
    if (!active || !locateRequest) {
      return;
    }
    void locatePath(locateRequest);
  }, [active, locateRequest?.id]);

  useEffect(() => {
    const pendingMode = pendingOpenMarkdownModeRef.current;
    pendingOpenMarkdownModeRef.current = null;
    if (
      pendingMode &&
      preview?.file.path &&
      isMarkdownPreviewFile(preview.file.path)
    ) {
      setMarkdownMode(pendingMode);
      return;
    }
    setMarkdownMode("source");
  }, [preview?.file.path]);

  useEffect(() => {
    setTerminalPanelOpen(loadFilesTerminalPanelOpen(terminalPanelContextKey));
  }, [terminalPanelContextKey]);

  useEffect(() => {
    const element = panelGroupShellRef.current;
    if (!element) {
      return;
    }
    const updateCollapsedSize = () => {
      setSidebarCollapsedSize((current) => {
        const next = collapsedPanelSizePercent(element, compactLayout);
        return Math.abs(current - next) < 0.01 ? current : next;
      });
    };
    updateCollapsedSize();
    if (typeof ResizeObserver === "undefined") {
      window.addEventListener("resize", updateCollapsedSize);
      return () => window.removeEventListener("resize", updateCollapsedSize);
    }
    const observer = new ResizeObserver(updateCollapsedSize);
    observer.observe(element);
    return () => observer.disconnect();
  }, [compactLayout]);

  useEffect(() => {
    if (treePaneCollapsed) {
      treePanelRef.current?.collapse();
    }
  }, [treePaneCollapsed]);

  useEffect(() => {
    if (!createDialog) {
      return;
    }
    const timeout = window.setTimeout(() => {
      createNameInputRef.current?.focus();
      createNameInputRef.current?.select();
    }, 0);
    return () => window.clearTimeout(timeout);
  }, [createDialog]);

  const closeContextMenu = useCallback(() => {
    setContextMenu(null);
  }, []);

  function collapseTreePane() {
    closeContextMenu();
    setTreePaneCollapsed(true);
  }

  function expandTreePane() {
    treePanelRef.current?.expand(sidebarDefaultSize);
    setTreePaneCollapsed(false);
  }

  function toggleTerminalPanelOpen() {
    setTerminalPanelOpen((current) => {
      const next = !current;
      saveFilesTerminalPanelOpen(terminalPanelContextKey, next);
      return next;
    });
  }

  async function loadRoot({ protectDirty }: { protectDirty: boolean }) {
    const contextKey = workspaceContextKey;
    if (!agent?.id || agent.runtimeStatus !== "running" || !workspacePath) {
      resetPanelState();
      loadedWorkspaceContextKeyRef.current = contextKey;
      return;
    }
    const agentId = agent.id;
    if (protectDirty && dirty) {
      requestDirtyAction({ kind: "refreshRoot" });
      return;
    }
    const sequence = requestSeqRef.current + 1;
    requestSeqRef.current = sequence;
    setLoading(true);
    try {
      const [nodes, status] = await Promise.all([
        api.workspaceTree(agentId, workspacePath, true),
        api.gitStatus(agentId, workspacePath).catch(() => null),
      ]);
      if (requestSeqRef.current !== sequence) {
        return;
      }
      setTree(nodes);
      setExpandedPaths(new Set());
      setGitStatus(status);
      loadedWorkspaceContextKeyRef.current = contextKey;
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
    } finally {
      if (requestSeqRef.current === sequence) {
        setLoading(false);
      }
    }
  }

  function requestDirtyAction(action: PendingAction) {
    if (dirty) {
      setPendingAction(action);
      setConfirmDiscardOpen(true);
      return;
    }
    void runPendingAction(action);
  }

  async function runPendingAction(action: PendingAction) {
    if (action.kind === "open") {
      await openNode(action.node, action.markdownMode);
      return;
    }
    if (action.kind === "createFile") {
      await createWorkspaceFile(action.parentPath, action.name);
      return;
    }
    if (action.kind === "createDirectory") {
      await createWorkspaceDirectory(action.parentPath, action.name);
      return;
    }
    if (action.kind === "reloadCurrent") {
      await reloadCurrent({ protectDirty: false });
      return;
    }
    if (action.kind === "refreshDirectory") {
      await refreshDirectory(action.path);
      return;
    }
    await loadRoot({ protectDirty: false });
  }

  async function openPreview(node: WorkspaceTreeNode) {
    if (!agentReady || node.type !== "file") {
      return;
    }
    requestDirtyAction({ kind: "open", node });
  }

  async function openNode(
    node: WorkspaceTreeNode,
    markdownMode?: MarkdownViewMode,
  ) {
    if (!agentReady || !agent?.id) {
      return;
    }
    setLoading(true);
    try {
      const nextPreview = await api.workspaceFile(agent.id, node.path);
      pendingOpenMarkdownModeRef.current = markdownMode ?? null;
      setPreview(nextPreview);
      setDraft(nextPreview.content || "");
      setSelectedPath(node.path);
      setSelectedNode(node);
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
    } finally {
      setLoading(false);
    }
  }

  function selectDirectory(node: WorkspaceTreeNode) {
    setSelectedPath(node.path);
    setSelectedNode(node);
  }

  async function toggleDirectory(node: WorkspaceTreeNode) {
    selectDirectory(node);
    if (expandedPaths.has(node.path)) {
      setExpandedPaths((current) => {
        const next = new Set(current);
        next.delete(node.path);
        return next;
      });
      return;
    }
    setExpandedPaths((current) => new Set(current).add(node.path));
    if (node.children) {
      return;
    }
    await loadDirectory(node.path);
  }

  async function loadDirectory(path: string) {
    if (!agentReady || !agent?.id) {
      return;
    }
    setLoadingPaths((current) => new Set(current).add(path));
    try {
      const children = await api.workspaceTree(agent.id, path, true);
      setTree((current) => updateTreeChildren(current, path, children));
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
    } finally {
      setLoadingPaths((current) => {
        const next = new Set(current);
        next.delete(path);
        return next;
      });
    }
  }

  async function refreshDirectory(path: string) {
    if (!agentReady || !agent?.id) {
      return;
    }
    setLoadingPaths((current) => new Set(current).add(path));
    try {
      const children = await api.workspaceTree(agent.id, path, true);
      if (path === workspacePath) {
        setTree(children);
      } else {
        setTree((current) => updateTreeChildren(current, path, children));
      }
      setExpandedPaths((current) => new Set(current).add(path));
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
    } finally {
      setLoadingPaths((current) => {
        const next = new Set(current);
        next.delete(path);
        return next;
      });
    }
  }

  async function reloadCurrent({ protectDirty }: { protectDirty: boolean }) {
    if (!agentReady || !agent?.id || !preview?.file.path) {
      return;
    }
    if (protectDirty && dirty) {
      requestDirtyAction({ kind: "reloadCurrent" });
      return;
    }
    setLoading(true);
    try {
      const nextPreview = await api.workspaceFile(agent.id, preview.file.path);
      setPreview(nextPreview);
      setDraft(nextPreview.content || "");
      setSelectedPath(nextPreview.file.path);
      setSelectedNode(nextPreview.file);
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
    } finally {
      setLoading(false);
    }
  }

  async function saveCurrent() {
    if (
      !agentReady ||
      !agent?.id ||
      !preview?.file.path ||
      !editable ||
      !dirty
    ) {
      return false;
    }
    setSaving(true);
    try {
      const saved = await api.saveWorkspaceFile(agent.id, {
        path: preview.file.path,
        content: draft,
        encoding: preview.encoding || "utf-8",
        baseHash: preview.contentHash,
      });
      setPreview(saved);
      setDraft(saved.content || "");
      setSelectedPath(saved.file.path);
      setSelectedNode(saved.file);
      await refreshLoadedTreeAfterChange(saved.file.path);
      return true;
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
      return false;
    } finally {
      setSaving(false);
    }
  }

  async function refreshLoadedTreeAfterChange(path: string) {
    if (!agentReady || !agent?.id) {
      return;
    }
    const targetDir = parentPath(path);
    const refreshPath = expandedPaths.has(targetDir)
      ? targetDir
      : workspacePath;
    const children = await api.workspaceTree(agent.id, refreshPath, true);
    if (refreshPath === workspacePath) {
      setTree(children);
    } else {
      setTree((current) => updateTreeChildren(current, refreshPath, children));
    }
    const status = await api
      .gitStatus(agent.id, workspacePath)
      .catch(() => null);
    setGitStatus(status);
  }

  async function refreshDirectoryAfterCreate(parentPath: string) {
    if (!agentReady || !agent?.id) {
      return;
    }
    const children = await api.workspaceTree(agent.id, parentPath, true);
    if (parentPath === workspacePath) {
      setTree(children);
    } else {
      setTree((current) => updateTreeChildren(current, parentPath, children));
    }
    setExpandedPaths((current) => new Set(current).add(parentPath));
    const status = await api
      .gitStatus(agent.id, workspacePath)
      .catch(() => null);
    setGitStatus(status);
  }

  async function createWorkspaceFile(parentPath: string, name: string) {
    if (!agentReady || !agent?.id || !workspacePath) {
      return;
    }
    setCreating(true);
    setLoading(true);
    try {
      const created = await api.createWorkspaceFile(agent.id, {
        parentPath,
        name,
      });
      setPreview(created);
      setDraft(created.content || "");
      setSelectedPath(created.file.path);
      setSelectedNode(created.file);
      await refreshDirectoryAfterCreate(parentPath);
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
    } finally {
      setCreating(false);
      setLoading(false);
    }
  }

  async function createWorkspaceDirectory(parentPath: string, name: string) {
    if (!agentReady || !agent?.id || !workspacePath) {
      return;
    }
    setCreating(true);
    setLoading(true);
    try {
      await api.createWorkspaceDirectory(agent.id, {
        parentPath,
        name,
      });
      await refreshDirectoryAfterCreate(parentPath);
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
    } finally {
      setCreating(false);
      setLoading(false);
    }
  }

  async function uploadFiles(
    files: FileList | null,
    explicitTargetPath?: string,
  ) {
    if (!agentReady || !agent?.id || !workspacePath || !files?.length) {
      return;
    }
    const targetPath =
      explicitTargetPath || uploadTargetPath(selectedNode, workspacePath);
    setLoading(true);
    try {
      await api.uploadWorkspaceFiles(agent.id, targetPath, Array.from(files));
      const refreshed = await api.workspaceTree(agent.id, targetPath, true);
      if (targetPath === workspacePath) {
        setTree(refreshed);
      } else {
        setTree((current) =>
          updateTreeChildren(current, targetPath, refreshed),
        );
        setExpandedPaths((current) => new Set(current).add(targetPath));
      }
      const status = await api
        .gitStatus(agent.id, workspacePath)
        .catch(() => null);
      setGitStatus(status);
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
    } finally {
      setLoading(false);
    }
  }

  function download(path: string) {
    if (!agentReady || !agent?.id) {
      return;
    }
    window.open(
      api.workspaceResourceUrl(agent.id, path, "attachment"),
      "_blank",
      "noopener,noreferrer",
    );
  }

  function openFileResource(path: string) {
    if (!agentReady || !agent?.id) {
      return;
    }
    window.open(
      isHtmlFile(path)
        ? api.workspaceHtmlPreviewUrl(agent.id, path)
        : api.workspaceResourceUrl(agent.id, path, "inline"),
      "_blank",
      "noopener,noreferrer",
    );
  }

  function openWorkspacePathInDetail(path: string) {
    if (!agentReady || !path) {
      return;
    }
    const requestedMarkdownMode =
      markdownMode !== "source" && isMarkdownPreviewFile(path)
        ? markdownMode
        : undefined;
    const node = findTreeNode(tree, path) ?? {
      name: basename(path),
      path,
      type: "file" as const,
      expandable: false,
    };
    requestDirtyAction({
      kind: "open",
      markdownMode: requestedMarkdownMode,
      node,
    });
  }

  function openContextMenu(
    event: ReactMouseEvent,
    target: FilesContextMenuTarget,
  ) {
    event.preventDefault();
    event.stopPropagation();
    setContextMenu({
      position: { x: event.clientX, y: event.clientY },
      target,
    });
  }

  function triggerContextUpload(targetPath: string) {
    setContextUploadTarget(targetPath);
    window.setTimeout(() => contextUploadButtonRef.current?.click(), 0);
  }

  async function uploadContextFiles(files: FileList | null) {
    const targetPath =
      contextUploadTarget || uploadTargetPath(selectedNode, workspacePath);
    setContextUploadTarget(null);
    await uploadFiles(files, targetPath);
  }

  function openCreateEntryDialog(kind: CreateEntryKind, parentPath: string) {
    if (!parentPath) {
      return;
    }
    setCreateDialog({ kind, parentPath });
    setCreateName(kind === "file" ? "untitled.txt" : "untitled");
    setCreateError("");
  }

  function closeCreateEntryDialog() {
    setCreateDialog(null);
    setCreateName("");
    setCreateError("");
  }

  function submitCreateEntry() {
    if (!createDialog) {
      return;
    }
    const name = createName.trim();
    const error = workspaceCreateNameError(name);
    if (error) {
      setCreateError(error);
      return;
    }
    const action: PendingAction =
      createDialog.kind === "file"
        ? { kind: "createFile", parentPath: createDialog.parentPath, name }
        : {
            kind: "createDirectory",
            parentPath: createDialog.parentPath,
            name,
          };
    closeCreateEntryDialog();
    if (action.kind === "createFile") {
      requestDirtyAction(action);
      return;
    }
    void runPendingAction(action);
  }

  async function locatePath(request: FilesLocateRequest) {
    if (!agentReady || !agent?.id || !workspacePath) {
      onLocateRequestHandled?.(request.id);
      return;
    }
    const targetPath = joinWorkspacePath(
      workspacePath,
      relativeWorkspacePath(request.path, workspacePath),
    );
    const parent = parentPath(targetPath);
    try {
      await expandDirectoryPath(
        request.type === "directory" ? targetPath : parent,
      );
      setSelectedPath(targetPath);
      setSelectedNode({
        path: targetPath,
        type: request.type === "directory" ? "directory" : "file",
      });
      if (request.deleted) {
        toast.info("该文件已在工作区删除");
      } else if (request.type !== "directory") {
        const node = findTreeNode(tree, targetPath) ?? {
          name: basename(targetPath),
          path: targetPath,
          type: "file" as const,
          expandable: false,
        };
        requestDirtyAction({ kind: "open", node });
      } else {
        toast.info("已在 Explorer 中定位目录");
      }
    } finally {
      onLocateRequestHandled?.(request.id);
    }
  }

  async function expandDirectoryPath(targetDir: string) {
    if (!agentReady || !agent?.id) {
      return;
    }
    const contextKey = workspaceContextKey;
    let nextTree =
      loadedWorkspaceContextKeyRef.current === contextKey ? tree : [];
    if (nextTree.length === 0) {
      nextTree = await api.workspaceTree(agent.id, workspacePath, true);
      setTree(nextTree);
      loadedWorkspaceContextKeyRef.current = contextKey;
    }
    const relative = relativeWorkspacePath(targetDir, workspacePath);
    const parts = relative ? relative.split("/").filter(Boolean) : [];
    let currentPath = workspacePath.replace(/\/+$/, "");
    for (const part of parts) {
      currentPath = `${currentPath}/${part}`;
      const children = await api.workspaceTree(agent.id, currentPath, true);
      nextTree = updateTreeChildren(nextTree, currentPath, children);
      setTree(nextTree);
      setExpandedPaths((current) => new Set(current).add(currentPath));
    }
  }

  function buildContextMenuItems(): WorkspaceContextMenuEntry[] {
    if (!contextMenu) {
      return [];
    }
    const disabled =
      !agentReady || !workspacePath || loading || saving || creating;
    if (contextMenu.target.kind === "blank") {
      const targetPath = uploadTargetPath(selectedNode, workspacePath);
      return [
        {
          id: "create-file",
          label: "新建文件",
          icon: <FilePlus2 />,
          disabled,
          testId: "files-context-create-file",
          onSelect: () => openCreateEntryDialog("file", targetPath),
        },
        {
          id: "create-directory",
          label: "新建文件夹",
          icon: <FolderPlus />,
          disabled,
          testId: "files-context-create-directory",
          onSelect: () => openCreateEntryDialog("directory", targetPath),
        },
        { kind: "separator", id: "blank-create-separator" },
        {
          id: "refresh-explorer",
          label: "刷新",
          icon: <RefreshCw />,
          disabled,
          testId: "files-context-refresh",
          onSelect: () => requestDirtyAction({ kind: "refreshRoot" }),
        },
        {
          id: "upload-current",
          label: "上传",
          icon: <Upload />,
          disabled,
          testId: "files-context-upload-current",
          onSelect: () => triggerContextUpload(targetPath),
        },
        { kind: "separator", id: "blank-separator" },
        {
          id: "collapse-all",
          label: "全部折叠",
          icon: <ChevronsUp />,
          testId: "files-context-collapse-all",
          onSelect: () => setExpandedPaths(new Set()),
        },
        {
          id: "copy-workspace",
          label: "复制绝对路径",
          icon: <Clipboard />,
          disabled: !workspacePath,
          testId: "files-context-copy-workspace",
          onSelect: () =>
            copyWorkspaceText(workspacePath, "workspace 路径已复制"),
        },
      ];
    }

    const node = contextMenu.target.node;
    const relativePath = relativeWorkspacePath(node.path, workspacePath);
    if (node.type === "directory") {
      const expanded = expandedPaths.has(node.path);
      return [
        {
          id: "create-file",
          label: "新建文件",
          icon: <FilePlus2 />,
          disabled,
          testId: "files-context-create-file",
          onSelect: () => openCreateEntryDialog("file", node.path),
        },
        {
          id: "create-directory",
          label: "新建文件夹",
          icon: <FolderPlus />,
          disabled,
          testId: "files-context-create-directory",
          onSelect: () => openCreateEntryDialog("directory", node.path),
        },
        { kind: "separator", id: "directory-create-separator" },
        {
          id: "toggle-directory",
          label: expanded ? "折叠" : "展开",
          icon: expanded ? <ChevronDown /> : <ChevronRight />,
          disabled: !agentReady,
          testId: "files-context-toggle-directory",
          onSelect: () => void toggleDirectory(node),
        },
        {
          id: "refresh-directory",
          label: "刷新目录",
          icon: <RefreshCw />,
          disabled,
          testId: "files-context-refresh-directory",
          onSelect: () =>
            requestDirtyAction({ kind: "refreshDirectory", path: node.path }),
        },
        {
          id: "upload-here",
          label: "上传到此处",
          icon: <Upload />,
          disabled,
          testId: "files-context-upload-here",
          onSelect: () => triggerContextUpload(node.path),
        },
        { kind: "separator", id: "directory-separator" },
        {
          id: "copy-absolute",
          label: "复制绝对路径",
          icon: <Clipboard />,
          testId: "files-context-copy-absolute",
          onSelect: () =>
            copyWorkspaceText(node.path, "workspace 绝对路径已复制"),
        },
        {
          id: "copy-relative",
          label: "复制相对路径",
          icon: <Clipboard />,
          testId: "files-context-copy-relative",
          onSelect: () => copyWorkspaceText(relativePath, "相对路径已复制"),
        },
      ];
    }

    return [
      {
        id: "open-file",
        label: "打开",
        icon: <ExternalLink />,
        disabled,
        testId: "files-context-open-file",
        onSelect: () => openFileResource(node.path),
      },
      {
        id: "download-file",
        label: "下载",
        icon: <Download />,
        disabled,
        testId: "files-context-download-file",
        onSelect: () => download(node.path),
      },
      ...(canShowInGit
        ? [
            {
              id: "show-in-git",
              label: "在 Source Control 中查看变化",
              icon: <GitBranch />,
              disabled: !agentReady || !workspacePath,
              testId: "files-context-show-in-git",
              onSelect: () => onShowInGit?.(node.path),
            } satisfies WorkspaceContextMenuEntry,
          ]
        : []),
      { kind: "separator", id: "file-separator" },
      {
        id: "copy-absolute",
        label: "复制绝对路径",
        icon: <Clipboard />,
        testId: "files-context-copy-absolute",
        onSelect: () =>
          copyWorkspaceText(node.path, "workspace 绝对路径已复制"),
      },
      {
        id: "copy-relative",
        label: "复制相对路径",
        icon: <Clipboard />,
        testId: "files-context-copy-relative",
        onSelect: () => copyWorkspaceText(relativePath, "相对路径已复制"),
      },
    ];
  }

  function resetPanelState() {
    requestSeqRef.current += 1;
    pendingOpenMarkdownModeRef.current = null;
    setTree([]);
    setPreview(null);
    setDraft("");
    setSelectedPath("");
    setSelectedNode(null);
    setPendingAction(null);
    setConfirmDiscardOpen(false);
    closeCreateEntryDialog();
    setExpandedPaths(new Set());
    setLoadingPaths(new Set());
    setGitStatus(null);
    setLoading(false);
    setSaving(false);
    setCreating(false);
  }

  async function discardAndContinue() {
    const next = pendingAction;
    setConfirmDiscardOpen(false);
    setPendingAction(null);
    setDraft(preview?.content || "");
    if (next) {
      await runPendingAction(next);
    }
  }

  async function saveAndContinue() {
    const next = pendingAction;
    const saved = await saveCurrent();
    if (!saved) {
      return;
    }
    setConfirmDiscardOpen(false);
    setPendingAction(null);
    if (next) {
      await runPendingAction(next);
    }
  }

  const fileEditorPane = (
    <FileEditorPane
      agentId={agentReady && agent?.id ? agent.id : ""}
      dirty={dirty}
      editable={Boolean(editable)}
      imageSrc={
        agentReady && agent?.id && preview
          ? api.workspaceDownloadUrl(agent.id, preview.file.path)
          : ""
      }
      offlineDescription={
        agent
          ? `当前状态：${runtimeStatusLabel(agent.runtimeStatus || "unknown")}。启动后即可浏览文件。`
          : "请选择一个运行中的智能体。"
      }
      offlineTitle={agentReady ? "" : "当前 Agent 未运行"}
      markdownMode={markdownMode}
      preview={preview}
      saving={saving}
      settings={settings}
      terminalOpen={terminalPanelOpen}
      terminalToggleDisabled={!agent || (!agentReady && !terminalPanelOpen)}
      value={draft}
      onDownload={download}
      onOpenFileResource={stableOpenFileResource}
      onOpenWorkspacePath={stableOpenWorkspacePathInDetail}
      onMarkdownModeChange={setMarkdownMode}
      onReload={() => reloadCurrent({ protectDirty: true })}
      onSave={() => void saveCurrent()}
      onToggleTerminalPanel={toggleTerminalPanelOpen}
      onValueChange={setDraft}
    />
  );

  return (
    <section
      className={cn(
        "grid h-full min-h-0 grid-rows-[auto_minmax(0,1fr)] gap-0 overflow-hidden max-[760px]:h-[720px] max-[760px]:min-h-[720px]",
        active ? "" : "hidden",
      )}
      data-testid="files-panel"
    >
      <div className="grid gap-2">
        {!workspacePath ? (
          <Alert>
            选择 {workspaceRootPath} 或 {sharedRootPath} 下的目录后展示文件树。
          </Alert>
        ) : null}
      </div>
      <div ref={panelGroupShellRef} className="h-full min-h-0 overflow-hidden">
        <ResizablePanelGroup
          key={`${compactLayout ? "files-compact" : "files-wide"}-${sidebarDefaultSize}`}
          className="workspace-dark-plus h-full min-h-0 overflow-hidden bg-card"
          data-sidebar-collapsed={treePaneCollapsed}
          data-sidebar-default-size={sidebarDefaultSize}
          data-workbench-theme="vscode-dark-plus"
          direction={compactLayout ? "vertical" : "horizontal"}
        >
          <ResizablePanel
            ref={treePanelRef}
            collapsible={treePaneCollapsed}
            collapsedSize={sidebarCollapsedSize}
            defaultSize={sidebarCurrentSize}
            minSize={treePaneCollapsed ? sidebarCollapsedSize : 0}
            onCollapse={() => setTreePaneCollapsed(true)}
            onExpand={() => setTreePaneCollapsed(false)}
          >
            {treePaneCollapsed ? (
              <div
                className="workbench-sidebar-pane flex h-full min-h-0 items-start justify-center pt-2"
                data-testid="files-tree-collapsed-rail"
              >
                <IconButton
                  className="workbench-icon-button"
                  data-testid="files-tree-expand-button"
                  title="展开 Explorer"
                  onClick={expandTreePane}
                >
                  <PanelLeftOpen />
                </IconButton>
              </div>
            ) : (
              <div
                className="workbench-sidebar-pane grid h-full min-h-0 grid-rows-[auto_minmax(0,1fr)]"
                data-testid="files-tree-pane"
              >
                <div className="workbench-pane-title flex min-h-10 items-center justify-between gap-2 border-b px-3 text-xs font-medium uppercase text-muted-foreground">
                  <span>Explorer</span>
                  <div
                    className="flex items-center gap-1"
                    data-testid="files-tree-actions"
                  >
                    <IconButton
                      className="workbench-icon-button"
                      data-testid="files-refresh-button"
                      disabled={
                        !agentReady || !workspacePath || loading || saving
                      }
                      title="刷新 Explorer"
                      onClick={() =>
                        requestDirtyAction({ kind: "refreshRoot" })
                      }
                    >
                      <RefreshCw className={cn(loading && "animate-spin")} />
                    </IconButton>
                    <FileUploadButton
                      buttonTestId="files-upload-button"
                      className="workbench-icon-button size-7 rounded-[min(var(--radius-md),12px)] border-0 bg-transparent p-0"
                      disabled={
                        !agentReady || !workspacePath || loading || saving
                      }
                      inputTestId="files-upload-input"
                      multiple
                      title="上传文件"
                      onFiles={(files) => void uploadFiles(files)}
                    >
                      <Upload />
                      <span className="sr-only">上传文件</span>
                    </FileUploadButton>
                    <IconButton
                      className="workbench-icon-button"
                      data-testid="files-tree-collapse-button"
                      title="折叠 Explorer"
                      onClick={collapseTreePane}
                    >
                      <PanelLeftClose />
                    </IconButton>
                  </div>
                </div>
                <div
                  className="workbench-pane-scroll min-h-0 overflow-auto py-1"
                  data-testid="files-tree-scroll"
                  onContextMenu={(event) =>
                    openContextMenu(event, { kind: "blank" })
                  }
                >
                  {!loading && tree.length === 0 ? (
                    <EmptyState icon={<FolderOpen />} title="目录为空" />
                  ) : null}
                  <FileTree
                    context={treeContext}
                    nodes={tree}
                    onContextMenu={(event, node) =>
                      openContextMenu(event, { kind: "node", node })
                    }
                    onPreview={(node) => void openPreview(node)}
                    onSelectDirectory={selectDirectory}
                    onToggleDirectory={(node) => void toggleDirectory(node)}
                  />
                </div>
              </div>
            )}
          </ResizablePanel>
          <ResizeHandle
            className="workbench-resize-handle"
            data-testid="files-resize-handle"
          />
          <ResizablePanel
            defaultSize={100 - sidebarCurrentSize}
            minSize={compactLayout ? 42 : 42}
          >
            <div
              className="h-full min-h-0"
              data-terminal-open={terminalPanelOpen ? "true" : "false"}
              data-testid="files-content-pane"
            >
              {terminalPanelOpen ? (
                <ResizablePanelGroup
                  autoSaveId="files-terminal-panel-layout"
                  className="h-full min-h-0"
                  data-testid="files-content-terminal-layout"
                  direction="vertical"
                >
                  <ResizablePanel defaultSize={66} minSize={28}>
                    {fileEditorPane}
                  </ResizablePanel>
                  <ResizeHandle
                    className="workbench-resize-handle"
                    data-testid="files-terminal-resize-handle"
                  />
                  <ResizablePanel defaultSize={34} minSize={18}>
                    <Suspense fallback={<Spinner label="加载终端" />}>
                      <TerminalPanelSurface
                        active={active && terminalPanelOpen}
                        agent={agent}
                        agents={agents}
                        chromeDensity="compact"
                        className="border-t border-border"
                        headerTestId="files-terminal-header"
                        panelTestId="files-terminal-panel"
                        settings={settings}
                        storageKey="john-ai-agentbox-files-terminal-tabs"
                        surfaceTestId="files-terminal-surface"
                        terminalIdPrefix="files-terminal"
                        workspacePath={workspacePath}
                      />
                    </Suspense>
                  </ResizablePanel>
                </ResizablePanelGroup>
              ) : (
                fileEditorPane
              )}
            </div>
          </ResizablePanel>
        </ResizablePanelGroup>
      </div>
      <FileUploadButton
        buttonRef={contextUploadButtonRef}
        className="sr-only"
        disabled={!agentReady || !workspacePath || loading || saving}
        inputTestId="files-context-upload-input"
        multiple
        title="右键上传文件"
        onFiles={(files) => void uploadContextFiles(files)}
      >
        <Upload />
        <span className="sr-only">右键上传文件</span>
      </FileUploadButton>
      <WorkspaceContextMenu
        items={buildContextMenuItems()}
        label="Explorer 右键菜单"
        position={contextMenu?.position ?? null}
        testId="files-context-menu"
        onClose={closeContextMenu}
      />
      <Dialog
        className="sm:max-w-md"
        footer={
          <>
            <Button
              type="button"
              variant="soft"
              onClick={closeCreateEntryDialog}
            >
              取消
            </Button>
            <Button
              disabled={creating || !createName.trim()}
              type="button"
              onClick={submitCreateEntry}
            >
              {createDialog?.kind === "directory" ? (
                <FolderPlus data-icon="inline-start" />
              ) : (
                <FilePlus2 data-icon="inline-start" />
              )}
              创建
            </Button>
          </>
        }
        open={Boolean(createDialog)}
        title={createDialog?.kind === "directory" ? "新建文件夹" : "新建文件"}
        onClose={closeCreateEntryDialog}
      >
        <div
          className="flex flex-col gap-3"
          data-testid="files-create-entry-form"
        >
          <Field
            error={createError || undefined}
            label={
              createDialog?.kind === "directory" ? "文件夹名称" : "文件名称"
            }
          >
            <Input
              ref={createNameInputRef}
              aria-label={
                createDialog?.kind === "directory" ? "文件夹名称" : "文件名称"
              }
              data-testid="files-create-entry-name-input"
              value={createName}
              onChange={(event) => {
                setCreateName(event.target.value);
                setCreateError("");
              }}
              onKeyDown={(event) => {
                if (event.key === "Enter") {
                  submitCreateEntry();
                }
              }}
            />
          </Field>
          <div className="flex min-w-0 flex-col gap-1 rounded-md border bg-muted/30 px-3 py-2 text-xs">
            <span className="font-medium text-muted-foreground">创建位置</span>
            <span
              className="truncate font-mono text-foreground"
              data-testid="files-create-entry-parent"
            >
              {createDialog?.parentPath}
            </span>
          </div>
        </div>
      </Dialog>
      <Dialog
        className="sm:max-w-md"
        footer={
          <>
            <Button
              type="button"
              variant="soft"
              onClick={() => {
                setConfirmDiscardOpen(false);
                setPendingAction(null);
              }}
            >
              取消
            </Button>
            <Button
              disabled={saving}
              type="button"
              variant="danger"
              onClick={() => void discardAndContinue()}
            >
              放弃修改
            </Button>
            <Button
              disabled={saving}
              type="button"
              onClick={() => void saveAndContinue()}
            >
              <Save />
              保存并继续
            </Button>
          </>
        }
        open={confirmDiscardOpen}
        title="处理未保存修改"
        onClose={() => {
          setConfirmDiscardOpen(false);
          setPendingAction(null);
        }}
      >
        <div className="text-sm leading-6 text-muted-foreground">
          当前文件存在未保存修改。可以先保存，再继续切换或刷新。
        </div>
      </Dialog>
    </section>
  );
}

function FileTree({
  context,
  nodes,
  level = 0,
  onContextMenu,
  onPreview,
  onSelectDirectory,
  onToggleDirectory,
}: FileTreeProps) {
  return (
    <div className="grid gap-0">
      {nodes.map((node) => {
        const expanded = context.expandedPaths.has(node.path);
        const loading = context.loadingPaths.has(node.path);
        const active = context.selectedPath === node.path;
        const gitState = context.gitStates.get(node.path);
        const Icon =
          node.type === "directory"
            ? expanded
              ? FolderOpen
              : Folder
            : fileIcon(node.name);
        const iconKind =
          node.type === "directory"
            ? expanded
              ? "folder-open"
              : "folder"
            : fileIconKind(node.name);
        return (
          <div key={node.path}>
            <div
              className={cn(
                "workbench-tree-row grid items-center gap-1.5 text-[13px] hover:bg-accent",
                node.type === "file"
                  ? "grid-cols-[minmax(0,1fr)_auto] pr-1"
                  : "grid-cols-[minmax(0,1fr)]",
                active &&
                  "is-active bg-primary/10 text-foreground hover:bg-primary/15",
              )}
              data-git-status={gitState?.status}
              data-icon-kind={iconKind}
              data-testid={`file-node-${node.path}`}
              style={{ paddingLeft: 8 + level * 18 }}
              onContextMenu={(event) => onContextMenu(event, node)}
            >
              {node.type === "directory" ? (
                <TreeButton
                  className="w-full overflow-hidden"
                  onClick={() => onToggleDirectory(node)}
                >
                  {node.expandable ? (
                    expanded ? (
                      <ChevronDown className="shrink-0 text-muted-foreground" />
                    ) : (
                      <ChevronRight className="shrink-0 text-muted-foreground" />
                    )
                  ) : (
                    <span className="w-4 shrink-0" />
                  )}
                  <Icon
                    className={cn(
                      "shrink-0 workbench-folder-icon",
                      expanded ? "text-primary" : "text-muted-foreground",
                    )}
                  />
                  <span
                    className={cn("truncate", gitTextClass(gitState?.status))}
                  >
                    {node.name}
                  </span>
                  {loading ? (
                    <RefreshCw className="animate-spin text-muted-foreground" />
                  ) : null}
                </TreeButton>
              ) : (
                <TreeButton
                  className="w-full overflow-hidden"
                  onClick={() => onPreview(node)}
                >
                  <span className="w-4 shrink-0" />
                  <Icon className="shrink-0 text-muted-foreground workbench-file-icon" />
                  <span
                    className={cn("truncate", gitTextClass(gitState?.status))}
                  >
                    {node.name}
                  </span>
                  {gitState?.staged && gitState.unstaged ? (
                    <Badge tone="info">S+U</Badge>
                  ) : null}
                  {node.size ? (
                    <span className="shrink-0 text-xs text-muted-foreground">
                      {formatBytes(node.size)}
                    </span>
                  ) : null}
                </TreeButton>
              )}
            </div>
            {node.type === "directory" && expanded && node.children?.length ? (
              <FileTree
                context={context}
                level={level + 1}
                nodes={node.children}
                onContextMenu={onContextMenu}
                onPreview={onPreview}
                onSelectDirectory={onSelectDirectory}
                onToggleDirectory={onToggleDirectory}
              />
            ) : null}
            {node.type === "directory" &&
            expanded &&
            node.children?.length === 0 &&
            !loading ? (
              <TreeButton
                className="workbench-tree-row w-full text-left text-[13px] text-muted-foreground hover:bg-accent"
                style={{ paddingLeft: 28 + (level + 1) * 18 }}
                onClick={() => onSelectDirectory(node)}
              >
                空目录
              </TreeButton>
            ) : null}
          </div>
        );
      })}
    </div>
  );
}

function FileEditorPane({
  agentId,
  dirty,
  editable,
  imageSrc,
  markdownMode,
  offlineDescription,
  offlineTitle,
  preview,
  saving,
  settings,
  terminalOpen,
  terminalToggleDisabled,
  value,
  onDownload,
  onMarkdownModeChange,
  onOpenFileResource,
  onOpenWorkspacePath,
  onReload,
  onSave,
  onToggleTerminalPanel,
  onValueChange,
}: {
  agentId: string;
  dirty: boolean;
  editable: boolean;
  imageSrc: string;
  markdownMode: MarkdownViewMode;
  offlineDescription: string;
  offlineTitle: string;
  preview: WorkspaceFilePreview | null;
  saving: boolean;
  settings: WorkbenchSettings;
  terminalOpen: boolean;
  terminalToggleDisabled: boolean;
  value: string;
  onDownload: (path: string) => void;
  onMarkdownModeChange: (mode: MarkdownViewMode) => void;
  onOpenFileResource: (path: string) => void;
  onOpenWorkspacePath: (path: string) => void;
  onReload: () => void;
  onSave: () => void;
  onToggleTerminalPanel: () => void;
  onValueChange: (value: string) => void;
}) {
  const effectivePreview = offlineTitle ? null : preview;
  const filePath = effectivePreview?.file.path ?? "";
  const title = effectivePreview?.file.name ?? "未选择文件";
  const canOpenFile = Boolean(effectivePreview?.file.path);
  const canDownload = Boolean(effectivePreview?.file.path);
  const editorLanguage =
    effectivePreview?.previewType === "text" && !effectivePreview.tooLarge
      ? settings.codeHighlighting
        ? languageFromPath(effectivePreview.file.path)
        : "plaintext"
      : undefined;
  const markdownAvailable = Boolean(
    effectivePreview?.previewType === "text" &&
    !effectivePreview.tooLarge &&
    isMarkdownPreviewFile(filePath),
  );
  const hideMarkdownSource = markdownAvailable && markdownMode === "preview";
  const showMarkdownPreview = markdownAvailable && markdownMode !== "source";
  const showPlainTextEditor =
    effectivePreview?.previewType === "text" &&
    !effectivePreview.tooLarge &&
    !markdownAvailable;
  const editorTheme =
    editorLanguage === "markdown" ? workbenchEditorTheme : "vs-dark";
  const icon = useMemo(() => {
    if (!effectivePreview) return <FileText />;
    if (effectivePreview.previewType === "image") return <FileImage />;
    if (
      effectivePreview.previewType === "unsupported" ||
      effectivePreview.tooLarge
    )
      return <FileX />;
    const Icon = fileIcon(effectivePreview.file.path);
    return <Icon />;
  }, [effectivePreview]);

  useEffect(() => {
    function handleSaveShortcut(event: KeyboardEvent) {
      if (
        !(event.metaKey || event.ctrlKey) ||
        event.key.toLowerCase() !== "s"
      ) {
        return;
      }
      event.preventDefault();
      if (editable && dirty && !saving) {
        onSave();
      }
    }
    window.addEventListener("keydown", handleSaveShortcut);
    return () => window.removeEventListener("keydown", handleSaveShortcut);
  }, [dirty, editable, onSave, saving]);

  return (
    <div className="workbench-detail-pane grid h-full min-h-0 grid-rows-[auto_minmax(0,1fr)]">
      <header className="workbench-detail-header flex min-h-10 items-center justify-between gap-2 border-b px-2 py-1 max-[760px]:flex-col max-[760px]:items-stretch">
        <div className="flex min-w-0 items-center gap-2">
          {icon}
          <div className="min-w-0">
            <div className="truncate text-sm font-medium">
              {dirty ? `${title} *` : title}
            </div>
            {filePath ? (
              <div className="truncate text-xs text-muted-foreground">
                {filePath}
              </div>
            ) : null}
          </div>
        </div>
        <div className="flex shrink-0 flex-wrap items-center justify-end gap-1.5 max-[760px]:justify-start">
          {effectivePreview ? (
            <>
              <Badge>{effectivePreview.file.contentType || "unknown"}</Badge>
              <Badge>{formatBytes(effectivePreview.file.size)}</Badge>
              {dirty ? <Badge tone="warning">未保存</Badge> : null}
            </>
          ) : null}
          {markdownAvailable ? (
            <div
              aria-label="Markdown 查看模式"
              className="flex items-center gap-0.5 rounded-[6px] border border-border bg-background/30 p-0.5"
              data-testid="markdown-mode-controls"
              role="group"
            >
              {markdownViewModes.map((mode) => {
                const ModeIcon = mode.icon;
                return (
                  <Button
                    key={mode.value}
                    aria-pressed={markdownMode === mode.value}
                    className="h-6 gap-1 rounded-[5px] px-2 text-xs"
                    data-testid={`markdown-mode-${mode.value}`}
                    type="button"
                    variant={
                      markdownMode === mode.value ? "secondary" : "ghost"
                    }
                    onClick={() => onMarkdownModeChange(mode.value)}
                  >
                    <ModeIcon />
                    {mode.label}
                  </Button>
                );
              })}
            </div>
          ) : null}
          <IconButton
            aria-pressed={terminalOpen}
            className="workbench-icon-button"
            data-testid="files-terminal-toggle"
            disabled={terminalToggleDisabled}
            title={terminalOpen ? "关闭终端" : "打开终端"}
            onClick={onToggleTerminalPanel}
          >
            <TerminalIcon />
          </IconButton>
          <IconButton
            className="workbench-icon-button"
            disabled={!effectivePreview || saving}
            title="重新加载"
            onClick={onReload}
          >
            <RotateCcw />
          </IconButton>
          <IconButton
            className="workbench-icon-button"
            disabled={!canOpenFile}
            title="打开文件"
            onClick={() =>
              effectivePreview && onOpenFileResource(effectivePreview.file.path)
            }
          >
            <ExternalLink />
          </IconButton>
          <IconButton
            className="workbench-icon-button"
            disabled={!canDownload}
            title="下载文件"
            onClick={() =>
              effectivePreview && onDownload(effectivePreview.file.path)
            }
          >
            <Download />
          </IconButton>
          <Button
            disabled={!editable || !dirty || saving}
            type="button"
            onClick={onSave}
          >
            <Save />
            保存
          </Button>
        </div>
      </header>
      <div
        className="workbench-editor-surface min-h-0 overflow-hidden"
        data-editor-font-size={settings.editorFontSize}
        data-editor-language={editorLanguage}
        data-editor-minimap={settings.editorMinimap}
        data-editor-tab-size={settings.editorTabSize}
        data-editor-word-wrap={settings.editorWordWrap}
      >
        {offlineTitle ? (
          <EmptyState
            icon={<HardDrive />}
            title={offlineTitle}
            description={offlineDescription}
          />
        ) : null}
        {!offlineTitle && !effectivePreview ? (
          <EmptyState icon={<FileText />} title="选择文件开始预览" />
        ) : null}
        {showPlainTextEditor ? (
          <Suspense fallback={<Spinner label="加载编辑器" />}>
            <MonacoEditor
              beforeMount={configureWorkbenchMonaco}
              height="100%"
              language={editorLanguage}
              options={{
                fontFamily: workbenchEditorFontFamily,
                fontSize: settings.editorFontSize,
                minimap: { enabled: settings.editorMinimap },
                scrollBeyondLastLine: false,
                tabSize: settings.editorTabSize,
                wordWrap: settings.editorWordWrap,
              }}
              theme={editorTheme}
              value={value}
              onChange={(nextValue) => onValueChange(nextValue ?? "")}
            />
          </Suspense>
        ) : null}
        {markdownAvailable ? (
          <div
            className={cn(
              "h-full min-h-0",
              markdownMode === "split"
                ? "grid grid-cols-[minmax(0,1fr)_minmax(0,1fr)] max-[900px]:grid-cols-1 max-[900px]:grid-rows-[minmax(0,1fr)_minmax(0,1fr)]"
                : "grid grid-cols-1",
            )}
            data-markdown-mode={markdownMode}
            data-testid="markdown-preview-layout"
          >
            {markdownAvailable ? (
              <div
                aria-hidden={hideMarkdownSource}
                className={cn(
                  "min-h-0 overflow-hidden",
                  hideMarkdownSource && "hidden",
                )}
                data-testid="markdown-source-pane"
              >
                <Suspense fallback={<Spinner label="加载编辑器" />}>
                  <MonacoEditor
                    beforeMount={configureWorkbenchMonaco}
                    height="100%"
                    language={editorLanguage}
                    options={{
                      fontFamily: workbenchEditorFontFamily,
                      fontSize: settings.editorFontSize,
                      minimap: { enabled: settings.editorMinimap },
                      scrollBeyondLastLine: false,
                      tabSize: settings.editorTabSize,
                      wordWrap: settings.editorWordWrap,
                    }}
                    theme={editorTheme}
                    value={value}
                    onChange={(nextValue) => onValueChange(nextValue ?? "")}
                  />
                </Suspense>
              </div>
            ) : null}
            {showMarkdownPreview ? (
              <MarkdownPreviewPane
                agentId={agentId}
                content={value}
                editorFontSize={settings.editorFontSize}
                filePath={filePath}
                onOpenFileResource={onOpenFileResource}
                onOpenWorkspacePath={onOpenWorkspacePath}
              />
            ) : null}
          </div>
        ) : null}
        {effectivePreview?.previewType === "image" ? (
          <div className="grid h-full place-items-center overflow-auto bg-muted/30 p-4">
            <img
              alt={effectivePreview.file.name}
              className="max-h-full max-w-full rounded-[8px] border border-border object-contain"
              src={effectivePreview.downloadUrl || imageSrc}
            />
          </div>
        ) : null}
        {effectivePreview &&
        (effectivePreview.previewType === "unsupported" ||
          effectivePreview.tooLarge) ? (
          <EmptyState
            icon={<FileX />}
            title="该文件不支持在线编辑"
            description={
              effectivePreview.tooLarge
                ? "文件超过在线编辑限制，可下载后查看。"
                : "当前类型仅支持下载查看。"
            }
          />
        ) : null}
      </div>
    </div>
  );
}

const MarkdownPreviewPane = memo(function MarkdownPreviewPane({
  agentId,
  content,
  editorFontSize,
  filePath,
  onOpenFileResource,
  onOpenWorkspacePath,
}: MarkdownPreviewPaneProps) {
  const markdownComponents = useMemo<MarkdownComponents>(
    () => ({
      pre: ({ children, node: _node, ...props }) => {
        const onlyChild =
          Array.isArray(children) && children.length === 1
            ? children[0]
            : children;
        if (isValidElement(onlyChild) && onlyChild.type === MermaidDiagram) {
          return onlyChild;
        }
        return <pre {...props}>{children}</pre>;
      },
      code: ({ children, className, node: _node, ...props }) => {
        const language = markdownCodeLanguage(className);
        if (language === "mermaid") {
          return (
            <MermaidDiagram
              fontSize={editorFontSize}
              source={markdownCodeText(children)}
            />
          );
        }
        return (
          <code {...props} className={className}>
            {children}
          </code>
        );
      },
      a: ({ children, href, node: _node, ...props }) => {
        const resolved = resolveMarkdownHref(filePath, href);
        if (resolved.kind === "blocked") {
          return (
            <span
              {...props}
              className="workbench-markdown-blocked-link"
              data-testid="markdown-blocked-link"
            >
              {children}
            </span>
          );
        }
        if (resolved.kind === "workspace") {
          return (
            <a
              {...props}
              href={resolved.href}
              onClick={(event) => {
                event.preventDefault();
                if (isWorkspaceTextPreviewLink(resolved.path)) {
                  onOpenWorkspacePath(resolved.path);
                  return;
                }
                onOpenFileResource(resolved.path);
              }}
            >
              {children}
            </a>
          );
        }
        return (
          <a
            {...props}
            href={resolved.href}
            rel={
              resolved.kind === "external" ? "noopener noreferrer" : undefined
            }
            target={resolved.kind === "external" ? "_blank" : undefined}
          >
            {children}
          </a>
        );
      },
      img: ({ alt, node: _node, src, ...props }) => {
        const resolved = resolveMarkdownHref(filePath, src);
        if (resolved.kind === "blocked" || resolved.kind === "fragment") {
          return null;
        }
        return (
          <img
            {...props}
            alt={alt ?? ""}
            loading="lazy"
            src={
              resolved.kind === "workspace"
                ? agentId
                  ? api.workspaceResourceUrl(agentId, resolved.path, "inline")
                  : ""
                : resolved.href
            }
          />
        );
      },
    }),
    [
      agentId,
      editorFontSize,
      filePath,
      onOpenFileResource,
      onOpenWorkspacePath,
    ],
  );
  const previewStyle = useMemo<MarkdownPreviewStyle>(
    () => ({
      "--markdown-font-size": `${editorFontSize}px`,
      "--markdown-line-height": `${markdownPreviewLineHeight(editorFontSize)}px`,
    }),
    [editorFontSize],
  );

  return (
    <div
      className="workbench-markdown-preview min-h-0 overflow-auto border-l border-border max-[900px]:border-l-0 max-[900px]:border-t"
      data-testid="markdown-rendered-pane"
      style={previewStyle}
    >
      <ReactMarkdown
        components={markdownComponents}
        rehypePlugins={markdownRehypePlugins}
        remarkPlugins={markdownRemarkPlugins}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}, areMarkdownPreviewPropsEqual);

const MermaidDiagram = memo(function MermaidDiagram({
  fontSize,
  source,
}: {
  fontSize: number;
  source: string;
}) {
  const diagramId = useRef(nextMermaidDiagramId());
  const [renderState, setRenderState] = useState<MermaidRenderState>({
    status: "loading",
  });

  useEffect(() => {
    const normalizedSource = source.trim();
    if (!normalizedSource) {
      setRenderState({
        status: "error",
        message: "Mermaid 图表源码为空。",
      });
      return;
    }
    if (normalizedSource.length > mermaidMaxSourceLength) {
      setRenderState({
        status: "error",
        message: "Mermaid 图表源码超过预览限制。",
      });
      return;
    }

    let cancelled = false;
    setRenderState({ status: "loading" });
    const timeout = window.setTimeout(() => {
      renderMermaidSource(
        `${diagramId.current}-${nextMermaidRenderSequence()}`,
        normalizedSource,
        fontSize,
      )
        .then((svg) => {
          if (!cancelled) {
            setRenderState({ status: "rendered", svg });
          }
        })
        .catch((error: unknown) => {
          if (!cancelled) {
            setRenderState({
              status: "error",
              message: mermaidErrorMessage(error),
            });
          }
        });
    }, mermaidRenderDebounceMs);

    return () => {
      cancelled = true;
      window.clearTimeout(timeout);
    };
  }, [fontSize, source]);

  return (
    <div
      className="workbench-mermaid-diagram"
      data-testid="markdown-mermaid-diagram"
    >
      {renderState.status === "loading" ? (
        <div
          className="workbench-mermaid-loading"
          data-testid="markdown-mermaid-loading"
        >
          图表渲染中
        </div>
      ) : null}
      {renderState.status === "rendered" ? (
        <div
          className="workbench-mermaid-svg"
          data-testid="markdown-mermaid-svg"
          dangerouslySetInnerHTML={{ __html: renderState.svg }}
        />
      ) : null}
      {renderState.status === "error" ? (
        <div
          className="workbench-mermaid-error"
          data-testid="markdown-mermaid-error"
        >
          <Alert tone="danger">
            <div className="font-semibold">Mermaid 图表渲染失败</div>
            <div>{renderState.message}</div>
            <pre className="workbench-mermaid-source">
              <code>{source}</code>
            </pre>
          </Alert>
        </div>
      ) : null}
    </div>
  );
});

function markdownCodeLanguage(className: string | undefined) {
  return (
    className?.match(/(?:^|\s)language-([^\s]+)/)?.[1]?.toLowerCase() ?? ""
  );
}

function markdownCodeText(children: unknown) {
  const value = Array.isArray(children)
    ? children.join("")
    : String(children ?? "");
  return value.replace(/\n$/, "");
}

function nextMermaidDiagramId() {
  mermaidRenderSequence += 1;
  return `workbench-mermaid-${mermaidRenderSequence}`;
}

function nextMermaidRenderSequence() {
  mermaidRenderSequence += 1;
  return mermaidRenderSequence;
}

async function renderMermaidSource(
  id: string,
  source: string,
  fontSize: number,
) {
  const [{ default: mermaid }, { default: DOMPurify }] = await Promise.all([
    import("mermaid"),
    import("dompurify"),
  ]);
  if (mermaidInitializedFontSize !== fontSize) {
    mermaid.initialize({
      startOnLoad: false,
      securityLevel: "strict",
      theme: "dark",
      fontFamily: workbenchEditorFontFamily,
      fontSize,
      themeVariables: {
        fontFamily: workbenchEditorFontFamily,
        fontSize: `${fontSize}px`,
      },
      htmlLabels: false,
      flowchart: {
        htmlLabels: false,
      },
    });
    mermaidInitializedFontSize = fontSize;
  }

  await mermaid.parse(source);
  const { svg } = await mermaid.render(id, source);
  return String(
    DOMPurify.sanitize(svg, {
      USE_PROFILES: { svg: true, svgFilters: true },
      ADD_TAGS: ["style"],
      ADD_ATTR: ["aria-roledescription", "class", "role", "style"],
      FORBID_TAGS: ["foreignObject", "script"],
    }),
  );
}

function mermaidErrorMessage(error: unknown) {
  if (error instanceof Error && error.message) {
    return truncateText(collapseWhitespace(error.message), 180);
  }
  return "Mermaid 语法无效或暂时无法渲染。";
}

function collapseWhitespace(value: string) {
  return value.replace(/\s+/g, " ").trim();
}

function truncateText(value: string, maxLength: number) {
  if (value.length <= maxLength) {
    return value;
  }
  return `${value.slice(0, maxLength)}...`;
}

function areMarkdownPreviewPropsEqual(
  prev: MarkdownPreviewPaneProps,
  next: MarkdownPreviewPaneProps,
) {
  // Callback props are intentionally ignored because FilesPanel passes stable
  // wrappers that read the latest parent state through refs.
  return (
    prev.agentId === next.agentId &&
    prev.content === next.content &&
    prev.editorFontSize === next.editorFontSize &&
    prev.filePath === next.filePath
  );
}

function markdownPreviewLineHeight(fontSize: number) {
  return Math.round(fontSize * (22 / 14));
}

function useLatestCallback<TArgs extends unknown[], TResult>(
  callback: (...args: TArgs) => TResult,
) {
  const callbackRef = useRef(callback);
  callbackRef.current = callback;
  return useCallback((...args: TArgs) => callbackRef.current(...args), []);
}

function updateTreeChildren(
  nodes: WorkspaceTreeNode[],
  targetPath: string,
  children: WorkspaceTreeNode[],
): WorkspaceTreeNode[] {
  return nodes.map((node) => {
    if (node.path === targetPath) {
      return { ...node, children };
    }
    if (node.children?.length) {
      return {
        ...node,
        children: updateTreeChildren(node.children, targetPath, children),
      };
    }
    return node;
  });
}

function collapsedPanelSizePercent(
  element: HTMLElement,
  compactLayout: boolean,
) {
  const rect = element.getBoundingClientRect();
  const availableSize = compactLayout ? rect.height : rect.width;
  if (availableSize <= 0) {
    return defaultCollapsedPanelSize;
  }
  const percent = (collapsedWorkbenchSidebarWidth / availableSize) * 100;
  return Math.min(80, Math.max(2, Number(percent.toFixed(2))));
}

function findTreeNode(
  nodes: WorkspaceTreeNode[],
  targetPath: string,
): WorkspaceTreeNode | null {
  for (const node of nodes) {
    if (node.path === targetPath) {
      return node;
    }
    const child = node.children
      ? findTreeNode(node.children, targetPath)
      : null;
    if (child) {
      return child;
    }
  }
  return null;
}

function uploadTargetPath(
  selectedNode: SelectedWorkspaceNode | null,
  workspacePath: string,
) {
  if (!selectedNode) {
    return workspacePath;
  }
  if (selectedNode.type === "directory") {
    return selectedNode.path;
  }
  return parentPath(selectedNode.path);
}

function workspaceCreateNameError(name: string) {
  if (!name || name === "." || name === "..") {
    return "请输入名称";
  }
  if (/[\\/]/.test(name) || /[\u0000\r\n\t]/.test(name)) {
    return "名称不能包含路径分隔符";
  }
  return "";
}

function basename(path: string) {
  const normalized = path.replace(/\/+$/, "");
  const index = normalized.lastIndexOf("/");
  return index >= 0 ? normalized.slice(index + 1) : normalized;
}

function parentPath(path: string) {
  const normalized = path.replace(/\/+$/, "");
  const index = normalized.lastIndexOf("/");
  if (index <= 0) {
    return workspaceRootPath;
  }
  return normalized.slice(0, index);
}

function relativeWorkspacePath(path: string, workspacePath: string) {
  const root = workspacePath.replace(/\/+$/, "");
  const normalized = path.replace(/\/+$/, "");
  if (normalized === root) {
    return "";
  }
  if (normalized.startsWith(`${root}/`)) {
    return normalized.slice(root.length + 1);
  }
  if (normalized.startsWith(`${workspaceRootPath}/`)) {
    return normalized.slice(workspaceRootPath.length + 1);
  }
  if (normalized.startsWith(`${sharedRootPath}/`)) {
    return normalized.slice(sharedRootPath.length + 1);
  }
  return normalized.replace(/^\/+/, "");
}

function isMarkdownPreviewFile(path: string) {
  return /\.(md|markdown)$/i.test(markdownPathWithoutUrlState(path));
}

function isWorkspaceTextPreviewLink(path: string) {
  const normalized = markdownPathWithoutUrlState(path).toLowerCase();
  if (
    /(\.md|\.markdown|\.mdx|\.txt|\.log|\.jsonc?|\.ya?ml|\.toml|\.ini|\.env|\.conf|\.config|\.go|\.ts|\.tsx|\.js|\.jsx|\.mjs|\.cjs|\.py|\.rs|\.java|\.kt|\.kts|\.cpp|\.c|\.h|\.hpp|\.sh|\.zsh|\.bash|\.css|\.scss|\.sass|\.less|\.html?|\.vue|\.svelte|\.sql)$/i.test(
      normalized,
    )
  ) {
    return true;
  }
  const name = basename(normalized);
  return name === "dockerfile" || name === "makefile";
}

function resolveMarkdownHref(
  currentFilePath: string,
  href: string | undefined,
): MarkdownHrefResolution {
  const value = href?.trim() ?? "";
  if (!value) {
    return { kind: "blocked" };
  }
  if (value.startsWith("#")) {
    return { kind: "fragment", href: value };
  }
  if (value.startsWith("//")) {
    return { kind: "blocked" };
  }
  const protocol = value
    .match(/^([a-zA-Z][a-zA-Z\d+.-]*):/)?.[1]
    ?.toLowerCase();
  if (protocol) {
    if (protocol === "http" || protocol === "https") {
      return { kind: "external", href: value };
    }
    return { kind: "blocked" };
  }
  const path = resolveMarkdownWorkspacePath(currentFilePath, value);
  if (!path) {
    return { kind: "blocked" };
  }
  return { kind: "workspace", href: path, path };
}

function resolveMarkdownWorkspacePath(currentFilePath: string, href: string) {
  const pathPart = markdownPathWithoutUrlState(href);
  if (!pathPart) {
    return "";
  }
  let decodedPath: string;
  try {
    decodedPath = decodeURIComponent(pathPart);
  } catch {
    return "";
  }
  const normalizedInput = decodedPath.replace(/\\/g, "/");
  const candidate = isControlledWorkspaceAbsolutePath(normalizedInput)
    ? normalizedInput
    : normalizedInput.startsWith("/")
      ? joinWorkspacePath(workspaceRootPath, normalizedInput)
      : joinWorkspacePath(parentPath(currentFilePath), normalizedInput);
  const normalizedPath = normalizeMarkdownWorkspacePath(candidate);
  return isControlledWorkspaceAbsolutePath(normalizedPath)
    ? normalizedPath
    : "";
}

function markdownPathWithoutUrlState(value: string) {
  const hashIndex = value.indexOf("#");
  const queryIndex = value.indexOf("?");
  const indexes = [hashIndex, queryIndex].filter((index) => index >= 0);
  const end = indexes.length ? Math.min(...indexes) : value.length;
  return value.slice(0, end);
}

function normalizeMarkdownWorkspacePath(path: string) {
  const parts = path.replace(/\\/g, "/").split("/");
  const stack: string[] = [];
  for (const part of parts) {
    if (!part || part === ".") {
      continue;
    }
    if (part === "..") {
      stack.pop();
      continue;
    }
    stack.push(part);
  }
  return `/${stack.join("/")}`;
}

function isControlledWorkspaceAbsolutePath(path: string) {
  return (
    path === workspaceRootPath ||
    path.startsWith(`${workspaceRootPath}/`) ||
    path === sharedRootPath ||
    path.startsWith(`${sharedRootPath}/`)
  );
}

function buildGitStateIndex(status: GitStatusResponse | null) {
  const states = new Map<string, GitVisualState>();
  if (!status?.root || (status.state !== "ok" && status.state !== "clean")) {
    return states;
  }
  for (const change of [
    ...(status.stagedChanges ?? []),
    ...(status.changes ?? []),
  ]) {
    addGitState(states, status.root, change);
  }
  for (const [path, state] of [...states.entries()]) {
    const parts = path.split("/").filter(Boolean);
    for (let i = 1; i < parts.length; i += 1) {
      const dirPath = `/${parts.slice(0, i).join("/")}`;
      const current = states.get(dirPath);
      if (
        !current ||
        gitStatusRank(state.status) > gitStatusRank(current.status)
      ) {
        states.set(dirPath, { ...state });
      } else {
        states.set(dirPath, {
          ...current,
          staged: current.staged || state.staged,
          unstaged: current.unstaged || state.unstaged,
        });
      }
    }
  }
  return states;
}

function addGitState(
  states: Map<string, GitVisualState>,
  root: string,
  change: GitChange,
) {
  const absolutePath = joinWorkspacePath(root, change.path);
  const next: GitVisualState = {
    status: normalizeGitStatus(change.status),
    staged: change.changeScope === "staged",
    unstaged: change.changeScope !== "staged",
  };
  const current = states.get(absolutePath);
  if (!current || gitStatusRank(next.status) > gitStatusRank(current.status)) {
    states.set(absolutePath, {
      ...next,
      staged: Boolean(current?.staged || next.staged),
      unstaged: Boolean(current?.unstaged || next.unstaged),
    });
    return;
  }
  states.set(absolutePath, {
    ...current,
    staged: current.staged || next.staged,
    unstaged: current.unstaged || next.unstaged,
  });
}

function joinWorkspacePath(root: string, relativePath: string) {
  return `${root.replace(/\/+$/, "")}/${relativePath.replace(/^\/+/, "")}`;
}

function normalizeGitStatus(status: string | undefined): GitStatusKind {
  const value = status || "unknown";
  if (
    value === "modified" ||
    value === "untracked" ||
    value === "deleted" ||
    value === "added" ||
    value === "renamed" ||
    value === "copied" ||
    value === "type_changed" ||
    value === "unmerged" ||
    value === "updated"
  ) {
    return value;
  }
  return "unknown";
}

function gitStatusRank(status: GitStatusKind) {
  const ranks: Record<GitStatusKind, number> = {
    unknown: 1,
    untracked: 2,
    renamed: 3,
    copied: 3,
    added: 4,
    updated: 5,
    modified: 6,
    type_changed: 7,
    deleted: 8,
    unmerged: 9,
  };
  return ranks[status];
}

function gitTextClass(status: GitStatusKind | undefined) {
  if (!status) {
    return "";
  }
  if (status === "untracked") {
    return "text-chart-2";
  }
  if (status === "added") {
    return "text-chart-3";
  }
  if (status === "deleted" || status === "unmerged") {
    return "text-destructive";
  }
  if (status === "renamed" || status === "copied") {
    return "text-primary";
  }
  return "text-chart-4";
}

function runtimeStatusLabel(value: string) {
  const labels: Record<string, string> = {
    created: "未启动",
    dead: "异常退出",
    exited: "已退出",
    missing: "容器离线",
    paused: "已暂停",
    stopped: "已停止",
    unavailable: "Docker 不可用",
  };
  return labels[value] ?? value;
}

function filesTerminalPanelContextKey(agentId: string, workspacePath: string) {
  return `${agentId}|${workspacePath}`;
}

function loadFilesTerminalPanelOpen(contextKey: string) {
  if (!contextKey || typeof window === "undefined") {
    return false;
  }
  try {
    const rawState = window.localStorage.getItem(
      filesTerminalPanelOpenStorageKey,
    );
    if (!rawState) {
      return false;
    }
    const parsed = JSON.parse(rawState) as Record<string, unknown>;
    return parsed[contextKey] === true;
  } catch {
    return false;
  }
}

function saveFilesTerminalPanelOpen(contextKey: string, open: boolean) {
  if (!contextKey || typeof window === "undefined") {
    return;
  }
  try {
    const rawState = window.localStorage.getItem(
      filesTerminalPanelOpenStorageKey,
    );
    const parsed = rawState
      ? (JSON.parse(rawState) as Record<string, unknown>)
      : {};
    const next = { ...parsed };
    if (open) {
      next[contextKey] = true;
    } else {
      delete next[contextKey];
    }
    if (Object.keys(next).length === 0) {
      window.localStorage.removeItem(filesTerminalPanelOpenStorageKey);
      return;
    }
    window.localStorage.setItem(
      filesTerminalPanelOpenStorageKey,
      JSON.stringify(next),
    );
  } catch {
    return;
  }
}

function fileIcon(path: string): ComponentType<{ className?: string }> {
  const kind = fileIconKind(path);
  if (kind === "image") return FileImage;
  if (kind === "json") return FileJson;
  if (kind === "text") return FileText;
  if (kind === "code") return FileCode2;
  if (kind === "markup") return Braces;
  if (kind === "config") return FileCog;
  if (kind === "hash") return Hash;
  return File;
}

function fileIconKind(path: string) {
  const lower = path.toLowerCase();
  if (/\.(png|jpg|jpeg|gif|webp|svg|ico)$/.test(lower)) return "image";
  if (/\.(json|jsonc)$/.test(lower)) return "json";
  if (/\.(md|mdx|txt|log)$/.test(lower)) return "text";
  if (
    /\.(go|ts|tsx|js|jsx|mjs|cjs|py|rs|java|kt|kts|cpp|c|h|hpp|sh|zsh|bash|php|rb|swift|lua|r|scala|sbt|proto|graphql|gql)$/.test(
      lower,
    )
  )
    return "code";
  if (/\.(css|scss|sass|less|html|vue|svelte)$/.test(lower)) return "markup";
  if (
    /\.(ya?ml|toml|ini|env|conf|config)$/.test(lower) ||
    lower.includes("dockerfile")
  )
    return "config";
  if (lower.startsWith("#")) return "hash";
  return "file";
}

function isHtmlFile(path: string) {
  return /\.html?$/i.test(path);
}
