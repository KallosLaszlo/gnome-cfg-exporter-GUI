export interface Category {
  id: string;
  name: string;
  desc: string;
}

export interface BackupInfo {
  path: string;
  dirName: string;
  timestamp: string;
  gnomeVersion: string;
  hostname: string;
  username: string;
  sessionType: string;
  size: string;
  fileCount: number;
  description: string;
  categories: string[];
}

export interface BackupResult {
  path: string;
  timestamp: string;
  size: string;
  fileCount: number;
  errors: string[];
}

export interface RestoreResult {
  safetyBackupPath: string;
  restoredCats: string[];
  warnings: string[];
  errors: string[];
}

export interface SystemInfo {
  gnomeVersion: string;
  sessionType: string;
  hostname: string;
  distro: {
    id: string;
    name: string;
    packageManager: string;
  };
}

export interface DepStatus {
  name: string;
  command: string;
  installed: boolean;
  required: boolean;
  desc: string;
}

export interface AppConfig {
  backupDir: string;
  scheduleEnabled: boolean;
  scheduleLevels: Record<string, number>;
  defaultCategories: string[];
  autoDeleteDays: number;
}

export interface ScheduleLevel {
  level: string;
  enabled: boolean;
  keep: number;
}

export interface ScheduleStatus {
  levels: ScheduleLevel[];
}

export interface ProgressEvent {
  percent: number;
  message: string;
}
