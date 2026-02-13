<script lang="ts">
  import { onMount } from 'svelte';
  import { safeCall } from './lib/utils/safe';
  import { GetAllItems, GetItemWithAssets } from './lib/wails/wailsjs/go/main/App';
  import ItemCard from './lib/components/ItemCard.svelte';
  import NotificationToast from './lib/components/NotificationToast.svelte';
  import ProgressBar from './lib/components/ProgressBar.svelte';
  import { eventBus } from './lib/stores/events.svelte';
  import { Package, Plus, Search } from 'lucide-svelte';

  let items = $state<any[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let searchQuery = $state('');

  async function loadItems() {
    loading = true;
    error = null;

    const [err, data] = await safeCall(GetAllItems());

    if (err) {
      error = err.message;
      loading = false;
      return;
    }

    items = data || [];
    loading = false;
  }

  onMount(() => {
    loadItems();

    // Cleanup on unmount
    return () => {
      eventBus.destroy();
    };
  });

  const filteredItems = $derived(
    items.filter(item =>
      item.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      item.brand.toLowerCase().includes(searchQuery.toLowerCase()) ||
      item.category.toLowerCase().includes(searchQuery.toLowerCase())
    )
  );
</script>

<div class="min-h-screen bg-background">
  <!-- Header -->
  <header class="border-b bg-card">
    <div class="container mx-auto px-4 py-6">
      <div class="flex items-center gap-3">
        <div class="p-2 bg-primary rounded-lg">
          <Package class="w-6 h-6 text-primary-foreground" />
        </div>
        <div>
          <h1 class="text-2xl font-bold">Brique</h1>
          <p class="text-sm text-muted-foreground">L'infrastructure de résilience pour la réparation</p>
        </div>
      </div>
    </div>
  </header>

  <main class="container mx-auto px-4 py-8">
    <!-- Search and Actions -->
    <div class="flex gap-4 mb-8">
      <div class="flex-1 relative">
        <Search class="absolute left-3 top-1/2 transform -translate-y-1/2 w-5 h-5 text-muted-foreground" />
        <input
          type="text"
          placeholder="Rechercher un objet..."
          class="w-full pl-10 pr-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
          bind:value={searchQuery}
        />
      </div>
      <button
        class="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition"
        onclick={() => alert('Ajouter un item (à implémenter)')}
      >
        <Plus class="w-5 h-5" />
        Ajouter
      </button>
    </div>

    <!-- Content -->
    {#if loading}
      <div class="flex items-center justify-center py-16">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    {:else if error}
      <div class="bg-destructive/10 text-destructive px-4 py-3 rounded-lg">
        <p class="font-semibold">Erreur</p>
        <p class="text-sm">{error}</p>
      </div>
    {:else if filteredItems.length === 0}
      <div class="text-center py-16">
        <Package class="w-16 h-16 text-muted-foreground mx-auto mb-4" />
        <p class="text-lg font-semibold text-foreground">
          {searchQuery ? 'Aucun résultat' : 'Aucun objet dans l\'inventaire'}
        </p>
        <p class="text-sm text-muted-foreground mt-2">
          {searchQuery ? 'Essayez avec d\'autres mots-clés' : 'Commencez par ajouter votre premier objet'}
        </p>
      </div>
    {:else}
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {#each filteredItems as item (item.id)}
          <ItemCard {item} />
        {/each}
      </div>

      <div class="mt-8 text-center text-sm text-muted-foreground">
        {filteredItems.length} objet{filteredItems.length > 1 ? 's' : ''}
        {searchQuery ? ` trouvé${filteredItems.length > 1 ? 's' : ''}` : ' dans l\'inventaire'}
      </div>
    {/if}
  </main>
</div>

<!-- Global UI Components -->
<NotificationToast />
<ProgressBar />
