<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { SaveConfig, CheckDependencies, GetSystemInfo, GetConfig } from '../../wailsjs/go/main/App';
  import type { DepStatus, SystemInfo } from '../types';

  const dispatch = createEventDispatcher();

  let step = 0;
  let backupDir = '';
  let deps: DepStatus[] = [];
  let sysInfo: SystemInfo | null = null;
  let hasCritical = false;

  onMount(async () => {
    const cfg = await GetConfig();
    backupDir = cfg.backupDir;
    sysInfo = await GetSystemInfo();
    deps = await CheckDependencies();
    hasCritical = deps.some(d => d.required && !d.installed);
  });

  async function finish() {
    await SaveConfig({
      backupDir,
      scheduleEnabled: false,
      scheduleLevels: {},
      defaultCategories: [],
      autoDeleteDays: 0,
    });
    dispatch('done');
  }
</script>

<div class="flex flex-col items-center justify-center h-full bg-base-300 p-8">
  <div class="card bg-base-100 shadow-xl w-full max-w-md">
    <div class="card-body">
      {#if step === 0}
        <h2 class="card-title">Welcome!</h2>
        <p class="text-sm text-base-content/70">GNOME Config Exporter backs up and restores your GNOME desktop settings.</p>
        {#if sysInfo}
          <div class="bg-base-200 rounded p-3 text-sm mt-2">
            <p><strong>Distro:</strong> {sysInfo.distro?.name || sysInfo.distro?.id}</p>
            <p><strong>GNOME:</strong> {sysInfo.gnomeVersion}</p>
            <p><strong>Session:</strong> {sysInfo.sessionType}</p>
          </div>
        {/if}
        <div class="card-actions justify-end mt-4">
          <button class="btn btn-sm btn-primary" on:click={() => step = 1}>Next</button>
        </div>

      {:else if step === 1}
        <h2 class="card-title">Dependencies</h2>
        <div class="space-y-1 mt-2">
          {#each deps as d}
            <div class="flex items-center gap-2 text-sm">
              {#if d.installed}
                <span class="text-success">&#10003;</span>
              {:else if d.required}
                <span class="text-error">&#10007;</span>
              {:else}
                <span class="text-warning">&#9888;</span>
              {/if}
              <span class="font-mono">{d.command}</span>
              <span class="text-base-content/60 text-xs">— {d.desc}</span>
            </div>
          {/each}
        </div>
        {#if hasCritical}
          <div class="alert alert-error text-sm mt-2">Required dependencies are missing. Please install them first.</div>
        {/if}
        <div class="card-actions justify-between mt-4">
          <button class="btn btn-sm btn-ghost" on:click={() => step = 0}>Back</button>
          <button class="btn btn-sm btn-primary" disabled={hasCritical} on:click={() => step = 2}>Next</button>
        </div>

      {:else if step === 2}
        <h2 class="card-title">Backup Location</h2>
        <div class="form-control mt-2">
          <label class="label pb-1"><span class="label-text">Where should backups be stored?</span></label>
          <input type="text" class="input input-sm input-bordered" bind:value={backupDir} />
        </div>
        <div class="card-actions justify-between mt-4">
          <button class="btn btn-sm btn-ghost" on:click={() => step = 1}>Back</button>
          <button class="btn btn-sm btn-primary" on:click={finish}>Finish Setup</button>
        </div>
      {/if}
    </div>
  </div>
</div>
