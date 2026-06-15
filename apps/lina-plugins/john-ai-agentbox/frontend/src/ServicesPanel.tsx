import { useEffect, useMemo, useState } from "react";
import {
  Copy,
  ExternalLink,
  Plug,
  PowerOff,
  RefreshCw,
  Unplug,
} from "lucide-react";
import { toast } from "sonner";
import { api } from "./api";
import type {
  AgentInfo,
  AgentRuntimeServiceInfo,
  AgentServiceAccessStatus,
  AgentServiceListenAddress,
  AgentServiceProtocol,
} from "./types";
import type { ColumnDef } from "@/components/ui";
import {
  Alert,
  Badge,
  Button,
  DataTable,
  EmptyState,
  Spinner,
  Toolbar,
} from "@/components/ui";
import { cn } from "@/lib/utils";

type Props = {
  active: boolean;
  agent?: AgentInfo;
};

const statusMeta: Record<
  AgentServiceAccessStatus,
  { label: string; tone: "success" | "warning" | "info" | "neutral" }
> = {
  direct: { label: "可直接访问", tone: "success" },
  bridge_required: { label: "需手动开启", tone: "warning" },
  bridged: { label: "桥接中", tone: "info" },
  unavailable: { label: "不可访问", tone: "neutral" },
};

const protocolLabel: Record<AgentServiceProtocol, string> = {
  http: "HTTP",
  https: "HTTPS",
  tcp: "TCP",
  unknown: "Unknown",
};

export default function ServicesPanel({ active, agent }: Props) {
  const [services, setServices] = useState<AgentRuntimeServiceInfo[]>([]);
  const [loading, setLoading] = useState(false);
  const [busyServiceId, setBusyServiceId] = useState("");
  const [error, setError] = useState("");
  const agentReady = Boolean(agent?.id && agent.runtimeStatus === "running");

  useEffect(() => {
    if (!active) {
      return;
    }
    if (!agentReady) {
      setServices([]);
      setError("");
      return;
    }
    void loadServices();
  }, [active, agent?.id, agent?.runtimeStatus]);

  const columns = useMemo<ColumnDef<AgentRuntimeServiceInfo>[]>(
    () => [
      {
        accessorKey: "port",
        header: "端口",
        cell: ({ row }) => (
          <div className="flex min-w-0 items-center gap-2">
            <span className="font-mono text-sm font-semibold">
              {row.original.port}
            </span>
            <Badge tone="neutral">
              {protocolLabel[row.original.protocol] ?? row.original.protocol}
            </Badge>
          </div>
        ),
      },
      {
        id: "listenAddresses",
        header: "监听地址",
        cell: ({ row }) => (
          <div className="flex max-w-[320px] flex-wrap gap-1.5">
            {row.original.listenAddresses.map((address) => (
              <Badge
                key={`${address.network}:${address.address}:${address.port}`}
                className="font-mono"
                tone={
                  address.accessStatus === "bridge_required"
                    ? "warning"
                    : address.accessStatus === "bridged"
                      ? "info"
                      : address.accessStatus === "direct"
                        ? "success"
                        : "neutral"
                }
              >
                {formatListenAddress(address)}
              </Badge>
            ))}
          </div>
        ),
      },
      {
        id: "process",
        header: "进程",
        cell: ({ row }) => (
          <div className="min-w-0">
            <div className="truncate text-sm">
              {row.original.processName || "未知"}
            </div>
            {row.original.pid ? (
              <div className="font-mono text-xs text-muted-foreground">
                PID {row.original.pid}
              </div>
            ) : null}
          </div>
        ),
      },
      {
        accessorKey: "accessStatus",
        header: "访问状态",
        cell: ({ row }) => <ServiceStatus service={row.original} />,
      },
      {
        id: "localAccess",
        header: "本地访问",
        cell: ({ row }) => (
          <LocalAccess service={row.original} onCopy={copyText} />
        ),
      },
      {
        id: "actions",
        header: "操作",
        cell: ({ row }) => (
          <ServiceActions
            agent={agent}
            busy={busyServiceId === row.original.id}
            service={row.original}
            onBridge={(service, address) => void createBridge(service, address)}
            onCloseBridge={(service) => void closeBridge(service)}
            onCopy={copyText}
            onOpen={openService}
          />
        ),
      },
    ],
    [agent?.id, busyServiceId],
  );

  async function loadServices({ notify = false }: { notify?: boolean } = {}) {
    if (!agent?.id || !agentReady) {
      return;
    }
    setLoading(true);
    setError("");
    try {
      const items = await api.listAgentServices(agent.id);
      setServices(items);
      if (notify) {
        toast.success("服务列表已刷新");
      }
    } catch (err) {
      const message = (err as Error).message;
      setError(message);
      setServices([]);
      toast.error(message);
    } finally {
      setLoading(false);
    }
  }

  async function createBridge(
    service: AgentRuntimeServiceInfo,
    address: AgentServiceListenAddress,
  ) {
    if (!agent?.id) {
      return;
    }
    setBusyServiceId(service.id);
    try {
      await api.createAgentServiceBridge(agent.id, {
        serviceId: service.id,
        listenAddress: address.address,
        port: address.port,
      });
      toast.success("桥接已开启");
      await loadServices();
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      setBusyServiceId("");
    }
  }

  async function closeBridge(service: AgentRuntimeServiceInfo) {
    if (!agent?.id || !service.bridgeId) {
      return;
    }
    setBusyServiceId(service.id);
    try {
      await api.deleteAgentServiceBridge(agent.id, service.bridgeId);
      toast.success("桥接已关闭");
      await loadServices();
    } catch (err) {
      toast.error((err as Error).message);
    } finally {
      setBusyServiceId("");
    }
  }

  async function copyText(value: string, success: string) {
    if (!value) {
      return;
    }
    const text = value.startsWith("/")
      ? new URL(value, window.location.origin).toString()
      : value;
    try {
      await navigator.clipboard.writeText(text);
      toast.success(success);
    } catch {
      toast.error("复制失败");
    }
  }

  function openService(service: AgentRuntimeServiceInfo) {
    if (!service.proxyUrl) {
      return;
    }
    window.open(
      new URL(service.proxyUrl, window.location.origin).toString(),
      "_blank",
      "noopener,noreferrer",
    );
  }

  if (!active) {
    return null;
  }

  return (
    <section
      className="flex h-full min-h-0 min-w-0 flex-col bg-background"
      data-testid="services-panel"
    >
      <div className="flex min-h-12 items-center justify-between gap-3 border-b px-3 py-2">
        <div className="min-w-0">
          <h2 className="truncate text-sm font-semibold">服务</h2>
          <p className="truncate text-xs text-muted-foreground">
            {agentReady
              ? `${services.length} 个监听端口`
              : "启动 Agent 后检测容器监听端口"}
          </p>
        </div>
        <Toolbar className="min-h-0 justify-end">
          <Button
            disabled={!agentReady || loading}
            size="sm"
            type="button"
            variant="soft"
            onClick={() => void loadServices({ notify: true })}
          >
            <RefreshCw
              className={cn(loading && "animate-spin")}
              data-icon="inline-start"
            />
            刷新
          </Button>
        </Toolbar>
      </div>

      <div className="flex min-h-0 flex-1 flex-col gap-3 overflow-hidden p-3">
        {!agentReady ? (
          <EmptyState
            description="当前 Agent 未运行，服务发现、桥接和隧道操作已禁用。"
            icon={<PowerOff className="h-5 w-5" />}
            title="需要启动 Agent"
          />
        ) : error ? (
          <Alert tone="danger">{error}</Alert>
        ) : loading && services.length === 0 ? (
          <div className="grid min-h-44 place-items-center rounded-lg border bg-card">
            <Spinner label="正在检测服务" />
          </div>
        ) : services.length === 0 ? (
          <EmptyState title="暂无监听服务" />
        ) : (
          <>
            <div className="flex flex-col gap-2 md:hidden">
              {services.map((service) => (
                <ServiceMobileCard
                  key={service.id}
                  agent={agent}
                  busy={busyServiceId === service.id}
                  service={service}
                  onBridge={(item, address) => void createBridge(item, address)}
                  onCloseBridge={(item) => void closeBridge(item)}
                  onCopy={copyText}
                  onOpen={openService}
                />
              ))}
            </div>
            <DataTable
              className="hidden min-h-0 flex-1 rounded-lg border shadow-xs md:block"
              columns={columns}
              data={services}
              emptyText="暂无监听服务"
              getRowId={(service) => service.id}
            />
          </>
        )}
      </div>
    </section>
  );
}

function ServiceMobileCard({
  agent,
  busy,
  service,
  onBridge,
  onCloseBridge,
  onCopy,
  onOpen,
}: {
  agent?: AgentInfo;
  busy: boolean;
  service: AgentRuntimeServiceInfo;
  onBridge: (
    service: AgentRuntimeServiceInfo,
    address: AgentServiceListenAddress,
  ) => void;
  onCloseBridge: (service: AgentRuntimeServiceInfo) => void;
  onCopy: (value: string, success: string) => void;
  onOpen: (service: AgentRuntimeServiceInfo) => void;
}) {
  return (
    <article className="flex min-w-0 flex-col gap-3 rounded-lg border bg-card p-3">
      <div className="flex min-w-0 items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex min-w-0 items-center gap-2">
            <span className="font-mono text-sm font-semibold">
              {service.port}
            </span>
            <Badge tone="neutral">
              {protocolLabel[service.protocol] ?? service.protocol}
            </Badge>
          </div>
          <div className="mt-1 flex flex-wrap gap-1.5">
            {service.listenAddresses.map((address) => (
              <Badge
                key={`${address.network}:${address.address}:${address.port}`}
                className="font-mono"
                tone={
                  address.accessStatus === "bridge_required"
                    ? "warning"
                    : address.accessStatus === "bridged"
                      ? "info"
                      : address.accessStatus === "direct"
                        ? "success"
                        : "neutral"
                }
              >
                {formatListenAddress(address)}
              </Badge>
            ))}
          </div>
        </div>
        <ServiceStatus service={service} />
      </div>
      <div className="min-w-0 text-sm">
        <span>{service.processName || "未知"}</span>
        {service.pid ? (
          <span className="ml-2 font-mono text-xs text-muted-foreground">
            PID {service.pid}
          </span>
        ) : null}
      </div>
      <LocalAccess service={service} onCopy={onCopy} />
      <ServiceActions
        agent={agent}
        busy={busy}
        service={service}
        onBridge={onBridge}
        onCloseBridge={onCloseBridge}
        onCopy={onCopy}
        onOpen={onOpen}
      />
    </article>
  );
}

function ServiceStatus({ service }: { service: AgentRuntimeServiceInfo }) {
  const meta = statusMeta[service.accessStatus] ?? statusMeta.unavailable;
  return (
    <div className="flex min-w-0 flex-col gap-1">
      <Badge tone={meta.tone}>{meta.label}</Badge>
      {service.unavailableReason ? (
        <span className="max-w-[220px] truncate text-xs text-muted-foreground">
          {service.unavailableReason}
        </span>
      ) : null}
    </div>
  );
}

function LocalAccess({
  service,
  onCopy,
}: {
  service: AgentRuntimeServiceInfo;
  onCopy: (value: string, success: string) => void;
}) {
  const localAddress = formatLocalAccess(service);
  if (!localAddress) {
    return <span className="text-xs text-muted-foreground">暂无</span>;
  }
  return (
    <div className="flex min-w-0 flex-wrap items-center gap-1.5">
      <span className="min-w-0 truncate font-mono text-xs" title={localAddress}>
        {localAddress}
      </span>
      <Button
        size="sm"
        type="button"
        variant="ghost"
        onClick={() => onCopy(localAddress, "本地地址已复制")}
      >
        <Copy data-icon="inline-start" />
        复制地址
      </Button>
    </div>
  );
}

function ServiceActions({
  agent,
  busy,
  service,
  onBridge,
  onCloseBridge,
  onCopy,
  onOpen,
}: {
  agent?: AgentInfo;
  busy: boolean;
  service: AgentRuntimeServiceInfo;
  onBridge: (
    service: AgentRuntimeServiceInfo,
    address: AgentServiceListenAddress,
  ) => void;
  onCloseBridge: (service: AgentRuntimeServiceInfo) => void;
  onCopy: (value: string, success: string) => void;
  onOpen: (service: AgentRuntimeServiceInfo) => void;
}) {
  const bridgeAddress = service.listenAddresses.find(
    (address) => address.accessStatus === "bridge_required",
  );
  const canOpen = Boolean(
    service.proxyUrl &&
    (service.protocol === "http" || service.protocol === "https"),
  );
  return (
    <div className="flex min-w-0 flex-wrap justify-start gap-1.5 md:min-w-[220px] md:justify-center">
      {canOpen ? (
        <>
          <Button
            disabled={!agent || busy}
            size="sm"
            type="button"
            variant="soft"
            onClick={() => onOpen(service)}
          >
            <ExternalLink data-icon="inline-start" />
            打开
          </Button>
          <Button
            disabled={!agent || busy}
            size="sm"
            type="button"
            variant="ghost"
            onClick={() => onCopy(service.proxyUrl ?? "", "链接已复制")}
          >
            <Copy data-icon="inline-start" />
            复制链接
          </Button>
        </>
      ) : null}
      {bridgeAddress ? (
        <Button
          disabled={!agent || busy}
          size="sm"
          type="button"
          variant="soft"
          onClick={() => onBridge(service, bridgeAddress)}
        >
          <Plug data-icon="inline-start" />
          开启桥接
        </Button>
      ) : null}
      {service.bridgeId ? (
        <Button
          disabled={!agent || busy}
          size="sm"
          type="button"
          variant="ghost"
          onClick={() => onCloseBridge(service)}
        >
          <Unplug data-icon="inline-start" />
          关闭桥接
        </Button>
      ) : null}
      {!canOpen && !bridgeAddress && !service.bridgeId ? (
        <span className="text-xs text-muted-foreground">无可用操作</span>
      ) : null}
    </div>
  );
}

function formatListenAddress(address: AgentServiceListenAddress) {
  return `${address.address}:${address.port}`;
}

function formatLocalAccess(service: AgentRuntimeServiceInfo) {
  if (!service.localHost || !service.localPort) {
    return "";
  }
  const host = formatHostForURL(service.localHost);
  if (service.protocol === "http" || service.protocol === "https") {
    return `${service.protocol}://${host}:${service.localPort}`;
  }
  return `${host}:${service.localPort}`;
}

function formatHostForURL(host: string) {
  return host.includes(":") && !host.startsWith("[") ? `[${host}]` : host;
}
