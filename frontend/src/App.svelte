<script lang="ts">
  import { onMount } from 'svelte';
  import Toolbar from './components/Toolbar.svelte';
  import BackupList from './components/BackupList.svelte';
  import Statusbar from './components/Statusbar.svelte';
  import BackupDialog from './components/BackupDialog.svelte';
  import RestoreDialog from './components/RestoreDialog.svelte';
  import SettingsDialog from './components/SettingsDialog.svelte';
  import SetupWizard from './components/SetupWizard.svelte';
  import ProgressOverlay from './components/ProgressOverlay.svelte';
  import {
    backups, categories, systemInfo, appConfig, freeSpace,
    showSettings, showSetup, progress, progressVisible
  } from './stores';
  import type { ProgressEvent } from './types';
  import { GetBackups, GetCategories, GetSystemInfo, GetConfig, ConfigExists, GetFreeSpace } from '../wailsjs/go/main/App';
  import { EventsOn } from '../wailsjs/runtime/runtime';

  let showBackupDialog = false;
  let showRestoreDialog = false;
  let loading = true;
  let loadingMsg = 'Initializing...';

  onMount(async () => {
    EventsOn('backup:progress', (data: ProgressEvent) => {
      progress.set(data);
    });
    EventsOn('restore:progress', (data: ProgressEvent) => {
      progress.set(data);
    });

    const exists = await ConfigExists();
    if (!exists) {
      loading = false;
      showSetup.set(true);
      return;
    }
    await loadData();
  });

  async function loadData() {
    loading = true;
    loadingMsg = 'Loading configuration...';

    const [cats, cfg] = await Promise.all([
      GetCategories(),
      GetConfig(),
    ]);
    categories.set(cats);
    appConfig.set(cfg);

    loadingMsg = 'Scanning backups...';
    const [sys, bkps, space] = await Promise.all([
      GetSystemInfo(),
      GetBackups(),
      GetFreeSpace(),
    ]);
    systemInfo.set(sys);
    backups.set(bkps || []);
    freeSpace.set(space);

    loading = false;
  }

  function handleSetupDone() {
    showSetup.set(false);
    loadData();
  }
</script>

{#if $showSetup}
  <SetupWizard on:done={handleSetupDone} />
{:else if loading}
  <div class="flex flex-col items-center justify-center h-full bg-base-300 gap-4">
    <div class="flex flex-col items-center gap-3">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12 text-primary animate-pulse" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4"/>
      </svg>
      <h2 class="text-lg font-semibold text-base-content">GNOME Config Exporter</h2>
    </div>
    <div class="flex flex-col items-center gap-2">
      <span class="loading loading-spinner loading-md text-primary"></span>
      <p class="text-sm text-base-content/60">{loadingMsg}</p>
    </div>
  </div>
{:else}
  <div class="flex flex-col h-full">
    <Toolbar
      on:backup={() => showBackupDialog = true}
      on:restore={() => showRestoreDialog = true}
      on:settings={() => showSettings.set(true)}
      on:refresh={loadData}
    />
    <BackupList />
    <Statusbar />
  </div>

  {#if showBackupDialog}
    <BackupDialog on:close={() => showBackupDialog = false} on:done={loadData} />
  {/if}

  {#if showRestoreDialog}
    <RestoreDialog on:close={() => showRestoreDialog = false} on:done={loadData} />
  {/if}

  {#if $showSettings}
    <SettingsDialog on:close={() => showSettings.set(false)} on:saved={loadData} />
  {/if}

  {#if $progressVisible}
    <ProgressOverlay />
  {/if}
{/if}
