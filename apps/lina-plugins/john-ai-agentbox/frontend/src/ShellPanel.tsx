import { Fragment, useEffect, useMemo, useRef, useState } from "react";
import type { MouseEvent as ReactMouseEvent, ReactNode } from "react";
import {
  ClipboardPaste,
  Columns2,
  Copy,
  Eraser,
  Fullscreen,
  Maximize2,
  Minimize2,
  PanelRight,
  PanelTop,
  Plus,
  RefreshCcw,
  Rows3,
  ScanText,
  Shrink,
  Terminal as TerminalIcon,
  X,
} from "lucide-react";
import { toast } from "sonner";
import { FitAddon } from "@xterm/addon-fit";
import { Terminal } from "@xterm/xterm";
import type { AgentInfo, ShellMessage } from "./types";
import { pluginWebSocketURL } from "./plugin-paths";
import {
  Button,
  IconButton,
  ResizablePanel,
  ResizablePanelGroup,
  ResizeHandle,
  TabButton,
} from "@/components/ui";
import type { WorkbenchSettings } from "@/lib/workbench-settings";
import { cn, normalizeWorkspacePath, workspaceRootPath } from "@/lib/utils";

const shellTabsStorageKey = "john-ai-agentbox-shell-tabs";
const maxPanesPerAgent = 8;
const darkPlusTerminalSelectionBackground = "#264f78";
const darkPlusTerminalSelectionForeground = "#ffffff";
const terminalBrowserControlledModeParams = new Set([
  "9",
  "47",
  "1000",
  "1002",
  "1003",
  "1005",
  "1006",
  "1007",
  "1015",
  "1047",
  "1048",
  "1049",
]);
const terminalGeneratedResponseSequences = ["\x1b[?1;2c", "\x1b[>0;276;0c"];
const terminalImeFallbackDelayMs = 25;
const terminalTmuxControlledScrollback = 0;
const darkPlusTerminalTheme = {
  background: "#1e1e1e",
  foreground: "#cccccc",
  cursor: "#cccccc",
  selectionBackground: darkPlusTerminalSelectionBackground,
  selectionForeground: darkPlusTerminalSelectionForeground,
  selectionInactiveBackground: darkPlusTerminalSelectionBackground,
  black: "#000000",
  red: "#cd3131",
  green: "#0dbc79",
  yellow: "#e5e510",
  blue: "#2472c8",
  magenta: "#bc3fbc",
  cyan: "#11a8cd",
  white: "#e5e5e5",
  brightBlack: "#666666",
  brightRed: "#f14c4c",
  brightGreen: "#23d18b",
  brightYellow: "#f5f543",
  brightBlue: "#3b8eea",
  brightMagenta: "#d670d6",
  brightCyan: "#29b8db",
  brightWhite: "#e5e5e5",
};

type TerminalPaneNode = {
  type: "pane";
  id: string;
  agentId: string;
  sequence: number;
  workingDir: string;
};

type TerminalSplitNode = {
  type: "split";
  id: string;
  direction: "horizontal" | "vertical";
  children: TerminalLayoutNode[];
};

type TerminalTabsNode = {
  type: "tabs";
  id: string;
  activePaneId: string;
  children: TerminalPaneNode[];
};

type TerminalLayoutNode =
  | TerminalPaneNode
  | TerminalSplitNode
  | TerminalTabsNode;

type ShellSession = {
  id: string;
  terminal: Terminal;
  fit: FitAddon;
  socket: WebSocket;
  resizeObserver: ResizeObserver;
  inputDisposable: { dispose: () => void };
  imeInputFallback: TerminalImeInputFallback;
  pendingTerminalOutput: string;
  disposed: boolean;
};

type TerminalImeInputFallback = {
  dispose: () => void;
  handleTerminalData: (data: string) => boolean;
};

type ShellContextMenu = {
  open: boolean;
  x: number;
  y: number;
  paneId: string;
};

type Props = {
  active: boolean;
  agent?: AgentInfo;
  agents: AgentInfo[];
  settings: WorkbenchSettings;
  workspacePath: string;
};

type TerminalPanelSurfaceProps = Props & {
  chromeDensity?: "default" | "compact";
  className?: string;
  headerTestId?: string;
  panelTestId?: string;
  storageKey?: string;
  surfaceTestId?: string;
  terminalIdPrefix?: string;
};

export default function ShellPanel(props: Props) {
  return <TerminalPanelSurface {...props} />;
}

export function TerminalPanelSurface({
  active,
  agent,
  agents,
  chromeDensity = "default",
  className,
  headerTestId = "shell-terminal-header",
  panelTestId = "shell-panel",
  settings,
  storageKey = shellTabsStorageKey,
  surfaceTestId = "shell-terminal-surface",
  terminalIdPrefix = "",
  workspacePath,
}: TerminalPanelSurfaceProps) {
  const [persistedShellState] = useState(() =>
    loadPersistedShellState(storageKey),
  );
  const [layoutsByAgent, setLayoutsByAgent] = useState<
    Record<string, TerminalLayoutNode>
  >(persistedShellState.layoutsByAgent);
  const [activeByAgent, setActiveByAgent] = useState<Record<string, string>>(
    persistedShellState.activeByAgent,
  );
  const [statuses, setStatuses] = useState<Record<string, string>>({});
  const [contextMenu, setContextMenu] = useState<ShellContextMenu>({
    open: false,
    x: 0,
    y: 0,
    paneId: "",
  });
  const [terminalFullscreen, setTerminalFullscreen] = useState(false);
  const [maximizedPaneId, setMaximizedPaneId] = useState<string | null>(null);
  const panelRef = useRef<HTMLDivElement | null>(null);
  const sessionsRef = useRef(new Map<string, ShellSession>());
  const hostsRef = useRef(new Map<string, HTMLDivElement>());
  const paneAgentsRef = useRef(new Map<string, string>());
  const uidRef = useRef(0);

  const currentLayout = agent ? layoutsByAgent[agent.id] : undefined;
  const effectiveWorkspacePath = normalizeWorkspacePath(
    workspacePath || workspaceRootPath,
  );
  const visiblePanes = useMemo(
    () => (currentLayout ? collectPanes(currentLayout) : []),
    [currentLayout],
  );
  const renderedPanes = useMemo(
    () => (currentLayout ? collectRenderedPanes(currentLayout) : []),
    [currentLayout],
  );
  const activePaneId = getActivePaneId(agent, renderedPanes, activeByAgent);
  const maximizedPane = maximizedPaneId
    ? visiblePanes.find((pane) => pane.id === maximizedPaneId)
    : undefined;
  const canOpenShell = active && agent?.runtimeStatus === "running";
  const compactChrome = chromeDensity === "compact";

  useEffect(() => {
    savePersistedShellState(storageKey, { layoutsByAgent, activeByAgent });
  }, [activeByAgent, layoutsByAgent, storageKey]);

  useEffect(() => {
    if (!agent?.id || !activePaneId) {
      return;
    }
    setActiveByAgent((current) =>
      current[agent.id] === activePaneId
        ? current
        : { ...current, [agent.id]: activePaneId },
    );
  }, [agent?.id, activePaneId]);

  useEffect(() => {
    if (!agent) {
      if (
        terminalFullscreen &&
        document.fullscreenElement === panelRef.current
      ) {
        void exitDocumentFullscreen().catch(() => undefined);
      }
      setTerminalFullscreen(false);
      setMaximizedPaneId(null);
      closeContextMenu();
      return;
    }
    if (!active) {
      if (
        terminalFullscreen &&
        document.fullscreenElement === panelRef.current
      ) {
        void exitDocumentFullscreen().catch(() => undefined);
      }
      setTerminalFullscreen(false);
      closeContextMenu();
      return;
    }
    if (agent.runtimeStatus !== "running") {
      collectPanes(layoutsByAgent[agent.id]).forEach((pane) =>
        closeSession(pane.id),
      );
      if (
        terminalFullscreen &&
        document.fullscreenElement === panelRef.current
      ) {
        void exitDocumentFullscreen().catch(() => undefined);
      }
      setTerminalFullscreen(false);
      setMaximizedPaneId(null);
      closeContextMenu();
      return;
    }
    ensureAgentLayout(agent);
  }, [active, agent?.id, agent?.runtimeStatus, terminalFullscreen]);

  useEffect(() => {
    if (
      !maximizedPaneId ||
      visiblePanes.some((pane) => pane.id === maximizedPaneId)
    ) {
      return;
    }
    setMaximizedPaneId(null);
  }, [maximizedPaneId, visiblePanes.map((pane) => pane.id).join("|")]);

  useEffect(() => {
    function syncTerminalFullscreen() {
      const nextTerminalFullscreen =
        document.fullscreenElement === panelRef.current;
      setTerminalFullscreen(nextTerminalFullscreen);
      window.requestAnimationFrame(() => {
        sessionsRef.current.forEach((session) => {
          fitSession(session);
        });
        if (nextTerminalFullscreen && activePaneId) {
          sessionsRef.current.get(activePaneId)?.terminal.focus();
        }
      });
    }
    document.addEventListener("fullscreenchange", syncTerminalFullscreen);
    return () =>
      document.removeEventListener("fullscreenchange", syncTerminalFullscreen);
  }, [activePaneId]);

  useEffect(() => {
    window.requestAnimationFrame(() => {
      sessionsRef.current.forEach((session) => {
        fitSession(session);
      });
      if (maximizedPaneId) {
        sessionsRef.current.get(maximizedPaneId)?.terminal.focus();
      }
    });
  }, [maximizedPaneId]);

  useEffect(() => {
    if (agents.length === 0) {
      return;
    }
    const runningIds = new Set(
      agents
        .filter((item) => item.runtimeStatus === "running")
        .map((item) => item.id),
    );
    const stalePaneIds = Object.entries(layoutsByAgent)
      .filter(([agentId]) => !runningIds.has(agentId))
      .flatMap(([, layout]) => collectPanes(layout).map((pane) => pane.id));
    if (stalePaneIds.length === 0) {
      return;
    }
    stalePaneIds.forEach((paneId) => closeSession(paneId));
    setLayoutsByAgent((current) => {
      const next = { ...current };
      Object.keys(next).forEach((agentId) => {
        if (!runningIds.has(agentId)) {
          delete next[agentId];
        }
      });
      return next;
    });
    setActiveByAgent((current) => {
      const next = { ...current };
      Object.keys(next).forEach((agentId) => {
        if (!runningIds.has(agentId)) {
          delete next[agentId];
        }
      });
      return next;
    });
  }, [
    agents.map((item) => `${item.id}:${item.runtimeStatus}`).join("|"),
    layoutsByAgent,
  ]);

  useEffect(() => {
    if (!active || !agent || agent.runtimeStatus !== "running") {
      return;
    }
    visiblePanes.forEach((pane) => {
      const host = hostsRef.current.get(pane.id);
      if (host) {
        createSession(agent, pane, host);
      }
    });
    window.requestAnimationFrame(() => {
      visiblePanes.forEach((pane) => {
        const session = sessionsRef.current.get(pane.id);
        if (session) {
          fitSession(session);
        }
      });
      sessionsRef.current.get(activePaneId)?.terminal.focus();
    });
  }, [
    active,
    agent?.id,
    agent?.runtimeStatus,
    activePaneId,
    visiblePanes.map((pane) => pane.id).join("|"),
  ]);

  useEffect(() => {
    sessionsRef.current.forEach((session) => {
      const host = hostsRef.current.get(session.id);
      session.terminal.options.fontSize = settings.terminalFontSize;
      session.terminal.options.lineHeight = settings.terminalLineHeight;
      session.terminal.options.cursorBlink = settings.terminalCursorBlink;
      session.terminal.options.cursorStyle = settings.terminalCursorStyle;
      session.terminal.options.cursorWidth = settings.terminalCursorWidth;
      session.terminal.options.scrollback = terminalTmuxControlledScrollback;
      if (host) {
        applyTerminalHostStyle(host, settings.terminalFontSize);
      }
      session.terminal.element?.style.setProperty(
        "font-size",
        `${settings.terminalFontSize}px`,
      );
      window.requestAnimationFrame(() => {
        if (!session.disposed) {
          session.terminal.refresh(0, session.terminal.rows - 1);
          fitSession(session);
        }
      });
    });
  }, [
    settings.terminalCursorBlink,
    settings.terminalCursorStyle,
    settings.terminalCursorWidth,
    settings.terminalFontSize,
    settings.terminalLineHeight,
  ]);

  useEffect(() => {
    if (!contextMenu.open) {
      return;
    }
    function close() {
      closeContextMenu();
    }
    window.addEventListener("click", close);
    window.addEventListener("keydown", close);
    window.addEventListener("resize", close);
    return () => {
      window.removeEventListener("click", close);
      window.removeEventListener("keydown", close);
      window.removeEventListener("resize", close);
    };
  }, [contextMenu.open]);

  useEffect(() => {
    return () => {
      sessionsRef.current.forEach((session) => disposeSession(session));
      sessionsRef.current.clear();
    };
  }, []);

  function createPaneSeed(
    agentId: string,
    existingLayout?: TerminalLayoutNode,
    workingDir = effectiveWorkspacePath,
  ) {
    uidRef.current += 1;
    const id = terminalIdPrefix
      ? `${terminalIdPrefix}-${agentId}-${Date.now()}-${uidRef.current}`
      : `${agentId}-${Date.now()}-${uidRef.current}`;
    return {
      id,
      agentId,
      sequence: getNextPaneSequence(existingLayout, agentId),
      workingDir,
    };
  }

  function ensureAgentLayout(target: AgentInfo) {
    setLayoutsByAgent((current) => {
      const existing = current[target.id];
      const existingPanes = collectPanes(existing);
      if (existing && existingPanes.length > 0) {
        setActiveByAgent((activeCurrent) => {
          if (
            existingPanes.some((pane) => pane.id === activeCurrent[target.id])
          ) {
            return activeCurrent;
          }
          return { ...activeCurrent, [target.id]: existingPanes[0].id };
        });
        return current;
      }
      const pane: TerminalPaneNode = {
        ...createPaneSeed(target.id),
        type: "pane",
      };
      setActiveByAgent((activeCurrent) => ({
        ...activeCurrent,
        [target.id]: pane.id,
      }));
      return { ...current, [target.id]: pane };
    });
  }

  function openPane(target: AgentInfo, basePaneId = activePaneId) {
    if (target.runtimeStatus !== "running") {
      return;
    }
    setLayoutsByAgent((current) => {
      const currentLayout = current[target.id];
      const panes = collectPanes(currentLayout);
      if (panes.length >= maxPanesPerAgent) {
        toast.warning(`每个智能体最多打开 ${maxPanesPerAgent} 个终端。`);
        return current;
      }
      const pane: TerminalPaneNode = {
        ...createPaneSeed(target.id, currentLayout),
        type: "pane",
      };
      setActiveByAgent((activeCurrent) => ({
        ...activeCurrent,
        [target.id]: pane.id,
      }));
      if (
        maximizedPaneId &&
        panes.some((item) => item.id === maximizedPaneId)
      ) {
        setMaximizedPaneId(pane.id);
      }
      if (!currentLayout) {
        return { ...current, [target.id]: pane };
      }
      return {
        ...current,
        [target.id]: appendPane(currentLayout, basePaneId, pane),
      };
    });
  }

  function splitPane(paneId: string, direction: "horizontal" | "vertical") {
    if (!agent || agent.runtimeStatus !== "running") {
      return;
    }
    setLayoutsByAgent((current) => {
      const layout = current[agent.id];
      const panes = collectPanes(layout);
      if (!layout || !panes.some((pane) => pane.id === paneId)) {
        return current;
      }
      if (panes.length >= maxPanesPerAgent) {
        toast.warning(`每个智能体最多打开 ${maxPanesPerAgent} 个终端。`);
        return current;
      }
      const pane: TerminalPaneNode = {
        ...createPaneSeed(agent.id, layout),
        type: "pane",
      };
      const nextLayout = splitLayoutPane(layout, paneId, direction, pane);
      setActiveByAgent((activeCurrent) => ({
        ...activeCurrent,
        [agent.id]: pane.id,
      }));
      return { ...current, [agent.id]: nextLayout };
    });
  }

  function closePane(pane: TerminalPaneNode) {
    if (!agent) {
      return;
    }
    setMaximizedPaneId((current) => (current === pane.id ? null : current));
    closeSession(pane.id, true);
    closeContextMenu();
    setLayoutsByAgent((current) => {
      const layout = current[pane.agentId];
      const panes = collectPanes(layout);
      if (!layout || panes.length <= 1) {
        const replacement: TerminalPaneNode = {
          ...createPaneSeed(pane.agentId),
          type: "pane",
        };
        setActiveByAgent((activeCurrent) => ({
          ...activeCurrent,
          [pane.agentId]: replacement.id,
        }));
        return { ...current, [pane.agentId]: replacement };
      }
      const nextLayout = removePaneFromLayout(layout, pane.id);
      if (!nextLayout) {
        return current;
      }
      const nextPanes = collectPanes(nextLayout);
      setActiveByAgent((activeCurrent) => {
        if (nextPanes.some((item) => item.id === activeCurrent[pane.agentId])) {
          return activeCurrent;
        }
        return { ...activeCurrent, [pane.agentId]: nextPanes[0]?.id ?? "" };
      });
      return { ...current, [pane.agentId]: nextLayout };
    });
  }

  function setActivePane(pane: TerminalPaneNode) {
    setActiveByAgent((current) => ({ ...current, [pane.agentId]: pane.id }));
    if (maximizedPaneId && visiblePanes.some((item) => item.id === pane.id)) {
      setMaximizedPaneId(pane.id);
    }
    setLayoutsByAgent((current) => {
      const layout = current[pane.agentId];
      if (!layout || !layoutContainsPane(layout, pane.id)) {
        return current;
      }
      const nextLayout = activatePaneInLayout(layout, pane.id);
      return nextLayout === layout
        ? current
        : { ...current, [pane.agentId]: nextLayout };
    });
    sessionsRef.current.get(pane.id)?.terminal.focus();
  }

  function setHostNode(pane: TerminalPaneNode, node: HTMLDivElement | null) {
    paneAgentsRef.current.set(pane.id, pane.agentId);
    if (!node) {
      hostsRef.current.delete(pane.id);
      sessionsRef.current.get(pane.id)?.resizeObserver.disconnect();
      return;
    }
    hostsRef.current.set(pane.id, node);
    if (
      active &&
      agent?.id === pane.agentId &&
      agent.runtimeStatus === "running"
    ) {
      createSession(agent, pane, node);
    }
  }

  function createSession(
    target: AgentInfo,
    pane: TerminalPaneNode,
    host: HTMLDivElement,
    mode?: "rebuild",
  ) {
    const current = sessionsRef.current.get(pane.id);
    if (current) {
      attachSessionToHost(current, host);
      return;
    }

    const terminal = new Terminal({
      cursorBlink: settings.terminalCursorBlink,
      cursorStyle: settings.terminalCursorStyle,
      cursorWidth: settings.terminalCursorWidth,
      fontFamily: '"JetBrains Mono", "SFMono-Regular", Consolas, monospace',
      fontSize: settings.terminalFontSize,
      lineHeight: settings.terminalLineHeight,
      macOptionClickForcesSelection: true,
      scrollback: terminalTmuxControlledScrollback,
      theme: darkPlusTerminalTheme,
    });
    const fit = new FitAddon();
    terminal.loadAddon(fit);
    terminal.attachCustomKeyEventHandler((event) => {
      if (!isTerminalCopyShortcut(event)) {
        return true;
      }
      const selection = terminal.getSelection();
      if (!selection) {
        return true;
      }
      event.preventDefault();
      event.stopPropagation();
      void copyTerminalText(selection, { showSuccess: false });
      return false;
    });
    terminal.attachCustomWheelEventHandler((event) =>
      handleTerminalWheel(terminal, event, {
        onPtyWheel: (data) => sendShellInput(pane.id, data),
      }),
    );
    applyTerminalHostStyle(host, settings.terminalFontSize);
    terminal.open(host);

    const params = new URLSearchParams({
      terminalId: pane.id,
      cwd: pane.workingDir || effectiveWorkspacePath,
    });
    if (mode) {
      params.set("mode", mode);
    }
    const socket = new WebSocket(
      pluginWebSocketURL(
        `/agents/${encodeURIComponent(target.id)}/shell?${params}`,
      ),
    );

    let session: ShellSession;
    const resizeObserver = new ResizeObserver(() => fitSession(session));
    const imeInputFallback = installTerminalImeInputFallback(terminal, (data) =>
      sendShellInput(pane.id, data),
    );
    const inputDisposable = terminal.onData((data) => {
      if (!imeInputFallback.handleTerminalData(data)) {
        return;
      }
      sendShellInput(pane.id, data);
    });

    session = {
      id: pane.id,
      terminal,
      fit,
      socket,
      resizeObserver,
      inputDisposable,
      imeInputFallback,
      pendingTerminalOutput: "",
      disposed: false,
    };
    sessionsRef.current.set(pane.id, session);
    paneAgentsRef.current.set(pane.id, pane.agentId);
    setSessionStatus(pane.id, mode === "rebuild" ? "rebuilding" : "connecting");
    resizeObserver.observe(host);

    socket.addEventListener("open", () => {
      if (session.disposed) return;
      fitSession(session);
    });
    let preserveTerminalStatusOnClose = false;
    socket.addEventListener("message", (event) => {
      if (session.disposed) return;
      const message = JSON.parse(event.data) as ShellMessage;
      if (
        (message.type === "output" || message.type === "replay") &&
        message.data
      ) {
        const output = sanitizeTerminalOutput(session, message.data);
        if (output) {
          terminal.write(output);
        }
      }
      if (message.type === "status" && message.status) {
        if (message.status === "detached" && mode !== "rebuild") {
          preserveTerminalStatusOnClose = true;
          rebuildSession(pane);
          return;
        }
        setSessionStatus(pane.id, message.status);
        if (message.status === "connected") {
          window.requestAnimationFrame(() => fitSession(session));
        }
        if (message.status === "error") {
          preserveTerminalStatusOnClose = true;
        }
      }
      if (message.type === "error") {
        preserveTerminalStatusOnClose = true;
        setSessionStatus(pane.id, "error");
        terminal.writeln(
          `\r\n${message.message || "终端错误"}\r\n可重建终端会话。`,
        );
      }
    });
    socket.addEventListener("close", () => {
      if (!session.disposed && !preserveTerminalStatusOnClose) {
        setSessionStatus(pane.id, "closed");
      }
    });
  }

  function attachSessionToHost(session: ShellSession, host: HTMLDivElement) {
    if (session.disposed) {
      return;
    }
    applyTerminalHostStyle(host, settings.terminalFontSize);
    if (
      session.terminal.element &&
      session.terminal.element.parentElement !== host
    ) {
      host.appendChild(session.terminal.element);
    }
    session.resizeObserver.disconnect();
    session.resizeObserver.observe(host);
    window.requestAnimationFrame(() => {
      fitSession(session);
      if (session.id === activePaneId) {
        session.terminal.focus();
      }
    });
  }

  function setSessionStatus(id: string, status: string) {
    setStatuses((current) =>
      current[id] === status ? current : { ...current, [id]: status },
    );
  }

  function fitSession(session: ShellSession) {
    const host = hostsRef.current.get(session.id);
    if (
      session.disposed ||
      !host ||
      host.clientWidth === 0 ||
      host.clientHeight === 0
    ) {
      return;
    }
    session.fit.fit();
    if (session.socket.readyState === WebSocket.OPEN) {
      session.socket.send(
        JSON.stringify({
          type: "resize",
          cols: session.terminal.cols,
          rows: session.terminal.rows,
        }),
      );
    }
  }

  function sendShellInput(paneId: string, data: string) {
    const session = sessionsRef.current.get(paneId);
    const input = stripTerminalGeneratedResponses(data);
    if (
      !session ||
      session.disposed ||
      session.socket.readyState !== WebSocket.OPEN ||
      !input
    ) {
      return;
    }
    session.socket.send(JSON.stringify({ type: "input", data: input }));
  }

  function closeSession(id: string, notifyServer = false) {
    const session = sessionsRef.current.get(id);
    if (!session && notifyServer) {
      notifyShellClose(id);
      setStatuses((current) => {
        const next = { ...current };
        delete next[id];
        return next;
      });
      return;
    }
    if (!session) {
      return;
    }
    if (notifyServer) {
      if (session.socket.readyState === WebSocket.OPEN) {
        session.socket.send(JSON.stringify({ type: "close" }));
      } else {
        notifyShellClose(id);
      }
    }
    disposeSession(session);
    sessionsRef.current.delete(id);
    setStatuses((current) => {
      const next = { ...current };
      delete next[id];
      return next;
    });
  }

  function notifyShellClose(id: string) {
    const agentId = paneAgentsRef.current.get(id);
    if (!agentId) {
      return;
    }
    const params = new URLSearchParams({ terminalId: id, mode: "close" });
    const socket = new WebSocket(
      pluginWebSocketURL(
        `/agents/${encodeURIComponent(agentId)}/shell?${params}`,
      ),
    );
    socket.addEventListener("open", () => socket.close());
  }

  function rebuildSession(pane: TerminalPaneNode) {
    const session = sessionsRef.current.get(pane.id);
    if (session) {
      disposeSession(session);
      sessionsRef.current.delete(pane.id);
    }
    setSessionStatus(pane.id, "rebuilding");
    const host = hostsRef.current.get(pane.id);
    const target = agents.find((item) => item.id === pane.agentId);
    if (host && target?.runtimeStatus === "running") {
      createSession(target, pane, host, "rebuild");
      return;
    }
    setSessionStatus(pane.id, "error");
  }

  function disposeSession(session: ShellSession) {
    session.disposed = true;
    session.resizeObserver.disconnect();
    session.inputDisposable.dispose();
    session.imeInputFallback.dispose();
    session.socket.close();
    session.terminal.dispose();
  }

  async function copySelection(paneId: string) {
    const selection =
      sessionsRef.current.get(paneId)?.terminal.getSelection() ?? "";
    if (!selection) {
      toast.info("终端中没有选中文本。");
      return;
    }
    try {
      await copyTerminalText(selection, { showSuccess: false });
      toast.success("已复制终端选中文本");
    } catch (error) {
      toast.error(`复制失败：${(error as Error).message}`);
    }
  }

  async function pasteClipboard(paneId: string) {
    try {
      const text = await navigator.clipboard.readText();
      if (text) {
        sendShellInput(paneId, text);
      }
    } catch (error) {
      toast.error(`粘贴失败：${(error as Error).message}`);
    }
  }

  function selectAll(paneId: string) {
    sessionsRef.current.get(paneId)?.terminal.selectAll();
  }

  function clearTerminal(paneId: string) {
    sessionsRef.current.get(paneId)?.terminal.clear();
  }

  function handlePaneContextMenu(
    event: ReactMouseEvent,
    pane: TerminalPaneNode,
  ) {
    event.preventDefault();
    event.stopPropagation();
    setActivePane(pane);
    if (settings.terminalRightClickBehavior === "paste") {
      void pasteClipboard(pane.id);
      return;
    }
    setContextMenu({
      open: true,
      x: event.clientX,
      y: event.clientY,
      paneId: pane.id,
    });
  }

  function closeContextMenu() {
    setContextMenu((current) =>
      current.open ? { ...current, open: false } : current,
    );
  }

  function runContextAction(action: () => void | Promise<void>) {
    closeContextMenu();
    void action();
  }

  function togglePaneMaximized(pane: TerminalPaneNode) {
    closeContextMenu();
    setActivePane(pane);
    setMaximizedPaneId((current) => (current === pane.id ? null : pane.id));
  }

  async function toggleTerminalFullscreen() {
    if (terminalFullscreen) {
      if (document.fullscreenElement) {
        await exitDocumentFullscreen();
      }
      return;
    }
    if (activePaneId) {
      closeContextMenu();
      const panelNode = panelRef.current;
      if (!panelNode) {
        return;
      }
      try {
        if (document.fullscreenElement) {
          await exitDocumentFullscreen();
        }
        await panelNode.requestFullscreen();
      } catch (error) {
        toast.error(`切换全屏失败：${(error as Error).message}`);
      }
    }
  }

  return (
    <div
      ref={panelRef}
      className={cn(
        "grid h-full min-h-0",
        "grid-rows-[auto_minmax(0,1fr)]",
        terminalFullscreen && "h-screen w-screen bg-background",
        active ? "" : "hidden",
        className,
      )}
      data-maximized-pane-id={maximizedPane?.id ?? ""}
      data-pane-maximized={maximizedPane ? "true" : "false"}
      data-terminal-fullscreen={terminalFullscreen ? "window" : "false"}
      data-testid={panelTestId}
    >
      <div
        className={cn(
          "flex items-center justify-between border-b border-border bg-muted/50",
          compactChrome ? "min-h-8 gap-1 px-1.5" : "min-h-10 gap-2 px-2",
        )}
        data-terminal-density={chromeDensity}
        data-testid={headerTestId}
      >
        <div
          className={cn(
            "flex min-w-0 flex-1 items-center overflow-auto",
            compactChrome ? "gap-0.5" : "gap-1",
          )}
          role="tablist"
          aria-label="终端 pane"
        >
          {visiblePanes.map((pane) => {
            const status = statuses[pane.id];
            const statusTone = terminalPaneStatusTone(status);
            return (
              <div
                key={pane.id}
                className={cn(
                  "flex min-w-0 items-center rounded-[6px] border border-border bg-card",
                  pane.id === activePaneId && "border-primary/30 bg-primary/10",
                )}
              >
                <TabButton
                  active={pane.id === activePaneId}
                  className={cn(
                    "min-w-0 rounded-r-none border-0",
                    compactChrome
                      ? "h-6 px-1.5 text-[11px]"
                      : "h-8 px-2 text-xs",
                  )}
                  role="tab"
                  onClick={() => setActivePane(pane)}
                >
                  <TerminalIcon
                    className={compactChrome ? "h-3 w-3" : "h-3.5 w-3.5"}
                  />
                  <span className="truncate">{getPaneLabel(pane)}</span>
                  <span
                    className={cn("dot", statusTone)}
                    data-terminal-status={status ?? "pending"}
                    data-terminal-status-tone={statusTone}
                    data-testid="shell-pane-status-dot"
                  />
                </TabButton>
                <IconButton
                  className={cn(
                    "border-0 bg-transparent",
                    compactChrome ? "h-6 w-6" : "h-7 w-7",
                  )}
                  disabled={visiblePanes.length <= 1}
                  title={`关闭${getPaneLabel(pane)}`}
                  onClick={() => closePane(pane)}
                >
                  <X className={compactChrome ? "h-3 w-3" : "h-3.5 w-3.5"} />
                </IconButton>
              </div>
            );
          })}
        </div>
        <div
          className={cn(
            "flex shrink-0 items-center",
            compactChrome ? "gap-0.5" : "gap-1",
          )}
        >
          <Button
            aria-label="新建终端"
            disabled={!canOpenShell || !agent}
            size={compactChrome ? "icon-xs" : "icon"}
            type="button"
            variant="soft"
            onClick={() => agent && openPane(agent)}
          >
            <Plus className={compactChrome ? "h-3.5 w-3.5" : "h-4 w-4"} />
          </Button>
          <Button
            aria-label="水平拆分终端"
            disabled={!canOpenShell || !activePaneId}
            size={compactChrome ? "icon-xs" : "icon"}
            type="button"
            variant="soft"
            onClick={() =>
              activePaneId && splitPane(activePaneId, "horizontal")
            }
          >
            <Columns2 className={compactChrome ? "h-3.5 w-3.5" : "h-4 w-4"} />
          </Button>
          <Button
            aria-label="垂直拆分终端"
            disabled={!canOpenShell || !activePaneId}
            size={compactChrome ? "icon-xs" : "icon"}
            type="button"
            variant="soft"
            onClick={() => activePaneId && splitPane(activePaneId, "vertical")}
          >
            <Rows3 className={compactChrome ? "h-3.5 w-3.5" : "h-4 w-4"} />
          </Button>
          <Button
            aria-label={terminalFullscreen ? "退出终端全屏" : "终端面板全屏"}
            disabled={!canOpenShell || !activePaneId}
            size={compactChrome ? "icon-xs" : "icon"}
            title={terminalFullscreen ? "退出全屏" : "全屏"}
            type="button"
            variant="soft"
            onClick={toggleTerminalFullscreen}
          >
            {terminalFullscreen ? (
              <Shrink className={compactChrome ? "h-3.5 w-3.5" : "h-4 w-4"} />
            ) : (
              <Fullscreen
                className={compactChrome ? "h-3.5 w-3.5" : "h-4 w-4"}
              />
            )}
          </Button>
        </div>
      </div>
      <div
        className="relative min-h-0 overflow-hidden"
        data-testid={surfaceTestId}
        style={{ backgroundColor: darkPlusTerminalTheme.background }}
      >
        {agent?.runtimeStatus !== "running" ? (
          <div className="grid h-full place-items-center text-sm text-slate-400">
            {agent
              ? `智能体 ${agent.name} 当前${statusLabel(agent.runtimeStatus || "not running")}`
              : "未选择智能体"}
          </div>
        ) : visiblePanes.length === 0 ? (
          <div className="grid h-full place-items-center text-sm text-slate-400">
            正在打开终端
          </div>
        ) : null}
        {agent?.runtimeStatus === "running" && maximizedPane ? (
          <ShellPaneRenderer
            active={active}
            activePaneId={activePaneId}
            key={maximizedPane.id}
            maximized
            pane={maximizedPane}
            status={statuses[maximizedPane.id] ?? "connecting"}
            onContextMenu={handlePaneContextMenu}
            onFocus={setActivePane}
            onHost={setHostNode}
            onToggleMaximize={togglePaneMaximized}
          />
        ) : agent?.runtimeStatus === "running" && currentLayout ? (
          <ShellLayoutRenderer
            active={active}
            activePaneId={activePaneId}
            key={agent.id}
            layout={currentLayout}
            statuses={statuses}
            onContextMenu={handlePaneContextMenu}
            onFocus={setActivePane}
            onHost={setHostNode}
            onToggleMaximize={togglePaneMaximized}
          />
        ) : null}
        {contextMenu.open ? (
          <TerminalContextMenu
            canClose={visiblePanes.length > 1}
            pane={visiblePanes.find((item) => item.id === contextMenu.paneId)}
            x={contextMenu.x}
            y={contextMenu.y}
            onAction={runContextAction}
            onClear={(paneId) => clearTerminal(paneId)}
            onClose={(pane) => closePane(pane)}
            onCopy={(paneId) => copySelection(paneId)}
            onNew={(paneId) => agent && openPane(agent, paneId)}
            onPaste={(paneId) => pasteClipboard(paneId)}
            onRebuild={(pane) => rebuildSession(pane)}
            onSelectAll={(paneId) => selectAll(paneId)}
            onSplit={(paneId, direction) => splitPane(paneId, direction)}
            status={
              contextMenu.paneId ? statuses[contextMenu.paneId] : undefined
            }
          />
        ) : null}
      </div>
    </div>
  );
}

function ShellLayoutRenderer({
  active,
  activePaneId,
  layout,
  statuses,
  onContextMenu,
  onFocus,
  onHost,
  onToggleMaximize,
}: {
  active: boolean;
  activePaneId: string;
  layout: TerminalLayoutNode;
  statuses: Record<string, string>;
  onContextMenu: (event: ReactMouseEvent, pane: TerminalPaneNode) => void;
  onFocus: (pane: TerminalPaneNode) => void;
  onHost: (pane: TerminalPaneNode, node: HTMLDivElement | null) => void;
  onToggleMaximize: (pane: TerminalPaneNode) => void;
}) {
  if (layout.type === "pane") {
    return (
      <ShellPaneRenderer
        active={active}
        activePaneId={activePaneId}
        maximized={false}
        pane={layout}
        status={statuses[layout.id] ?? "connecting"}
        onContextMenu={onContextMenu}
        onFocus={onFocus}
        onHost={onHost}
        onToggleMaximize={onToggleMaximize}
      />
    );
  }
  if (layout.type === "tabs") {
    const activeTabPane = getActiveTabPane(layout);
    if (!activeTabPane) {
      return null;
    }
    return (
      <ShellPaneRenderer
        active={active}
        activePaneId={activePaneId}
        key={activeTabPane.id}
        maximized={false}
        pane={activeTabPane}
        status={statuses[activeTabPane.id] ?? "connecting"}
        onContextMenu={onContextMenu}
        onFocus={onFocus}
        onHost={onHost}
        onToggleMaximize={onToggleMaximize}
      />
    );
  }

  return (
    <ResizablePanelGroup
      autoSaveId={`shell-layout-${layout.id}`}
      className="h-full min-h-0"
      data-testid={`shell-split-${layout.direction}`}
      direction={layout.direction}
    >
      {layout.children.map((child, index) => (
        <Fragment key={child.id}>
          <ResizablePanel
            defaultSize={100 / layout.children.length}
            minSize={18}
          >
            <ShellLayoutRenderer
              active={active}
              activePaneId={activePaneId}
              layout={child}
              statuses={statuses}
              onContextMenu={onContextMenu}
              onFocus={onFocus}
              onHost={onHost}
              onToggleMaximize={onToggleMaximize}
            />
          </ResizablePanel>
          {index < layout.children.length - 1 ? (
            <ResizeHandle
              className="workbench-resize-handle shell-resize-handle"
              data-testid="shell-resize-handle"
            />
          ) : null}
        </Fragment>
      ))}
    </ResizablePanelGroup>
  );
}

function ShellPaneRenderer({
  active,
  activePaneId,
  maximized,
  pane,
  status,
  onContextMenu,
  onFocus,
  onHost,
  onToggleMaximize,
}: {
  active: boolean;
  activePaneId: string;
  maximized: boolean;
  pane: TerminalPaneNode;
  status: string;
  onContextMenu: (event: ReactMouseEvent, pane: TerminalPaneNode) => void;
  onFocus: (pane: TerminalPaneNode) => void;
  onHost: (pane: TerminalPaneNode, node: HTMLDivElement | null) => void;
  onToggleMaximize: (pane: TerminalPaneNode) => void;
}) {
  const label = getPaneLabel(pane);
  return (
    <div
      className={cn(
        "agent-shell-pane relative h-full min-h-0 overflow-hidden border border-transparent bg-[#1e1e1e] transition-colors hover:border-primary/35",
        pane.id === activePaneId &&
          "is-active border-primary/60 hover:border-primary/70",
        maximized && "border-primary/70 hover:border-primary/80",
      )}
      data-active={pane.id === activePaneId}
      data-fullscreen={maximized ? "pane" : "false"}
      data-shell-agent-id={pane.agentId}
      data-shell-pane-id={pane.id}
      data-shell-pane-status={status}
      data-testid="shell-pane"
      onClick={() => onFocus(pane)}
      onContextMenu={(event) => onContextMenu(event, pane)}
    >
      <Button
        aria-label={maximized ? `还原${label}` : `最大化${label}`}
        className="absolute right-2 top-2 z-20 h-7 w-7 border-0 bg-zinc-950/20 text-slate-200 opacity-50 shadow-none backdrop-blur-[1px] hover:bg-zinc-950/35 hover:opacity-90 focus-visible:opacity-100 focus-visible:ring-2 focus-visible:ring-ring/35"
        data-testid="shell-pane-maximize-button"
        size="icon"
        title={maximized ? "还原" : "最大化"}
        type="button"
        variant="soft"
        onClick={(event) => {
          event.stopPropagation();
          onToggleMaximize(pane);
        }}
      >
        {maximized ? (
          <Minimize2 className="h-3.5 w-3.5" />
        ) : (
          <Maximize2 className="h-3.5 w-3.5" />
        )}
      </Button>
      <div
        ref={(node) => onHost(pane, node)}
        className={cn(
          "agent-shell-terminal-host absolute inset-0",
          active ? "block" : "hidden",
        )}
        data-testid="shell-terminal-host"
      />
    </div>
  );
}

function TerminalContextMenu({
  canClose,
  pane,
  x,
  y,
  onAction,
  onClear,
  onClose,
  onCopy,
  onNew,
  onPaste,
  onRebuild,
  onSelectAll,
  onSplit,
  status,
}: {
  canClose: boolean;
  pane?: TerminalPaneNode;
  x: number;
  y: number;
  onAction: (action: () => void | Promise<void>) => void;
  onClear: (paneId: string) => void;
  onClose: (pane: TerminalPaneNode) => void;
  onCopy: (paneId: string) => Promise<void>;
  onNew: (paneId: string) => void;
  onPaste: (paneId: string) => Promise<void>;
  onRebuild: (pane: TerminalPaneNode) => void;
  onSelectAll: (paneId: string) => void;
  onSplit: (paneId: string, direction: "horizontal" | "vertical") => void;
  status?: string;
}) {
  if (!pane) {
    return null;
  }
  const viewportPadding = 8;
  const menuWidth = 220;
  const menuHeight = 292;
  const left = Math.min(
    Math.max(viewportPadding, x),
    window.innerWidth - menuWidth - viewportPadding,
  );
  const top = Math.min(
    Math.max(viewportPadding, y),
    window.innerHeight - menuHeight - viewportPadding,
  );
  return (
    <div
      className="fixed z-50 grid min-w-56 justify-items-stretch rounded-[6px] border border-border bg-popover p-1 text-left text-sm text-popover-foreground shadow-lg"
      data-testid="terminal-context-menu"
      role="menu"
      style={{ left, top }}
      onClick={(event) => event.stopPropagation()}
      onContextMenu={(event) => event.preventDefault()}
    >
      <TerminalMenuButton
        icon={<Copy />}
        label="复制"
        onClick={() => onAction(() => onCopy(pane.id))}
      />
      <TerminalMenuButton
        icon={<ClipboardPaste />}
        label="粘贴"
        onClick={() => onAction(() => onPaste(pane.id))}
      />
      <TerminalMenuButton
        icon={<ScanText />}
        label="选择全部"
        onClick={() => onAction(() => onSelectAll(pane.id))}
      />
      <TerminalMenuButton
        icon={<Eraser />}
        label="清屏"
        onClick={() => onAction(() => onClear(pane.id))}
      />
      <TerminalMenuSeparator />
      <TerminalMenuButton
        icon={<Plus />}
        label="新建终端"
        onClick={() => onAction(() => onNew(pane.id))}
      />
      <TerminalMenuButton
        icon={<PanelRight />}
        label="水平拆分终端"
        onClick={() => onAction(() => onSplit(pane.id, "horizontal"))}
      />
      <TerminalMenuButton
        icon={<PanelTop />}
        label="垂直拆分终端"
        onClick={() => onAction(() => onSplit(pane.id, "vertical"))}
      />
      <TerminalMenuButton
        disabled={status !== "error" && status !== "closed"}
        icon={<RefreshCcw />}
        label="重建终端"
        onClick={() => onAction(() => onRebuild(pane))}
      />
      <TerminalMenuSeparator />
      <TerminalMenuButton
        danger
        disabled={!canClose}
        icon={<X />}
        label="关闭终端"
        onClick={() => onAction(() => onClose(pane))}
      />
    </div>
  );
}

function TerminalMenuButton({
  danger,
  disabled,
  icon,
  label,
  onClick,
}: {
  danger?: boolean;
  disabled?: boolean;
  icon: ReactNode;
  label: string;
  onClick: () => void;
}) {
  return (
    <Button
      className={cn(
        "flex h-8 w-full min-w-0 justify-start gap-2 rounded-[4px] px-2 text-left text-sm hover:bg-accent disabled:pointer-events-none disabled:opacity-45 [&_svg]:h-4 [&_svg]:w-4 [&_svg]:shrink-0",
        danger && "text-destructive hover:bg-destructive/10",
      )}
      disabled={disabled}
      role="menuitem"
      type="button"
      variant="ghost"
      onClick={onClick}
    >
      {icon}
      <span className="truncate">{label}</span>
    </Button>
  );
}

function TerminalMenuSeparator() {
  return <div className="my-1 h-px bg-border" role="separator" />;
}

function getActivePaneId(
  agent: AgentInfo | undefined,
  panes: TerminalPaneNode[],
  activeByAgent: Record<string, string>,
) {
  if (!agent || agent.runtimeStatus !== "running") {
    return "";
  }
  const current = activeByAgent[agent.id];
  return panes.some((pane) => pane.id === current)
    ? current
    : (panes[0]?.id ?? "");
}

function getNextPaneSequence(
  layout: TerminalLayoutNode | undefined,
  agentId: string,
) {
  return (
    collectPanes(layout).reduce(
      (max, pane) =>
        pane.agentId === agentId ? Math.max(max, pane.sequence) : max,
      0,
    ) + 1
  );
}

function getPaneLabel(pane: TerminalPaneNode) {
  return `终端 ${pane.sequence}`;
}

function terminalPaneStatusTone(status: string | undefined) {
  if (status === "connected") {
    return "online";
  }
  if (status === "error") {
    return "error";
  }
  return "waiting";
}

function statusLabel(value: string) {
  const labels: Record<string, string> = {
    idle: "空闲",
    offline: "离线",
    starting: "启动中",
    connecting: "连接中",
    connected: "已连接",
    recovering: "恢复中",
    detached: "待重建",
    rebuilding: "重建中",
    closed: "已关闭",
    error: "异常",
    running: "运行中",
    stopped: "已停止",
    exited: "已退出",
    paused: "已暂停",
    missing: "容器离线",
    unavailable: "Docker 不可用",
    dead: "异常退出",
    created: "未启动",
  };
  return labels[value] ?? value;
}

type PersistedShellState = {
  layoutsByAgent: Record<string, TerminalLayoutNode>;
  activeByAgent: Record<string, string>;
};

type LegacyShellTab = {
  id: string;
  agentId: string;
  sequence: number;
};

function loadPersistedShellState(
  storageKey = shellTabsStorageKey,
): PersistedShellState {
  if (typeof window === "undefined") {
    return { layoutsByAgent: {}, activeByAgent: {} };
  }

  try {
    const rawState = window.localStorage.getItem(storageKey);
    if (!rawState) {
      return { layoutsByAgent: {}, activeByAgent: {} };
    }
    const parsed = JSON.parse(rawState) as Record<string, unknown>;
    const layoutsByAgent = parsePersistedLayouts(parsed.layoutsByAgent);
    const activeByAgent = parsePersistedActiveShellPanes(parsed.activeByAgent);
    if (Object.keys(layoutsByAgent).length > 0) {
      return { layoutsByAgent, activeByAgent };
    }
    return migrateLegacyTabs(
      parsePersistedShellTabs(parsed.tabs),
      activeByAgent,
    );
  } catch {
    return { layoutsByAgent: {}, activeByAgent: {} };
  }
}

function savePersistedShellState(
  storageKey: string,
  state: PersistedShellState,
) {
  if (typeof window === "undefined") {
    return;
  }

  try {
    window.localStorage.setItem(storageKey, JSON.stringify(state));
  } catch {
    return;
  }
}

function parsePersistedLayouts(
  value: unknown,
): Record<string, TerminalLayoutNode> {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return {};
  }
  return Object.entries(value as Record<string, unknown>).reduce<
    Record<string, TerminalLayoutNode>
  >((result, [agentId, node]) => {
    const parsed = parseLayoutNode(node, agentId);
    if (parsed && collectPanes(parsed).length > 0) {
      result[agentId] = parsed;
    }
    return result;
  }, {});
}

function parseLayoutNode(
  value: unknown,
  fallbackAgentId: string,
  path = "root",
): TerminalLayoutNode | null {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return null;
  }
  const candidate = value as Record<string, unknown>;
  if (candidate.type === "pane") {
    const id = typeof candidate.id === "string" ? candidate.id : "";
    const agentId =
      typeof candidate.agentId === "string"
        ? candidate.agentId
        : fallbackAgentId;
    const sequence =
      typeof candidate.sequence === "number" &&
      Number.isFinite(candidate.sequence)
        ? Math.max(1, Math.floor(candidate.sequence))
        : 1;
    const workingDir =
      typeof candidate.workingDir === "string" && candidate.workingDir
        ? normalizeWorkspacePath(candidate.workingDir)
        : workspaceRootPath;
    return id && agentId
      ? { type: "pane", id, agentId, sequence, workingDir }
      : null;
  }
  if (candidate.type === "split") {
    const id =
      typeof candidate.id === "string" && candidate.id
        ? candidate.id
        : `split-${fallbackAgentId}-${path}`;
    const direction =
      candidate.direction === "vertical" ? "vertical" : "horizontal";
    const children = Array.isArray(candidate.children)
      ? candidate.children
          .map((child, index) =>
            parseLayoutNode(child, fallbackAgentId, `${path}-${index}`),
          )
          .filter(Boolean)
      : [];
    return children.length > 0
      ? collapseLayout({
          type: "split",
          id,
          direction,
          children: children as TerminalLayoutNode[],
        })
      : null;
  }
  if (candidate.type === "tabs") {
    const id =
      typeof candidate.id === "string" && candidate.id
        ? candidate.id
        : `tabs-${fallbackAgentId}-${path}`;
    const children = Array.isArray(candidate.children)
      ? candidate.children
          .map((child, index) =>
            parseLayoutNode(child, fallbackAgentId, `${path}-${index}`),
          )
          .flatMap((child) => collectPanes(child ?? undefined))
      : [];
    if (children.length === 0) {
      return null;
    }
    const activePaneId =
      typeof candidate.activePaneId === "string" &&
      children.some((pane) => pane.id === candidate.activePaneId)
        ? candidate.activePaneId
        : children[0].id;
    return collapseTabs({ type: "tabs", id, activePaneId, children });
  }
  return null;
}

function parsePersistedShellTabs(value: unknown): LegacyShellTab[] {
  if (!Array.isArray(value)) {
    return [];
  }
  const counts = new Map<string, number>();
  return value.reduce<LegacyShellTab[]>((result, item) => {
    if (!item || typeof item !== "object" || Array.isArray(item)) {
      return result;
    }
    const candidate = item as Record<string, unknown>;
    const id = typeof candidate.id === "string" ? candidate.id : "";
    const agentId =
      typeof candidate.agentId === "string" ? candidate.agentId : "";
    const sequence =
      typeof candidate.sequence === "number" &&
      Number.isFinite(candidate.sequence)
        ? Math.max(1, Math.floor(candidate.sequence))
        : 1;
    if (!id || !agentId) {
      return result;
    }
    const currentCount = counts.get(agentId) ?? 0;
    if (currentCount >= maxPanesPerAgent) {
      return result;
    }
    counts.set(agentId, currentCount + 1);
    result.push({ id, agentId, sequence });
    return result;
  }, []);
}

function parsePersistedActiveShellPanes(
  value: unknown,
): Record<string, string> {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return {};
  }
  return Object.entries(value as Record<string, unknown>).reduce<
    Record<string, string>
  >((result, [agentId, terminalId]) => {
    if (agentId && typeof terminalId === "string" && terminalId) {
      result[agentId] = terminalId;
    }
    return result;
  }, {});
}

function migrateLegacyTabs(
  tabs: LegacyShellTab[],
  activeByAgent: Record<string, string>,
): PersistedShellState {
  const grouped = tabs.reduce<Record<string, LegacyShellTab[]>>(
    (result, tab) => {
      result[tab.agentId] = [...(result[tab.agentId] ?? []), tab];
      return result;
    },
    {},
  );
  const layoutsByAgent = Object.entries(grouped).reduce<
    Record<string, TerminalLayoutNode>
  >((result, [agentId, agentTabs]) => {
    const panes = agentTabs
      .slice(0, maxPanesPerAgent)
      .map<TerminalPaneNode>((tab) => ({
        type: "pane",
        id: tab.id,
        agentId: tab.agentId,
        sequence: tab.sequence,
        workingDir: workspaceRootPath,
      }));
    if (panes.length === 1) {
      result[agentId] = panes[0];
    } else if (panes.length > 1) {
      const activePaneId = panes.some(
        (pane) => pane.id === activeByAgent[agentId],
      )
        ? activeByAgent[agentId]
        : panes[0].id;
      result[agentId] = {
        type: "tabs",
        id: `migrated-${agentId}`,
        activePaneId,
        children: panes,
      };
    }
    return result;
  }, {});
  return { layoutsByAgent, activeByAgent };
}

function collectPanes(
  node: TerminalLayoutNode | undefined,
): TerminalPaneNode[] {
  if (!node) {
    return [];
  }
  if (node.type === "pane") {
    return [node];
  }
  if (node.type === "tabs") {
    return node.children;
  }
  return node.children.flatMap(collectPanes);
}

function collectRenderedPanes(node: TerminalLayoutNode): TerminalPaneNode[] {
  if (node.type === "pane") {
    return [node];
  }
  if (node.type === "tabs") {
    const activeTabPane = getActiveTabPane(node);
    return activeTabPane ? [activeTabPane] : [];
  }
  return node.children.flatMap(collectRenderedPanes);
}

function appendPane(
  layout: TerminalLayoutNode,
  activePaneId: string,
  pane: TerminalPaneNode,
): TerminalLayoutNode {
  if (layout.type === "pane") {
    return {
      type: "tabs",
      id: `tabs-${layout.id}-${pane.id}`,
      activePaneId: pane.id,
      children: [layout, pane],
    };
  }
  if (layout.type === "tabs") {
    return {
      ...layout,
      activePaneId: pane.id,
      children: [...layout.children, pane],
    };
  }
  const targetChildIndex = layout.children.findIndex((child) =>
    layoutContainsPane(child, activePaneId),
  );
  if (targetChildIndex < 0) {
    return {
      type: "tabs",
      id: `tabs-${pane.id}`,
      activePaneId: pane.id,
      children: [...collectPanes(layout), pane],
    };
  }
  return {
    ...layout,
    children: layout.children.map((child, index) =>
      index === targetChildIndex
        ? appendPane(child, activePaneId, pane)
        : child,
    ),
  };
}

function splitLayoutPane(
  layout: TerminalLayoutNode,
  paneId: string,
  direction: "horizontal" | "vertical",
  pane: TerminalPaneNode,
): TerminalLayoutNode {
  if (layout.type === "pane") {
    if (layout.id !== paneId) {
      return layout;
    }
    return {
      type: "split",
      id: `split-${layout.id}-${pane.id}`,
      direction,
      children: [layout, pane],
    };
  }
  if (layout.type === "tabs") {
    if (!layout.children.some((child) => child.id === paneId)) {
      return layout;
    }
    return {
      type: "split",
      id: `split-${paneId}-${pane.id}`,
      direction,
      children: [layout, pane],
    };
  }
  return {
    ...layout,
    children: layout.children.map((child) =>
      splitLayoutPane(child, paneId, direction, pane),
    ),
  };
}

function removePaneFromLayout(
  layout: TerminalLayoutNode,
  paneId: string,
): TerminalLayoutNode | null {
  if (layout.type === "pane") {
    return layout.id === paneId ? null : layout;
  }
  if (layout.type === "tabs") {
    const children = layout.children.filter((child) => child.id !== paneId);
    if (children.length === 0) {
      return null;
    }
    const activePaneId = children.some(
      (child) => child.id === layout.activePaneId,
    )
      ? layout.activePaneId
      : children[0].id;
    return collapseTabs({ ...layout, activePaneId, children });
  }
  const children = layout.children
    .map((child) => removePaneFromLayout(child, paneId))
    .filter(Boolean) as TerminalLayoutNode[];
  if (children.length === 0) {
    return null;
  }
  return collapseLayout({ ...layout, children });
}

function collapseLayout(layout: TerminalSplitNode): TerminalLayoutNode {
  if (layout.children.length === 1) {
    return layout.children[0];
  }
  return layout;
}

function collapseTabs(layout: TerminalTabsNode): TerminalLayoutNode {
  if (layout.children.length === 1) {
    return layout.children[0];
  }
  return layout;
}

function getActiveTabPane(layout: TerminalTabsNode) {
  return (
    layout.children.find((pane) => pane.id === layout.activePaneId) ??
    layout.children[0]
  );
}

function layoutContainsPane(
  layout: TerminalLayoutNode,
  paneId: string,
): boolean {
  if (layout.type === "pane") {
    return layout.id === paneId;
  }
  return layout.children.some((child) => layoutContainsPane(child, paneId));
}

function activatePaneInLayout(
  layout: TerminalLayoutNode,
  paneId: string,
): TerminalLayoutNode {
  if (layout.type === "pane") {
    return layout;
  }
  if (layout.type === "tabs") {
    if (
      !layout.children.some((pane) => pane.id === paneId) ||
      layout.activePaneId === paneId
    ) {
      return layout;
    }
    return { ...layout, activePaneId: paneId };
  }
  let changed = false;
  const children = layout.children.map((child) => {
    if (!layoutContainsPane(child, paneId)) {
      return child;
    }
    const nextChild = activatePaneInLayout(child, paneId);
    changed = changed || nextChild !== child;
    return nextChild;
  });
  return changed ? { ...layout, children } : layout;
}

async function exitDocumentFullscreen() {
  if (document.fullscreenElement) {
    await document.exitFullscreen();
  }
}

function applyTerminalHostStyle(host: HTMLDivElement, fontSize: number) {
  host.style.setProperty("--agent-shell-terminal-font-size", `${fontSize}px`);
  host.style.setProperty(
    "--agent-shell-terminal-selection-background",
    darkPlusTerminalSelectionBackground,
  );
  host.style.setProperty(
    "--agent-shell-terminal-selection-foreground",
    darkPlusTerminalSelectionForeground,
  );
}

function isTerminalCopyShortcut(event: KeyboardEvent) {
  if (event.defaultPrevented || event.altKey) {
    return false;
  }
  return (event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "c";
}

async function copyTerminalText(
  text: string,
  options: { showSuccess: boolean },
) {
  let clipboardError: Error | undefined;
  if (navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(text);
      if (options.showSuccess) {
        toast.success("已复制终端选中文本");
      }
      return;
    } catch (error) {
      clipboardError = error as Error;
    }
  }
  if (copyTerminalTextWithTextarea(text)) {
    if (options.showSuccess) {
      toast.success("已复制终端选中文本");
    }
    return;
  }
  throw clipboardError ?? new Error("浏览器拒绝访问剪贴板");
}

function copyTerminalTextWithTextarea(text: string) {
  const previousActiveElement =
    document.activeElement instanceof HTMLElement
      ? document.activeElement
      : null;
  const textarea = document.createElement("textarea");
  textarea.value = text;
  textarea.setAttribute("readonly", "true");
  textarea.style.position = "fixed";
  textarea.style.left = "-9999px";
  textarea.style.top = "0";
  textarea.style.opacity = "0";
  textarea.style.pointerEvents = "none";
  document.body.appendChild(textarea);
  textarea.focus();
  textarea.select();
  try {
    return document.execCommand("copy");
  } catch {
    return false;
  } finally {
    document.body.removeChild(textarea);
    try {
      previousActiveElement?.focus({ preventScroll: true });
    } catch {
      // The previous focus target may have been removed with the terminal menu.
    }
  }
}

function stripTerminalBrowserControlledModes(data: string) {
  return data.replace(
    /\x1b\[\?([0-9;]+)([hl])/g,
    (sequence, rawParams: string, command: string) => {
      const params = rawParams.split(";");
      const retainedParams = params.filter(
        (param) => !terminalBrowserControlledModeParams.has(param),
      );
      if (retainedParams.length === params.length) {
        return sequence;
      }
      return retainedParams.length > 0
        ? `\x1b[?${retainedParams.join(";")}${command}`
        : "";
    },
  );
}

function sanitizeTerminalOutput(session: ShellSession, data: string) {
  const combined = session.pendingTerminalOutput + data;
  const { complete, pending } = splitTrailingIncompleteAnsiSequence(combined);
  session.pendingTerminalOutput = pending;
  return stripTerminalBrowserControlledModes(complete);
}

function stripTerminalGeneratedResponses(data: string) {
  let input = data;
  for (const sequence of terminalGeneratedResponseSequences) {
    input = input.replaceAll(sequence, "");
  }
  return input;
}

function isTerminalImeInputEvent(event: InputEvent) {
  return (
    event.isComposing ||
    event.inputType === "insertCompositionText" ||
    event.inputType === "insertFromComposition"
  );
}

function isTerminalImeCommitInput(event: InputEvent) {
  return event.inputType === "insertFromComposition";
}

function installTerminalImeInputFallback(
  terminal: Terminal,
  sendInput: (data: string) => void,
): TerminalImeInputFallback {
  const textarea = terminal.textarea;
  if (!textarea) {
    return {
      dispose: () => undefined,
      handleTerminalData: () => true,
    };
  }

  let compositionSequence = 0;
  let pendingCompositionSequence = 0;
  let compositionStartValue = "";
  let compositionPreviewText = "";
  let compositionInputText = "";
  let compositionCommitText = "";
  let terminalDataAfterCompositionStart = "";
  let finalizeTimer = 0;
  let verifyTimer = 0;

  const clearTimers = () => {
    window.clearTimeout(finalizeTimer);
    window.clearTimeout(verifyTimer);
    finalizeTimer = 0;
    verifyTimer = 0;
  };

  const handleCompositionStart = () => {
    clearTimers();
    compositionSequence += 1;
    pendingCompositionSequence = compositionSequence;
    compositionStartValue = textarea.value;
    compositionPreviewText = "";
    compositionInputText = "";
    compositionCommitText = "";
    terminalDataAfterCompositionStart = "";
  };

  const ensureCompositionSequence = () => {
    if (pendingCompositionSequence) {
      return pendingCompositionSequence;
    }
    compositionSequence += 1;
    pendingCompositionSequence = compositionSequence;
    compositionStartValue = textarea.value;
    compositionPreviewText = "";
    compositionInputText = "";
    compositionCommitText = "";
    terminalDataAfterCompositionStart = "";
    return pendingCompositionSequence;
  };

  const compositionValueDelta = (startValue = compositionStartValue) => {
    const value = textarea.value;
    return value.startsWith(startValue) ? value.slice(startValue.length) : "";
  };

  const rememberCompositionInput = (eventData: string) => {
    const valueDelta = compositionValueDelta();
    const candidate = valueDelta || eventData;
    if (candidate) {
      compositionInputText = candidate;
    }
  };

  const rememberCompositionPreview = (eventData: string) => {
    if (eventData) {
      compositionPreviewText = eventData;
    }
  };

  const rememberCompositionCommit = (
    eventData: string,
    options: { useValueDelta?: boolean } = {},
  ) => {
    const candidate =
      eventData || (options.useValueDelta ? compositionValueDelta() : "");
    if (candidate && (!compositionCommitText || eventData)) {
      compositionCommitText = candidate;
    }
  };

  const scheduleCompositionCommit = (sequence: number, eventData: string) => {
    const startValue = compositionStartValue;
    clearTimers();
    finalizeTimer = window.setTimeout(() => {
      if (pendingCompositionSequence !== sequence) {
        return;
      }
      const valueDelta = compositionValueDelta(startValue);
      const committedText =
        compositionCommitText ||
        valueDelta ||
        eventData ||
        compositionInputText;
      if (!committedText) {
        pendingCompositionSequence = 0;
        compositionPreviewText = "";
        compositionInputText = "";
        compositionCommitText = "";
        terminalDataAfterCompositionStart = "";
        return;
      }
      verifyTimer = window.setTimeout(() => {
        if (pendingCompositionSequence !== sequence) {
          return;
        }
        if (!terminalDataAfterCompositionStart.includes(committedText)) {
          sendInput(committedText);
        }
        pendingCompositionSequence = 0;
        compositionPreviewText = "";
        compositionInputText = "";
        compositionCommitText = "";
        terminalDataAfterCompositionStart = "";
      }, terminalImeFallbackDelayMs);
    }, 0);
  };

  const handleCompositionEnd = (event: CompositionEvent) => {
    const sequence = pendingCompositionSequence || (compositionSequence += 1);
    pendingCompositionSequence = sequence;
    rememberCompositionCommit(event.data || "");
    scheduleCompositionCommit(sequence, event.data || "");
  };

  const isLikelyCompositionCommit = (event: InputEvent) => {
    return (
      isTerminalImeCommitInput(event) ||
      Boolean(pendingCompositionSequence && event.data && !event.isComposing)
    );
  };

  const handleCompositionUpdate = (event: CompositionEvent) => {
    rememberCompositionPreview(event.data || "");
  };

  const handleBeforeInput = (event: Event) => {
    const inputEvent = event as InputEvent;
    if (
      !isTerminalImeInputEvent(inputEvent) &&
      !isLikelyCompositionCommit(inputEvent)
    ) {
      return;
    }
    const sequence = ensureCompositionSequence();
    if (isLikelyCompositionCommit(inputEvent)) {
      rememberCompositionCommit(inputEvent.data || "", { useValueDelta: true });
      scheduleCompositionCommit(sequence, inputEvent.data || "");
      return;
    }
    rememberCompositionInput(inputEvent.data || "");
  };

  const handleInput = (event: Event) => {
    const inputEvent = event as InputEvent;
    if (!isTerminalImeInputEvent(inputEvent)) {
      if (isLikelyCompositionCommit(inputEvent)) {
        const sequence = ensureCompositionSequence();
        rememberCompositionCommit(inputEvent.data || "", {
          useValueDelta: true,
        });
        scheduleCompositionCommit(sequence, inputEvent.data || "");
      }
      return;
    }
    const sequence = ensureCompositionSequence();
    if (isLikelyCompositionCommit(inputEvent)) {
      rememberCompositionCommit(inputEvent.data || "", { useValueDelta: true });
      scheduleCompositionCommit(sequence, inputEvent.data || "");
      return;
    }
    rememberCompositionInput(inputEvent.data || "");
  };

  textarea.addEventListener("compositionstart", handleCompositionStart);
  textarea.addEventListener("compositionupdate", handleCompositionUpdate);
  textarea.addEventListener("compositionend", handleCompositionEnd);
  textarea.addEventListener("beforeinput", handleBeforeInput, true);
  textarea.addEventListener("input", handleInput, true);

  return {
    dispose: () => {
      clearTimers();
      textarea.removeEventListener("compositionstart", handleCompositionStart);
      textarea.removeEventListener(
        "compositionupdate",
        handleCompositionUpdate,
      );
      textarea.removeEventListener("compositionend", handleCompositionEnd);
      textarea.removeEventListener("beforeinput", handleBeforeInput, true);
      textarea.removeEventListener("input", handleInput, true);
    },
    handleTerminalData: (data: string) => {
      if (pendingCompositionSequence) {
        terminalDataAfterCompositionStart += data;
        const intermediateText = compositionInputText || compositionPreviewText;
        if (
          compositionCommitText &&
          intermediateText &&
          data.includes(intermediateText) &&
          !data.includes(compositionCommitText)
        ) {
          return false;
        }
      }
      return true;
    },
  };
}

function handleTerminalWheel(
  terminal: Terminal,
  event: WheelEvent,
  options: {
    onPtyWheel: (data: string) => void;
  },
) {
  if (event.deltaY === 0) {
    return true;
  }
  return forwardTerminalWheelToPty(terminal, event, options);
}

function forwardTerminalWheelToPty(
  terminal: Terminal,
  event: WheelEvent,
  options: {
    onPtyWheel: (data: string) => void;
  },
) {
  const sequence = terminalWheelSgrSequence(terminal, event);
  if (!sequence) {
    return true;
  }
  options.onPtyWheel(sequence);
  event.preventDefault();
  event.stopPropagation();
  return false;
}

function terminalWheelSgrSequence(terminal: Terminal, event: WheelEvent) {
  const element = terminal.element;
  const rowElement = element?.querySelector(".xterm-rows > div");
  if (!element || !rowElement) {
    return "";
  }
  const viewport = element.getBoundingClientRect();
  const row = rowElement.getBoundingClientRect();
  const rowHeight = row.height || 16;
  const colWidth = row.width / Math.max(1, terminal.cols) || 8;
  const col = clampInteger(
    Math.floor((event.clientX - viewport.left) / colWidth) + 1,
    1,
    Math.max(1, terminal.cols),
  );
  const line = clampInteger(
    Math.floor((event.clientY - viewport.top) / rowHeight) + 1,
    1,
    Math.max(1, terminal.rows),
  );
  const buttonCode = event.deltaY < 0 ? 64 : 65;
  return `\x1b[<${buttonCode};${col};${line}M`;
}

function clampInteger(value: number, min: number, max: number) {
  if (!Number.isFinite(value)) {
    return min;
  }
  return Math.min(max, Math.max(min, value));
}

function splitTrailingIncompleteAnsiSequence(data: string) {
  const escapeIndex = data.lastIndexOf("\x1b");
  if (escapeIndex < 0) {
    return { complete: data, pending: "" };
  }
  const tail = data.slice(escapeIndex);
  if (/^\x1b(?:\[?\??[0-9;]*)?$/.test(tail)) {
    return { complete: data.slice(0, escapeIndex), pending: tail };
  }
  return { complete: data, pending: "" };
}
