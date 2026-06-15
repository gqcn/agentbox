import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export const workspaceRootPath = '/home/agent/workspace';
export const sharedRootPath = '/home/agent/shared';

export function truncateMiddle(value: string, max = 44) {
  if (value.length <= max) {
    return value;
  }
  const keep = Math.max(8, Math.floor((max - 3) / 2));
  return `${value.slice(0, keep)}...${value.slice(-keep)}`;
}

export function formatBytes(value?: number) {
  if (!value) {
    return '0 B';
  }
  const units = ['B', 'KB', 'MB', 'GB'];
  let size = value;
  let unit = 0;
  while (size >= 1024 && unit < units.length - 1) {
    size /= 1024;
    unit += 1;
  }
  return `${size.toFixed(unit === 0 ? 0 : 1)} ${units[unit]}`;
}

export function normalizeWorkspacePath(value: string) {
  const trimmed = value.trim();
  if (!trimmed || trimmed === '/') {
    return workspaceRootPath;
  }
  if (trimmed === workspaceRootPath || trimmed.startsWith(`${workspaceRootPath}/`)) {
    return trimmed;
  }
  if (trimmed === sharedRootPath || trimmed.startsWith(`${sharedRootPath}/`)) {
    return trimmed;
  }
  if (trimmed.startsWith('/')) {
    return trimmed;
  }
  return `${workspaceRootPath}/${trimmed.replace(/^\/+/, '')}`;
}
