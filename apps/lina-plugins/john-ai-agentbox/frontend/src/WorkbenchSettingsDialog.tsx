import {
  ArrowDown,
  ArrowUp,
  Boxes,
  Code2,
  Columns2,
  FileText,
  GitCompareArrows,
  GripVertical,
  MessageSquare,
  Network,
  PanelLeft,
  RotateCcw,
  Terminal,
} from "lucide-react";
import { useRef, useState } from "react";
import type { PointerEvent as ReactPointerEvent, ReactNode } from "react";
import { createPortal } from "react-dom";
import { toast } from "sonner";
import {
  Button,
  CheckboxField,
  Dialog,
  Field,
  Input,
  Select,
  SelectOption,
} from "@/components/ui";
import { useFullscreenPortalContainer } from "@/hooks/useFullscreenPortalContainer";
import type {
  AgentDetailPanelId,
  AgentDetailPanelPreference,
  WorkbenchSettings,
} from "@/lib/workbench-settings";
import {
  defaultWorkbenchSettings,
  normalizeWorkbenchSettings,
} from "@/lib/workbench-settings";
import { cn } from "@/lib/utils";

type SettingsCategory =
  | "workbench"
  | "editor"
  | "source-control"
  | "terminal"
  | "code-display";
type AgentDetailPanelDropPosition = "before" | "after";
type AgentDetailPanelDropIndicator = {
  id: AgentDetailPanelId;
  position: AgentDetailPanelDropPosition;
};
type AgentDetailPanelDragPreview = {
  id: AgentDetailPanelId;
  label: string;
  x: number;
  y: number;
};
type AgentDetailPanelDragGesture = {
  id: AgentDetailPanelId;
  active: boolean;
  startX: number;
  startY: number;
};

type Props = {
  open: boolean;
  settings: WorkbenchSettings;
  onClose: () => void;
  onReset: () => void;
  onSettingsChange: (settings: WorkbenchSettings) => void;
};

const categories: Array<{
  id: SettingsCategory;
  label: string;
  icon: typeof PanelLeft;
}> = [
  { id: "workbench", label: "工作台", icon: PanelLeft },
  { id: "editor", label: "编辑器", icon: Code2 },
  { id: "source-control", label: "源代码管理", icon: GitCompareArrows },
  { id: "terminal", label: "终端", icon: Terminal },
  { id: "code-display", label: "代码显示", icon: Columns2 },
];

const panelSettingMeta: Record<
  AgentDetailPanelId,
  { label: string; icon: typeof PanelLeft }
> = {
  chat: { label: "对话", icon: MessageSquare },
  shell: { label: "终端", icon: Terminal },
  services: { label: "服务", icon: Network },
  skills: { label: "技能", icon: Boxes },
  files: { label: "文件", icon: FileText },
  git: { label: "变更", icon: GitCompareArrows },
};

export default function WorkbenchSettingsDialog({
  open,
  settings,
  onClose,
  onReset,
  onSettingsChange,
}: Props) {
  function patchSettings(patch: Partial<WorkbenchSettings>) {
    onSettingsChange(normalizeWorkbenchSettings({ ...settings, ...patch }));
  }

  function updateAgentDetailPanels(
    agentDetailPanels: AgentDetailPanelPreference[],
  ) {
    patchSettings({ agentDetailPanels });
  }

  return (
    <Dialog
      className="sm:max-w-4xl"
      footer={
        <>
          <Button type="button" variant="soft" onClick={onReset}>
            <RotateCcw />
            恢复默认
          </Button>
          <Button type="button" variant="soft" onClick={onClose}>
            取消
          </Button>
          <Button type="button" onClick={onClose}>
            完成
          </Button>
        </>
      }
      open={open}
      title="工作台设置"
      onClose={onClose}
    >
      <div
        className="grid h-[520px] max-h-[calc(90dvh_-_10rem)] min-h-0 grid-cols-[190px_minmax(0,1fr)] gap-0 overflow-hidden rounded-[6px] border max-[760px]:grid-cols-1 max-[760px]:grid-rows-[auto_minmax(0,1fr)]"
        data-testid="workbench-settings-dialog"
      >
        <nav
          className="flex min-h-0 flex-col gap-1 border-r bg-muted/40 p-2 max-[760px]:border-b max-[760px]:border-r-0"
          data-testid="workbench-settings-categories"
        >
          {categories.map((category) => {
            const Icon = category.icon;
            return (
              <a
                key={category.id}
                className="flex h-9 items-center gap-2 rounded-[4px] px-2 text-sm text-muted-foreground hover:bg-accent hover:text-foreground"
                href={`#workbench-settings-${category.id}`}
              >
                <Icon />
                <span className="truncate">{category.label}</span>
              </a>
            );
          })}
        </nav>
        <div
          className="min-h-0 overflow-auto bg-background"
          data-testid="workbench-settings-content"
        >
          <SettingsSection id="workbench-settings-workbench" title="工作台">
            <NumberSetting
              description="控制文件页面首次打开时资源管理器的默认比例。"
              label="文件：资源管理器默认宽度"
              step="any"
              suffix="%"
              value={settings.filesSidebarSize}
              onChange={(filesSidebarSize) =>
                patchSettings({ filesSidebarSize })
              }
            />
            <NumberSetting
              description="控制 Git 页面首次打开时源代码管理面板的默认比例。"
              label="Git：源代码管理默认宽度"
              step="any"
              suffix="%"
              value={settings.gitSidebarSize}
              onChange={(gitSidebarSize) => patchSettings({ gitSidebarSize })}
            />
            <AgentDetailPanelsSetting
              panels={settings.agentDetailPanels}
              onChange={updateAgentDetailPanels}
            />
          </SettingsSection>

          <SettingsSection id="workbench-settings-editor" title="编辑器">
            <NumberSetting
              description="应用到文件编辑器、Git 文件编辑器和 Git 差异编辑器。"
              label="字号"
              max={20}
              min={11}
              suffix="px"
              value={settings.editorFontSize}
              onChange={(editorFontSize) => patchSettings({ editorFontSize })}
            />
            <NumberSetting
              description="应用到可编辑文件和左右对照差异视图。"
              label="制表符大小"
              max={8}
              min={2}
              suffix="空格"
              value={settings.editorTabSize}
              onChange={(editorTabSize) => patchSettings({ editorTabSize })}
            />
            <SelectSetting
              description="控制长行是否在编辑器视口内自动换行。"
              label="自动换行"
              value={settings.editorWordWrap}
              onChange={(editorWordWrap) => patchSettings({ editorWordWrap })}
              options={[
                { label: "开启", value: "on" },
                { label: "关闭", value: "off" },
              ]}
            />
            <CheckboxSetting
              checked={settings.editorMinimap}
              description="在编辑器右侧显示代码缩略图。"
              label="缩略图"
              onChange={(editorMinimap) => patchSettings({ editorMinimap })}
            />
          </SettingsSection>

          <SettingsSection
            id="workbench-settings-source-control"
            title="源代码管理"
          >
            <SelectSetting
              description="桌面视口默认使用左右对照，可按个人习惯切换内联。"
              label="差异展示方式"
              value={settings.gitDiffDisplay}
              onChange={(gitDiffDisplay) => patchSettings({ gitDiffDisplay })}
              options={[
                { label: "左右对照", value: "side-by-side" },
                { label: "内联", value: "inline" },
              ]}
            />
            <CheckboxSetting
              checked={settings.gitDiffInlineOnNarrow}
              description="窄屏下自动使用内联差异视图，避免左右内容互相挤压。"
              label="窄屏自动内联差异"
              onChange={(gitDiffInlineOnNarrow) =>
                patchSettings({ gitDiffInlineOnNarrow })
              }
            />
          </SettingsSection>

          <SettingsSection id="workbench-settings-terminal" title="终端">
            <NumberSetting
              description="应用到终端内容字号，已打开的终端会重新适配当前 pane 尺寸。"
              label="终端字号"
              max={24}
              min={10}
              suffix="px"
              value={settings.terminalFontSize}
              onChange={(terminalFontSize) =>
                patchSettings({ terminalFontSize })
              }
            />
            <NumberSetting
              description="控制终端行距密度，适合在紧凑和可读之间切换。"
              label="终端行高"
              max={2}
              min={1}
              step={0.05}
              suffix="倍"
              value={settings.terminalLineHeight}
              onChange={(terminalLineHeight) =>
                patchSettings({ terminalLineHeight })
              }
            />
            <SelectSetting
              description="设置终端光标展示样式，与 VSCode 常用终端配置保持一致。"
              label="终端光标样式"
              value={settings.terminalCursorStyle}
              onChange={(terminalCursorStyle) =>
                patchSettings({ terminalCursorStyle })
              }
              options={[
                { label: "块状", value: "block" },
                { label: "下划线", value: "underline" },
                { label: "竖线", value: "bar" },
              ]}
            />
            <CheckboxSetting
              checked={settings.terminalCursorBlink}
              description="开启后终端光标会闪烁，关闭后保持常亮。"
              label="终端光标闪烁"
              onChange={(terminalCursorBlink) =>
                patchSettings({ terminalCursorBlink })
              }
            />
            <NumberSetting
              description="控制竖线光标的宽度，块状和下划线光标会保留 xterm 默认表现。"
              label="终端光标宽度"
              max={6}
              min={1}
              suffix="px"
              value={settings.terminalCursorWidth}
              onChange={(terminalCursorWidth) =>
                patchSettings({ terminalCursorWidth })
              }
            />
            <SelectSetting
              description="菜单模式展示终端右键菜单；直接粘贴模式会把剪贴板文本发送到当前 pane。"
              label="终端右键行为"
              value={settings.terminalRightClickBehavior}
              onChange={(terminalRightClickBehavior) =>
                patchSettings({ terminalRightClickBehavior })
              }
              options={[
                { label: "显示菜单", value: "menu" },
                { label: "直接粘贴", value: "paste" },
              ]}
            />
          </SettingsSection>

          <SettingsSection
            id="workbench-settings-code-display"
            title="代码显示"
          >
            <CheckboxSetting
              checked={settings.codeHighlighting}
              description="关闭后所有文件和差异内容按纯文本渲染。"
              label="代码高亮"
              onChange={(codeHighlighting) =>
                patchSettings({ codeHighlighting })
              }
            />
            <div
              className="grid gap-1 rounded-[4px] border bg-muted/30 p-3 text-xs text-muted-foreground"
              data-testid="workbench-settings-defaults"
            >
              <span>
                默认值：文件 {defaultWorkbenchSettings.filesSidebarSize}% · Git{" "}
                {defaultWorkbenchSettings.gitSidebarSize}% · 编辑器字号{" "}
                {defaultWorkbenchSettings.editorFontSize}px · 终端字号{" "}
                {defaultWorkbenchSettings.terminalFontSize}px
              </span>
            </div>
          </SettingsSection>
        </div>
      </div>
    </Dialog>
  );
}

function AgentDetailPanelsSetting({
  panels,
  onChange,
}: {
  panels: AgentDetailPanelPreference[];
  onChange: (panels: AgentDetailPanelPreference[]) => void;
}) {
  const [draggingPanelId, setDraggingPanelId] =
    useState<AgentDetailPanelId | null>(null);
  const [dropIndicator, setDropIndicator] =
    useState<AgentDetailPanelDropIndicator | null>(null);
  const [dragPreview, setDragPreview] =
    useState<AgentDetailPanelDragPreview | null>(null);
  const dragGestureRef = useRef<AgentDetailPanelDragGesture | null>(null);
  const rowRefs = useRef(new Map<AgentDetailPanelId, HTMLDivElement>());
  const fullscreenPortalContainer = useFullscreenPortalContainer();
  const dragPreviewPortalContainer =
    fullscreenPortalContainer ??
    (typeof document !== "undefined" ? document.body : null);
  const visibleCount = panels.filter((panel) => panel.visible).length;
  const DragPreviewIcon = dragPreview
    ? panelSettingMeta[dragPreview.id].icon
    : null;

  function movePanel(index: number, direction: -1 | 1) {
    const nextIndex = index + direction;
    if (nextIndex < 0 || nextIndex >= panels.length) {
      return;
    }
    const next = [...panels];
    [next[index], next[nextIndex]] = [next[nextIndex], next[index]];
    onChange(next);
  }

  function reorderPanel(
    sourceId: AgentDetailPanelId,
    targetId: AgentDetailPanelId,
    position: AgentDetailPanelDropPosition,
  ) {
    if (sourceId === targetId) {
      return;
    }
    const sourceIndex = panels.findIndex((panel) => panel.id === sourceId);
    if (sourceIndex < 0) {
      return;
    }
    const moved = panels[sourceIndex];
    const next = panels.filter((panel) => panel.id !== sourceId);
    const targetIndex = next.findIndex((panel) => panel.id === targetId);
    if (targetIndex < 0) {
      return;
    }
    const insertIndex = position === "after" ? targetIndex + 1 : targetIndex;
    next.splice(insertIndex, 0, moved);
    onChange(next);
  }

  function dropIndicatorFromPoint(
    sourceId: AgentDetailPanelId,
    clientX: number,
    clientY: number,
  ): AgentDetailPanelDropIndicator | null {
    for (const [id, row] of rowRefs.current) {
      if (id === sourceId) {
        continue;
      }
      const bounds = row.getBoundingClientRect();
      if (
        clientX < bounds.left ||
        clientX > bounds.right ||
        clientY < bounds.top ||
        clientY > bounds.bottom
      ) {
        continue;
      }
      return {
        id,
        position:
          clientY >= bounds.top + bounds.height / 2 ? "after" : "before",
      };
    }
    return null;
  }

  function setRowRef(id: AgentDetailPanelId, row: HTMLDivElement | null) {
    if (row) {
      rowRefs.current.set(id, row);
      return;
    }
    rowRefs.current.delete(id);
  }

  function canStartRowDrag(target: EventTarget | null) {
    if (!(target instanceof Element)) {
      return false;
    }
    if (target.closest('[data-panel-drag-handle="true"]')) {
      return true;
    }
    return !target.closest(
      'button,input,label,[role="checkbox"],[data-panel-row-action="true"]',
    );
  }

  function clearDragState() {
    dragGestureRef.current = null;
    setDraggingPanelId(null);
    setDropIndicator(null);
    setDragPreview(null);
  }

  function handlePointerDown(
    event: ReactPointerEvent<HTMLDivElement>,
    id: AgentDetailPanelId,
  ) {
    if (event.button !== 0 || !canStartRowDrag(event.target)) {
      return;
    }
    event.preventDefault();
    dragGestureRef.current = {
      id,
      active: false,
      startX: event.clientX,
      startY: event.clientY,
    };
    event.currentTarget.setPointerCapture(event.pointerId);
  }

  function handlePointerMove(event: ReactPointerEvent<HTMLDivElement>) {
    const gesture = dragGestureRef.current;
    if (!gesture) {
      return;
    }
    const moved = Math.hypot(
      event.clientX - gesture.startX,
      event.clientY - gesture.startY,
    );
    if (!gesture.active && moved < 4) {
      return;
    }
    if (!gesture.active) {
      gesture.active = true;
      setDraggingPanelId(gesture.id);
      setDragPreview({
        id: gesture.id,
        label: panelSettingMeta[gesture.id].label,
        x: event.clientX,
        y: event.clientY,
      });
    } else {
      setDragPreview((preview) =>
        preview ? { ...preview, x: event.clientX, y: event.clientY } : preview,
      );
    }
    setDropIndicator(
      dropIndicatorFromPoint(gesture.id, event.clientX, event.clientY),
    );
  }

  function handlePointerUp(event: ReactPointerEvent<HTMLDivElement>) {
    const gesture = dragGestureRef.current;
    if (!gesture) {
      return;
    }
    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId);
    }
    const indicator =
      dropIndicatorFromPoint(gesture.id, event.clientX, event.clientY) ??
      dropIndicator;
    if (gesture.active && indicator) {
      reorderPanel(gesture.id, indicator.id, indicator.position);
    }
    clearDragState();
  }

  function handlePointerCancel(event: ReactPointerEvent<HTMLDivElement>) {
    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId);
    }
    clearDragState();
  }

  function setPanelVisible(id: AgentDetailPanelId, visible: boolean) {
    if (!visible && visibleCount <= 1) {
      toast.warning("至少保留一个功能入口。");
      return;
    }
    onChange(
      panels.map((panel) => (panel.id === id ? { ...panel, visible } : panel)),
    );
  }

  return (
    <div
      className="grid gap-2 rounded-[4px] px-2 py-2 hover:bg-muted/50"
      data-testid="agent-detail-panel-preferences"
    >
      <div className="min-w-0">
        <div className="text-sm font-medium">功能展示</div>
        <p className="mt-1 text-xs leading-5 text-muted-foreground">
          调整 Agent 详情页顶部功能入口的顺序和可见性，对所有智能体生效。
        </p>
      </div>
      <div className="grid gap-1 rounded-[4px] border bg-muted/20 p-1">
        {panels.map((panel, index) => {
          const meta = panelSettingMeta[panel.id];
          const Icon = meta.icon;
          const currentDropIndicator =
            dropIndicator?.id === panel.id && draggingPanelId !== panel.id
              ? dropIndicator
              : null;
          return (
            <div
              className={cn(
                "relative grid cursor-grab touch-none select-none grid-cols-[auto_minmax(0,1fr)_auto] items-center gap-2 rounded-[4px] border border-transparent bg-background px-2 py-1.5 shadow-sm transition-[background-color,border-color,box-shadow,opacity] max-[760px]:grid-cols-[minmax(0,1fr)_auto]",
                draggingPanelId === panel.id &&
                  "cursor-grabbing border-dashed border-muted-foreground/50 bg-muted/30 opacity-70 shadow-none",
                currentDropIndicator && "border-primary/40 bg-primary/5",
              )}
              data-drop-position={currentDropIndicator?.position}
              data-testid="agent-detail-panel-setting-row"
              data-panel-id={panel.id}
              key={panel.id}
              ref={(row) => setRowRef(panel.id, row)}
              onPointerCancel={handlePointerCancel}
              onPointerDown={(event) => handlePointerDown(event, panel.id)}
              onPointerMove={handlePointerMove}
              onPointerUp={handlePointerUp}
            >
              {currentDropIndicator ? (
                <span
                  aria-hidden="true"
                  className={cn(
                    "pointer-events-none absolute left-2 right-2 z-10 h-0 border-t-2 border-primary",
                    currentDropIndicator.position === "before"
                      ? "-top-[3px]"
                      : "-bottom-[3px]",
                  )}
                  data-testid="agent-detail-panel-drop-indicator"
                />
              ) : null}
              <div className="flex min-w-0 items-center gap-2 max-[760px]:col-span-2">
                <Button
                  aria-label={`拖动${meta.label}排序`}
                  className="inline-flex h-6 w-6 cursor-grab items-center justify-center rounded-[4px] text-muted-foreground outline-none hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring active:cursor-grabbing"
                  data-panel-drag-handle="true"
                  data-testid={`agent-detail-panel-drag-handle-${panel.id}`}
                  size="icon-xs"
                  type="button"
                  variant="ghost"
                >
                  <GripVertical className="h-4 w-4" aria-hidden="true" />
                </Button>
                <Icon
                  className="h-4 w-4 text-muted-foreground"
                  aria-hidden="true"
                />
                <span className="truncate text-sm font-medium">
                  {meta.label}
                </span>
              </div>
              <div
                className="flex items-center gap-1 justify-self-end max-[760px]:justify-self-start"
                data-panel-row-action="true"
              >
                <Button
                  aria-label={`上移${meta.label}`}
                  disabled={index === 0}
                  size="icon-xs"
                  type="button"
                  variant="ghost"
                  onClick={() => movePanel(index, -1)}
                >
                  <ArrowUp />
                </Button>
                <Button
                  aria-label={`下移${meta.label}`}
                  disabled={index === panels.length - 1}
                  size="icon-xs"
                  type="button"
                  variant="ghost"
                  onClick={() => movePanel(index, 1)}
                >
                  <ArrowDown />
                </Button>
              </div>
              <CheckboxField
                aria-label={`显示${meta.label}`}
                checked={panel.visible}
                className="justify-self-end"
                data-testid={`agent-detail-panel-visible-${panel.id}`}
                onChange={(event) =>
                  setPanelVisible(panel.id, event.currentTarget.checked)
                }
              >
                显示{meta.label}
              </CheckboxField>
            </div>
          );
        })}
      </div>
      {dragPreview && DragPreviewIcon && dragPreviewPortalContainer
        ? createPortal(
            <div
              className="pointer-events-none fixed z-[10000] grid max-w-52 grid-cols-[auto_auto_minmax(0,1fr)] items-center gap-2 rounded-[4px] border border-dashed border-primary/70 bg-background/95 px-3 py-2 text-sm font-medium text-foreground shadow-lg backdrop-blur"
              data-testid="agent-detail-panel-drag-preview"
              style={{
                left: dragPreview.x,
                top: dragPreview.y,
                transform: "translate3d(8px, -50%, 0)",
              }}
            >
              <GripVertical
                className="h-4 w-4 text-muted-foreground"
                aria-hidden="true"
              />
              <DragPreviewIcon
                className="h-4 w-4 text-muted-foreground"
                aria-hidden="true"
              />
              <span className="truncate">{dragPreview.label}</span>
            </div>,
            dragPreviewPortalContainer,
          )
        : null}
    </div>
  );
}

function SettingsSection({
  children,
  id,
  title,
}: {
  children: ReactNode;
  id: string;
  title: string;
}) {
  return (
    <section className="grid gap-3 border-b p-4 last:border-b-0" id={id}>
      <h2 className="text-sm font-semibold">{title}</h2>
      <div className="grid gap-2">{children}</div>
    </section>
  );
}

function NumberSetting({
  description,
  label,
  max,
  min,
  suffix,
  step = 1,
  value,
  onChange,
}: {
  description: string;
  label: string;
  max?: number;
  min?: number;
  step?: number | "any";
  suffix: string;
  value: number;
  onChange: (value: number) => void;
}) {
  return (
    <div className="grid grid-cols-[minmax(0,1fr)_148px] items-center gap-3 rounded-[4px] px-2 py-2 hover:bg-muted/50 max-[760px]:grid-cols-1">
      <div className="min-w-0">
        <div className="text-sm font-medium">{label}</div>
        <p className="mt-1 text-xs leading-5 text-muted-foreground">
          {description}
        </p>
      </div>
      <Field className="min-w-0" label={suffix}>
        <Input
          aria-label={label}
          data-testid={`workbench-setting-${settingKey(label)}`}
          max={max}
          min={min}
          step={step}
          type="number"
          value={value}
          onChange={(event) => onChange(Number(event.currentTarget.value))}
        />
      </Field>
    </div>
  );
}

function SelectSetting<T extends string>({
  description,
  label,
  options,
  value,
  onChange,
}: {
  description: string;
  label: string;
  options: Array<{ label: string; value: T }>;
  value: T;
  onChange: (value: T) => void;
}) {
  return (
    <div className="grid grid-cols-[minmax(0,1fr)_190px] items-center gap-3 rounded-[4px] px-2 py-2 hover:bg-muted/50 max-[760px]:grid-cols-1">
      <div className="min-w-0">
        <div className="text-sm font-medium">{label}</div>
        <p className="mt-1 text-xs leading-5 text-muted-foreground">
          {description}
        </p>
      </div>
      <Select
        aria-label={label}
        value={value}
        onChange={(event) => onChange(event.currentTarget.value as T)}
      >
        {options.map((option) => (
          <SelectOption key={option.value} value={option.value}>
            {option.label}
          </SelectOption>
        ))}
      </Select>
    </div>
  );
}

function CheckboxSetting({
  checked,
  description,
  label,
  onChange,
}: {
  checked: boolean;
  description: string;
  label: string;
  onChange: (checked: boolean) => void;
}) {
  return (
    <div
      className={cn(
        "grid grid-cols-[minmax(0,1fr)_auto] items-center gap-3 rounded-[4px] px-2 py-2 hover:bg-muted/50 max-[760px]:grid-cols-1",
      )}
    >
      <div className="min-w-0">
        <div className="text-sm font-medium">{label}</div>
        <p className="mt-1 text-xs leading-5 text-muted-foreground">
          {description}
        </p>
      </div>
      <CheckboxField
        aria-label={label}
        checked={checked}
        data-testid={`workbench-setting-${settingKey(label)}`}
        onChange={(event) => onChange(event.currentTarget.checked)}
      >
        启用 {label}
      </CheckboxField>
    </div>
  );
}

function settingKey(label: string) {
  return encodeURIComponent(label);
}
