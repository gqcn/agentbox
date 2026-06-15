import type {
  ButtonHTMLAttributes,
  ComponentProps,
  FormHTMLAttributes,
  HTMLAttributes,
  InputHTMLAttributes,
  LabelHTMLAttributes,
  MouseEvent,
  OptionHTMLAttributes,
  Ref,
  ReactElement,
  ReactNode,
  SelectHTMLAttributes,
  TextareaHTMLAttributes,
} from 'react';
import { Children, cloneElement, forwardRef, isValidElement, useId, useMemo, useState } from 'react';
import {
  AlertTriangle,
  CheckCircle2,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  ChevronUp,
  ChevronsUpDown,
  Loader2,
  Search,
  X,
} from 'lucide-react';
import {
  type ColumnDef,
  type ColumnSizingState,
  type VisibilityState,
  type Row,
  type SortingState,
  flexRender,
  getCoreRowModel,
  getSortedRowModel,
  useReactTable,
} from '@tanstack/react-table';
import { Panel, PanelGroup, PanelResizeHandle } from 'react-resizable-panels';
import { cn } from '@/lib/utils';
import { Button as ShadcnButton, buttonVariants } from './button';
import { Input as ShadcnInput } from './input';
import { Textarea as ShadcnTextarea } from './textarea';
import { Checkbox as ShadcnCheckbox } from './checkbox';
import { Badge as ShadcnBadge } from './badge';
import {
  Alert as ShadcnAlert,
  AlertAction,
  AlertDescription,
  AlertTitle,
} from './alert';
import {
  Dialog as ShadcnDialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from './dialog';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from './dropdown-menu';
export {
  ContextMenu,
  ContextMenuCheckboxItem,
  ContextMenuContent,
  ContextMenuGroup,
  ContextMenuItem,
  ContextMenuLabel,
  ContextMenuPortal,
  ContextMenuRadioGroup,
  ContextMenuRadioItem,
  ContextMenuSeparator,
  ContextMenuShortcut,
  ContextMenuSub,
  ContextMenuSubContent,
  ContextMenuSubTrigger,
  ContextMenuTrigger,
} from './context-menu';
import { Skeleton as ShadcnSkeleton } from './skeleton';
import { Label as ShadcnLabel } from './label';
import {
  Select as ShadcnSelect,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from './select';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from './table';
import {
  Field as ShadcnField,
  FieldError,
  FieldLabel,
} from './field';
export { Toaster } from './sonner';
export { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from './tooltip';

export type { ColumnDef, Row };
export { buttonVariants };

// ─── Button ──────────────────────────────────────────────────────────────────

type ShadcnButtonProps = ComponentProps<typeof ShadcnButton>;

export type ButtonProps = Omit<ShadcnButtonProps, 'variant'> & {
  variant?: ComponentProps<typeof ShadcnButton>['variant'] | 'primary' | 'soft' | 'danger';
};

function normalizeButtonVariant(variant: ButtonProps['variant']): ComponentProps<typeof ShadcnButton>['variant'] {
  if (variant === 'primary') return 'default';
  if (variant === 'soft') return 'secondary';
  if (variant === 'danger') return 'destructive';
  return variant ?? 'default';
}

export function Button({ className, variant = 'default', size = 'default', ...props }: ButtonProps) {
  return (
    <ShadcnButton
      className={className}
      size={size}
      variant={normalizeButtonVariant(variant)}
      {...props}
    />
  );
}

export function ControlButton({ className, ...props }: ButtonHTMLAttributes<HTMLButtonElement>) {
  return <Button className={className} type="button" variant="ghost" {...props} />;
}

export function IconButton({ className, ...props }: ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <Button
      aria-label={props['aria-label'] ?? props.title}
      className={className}
      size="icon-sm"
      type="button"
      variant="ghost"
      {...props}
    />
  );
}

export function TreeButton({ className, ...props }: ButtonHTMLAttributes<HTMLButtonElement>) {
  return (
    <ControlButton
      className={cn('h-auto min-w-0 justify-start bg-transparent px-0 py-0 text-left font-normal hover:bg-transparent', className)}
      {...props}
    />
  );
}

export function TabButton({
  active,
  className,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & { active?: boolean }) {
  return (
    <Button
      aria-selected={active}
      className={cn(
        'border-transparent hover:border-primary/60 hover:bg-primary/15 hover:text-foreground hover:shadow-sm hover:ring-1 hover:ring-primary/20',
        active && 'border-primary/40 bg-primary/10 text-foreground shadow-xs hover:border-primary/80 hover:bg-primary/30 hover:ring-2 hover:ring-primary/25',
        className,
      )}
      type="button"
      variant={active ? 'outline' : 'ghost'}
      {...props}
    />
  );
}

export function ListButton({
  active,
  className,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & { active?: boolean }) {
  return (
    <Button
      className={cn(
        'h-auto w-full items-start justify-start p-2.5 text-left font-normal whitespace-normal',
        active && 'border-primary/30 bg-primary/10 text-foreground hover:bg-primary/10',
        className,
      )}
      type="button"
      variant="outline"
      {...props}
    />
  );
}

// ─── Form Inputs ─────────────────────────────────────────────────────────────

export const Input = forwardRef<HTMLInputElement, InputHTMLAttributes<HTMLInputElement>>(
  ({ className, ...props }, ref) => (
    <ShadcnInput ref={ref} className={cn('h-9', className)} {...props} />
  ),
);
Input.displayName = 'Input';

export const Textarea = forwardRef<HTMLTextAreaElement, TextareaHTMLAttributes<HTMLTextAreaElement>>(
  ({ className, ...props }, ref) => (
    <ShadcnTextarea ref={ref} className={cn('min-h-24 resize-y', className)} {...props} />
  ),
);
Textarea.displayName = 'Textarea';

export const Select = forwardRef<HTMLButtonElement, Omit<SelectHTMLAttributes<HTMLSelectElement>, 'onChange' | 'value' | 'defaultValue'> & {
  value?: string | number;
  defaultValue?: string | number;
  placeholder?: string;
  onChange?: (event: { target: { value: string }; currentTarget: { value: string } }) => void;
}>(
  ({ className, children, disabled, value, defaultValue, placeholder, onChange, ...props }, ref) => {
    const options = useMemo(() => selectOptionsFromChildren(children), [children]);
    const selectedValue = value == null ? undefined : String(value);
    const initialValue = defaultValue == null ? undefined : String(defaultValue);
    const currentLabel = options.find((option) => option.value === selectedValue)?.label;

    return (
      <ShadcnSelect
        defaultValue={initialValue}
        disabled={disabled}
        value={selectedValue}
        onValueChange={(nextValue) => {
          const stringValue = nextValue ?? '';
          onChange?.({
            target: { value: stringValue },
            currentTarget: { value: stringValue },
          });
        }}
      >
        <SelectTrigger
          ref={ref}
          aria-label={props['aria-label']}
          className={cn('h-9 w-full', className)}
          id={props.id}
          name={props.name}
          size="default"
        >
          {selectedValue != null
            ? <SelectValue>{currentLabel ?? selectedValue}</SelectValue>
            : <SelectValue placeholder={placeholder ?? '请选择'} />
          }
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            {options.map((option) => (
              <SelectItem disabled={option.disabled} key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectGroup>
        </SelectContent>
      </ShadcnSelect>
    );
  },
);
Select.displayName = 'Select';

export function SelectOption(props: OptionHTMLAttributes<HTMLOptionElement>) {
  return <option {...props} />;
}

function selectOptionsFromChildren(children: ReactNode) {
  const options: Array<{ value: string; label: ReactNode; disabled?: boolean }> = [];
  Children.forEach(children, (child) => {
    if (!child || typeof child !== 'object' || !('props' in child)) {
      return;
    }
    const props = child.props as OptionHTMLAttributes<HTMLOptionElement>;
    options.push({
      value: props.value == null ? String(props.children ?? '') : String(props.value),
      label: props.children,
      disabled: props.disabled,
    });
  });
  return options;
}

export function Checkbox({ className, checked, onChange, ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <ShadcnCheckbox
      checked={Boolean(checked)}
      className={className}
      disabled={props.disabled}
      name={props.name}
      value={typeof props.value === 'string' || typeof props.value === 'number' ? String(props.value) : undefined}
      onCheckedChange={(nextChecked) => {
        const syntheticEvent = {
          target: { checked: Boolean(nextChecked) },
          currentTarget: { checked: Boolean(nextChecked) },
        } as unknown as Parameters<NonNullable<typeof onChange>>[0];
        onChange?.(syntheticEvent);
      }}
    />
  );
}

export function CheckboxField({
  children,
  className,
  inputClassName,
  ...props
}: InputHTMLAttributes<HTMLInputElement> & { children: ReactNode; inputClassName?: string }) {
  return (
    <label className={cn('group/field inline-flex cursor-pointer items-center gap-2 text-sm text-foreground', className)}>
      <Checkbox className={inputClassName} {...props} />
      <span>{children}</span>
    </label>
  );
}

export function Switch({ checked, className, ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <label
      className={cn(
        'relative inline-flex h-5 w-9 cursor-pointer items-center rounded-full border transition-colors',
        checked ? 'border-primary bg-primary' : 'border-input bg-muted',
        className,
      )}
    >
      <input className="sr-only" checked={checked} type="checkbox" {...props} />
      <span
        className={cn(
          'size-4 rounded-full bg-background shadow-sm transition-transform',
          checked ? 'translate-x-4' : 'translate-x-0.5',
        )}
      />
    </label>
  );
}

export function Label({ className, ...props }: LabelHTMLAttributes<HTMLLabelElement>) {
  return <ShadcnLabel className={cn('text-xs text-muted-foreground', className)} {...props} />;
}

export function Field({
  label,
  error,
  children,
  className,
}: {
  label: string;
  error?: string;
  children: ReactNode;
  className?: string;
}) {
  const generatedId = useId();
  const child = attachFieldId(children, generatedId);
  const controlId = fieldChildId(child);

  return (
    <ShadcnField className={cn('gap-1.5', className)} data-invalid={Boolean(error) || undefined}>
      <FieldLabel className="text-xs font-semibold text-muted-foreground" htmlFor={controlId}>{label}</FieldLabel>
      {child}
      {error ? <FieldError className="text-xs">{error}</FieldError> : null}
    </ShadcnField>
  );
}

function attachFieldId(children: ReactNode, generatedId: string) {
  if (!isValidElement<{ id?: string; 'aria-label'?: string }>(children)) {
    return children;
  }
  if (children.props.id || children.props['aria-label']) {
    return children;
  }
  return cloneElement(children, { id: generatedId });
}

function fieldChildId(children: ReactNode) {
  if (!isValidElement<{ id?: string }>(children)) {
    return undefined;
  }
  return typeof children.props.id === 'string' ? children.props.id : undefined;
}

export const Form = forwardRef<HTMLFormElement, FormHTMLAttributes<HTMLFormElement>>(
  ({ className, ...props }, ref) => (
    <form ref={ref} className={cn('grid gap-4', className)} {...props} />
  ),
);
Form.displayName = 'Form';

export function FileUploadButton({
  children,
  className,
  disabled,
  multiple,
  accept,
  buttonRef,
  buttonTestId,
  inputTestId,
  title,
  onFiles,
}: {
  children: ReactNode;
  className?: string;
  disabled?: boolean;
  multiple?: boolean;
  accept?: string;
  buttonRef?: Ref<HTMLLabelElement>;
  buttonTestId?: string;
  inputTestId?: string;
  title?: string;
  onFiles: (files: FileList | null) => void;
}) {
  return (
    <label
      ref={buttonRef}
      aria-label={title}
      aria-disabled={disabled}
      className={cn(
        buttonVariants({ variant: 'outline', size: 'default' }),
        'cursor-pointer focus-within:border-ring focus-within:ring-3 focus-within:ring-ring/50',
        disabled && 'pointer-events-none cursor-not-allowed opacity-50',
        className,
      )}
      data-testid={buttonTestId}
      role="button"
      title={title}
    >
      {children}
      <input
        className="sr-only"
        accept={accept}
        aria-hidden="true"
        data-testid={inputTestId}
        disabled={disabled}
        multiple={multiple}
        tabIndex={-1}
        type="file"
        onChange={(event) => onFiles(event.target.files)}
      />
    </label>
  );
}

// ─── Badge ────────────────────────────────────────────────────────────────────

export function Badge({
  className,
  tone = 'neutral',
  ...props
}: ComponentProps<typeof ShadcnBadge> & {
  tone?: 'neutral' | 'success' | 'warning' | 'danger' | 'info' | 'primary';
}) {
  const variant =
    tone === 'danger' ? 'destructive' :
    tone === 'primary' ? 'default' :
    tone === 'neutral' ? 'secondary' :
    'outline';
  return (
    <ShadcnBadge
      className={cn(
        tone === 'success' && 'border-chart-2/30 bg-chart-2/10 text-foreground',
        tone === 'warning' && 'border-chart-4/30 bg-chart-4/10 text-foreground',
        tone === 'info' && 'border-chart-3/30 bg-chart-3/10 text-foreground',
        className,
      )}
      variant={variant}
      {...props}
    />
  );
}

// ─── Layout helpers ───────────────────────────────────────────────────────────

export function Toolbar({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('flex min-h-10 flex-wrap items-center gap-2', className)} {...props} />;
}

export function PanelHeader({
  title,
  description,
  actions,
  className,
}: {
  title: string;
  description?: string;
  actions?: ReactNode;
  className?: string;
}) {
  return (
    <div className={cn('flex min-h-12 items-center justify-between gap-3 border-b px-4 py-3', className)}>
      <div className="min-w-0">
        <h2 className="truncate text-sm font-semibold text-foreground">{title}</h2>
        {description ? <p className="mt-0.5 truncate text-xs text-muted-foreground">{description}</p> : null}
      </div>
      {actions ? <div className="flex shrink-0 items-center gap-1.5">{actions}</div> : null}
    </div>
  );
}

export function EmptyState({
  title,
  description,
  action,
  icon,
  className,
}: {
  title: string;
  description?: string;
  action?: ReactNode;
  icon?: ReactNode;
  className?: string;
}) {
  return (
    <div className={cn('grid min-h-44 place-items-center rounded-lg bg-muted/50 p-8 text-center', className)}>
      <div className="max-w-sm">
        {icon ? (
          <div className="mx-auto mb-4 flex size-12 items-center justify-center rounded-lg bg-muted text-muted-foreground">
            {icon}
          </div>
        ) : null}
        <div className="text-sm font-semibold text-foreground">{title}</div>
        {description ? <p className="mt-2 text-sm leading-6 text-muted-foreground">{description}</p> : null}
        {action ? <div className="mt-4 flex justify-center">{action}</div> : null}
      </div>
    </div>
  );
}

export function Spinner({ label = '加载中' }: { label?: string }) {
  return (
    <span className="inline-flex items-center gap-2 text-sm text-muted-foreground">
      <Loader2 className="size-4 animate-spin" />
      {label}
    </span>
  );
}

export function Skeleton({ className }: { className?: string }) {
  return <ShadcnSkeleton className={className} />;
}

export function Alert({
  tone = 'info',
  children,
  className,
  onDismiss,
}: {
  tone?: 'info' | 'success' | 'warning' | 'danger';
  children: ReactNode;
  className?: string;
  onDismiss?: () => void;
}) {
  const Icon = tone === 'success' ? CheckCircle2 : AlertTriangle;
  return (
    <ShadcnAlert
      className={cn(
        'grid-cols-[auto_minmax(0,1fr)_auto] items-start gap-x-2.5 gap-y-0 rounded-lg px-3 py-2.5 shadow-xs',
        'has-data-[slot=alert-action]:pr-3',
        tone === 'success' && 'border-chart-2/30 bg-chart-2/5 text-foreground',
        tone === 'warning' && 'border-chart-4/30 bg-chart-4/5 text-foreground',
        tone === 'info' && 'border-chart-3/30 bg-chart-3/5 text-foreground',
        tone === 'danger' && 'border-destructive/25 bg-destructive/5 text-foreground',
        tone === 'danger' && '*:data-[slot=alert-description]:text-destructive/90 *:[svg]:text-destructive',
        className,
      )}
      variant="default"
    >
      <span className="mt-0.5 grid size-6 place-items-center rounded-md bg-background/80 ring-1 ring-current/10">
        <Icon className="size-4" />
      </span>
      <div className="min-w-0">
        <AlertTitle className={cn('text-sm font-semibold leading-5', tone === 'danger' && 'text-destructive')}>
          {toneLabel(tone)}
        </AlertTitle>
        <AlertDescription className="mt-0.5 text-sm leading-5 text-pretty">
          {children}
        </AlertDescription>
      </div>
      {onDismiss ? (
        <AlertAction className="static col-start-3 row-span-2 row-start-1 self-start">
          <button
            aria-label="关闭提示"
            type="button"
            className={cn(
              'grid size-7 place-items-center rounded-md text-muted-foreground transition-colors',
              'hover:bg-background hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring/50 focus-visible:outline-none',
              tone === 'danger' && 'text-destructive/70 hover:text-destructive',
            )}
            onClick={onDismiss}
          >
            <X className="size-4" />
          </button>
        </AlertAction>
      ) : null}
    </ShadcnAlert>
  );
}

function toneLabel(tone: 'info' | 'success' | 'warning' | 'danger') {
  if (tone === 'success') return '成功';
  if (tone === 'warning') return '注意';
  if (tone === 'danger') return '错误';
  return '提示';
}

// ─── Dialog ──────────────────────────────────────────────────────────────────

export function Dialog({
  open,
  title,
  children,
  footer,
  onClose,
  className,
}: {
  open: boolean;
  title: string;
  children: ReactNode;
  footer?: ReactNode;
  onClose: () => void;
  className?: string;
}) {
  return (
    <ShadcnDialog open={open} onOpenChange={(isOpen) => { if (!isOpen) onClose(); }}>
      <DialogContent className={cn('grid max-h-[90dvh] max-w-2xl grid-rows-[auto_minmax(0,1fr)_auto] gap-0 overflow-hidden p-0 sm:max-w-2xl', className)}>
        <DialogHeader className="border-b px-5 py-4">
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription className="sr-only">{title}</DialogDescription>
        </DialogHeader>
        <div className="min-h-0 overflow-auto px-5 py-4">{children}</div>
        {footer ? <DialogFooter className="border-t px-5 py-4">{footer}</DialogFooter> : null}
      </DialogContent>
    </ShadcnDialog>
  );
}

export function ConfirmDialog({
  open,
  title,
  description,
  confirmText = '确认',
  cancelText = '取消',
  danger,
  disabled,
  onConfirm,
  onClose,
}: {
  open: boolean;
  title: string;
  description: ReactNode;
  confirmText?: string;
  cancelText?: string;
  danger?: boolean;
  disabled?: boolean;
  onConfirm: () => void;
  onClose: () => void;
}) {
  return (
    <Dialog
      className="sm:max-w-md"
      footer={
        <>
          <DialogClose render={<Button type="button" variant="soft" onClick={onClose} />}>
            {cancelText}
          </DialogClose>
          <Button disabled={disabled} type="button" variant={danger ? 'danger' : 'primary'} onClick={onConfirm}>
            {confirmText}
          </Button>
        </>
      }
      open={open}
      title={title}
      onClose={onClose}
    >
      <div className="text-sm leading-6 text-muted-foreground">{description}</div>
    </Dialog>
  );
}

// ─── Search / Pagination ─────────────────────────────────────────────────────

export function SearchField({ className, ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <div className={cn('relative min-w-0', className)}>
      <Search className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground" />
      <Input className="w-full pl-9" {...props} />
    </div>
  );
}

export function Pagination({
  page,
  pages,
  label,
  onPrev,
  onNext,
}: {
  page: number;
  pages: number;
  label: string;
  onPrev: () => void;
  onNext: () => void;
}) {
  return (
    <div className="flex min-h-11 items-center justify-between gap-3 border-t px-4 py-2 text-sm text-muted-foreground">
      <span>{label}</span>
      <div className="flex items-center gap-1">
        <IconButton disabled={page <= 1} title="上一页" onClick={onPrev}>
          <ChevronLeft />
        </IconButton>
        <span className="min-w-16 text-center tabular-nums">{page} / {pages}</span>
        <IconButton disabled={page >= pages} title="下一页" onClick={onNext}>
          <ChevronRight />
        </IconButton>
      </div>
    </div>
  );
}

// ─── Chat ─────────────────────────────────────────────────────────────────────

export type ChatUiExecutionEvent = {
  id: string;
  kind: 'status' | 'tool' | 'command' | 'file' | 'test' | 'thinking';
  title: string;
  detail?: string;
  status: 'running' | 'complete' | 'error';
  createdAt: string;
};

export type ChatUiMessage = {
  id: number;
  role: string;
  content: string;
  status: string;
  executionEvents?: ChatUiExecutionEvent[];
};

export const ChatThread = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ children, className, ...props }, ref) => (
    <div ref={ref} className={cn('flex min-h-0 flex-col gap-3 overflow-auto p-4', className)} {...props}>
      {children}
    </div>
  ),
);
ChatThread.displayName = 'ChatThread';

export function ChatMessage({
  message,
  label,
  statusLabel,
  children,
}: {
  message: ChatUiMessage;
  label: string;
  statusLabel: string;
  children: ReactNode;
}) {
  return (
    <article
      className={cn(
        'grid w-fit max-w-[min(760px,88%)] gap-2 rounded-xl px-4 py-3 text-sm',
        message.role === 'user' && 'ml-auto bg-primary text-primary-foreground',
        message.role === 'assistant' && 'mr-auto border-l-[3px] border-l-primary bg-muted/50',
        message.role === 'terminal' && 'w-full max-w-full self-stretch rounded-lg bg-zinc-950 font-mono text-xs text-zinc-100',
        message.role === 'error' && 'border border-destructive/30 bg-destructive/10',
        !['user', 'assistant', 'terminal', 'error'].includes(message.role) && 'border bg-card',
      )}
    >
      <header
        className={cn(
          'flex justify-between gap-3 text-xs',
          message.role === 'user' ? 'text-primary-foreground/70' : 'text-muted-foreground',
        )}
      >
        <strong>{label}</strong>
        <span>{statusLabel}</span>
      </header>
      {message.role === 'assistant' && message.executionEvents?.length ? (
        <ChatExecutionPanel events={message.executionEvents} processing={message.status === 'streaming'} />
      ) : null}
      {children}
    </article>
  );
}

export function ChatExecutionPanel({ events, processing }: { events: ChatUiExecutionEvent[]; processing: boolean }) {
  const [expanded, setExpanded] = useState(true);
  const visibleEvents = events.slice(-24);
  const running = processing && visibleEvents.some((event) => event.status === 'running');
  const logText = visibleEvents.map(formatExecutionLogLine).filter(Boolean).join('\n');
  const toggleTitle = expanded ? '收起执行过程' : '展开执行过程';
  return (
    <div className="grid gap-2 rounded-lg border bg-card p-2" data-testid="chat-execution-panel">
      <div className="flex items-center justify-between gap-2 text-[11px] font-medium text-muted-foreground">
        <span className="flex min-w-0 items-center gap-1.5">
          {running ? <Loader2 aria-label="执行中" className="size-3 animate-spin" data-testid="chat-execution-loading" /> : null}
          <span>执行过程</span>
        </span>
        <div className="flex min-w-0 items-center gap-1.5">
          {events.length > visibleEvents.length ? <span className="truncate">最近 {visibleEvents.length} 条</span> : null}
          {!expanded ? <span className="truncate">已收起</span> : null}
          <IconButton
            aria-expanded={expanded}
            aria-label={toggleTitle}
            className="size-6"
            title={toggleTitle}
            onClick={() => setExpanded((value) => !value)}
          >
            <ChevronDown className={cn('size-3.5 transition-transform', !expanded && '-rotate-90')} />
          </IconButton>
        </div>
      </div>
      {expanded ? (
        <pre className="max-h-48 overflow-auto whitespace-pre-wrap break-words rounded-md bg-muted/70 px-2 py-1.5 font-mono text-[11px] leading-5 text-foreground" data-testid="chat-execution-log">
          <code>{logText}</code>
        </pre>
      ) : null}
    </div>
  );
}

function formatExecutionLogLine(event: ChatUiExecutionEvent) {
  const title = event.title.trim();
  const detail = (event.detail ?? '').trim();
  if (title && detail && title !== detail) {
    return `${title}: ${detail}`;
  }
  return detail || title;
}

export function ChatComposer({ children, className, ...props }: FormHTMLAttributes<HTMLFormElement>) {
  return (
    <Form
      className={cn('grid grid-cols-[minmax(0,1fr)_auto] gap-3 border-t p-3 max-[640px]:grid-cols-1', className)}
      {...props}
    >
      {children}
    </Form>
  );
}

// ─── DataTable ────────────────────────────────────────────────────────────────

export function DataTable<T>({
  columns,
  data,
  className,
  getRowId,
  onRowClick,
  selectedRowId,
  emptyText = '暂无数据',
  columnVisibility,
}: {
  columns: ColumnDef<T>[];
  data: T[];
  className?: string;
  getRowId?: (row: T, index: number) => string;
  onRowClick?: (row: Row<T>) => void;
  selectedRowId?: string;
  emptyText?: string;
  columnVisibility?: VisibilityState;
}) {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnSizing, setColumnSizing] = useState<ColumnSizingState>({});

  const table = useReactTable({
    data,
    columns,
    state: { sorting, columnSizing, columnVisibility },
    onSortingChange: setSorting,
    onColumnSizingChange: setColumnSizing,
    columnResizeMode: 'onChange',
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getRowId,
  });

  return (
    <div className={cn('min-h-0 overflow-auto', className)}>
      <Table className="min-w-full text-left">
        <TableHeader className="sticky top-0 z-10">
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow key={headerGroup.id}>
              {headerGroup.headers.map((header) => {
                const canSort = header.column.getCanSort();
                const sorted = header.column.getIsSorted();
                return (
                  <TableHead
                    key={header.id}
                    className={cn(
                      'bg-muted px-2 py-2 text-xs font-semibold text-muted-foreground sm:px-4 sm:py-3',
                      canSort && 'cursor-pointer select-none transition-colors hover:text-foreground',
                      header.column.id === 'actions' && 'text-center',
                    )}
                    style={{ width: header.column.getSize() }}
                    onClick={canSort ? header.column.getToggleSortingHandler() : undefined}
                  >
                    <span className={cn('inline-flex items-center gap-1', header.column.id === 'actions' && 'w-full justify-center')}>
                      {flexRender(header.column.columnDef.header, header.getContext())}
                      {canSort && (
                        sorted === 'asc' ? <ChevronUp className="size-3" /> :
                        sorted === 'desc' ? <ChevronDown className="size-3" /> :
                        <ChevronsUpDown className="size-3 opacity-30" />
                      )}
                    </span>
                  </TableHead>
                );
              })}
            </TableRow>
          ))}
        </TableHeader>
        <TableBody>
          {table.getRowModel().rows.length === 0 ? (
            <TableRow>
              <TableCell
                className="h-48 text-center align-middle text-sm text-muted-foreground"
                colSpan={columns.length}
              >
                {emptyText}
              </TableCell>
            </TableRow>
          ) : null}
          {table.getRowModel().rows.map((row) => (
            <TableRow
              key={row.id}
              className={cn(
                onRowClick && 'cursor-pointer',
                row.id === selectedRowId && 'bg-primary/10 hover:bg-primary/10',
              )}
              onClick={() => onRowClick?.(row)}
            >
              {row.getVisibleCells().map((cell) => (
                <TableCell key={cell.id} className={cn('px-2 py-2 align-middle sm:px-4 sm:py-3', cell.column.id === 'actions' && 'text-center')} style={{ width: cell.column.getSize() }}>
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

// ─── Resizable panels ─────────────────────────────────────────────────────────

export function ResizablePanelGroup(props: ComponentProps<typeof PanelGroup>) {
  return <PanelGroup {...props} />;
}

export function ResizablePanel(props: ComponentProps<typeof Panel>) {
  return <Panel {...props} />;
}

export function ResizeHandle({ className, ...props }: ComponentProps<typeof PanelResizeHandle>) {
  return (
    <PanelResizeHandle
      className={cn(
        'relative flex items-center justify-center bg-border transition-colors hover:bg-primary active:bg-primary',
        'data-[panel-group-direction=horizontal]:w-px data-[panel-group-direction=vertical]:h-px',
        className,
      )}
      {...props}
    />
  );
}

// ─── Dropdown ─────────────────────────────────────────────────────────────────

export function Dropdown({
  trigger,
  children,
  className,
}: {
  trigger: ReactNode;
  children: ReactNode;
  className?: string;
}) {
  const dropdownTrigger = isValidElement(trigger)
    ? cloneElement(trigger as ReactElement<{ onClick?: (event: MouseEvent<HTMLElement>) => void }>, {
        onClick: (event: MouseEvent<HTMLElement>) => {
          event.stopPropagation();
          (trigger as ReactElement<{ onClick?: (event: MouseEvent<HTMLElement>) => void }>).props.onClick?.(event);
        },
      })
    : trigger;

  return (
    <DropdownMenu>
      <DropdownMenuTrigger render={dropdownTrigger as ReactElement<object>} />
      <DropdownMenuContent align="end" sideOffset={6} className={className}>
        <DropdownMenuGroup>{children}</DropdownMenuGroup>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export function DropdownItem({
  children,
  danger,
  onClick,
  className,
}: {
  children: ReactNode;
  danger?: boolean;
  onClick?: () => void;
  className?: string;
}) {
  return (
    <DropdownMenuItem
      className={className}
      variant={danger ? 'destructive' : 'default'}
      onClick={onClick}
    >
      {children}
    </DropdownMenuItem>
  );
}
