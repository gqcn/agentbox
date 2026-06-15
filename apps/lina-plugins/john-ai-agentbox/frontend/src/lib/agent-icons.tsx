import type { ComponentType, SVGProps } from "react";
import {
  Activity,
  Blocks,
  Bot,
  Box,
  Boxes,
  BookOpen,
  Braces,
  BrainCircuit,
  Bug,
  CircleCheckBig,
  ClipboardCheck,
  Cloud,
  CloudDownload,
  CloudUpload,
  Code2,
  Container,
  Cpu,
  Database,
  FileCode2,
  FileDiff,
  FileSearch,
  FileTerminal,
  Files,
  FlaskConical,
  FolderCode,
  FolderTree,
  Gauge,
  GitBranch,
  GitCommitHorizontal,
  GitCompare,
  GitMerge,
  GitPullRequest,
  Hammer,
  HardDrive,
  KeyRound,
  Lightbulb,
  ListChecks,
  LockKeyhole,
  Logs,
  MemoryStick,
  MessageSquareCode,
  MessagesSquare,
  MonitorCog,
  Network,
  Package,
  PlugZap,
  Regex,
  Replace,
  Rocket,
  Route,
  ScrollText,
  SearchCode,
  Server,
  ServerCog,
  Settings,
  ShieldCheck,
  Sparkles,
  TableProperties,
  TerminalSquare,
  TestTube2,
  WandSparkles,
  Waypoints,
  Webhook,
  Workflow,
  Wrench,
} from "lucide-react";

export type SelectableAgentIconKey =
  | "terminal"
  | "code"
  | "braces"
  | "file-code"
  | "file-terminal"
  | "file-search"
  | "files"
  | "folder-tree"
  | "folder-code"
  | "search-code"
  | "regex"
  | "replace"
  | "file-diff"
  | "git-branch"
  | "git-pull-request"
  | "git-commit"
  | "git-merge"
  | "git-compare"
  | "bug"
  | "test-tube"
  | "flask"
  | "clipboard-check"
  | "list-checks"
  | "circle-check"
  | "package"
  | "boxes"
  | "box"
  | "database"
  | "table-properties"
  | "server-cog"
  | "server"
  | "container"
  | "cloud"
  | "cloud-upload"
  | "cloud-download"
  | "workflow"
  | "waypoints"
  | "network"
  | "route"
  | "plug-zap"
  | "webhook"
  | "wrench"
  | "hammer"
  | "settings"
  | "shield-check"
  | "lock-keyhole"
  | "key-round"
  | "rocket"
  | "messages-square"
  | "message-square-code"
  | "brain-circuit"
  | "sparkles"
  | "lightbulb"
  | "cpu"
  | "memory-stick"
  | "hard-drive"
  | "monitor-cog"
  | "gauge"
  | "activity"
  | "logs"
  | "scroll-text"
  | "book-open"
  | "wand-sparkles"
  | "blocks";

export type AgentIconKey = "bot" | SelectableAgentIconKey;

export type AgentIconComponent = ComponentType<SVGProps<SVGSVGElement>>;

export type AgentIconCategoryId =
  | "development"
  | "search"
  | "collaboration"
  | "quality"
  | "delivery"
  | "runtime"
  | "integration-security"
  | "data-docs";

export type AgentIconCategory = {
  id: AgentIconCategoryId;
  label: string;
};

export const agentIconCategories: AgentIconCategory[] = [
  { id: "development", label: "开发实现" },
  { id: "search", label: "检索重构" },
  { id: "collaboration", label: "代码协作" },
  { id: "quality", label: "测试质量" },
  { id: "delivery", label: "构建发布" },
  { id: "runtime", label: "运行环境" },
  { id: "integration-security", label: "集成安全" },
  { id: "data-docs", label: "数据文档" },
];

export type AgentIconOption = {
  key: SelectableAgentIconKey;
  label: string;
  Icon: AgentIconComponent;
  category: AgentIconCategoryId;
  recommended?: boolean;
};

export const agentIconOptions: AgentIconOption[] = [
  {
    key: "terminal",
    label: "终端执行",
    Icon: TerminalSquare,
    category: "development",
    recommended: true,
  },
  {
    key: "code",
    label: "功能开发",
    Icon: Code2,
    category: "development",
    recommended: true,
  },
  {
    key: "braces",
    label: "结构化代码",
    Icon: Braces,
    category: "development",
  },
  {
    key: "file-code",
    label: "代码文件",
    Icon: FileCode2,
    category: "development",
  },
  {
    key: "file-terminal",
    label: "脚本自动化",
    Icon: FileTerminal,
    category: "development",
  },
  {
    key: "file-search",
    label: "文件定位",
    Icon: FileSearch,
    category: "search",
  },
  {
    key: "files",
    label: "多文件修改",
    Icon: Files,
    category: "development",
    recommended: true,
  },
  {
    key: "folder-tree",
    label: "项目结构",
    Icon: FolderTree,
    category: "development",
  },
  {
    key: "folder-code",
    label: "代码仓库",
    Icon: FolderCode,
    category: "development",
  },
  {
    key: "search-code",
    label: "代码检索",
    Icon: SearchCode,
    category: "search",
    recommended: true,
  },
  {
    key: "regex",
    label: "规则检索",
    Icon: Regex,
    category: "search",
  },
  {
    key: "replace",
    label: "批量重构",
    Icon: Replace,
    category: "search",
  },
  {
    key: "file-diff",
    label: "差异审查",
    Icon: FileDiff,
    category: "collaboration",
    recommended: true,
  },
  {
    key: "git-branch",
    label: "分支管理",
    Icon: GitBranch,
    category: "collaboration",
  },
  {
    key: "git-pull-request",
    label: "PR 审查",
    Icon: GitPullRequest,
    category: "collaboration",
    recommended: true,
  },
  {
    key: "git-commit",
    label: "提交整理",
    Icon: GitCommitHorizontal,
    category: "collaboration",
  },
  {
    key: "git-merge",
    label: "合并变更",
    Icon: GitMerge,
    category: "collaboration",
  },
  {
    key: "git-compare",
    label: "版本对比",
    Icon: GitCompare,
    category: "collaboration",
  },
  {
    key: "bug",
    label: "缺陷修复",
    Icon: Bug,
    category: "quality",
    recommended: true,
  },
  {
    key: "test-tube",
    label: "测试验证",
    Icon: TestTube2,
    category: "quality",
    recommended: true,
  },
  {
    key: "flask",
    label: "实验验证",
    Icon: FlaskConical,
    category: "quality",
  },
  {
    key: "clipboard-check",
    label: "验收检查",
    Icon: ClipboardCheck,
    category: "quality",
  },
  {
    key: "list-checks",
    label: "任务拆解",
    Icon: ListChecks,
    category: "development",
  },
  {
    key: "circle-check",
    label: "完成校验",
    Icon: CircleCheckBig,
    category: "quality",
  },
  {
    key: "package",
    label: "依赖构建",
    Icon: Package,
    category: "delivery",
    recommended: true,
  },
  {
    key: "boxes",
    label: "模块装配",
    Icon: Boxes,
    category: "delivery",
  },
  {
    key: "box",
    label: "构建产物",
    Icon: Box,
    category: "delivery",
  },
  {
    key: "database",
    label: "数据访问",
    Icon: Database,
    category: "data-docs",
    recommended: true,
  },
  {
    key: "table-properties",
    label: "数据表建模",
    Icon: TableProperties,
    category: "data-docs",
  },
  {
    key: "server-cog",
    label: "服务运维",
    Icon: ServerCog,
    category: "runtime",
    recommended: true,
  },
  {
    key: "server",
    label: "后端服务",
    Icon: Server,
    category: "runtime",
  },
  {
    key: "container",
    label: "容器环境",
    Icon: Container,
    category: "runtime",
    recommended: true,
  },
  {
    key: "cloud",
    label: "云端执行",
    Icon: Cloud,
    category: "runtime",
  },
  {
    key: "cloud-upload",
    label: "交付上传",
    Icon: CloudUpload,
    category: "delivery",
  },
  {
    key: "cloud-download",
    label: "环境同步",
    Icon: CloudDownload,
    category: "runtime",
  },
  {
    key: "workflow",
    label: "流程编排",
    Icon: Workflow,
    category: "development",
    recommended: true,
  },
  {
    key: "waypoints",
    label: "任务路径",
    Icon: Waypoints,
    category: "development",
  },
  {
    key: "network",
    label: "服务拓扑",
    Icon: Network,
    category: "runtime",
  },
  {
    key: "route",
    label: "路由调试",
    Icon: Route,
    category: "runtime",
  },
  {
    key: "plug-zap",
    label: "工具接入",
    Icon: PlugZap,
    category: "integration-security",
    recommended: true,
  },
  {
    key: "webhook",
    label: "事件回调",
    Icon: Webhook,
    category: "integration-security",
  },
  {
    key: "wrench",
    label: "工具配置",
    Icon: Wrench,
    category: "integration-security",
  },
  {
    key: "hammer",
    label: "构建工具",
    Icon: Hammer,
    category: "delivery",
  },
  {
    key: "settings",
    label: "运行配置",
    Icon: Settings,
    category: "integration-security",
  },
  {
    key: "shield-check",
    label: "安全审查",
    Icon: ShieldCheck,
    category: "integration-security",
    recommended: true,
  },
  {
    key: "lock-keyhole",
    label: "权限控制",
    Icon: LockKeyhole,
    category: "integration-security",
  },
  {
    key: "key-round",
    label: "凭据管理",
    Icon: KeyRound,
    category: "integration-security",
  },
  {
    key: "rocket",
    label: "发布部署",
    Icon: Rocket,
    category: "delivery",
    recommended: true,
  },
  {
    key: "messages-square",
    label: "协作沟通",
    Icon: MessagesSquare,
    category: "collaboration",
    recommended: true,
  },
  {
    key: "message-square-code",
    label: "代码答疑",
    Icon: MessageSquareCode,
    category: "collaboration",
  },
  {
    key: "brain-circuit",
    label: "架构规划",
    Icon: BrainCircuit,
    category: "development",
    recommended: true,
  },
  {
    key: "sparkles",
    label: "智能生成",
    Icon: Sparkles,
    category: "development",
  },
  {
    key: "lightbulb",
    label: "方案建议",
    Icon: Lightbulb,
    category: "development",
  },
  {
    key: "cpu",
    label: "算力评估",
    Icon: Cpu,
    category: "runtime",
  },
  {
    key: "memory-stick",
    label: "内存诊断",
    Icon: MemoryStick,
    category: "runtime",
  },
  {
    key: "hard-drive",
    label: "存储排查",
    Icon: HardDrive,
    category: "runtime",
  },
  {
    key: "monitor-cog",
    label: "工作台配置",
    Icon: MonitorCog,
    category: "runtime",
  },
  {
    key: "gauge",
    label: "性能优化",
    Icon: Gauge,
    category: "runtime",
  },
  {
    key: "activity",
    label: "运行观测",
    Icon: Activity,
    category: "runtime",
  },
  {
    key: "logs",
    label: "日志分析",
    Icon: Logs,
    category: "runtime",
    recommended: true,
  },
  {
    key: "scroll-text",
    label: "技术文档",
    Icon: ScrollText,
    category: "data-docs",
  },
  {
    key: "book-open",
    label: "知识阅读",
    Icon: BookOpen,
    category: "data-docs",
    recommended: true,
  },
  {
    key: "wand-sparkles",
    label: "文档润色",
    Icon: WandSparkles,
    category: "data-docs",
  },
  {
    key: "blocks",
    label: "插件扩展",
    Icon: Blocks,
    category: "integration-security",
  },
];

export const recommendedAgentIconOptions = agentIconOptions.filter(
  (option) => option.recommended,
);

const agentIconOptionByKey = new Map(
  agentIconOptions.map((option) => [option.key, option]),
);

const defaultIconByAgentType: Record<string, AgentIconKey> = {
  claude_code: "bot",
  codex: "terminal",
  custom: "sparkles",
};

export function normalizeAgentIconKey(
  value: string | undefined,
): SelectableAgentIconKey | "" {
  return agentIconOptionByKey.has(value as SelectableAgentIconKey)
    ? (value as SelectableAgentIconKey)
    : "";
}

export function resolveAgentIconKey(
  iconKey: string | undefined,
  agentType: string | undefined,
): AgentIconKey {
  return (
    normalizeAgentIconKey(iconKey) ||
    defaultIconByAgentType[agentType ?? ""] ||
    "bot"
  );
}

export function resolveAgentIcon(
  iconKey: string | undefined,
  agentType: string | undefined,
): AgentIconComponent {
  const resolvedIconKey = resolveAgentIconKey(iconKey, agentType);
  return (
    agentIconOptionByKey.get(resolvedIconKey as SelectableAgentIconKey)?.Icon ??
    Bot
  );
}
