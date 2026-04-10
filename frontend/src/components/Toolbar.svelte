<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { selectedBackups, backups } from '../stores';
  import { DeleteBackups, BrowseBackup, CompareBackups } from '../../wailsjs/go/main/App';

  const dispatch = createEventDispatcher();
  let compareResult = '';
  let showCompare = false;
  let deleting = false;

  $: selected = $selectedBackups;
  $: selectedCount = selected.size;

  async function handleDelete() {
    if (selectedCount === 0) return;
    const paths = Array.from(selected);
    if (!confirm(`Delete ${selectedCount} backup(s)?`)) return;
    deleting = true;
    try {
      await DeleteBackups(paths);
      selectedBackups.set(new Set());
      dispatch('refresh');
    } finally {
      deleting = false;
    }
  }

  async function handleBrowse() {
    const paths = Array.from(selected);
    if (paths.length === 1) await BrowseBackup(paths[0]);
  }

  async function handleCompare() {
    const paths = Array.from(selected);
    if (paths.length !== 2) return;
    try {
      compareResult = await CompareBackups(paths[0], paths[1]);
      showCompare = true;
    } catch (e) {
      compareResult = 'Error: ' + e;
      showCompare = true;
    }
  }
</script>

<div class="navbar bg-base-200 border-b border-base-300 px-2 min-h-0 h-12 gap-1">
  <div class="flex gap-1">
    <button class="btn btn-sm btn-primary gap-1" on:click={() => dispatch('backup')}>
      <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" /></svg>
      Create
    </button>
    <button
      class="btn btn-sm btn-accent gap-1"
      disabled={selectedCount !== 1}
      on:click={() => dispatch('restore')}
    >
      <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/></svg>
      Restore
    </button>
  </div>

  <div class="divider divider-horizontal mx-0 w-0"></div>

  <div class="flex gap-1">
    <button
      class="btn btn-sm btn-ghost gap-1"
      disabled={selectedCount !== 1}
      on:click={handleBrowse}
    >Browse</button>
    <button
      class="btn btn-sm btn-ghost gap-1"
      disabled={selectedCount !== 2}
      on:click={handleCompare}
    >Compare</button>
    <button
      class="btn btn-sm btn-error btn-outline gap-1"
      disabled={selectedCount === 0 || deleting}
      on:click={handleDelete}
    >{#if deleting}<span class="loading loading-spinner loading-xs"></span> Deleting...{:else}Delete{/if}</button>
  </div>

  <div class="flex-1"></div>

  <div class="flex gap-1">
    <button class="btn btn-sm btn-ghost btn-square" on:click={() => dispatch('refresh')} title="Refresh">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/></svg>
    </button>
    <button class="btn btn-sm btn-ghost btn-square" on:click={() => dispatch('settings')} title="Settings">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/></svg>
    </button>
  </div>
</div>

{#if showCompare}
  <div class="modal modal-open">
    <div class="modal-box max-w-2xl">
      <h3 class="font-bold text-lg mb-3">Backup Comparison</h3>
      <pre class="bg-base-300 p-3 rounded text-sm overflow-auto max-h-96 whitespace-pre-wrap">{compareResult}</pre>
      <div class="modal-action">
        <button class="btn btn-sm" on:click={() => showCompare = false}>Close</button>
      </div>
    </div>
    <div class="modal-backdrop" on:click={() => showCompare = false} on:keydown={() => {}}></div>
  </div>
{/if}
