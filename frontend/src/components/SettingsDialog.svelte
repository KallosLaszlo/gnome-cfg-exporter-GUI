<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { appConfig } from '../stores';
  import type { AppConfig, ScheduleLevel } from '../types';
  import { SaveConfig, GetScheduleStatus, UpdateSchedule, CheckDependencies } from '../../wailsjs/go/main/App';
  import type { DepStatus } from '../types';

  const dispatch = createEventDispatcher();

  let cfg: AppConfig | null = null;
  let schedLevels: ScheduleLevel[] = [];
  let deps: DepStatus[] = [];
  let tab: 'general' | 'schedule' | 'deps' = 'general';

  onMount(async () => {
    cfg = { ...$appConfig! };
    const status = await GetScheduleStatus();
    schedLevels = status.levels || [
      { level: 'hourly', enabled: false, keep: 6 },
      { level: 'daily', enabled: false, keep: 5 },
      { level: 'weekly', enabled: false, keep: 3 },
      { level: 'monthly', enabled: false, keep: 2 },
    ];
    deps = await CheckDependencies();
  });

  async function save() {
    if (!cfg) return;
    await SaveConfig(cfg);
    await UpdateSchedule(schedLevels);
    dispatch('saved');
    dispatch('close');
  }
</script>

<div class="modal modal-open">
  <div class="modal-box max-w-lg">
    <h3 class="font-bold text-lg mb-3">Settings</h3>

    <div class="tabs tabs-boxed mb-3">
      <button class="tab tab-sm" class:tab-active={tab === 'general'} on:click={() => tab = 'general'}>General</button>
      <button class="tab tab-sm" class:tab-active={tab === 'schedule'} on:click={() => tab = 'schedule'}>Schedule</button>
      <button class="tab tab-sm" class:tab-active={tab === 'deps'} on:click={() => tab = 'deps'}>Dependencies</button>
    </div>

    {#if cfg}
      {#if tab === 'general'}
        <div class="form-control mb-2">
          <label class="label pb-1"><span class="label-text">Backup directory</span></label>
          <input type="text" class="input input-sm input-bordered" bind:value={cfg.backupDir} />
        </div>
        <div class="form-control mb-2">
          <label class="label pb-1"><span class="label-text">Auto-delete backups older than (days, 0=never)</span></label>
          <input type="number" class="input input-sm input-bordered w-24" bind:value={cfg.autoDeleteDays} min="0" />
        </div>
      {/if}

      {#if tab === 'schedule'}
        <div class="space-y-2">
          {#each schedLevels as lvl, i}
            <div class="flex items-center gap-3 bg-base-200 rounded px-3 py-2">
              <input type="checkbox" class="toggle toggle-sm toggle-primary" bind:checked={schedLevels[i].enabled} />
              <span class="capitalize w-20">{lvl.level}</span>
              <span class="text-xs text-base-content/60">Keep</span>
              <input type="number" class="input input-xs input-bordered w-16" bind:value={schedLevels[i].keep} min="1" max="100" />
            </div>
          {/each}
        </div>
      {/if}

      {#if tab === 'deps'}
        <div class="space-y-1">
          {#each deps as d}
            <div class="flex items-center gap-2 text-sm px-2 py-1">
              {#if d.installed}
                <span class="badge badge-success badge-xs">OK</span>
              {:else if d.required}
                <span class="badge badge-error badge-xs">MISSING</span>
              {:else}
                <span class="badge badge-warning badge-xs">OPT</span>
              {/if}
              <span class="font-mono">{d.command}</span>
              <span class="text-base-content/60">— {d.desc}</span>
            </div>
          {/each}
        </div>
      {/if}
    {/if}

    <div class="modal-action">
      <button class="btn btn-sm btn-ghost" on:click={() => dispatch('close')}>Cancel</button>
      <button class="btn btn-sm btn-primary" on:click={save}>Save</button>
    </div>
  </div>
  <div class="modal-backdrop" on:click={() => dispatch('close')} on:keydown={() => {}}></div>
</div>
