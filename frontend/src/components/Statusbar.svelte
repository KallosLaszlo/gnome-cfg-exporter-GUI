<script lang="ts">
  import { backups, systemInfo, freeSpace } from '../stores';

  $: info = $systemInfo;
  $: count = $backups.length;
  $: space = $freeSpace;
</script>

<div class="bg-base-200 border-t border-base-300 px-3 py-1 flex items-center text-xs text-base-content/70 gap-4 min-h-[28px]">
  <span>{count} backup{count !== 1 ? 's' : ''}</span>
  <span class="divider divider-horizontal mx-0 w-0 h-3"></span>
  {#if info}
    <span>{info.distro?.name || info.distro?.id || 'Unknown'}</span>
    <span>{info.gnomeVersion || ''}</span>
    <span>{info.sessionType || ''}</span>
  {/if}
  <div class="flex-1"></div>
  <span>Free: {space || '...'}</span>
</div>
