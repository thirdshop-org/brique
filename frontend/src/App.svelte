<script lang="ts">
  import { onMount } from 'svelte';
  import { safeCall } from './lib/utils/safe';
  import { GetAllItems } from './lib/wails/wailsjs/go/main/App';
  import { main } from './lib/wails/wailsjs/go/models';
  import ItemCard from './lib/components/ItemCard.svelte';
  import ItemDetailModal from './lib/components/ItemDetailModal.svelte';
  import ItemForm from './lib/components/ItemForm.svelte';
  import AssetManager from './lib/components/AssetManager.svelte';
  import Dashboard from './lib/components/Dashboard.svelte';
  import SyncView from './lib/components/SyncView.svelte';
  import QRCodeModal from './lib/components/QRCodeModal.svelte';
  import NotificationToast from './lib/components/NotificationToast.svelte';
  import ProgressBar from './lib/components/ProgressBar.svelte';
  import { eventBus } from './lib/stores/events.svelte';
  import { Package, Plus, Search, BarChart3, List, Network } from 'lucide-svelte';

  let items = $state<main.ItemDTO[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let searchQuery = $state('');

  // Navigation
  let currentView = $state<'inventory' | 'dashboard' | 'sync'>('inventory');

  // Modal states
  let detailModalItemId = $state<number | null>(null);
  let formModalItemId = $state<number | null | undefined>(undefined); // null = create, number = edit
  let assetManagerItemId = $state<number | null>(null);
  let assetManagerItemName = $state<string>('');
  let qrCodeModalItemId = $state<number | null>(null);
  let qrCodeModalItemName = $state<string>('');

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

  function openDetailModal(itemId: number) {
    detailModalItemId = itemId;
  }

  function closeDetailModal() {
    detailModalItemId = null;
  }

  function openFormModal(itemId: number | null = null) {
    formModalItemId = itemId;
  }

  function closeFormModal() {
    formModalItemId = undefined;
  }

  function openAssetManager(itemId: number, itemName: string) {
    assetManagerItemId = itemId;
    assetManagerItemName = itemName;
  }

  function closeAssetManager() {
    assetManagerItemId = null;
    assetManagerItemName = '';
  }

  function openQRCodeModal(itemId: number, itemName: string) {
    qrCodeModalItemId = itemId;
    qrCodeModalItemName = itemName;
  }

  function closeQRCodeModal() {
    qrCodeModalItemId = null;
    qrCodeModalItemName = '';
  }

  function handleItemEdit(itemId: number) {
    closeDetailModal();
    openFormModal(itemId);
  }

  function handleItemDelete() {
    closeDetailModal();
    loadItems(); // Reload list after delete
  }

  function handleFormSuccess() {
    loadItems(); // Reload list after create/edit
  }
</script>

<div class="min-h-screen bg-background">
  <!-- Header -->
  <header class="border-b bg-card">
    <div class="container mx-auto px-4 py-6">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="p-2 bg-primary rounded-lg">
            <Package class="w-6 h-6 text-primary-foreground" />
          </div>
          <div>
            <h1 class="text-2xl font-bold">Brique</h1>
            <p class="text-sm text-muted-foreground">L'infrastructure de résilience pour la réparation</p>
          </div>
        </div>

        <!-- Navigation Tabs -->
        <div class="flex gap-2">
          <button
            class="flex items-center gap-2 px-4 py-2 rounded-lg transition {currentView === 'inventory' ? 'bg-primary text-primary-foreground' : 'hover:bg-secondary'}"
            onclick={() => currentView = 'inventory'}
          >
            <List class="w-4 h-4" />
            Inventaire
          </button>
          <button
            class="flex items-center gap-2 px-4 py-2 rounded-lg transition {currentView === 'dashboard' ? 'bg-primary text-primary-foreground' : 'hover:bg-secondary'}"
            onclick={() => currentView = 'dashboard'}
          >
            <BarChart3 class="w-4 h-4" />
            Tableau de bord
          </button>
          <button
            class="flex items-center gap-2 px-4 py-2 rounded-lg transition {currentView === 'sync' ? 'bg-primary text-primary-foreground' : 'hover:bg-secondary'}"
            onclick={() => currentView = 'sync'}
          >
            <Network class="w-4 h-4" />
            Synchronisation
          </button>
        </div>
      </div>
    </div>
  </header>

  <main class="container mx-auto px-4 py-8">
    {#if currentView === 'inventory'}
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
          onclick={() => openFormModal(null)}
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
            <ItemCard {item} onclick={() => openDetailModal(item.id)} />
          {/each}
        </div>

        <div class="mt-8 text-center text-sm text-muted-foreground">
          {filteredItems.length} objet{filteredItems.length > 1 ? 's' : ''}
          {searchQuery ? ` trouvé${filteredItems.length > 1 ? 's' : ''}` : ' dans l\'inventaire'}
        </div>
      {/if}
    {:else if currentView === 'dashboard'}
      <Dashboard />
    {:else if currentView === 'sync'}
      <SyncView />
    {/if}
  </main>
</div>

<!-- Global UI Components -->
<NotificationToast />
<ProgressBar />

<!-- Modals -->
{#if detailModalItemId !== null}
  <ItemDetailModal
    itemId={detailModalItemId}
    onClose={closeDetailModal}
    onEdit={handleItemEdit}
    onDelete={handleItemDelete}
    onGenerateQR={openQRCodeModal}
  />
{/if}

{#if formModalItemId !== undefined }
  <ItemForm
    itemId={formModalItemId}
    onClose={closeFormModal}
    onSuccess={handleFormSuccess}
  />
{/if}

{#if assetManagerItemId !== null}
  <AssetManager
    itemId={assetManagerItemId}
    itemName={assetManagerItemName}
    onClose={closeAssetManager}
  />
{/if}

{#if qrCodeModalItemId !== null}
  <QRCodeModal
    itemId={qrCodeModalItemId}
    itemName={qrCodeModalItemName}
    onClose={closeQRCodeModal}
  />
{/if}
