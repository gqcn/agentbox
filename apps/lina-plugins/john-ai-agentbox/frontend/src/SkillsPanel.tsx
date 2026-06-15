import type { ReactNode } from 'react';
import { useEffect, useMemo, useState } from 'react';
import { Inbox, PowerOff, RefreshCw, SearchX, Upload } from 'lucide-react';
import { toast } from 'sonner';
import { api } from './api';
import type { AgentInfo, WorkspaceSkillInfo } from './types';
import type { ColumnDef } from '@/components/ui';
import { Badge, Button, DataTable, Dialog, EmptyState, FileUploadButton, Pagination, SearchField, Spinner, TabButton, Toolbar } from '@/components/ui';
import { cn } from '@/lib/utils';

type Props = {
  active: boolean;
  agent?: AgentInfo;
  workspacePath: string;
};

type SkillFilter = 'all' | 'global' | 'project';

const PAGE_SIZE = 10;

export default function SkillsPanel({ active, agent, workspacePath }: Props) {
  const [filter, setFilter] = useState<SkillFilter>('all');
  const [query, setQuery] = useState('');
  const [allItems, setAllItems] = useState<WorkspaceSkillInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [detailSkill, setDetailSkill] = useState<WorkspaceSkillInfo | null>(null);
  const agentReady = Boolean(agent?.id && agent.runtimeStatus === 'running');

  const filteredSkills = useMemo(() => {
    let items = allItems;
    const normalized = query.trim().toLowerCase();
    if (normalized) {
      items = items.filter((skill) =>
        [skill.name, skill.description, skill.path].some((value) => value?.toLowerCase().includes(normalized)),
      );
    }
    if (filter !== 'all') {
      items = items.filter((skill) => skill.scope === filter);
    }
    return items;
  }, [allItems, query, filter]);

  const totalPages = Math.max(1, Math.ceil(filteredSkills.length / PAGE_SIZE));
  const currentPage = Math.min(page, totalPages);
  const pageItems = useMemo(() => {
    const start = (currentPage - 1) * PAGE_SIZE;
    return filteredSkills.slice(start, start + PAGE_SIZE);
  }, [filteredSkills, currentPage]);

  useEffect(() => {
    setPage(1);
  }, [query, filter]);

  useEffect(() => {
    if (active) {
      void loadSkills();
    }
  }, [active, agent?.id, agent?.runtimeStatus, workspacePath]);

  async function loadSkills({ notify = false }: { notify?: boolean } = {}) {
    if (!agentReady || !agent?.id) {
      resetPanelState();
      return;
    }
    setLoading(true);
    try {
      const [globalRes, projectRes] = await Promise.allSettled([
        api.listSkills(agent.id, 'global', workspacePath, ''),
        api.listSkills(agent.id, 'project', workspacePath, ''),
      ]);
      const items: WorkspaceSkillInfo[] = [];
      if (globalRes.status === 'fulfilled') {
        items.push(...(globalRes.value.items ?? []));
      }
      if (projectRes.status === 'fulfilled') {
        items.push(...(projectRes.value.items ?? []));
      }
      setAllItems(items);
      if (notify) {
        toast.success('技能列表已刷新');
      }
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
    } finally {
      setLoading(false);
    }
  }

  async function uploadProjectSkills(files: FileList | null) {
    if (!agentReady || !agent?.id || !files?.length) {
      return;
    }
    setLoading(true);
    try {
      await api.uploadProjectSkills(agent.id, workspacePath, Array.from(files));
      await loadSkills();
      toast.success('技能已上传');
    } catch (err) {
      const message = (err as Error).message;
      toast.error(message);
    } finally {
      setLoading(false);
    }
  }

  const tabs: { key: SkillFilter; label: string }[] = [
    { key: 'all', label: '全部技能' },
    { key: 'global', label: '全局技能' },
    { key: 'project', label: '项目技能' },
  ];

  function resetPanelState() {
    setAllItems([]);
    setLoading(false);
    setPage(1);
    setDetailSkill(null);
  }

  return (
    <section
      className={cn('grid min-h-0 grid-rows-[auto_minmax(0,1fr)] gap-3 overflow-hidden p-4', active ? '' : 'hidden')}
      data-testid="skills-panel"
    >
      <Toolbar>
        <div className="inline-flex rounded-[6px] border border-border bg-muted/50 p-1">
          {tabs.map((item) => (
            <TabButton
              key={item.key}
              active={filter === item.key}
              className="h-8 rounded-[5px]"
              onClick={() => setFilter(item.key)}
            >
              {item.label}
            </TabButton>
          ))}
        </div>
        <SearchField className="w-full max-w-sm" placeholder="搜索技能名称、描述或路径" value={query} onChange={(event) => setQuery(event.target.value)} />
        <Button disabled={!agentReady || loading} type="button" variant="soft" onClick={() => void loadSkills({ notify: true })}>
          <RefreshCw className="h-4 w-4" />
          刷新
        </Button>
        <FileUploadButton disabled={!agentReady || loading} multiple onFiles={(files) => void uploadProjectSkills(files)}>
          <Upload className="h-4 w-4" />
          上传技能
        </FileUploadButton>
      </Toolbar>

      <div className="flex min-h-0 flex-col gap-3 overflow-auto">
        {loading && allItems.length === 0 ? <Spinner /> : null}
        {!loading && !agentReady ? (
          <EmptyState
            icon={<PowerOff className="h-5 w-5" />}
            title="当前 Agent 未运行"
            description={agent ? `当前状态：${statusLabel(agent.runtimeStatus || 'unknown')}。启动后即可查看技能。` : '请选择一个运行中的智能体。'}
          />
        ) : null}
        {!loading && agentReady && filteredSkills.length === 0 ? (
          <EmptyState icon={<SearchX className="h-5 w-5" />} title="暂无匹配技能" description="调整搜索条件或切换技能范围。" />
        ) : null}
        {agentReady && filteredSkills.length > 0 ? (
          <>
            <SkillDataTable pageItems={pageItems} onDetail={setDetailSkill} />
            {totalPages > 1 ? (
              <Pagination
                page={currentPage}
                pages={totalPages}
                label={`共 ${filteredSkills.length} 条`}
                onPrev={() => setPage((p) => Math.max(1, p - 1))}
                onNext={() => setPage((p) => Math.min(totalPages, p + 1))}
              />
            ) : null}
          </>
        ) : null}
      </div>

      <Dialog
        open={detailSkill !== null}
        title="技能详情"
        onClose={() => setDetailSkill(null)}
      >
        {detailSkill ? (
          <div className="grid gap-4 text-sm">
            <DetailRow label="名称" value={detailSkill.name} />
            <DetailRow
              label="范围"
              value={
                <Badge tone={detailSkill.scope === 'global' ? 'info' : 'success'}>
                  {detailSkill.scope === 'global' ? '全局技能' : '项目技能'}
                </Badge>
              }
            />
            {detailSkill.description ? <DetailRow label="描述" value={detailSkill.description} /> : null}
            <DetailRow
              label="路径"
              value={
                <span className="inline-flex items-center gap-1 break-all font-mono text-xs text-muted-foreground">
                  <Inbox className="h-3.5 w-3.5 shrink-0" />
                  {detailSkill.path}
                </span>
              }
            />
            {detailSkill.source ? (
              <DetailRow
                label="来源"
                value={<span className="break-all font-mono text-xs text-muted-foreground">{detailSkill.source}</span>}
              />
            ) : null}
            <DetailRow
              label="SKILL.md"
              value={
                <Badge tone={detailSkill.hasManifest ? 'success' : 'neutral'}>
                  {detailSkill.hasManifest ? '已配置' : '未配置'}
                </Badge>
              }
            />
          </div>
        ) : null}
      </Dialog>
    </section>
  );
}

function SkillDataTable({ pageItems, onDetail }: { pageItems: WorkspaceSkillInfo[]; onDetail: (skill: WorkspaceSkillInfo) => void }) {
  const columns = useMemo<ColumnDef<WorkspaceSkillInfo>[]>(() => [
    {
      accessorKey: 'name',
      header: '技能名称',
      cell: ({ row }) => <span className="font-medium whitespace-nowrap">{row.original.name}</span>,
    },
    {
      accessorKey: 'description',
      header: '描述',
      enableSorting: false,
      cell: ({ row }) => <span className="block max-w-xs truncate text-muted-foreground">{row.original.description || '暂无描述'}</span>,
    },
    {
      accessorKey: 'scope',
      header: '范围',
      cell: ({ row }) => (
        <Badge tone={row.original.scope === 'global' ? 'info' : 'success'}>
          {row.original.scope === 'global' ? '全局' : '项目'}
        </Badge>
      ),
    },
    {
      id: 'actions',
      header: '操作',
      enableSorting: false,
      cell: ({ row }) => (
        <Button size="sm" type="button" variant="ghost" onClick={() => onDetail(row.original)}>
          详情
        </Button>
      ),
    },
  ], [onDetail]);

  return (
    <DataTable<WorkspaceSkillInfo>
      className="rounded-lg border shadow-xs"
      columns={columns}
      data={pageItems}
      getRowId={(row) => `${row.scope}:${row.path}`}
      emptyText="暂无匹配技能"
    />
  );
}

function DetailRow({ label, value }: { label: string; value: ReactNode }) {
  return (
    <div className="grid grid-cols-[6rem_1fr] gap-3">
      <span className="text-xs font-medium text-muted-foreground">{label}</span>
      <span className="min-w-0">{value}</span>
    </div>
  );
}

function statusLabel(value: string) {
  const labels: Record<string, string> = {
    created: '未启动',
    dead: '异常退出',
    exited: '已退出',
    missing: '容器离线',
    paused: '已暂停',
    stopped: '已停止',
    unavailable: 'Docker 不可用',
  };
  return labels[value] ?? value;
}
