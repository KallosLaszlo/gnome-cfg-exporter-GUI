<script lang="ts">
  import { backups, selectedBackups, toggleSelect } from '../stores';
  import type { BackupInfo } from '../types';

  $: items = $backups;
  $: selected = $selectedBackups;

  function formatDate(ts: string): string {
    if (!ts) return '';
    try {
      return new Date(ts).toLocaleString();
    } catch {
      return ts;
    }
  }

  function catIcons(cats: string[]): string {
    return cats?.join(', ') || '';
  }
</script>

<div class="flex-1 overflow-auto bg-base-100">
  {#if items.length === 0}
    <div class="flex flex-col items-center justify-center h-full text-base-content/50">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4"/>
      </svg>
      <p class="text-lg">No backups yet</p>
      <p class="text-sm">Click <strong>Create</strong> to make your first backup</p>
    </div>
  {:else}
    <table class="table table-sm table-zebra w-full">
      <thead class="sticky top-0 bg-base-200 z-10">
        <tr>
          <th class="w-8"></th>
          <th>Date</th>
          <th>GNOME</th>
          <th>Host</th>
          <th>Size</th>
          <th>Files</th>
          <th>Description</th>
          <th>Categories</th>
        </tr>
      </thead>
      <tbody>
        {#each items as b (b.path)}
          <tr
            class="hover cursor-pointer"
            class:active={selected.has(b.path)}
            on:click={() => toggleSelect(b.path)}
          >
            <td>
              <input
                type="checkbox"
                class="checkbox checkbox-xs"
                checked={selected.has(b.path)}
                on:click|stopPropagation={() => toggleSelect(b.path)}
              />
            </td>
            <td class="whitespace-nowrap">{formatDate(b.timestamp)}</td>
            <td>{b.gnomeVersion || '-'}</td>
            <td>{b.hostname || '-'}</td>
            <td>{b.size}</td>
            <td>{b.fileCount}</td>
            <td class="max-w-[200px] truncate">{b.description || ''}</td>
            <td class="text-xs max-w-[200px] truncate">{catIcons(b.categories)}</td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>
