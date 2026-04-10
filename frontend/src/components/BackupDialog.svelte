<script lang="ts">
  import { createEventDispatcher, onMount } from 'svelte';
  import { categories, progress, progressVisible } from '../stores';
  import type { Category } from '../types';
  import { CreateBackup } from '../../wailsjs/go/main/App';

  const dispatch = createEventDispatcher();

  let allCats: Category[] = [];
  let selectedCats: Set<string> = new Set();
  let description = '';
  let running = false;
  let resultMsg = '';

  onMount(() => {
    allCats = $categories;
    allCats.forEach(c => selectedCats.add(c.id));
    selectedCats = selectedCats;
  });

  function toggleCat(id: string) {
    if (selectedCats.has(id)) selectedCats.delete(id);
    else selectedCats.add(id);
    selectedCats = selectedCats;
  }

  function selectAll() {
    allCats.forEach(c => selectedCats.add(c.id));
    selectedCats = selectedCats;
  }

  function selectNone() {
    selectedCats.clear();
    selectedCats = selectedCats;
  }

  async function startBackup() {
    if (selectedCats.size === 0) return;
    running = true;
    progressVisible.set(true);
    try {
      const result = await CreateBackup(Array.from(selectedCats), description);
      resultMsg = `Backup created: ${result.size}, ${result.fileCount} files`;
      if (result.errors?.length) {
        resultMsg += '\nWarnings: ' + result.errors.join('; ');
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
    <h3 class="font-bold text-lg mb-3">Create Backup</h3>

    {#if resultMsg}
      <div class="alert alert-info mb-3 text-sm whitespace-pre-line">{resultMsg}</div>
      <div class="modal-action">
        <button class="btn btn-sm" on:click={() => dispatch('close')}>Close</button>
      </div>
    {:else if running}
      <div class="flex flex-col items-center gap-4 py-6">
        <span class="loading loading-spinner loading-lg text-primary"></span>
        <p class="text-sm font-medium">{$progress?.message || 'Starting backup...'}</p>
        <progress class="progress progress-primary w-full" value={$progress?.percent || 0} max="100"></progress>
        <p class="text-xs text-base-content/50">{$progress?.percent || 0}%</p>
      </div>
    {:else}
      <div class="form-control mb-3">
        <label class="label pb-1"><span class="label-text">Description (optional)</span></label>
        <input type="text" class="input input-sm input-bordered" bind:value={description} placeholder="e.g. Before theme change" />
      </div>

      <div class="flex items-center justify-between mb-2">
        <span class="label-text font-medium">Categories ({selectedCats.size}/{allCats.length})</span>
        <div class="flex gap-1">
          <button class="btn btn-xs btn-ghost" on:click={selectAll}>All</button>
          <button class="btn btn-xs btn-ghost" on:click={selectNone}>None</button>
        </div>
      </div>

      <div class="grid grid-cols-2 gap-1 max-h-60 overflow-y-auto pr-1">
        {#each allCats as cat (cat.id)}
          <label class="flex items-center gap-2 px-2 py-1 rounded hover:bg-base-200 cursor-pointer text-sm">
            <input type="checkbox" class="checkbox checkbox-xs" checked={selectedCats.has(cat.id)} on:change={() => toggleCat(cat.id)} />
            <span>{cat.name}</span>
          </label>
        {/each}
      </div>

      <div class="modal-action">
        <button class="btn btn-sm btn-ghost" on:click={() => dispatch('close')}>Cancel</button>
        <button class="btn btn-sm btn-primary" disabled={running || selectedCats.size === 0} on:click={startBackup}>
          {#if running}Backing up...{:else}Start Backup{/if}
        </button>
      </div>
    {/if}
  </div>
  <div class="modal-backdrop" on:click={() => { if (!running) dispatch('close') }} on:keydown={() => {}}></div>
</div>
