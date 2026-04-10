<script lang="ts">
  import { progress } from '../stores';

  $: p = $progress;
</script>

{#if p}
  <div class="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
    <div class="card bg-base-100 shadow-xl w-80">
      <div class="card-body p-5">
        <p class="text-sm font-medium mb-2">{p.message || 'Working...'}</p>
        <progress class="progress progress-primary w-full" value={p.percent} max="100"></progress>
        <p class="text-xs text-base-content/60 mt-1 text-right">{p.percent}%</p>
      </div>
    </div>
  </div>
{/if}
