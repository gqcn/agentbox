import {
  lazy,
  Suspense,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { MouseEvent as ReactMouseEvent, ReactNode } from "react";
import type { MonacoDiffEditor } from "@monaco-editor/react";
import type { editor as MonacoEditorApi } from "monaco-editor";
import type { ImperativePanelHandle } from "react-resizable-panels";
import {
  ChevronDown,
  ChevronRight,
  ChevronsDownUp,
  ChevronsUpDown,
  FileText,
  FileX,
  Folder,
  FolderOpen,
  GitBranch,
  GitPullRequest,
  LocateFixed,
  PanelLeftClose,
  PanelLeftOpen,
  PowerOff,
  RefreshCw,
  RotateCcw,
  Save,
} from "lucide-react";
import { toast } from "sonner";
import { api } from "./api";
import type {
  AgentInfo,
  AICapabilityTierInfo,
  GitChangeScope,
  GitCommitMessageSuggestionResponse,
  GitDiffResponse,
  GitFileResponse,
  GitStatusResponse,
  GitTreeNode,
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
  EmptyState,
  IconButton,
  ResizablePanel,
  ResizablePanelGroup,
  ResizeHandle,
  Spinner,
  Textarea,
  TreeButton,
} from "@/components/ui";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useMediaQuery } from "@/hooks/useMediaQuery";
import { languageFromPath } from "@/lib/editor-language";
import { loadMonacoDiffEditor, loadMonacoEditor } from "@/lib/monaco-loader";
import type { WorkbenchSettings } from "@/lib/workbench-settings";
import type { WordWrapMode } from "@/lib/workbench-settings";
import { cn, sharedRootPath, workspaceRootPath } from "@/lib/utils";

const MonacoDiffEditor = lazy(loadMonacoDiffEditor);
const MonacoEditor = lazy(loadMonacoEditor);
const collapsedWorkbenchSidebarWidth = 44;
const defaultCollapsedPanelSize = 6;
const gitCommitDraftStore = new Map<string, GitCommitDraftSnapshot>();

type Props = {
  active: boolean;
  agent?: AgentInfo;
  settings: WorkbenchSettings;
  workspacePath: string;
  canLocateInFiles?: boolean;
  locateRequest?: GitLocateRequest | null;
  onDirtyStateChange?: (state: {
    agentId: string;
    workspacePath: string;
    hasChanges: boolean;
  }) => void;
  onLocateInFiles?: (
    path: string,
    deleted?: boolean,
    type?: "file" | "directory",
  ) => void;
  onLocateRequestHandled?: (id: number) => void;
};

type OpenedGitResource =
  | { kind: "empty" }
  | { kind: "diff"; path: string; scope: GitChangeScope; diff: GitDiffResponse }
  | {
      kind: "file";
      path: string;
      file: GitFileResponse;
      draft: string;
      sourceScope?: GitChangeScope;
    };

export type GitLocateRequest = {
  id: number;
  workspacePath: string;
};

type GitCommitDraftSnapshot = {
  draft: string;
  suggestion: GitCommitMessageSuggestionResponse | null;
};

function gitCommitDraftStoreKey(agentId: string, workspacePath: string) {
  return `${agentId}\u0000${workspacePath}`;
}

type DiffEditorWordWrapSnapshot = {
  original: WordWrapMode | "inherit";
  modified: WordWrapMode | "inherit";
};

type GitContextMenuTarget =
  | { kind: "file"; node: GitTreeNode; scope: GitChangeScope }
  | { kind: "directory"; node: GitTreeNode; scope: GitChangeScope }
  | { kind: "group"; scope: GitChangeScope; nodes: GitTreeNode[] }
  | { kind: "blank" };

export default function GitPanel({
  active,
  agent,
  canLocateInFiles = true,
  settings,
  workspacePath,
  locateRequest,
  onDirtyStateChange,
  onLocateInFiles,
  onLocateRequestHandled,
}: Props) {
  const commitDraftStoreKey =
    agent?.id && workspacePath
      ? gitCommitDraftStoreKey(agent.id, workspacePath)
      : "";
  const initialCommitDraftSnapshot = commitDraftStoreKey
    ? gitCommitDraftStore.get(commitDraftStoreKey)
    : undefined;
  const [status, setStatus] = useState<GitStatusResponse | null>(null);
  const [opened, setOpened] = useState<OpenedGitResource>({ kind: "empty" });
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [committing, setCommitting] = useState(false);
  const [capabilityTiers, setCapabilityTiers] = useState<
    AICapabilityTierInfo[]
  >([]);
  const [generatingCommitMessage, setGeneratingCommitMessage] = useState(false);
  const [commitSuggestion, setCommitSuggestion] =
    useState<GitCommitMessageSuggestionResponse | null>(
      initialCommitDraftSnapshot?.suggestion ?? null,
    );
  const [commitMessageDraft, setCommitMessageDraft] = useState(
    initialCommitDraftSnapshot?.draft ?? "",
  );
  const [collapsedPaths, setCollapsedPaths] = useState<Set<string>>(
    () => new Set(),
  );
  const [diffWordWrapOverride, setDiffWordWrapOverride] =
    useState<WordWrapMode | null>(null);
  const [contextMenu, setContextMenu] = useState<{
    position: WorkspaceContextMenuPosition;
    target: GitContextMenuTarget;
  } | null>(null);
  const [sourceControlCollapsed, setSourceControlCollapsed] = useState(false);
  const [sidebarCollapsedSize, setSidebarCollapsedSize] = useState(
    defaultCollapsedPanelSize,
  );
  const compactLayout = useMediaQuery("(max-width: 760px)");
  const panelGroupShellRef = useRef<HTMLDivElement | null>(null);
  const sourceControlPanelRef = useRef<ImperativePanelHandle | null>(null);
  const restoredCommitDraftKeyRef = useRef(commitDraftStoreKey);
  const skipCommitDraftPersistRef = useRef(false);
  const sidebarDefaultSize = compactLayout ? 34 : settings.gitSidebarSize;
  const sidebarCurrentSize = sourceControlCollapsed
    ? sidebarCollapsedSize
    : sidebarDefaultSize;
  const loadedGitContextKeyRef = useRef("");

  const agentReady = Boolean(agent?.id && agent.runtimeStatus === "running");
  const gitContextKey = `${agent?.id ?? ""}|${agent?.runtimeStatus ?? ""}|${workspacePath}`;
  const basicCapability = capabilityTiers.find((tier) => tier.code === "basic");
  const basicCapabilityAvailable = Boolean(basicCapability?.available);
  const dirty =
    agentReady &&
    opened.kind === "file" &&
    opened.draft !== (opened.file.file.content || "");
  const selectedKey =
    opened.kind === "diff"
      ? `${opened.scope}:${opened.path}`
      : opened.kind === "file"
        ? `file:${opened.path}`
        : "";
  const hasStatusAlert = !workspacePath;

  useEffect(() => {
    if (!active || locateRequest) {
      return;
    }
    if (!agentReady || !workspacePath) {
      if (loadedGitContextKeyRef.current !== gitContextKey) {
        resetPanelState();
        loadedGitContextKeyRef.current = gitContextKey;
      }
      return;
    }
    if (loadedGitContextKeyRef.current !== gitContextKey) {
      void loadStatus();
    }
    void loadCapabilityTiers();
  }, [active, agentReady, gitContextKey, locateRequest, workspacePath]);

  useEffect(() => {
    if (restoredCommitDraftKeyRef.current === commitDraftStoreKey) {
      return;
    }
    restoredCommitDraftKeyRef.current = commitDraftStoreKey;
    skipCommitDraftPersistRef.current = true;
    const snapshot = commitDraftStoreKey
      ? gitCommitDraftStore.get(commitDraftStoreKey)
      : undefined;
    setCommitMessageDraft(snapshot?.draft ?? "");
    setCommitSuggestion(snapshot?.suggestion ?? null);
  }, [commitDraftStoreKey]);

  useEffect(() => {
    if (!commitDraftStoreKey) {
      return;
    }
    if (skipCommitDraftPersistRef.current) {
      skipCommitDraftPersistRef.current = false;
      return;
    }
    if (!commitMessageDraft && !commitSuggestion) {
      gitCommitDraftStore.delete(commitDraftStoreKey);
      return;
    }
    gitCommitDraftStore.set(commitDraftStoreKey, {
      draft: commitMessageDraft,
      suggestion: commitSuggestion,
    });
  }, [commitDraftStoreKey, commitMessageDraft, commitSuggestion]);

  useEffect(() => {
    if (!active || !locateRequest) {
      return;
    }
    void openWorkspacePathInGit(locateRequest);
  }, [active, locateRequest?.id]);

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
    if (sourceControlCollapsed) {
      sourceControlPanelRef.current?.collapse();
    }
  }, [sourceControlCollapsed]);

  const closeContextMenu = useCallback(() => {
    setContextMenu(null);
  }, []);

  function collapseSourceControlPane() {
    closeContextMenu();
    setSourceControlCollapsed(true);
  }

  function expandSourceControlPane() {
    sourceControlPanelRef.current?.expand(sidebarDefaultSize);
    setSourceControlCollapsed(false);
  }

  async function loadStatus() {
    const contextKey = gitContextKey;
    if (!agentReady || !agent?.id || !workspacePath) {
      resetPanelState();
      loadedGitContextKeyRef.current = contextKey;
      return;
    }
    setLoading(true);
    try {
      const nextStatus = await api.gitStatus(agent.id, workspacePath);
      setStatus(nextStatus);
      reportDirtyState(nextStatus);
      setCollapsedPaths(new Set());
      loadedGitContextKeyRef.current = contextKey;
    } catch (err) {
      const message = (err as Error).message;
      reportDirtyState(null);
      toast.error(message);
    } finally {
      setLoading(false);
    }
  }

  async function loadCapabilityTiers() {
    try {
      const items = await api.listAICapabilityTiers();
      setCapabilityTiers(items);
    } catch {
      setCapabilityTiers([]);
    }
  }

  async function generateCommitMessage() {
    if (!agentReady || !agent?.id || !workspacePath) {
      return;
    }
    if (!basicCapabilityAvailable) {
      toast.error("请先在 AI 能力页面配置基础档位");
      return;
    }
    setGeneratingCommitMessage(true);
    try {
      const suggestion = await api.gitCommitMessageSuggestion(
        agent.id,
        workspacePath,
      );
      setCommitSuggestion(suggestion);
      setCommitMessageDraft(suggestion.message);
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
    } finally {
      setGeneratingCommitMessage(false);
    }
  }

  async function openDiff(path: string, scope: GitChangeScope) {
    if (!agentReady || !agent?.id) {
      return;
    }
    if (dirty && !window.confirm("当前文件存在未保存修改，确认放弃并切换？")) {
      return;
    }
    setLoading(true);
    try {
      const diff = await api.gitDiff(agent.id, workspacePath, path, scope);
      setDiffWordWrapOverride(null);
      setOpened({ kind: "diff", path, scope, diff });
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
    } finally {
      setLoading(false);
    }
  }

  async function openFile(path: string, sourceScope?: GitChangeScope) {
    if (!agentReady || !agent?.id) {
      return;
    }
    setLoading(true);
    try {
      const file = await api.gitFile(agent.id, workspacePath, path);
      if (
        file.status === "deleted" ||
        file.file?.tooLarge ||
        file.file?.previewType !== "text"
      ) {
        setOpened({
          kind: "file",
          path,
          file,
          draft: file.file?.content || "",
          sourceScope,
        });
        return;
      }
      setOpened({
        kind: "file",
        path,
        file,
        draft: file.file.content || "",
        sourceScope,
      });
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
    } finally {
      setLoading(false);
    }
  }

  async function saveOpenedFile() {
    if (
      !agentReady ||
      !agent?.id ||
      opened.kind !== "file" ||
      opened.file.status === "deleted"
    ) {
      return;
    }
    setSaving(true);
    try {
      const saved = await api.saveWorkspaceFile(agent.id, {
        path: opened.file.file.file.path,
        content: opened.draft,
        encoding: opened.file.file.encoding || "utf-8",
        baseHash: opened.file.file.contentHash,
      });
      setOpened({
        kind: "file",
        path: opened.path,
        file: { ...opened.file, file: saved },
        draft: saved.content || "",
        sourceScope: opened.sourceScope,
      });
      const nextStatus = await api.gitStatus(agent.id, workspacePath);
      setStatus(nextStatus);
      loadedGitContextKeyRef.current = gitContextKey;
      reportDirtyState(nextStatus);
      if (opened.sourceScope) {
        const scope = opened.sourceScope;
        const diff = await api.gitDiff(
          agent.id,
          workspacePath,
          opened.path,
          scope,
        );
        setOpened({ kind: "diff", path: opened.path, scope, diff });
      }
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
    } finally {
      setSaving(false);
    }
  }

  async function reloadOpenedFile() {
    if (opened.kind !== "file") {
      return;
    }
    await openFile(opened.path, opened.sourceScope);
  }

  async function stage(path: string) {
    if (!agentReady || !agent?.id) {
      return;
    }
    setLoading(true);
    try {
      const nextStatus = await api.stageGitFiles(agent.id, workspacePath, [
        path,
      ]);
      setStatus(nextStatus);
      reportDirtyState(nextStatus);
      if (opened.kind === "diff") {
        await openDiff(path, "staged");
      }
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
    } finally {
      setLoading(false);
    }
  }

  async function stageMany(paths: string[]) {
    if (!agentReady || !agent?.id || paths.length === 0) {
      return;
    }
    setLoading(true);
    try {
      const nextStatus = await api.stageGitFiles(
        agent.id,
        workspacePath,
        paths,
      );
      setStatus(nextStatus);
      reportDirtyState(nextStatus);
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
    } finally {
      setLoading(false);
    }
  }

  async function unstage(path: string) {
    if (!agentReady || !agent?.id) {
      return;
    }
    setLoading(true);
    try {
      const nextStatus = await api.unstageGitFiles(agent.id, workspacePath, [
        path,
      ]);
      setStatus(nextStatus);
      reportDirtyState(nextStatus);
      if (opened.kind === "diff") {
        await openDiff(path, "unstaged");
      }
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
    } finally {
      setLoading(false);
    }
  }

  async function unstageMany(paths: string[]) {
    if (!agentReady || !agent?.id || paths.length === 0) {
      return;
    }
    setLoading(true);
    try {
      const nextStatus = await api.unstageGitFiles(
        agent.id,
        workspacePath,
        paths,
      );
      setStatus(nextStatus);
      reportDirtyState(nextStatus);
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
    } finally {
      setLoading(false);
    }
  }

  async function stageAll() {
    if (!agentReady || !agent?.id) {
      return;
    }
    setLoading(true);
    try {
      const nextStatus = await api.stageAllGitFiles(agent.id, workspacePath);
      setStatus(nextStatus);
      reportDirtyState(nextStatus);
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
    } finally {
      setLoading(false);
    }
  }

  async function unstageAll() {
    if (!agentReady || !agent?.id) {
      return;
    }
    setLoading(true);
    try {
      const nextStatus = await api.unstageAllGitFiles(agent.id, workspacePath);
      setStatus(nextStatus);
      reportDirtyState(nextStatus);
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
    } finally {
      setLoading(false);
    }
  }

  async function discardMany(paths: string[]) {
    if (!agentReady || !agent?.id || paths.length === 0) {
      return;
    }
    const targetLabel = paths.length === 1 ? paths[0] : `${paths.length} 个变更`;
    if (
      !window.confirm(
        `确认取消 ${targetLabel} 的未暂存变更？该操作会还原已跟踪文件或删除未跟踪文件。`,
      )
    ) {
      return;
    }
    setLoading(true);
    try {
      const nextStatus = await api.discardGitFiles(agent.id, workspacePath, paths);
      setStatus(nextStatus);
      reportDirtyState(nextStatus);
      if (opened.kind === "diff" && opened.scope === "unstaged" && paths.includes(opened.path)) {
        setOpened({ kind: "empty" });
      }
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
    } finally {
      setLoading(false);
    }
  }

  async function commitChanges(push = false) {
    if (!agentReady || !agent?.id || committing) {
      return;
    }
    const message = commitMessageDraft.trim();
    if (!message) {
      toast.error("提交信息为空");
      return;
    }
    setCommitting(true);
    try {
      const result = await api.gitCommit(agent.id, workspacePath, message, push);
      setStatus(result.status);
      reportDirtyState(result.status);
      setCommitMessageDraft("");
      setCommitSuggestion(null);
      if (commitDraftStoreKey) {
        gitCommitDraftStore.delete(commitDraftStoreKey);
      }
    } catch (err) {
      const errorMessage = (err as Error).message;
      toast.error(errorMessage);
    } finally {
      setCommitting(false);
    }
  }

  function toggleDiffWordWrap() {
    setDiffWordWrapOverride((current) => {
      const base = current ?? settings.editorWordWrap;
      return base === "on" ? "off" : "on";
    });
  }

  function openContextMenu(
    event: ReactMouseEvent,
    target: GitContextMenuTarget,
  ) {
    event.preventDefault();
    event.stopPropagation();
    setContextMenu({
      position: { x: event.clientX, y: event.clientY },
      target,
    });
  }

  function expandAll(scope?: GitChangeScope) {
    setCollapsedPaths((current) => {
      if (!scope) {
        return new Set();
      }
      const next = new Set(current);
      for (const key of [...next]) {
        if (key.startsWith(`${scope}:`)) {
          next.delete(key);
        }
      }
      return next;
    });
  }

  function collapseAll(scope?: GitChangeScope) {
    const next = new Set<string>();
    if (!scope || scope === "staged") {
      collectDirectoryKeys(status?.stagedTree ?? [], "staged", next);
    }
    if (!scope || scope === "unstaged") {
      collectDirectoryKeys(status?.changeTree ?? [], "unstaged", next);
    }
    setCollapsedPaths(next);
  }

  async function openWorkspacePathInGit(request: GitLocateRequest) {
    if (!agentReady || !agent?.id || !workspacePath) {
      onLocateRequestHandled?.(request.id);
      return;
    }
    try {
      const nextStatus = await api.gitStatus(agent.id, workspacePath);
      setStatus(nextStatus);
      loadedGitContextKeyRef.current = gitContextKey;
      reportDirtyState(nextStatus);
      const relativePath = relativeGitPath(
        request.workspacePath,
        nextStatus.root || workspacePath,
      );
      const unstaged = findGitNode(nextStatus.changeTree ?? [], relativePath);
      const staged = findGitNode(nextStatus.stagedTree ?? [], relativePath);
      if (unstaged?.type === "file") {
        await openDiff(unstaged.path, "unstaged");
        return;
      }
      if (staged?.type === "file") {
        await openDiff(staged.path, "staged");
        return;
      }
      toast.info("该文件当前没有 Git 变化");
    } finally {
      onLocateRequestHandled?.(request.id);
    }
  }

  function buildContextMenuItems(): WorkspaceContextMenuEntry[] {
    if (!contextMenu) {
      return [];
    }
    const gitReady =
      agentReady && Boolean(workspacePath) && status?.state === "ok";
    if (contextMenu.target.kind === "blank") {
      return [
        {
          id: "refresh",
          label: "刷新",
          icon: <RefreshCw />,
          disabled: !agentReady,
          testId: "git-context-refresh",
          onSelect: () => void loadStatus(),
        },
        { kind: "separator", id: "blank-separator" },
        {
          id: "expand-all",
          label: "全部展开",
          icon: <ChevronsUpDown />,
          disabled: !gitReady,
          testId: "git-context-expand-all",
          onSelect: () => expandAll(),
        },
        {
          id: "collapse-all",
          label: "全部折叠",
          icon: <ChevronsDownUp />,
          disabled: !gitReady,
          testId: "git-context-collapse-all",
          onSelect: () => collapseAll(),
        },
      ];
    }
    if (contextMenu.target.kind === "group") {
      const target = contextMenu.target;
      const paths = collectFilePaths(target.nodes);
      const staged = target.scope === "staged";
      return [
        {
          id: staged ? "unstage-all" : "stage-all",
          label: staged ? "取消暂存全部 Staged Changes" : "暂存全部 Changes",
          icon: staged ? <Codicon name="remove" /> : <Codicon name="add" />,
          disabled: !gitReady || paths.length === 0,
          testId: staged ? "git-context-unstage-all" : "git-context-stage-all",
          onSelect: () => (staged ? void unstageAll() : void stageAll()),
        },
        ...(staged
          ? []
          : [
              {
                id: "discard-all",
                label: "取消全部 Changes",
                icon: <Codicon name="discard" />,
                disabled: !gitReady || paths.length === 0,
                testId: "git-context-discard-all",
                onSelect: () => void discardMany(paths),
              } satisfies WorkspaceContextMenuEntry,
            ]),
        { kind: "separator", id: "group-separator" },
        {
          id: "expand-group",
          label: "全部展开",
          icon: <ChevronsUpDown />,
          disabled: !gitReady,
          testId: "git-context-expand-group",
          onSelect: () => expandAll(target.scope),
        },
        {
          id: "collapse-group",
          label: "全部折叠",
          icon: <ChevronsDownUp />,
          disabled: !gitReady,
          testId: "git-context-collapse-group",
          onSelect: () => collapseAll(target.scope),
        },
        {
          id: "refresh",
          label: "刷新",
          icon: <RefreshCw />,
          disabled: !agentReady,
          testId: "git-context-refresh",
          onSelect: () => void loadStatus(),
        },
      ];
    }
    const target = contextMenu.target;
    const node = target.node;
    const staged = target.scope === "staged";
    const absolutePath = joinWorkspacePath(
      status?.root || workspacePath,
      node.path,
    );
    const paths = collectFilePaths([node]);
    if (target.kind === "directory") {
      const expanded = !collapsedPaths.has(
        gitDirectoryKey(target.scope, node.path),
      );
      return [
        {
          id: "toggle-directory",
          label: expanded ? "折叠" : "展开",
          icon: expanded ? <ChevronDown /> : <ChevronRight />,
          disabled: !gitReady,
          testId: "git-context-toggle-directory",
          onSelect: () => toggleGitDirectory(target.scope, node.path),
        },
        {
          id: staged ? "unstage-directory" : "stage-directory",
          label: staged ? "取消暂存此目录变化" : "暂存此目录变化",
          icon: staged ? <Codicon name="remove" /> : <Codicon name="add" />,
          disabled: !gitReady || paths.length === 0,
          testId: staged
            ? "git-context-unstage-directory"
            : "git-context-stage-directory",
          onSelect: () =>
            staged ? void unstageMany(paths) : void stageMany(paths),
        },
        ...(staged
          ? []
          : [
              {
                id: "discard-directory",
                label: "取消此目录变化",
                icon: <Codicon name="discard" />,
                disabled: !gitReady || paths.length === 0,
                testId: "git-context-discard-directory",
                onSelect: () => void discardMany(paths),
              } satisfies WorkspaceContextMenuEntry,
            ]),
        { kind: "separator", id: "directory-separator" },
        ...(canLocateInFiles
          ? [
              {
                id: "locate-files",
                label: "在 Explorer 中定位",
                icon: <LocateFixed />,
                disabled: !workspacePath,
                testId: "git-context-locate-files",
                onSelect: () => onLocateInFiles?.(absolutePath, false, "directory"),
              } satisfies WorkspaceContextMenuEntry,
            ]
          : []),
        {
          id: "copy-path",
          label: "复制路径",
          icon: <Codicon name="copy" />,
          testId: "git-context-copy-path",
          onSelect: () => copyWorkspaceText(node.path, "路径已复制"),
        },
        {
          id: "refresh",
          label: "刷新",
          icon: <RefreshCw />,
          disabled: !agentReady,
          testId: "git-context-refresh",
          onSelect: () => void loadStatus(),
        },
      ];
    }
    const deleted = node.status === "deleted";
    return [
      {
        id: "open-diff",
        label: "打开 diff",
        icon: <Codicon name="diff" />,
        disabled: !gitReady,
        testId: "git-context-open-diff",
        onSelect: () => void openDiff(node.path, target.scope),
      },
      {
        id: "open-file",
        label: "打开文件",
        icon: <FileText />,
        disabled: !gitReady,
        testId: "git-context-open-file",
        onSelect: () => void openFile(node.path, target.scope),
      },
      {
        id: staged ? "unstage-file" : "stage-file",
        label: staged ? "取消暂存" : "暂存",
        icon: staged ? <Codicon name="remove" /> : <Codicon name="add" />,
        disabled: !gitReady,
        testId: staged ? "git-context-unstage-file" : "git-context-stage-file",
        onSelect: () =>
          staged ? void unstage(node.path) : void stage(node.path),
      },
      ...(staged
        ? []
        : [
            {
              id: "discard-file",
              label: "取消变更",
              icon: <Codicon name="discard" />,
              disabled: !gitReady,
              testId: "git-context-discard-file",
              onSelect: () => void discardMany([node.path]),
            } satisfies WorkspaceContextMenuEntry,
          ]),
      { kind: "separator", id: "file-separator" },
      ...(canLocateInFiles
        ? [
            {
              id: "locate-files",
              label: "在 Explorer 中定位",
              icon: <LocateFixed />,
              disabled: !workspacePath,
              testId: "git-context-locate-files",
              onSelect: () => onLocateInFiles?.(absolutePath, deleted, "file"),
            } satisfies WorkspaceContextMenuEntry,
          ]
        : []),
      {
        id: "copy-relative",
        label: "复制仓库相对路径",
        icon: <Codicon name="copy" />,
        testId: "git-context-copy-relative",
        onSelect: () => copyWorkspaceText(node.path, "仓库相对路径已复制"),
      },
      {
        id: "copy-absolute",
        label: "复制绝对路径",
        icon: <Codicon name="copy" />,
        testId: "git-context-copy-absolute",
        onSelect: () =>
          copyWorkspaceText(absolutePath, "workspace 绝对路径已复制"),
      },
      {
        id: "refresh",
        label: "刷新",
        icon: <RefreshCw />,
        disabled: !agentReady,
        testId: "git-context-refresh",
        onSelect: () => void loadStatus(),
      },
    ];
  }

  return (
    <section
      className={cn(
        "grid h-full min-h-0 grid-rows-[auto_minmax(0,1fr)] overflow-hidden max-[760px]:h-[720px] max-[760px]:min-h-[720px]",
        hasStatusAlert ? "gap-1" : "gap-0",
        active ? "" : "hidden",
      )}
      data-testid="git-panel"
    >
      <div className="grid gap-2">
        {!workspacePath ? (
          <Alert>选择 {workspaceRootPath} 或 {sharedRootPath} 下的 Git 仓库目录后展示变化。</Alert>
        ) : null}
      </div>
      <div
        ref={panelGroupShellRef}
        className="h-full min-h-0 overflow-hidden"
      >
        <ResizablePanelGroup
          key={`${compactLayout ? "git-compact" : "git-wide"}-${sidebarDefaultSize}`}
          className="workspace-dark-plus h-full min-h-0 overflow-hidden bg-card"
          data-sidebar-collapsed={sourceControlCollapsed}
          data-sidebar-default-size={sidebarDefaultSize}
          data-workbench-theme="vscode-dark-plus"
          direction={compactLayout ? "vertical" : "horizontal"}
        >
          <ResizablePanel
            ref={sourceControlPanelRef}
            collapsible={sourceControlCollapsed}
            collapsedSize={sidebarCollapsedSize}
            defaultSize={sidebarCurrentSize}
            minSize={sourceControlCollapsed ? sidebarCollapsedSize : 0}
            onCollapse={() => setSourceControlCollapsed(true)}
            onExpand={() => setSourceControlCollapsed(false)}
          >
            {sourceControlCollapsed ? (
              <div
                className="workbench-sidebar-pane flex h-full min-h-0 items-start justify-center pt-2"
                data-testid="git-source-control-collapsed-rail"
              >
                <IconButton
                  className="workbench-icon-button"
                  data-testid="git-source-control-expand-button"
                  title="展开 Source Control"
                  onClick={expandSourceControlPane}
                >
                  <PanelLeftOpen />
                </IconButton>
              </div>
            ) : (
              <SourceControlPane
                agentReady={agentReady}
                basicCapabilityAvailable={basicCapabilityAvailable}
                collapsedPaths={collapsedPaths}
                commitMessageDraft={commitMessageDraft}
                committing={committing}
                generatingCommitMessage={generatingCommitMessage}
                loading={loading}
                selectedKey={selectedKey}
                status={status}
                onCollapseSidebar={collapseSourceControlPane}
                onCommit={() => void commitChanges()}
                onCommitPush={() => void commitChanges(true)}
                onCommitMessageChange={setCommitMessageDraft}
                onContextMenu={openContextMenu}
                onDiscard={(paths) => void discardMany(paths)}
                onDiff={(path, scope) => void openDiff(path, scope)}
                onGenerateCommitMessage={() => void generateCommitMessage()}
                onOpenFile={(path, scope) => void openFile(path, scope)}
                onRefresh={() => void loadStatus()}
                onStage={(path) => void stage(path)}
                onStageAll={() => void stageAll()}
                onToggleDirectory={(scope, path) =>
                  toggleGitDirectory(scope, path)
                }
                onUnstage={(path) => void unstage(path)}
                onUnstageAll={() => void unstageAll()}
              />
            )}
          </ResizablePanel>
          <ResizeHandle
            className="workbench-resize-handle"
            data-testid="git-resize-handle"
          />
          <ResizablePanel
            defaultSize={100 - sidebarCurrentSize}
            minSize={compactLayout ? 42 : 42}
          >
            <GitDetailPane
              compactLayout={compactLayout}
              dirty={dirty}
              offlineDescription={
                agent
                  ? `当前状态：${runtimeStatusLabel(agent.runtimeStatus || "unknown")}。启动后即可查看 Git 变化。`
                  : "请选择一个运行中的智能体。"
              }
              offlineTitle={agentReady ? "" : "当前 Agent 未运行"}
              opened={opened}
              saving={saving}
              settings={settings}
              wordWrapOverride={diffWordWrapOverride}
              onDraftChange={(draft) =>
                setOpened((current) =>
                  current.kind === "file" ? { ...current, draft } : current,
                )
              }
              onOpenFile={(path, sourceScope) =>
                void openFile(path, sourceScope)
              }
              onReloadFile={() => void reloadOpenedFile()}
              onSaveFile={() => void saveOpenedFile()}
              onToggleDiffWordWrap={toggleDiffWordWrap}
            />
          </ResizablePanel>
        </ResizablePanelGroup>
      </div>
      <WorkspaceContextMenu
        items={buildContextMenuItems()}
        label="Source Control 右键菜单"
        position={contextMenu?.position ?? null}
        testId="git-context-menu"
        onClose={closeContextMenu}
      />
    </section>
  );

  function toggleGitDirectory(scope: GitChangeScope, path: string) {
    const key = gitDirectoryKey(scope, path);
    setCollapsedPaths((current) => {
      const next = new Set(current);
      if (next.has(key)) {
        next.delete(key);
      } else {
        next.add(key);
      }
      return next;
    });
  }

  function resetPanelState() {
    setStatus(null);
    setOpened({ kind: "empty" });
    setLoading(false);
    setSaving(false);
    setCollapsedPaths(new Set());
    if (agent?.id && workspacePath) {
      reportDirtyState(null);
    }
  }

  function reportDirtyState(nextStatus: GitStatusResponse | null) {
    if (!agent?.id || !workspacePath) {
      return;
    }
    onDirtyStateChange?.({
      agentId: agent.id,
      workspacePath,
      hasChanges: hasGitChanges(nextStatus),
    });
  }
}

export function hasGitChanges(status: GitStatusResponse | null | undefined) {
  if (status?.state !== "ok") {
    return false;
  }
  return (
    (status.changes?.length ?? 0) > 0 ||
    (status.stagedChanges?.length ?? 0) > 0 ||
    countFiles(status.changeTree ?? []) > 0 ||
    countFiles(status.stagedTree ?? []) > 0
  );
}

function SourceControlPane({
  agentReady,
  basicCapabilityAvailable,
  collapsedPaths,
  commitMessageDraft,
  committing,
  generatingCommitMessage,
  loading,
  selectedKey,
  status,
  onCommit,
  onCommitPush,
  onCommitMessageChange,
  onCollapseSidebar,
  onContextMenu,
  onDiscard,
  onDiff,
  onGenerateCommitMessage,
  onOpenFile,
  onRefresh,
  onStage,
  onStageAll,
  onToggleDirectory,
  onUnstage,
  onUnstageAll,
}: {
  agentReady: boolean;
  basicCapabilityAvailable: boolean;
  collapsedPaths: Set<string>;
  commitMessageDraft: string;
  committing: boolean;
  generatingCommitMessage: boolean;
  loading: boolean;
  selectedKey: string;
  status: GitStatusResponse | null;
  onCommit: () => void;
  onCommitPush: () => void;
  onCommitMessageChange: (value: string) => void;
  onCollapseSidebar: () => void;
  onContextMenu: (event: ReactMouseEvent, target: GitContextMenuTarget) => void;
  onDiscard: (paths: string[]) => void;
  onDiff: (path: string, scope: GitChangeScope) => void;
  onGenerateCommitMessage: () => void;
  onOpenFile: (path: string, scope?: GitChangeScope) => void;
  onRefresh: () => void;
  onStage: (path: string) => void;
  onStageAll: () => void;
  onToggleDirectory: (scope: GitChangeScope, path: string) => void;
  onUnstage: (path: string) => void;
  onUnstageAll: () => void;
}) {
  const stagedTree = status?.stagedTree ?? [];
  const changeTree = status?.changeTree ?? [];
  const empty = !loading && stagedTree.length === 0 && changeTree.length === 0;
  const hasStagedChanges = countFiles(stagedTree) > 0;
  return (
    <div className="workbench-sidebar-pane grid h-full min-h-0 grid-rows-[auto_minmax(0,1fr)]">
      <div className="workbench-pane-title flex min-h-10 items-center justify-between gap-2 border-b px-3 text-xs font-medium uppercase text-muted-foreground">
        <div className="min-w-0">
          <span>Source Control</span>
          {status?.root ? (
            <div className="truncate normal-case text-[11px] font-normal text-muted-foreground">
              仓库 {status.root}
            </div>
          ) : null}
        </div>
        <div
          className="flex items-center gap-1"
          data-testid="git-source-control-actions"
        >
          <IconButton
            className="workbench-icon-button"
            data-testid="git-source-control-refresh-button"
            disabled={!agentReady || loading}
            title="刷新 Source Control"
            onClick={onRefresh}
          >
            {loading ? (
              <RefreshCw className="animate-spin" />
            ) : (
              <Codicon name="refresh" />
            )}
          </IconButton>
          <IconButton
            className="workbench-icon-button"
            data-testid="git-source-control-collapse-button"
            title="折叠 Source Control"
            onClick={onCollapseSidebar}
          >
            <PanelLeftClose />
          </IconButton>
        </div>
      </div>
      <div
        className="workbench-pane-scroll min-h-0 overflow-auto py-1"
        data-testid="git-source-control-scroll"
        onContextMenu={(event) => onContextMenu(event, { kind: "blank" })}
      >
        {!agentReady ? (
          <EmptyState icon={<PowerOff />} title="Agent 未运行" />
        ) : null}
        {agentReady && !basicCapabilityAvailable ? (
          <div
            className="mx-2 mb-2 rounded-[6px] border border-chart-4/30 bg-chart-4/10 px-3 py-2 text-xs text-foreground"
            data-testid="git-basic-capability-warning"
          >
            基础 AI 能力未配置，无法生成提交信息。
          </div>
        ) : null}
        {agentReady ? (
          <CommitBox
            basicCapabilityAvailable={basicCapabilityAvailable}
            committing={committing}
            draft={commitMessageDraft}
            generating={generatingCommitMessage}
            hasStagedChanges={hasStagedChanges}
            onChange={onCommitMessageChange}
            onCommit={onCommit}
            onCommitPush={onCommitPush}
            onGenerate={onGenerateCommitMessage}
          />
        ) : null}
        {agentReady && empty ? (
          <EmptyState icon={<GitBranch />} title="暂无 Git 变化" />
        ) : null}
        {agentReady ? (
          <>
            <GitGroup
              actionLabel="取消暂存"
              actionIcon={<Codicon name="remove" />}
              collapsedPaths={collapsedPaths}
              nodes={stagedTree}
              scope="staged"
              selectedKey={selectedKey}
              title="Staged Changes"
              onAction={onUnstage}
              onContextMenu={onContextMenu}
              onDiscard={onDiscard}
              onDiff={onDiff}
              onOpenFile={onOpenFile}
              onStageAll={onStageAll}
              onToggleDirectory={onToggleDirectory}
              onUnstageAll={onUnstageAll}
            />
            <GitGroup
              actionLabel="暂存"
              actionIcon={<Codicon name="add" />}
              collapsedPaths={collapsedPaths}
              nodes={changeTree}
              scope="unstaged"
              selectedKey={selectedKey}
              title="Changes"
              onAction={onStage}
              onContextMenu={onContextMenu}
              onDiscard={onDiscard}
              onDiff={onDiff}
              onOpenFile={onOpenFile}
              onStageAll={onStageAll}
              onToggleDirectory={onToggleDirectory}
              onUnstageAll={onUnstageAll}
            />
          </>
        ) : null}
      </div>
    </div>
  );
}

function GitGroup({
  actionIcon,
  actionLabel,
  collapsedPaths,
  nodes,
  scope,
  selectedKey,
  title,
  onAction,
  onContextMenu,
  onDiscard,
  onDiff,
  onOpenFile,
  onStageAll,
  onToggleDirectory,
  onUnstageAll,
}: {
  actionIcon: ReactNode;
  actionLabel: string;
  collapsedPaths: Set<string>;
  nodes: GitTreeNode[];
  scope: GitChangeScope;
  selectedKey: string;
  title: string;
  onAction: (path: string) => void;
  onContextMenu: (event: ReactMouseEvent, target: GitContextMenuTarget) => void;
  onDiscard: (paths: string[]) => void;
  onDiff: (path: string, scope: GitChangeScope) => void;
  onOpenFile: (path: string, scope?: GitChangeScope) => void;
  onStageAll: () => void;
  onToggleDirectory: (scope: GitChangeScope, path: string) => void;
  onUnstageAll: () => void;
}) {
  const fileCount = countFiles(nodes);
  return (
    <section className="grid gap-0 py-0.5">
      <div
        className="workbench-section-title flex h-6 items-center justify-between px-2 text-xs font-semibold text-muted-foreground"
        data-testid={`git-group-${scope}`}
        onContextMenu={(event) =>
          onContextMenu(event, { kind: "group", nodes, scope })
        }
      >
        <span>{title}</span>
        <div className="flex items-center gap-0.5">
          {scope === "unstaged" ? (
            <>
              <IconButton
                className="workbench-icon-button"
                disabled={fileCount === 0}
                title="取消所有变更"
                onClick={() => onDiscard(collectFilePaths(nodes))}
              >
                <Codicon name="discard" />
              </IconButton>
              <IconButton
                className="workbench-icon-button"
                disabled={fileCount === 0}
                title="添加所有变更"
                onClick={onStageAll}
              >
                <Codicon name="add" />
              </IconButton>
            </>
          ) : (
            <IconButton
              className="workbench-icon-button"
              disabled={fileCount === 0}
              title="取消所有暂存"
              onClick={onUnstageAll}
            >
              <Codicon name="remove" />
            </IconButton>
          )}
          <Badge>{fileCount}</Badge>
        </div>
      </div>
      {nodes.length === 0 ? (
        <div className="px-2 py-0.5 text-xs text-muted-foreground">无变化</div>
      ) : null}
      <GitTree
        actionIcon={actionIcon}
        actionLabel={actionLabel}
        collapsedPaths={collapsedPaths}
        nodes={nodes}
        scope={scope}
        selectedKey={selectedKey}
        onAction={onAction}
        onContextMenu={onContextMenu}
        onDiscard={onDiscard}
        onDiff={onDiff}
        onOpenFile={onOpenFile}
        onToggleDirectory={onToggleDirectory}
      />
    </section>
  );
}

function CommitBox({
  basicCapabilityAvailable,
  committing,
  draft,
  generating,
  hasStagedChanges,
  onChange,
  onCommit,
  onCommitPush,
  onGenerate,
}: {
  basicCapabilityAvailable: boolean;
  committing: boolean;
  draft: string;
  generating: boolean;
  hasStagedChanges: boolean;
  onChange: (value: string) => void;
  onCommit: () => void;
  onCommitPush: () => void;
  onGenerate: () => void;
}) {
  const canCommit = hasStagedChanges && draft.trim().length > 0 && !committing;
  const generateTitle = basicCapabilityAvailable
    ? "生成提交信息"
    : "配置 AI 能力基础档位后可生成提交信息";
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    const input = textareaRef.current;
    if (!input) {
      return;
    }
    const maxHeight = 154;
    input.style.height = "0px";
    const nextHeight = Math.min(Math.max(input.scrollHeight, 34), maxHeight);
    input.style.height = `${nextHeight}px`;
    input.style.overflowY = input.scrollHeight > maxHeight ? "auto" : "hidden";
  }, [draft, generating]);

  return (
    <div
      className="workbench-commit-box mx-2 mb-2 grid gap-2"
      data-testid="git-commit-box"
    >
      <div className="workbench-commit-input-wrap">
        <Textarea
          ref={textareaRef}
          className="workbench-commit-input"
          disabled={generating}
          placeholder={generating ? "正在生成提交信息..." : 'Message (⌘Enter to commit)'}
          rows={1}
          value={draft}
          onChange={(event) => onChange(event.target.value)}
          onKeyDown={(event) => {
            if ((event.metaKey || event.ctrlKey) && event.key === "Enter") {
              event.preventDefault();
              if (canCommit) {
                onCommit();
              }
            }
          }}
        />
        <IconButton
          className="workbench-icon-button workbench-commit-generate"
          disabled={!basicCapabilityAvailable || generating}
          title={generateTitle}
          onClick={onGenerate}
        >
          <Codicon
            className={generating ? "codicon-modifier-spin" : ""}
            name={generating ? "loading" : "sparkle"}
          />
        </IconButton>
      </div>
      <div className="workbench-commit-split-button" data-testid="git-commit-split-button">
        <Button
          className="workbench-commit-button workbench-commit-primary"
          disabled={!canCommit}
          type="button"
          onClick={onCommit}
        >
          <Codicon name="check" />
          Commit
        </Button>
        <DropdownMenu>
          <DropdownMenuTrigger
            render={
              <Button
                aria-label="打开 Commit 操作菜单"
                className="workbench-commit-menu-trigger"
                disabled={!canCommit}
                title="打开 Commit 操作菜单"
                type="button"
              >
                <ChevronDown aria-hidden="true" />
              </Button>
            }
          />
          <DropdownMenuContent
            align="end"
            className="workbench-commit-menu w-56"
            sideOffset={4}
          >
            <DropdownMenuGroup>
              <DropdownMenuItem
                disabled={!canCommit}
                data-testid="git-commit-menu-commit"
                onClick={onCommit}
              >
                <Codicon name="check" />
                Commit
              </DropdownMenuItem>
              <DropdownMenuItem
                disabled={!canCommit}
                data-testid="git-commit-menu-commit-staged"
                onClick={onCommit}
              >
                <Codicon name="git-commit" />
                Commit Staged
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                disabled={!canCommit}
                data-testid="git-commit-menu-push"
                onClick={onCommitPush}
              >
                <Codicon name="cloud-upload" />
                Commit & Push
              </DropdownMenuItem>
              <DropdownMenuItem disabled data-testid="git-commit-menu-sync">
                <Codicon name="sync" />
                Commit & Sync
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem disabled data-testid="git-commit-menu-amend">
                <Codicon name="history" />
                Commit Staged (Amend)
              </DropdownMenuItem>
              <DropdownMenuItem disabled data-testid="git-commit-menu-undo">
                <Codicon name="discard" />
                Undo Last Commit
              </DropdownMenuItem>
            </DropdownMenuGroup>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
      {!hasStagedChanges ? (
        <div className="text-[11px] text-muted-foreground">暂存变更后可提交。</div>
      ) : null}
    </div>
  );
}

function GitTree({
  actionIcon,
  actionLabel,
  collapsedPaths,
  nodes,
  level = 0,
  scope,
  selectedKey,
  onAction,
  onContextMenu,
  onDiscard,
  onDiff,
  onOpenFile,
  onToggleDirectory,
}: {
  actionIcon: ReactNode;
  actionLabel: string;
  collapsedPaths: Set<string>;
  nodes: GitTreeNode[];
  level?: number;
  scope: GitChangeScope;
  selectedKey: string;
  onAction: (path: string) => void;
  onContextMenu: (event: ReactMouseEvent, target: GitContextMenuTarget) => void;
  onDiscard: (paths: string[]) => void;
  onDiff: (path: string, scope: GitChangeScope) => void;
  onOpenFile: (path: string, scope?: GitChangeScope) => void;
  onToggleDirectory: (scope: GitChangeScope, path: string) => void;
}) {
  return (
    <div className="grid gap-0">
      {nodes.map((node) => {
        const active = selectedKey === `${scope}:${node.path}`;
        const expanded =
          node.type !== "directory" ||
          !collapsedPaths.has(gitDirectoryKey(scope, node.path));
        const hasChildren = Boolean(node.children?.length);
        return (
          <div key={`${scope}:${node.path}`}>
            <div
              className={cn(
                "workbench-tree-row grid grid-cols-[minmax(0,1fr)_auto] items-center gap-1.5 pr-2 text-[13px] hover:bg-accent",
                active && "is-active bg-primary/10 text-foreground hover:bg-primary/15",
              )}
              data-testid={`git-node-${scope}-${node.path}`}
              data-tree-state={
                node.type === "directory"
                  ? expanded
                    ? "expanded"
                    : "collapsed"
                  : undefined
              }
              style={{ paddingLeft: 8 + level * 18 }}
              onContextMenu={(event) =>
                onContextMenu(event, {
                  kind: node.type === "directory" ? "directory" : "file",
                  node,
                  scope,
                })
              }
            >
              {node.type === "directory" ? (
                <TreeButton
                  aria-label={`${expanded ? "折叠" : "展开"} ${node.name}`}
                  aria-expanded={expanded}
                  className="w-full overflow-hidden"
                  title={`${expanded ? "折叠" : "展开"} ${node.name}`}
                  onClick={() => onToggleDirectory(scope, node.path)}
                >
                  {hasChildren && expanded ? (
                    <ChevronDown className="shrink-0 text-muted-foreground" />
                  ) : (
                    <ChevronRight className="shrink-0 text-muted-foreground" />
                  )}
                  {expanded ? (
                    <FolderOpen className="shrink-0 text-primary workbench-folder-icon" />
                  ) : (
                    <Folder className="shrink-0 text-primary workbench-folder-icon" />
                  )}
                  <span className="truncate">{node.name}</span>
                </TreeButton>
              ) : (
                <TreeButton
                  aria-label={`Diff ${node.name}`}
                  className="w-full overflow-hidden"
                  title={`Diff ${node.name}`}
                  onClick={() => onDiff(node.path, scope)}
                >
                  <span className="w-4 shrink-0" />
                  <Codicon className="workbench-file-icon" name="file" />
                  <span className="truncate">{node.name}</span>
                  {node.status ? (
                    <Badge tone="warning">{statusLabel(node.status)}</Badge>
                  ) : null}
                </TreeButton>
              )}
              {node.type === "file" ? (
                <div className="flex items-center gap-0.5">
                  <IconButton
                    className="workbench-icon-button border-0 bg-transparent"
                    data-testid={`git-action-open-file-${scope}-${node.path}`}
                    title={`打开文件 ${node.name}`}
                    onClick={() => onOpenFile(node.path, scope)}
                  >
                    <Codicon name="open-preview" />
                  </IconButton>
                  {scope === "unstaged" ? (
                    <IconButton
                      className="workbench-icon-button border-0 bg-transparent"
                      data-testid={`git-action-discard-${scope}-${node.path}`}
                      title={`取消变更 ${node.name}`}
                      onClick={() => onDiscard([node.path])}
                    >
                      <Codicon name="discard" />
                    </IconButton>
                  ) : null}
                  <IconButton
                    className="workbench-icon-button border-0 bg-transparent"
                    data-testid={`git-action-${scope === "staged" ? "unstage" : "stage"}-${scope}-${node.path}`}
                    title={`${actionLabel} ${node.name}`}
                    onClick={() => onAction(node.path)}
                  >
                    {actionIcon}
                  </IconButton>
                </div>
              ) : null}
            </div>
            {hasChildren && expanded ? (
              <GitTree
                actionIcon={actionIcon}
                actionLabel={actionLabel}
                collapsedPaths={collapsedPaths}
                level={level + 1}
                nodes={node.children ?? []}
                scope={scope}
                selectedKey={selectedKey}
                onAction={onAction}
                onContextMenu={onContextMenu}
                onDiscard={onDiscard}
                onDiff={onDiff}
                onOpenFile={onOpenFile}
                onToggleDirectory={onToggleDirectory}
              />
            ) : null}
          </div>
        );
      })}
    </div>
  );
}

function GitDetailPane({
  compactLayout,
  dirty,
  offlineDescription,
  offlineTitle,
  opened,
  saving,
  settings,
  wordWrapOverride,
  onDraftChange,
  onOpenFile,
  onReloadFile,
  onSaveFile,
  onToggleDiffWordWrap,
}: {
  compactLayout: boolean;
  dirty: boolean;
  offlineDescription: string;
  offlineTitle: string;
  opened: OpenedGitResource;
  saving: boolean;
  settings: WorkbenchSettings;
  wordWrapOverride: WordWrapMode | null;
  onDraftChange: (draft: string) => void;
  onOpenFile: (path: string, sourceScope?: GitChangeScope) => void;
  onReloadFile: () => void;
  onSaveFile: () => void;
  onToggleDiffWordWrap: () => void;
}) {
  const effectiveOpened: OpenedGitResource = offlineTitle
    ? { kind: "empty" }
    : opened;
  const title =
    effectiveOpened.kind === "empty" ? "选择一个变更" : effectiveOpened.path;
  const subtitle =
    effectiveOpened.kind === "diff"
      ? effectiveOpened.scope === "staged"
        ? "Staged Changes"
        : "Changes"
      : effectiveOpened.kind === "file"
        ? "工作区文件"
        : "";
  const editable =
    effectiveOpened.kind === "file" &&
    effectiveOpened.file.status !== "deleted" &&
    effectiveOpened.file.file.previewType === "text" &&
    !effectiveOpened.file.file.tooLarge;
  const diffState =
    effectiveOpened.kind === "diff"
      ? buildDiffViewState(effectiveOpened, settings, compactLayout)
      : null;
  const effectiveDiffWordWrap = wordWrapOverride ?? settings.editorWordWrap;
  const diffEditorRef = useRef<MonacoDiffEditor | null>(null);
  const [diffEditorWordWrapSnapshot, setDiffEditorWordWrapSnapshot] =
    useState<DiffEditorWordWrapSnapshot | null>(null);
  const fileLanguage =
    effectiveOpened.kind === "file" && editable
      ? settings.codeHighlighting
        ? languageFromPath(effectiveOpened.file.file.file.path)
        : "plaintext"
      : undefined;
  const diffShortcutActive = Boolean(diffState && !diffState.fallbackMessage);
  const syncDiffEditorSettings = useCallback(
    (editor: MonacoDiffEditor) => {
      applyDiffEditorSettings(
        editor,
        settings.editorTabSize,
        effectiveDiffWordWrap,
      );
      setDiffEditorWordWrapSnapshot(readDiffEditorWordWrapSnapshot(editor));
    },
    [effectiveDiffWordWrap, settings.editorTabSize],
  );

  useEffect(() => {
    if (!diffShortcutActive) {
      return;
    }
    function handleDiffShortcut(event: KeyboardEvent) {
      if (
        event.altKey &&
        !event.ctrlKey &&
        !event.metaKey &&
        !event.shiftKey &&
        event.code === "KeyZ"
      ) {
        event.preventDefault();
        onToggleDiffWordWrap();
      }
    }
    window.addEventListener("keydown", handleDiffShortcut, true);
    return () => {
      window.removeEventListener("keydown", handleDiffShortcut, true);
    };
  }, [diffShortcutActive, onToggleDiffWordWrap]);

  useEffect(() => {
    if (!diffState || diffState.fallbackMessage) {
      diffEditorRef.current = null;
      setDiffEditorWordWrapSnapshot(null);
    }
  }, [
    diffState?.fallbackMessage,
    diffState?.modifiedModelPath,
    diffState?.originalModelPath,
  ]);

  useEffect(() => {
    if (!diffEditorRef.current || !diffState || diffState.fallbackMessage) {
      return;
    }
    syncDiffEditorSettings(diffEditorRef.current);
  }, [
    diffState?.fallbackMessage,
    diffState?.modifiedModelPath,
    diffState?.originalModelPath,
    syncDiffEditorSettings,
  ]);

  return (
    <div className="workbench-detail-pane grid h-full min-h-0 grid-rows-[auto_minmax(0,1fr)]">
      <header className="workbench-detail-header flex min-h-10 items-center justify-between gap-2 border-b px-2 py-1 max-[760px]:flex-col max-[760px]:items-stretch">
        <div className="flex min-w-0 items-center gap-2">
          {effectiveOpened.kind === "diff" ? <GitPullRequest /> : <FileText />}
          <div className="min-w-0">
            <div className="truncate text-sm font-medium">
              {dirty ? `${title} *` : title}
            </div>
            {subtitle ? (
              <div className="truncate text-xs text-muted-foreground">
                {subtitle}
              </div>
            ) : null}
          </div>
        </div>
        <div className="flex shrink-0 flex-wrap items-center justify-end gap-1.5 max-[760px]:justify-start">
          {effectiveOpened.kind === "diff" ? (
            <Button
              type="button"
              variant="soft"
              onClick={() =>
                onOpenFile(effectiveOpened.path, effectiveOpened.scope)
              }
            >
              <Codicon name="open-preview" />
              打开文件
            </Button>
          ) : null}
          {effectiveOpened.kind === "file" ? (
            <>
              {dirty ? <Badge tone="warning">未保存</Badge> : null}
              <IconButton
                className="workbench-icon-button"
                disabled={saving}
                title="重新加载"
                onClick={onReloadFile}
              >
                <RotateCcw />
              </IconButton>
              <Button
                disabled={!editable || !dirty || saving}
                type="button"
                onClick={onSaveFile}
              >
                <Save />
                保存
              </Button>
            </>
          ) : null}
        </div>
      </header>
      <div
        className="workbench-editor-surface min-h-0 overflow-hidden"
        data-diff-layout={diffState?.layout}
        data-diff-modified-word-wrap={diffEditorWordWrapSnapshot?.modified}
        data-diff-original-word-wrap={diffEditorWordWrapSnapshot?.original}
        data-editor-font-size={settings.editorFontSize}
        data-editor-language={diffState?.language ?? fileLanguage}
        data-editor-minimap={settings.editorMinimap}
        data-editor-tab-size={settings.editorTabSize}
        data-editor-word-wrap={diffState ? effectiveDiffWordWrap : settings.editorWordWrap}
      >
        {offlineTitle ? (
          <EmptyState
            icon={<PowerOff />}
            title={offlineTitle}
            description={offlineDescription}
          />
        ) : null}
        {!offlineTitle && effectiveOpened.kind === "empty" ? (
          <EmptyState icon={<GitBranch />} title="选择变更查看 diff" />
        ) : null}
        {effectiveOpened.kind === "diff" && diffState?.fallbackMessage ? (
          <EmptyState
            icon={<FileX />}
            title="无法在线展示该 diff"
            description={diffState.fallbackMessage}
          />
        ) : null}
        {effectiveOpened.kind === "diff" &&
        diffState &&
        !diffState.fallbackMessage ? (
          <Suspense fallback={<Spinner label="加载 diff" />}>
            <MonacoDiffEditor
              height="100%"
              keepCurrentModifiedModel
              keepCurrentOriginalModel
              language={diffState.language}
              modified={diffState.modified}
              modifiedModelPath={diffState.modifiedModelPath}
              options={{
                fontSize: settings.editorFontSize,
                minimap: { enabled: settings.editorMinimap },
                readOnly: true,
                renderSideBySide: diffState.layout === "side-by-side",
                diffWordWrap: effectiveDiffWordWrap,
                wordWrap: effectiveDiffWordWrap,
              }}
              original={diffState.original}
              originalModelPath={diffState.originalModelPath}
              theme="vs-dark"
              onMount={(editor) => {
                diffEditorRef.current = editor;
                syncDiffEditorSettings(editor);
              }}
            />
          </Suspense>
        ) : null}
        {effectiveOpened.kind === "file" && editable ? (
          <Suspense fallback={<Spinner label="加载编辑器" />}>
            <MonacoEditor
              height="100%"
              language={fileLanguage}
              options={{
                fontSize: settings.editorFontSize,
                minimap: { enabled: settings.editorMinimap },
                scrollBeyondLastLine: false,
                tabSize: settings.editorTabSize,
                wordWrap: settings.editorWordWrap,
              }}
              theme="vs-dark"
              value={effectiveOpened.draft}
              onChange={(nextValue) => onDraftChange(nextValue ?? "")}
            />
          </Suspense>
        ) : null}
        {effectiveOpened.kind === "file" && !editable ? (
          <EmptyState
            icon={<FileX />}
            title={effectiveOpened.file.message || "该文件不可编辑"}
            description="已删除、超大或非文本文件只能查看 diff 或下载处理。"
          />
        ) : null}
      </div>
    </div>
  );
}

function countFiles(nodes: GitTreeNode[]): number {
  return nodes.reduce(
    (total, node) =>
      total + (node.type === "file" ? 1 : countFiles(node.children ?? [])),
    0,
  );
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

function collectFilePaths(nodes: GitTreeNode[]): string[] {
  return nodes.flatMap((node) =>
    node.type === "file" ? [node.path] : collectFilePaths(node.children ?? []),
  );
}

function collectDirectoryKeys(
  nodes: GitTreeNode[],
  scope: GitChangeScope,
  output: Set<string>,
) {
  for (const node of nodes) {
    if (node.type !== "directory") {
      continue;
    }
    output.add(gitDirectoryKey(scope, node.path));
    collectDirectoryKeys(node.children ?? [], scope, output);
  }
}

function findGitNode(nodes: GitTreeNode[], path: string): GitTreeNode | null {
  for (const node of nodes) {
    if (node.path === path) {
      return node;
    }
    const child = node.children ? findGitNode(node.children, path) : null;
    if (child) {
      return child;
    }
  }
  return null;
}

function joinWorkspacePath(root: string, relativePath: string) {
  return `${root.replace(/\/+$/, "")}/${relativePath.replace(/^\/+/, "")}`;
}

function relativeGitPath(path: string, root: string) {
  const normalizedRoot = root.replace(/\/+$/, "");
  const normalizedPath = path.replace(/\/+$/, "");
  if (normalizedPath === normalizedRoot) {
    return "";
  }
  if (normalizedPath.startsWith(`${normalizedRoot}/`)) {
    return normalizedPath.slice(normalizedRoot.length + 1);
  }
  if (normalizedPath.startsWith(`${workspaceRootPath}/`)) {
    return normalizedPath.slice(workspaceRootPath.length + 1);
  }
  if (normalizedPath.startsWith(`${sharedRootPath}/`)) {
    return normalizedPath.slice(sharedRootPath.length + 1);
  }
  return normalizedPath.replace(/^\/+/, "");
}

function gitDirectoryKey(scope: GitChangeScope, path: string) {
  return `${scope}:${path}`;
}

function statusLabel(value: string) {
  const labels: Record<string, string> = {
    modified: "修改",
    untracked: "未跟踪",
    deleted: "删除",
    added: "新增",
    renamed: "重命名",
  };
  return labels[value] ?? value;
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

function buildDiffViewState(
  opened: Extract<OpenedGitResource, { kind: "diff" }>,
  settings: WorkbenchSettings,
  compactLayout: boolean,
) {
  const hasTextModels = Boolean(
    opened.diff.originalContent || opened.diff.modifiedContent,
  );
  const unsafeMessage = Boolean(
    opened.diff.message &&
    /not editable text|exceeds editable size limit|too large|binary/i.test(
      opened.diff.message,
    ),
  );
  const fallbackOnly = Boolean(
    !hasTextModels && opened.diff.diff && !unsafeMessage,
  );
  const fallbackMessage =
    opened.diff.message &&
    !hasTextModels &&
    (!opened.diff.diff || unsafeMessage)
      ? opened.diff.message
      : "";
  const sideBySide =
    settings.gitDiffDisplay === "side-by-side" &&
    !(compactLayout && settings.gitDiffInlineOnNarrow) &&
    !fallbackOnly;
  const language = fallbackOnly
    ? "diff"
    : settings.codeHighlighting
      ? opened.diff.language ||
        languageFromPath(opened.diff.modifiedPath || opened.path)
      : "plaintext";
  const originalPath = opened.diff.originalPath || opened.path;
  const modifiedPath = opened.diff.modifiedPath || opened.path;

  return {
    fallbackMessage,
    language,
    layout: sideBySide ? "side-by-side" : "inline",
    modified: fallbackOnly ? opened.diff.diff : opened.diff.modifiedContent,
    modifiedModelPath: modelPath(
      "modified",
      opened.scope,
      modifiedPath,
      fallbackOnly,
    ),
    original: fallbackOnly ? "" : opened.diff.originalContent,
    originalModelPath: modelPath(
      "original",
      opened.scope,
      originalPath,
      fallbackOnly,
    ),
  };
}

function modelPath(
  side: "original" | "modified",
  scope: GitChangeScope,
  filePath: string,
  fallbackOnly: boolean,
) {
  const suffix = fallbackOnly ? ".diff" : "";
  return `john-ai-agentbox://git-diff/${scope}/${side}/${encodeURIComponent(filePath)}${suffix}`;
}

function applyDiffEditorSettings(
  editor: MonacoDiffEditor,
  tabSize: number,
  wordWrap: WordWrapMode,
) {
  const model = editor.getModel();
  model?.original.updateOptions({ tabSize });
  model?.modified.updateOptions({ tabSize });
  editor.updateOptions({ diffWordWrap: wordWrap, wordWrap });
  editor
    .getOriginalEditor()
    .updateOptions({
      wordWrap,
      wordWrapOverride1: wordWrap,
      wordWrapOverride2: wordWrap,
    });
  editor
    .getModifiedEditor()
    .updateOptions({
      wordWrap,
      wordWrapOverride1: wordWrap,
      wordWrapOverride2: wordWrap,
    });
}

function readDiffEditorWordWrapSnapshot(
  editor: MonacoDiffEditor,
): DiffEditorWordWrapSnapshot {
  return {
    original: readEditorWordWrap(editor.getOriginalEditor()),
    modified: readEditorWordWrap(editor.getModifiedEditor()),
  };
}

function readEditorWordWrap(
  editor: MonacoEditorApi.IStandaloneCodeEditor,
): WordWrapMode | "inherit" {
  const rawOptions = editor.getRawOptions();
  const override = rawOptions.wordWrapOverride1;
  if (override === "on" || override === "off" || override === "inherit") {
    return override;
  }
  return rawOptions.wordWrap === "on" ? "on" : "off";
}

function Codicon({
  className,
  name,
}: {
  className?: string;
  name: string;
}) {
  return (
    <span
      aria-hidden="true"
      className={cn("codicon", `codicon-${name}`, className)}
    />
  );
}
