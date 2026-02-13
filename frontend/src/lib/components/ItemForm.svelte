<script lang="ts">
  import { X, Save, Package } from 'lucide-svelte';
  import { safeCall } from '../utils/safe';
  import { CreateItem, UpdateItem, GetItem } from '../wails/wailsjs/go/main/App';
  import { main } from '../wails/wailsjs/go/models';
  import { eventBus } from '../stores/events.svelte';

  interface Props {
    itemId: number | null; // null = création, number = édition
    onClose: () => void;
    onSuccess?: () => void;
  }

  let { itemId, onClose, onSuccess }: Props = $props();

  // Form fields
  let name = $state('');
  let category = $state('');
  let brand = $state('');
  let model = $state('');
  let serialNumber = $state('');
  let purchaseDate = $state('');
  let notes = $state('');

  // Form state
  let loading = $state(false);
  let submitting = $state(false);
  let errors = $state<Record<string, string>>({});

  const isEditMode = $derived(itemId !== null);
  const title = $derived(isEditMode ? 'Modifier l\'objet' : 'Ajouter un objet');

  // Load item data if in edit mode
  $effect(() => {
    if (itemId !== null) {
      loadItemData();
    }
  });

  async function loadItemData() {
    if (itemId === null) return;

    loading = true;
    const [err, data] = await safeCall(GetItem(itemId));

    if (err) {
      eventBus.emit({
        type: 'error',
        message: `Erreur lors du chargement: ${err.message}`
      });
      loading = false;
      return;
    }

    if (data) {
      name = data.name;
      category = data.category;
      brand = data.brand;
      model = data.model;
      serialNumber = data.serialNumber || '';
      purchaseDate = data.purchaseDate || '';
      notes = data.notes || '';
    }

    loading = false;
  }

  function validateForm(): boolean {
    errors = {};

    if (!name.trim()) {
      errors.name = 'Le nom est requis';
    }

    if (!category.trim()) {
      errors.category = 'La catégorie est requise';
    }

    if (!brand.trim()) {
      errors.brand = 'La marque est requise';
    }

    if (!model.trim()) {
      errors.model = 'Le modèle est requis';
    }

    if (purchaseDate && !isValidDate(purchaseDate)) {
      errors.purchaseDate = 'Format de date invalide (YYYY-MM-DD)';
    }

    return Object.keys(errors).length === 0;
  }

  function isValidDate(dateStr: string): boolean {
    if (!dateStr) return true;
    const regex = /^\d{4}-\d{2}-\d{2}$/;
    if (!regex.test(dateStr)) return false;
    const date = new Date(dateStr);
    return !isNaN(date.getTime());
  }

  async function handleSubmit() {
    if (!validateForm()) {
      return;
    }

    submitting = true;

    if (isEditMode && itemId !== null) {
      // Update existing item
      const [err] = await safeCall(
        UpdateItem(
          itemId,
          name.trim(),
          category.trim(),
          brand.trim(),
          model.trim(),
          serialNumber.trim(),
          purchaseDate.trim(),
          notes.trim()
        )
      );

      if (err) {
        eventBus.emit({
          type: 'error',
          message: `Erreur lors de la modification: ${err.message}`
        });
        submitting = false;
        return;
      }

      eventBus.emit({
        type: 'success',
        message: `"${name}" modifié avec succès`
      });
    } else {
      // Create new item
      const [err, data] = await safeCall(
        CreateItem(
          name.trim(),
          category.trim(),
          brand.trim(),
          model.trim(),
          serialNumber.trim(),
          notes.trim()
        )
      );

      if (err) {
        eventBus.emit({
          type: 'error',
          message: `Erreur lors de la création: ${err.message}`
        });
        submitting = false;
        return;
      }

      eventBus.emit({
        type: 'success',
        message: `"${name}" ajouté avec succès`
      });
    }

    submitting = false;
    onSuccess?.();
    onClose();
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      onClose();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

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
    class="bg-card border rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-hidden animate-slideUp"
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
        <h2 class="text-2xl font-bold">{title}</h2>
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
    <div class="overflow-y-auto max-h-[calc(90vh-180px)]">
      {#if loading}
        <div class="flex items-center justify-center py-16">
          <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
        </div>
      {:else}
        <form onsubmit={(e) => { e.preventDefault(); handleSubmit(); }} class="p-6 space-y-4">
          <!-- Name -->
          <div>
            <label for="name" class="block text-sm font-medium mb-2">
              Nom <span class="text-destructive">*</span>
            </label>
            <input
              id="name"
              type="text"
              bind:value={name}
              placeholder="Ex: Machine à laver"
              class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring {errors.name ? 'border-destructive' : ''}"
              required
            />
            {#if errors.name}
              <p class="text-destructive text-xs mt-1">{errors.name}</p>
            {/if}
          </div>

          <!-- Category -->
          <div>
            <label for="category" class="block text-sm font-medium mb-2">
              Catégorie <span class="text-destructive">*</span>
            </label>
            <input
              id="category"
              type="text"
              bind:value={category}
              placeholder="Ex: Électroménager"
              class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring {errors.category ? 'border-destructive' : ''}"
              required
            />
            {#if errors.category}
              <p class="text-destructive text-xs mt-1">{errors.category}</p>
            {/if}
          </div>

          <!-- Brand and Model -->
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label for="brand" class="block text-sm font-medium mb-2">
                Marque <span class="text-destructive">*</span>
              </label>
              <input
                id="brand"
                type="text"
                bind:value={brand}
                placeholder="Ex: Bosch"
                class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring {errors.brand ? 'border-destructive' : ''}"
                required
              />
              {#if errors.brand}
                <p class="text-destructive text-xs mt-1">{errors.brand}</p>
              {/if}
            </div>

            <div>
              <label for="model" class="block text-sm font-medium mb-2">
                Modèle <span class="text-destructive">*</span>
              </label>
              <input
                id="model"
                type="text"
                bind:value={model}
                placeholder="Ex: WAE28210FF"
                class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring {errors.model ? 'border-destructive' : ''}"
                required
              />
              {#if errors.model}
                <p class="text-destructive text-xs mt-1">{errors.model}</p>
              {/if}
            </div>
          </div>

          <!-- Serial Number -->
          <div>
            <label for="serialNumber" class="block text-sm font-medium mb-2">
              Numéro de série
            </label>
            <input
              id="serialNumber"
              type="text"
              bind:value={serialNumber}
              placeholder="Ex: SN123456789"
              class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </div>

          <!-- Purchase Date -->
          <div>
            <label for="purchaseDate" class="block text-sm font-medium mb-2">
              Date d'achat
            </label>
            <input
              id="purchaseDate"
              type="date"
              bind:value={purchaseDate}
              class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring {errors.purchaseDate ? 'border-destructive' : ''}"
            />
            {#if errors.purchaseDate}
              <p class="text-destructive text-xs mt-1">{errors.purchaseDate}</p>
            {/if}
          </div>

          <!-- Notes -->
          <div>
            <label for="notes" class="block text-sm font-medium mb-2">
              Notes
            </label>
            <textarea
              id="notes"
              bind:value={notes}
              placeholder="Notes additionnelles..."
              rows="4"
              class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring resize-none"
            ></textarea>
          </div>
        </form>
      {/if}
    </div>

    <!-- Footer -->
    {#if !loading}
      <div class="flex items-center justify-end gap-3 p-6 border-t bg-secondary/30">
        <button
          onclick={onClose}
          class="px-4 py-2 border rounded-lg hover:bg-secondary transition"
          disabled={submitting}
        >
          Annuler
        </button>
        <button
          onclick={handleSubmit}
          disabled={submitting}
          class="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {#if submitting}
            <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-current"></div>
            En cours...
          {:else}
            <Save class="w-4 h-4" />
            {isEditMode ? 'Enregistrer' : 'Créer'}
          {/if}
        </button>
      </div>
    {/if}
  </div>
</div>

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
