<script lang="ts">
  import { X, Package, Edit, Trash2, FileText, Calendar, Tag, Hash, QrCode } from 'lucide-svelte';
  import { safeCall } from '../utils/safe';
  import { GetItemWithAssets, DeleteItem } from '../wails/wailsjs/go/main/App';
  import { main } from '../wails/wailsjs/go/models';
  import { eventBus } from '../stores/events.svelte';

  interface Props {
    itemId: number | null;
    onClose: () => void;
    onEdit?: (itemId: number) => void;
    onDelete?: () => void;
    onGenerateQR?: (itemId: number, itemName: string) => void;
  }

  let { itemId, onClose, onEdit, onDelete, onGenerateQR }: Props = $props();

  let itemWithAssets = $state<main.ItemWithAssetsDTO | null>(null);
  let loading = $state(false);
  let error = $state<string | null>(null);
  let showDeleteConfirm = $state(false);

  // Load item data when itemId changes
  $effect(() => {
    if (itemId !== null) {
      loadItemData();
    }
  });

  async function loadItemData() {
    if (itemId === null) return;

    loading = true;
    error = null;

    const [err, data] = await safeCall(GetItemWithAssets(itemId));

    if (err) {
      error = err.message;
      loading = false;
      return;
    }

    itemWithAssets = data;
    loading = false;
  }

  async function handleDelete() {
    if (itemId === null || !itemWithAssets) return;

    const [err] = await safeCall(DeleteItem(itemId));

    if (err) {
      eventBus.emit({
        type: 'error',
        message: `Erreur lors de la suppression: ${err.message}`
      });
      return;
    }

    eventBus.emit({
      type: 'success',
      message: `"${itemWithAssets.item.name}" supprimé avec succès`
    });

    onDelete?.();
    onClose();
  }

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

  function formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`;
  }

  function formatDate(dateStr: string): string {
    if (!dateStr) return 'N/A';
    const date = new Date(dateStr);
    return date.toLocaleDateString('fr-FR', { year: 'numeric', month: 'long', day: 'numeric' });
  }

  function getAssetTypeLabel(type: string): string {
    const labels: Record<string, string> = {
      manual: 'Manuel utilisateur',
      service_manual: 'Manuel de service',
      exploded_view: 'Vue éclatée',
      stl: 'Fichier 3D (STL)',
      firmware: 'Firmware',
      driver: 'Driver',
      schematic: 'Schéma',
      other: 'Autre'
    };
    return labels[type] || type;
  }
</script>

{#if itemId !== null}
  <!-- Modal Backdrop -->
  <div
    class="fixed inset-0 bg-black/50 z-40 animate-fadeIn"
    onclick={onClose}
    role="button"
    tabindex="-1"
  ></div>

  <!-- Modal Content -->
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <div
      class="bg-card border rounded-lg shadow-xl max-w-3xl w-full max-h-[90vh] overflow-hidden animate-slideUp"
      onclick={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
    >
      <!-- Header -->
      <div class="flex items-center justify-between p-6 border-b">
        <div class="flex items-center gap-3">
          <div class="p-2 bg-primary/10 rounded-lg">
            <Package class="w-6 h-6 text-primary" />
          </div>
          <div>
            <h2 class="text-2xl font-bold">Détails de l'objet</h2>
            {#if itemWithAssets}
              <p class="text-sm text-muted-foreground">{itemWithAssets.item.brand} {itemWithAssets.item.model}</p>
            {/if}
          </div>
        </div>
        <button
          onclick={onClose}
          class="p-2 hover:bg-secondary rounded-lg transition"
          aria-label="Fermer"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Body -->
      <div class="overflow-y-auto max-h-[calc(90vh-200px)]">
        {#if loading}
          <div class="flex items-center justify-center py-16">
            <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
          </div>
        {:else if error}
          <div class="p-6">
            <div class="bg-destructive/10 text-destructive px-4 py-3 rounded-lg">
              <p class="font-semibold">Erreur</p>
              <p class="text-sm">{error}</p>
            </div>
          </div>
        {:else if itemWithAssets}
          <div class="p-6 space-y-6">
            <!-- Health Status -->
            <div class="flex items-center justify-between">
              <span class="text-sm font-medium">Santé documentaire:</span>
              <span class="text-sm px-3 py-1 rounded-full {getHealthColor(itemWithAssets.health)}">
                {getHealthEmoji(itemWithAssets.health)} {getHealthLabel(itemWithAssets.health)}
              </span>
            </div>

            <!-- Item Details -->
            <div class="space-y-4">
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="text-xs text-muted-foreground flex items-center gap-1 mb-1">
                    <Package class="w-3 h-3" />
                    Nom
                  </label>
                  <p class="font-medium">{itemWithAssets.item.name}</p>
                </div>

                <div>
                  <label class="text-xs text-muted-foreground flex items-center gap-1 mb-1">
                    <Tag class="w-3 h-3" />
                    Catégorie
                  </label>
                  <p class="font-medium">{itemWithAssets.item.category}</p>
                </div>

                <div>
                  <label class="text-xs text-muted-foreground mb-1 block">
                    Marque
                  </label>
                  <p class="font-medium">{itemWithAssets.item.brand}</p>
                </div>

                <div>
                  <label class="text-xs text-muted-foreground mb-1 block">
                    Modèle
                  </label>
                  <p class="font-medium">{itemWithAssets.item.model}</p>
                </div>

                {#if itemWithAssets.item.serialNumber}
                  <div>
                    <label class="text-xs text-muted-foreground flex items-center gap-1 mb-1">
                      <Hash class="w-3 h-3" />
                      Numéro de série
                    </label>
                    <p class="font-medium font-mono">{itemWithAssets.item.serialNumber}</p>
                  </div>
                {/if}

                {#if itemWithAssets.item.purchaseDate}
                  <div>
                    <label class="text-xs text-muted-foreground flex items-center gap-1 mb-1">
                      <Calendar class="w-3 h-3" />
                      Date d'achat
                    </label>
                    <p class="font-medium">{formatDate(itemWithAssets.item.purchaseDate)}</p>
                  </div>
                {/if}
              </div>

              {#if itemWithAssets.item.notes}
                <div>
                  <label class="text-xs text-muted-foreground flex items-center gap-1 mb-1">
                    <FileText class="w-3 h-3" />
                    Notes
                  </label>
                  <p class="text-sm bg-secondary p-3 rounded-lg">{itemWithAssets.item.notes}</p>
                </div>
              {/if}

              <div class="text-xs text-muted-foreground space-y-1">
                <p>Créé le: {formatDate(itemWithAssets.item.createdAt)}</p>
                <p>Modifié le: {formatDate(itemWithAssets.item.updatedAt)}</p>
              </div>
            </div>

            <!-- Assets Section -->
            <div class="border-t pt-6">
              <h3 class="font-semibold mb-4 flex items-center gap-2">
                <FileText class="w-5 h-5" />
                Documents et fichiers ({itemWithAssets.assets?.length || 0})
              </h3>

              {#if itemWithAssets.assets && itemWithAssets.assets.length > 0}
                <div class="space-y-2">
                  {#each itemWithAssets.assets as asset}
                    <div class="flex items-center justify-between p-3 bg-secondary rounded-lg">
                      <div class="flex-1 min-w-0">
                        <p class="font-medium truncate">{asset.name}</p>
                        <div class="flex items-center gap-3 text-xs text-muted-foreground mt-1">
                          <span class="px-2 py-0.5 bg-background rounded">{getAssetTypeLabel(asset.type)}</span>
                          <span>{formatFileSize(asset.fileSize)}</span>
                          <span>{formatDate(asset.createdAt)}</span>
                        </div>
                      </div>
                    </div>
                  {/each}
                </div>
              {:else}
                <div class="text-center py-8 text-muted-foreground">
                  <FileText class="w-12 h-12 mx-auto mb-2 opacity-50" />
                  <p class="text-sm">Aucun document associé</p>
                </div>
              {/if}
            </div>
          </div>
        {/if}
      </div>

      <!-- Footer -->
      {#if !loading && itemWithAssets && !showDeleteConfirm}
        <div class="flex items-center justify-between gap-3 p-6 border-t bg-secondary/30">
          {#if onGenerateQR}
            <button
              onclick={() => onGenerateQR?.(itemId!, itemWithAssets!.item.name)}
              class="flex items-center gap-2 px-4 py-2 border rounded-lg hover:bg-secondary transition"
            >
              <QrCode class="w-4 h-4" />
              QR Code
            </button>
          {/if}
          <div class="flex items-center gap-3 ml-auto">
            {#if onEdit}
              <button
                onclick={() => onEdit?.(itemId!)}
                class="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition"
              >
                <Edit class="w-4 h-4" />
                Modifier
              </button>
            {/if}
            <button
              onclick={() => showDeleteConfirm = true}
              class="flex items-center gap-2 px-4 py-2 bg-destructive text-destructive-foreground rounded-lg hover:bg-destructive/90 transition"
            >
              <Trash2 class="w-4 h-4" />
              Supprimer
            </button>
          </div>
        </div>
      {/if}

      {#if showDeleteConfirm}
        <div class="p-6 border-t bg-destructive/10">
          <p class="font-semibold mb-2">⚠️ Confirmer la suppression</p>
          <p class="text-sm mb-4">
            Êtes-vous sûr de vouloir supprimer "{itemWithAssets?.item.name}" ?
            Cette action est irréversible et supprimera également tous les documents associés.
          </p>
          <div class="flex items-center justify-end gap-3">
            <button
              onclick={() => showDeleteConfirm = false}
              class="px-4 py-2 border rounded-lg hover:bg-secondary transition"
            >
              Annuler
            </button>
            <button
              onclick={handleDelete}
              class="px-4 py-2 bg-destructive text-destructive-foreground rounded-lg hover:bg-destructive/90 transition"
            >
              Confirmer la suppression
            </button>
          </div>
        </div>
      {/if}
    </div>
  </div>
{/if}

<style>
  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }

  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .animate-fadeIn {
    animation: fadeIn 0.2s ease-out;
  }

  .animate-slideUp {
    animation: slideUp 0.3s ease-out;
  }
</style>
