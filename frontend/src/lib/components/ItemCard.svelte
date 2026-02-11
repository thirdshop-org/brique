<script lang="ts">
  import { Package, FileText } from 'lucide-svelte';

  interface Props {
    item: any;
  }

  let { item }: Props = $props();

  function getHealthColor(health: string) {
    switch(health) {
      case 'secured': return 'bg-green-100 text-green-800';
      case 'partial': return 'bg-yellow-100 text-yellow-800';
      case 'incomplete': return 'bg-red-100 text-red-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  }

  function getHealthEmoji(health: string) {
    switch(health) {
      case 'secured': return '🟢';
      case 'partial': return '🟡';
      case 'incomplete': return '🔴';
      default: return '❓';
    }
  }

  function getHealthLabel(health: string) {
    switch(health) {
      case 'secured': return 'Sécurisé';
      case 'partial': return 'Partiel';
      case 'incomplete': return 'Incomplet';
      default: return 'Inconnu';
    }
  }
</script>

<div class="bg-card border rounded-lg p-5 hover:shadow-lg transition cursor-pointer">
  <div class="flex items-start justify-between mb-3">
    <div class="p-2 bg-primary/10 rounded-lg">
      <Package class="w-5 h-5 text-primary" />
    </div>
    <span class="text-xs px-2 py-1 rounded-full {getHealthColor(item.health || 'incomplete')}">
      {getHealthEmoji(item.health || 'incomplete')} {getHealthLabel(item.health || 'incomplete')}
    </span>
  </div>

  <h3 class="font-semibold text-lg mb-1">{item.name}</h3>
  <p class="text-sm text-muted-foreground mb-3">
    {item.brand} {item.model}
  </p>

  <div class="flex items-center justify-between text-xs text-muted-foreground">
    <span class="px-2 py-1 bg-secondary rounded">{item.category}</span>
    {#if item.serialNumber}
      <span>#{item.serialNumber}</span>
    {/if}
  </div>

  {#if item.notes}
    <p class="text-xs text-muted-foreground mt-3 line-clamp-2">{item.notes}</p>
  {/if}
</div>
