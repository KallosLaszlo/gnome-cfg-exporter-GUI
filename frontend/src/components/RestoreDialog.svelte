<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { categories, selectedBackups, progress, progressVisible } from '../stores';
  import type { Category } from '../types';
  import { RestoreBackup, GetRestoreWarnings } from '../../wailsjs/go/main/App';

  const dispatch = createEventDispatcher();

  let allCats: Category[] = [];
  let selectedCats: Set<string> = new Set();
  let warnings: string[] = [];
  let running = false;
  let resultMsg = '';
  let confirmed = false;

  $: backupPath = Array.from($selectedBackups)[0] || '';

  onMount(async () => {
    allCats = $categories;
    allCats.forEach(c => selectedCats.add(c.id));
    selectedCats = selectedCats;
    if (backupPath) {
      warnings = await GetRestoreWarnings(backupPath);
    }
  });

  function toggleCat(id: string) {
    if (selectedCats.has(id)) selectedCats.delete(id);
    else selectedCats.add(id);
    selectedCats = selectedCats;
  }

  async function startRestore() {
    if (selectedCats.size === 0 || !backupPath) return;
    running = true;
    progressVisible.set(true);
    try {
      const result = await RestoreBackup(backupPath, Array.from(selectedCats));
      resultMsg = `Restored ${result.restoredCats?.length || 0} categories.`;
      if (result.safetyBackupPath) {
        resultMsg += `\nSafety backup: ${result.safetyBackupPath}`;
      }
      if (result.warnings?.length) {
        resultMsg += '\nWarnings: ' + result.warnings.join('; ');
      }
      if (result.errors?.length) {
        resultMsg += '\nErrors: ' + result.errors.join('; ');
      }
    } catch (e) {
      resultMsg = 'Error: ' + e;
    } finally {
      running = false;
      progressVisible.set(false);
      progress.set(null);
      dispatch('done');
    }
  }
</script>

<div class="modal modal-open">
  <div class="modal-box max-w-xl">
    <h3 class="font-bold text-lg mb-3">Restore Backup</h3>

    {#if resultMsg}
      <div class="alert alert-info mb-3 text-sm whitespace-pre-line">{resultMsg}</div>
      <div class="modal-action">
        <button class="btn btn-sm" on:click={() => dispatch('close')}>Close</button>
      </div>
    {:else}
      {#if warnings.length > 0}
        <div class="alert alert-warning mb-3 text-sm">
          <div>
            <strong>Warnings:</strong>
            <ul class="list-disc ml-4">
              {#each warnings as w}
                <li>{w}</li>
              {/each}
            </ul>
          </div>
        </div>
      {/if}

      <p class="text-sm mb-2 text-base-content/70">Restoring from: <code class="text-xs">{backupPath}</code></p>

      <div class="flex items-center justify-between mb-2">
        <span class="label-text font-medium">Categories ({selectedCats.size}/{allCats.length})</span>
      </div>

      <div class="grid grid-cols-2 gap-1 max-h-52 overflow-y-auto pr-1">
        {#each allCats as cat (cat.id)}
          <label class="flex items-center gap-2 px-2 py-1 rounded hover:bg-base-200 cursor-pointer text-sm">
            <input type="checkbox" class="checkbox checkbox-xs" checked={selectedCats.has(cat.id)} on:change={() => toggleCat(cat.id)} />
            <span>{cat.name}</span>
          </label>
        {/each}
      </div>

      {#if !confirmed}
        <div class="alert alert-error mt-3 text-sm">
          <label class="flex items-center gap-2 cursor-pointer">
            <input type="checkbox" class="checkbox checkbox-sm" bind:checked={confirmed} />
            <span>I understand this will overwrite my current GNOME settings. A safety backup will be created first.</span>
          </label>
        </div>
      {/if}

      <div class="modal-action">
        <button class="btn btn-sm btn-ghost" on:click={() => dispatch('close')}>Cancel</button>
        <button
          class="btn btn-sm btn-accent"
          disabled={running || selectedCats.size === 0 || !confirmed}
          on:click={startRestore}
        >
          {#if running}Restoring...{:else}Start Restore{/if}
        </button>
      </div>
    {/if}
  </div>
  <div class="modal-backdrop" on:click={() => { if (!running) dispatch('close') }} on:keydown={() => {}}></div>
</div>
