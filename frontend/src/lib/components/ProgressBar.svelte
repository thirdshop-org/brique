<script lang="ts">
  import { eventBus } from '../stores/events.svelte';
  import { Loader2 } from 'lucide-svelte';

  const progressArray = $derived(Array.from(eventBus.progressEvents.values()));
</script>

{#if progressArray.length > 0}
  <div class="fixed bottom-4 right-4 z-50 w-96 max-w-[calc(100vw-2rem)]">
    <div class="bg-card border rounded-lg shadow-lg p-4 space-y-3">
      {#each progressArray as progress (progress.id)}
        <div class="space-y-2">
          <div class="flex items-center justify-between text-sm">
            <div class="flex items-center gap-2">
              <Loader2 class="w-4 h-4 animate-spin text-primary" />
              <span class="font-medium">{progress.operation}</span>
            </div>
            <span class="text-muted-foreground">
              {Math.round((progress.current / progress.total) * 100)}%
            </span>
          </div>

          {#if progress.filename}
            <p class="text-xs text-muted-foreground truncate">{progress.filename}</p>
          {/if}

          <!-- Progress bar -->
          <div class="h-2 bg-secondary rounded-full overflow-hidden">
            <div
              class="h-full bg-primary transition-all duration-300 ease-out"
              style="width: {(progress.current / progress.total) * 100}%"
            ></div>
          </div>

          <p class="text-xs text-muted-foreground">
            {progress.current} / {progress.total}
          </p>
        </div>
      {/each}
    </div>
  </div>
{/if}
