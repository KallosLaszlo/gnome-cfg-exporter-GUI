import { writable, derived } from 'svelte/store';
import type { BackupInfo, Category, SystemInfo, AppConfig, ProgressEvent } from './types';

export const backups = writable<BackupInfo[]>([]);
export const categories = writable<Category[]>([]);
export const systemInfo = writable<SystemInfo | null>(null);
export const appConfig = writable<AppConfig | null>(null);
export const selectedBackups = writable<Set<string>>(new Set());
export const freeSpace = writable<string>('');

export const progress = writable<ProgressEvent | null>(null);
export const progressVisible = writable<boolean>(false);

export const showSettings = writable<boolean>(false);
export const showSetup = writable<boolean>(false);

export const backupCount = derived(backups, $b => $b.length);

export function toggleSelect(path: string) {
  selectedBackups.update(s => {
    const next = new Set(s);
    if (next.has(path)) next.delete(path);
    else next.add(path);
    return next;
  });
}

export function clearSelection() {
  selectedBackups.set(new Set());
}
