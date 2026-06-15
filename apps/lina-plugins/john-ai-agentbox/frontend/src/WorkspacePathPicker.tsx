import { useEffect, useState } from 'react';
import { ChevronDown, ChevronRight, Folder, FolderOpen, RefreshCw } from 'lucide-react';
import { api } from './api';
import type { AgentInfo, WorkspacePathSuggestion, WorkspaceTreeNode } from './types';
import { Button, ControlButton, Dialog, Input, Spinner } from '@/components/ui';
import { cn, normalizeWorkspacePath, sharedRootPath, workspaceRootPath } from '@/lib/utils';

type Props = {
  agent?: AgentInfo;
  value: string;
  onChange: (value: string) => void;
  variant?: 'bar' | 'inline';
};

export default function WorkspacePathPicker({ agent, value, onChange, variant = 'bar' }: Props) {
  const [localValue, setLocalValue] = useState(value || workspaceRootPath);
  const [suggestions, setSuggestions] = useState<WorkspacePathSuggestion[]>([]);
  const [browserOpen, setBrowserOpen] = useState(false);
  const [tree, setTree] = useState<WorkspaceTreeNode[]>([]);
  const [loading, setLoading] = useState(false);
  const [loadingPaths, setLoadingPaths] = useState<Set<string>>(new Set());
  const [expandedPaths, setExpandedPaths] = useState<Set<string>>(new Set([workspaceRootPath, sharedRootPath]));
  const [browserError, setBrowserError] = useState('');
  const [suggestionsOpen, setSuggestionsOpen] = useState(false);
  const agentReady = Boolean(agent?.id && agent.runtimeStatus === 'running');

  useEffect(() => {
    setLocalValue(value || workspaceRootPath);
  }, [value]);

  useEffect(() => {
    setSuggestions([]);
    setSuggestionsOpen(false);
    setTree([]);
    setBrowserOpen(false);
    setLoading(false);
    setLoadingPaths(new Set());
    setExpandedPaths(new Set([workspaceRootPath, sharedRootPath]));
    setBrowserError('');
  }, [agent?.id, agent?.runtimeStatus]);

  async function search(nextValue: string) {
    setLocalValue(nextValue);
    if (!agentReady || !agent?.id) {
      setSuggestions([]);
      setSuggestionsOpen(false);
      return;
    }
    try {
      const items = await api.workspacePathSuggestions(agent.id, nextValue || workspaceRootPath);
      setSuggestions(items);
      setSuggestionsOpen(items.length > 0);
    } catch {
      setSuggestions([]);
      setSuggestionsOpen(false);
    }
  }

  function commit(nextValue = localValue) {
    const normalized = normalizeWorkspacePath(nextValue);
    setLocalValue(normalized);
    setSuggestionsOpen(false);
    onChange(normalized);
  }

  async function openBrowser() {
    setBrowserOpen(true);
    setBrowserError('');
    setExpandedPaths((current) => new Set(current).add(workspaceRootPath).add(sharedRootPath));
    if (!agentReady || !agent?.id || tree.length > 0) {
      return;
    }
    setLoading(true);
    try {
      setTree(await loadRootNodes(agent.id));
    } catch {
      setTree([]);
      setBrowserError('目录树加载失败');
    } finally {
      setLoading(false);
    }
  }

  function selectPath(path: string) {
    const normalized = normalizeWorkspacePath(path);
    setLocalValue(normalized);
    onChange(normalized);
    setBrowserOpen(false);
  }

  async function toggleDirectory(node: WorkspaceTreeNode) {
    if (node.type !== 'directory') {
      return;
    }
    const expanded = expandedPaths.has(node.path);
    setExpandedPaths((current) => {
      const next = new Set(current);
      if (expanded) {
        next.delete(node.path);
      } else {
        next.add(node.path);
      }
      return next;
    });
    if (expanded || !agentReady || !agent?.id || !node.expandable || Array.isArray(node.children)) {
      return;
    }
    setLoadingPaths((current) => new Set(current).add(node.path));
    try {
      const children = directoryNodesOnly(await api.workspaceTree(agent.id, node.path, false));
      setTree((current) => updateWorkspaceTreeChildren(current, node.path, children));
    } catch {
      setBrowserError('目录树加载失败');
    } finally {
      setLoadingPaths((current) => {
        const next = new Set(current);
        next.delete(node.path);
        return next;
      });
    }
  }

  const browserTree = tree.length > 0 ? tree : workspaceRootNodes();

  return (
    <div
      className={cn(
        'flex min-h-14 items-center gap-2 border-b border-border bg-muted/50 px-3',
        variant === 'inline' && 'min-h-0 w-full min-w-0 max-w-[320px] flex-1 gap-0 overflow-hidden rounded-lg border border-border bg-background px-0 py-0 shadow-xs min-[1360px]:min-w-[300px] max-[760px]:max-w-none max-[760px]:basis-full',
      )}
      data-testid="workspace-path-picker"
    >
      <div className="relative min-w-0 flex-1">
        <Input
          aria-label="workspace 路径"
          className={cn('w-full', variant === 'inline' && 'h-8 rounded-none border-0 bg-transparent shadow-none focus-visible:ring-0')}
          disabled={!agentReady}
          placeholder={`${workspaceRootPath}/project`}
          value={localValue}
          onBlur={() => commit()}
          onChange={(event) => void search(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === 'Enter') {
              commit();
            }
          }}
        />
        {suggestionsOpen ? (
          <div className="absolute left-0 right-0 top-[calc(100%+4px)] z-20 grid max-h-56 overflow-auto rounded-[6px] border border-border bg-card p-1 shadow-xl">
            {suggestions.map((item) => (
              <ControlButton
                key={item.path}
                className="h-auto justify-start px-3 py-2 text-left"
                onMouseDown={(event) => event.preventDefault()}
                onClick={() => commit(item.path)}
              >
                {item.path}
              </ControlButton>
            ))}
          </div>
        ) : null}
      </div>
      <Button
        className={cn(variant === 'inline' && 'h-8 rounded-none border-0 border-l border-border bg-transparent shadow-none')}
        data-testid="workspace-path-browse-button"
        disabled={!agentReady}
        type="button"
        variant="ghost"
        onClick={openBrowser}
      >
        <FolderOpen data-icon="inline-start" />
        浏览
      </Button>
      <Dialog
        className="sm:max-w-3xl"
        footer={null}
        open={browserOpen}
        title="选择工作目录或共享目录"
        onClose={() => setBrowserOpen(false)}
      >
        {loading ? <Spinner label="加载目录" /> : null}
        {browserError ? <div className="rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive">{browserError}</div> : null}
        {!loading ? (
          <WorkspaceBrowserTree
            expandedPaths={expandedPaths}
            loadingPaths={loadingPaths}
            nodes={browserTree}
            selectedPath={normalizeWorkspacePath(localValue)}
            onSelectDirectory={selectPath}
            onToggleDirectory={(node) => void toggleDirectory(node)}
          />
        ) : null}
      </Dialog>
    </div>
  );
}

function WorkspaceBrowserTree({
  expandedPaths,
  loadingPaths,
  nodes,
  selectedPath,
  level = 0,
  onSelectDirectory,
  onToggleDirectory,
}: {
  expandedPaths: Set<string>;
  loadingPaths: Set<string>;
  nodes: WorkspaceTreeNode[];
  selectedPath: string;
  level?: number;
  onSelectDirectory: (path: string) => void;
  onToggleDirectory: (node: WorkspaceTreeNode) => void;
}) {
  return (
    <div className={cn('grid gap-0', level === 0 && 'rounded-md border border-border bg-card p-1')} role={level === 0 ? 'tree' : 'group'} data-testid={level === 0 ? 'workspace-browser-tree' : undefined}>
      {nodes.map((node) => {
        const expanded = expandedPaths.has(node.path);
        const loading = loadingPaths.has(node.path);
        const children = node.children ?? [];
        const active = selectedPath === node.path;
        const Icon = expanded ? FolderOpen : Folder;
        return (
          <div key={node.path}>
            <div
              aria-expanded={node.expandable ? expanded : undefined}
              className={cn(
                'grid min-h-8 grid-cols-[1.5rem_minmax(0,1fr)] items-center gap-1 rounded-[5px] text-sm',
                'hover:bg-accent',
                active && 'bg-primary/10 text-foreground hover:bg-primary/15',
              )}
              data-node-type="directory"
              data-testid={`workspace-browser-node-${node.path}`}
              role="treeitem"
              style={{ paddingLeft: 4 + level * 18 }}
            >
              {node.expandable ? (
                <ControlButton
                  aria-label={`${expanded ? '折叠' : '展开'} ${node.name}`}
                  className="h-6 w-6 shrink-0 justify-center p-0"
                  onClick={() => onToggleDirectory(node)}
                >
                  {loading ? <RefreshCw className="animate-spin text-muted-foreground" /> : expanded ? <ChevronDown /> : <ChevronRight />}
                </ControlButton>
              ) : (
                <span className="h-6 w-6" />
              )}
              <ControlButton
                aria-label={`选择 ${node.path}`}
                className="h-7 min-w-0 justify-start gap-2 px-1.5 py-1 text-left"
                onClick={() => onSelectDirectory(node.path)}
              >
                <Icon className={cn('shrink-0', expanded ? 'text-primary' : 'text-muted-foreground')} />
                <span className="min-w-0 flex-1 truncate">{node.name}</span>
                <span className="hidden min-w-0 max-w-[50%] truncate text-xs text-muted-foreground sm:block">{node.path}</span>
              </ControlButton>
            </div>
            {expanded && children.length > 0 ? (
              <WorkspaceBrowserTree
                expandedPaths={expandedPaths}
                level={level + 1}
                loadingPaths={loadingPaths}
                nodes={children}
                selectedPath={selectedPath}
                onSelectDirectory={onSelectDirectory}
                onToggleDirectory={onToggleDirectory}
              />
            ) : null}
            {expanded && Array.isArray(node.children) && children.length === 0 && !loading ? (
              <div
                className="min-h-7 rounded-[5px] px-2 py-1 text-sm text-muted-foreground"
                style={{ paddingLeft: 34 + (level + 1) * 18 }}
              >
                空目录
              </div>
            ) : null}
          </div>
        );
      })}
    </div>
  );
}

function updateWorkspaceTreeChildren(
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
        children: updateWorkspaceTreeChildren(node.children, targetPath, children),
      };
    }
    return node;
  });
}

async function loadRootNodes(agentId: string): Promise<WorkspaceTreeNode[]> {
  const [workspaceChildren, sharedChildren] = await Promise.all([
    api.workspaceTree(agentId, workspaceRootPath, false).then(directoryNodesOnly),
    api.workspaceTree(agentId, sharedRootPath, false).then(directoryNodesOnly),
  ]);
  return workspaceRootNodes(workspaceChildren, sharedChildren);
}

function workspaceRootNodes(workspaceChildren?: WorkspaceTreeNode[], sharedChildren?: WorkspaceTreeNode[]): WorkspaceTreeNode[] {
  return [
    {
      name: 'workspace',
      path: workspaceRootPath,
      type: 'directory',
      expandable: true,
      children: workspaceChildren,
    },
    {
      name: 'shared',
      path: sharedRootPath,
      type: 'directory',
      expandable: true,
      children: sharedChildren,
    },
  ];
}

function directoryNodesOnly(nodes: WorkspaceTreeNode[]): WorkspaceTreeNode[] {
  return nodes
    .filter((node) => node.type === 'directory')
    .map((node) => ({
      ...node,
      children: node.children ? directoryNodesOnly(node.children) : undefined,
    }));
}
