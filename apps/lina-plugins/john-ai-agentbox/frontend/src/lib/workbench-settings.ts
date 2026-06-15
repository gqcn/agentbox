export type GitDiffDisplayMode = "side-by-side" | "inline";
export type TerminalCursorStyle = "block" | "underline" | "bar";
export type TerminalRightClickBehavior = "menu" | "paste";
export type WordWrapMode = "off" | "on";
export type AgentDetailPanelId =
  | "chat"
  | "shell"
  | "services"
  | "skills"
  | "files"
  | "git";

export type AgentDetailPanelPreference = {
  id: AgentDetailPanelId;
  visible: boolean;
};

export type WorkbenchSettings = {
  filesSidebarSize: number;
  gitSidebarSize: number;
  agentDetailPanels: AgentDetailPanelPreference[];
  editorFontSize: number;
  editorTabSize: number;
  editorWordWrap: WordWrapMode;
  editorMinimap: boolean;
  gitDiffDisplay: GitDiffDisplayMode;
  gitDiffInlineOnNarrow: boolean;
  codeHighlighting: boolean;
  terminalFontSize: number;
  terminalLineHeight: number;
  terminalCursorStyle: TerminalCursorStyle;
  terminalCursorBlink: boolean;
  terminalCursorWidth: number;
  terminalRightClickBehavior: TerminalRightClickBehavior;
};

export const workbenchSettingsStorageKey =
  "john-ai-agentbox-workbench-settings";
export const workbenchSettingsServerKey = "workbench";
export const agentDetailPanelIds = [
  "chat",
  "shell",
  "files",
  "git",
  "services",
  "skills",
] as const satisfies readonly AgentDetailPanelId[];
const workspaceBrowserAdjacentPanelIds = [
  "services",
  "skills",
] as const satisfies readonly AgentDetailPanelId[];
export const defaultAgentDetailPanelPreferences: AgentDetailPanelPreference[] =
  agentDetailPanelIds.map((id) => ({ id, visible: true }));

export const defaultWorkbenchSettings: WorkbenchSettings = {
  filesSidebarSize: 20,
  gitSidebarSize: 20,
  agentDetailPanels: defaultAgentDetailPanelPreferences,
  editorFontSize: 14,
  editorTabSize: 4,
  editorWordWrap: "on",
  editorMinimap: false,
  gitDiffDisplay: "side-by-side",
  gitDiffInlineOnNarrow: true,
  codeHighlighting: true,
  terminalFontSize: 13,
  terminalLineHeight: 1.15,
  terminalCursorStyle: "block",
  terminalCursorBlink: true,
  terminalCursorWidth: 1,
  terminalRightClickBehavior: "menu",
};

export function loadWorkbenchSettings(): WorkbenchSettings {
  return loadLocalWorkbenchSettings() ?? defaultWorkbenchSettings;
}

export function loadLocalWorkbenchSettings(): WorkbenchSettings | null {
  try {
    const raw = localStorage.getItem(workbenchSettingsStorageKey);
    if (!raw) return null;
    return normalizeWorkbenchSettings(
      JSON.parse(raw) as Partial<WorkbenchSettings>,
    );
  } catch {
    return null;
  }
}

export function saveWorkbenchSettings(settings: WorkbenchSettings) {
  localStorage.setItem(workbenchSettingsStorageKey, JSON.stringify(settings));
}

export function encodeWorkbenchSettings(settings: WorkbenchSettings) {
  return JSON.stringify(normalizeWorkbenchSettings(settings));
}

export function decodeWorkbenchSettings(value: string): WorkbenchSettings {
  return normalizeWorkbenchSettings(
    JSON.parse(value) as Partial<WorkbenchSettings>,
  );
}

export function equalWorkbenchSettings(
  left: WorkbenchSettings,
  right: WorkbenchSettings,
) {
  return encodeWorkbenchSettings(left) === encodeWorkbenchSettings(right);
}

export function normalizeWorkbenchSettings(
  candidate: Partial<WorkbenchSettings>,
): WorkbenchSettings {
  return {
    filesSidebarSize: numberOrFallback(
      candidate.filesSidebarSize,
      defaultWorkbenchSettings.filesSidebarSize,
    ),
    gitSidebarSize: numberOrFallback(
      candidate.gitSidebarSize,
      defaultWorkbenchSettings.gitSidebarSize,
    ),
    agentDetailPanels: normalizeAgentDetailPanelPreferences(
      candidate.agentDetailPanels,
    ),
    editorFontSize: clampNumber(
      candidate.editorFontSize,
      11,
      20,
      defaultWorkbenchSettings.editorFontSize,
    ),
    editorTabSize: clampNumber(
      candidate.editorTabSize,
      2,
      8,
      defaultWorkbenchSettings.editorTabSize,
    ),
    editorWordWrap: candidate.editorWordWrap === "off" ? "off" : "on",
    editorMinimap: Boolean(candidate.editorMinimap),
    gitDiffDisplay:
      candidate.gitDiffDisplay === "inline" ? "inline" : "side-by-side",
    gitDiffInlineOnNarrow: candidate.gitDiffInlineOnNarrow !== false,
    codeHighlighting: candidate.codeHighlighting !== false,
    terminalFontSize: clampNumber(
      candidate.terminalFontSize,
      10,
      24,
      defaultWorkbenchSettings.terminalFontSize,
    ),
    terminalLineHeight: clampDecimal(
      candidate.terminalLineHeight,
      1,
      2,
      defaultWorkbenchSettings.terminalLineHeight,
      2,
    ),
    terminalCursorStyle: isTerminalCursorStyle(candidate.terminalCursorStyle)
      ? candidate.terminalCursorStyle
      : defaultWorkbenchSettings.terminalCursorStyle,
    terminalCursorBlink: candidate.terminalCursorBlink !== false,
    terminalCursorWidth: clampNumber(
      candidate.terminalCursorWidth,
      1,
      6,
      defaultWorkbenchSettings.terminalCursorWidth,
    ),
    terminalRightClickBehavior:
      candidate.terminalRightClickBehavior === "paste" ? "paste" : "menu",
  };
}

export function normalizeAgentDetailPanelPreferences(
  value: unknown,
): AgentDetailPanelPreference[] {
  if (!Array.isArray(value)) {
    return defaultAgentDetailPanelPreferences.map(
      copyAgentDetailPanelPreference,
    );
  }
  const seen = new Set<AgentDetailPanelId>();
  const normalized: AgentDetailPanelPreference[] = [];
  for (const item of value) {
    if (!item || typeof item !== "object" || Array.isArray(item)) {
      continue;
    }
    const candidate = item as Partial<AgentDetailPanelPreference>;
    if (!isAgentDetailPanelId(candidate.id) || seen.has(candidate.id)) {
      continue;
    }
    seen.add(candidate.id);
    normalized.push({ id: candidate.id, visible: candidate.visible !== false });
  }
  for (const fallback of defaultAgentDetailPanelPreferences) {
    if (!seen.has(fallback.id)) {
      insertMissingAgentDetailPanelPreference(
        normalized,
        copyAgentDetailPanelPreference(fallback),
      );
      seen.add(fallback.id);
    }
  }
  if (!normalized.some((item) => item.visible)) {
    return defaultAgentDetailPanelPreferences.map(
      copyAgentDetailPanelPreference,
    );
  }
  return pinWorkspaceBrowserAdjacentPanels(normalized);
}

function insertMissingAgentDetailPanelPreference(
  items: AgentDetailPanelPreference[],
  fallback: AgentDetailPanelPreference,
) {
  const fallbackIndex = agentDetailPanelIds.indexOf(fallback.id);
  if (fallbackIndex < 0) {
    items.push(fallback);
    return;
  }
  const successorId = agentDetailPanelIds
    .slice(fallbackIndex + 1)
    .find((id) => items.some((item) => item.id === id));
  if (successorId) {
    const successorIndex = items.findIndex((item) => item.id === successorId);
    items.splice(successorIndex, 0, fallback);
    return;
  }
  const predecessorId = agentDetailPanelIds
    .slice(0, fallbackIndex)
    .reverse()
    .find((id) => items.some((item) => item.id === id));
  if (predecessorId) {
    const predecessorIndex = items.findIndex(
      (item) => item.id === predecessorId,
    );
    items.splice(predecessorIndex + 1, 0, fallback);
    return;
  }
  items.push(fallback);
}

function pinWorkspaceBrowserAdjacentPanels(
  items: AgentDetailPanelPreference[],
) {
  const pinned = workspaceBrowserAdjacentPanelIds
    .map((id) => items.find((item) => item.id === id))
    .filter((item): item is AgentDetailPanelPreference => Boolean(item));
  if (pinned.length === 0) {
    return items;
  }
  const pinnedIds = new Set<AgentDetailPanelId>(
    workspaceBrowserAdjacentPanelIds,
  );
  return [
    ...items.filter((item) => !pinnedIds.has(item.id)),
    ...pinned.map(copyAgentDetailPanelPreference),
  ];
}

function copyAgentDetailPanelPreference(
  value: AgentDetailPanelPreference,
): AgentDetailPanelPreference {
  return { id: value.id, visible: value.visible };
}

function clampNumber(
  value: unknown,
  min: number,
  max: number,
  fallback: number,
) {
  const numeric = typeof value === "number" ? value : Number(value);
  if (!Number.isFinite(numeric)) return fallback;
  return Math.min(max, Math.max(min, Math.round(numeric)));
}

function numberOrFallback(value: unknown, fallback: number) {
  if (value === undefined || value === null || value === "") return fallback;
  const numeric = typeof value === "number" ? value : Number(value);
  return Number.isFinite(numeric) ? numeric : fallback;
}

function clampDecimal(
  value: unknown,
  min: number,
  max: number,
  fallback: number,
  precision: number,
) {
  const numeric = typeof value === "number" ? value : Number(value);
  if (!Number.isFinite(numeric)) return fallback;
  const multiplier = 10 ** precision;
  return Math.min(
    max,
    Math.max(min, Math.round(numeric * multiplier) / multiplier),
  );
}

function isTerminalCursorStyle(value: unknown): value is TerminalCursorStyle {
  return value === "block" || value === "underline" || value === "bar";
}

function isAgentDetailPanelId(value: unknown): value is AgentDetailPanelId {
  return (
    typeof value === "string" &&
    agentDetailPanelIds.includes(value as AgentDetailPanelId)
  );
}
