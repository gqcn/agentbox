import { useEffect, useMemo, useRef } from 'react';
import type { ReactNode } from 'react';
import { toast } from 'sonner';
import { Button } from '@/components/ui';
import { cn } from '@/lib/utils';

export type WorkspaceContextMenuPosition = {
  x: number;
  y: number;
};

export type WorkspaceContextMenuAction = {
  kind?: 'item';
  id: string;
  label: string;
  icon?: ReactNode;
  shortcut?: string;
  danger?: boolean;
  disabled?: boolean;
  hidden?: boolean;
  testId?: string;
  onSelect: () => void | Promise<void>;
};

export type WorkspaceContextMenuSeparator = {
  kind: 'separator';
  id: string;
  hidden?: boolean;
};

export type WorkspaceContextMenuEntry = WorkspaceContextMenuAction | WorkspaceContextMenuSeparator;

type Props = {
  position: WorkspaceContextMenuPosition | null;
  items: WorkspaceContextMenuEntry[];
  label: string;
  testId: string;
  onClose: () => void;
};

const viewportPadding = 8;
const menuWidth = 232;
const menuMaxHeight = 360;

export function WorkspaceContextMenu({ position, items, label, testId, onClose }: Props) {
  const menuRef = useRef<HTMLDivElement | null>(null);
  const visibleItems = useMemo(() => items.filter((item) => !item.hidden), [items]);

  useEffect(() => {
    if (!position) {
      return;
    }

    function handlePointerDown(event: PointerEvent) {
      if (menuRef.current?.contains(event.target as Node)) {
        return;
      }
      onClose();
    }

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        onClose();
      }
    }

    window.addEventListener('pointerdown', handlePointerDown);
    window.addEventListener('keydown', handleKeyDown);
    window.addEventListener('resize', onClose);
    window.addEventListener('scroll', onClose, true);
    return () => {
      window.removeEventListener('pointerdown', handlePointerDown);
      window.removeEventListener('keydown', handleKeyDown);
      window.removeEventListener('resize', onClose);
      window.removeEventListener('scroll', onClose, true);
    };
  }, [onClose, position]);

  if (!position || visibleItems.length === 0) {
    return null;
  }

  const availableWidth = Math.max(0, window.innerWidth - viewportPadding * 2);
  const availableHeight = Math.max(0, window.innerHeight - viewportPadding * 2);
  const effectiveWidth = Math.min(menuWidth, availableWidth);
  const effectiveMaxHeight = Math.min(menuMaxHeight, availableHeight);
  const minLeft = Math.min(viewportPadding, Math.max(0, window.innerWidth - effectiveWidth));
  const maxLeft = Math.max(minLeft, window.innerWidth - effectiveWidth - viewportPadding);
  const minTop = Math.min(viewportPadding, Math.max(0, window.innerHeight - effectiveMaxHeight));
  const maxTop = Math.max(minTop, window.innerHeight - effectiveMaxHeight - viewportPadding);
  const left = Math.min(Math.max(minLeft, position.x), maxLeft);
  const top = Math.min(Math.max(minTop, position.y), maxTop);

  return (
    <div
      ref={menuRef}
      aria-label={label}
      className="fixed z-50 grid min-w-56 max-w-[calc(100vw-16px)] justify-items-stretch overflow-y-auto rounded-[6px] border border-border bg-popover p-1 text-left text-sm text-popover-foreground shadow-lg"
      data-testid={testId}
      role="menu"
      style={{ left, maxHeight: effectiveMaxHeight, top, width: menuWidth }}
      onClick={(event) => event.stopPropagation()}
      onContextMenu={(event) => event.preventDefault()}
    >
      {visibleItems.map((item) => {
        if (item.kind === 'separator') {
          return <div key={item.id} className="my-1 h-px bg-border" role="separator" />;
        }
        return (
          <Button
            key={item.id}
            className={cn(
              'flex h-8 w-full min-w-0 justify-start gap-2 rounded-[4px] px-2 text-left text-sm hover:bg-accent disabled:pointer-events-none disabled:opacity-45 [&_svg]:h-4 [&_svg]:w-4 [&_svg]:shrink-0',
              item.danger && 'text-destructive hover:bg-destructive/10',
            )}
            data-testid={item.testId}
            disabled={item.disabled}
            role="menuitem"
            type="button"
            variant="ghost"
            onClick={() => {
              if (item.disabled) {
                return;
              }
              onClose();
              void item.onSelect();
            }}
          >
            {item.icon}
            <span className="truncate">{item.label}</span>
            {item.shortcut ? <span className="ml-auto text-xs text-muted-foreground">{item.shortcut}</span> : null}
          </Button>
        );
      })}
    </div>
  );
}

export async function copyWorkspaceText(value: string, successMessage: string) {
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(value);
    } else {
      copyWithTextarea(value);
    }
    toast.success(successMessage);
  } catch (error) {
    toast.error(`复制失败：${(error as Error).message}`);
  }
}

function copyWithTextarea(value: string) {
  const textarea = document.createElement('textarea');
  textarea.value = value;
  textarea.style.position = 'fixed';
  textarea.style.left = '-9999px';
  textarea.style.top = '0';
  document.body.appendChild(textarea);
  textarea.focus();
  textarea.select();
  const copied = document.execCommand('copy');
  document.body.removeChild(textarea);
  if (!copied) {
    throw new Error('当前浏览器不允许访问剪贴板');
  }
}
