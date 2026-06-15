import { lazy, Suspense, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import type { OnMount } from '@monaco-editor/react';
import { RefreshCw, RotateCcw, Save, ScrollText, WandSparkles } from 'lucide-react';
import { toast } from 'sonner';
import { api } from './api';
import type { PromptTemplateInfo } from './types';
import {
  Alert,
  Badge,
  Button,
  ConfirmDialog,
  EmptyState,
  Field,
  IconButton,
  Skeleton,
  Spinner,
} from '@/components/ui';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { loadMonacoEditor } from '@/lib/monaco-loader';
import { cn } from '@/lib/utils';

type PromptForm = {
  content: string;
};

const MonacoEditor = lazy(loadMonacoEditor);
const gitCommitPromptCode = 'git_commit_message';

type PromptCodeEditorHost = HTMLDivElement & {
  __agentBoxPromptEditor?: {
    focus: () => void;
    getValue: () => string;
    setValue: (content: string) => void;
  };
};

export default function PromptsPage() {
  const [templates, setTemplates] = useState<PromptTemplateInfo[]>([]);
  const [forms, setForms] = useState<Record<string, PromptForm>>({});
  const [activeCode, setActiveCode] = useState(gitCommitPromptCode);
  const [loading, setLoading] = useState(false);
  const [savingCode, setSavingCode] = useState('');
  const [restoringCode, setRestoringCode] = useState('');
  const [errorByCode, setErrorByCode] = useState<Record<string, string>>({});

  const activeTemplate = useMemo(
    () => templates.find((template) => template.code === activeCode) ?? templates[0],
    [activeCode, templates],
  );

  useEffect(() => {
    void loadTemplates();
  }, []);

  async function loadTemplates() {
    setLoading(true);
    try {
      const items = await api.listPromptTemplates();
      setTemplates(items);
      setForms(formsFromTemplates(items));
      if (items.length > 0 && !items.some((item) => item.code === activeCode)) {
        setActiveCode(items[0].code);
      }
    } catch (error) {
      toast.error((error as Error).message);
    } finally {
      setLoading(false);
    }
  }

  function updateForm(code: string, patch: Partial<PromptForm>) {
    setForms((current) => ({
      ...current,
      [code]: {
        ...(current[code] ?? blankPromptForm()),
        ...patch,
      },
    }));
    setErrorByCode((current) => ({ ...current, [code]: '' }));
  }

  async function saveTemplate(template: PromptTemplateInfo) {
    const form = forms[template.code] ?? formFromTemplate(template);
    if (form.content.trim() === '') {
      setErrorByCode((current) => ({ ...current, [template.code]: '模板内容不能为空' }));
      return;
    }
    setSavingCode(template.code);
    try {
      const updated = await api.updatePromptTemplate(template.code, {
        content: form.content,
      });
      replaceTemplate(updated);
      setForms((current) => ({ ...current, [updated.code]: formFromTemplate(updated) }));
      toast.success('提示词模板已保存');
    } catch (error) {
      const message = (error as Error).message;
      setErrorByCode((current) => ({ ...current, [template.code]: message }));
      toast.error(message);
    } finally {
      setSavingCode('');
    }
  }

  async function restoreTemplate(template: PromptTemplateInfo) {
    setRestoringCode(template.code);
    try {
      const restored = await api.restorePromptTemplate(template.code);
      replaceTemplate(restored);
      setForms((current) => ({ ...current, [restored.code]: formFromTemplate(restored) }));
      setErrorByCode((current) => ({ ...current, [restored.code]: '' }));
      toast.success('已恢复默认提示词');
    } catch (error) {
      toast.error((error as Error).message);
    } finally {
      setRestoringCode('');
    }
  }

  function replaceTemplate(updated: PromptTemplateInfo) {
    setTemplates((current) => current.map((template) => template.code === updated.code ? updated : template));
  }

  if (loading && templates.length === 0) {
    return (
      <section className="grid min-h-0 grid-rows-[auto_minmax(0,1fr)] overflow-hidden rounded-[8px] border border-border bg-card" data-testid="prompts-page">
        <div className="flex min-h-14 items-center justify-between border-b border-border px-4 py-3">
          <Skeleton className="h-5 w-36" />
          <Skeleton className="h-9 w-24" />
        </div>
        <div className="grid gap-4 p-4 lg:grid-cols-[minmax(0,1fr)_320px]">
          <Skeleton className="h-[520px] rounded-[8px]" />
          <Skeleton className="h-[520px] rounded-[8px]" />
        </div>
      </section>
    );
  }

  if (!activeTemplate) {
    return (
      <section className="grid min-h-0 place-items-center rounded-[8px] border border-border bg-card p-4" data-testid="prompts-page">
        <EmptyState
          description="当前注册表没有可管理的系统提示词模板。"
          icon={<ScrollText className="h-5 w-5" />}
          title="暂无提示词模板"
          action={(
            <Button disabled={loading} type="button" variant="soft" onClick={() => void loadTemplates()}>
              <RefreshCw data-icon="inline-start" />
              刷新
            </Button>
          )}
        />
      </section>
    );
  }

  return (
    <section className="grid h-full min-h-0 grid-rows-[auto_minmax(0,1fr)] overflow-hidden rounded-[8px] border border-border bg-card max-[980px]:h-auto" data-testid="prompts-page">
      <div className="flex min-h-14 items-center justify-between gap-3 border-b border-border px-4 py-3 max-[760px]:flex-col max-[760px]:items-stretch">
        <div className="grid gap-1">
          <div className="text-sm font-semibold text-foreground">系统提示词</div>
          <div className="text-xs text-muted-foreground">已注册 {templates.length} 个，首期用于 Git 提交信息生成</div>
        </div>
        <div className="flex flex-wrap items-center gap-2">
          <Badge tone={activeTemplate.content === activeTemplate.defaultContent ? 'neutral' : 'primary'}>
            {activeTemplate.content === activeTemplate.defaultContent ? '默认' : '已编辑'}
          </Badge>
          <IconButton disabled={loading} title="刷新提示词模板" onClick={() => void loadTemplates()}>
            <RefreshCw className={cn('h-4 w-4', loading && 'animate-spin')} />
          </IconButton>
        </div>
      </div>

      <Tabs className="h-full min-h-0 overflow-hidden p-4 max-[980px]:overflow-visible" defaultValue={activeTemplate.code} value={activeTemplate.code} onValueChange={setActiveCode}>
        <div className="flex shrink-0 min-w-0 items-center justify-between gap-3 max-[760px]:flex-col max-[760px]:items-stretch">
          <TabsList className="max-w-full overflow-x-auto" variant="line">
            {templates.map((template) => (
              <TabsTrigger data-testid={`prompt-tab-${template.code}`} key={template.code} value={template.code}>
                <WandSparkles data-icon="inline-start" />
                {promptTemplateTitle(template)}
              </TabsTrigger>
            ))}
          </TabsList>
        </div>

        {templates.map((template) => {
          const form = forms[template.code] ?? formFromTemplate(template);
          const isActiveTemplate = template.code === activeTemplate.code;
          return (
            <TabsContent className="min-h-0 overflow-auto xl:overflow-hidden" key={template.code} value={template.code}>
              {isActiveTemplate ? (
                <PromptTemplateEditor
                  error={errorByCode[template.code]}
                  form={form}
                  restoring={restoringCode === template.code}
                  saving={savingCode === template.code}
                  template={template}
                  onForm={(patch) => updateForm(template.code, patch)}
                  onRestore={() => void restoreTemplate(template)}
                  onSave={() => void saveTemplate(template)}
                />
              ) : null}
            </TabsContent>
          );
        })}
      </Tabs>
    </section>
  );
}

function PromptTemplateEditor({
  error,
  form,
  restoring,
  saving,
  template,
  onForm,
  onRestore,
  onSave,
}: {
  error?: string;
  form: PromptForm;
  restoring: boolean;
  saving: boolean;
  template: PromptTemplateInfo;
  onForm: (patch: Partial<PromptForm>) => void;
  onRestore: () => void;
  onSave: () => void;
}) {
  const [showRestoreConfirm, setShowRestoreConfirm] = useState(false);

  return (
    <div className="grid min-h-0 gap-4 xl:h-full xl:grid-cols-[minmax(0,1fr)_360px] xl:overflow-hidden">
      <div className="grid min-h-[560px] min-w-0 grid-rows-[auto_minmax(0,1fr)_auto] rounded-[8px] border border-border bg-background xl:h-full xl:min-h-0" data-testid="prompt-template-editor">
        <div className="flex items-start justify-between gap-3 border-b border-border p-4 max-[720px]:flex-col">
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-2">
              <h3 className="text-base font-semibold">{promptTemplateTitle(template)}</h3>
              <Badge tone={template.content === template.defaultContent ? 'neutral' : 'primary'}>
                {template.content === template.defaultContent ? '默认' : '已编辑'}
              </Badge>
            </div>
            <p className="mt-1 text-xs leading-5 text-muted-foreground">{promptTemplateDescription(template)}</p>
          </div>
        </div>

        <div className="flex min-h-0 flex-col overflow-hidden p-4">
          {error ? <Alert className="mb-3" tone="danger">{error}</Alert> : null}
          <Field className="min-h-0 flex-1" label="模板内容">
            <PromptCodeEditor
              code={template.code}
              disabled={saving || restoring}
              value={form.content}
              onChange={(content) => onForm({ content })}
              onSave={onSave}
            />
          </Field>
        </div>

        <div className="flex flex-wrap items-center justify-between gap-2 border-t border-border p-4">
          <div className="text-xs text-muted-foreground">
            用途：{promptPurposeLabel(template.purpose)}
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <Button disabled={saving || restoring} type="button" variant="soft" onClick={() => setShowRestoreConfirm(true)}>
              <RotateCcw data-icon="inline-start" />
              {restoring ? '恢复中' : '恢复默认'}
            </Button>
            <Button disabled={saving || restoring} type="button" variant="primary" onClick={onSave}>
              <Save data-icon="inline-start" />
              {saving ? '保存中' : '保存'}
            </Button>
          </div>
        </div>
      </div>

      <ConfirmDialog
        confirmText="恢复默认"
        description="恢复后，当前编辑的内容将被覆盖且无法撤销。确定要继续吗？"
        disabled={restoring}
        open={showRestoreConfirm}
        title="恢复默认提示词"
        onClose={() => setShowRestoreConfirm(false)}
        onConfirm={() => {
          setShowRestoreConfirm(false);
          onRestore();
        }}
      />

      <aside className="grid min-h-[560px] min-w-0 grid-rows-[auto_minmax(0,1fr)] rounded-[8px] border border-border bg-background xl:h-full xl:min-h-0" data-testid="prompt-variable-panel">
        <div className="border-b border-border p-4">
          <div className="text-sm font-semibold">变量</div>
          <div className="mt-1 text-xs text-muted-foreground">{template.variables.length} 个可用变量</div>
        </div>
        <div className="min-h-0 overflow-auto p-4" data-testid="prompt-variable-scroll">
          <div className="grid gap-2" data-testid="prompt-variable-list">
            {template.variables.map((variable) => (
              <div className="rounded-[8px] border border-border p-3" key={variable.name}>
                <div className="flex items-center justify-between gap-2">
                  <code className="rounded bg-muted px-1.5 py-0.5 text-xs">{`{{.${variable.name}}}`}</code>
                  <Badge tone={variable.required ? 'warning' : 'neutral'}>
                    {variable.required ? '必需' : '可选'}
                  </Badge>
                </div>
                <p className="mt-2 text-xs leading-5 text-muted-foreground">{promptVariableDescription(template.code, variable.name, variable.description)}</p>
                <div className="mt-3 grid gap-1.5">
                  <div className="text-[11px] font-medium text-muted-foreground">示例值</div>
                  <pre
                    className="max-h-24 overflow-auto whitespace-pre-wrap break-words rounded-[6px] bg-muted/50 p-2 text-[11px] leading-4 text-muted-foreground"
                    data-testid={`prompt-variable-${variable.name}-sample`}
                  >{promptVariableSampleValue(template.code, variable.name, variable.sampleValue)}</pre>
                </div>
              </div>
            ))}
          </div>
        </div>
      </aside>
    </div>
  );
}

function PromptCodeEditor({
  code,
  disabled,
  value,
  onChange,
  onSave,
}: {
  code: string;
  disabled: boolean;
  value: string;
  onChange: (content: string) => void;
  onSave: () => void;
}) {
  const saveRef = useRef(onSave);
  const disabledRef = useRef(disabled);
  const hostRef = useRef<PromptCodeEditorHost>(null);

  useEffect(() => {
    saveRef.current = onSave;
  }, [onSave]);

  useEffect(() => {
    disabledRef.current = disabled;
  }, [disabled]);

  useEffect(() => () => {
    if (hostRef.current) {
      delete hostRef.current.__agentBoxPromptEditor;
    }
  }, []);

  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      const host = hostRef.current;
      if (!host?.contains(document.activeElement) || disabledRef.current) {
        return;
      }
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 's') {
        event.preventDefault();
        saveRef.current();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  const handleMount = useCallback<OnMount>((editor, monaco) => {
    if (hostRef.current) {
      hostRef.current.__agentBoxPromptEditor = editor;
    }
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
      if (!disabledRef.current) {
        saveRef.current();
      }
    });
  }, []);

  return (
    <div
      ref={hostRef}
      className="h-full min-h-[360px] overflow-hidden rounded-[8px] border border-input bg-background xl:min-h-0"
      data-editor-font-size="14"
      data-editor-language="handlebars"
      data-editor-value={value}
      data-testid="prompt-template-code-editor"
    >
      <Suspense fallback={<PromptEditorSpinner />}>
        <MonacoEditor
          height="100%"
          language="handlebars"
          loading={<PromptEditorSpinner />}
          options={{
            ariaLabel: '模板内容',
            automaticLayout: true,
            fixedOverflowWidgets: true,
            folding: false,
            fontSize: 14,
            lineHeight: 22,
            minimap: { enabled: false },
            overviewRulerBorder: false,
            padding: { top: 12, bottom: 12 },
            quickSuggestions: false,
            renderLineHighlight: 'line',
            scrollBeyondLastLine: false,
            tabSize: 2,
            wordWrap: 'on',
          }}
          path={`john-ai-agentbox://prompt-template/${encodeURIComponent(code)}.gotmpl`}
          theme="vs-dark"
          value={value}
          onChange={(nextValue) => onChange(nextValue ?? '')}
          onMount={handleMount}
        />
      </Suspense>
    </div>
  );
}

function PromptEditorSpinner() {
  return (
    <div className="grid h-full min-h-[360px] place-items-center bg-muted/30">
      <Spinner label="加载编辑器" />
    </div>
  );
}

function formsFromTemplates(templates: PromptTemplateInfo[]) {
  return templates.reduce<Record<string, PromptForm>>((acc, template) => {
    acc[template.code] = formFromTemplate(template);
    return acc;
  }, {});
}

function formFromTemplate(template: PromptTemplateInfo): PromptForm {
  return {
    content: template.content || template.defaultContent,
  };
}

function blankPromptForm(): PromptForm {
  return {
    content: '',
  };
}

function promptPurposeLabel(purpose: string) {
  if (purpose === 'git_commit_message') {
    return 'Git 提交信息生成';
  }
  return purpose;
}

function promptTemplateTitle(template: PromptTemplateInfo) {
  if (template.code === gitCommitPromptCode) {
    return 'Git 提交信息';
  }
  return template.displayName;
}

function promptTemplateDescription(template: PromptTemplateInfo) {
  if (template.code === gitCommitPromptCode) {
    return '根据已暂存、未暂存或未跟踪的仓库变更生成一条简洁的 Git 提交信息。';
  }
  return template.description;
}

function promptVariableDescription(templateCode: string, name: string, fallback: string) {
  if (templateCode !== gitCommitPromptCode) {
    return fallback;
  }
  if (name === 'DiffScope') {
    return '生成提交信息时使用的 Git diff 范围，通常是已暂存或未暂存。';
  }
  if (name === 'Diff') {
    return '用于生成提交信息的 Git diff 内容，或安全的未跟踪文件摘要。';
  }
  if (name === 'TruncatedNotice') {
    return 'diff 被截断时展示的一行提示；完整 diff 可见时为空。';
  }
  return fallback;
}

function promptVariableSampleValue(templateCode: string, name: string, fallback: string) {
  if (templateCode !== gitCommitPromptCode) {
    return fallback.trim() || '暂无示例值';
  }
  if (name === 'DiffScope') {
    return 'staged（已暂存变更）';
  }
  if (name === 'Diff') {
    return 'diff --git a/web/src/App.tsx b/web/src/App.tsx\n+新增提示词管理导航';
  }
  if (name === 'TruncatedNotice') {
    return '- diff 已被截断，请只总结当前可见的变更。';
  }
  return fallback.trim() || '暂无示例值';
}
