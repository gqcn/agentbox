import { useEffect, useMemo, useState } from 'react';
import { BrainCircuit, FlaskConical, RefreshCw, Save, Sparkles, Zap } from 'lucide-react';
import { toast } from 'sonner';
import { api } from './api';
import type {
  AICapabilityTestResult,
  AICapabilityTierCode,
  AICapabilityTierInfo,
  ProviderInfo,
  ProviderModel,
} from './types';
import {
  Alert,
  Badge,
  Button,
  EmptyState,
  Field,
  IconButton,
  Select,
  SelectOption,
  Switch,
} from '@/components/ui';
import { cn } from '@/lib/utils';

type TierForm = {
  enabled: boolean;
  providerId: number;
  providerModelId: number;
  protocol: ProviderModel['protocol'] | '';
};

type CapabilityModelOption = Pick<ProviderModel, 'id' | 'providerId' | 'name' | 'protocol'> & {
  synthetic: boolean;
};

const tierOrder: AICapabilityTierCode[] = ['basic', 'standard', 'advanced'];
const tierCopy: Record<AICapabilityTierCode, { label: string; subtitle: string; icon: typeof Zap }> = {
  basic: { label: '基础', subtitle: 'Git commit message、摘要、短文本', icon: Zap },
  standard: { label: '标准', subtitle: '常规代码生成、解释与优化', icon: Sparkles },
  advanced: { label: '高级', subtitle: '复杂生成、多文件推理与架构优化', icon: BrainCircuit },
};
export default function AICapabilitiesPage({ providers }: { providers: ProviderInfo[] }) {
  const [tiers, setTiers] = useState<AICapabilityTierInfo[]>([]);
  const [forms, setForms] = useState<Record<string, TierForm>>({});
  const [loading, setLoading] = useState(false);
  const [savingTier, setSavingTier] = useState('');
  const [testingTier, setTestingTier] = useState('');
  const [testResults, setTestResults] = useState<Record<string, AICapabilityTestResult>>({});

  const configuredCount = tiers.filter((tier) => tier.configured).length;
  const availableCount = tiers.filter((tier) => tier.available).length;

  useEffect(() => {
    void loadTiers();
  }, []);

  async function loadTiers() {
    setLoading(true);
    try {
      const items = await api.listAICapabilityTiers();
      setTiers(sortTiers(items));
      setForms(formsFromTiers(items));
    } catch (error) {
      toast.error((error as Error).message);
    } finally {
      setLoading(false);
    }
  }

  async function saveTier(code: AICapabilityTierCode) {
    const form = forms[code];
    if (!form) {
      return;
    }
    setSavingTier(code);
    try {
      const updated = await api.updateAICapabilityTier(code, {
        enabled: form.enabled,
        providerId: form.providerId,
        providerModelId: form.providerModelId,
        protocol: protocolForPayload(providers, form),
      });
      setTiers((current) => sortTiers(current.map((tier) => tier.code === code ? updated : tier)));
      setForms((current) => ({ ...current, [code]: formFromTier(updated) }));
      toast.success(`${tierCopy[code].label}档位已保存`);
    } catch (error) {
      toast.error((error as Error).message);
    } finally {
      setSavingTier('');
    }
  }

  async function testTier(code: AICapabilityTierCode) {
    const form = forms[code];
    if (!form) {
      return;
    }
    setTestingTier(code);
    try {
      const result = await api.testAICapabilityTier(code, {
        providerId: form.providerId,
        providerModelId: form.providerModelId,
        protocol: protocolForPayload(providers, form),
      });
      setTestResults((current) => ({ ...current, [code]: result }));
      const refreshed = await api.listAICapabilityTiers();
      setTiers(sortTiers(refreshed));
      setForms((current) => ({ ...formsFromTiers(refreshed), ...current }));
      if (result.status === 'success') {
        toast.success(`${tierCopy[code].label}档位测试通过`);
      } else {
        toast.error(result.errorMessage || `${tierCopy[code].label}档位测试失败`);
      }
    } catch (error) {
      toast.error((error as Error).message);
    } finally {
      setTestingTier('');
    }
  }

  function updateForm(code: AICapabilityTierCode, patch: Partial<TierForm>) {
    setForms((current) => {
      const previous = current[code] ?? blankTierForm();
      const next = { ...previous, ...patch };
      if ('providerId' in patch) {
        const provider = providers.find((item) => item.id === next.providerId);
        const firstModel = providerModelOptions(provider)[0];
        next.providerModelId = firstModel?.id ?? 0;
        next.protocol = firstModel?.protocol ?? '';
      }
      return { ...current, [code]: next };
    });
  }

  return (
    <section className="grid min-h-0 grid-rows-[auto_minmax(0,1fr)] overflow-hidden rounded-[8px] border border-border bg-card" data-testid="ai-capabilities-page">
      <div className="flex min-h-14 items-center justify-between gap-3 border-b border-border px-4 py-3 max-[760px]:flex-col max-[760px]:items-stretch">
        <div className="grid gap-1">
          <div className="text-sm font-semibold text-foreground">能力档位</div>
          <div className="text-xs text-muted-foreground">已配置 {configuredCount}/3，可用 {availableCount}/3</div>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <StatPill label="基础" value={tierStateLabel(tiers.find((tier) => tier.code === 'basic'))} />
          <StatPill label="标准" value={tierStateLabel(tiers.find((tier) => tier.code === 'standard'))} />
          <StatPill label="高级" value={tierStateLabel(tiers.find((tier) => tier.code === 'advanced'))} />
          <IconButton disabled={loading} title="刷新 AI 能力" onClick={() => void loadTiers()}>
            <RefreshCw className={cn('h-4 w-4', loading && 'animate-spin')} />
          </IconButton>
        </div>
      </div>

      <div className="min-h-0 overflow-auto p-4">
        {providers.length === 0 ? (
          <Alert tone="warning">需要先在“供应商”中配置供应商和模型，AI 能力档位才能绑定可用模型。</Alert>
        ) : null}
        <div className="mt-0 grid gap-4 xl:grid-cols-3">
          {tierOrder.map((code) => (
            <TierPanel
              form={forms[code] ?? blankTierForm()}
              key={code}
              loading={loading}
              providers={providers}
              saving={savingTier === code}
              testing={testingTier === code}
              testResult={testResults[code]}
              tier={tiers.find((item) => item.code === code)}
              tierCode={code}
              onSave={() => void saveTier(code)}
              onTest={() => void testTier(code)}
              onUpdate={(patch) => updateForm(code, patch)}
            />
          ))}
        </div>

      </div>
    </section>
  );
}

function TierPanel({
  form,
  loading,
  providers,
  saving,
  testing,
  testResult,
  tier,
  tierCode,
  onSave,
  onTest,
  onUpdate,
}: {
  form: TierForm;
  loading: boolean;
  providers: ProviderInfo[];
  saving: boolean;
  testing: boolean;
  testResult?: AICapabilityTestResult;
  tier?: AICapabilityTierInfo;
  tierCode: AICapabilityTierCode;
  onSave: () => void;
  onTest: () => void;
  onUpdate: (patch: Partial<TierForm>) => void;
}) {
  const copy = tierCopy[tierCode];
  const Icon = copy.icon;
  const selectedProvider = providers.find((provider) => provider.id === form.providerId);
  const models = providerModelOptions(selectedProvider);
  const selectedModelValue = selectedModelOptionValue(models, form, tier?.binding?.modelName);
  const selectedModel = selectedModelOption(models, providers, form, tier?.binding?.modelName);
  const effectiveResult = testResult ?? tier?.lastTest;
  const canSave = Boolean(form.providerId && form.providerModelId) || Boolean(tier?.binding);
  const canTest = Boolean(form.enabled && form.providerId && form.providerModelId);

  return (
    <article className="grid min-h-[420px] grid-rows-[auto_auto_minmax(0,1fr)_auto] rounded-[8px] border border-border bg-background" data-testid={`ai-tier-${tierCode}`}>
      <header className="flex items-start justify-between gap-3 border-b border-border p-4">
        <div className="flex min-w-0 items-start gap-3">
          <div className="grid h-10 w-10 shrink-0 place-items-center rounded-[8px] border border-primary/25 bg-primary/10 text-primary">
            <Icon className="h-5 w-5" />
          </div>
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-2">
              <h3 className="text-base font-semibold">{copy.label}</h3>
              <Badge tone={tier?.available ? 'success' : form.enabled ? 'warning' : 'neutral'}>
                {tierStateLabel(tier)}
              </Badge>
            </div>
            <p className="mt-1 text-xs leading-5 text-muted-foreground">{copy.subtitle}</p>
          </div>
        </div>
        <Switch
          checked={form.enabled}
          title={`${copy.label}档位启用状态`}
          onChange={(event) => onUpdate({ enabled: event.currentTarget.checked })}
        />
      </header>

      <div className="grid gap-2 border-b border-border px-4 py-3 text-xs">
        <div className="flex items-center justify-between gap-2">
          <span className="text-muted-foreground">协议</span>
          <Badge tone={selectedModel?.protocol ? 'info' : 'neutral'}>{selectedModel ? protocolLabel(selectedModel.protocol) : '未选择'}</Badge>
        </div>
        <div className="flex items-center justify-between gap-2">
          <span className="text-muted-foreground">当前模型</span>
          <span className="truncate text-right text-foreground">{tier?.binding?.modelName ?? '未配置'}</span>
        </div>
      </div>

      <div className="grid content-start gap-3 p-4">
        <Field label="供应商">
          <Select
            disabled={providers.length === 0}
            value={form.providerId || ''}
            onChange={(event) => onUpdate({ providerId: Number(event.target.value) })}
          >
            <SelectOption value="" disabled>选择供应商</SelectOption>
            {providers.map((provider) => (
              <SelectOption disabled={providerModelOptions(provider).length === 0} key={provider.id} value={provider.id}>
                {provider.name}
              </SelectOption>
            ))}
          </Select>
        </Field>
        <Field label="模型">
          <Select
            disabled={models.length === 0}
            value={selectedModelValue}
            onChange={(event) => {
              const option = modelOptionFromValue(models, event.target.value);
              onUpdate({
                providerModelId: option?.id ?? 0,
                protocol: option?.protocol ?? '',
              });
            }}
          >
            <SelectOption value="" disabled>选择模型</SelectOption>
            {models.map((model) => (
              <SelectOption key={modelOptionValue(model)} value={modelOptionValue(model)}>
                {model.name} / {protocolLabel(model.protocol)}
              </SelectOption>
            ))}
          </Select>
        </Field>

        {effectiveResult ? (
          <div className="rounded-[6px] border border-border bg-muted/40 p-3 text-xs">
            <div className="mb-2 flex items-center justify-between gap-2">
              <span className="font-medium text-foreground">最近测试</span>
              <Badge tone={effectiveResult.status === 'success' ? 'success' : 'danger'}>
                {effectiveResult.status === 'success' ? '通过' : '失败'}
              </Badge>
            </div>
            <div className="grid gap-1 text-muted-foreground">
              <div>{formatTimestamp(resultTimestamp(effectiveResult, tier))}</div>
              <div>{effectiveResult.modelName || tier?.lastTest?.modelName || selectedModel?.name || '-'}</div>
              <div>{effectiveResult.latencyMs || tier?.lastTest?.latencyMs || 0} ms</div>
              {effectiveResult.errorMessage ? <div className="text-destructive">{effectiveResult.errorMessage}</div> : null}
            </div>
          </div>
        ) : (
          <EmptyState className="min-h-28 p-4" icon={<FlaskConical />} title="尚未测试" description="选择供应商和模型后可先测试，再按需保存档位。" />
        )}
      </div>

      <footer className="flex items-center justify-between gap-2 border-t border-border p-4">
        <Button disabled={!canTest || testing || saving || loading} type="button" variant="soft" onClick={onTest}>
          <FlaskConical className={cn('h-4 w-4', testing && 'animate-pulse')} />
          测试
        </Button>
        <Button disabled={!canSave || saving || loading} type="button" variant="primary" onClick={onSave}>
          {saving ? <RefreshCw className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
          保存
        </Button>
      </footer>
    </article>
  );
}

function StatPill({ label, value }: { label: string; value: string }) {
  return (
    <div className="inline-flex h-8 items-center gap-2 rounded-[6px] border border-border bg-background px-2 text-xs">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-medium text-foreground">{value}</span>
    </div>
  );
}

function sortTiers(items: AICapabilityTierInfo[]) {
  return [...items].sort((a, b) => tierOrder.indexOf(a.code) - tierOrder.indexOf(b.code));
}

function formsFromTiers(items: AICapabilityTierInfo[]) {
  return items.reduce<Record<string, TierForm>>((result, tier) => {
    result[tier.code] = formFromTier(tier);
    return result;
  }, {});
}

function formFromTier(tier: AICapabilityTierInfo): TierForm {
  return {
    enabled: tier.enabled,
    providerId: tier.binding?.providerId ?? 0,
    providerModelId: tier.binding?.providerModelId ?? 0,
    protocol: tier.binding?.protocol ?? '',
  };
}

function blankTierForm(): TierForm {
  return {
    enabled: true,
    providerId: 0,
    providerModelId: 0,
    protocol: '',
  };
}

function protocolForPayload(providers: ProviderInfo[], form: TierForm): ProviderModel['protocol'] | undefined {
  if (form.protocol) {
    return form.protocol;
  }
  return selectedModelOption(providerModelOptions(providers.find((provider) => provider.id === form.providerId)), providers, form)?.protocol;
}

function providerModelById(providers: ProviderInfo[], modelId: number): ProviderModel | undefined {
  for (const provider of providers) {
    const model = provider.models?.find((item) => item.id === modelId);
    if (model) {
      return model;
    }
  }
  return undefined;
}

function selectedModelOption(models: CapabilityModelOption[], providers: ProviderInfo[], form: TierForm, bindingModelName = ''): CapabilityModelOption | ProviderModel | undefined {
  return models.find((model) => model.id === form.providerModelId && model.protocol === form.protocol)
    ?? models.find((model) => model.id === form.providerModelId)
    ?? models.find((model) => bindingModelName && model.name === bindingModelName && model.protocol === form.protocol)
    ?? providerModelById(providers, form.providerModelId);
}

function selectedModelOptionValue(models: CapabilityModelOption[], form: TierForm, bindingModelName = '') {
  const exact = models.find((model) => model.id === form.providerModelId && model.protocol === form.protocol);
  if (exact) {
    return modelOptionValue(exact);
  }
  const fallback = models.find((model) => model.id === form.providerModelId);
  if (fallback) {
    return modelOptionValue(fallback);
  }
  const sameNameProtocol = models.find((model) => bindingModelName && model.name === bindingModelName && model.protocol === form.protocol);
  return sameNameProtocol ? modelOptionValue(sameNameProtocol) : '';
}

function modelOptionFromValue(models: CapabilityModelOption[], value: string): CapabilityModelOption | undefined {
  return models.find((model) => modelOptionValue(model) === value);
}

function modelOptionValue(model: Pick<ProviderModel, 'id' | 'protocol'>) {
  return `${model.id}:${model.protocol}`;
}

function providerModelOptions(provider?: ProviderInfo): CapabilityModelOption[] {
  if (!provider) {
    return [];
  }
  const options = new Map<string, CapabilityModelOption>();
  const nameProtocolKeys = new Set<string>();
  for (const model of provider.models ?? []) {
    options.set(modelOptionValue(model), { ...model, synthetic: false });
    nameProtocolKeys.add(modelNameProtocolKey(model.name, model.protocol));
  }
  const supportedProtocols = providerSupportedProtocols(provider);
  for (const model of provider.models ?? []) {
    for (const protocol of supportedProtocols) {
      if (nameProtocolKeys.has(modelNameProtocolKey(model.name, protocol))) {
        continue;
      }
      const key = `${model.id}:${protocol}`;
      if (!options.has(key)) {
        options.set(key, {
          id: model.id,
          providerId: model.providerId,
          name: model.name,
          protocol,
          synthetic: true,
        });
        nameProtocolKeys.add(modelNameProtocolKey(model.name, protocol));
      }
    }
  }
  return [...options.values()].sort((left, right) => {
    const byName = left.name.localeCompare(right.name);
    if (byName !== 0) {
      return byName;
    }
    return protocolOrder(left.protocol) - protocolOrder(right.protocol);
  });
}

function modelNameProtocolKey(name: string, protocol: ProviderModel['protocol']) {
  return `${name}\u0000${protocol}`;
}

function providerSupportedProtocols(provider: ProviderInfo): ProviderModel['protocol'][] {
  const protocols = new Set<ProviderModel['protocol']>();
  if (provider.openaiBaseUrl.trim()) {
    protocols.add('openai');
  }
  if (provider.anthropicBaseUrl.trim()) {
    protocols.add('anthropic');
  }
  if (protocols.size === 0) {
    provider.models?.forEach((model) => protocols.add(model.protocol));
  }
  return (['openai', 'anthropic'] as const).filter((protocol) => protocols.has(protocol));
}

function tierStateLabel(tier?: AICapabilityTierInfo) {
  if (!tier) {
    return '加载中';
  }
  if (!tier.enabled) {
    return '已禁用';
  }
  if (!tier.configured) {
    return '未配置';
  }
  return tier.available ? '可用' : '不可用';
}

function protocolLabel(value: string | undefined) {
  return value === 'openai' ? 'OpenAI' : value === 'anthropic' ? 'Anthropic' : '未选择';
}

function protocolOrder(value: ProviderModel['protocol']) {
  return value === 'openai' ? 0 : 1;
}

function formatTimestamp(value?: number) {
  if (!value) {
    return '-';
  }
  return new Intl.DateTimeFormat(undefined, {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value));
}

function resultTimestamp(result: AICapabilityTestResult | NonNullable<AICapabilityTierInfo['lastTest']>, tier?: AICapabilityTierInfo) {
  if ('testedAt' in result) {
    return result.testedAt;
  }
  return result.createdAt || tier?.lastTest?.createdAt;
}
