<script lang="ts">
  import { X, Upload, FileText, Trash2, Check } from 'lucide-svelte';
  import { safeCall } from '../utils/safe';
  import { GetAssets, AddAsset, DeleteAsset } from '../wails/wailsjs/go/main/App';
  import { main } from '../wails/wailsjs/go/models';
  import { eventBus } from '../stores/events.svelte';

  interface Props {
    itemId: number | null;
    itemName?: string;
    onClose: () => void;
  }

  let { itemId, itemName, onClose }: Props = $props();

  let assets = $state<main.AssetDTO[]>([]);
  let loading = $state(false);
  let uploading = $state(false);
  let isDragging = $state(false);

  // Upload form state
  let selectedFile: File | null = $state(null);
  let assetType = $state('manual');
  let assetName = $state('');
  let fileInput: HTMLInputElement | undefined = $state(undefined);

  const assetTypes = [
    { value: 'manual', label: 'Manuel utilisateur' },
    { value: 'service_manual', label: 'Manuel de service' },
    { value: 'exploded_view', label: 'Vue éclatée' },
    { value: 'stl', label: 'Fichier 3D (STL)' },
    { value: 'firmware', label: 'Firmware' },
    { value: 'driver', label: 'Driver' },
    { value: 'schematic', label: 'Schéma' },
    { value: 'other', label: 'Autre' }
  ];

  // Load assets when itemId changes
  $effect(() => {
    if (itemId !== null) {
      loadAssets();
    }
  });

  async function loadAssets() {
    if (itemId === null) return;

    loading = true;
    const [err, data] = await safeCall(GetAssets(itemId));

    if (err) {
      eventBus.emit({
        type: 'error',
        message: `Erreur lors du chargement des assets: ${err.message}`
      });
      loading = false;
      return;
    }

    assets = data || [];
    loading = false;
  }

  function handleDragEnter(e: DragEvent) {
    e.preventDefault();
    isDragging = true;
  }

  function handleDragOver(e: DragEvent) {
    e.preventDefault();
    e.dataTransfer!.dropEffect = 'copy';
  }

  function handleDragLeave(e: DragEvent) {
    e.preventDefault();
    // Only set isDragging to false if we're leaving the dropzone itself
    const target = e.currentTarget as HTMLElement;
    const related = e.relatedTarget as HTMLElement;
    if (!target.contains(related)) {
      isDragging = false;
    }
  }

  function handleDrop(e: DragEvent) {
    e.preventDefault();
    isDragging = false;

    const files = e.dataTransfer?.files;
    if (files && files.length > 0) {
      handleFileSelect(files[0]);
    }
  }

  function handleFileInputChange(e: Event) {
    const input = e.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      handleFileSelect(input.files[0]);
    }
  }

  function handleFileSelect(file: File) {
    selectedFile = file;
    // Auto-fill asset name with filename (without extension)
    if (!assetName) {
      assetName = file.name.replace(/\.[^/.]+$/, '');
    }
  }

  function clearSelection() {
    selectedFile = null;
    assetName = '';
    if (fileInput) {
      fileInput.value = '';
    }
  }

  async function handleUpload() {
    if (!selectedFile || itemId === null) return;

    uploading = true;

    // Note: In a real implementation, we would need to handle file upload differently
    // since Wails requires file paths, not file objects. For now, we'll show an error.
    eventBus.emit({
      type: 'warning',
      message: 'Upload de fichiers via drag & drop non implémenté. Utilisez la CLI pour ajouter des assets.'
    });

    uploading = false;

    // TODO: Implement file upload through Wails
    // This requires either:
    // 1. Writing the file to a temp location first, then passing the path
    // 2. Using Wails' file dialog to select files
    // 3. Creating a new Go handler that accepts file content directly
  }

  async function handleDelete(assetId: number, assetName: string) {
    const confirmed = confirm(`Voulez-vous vraiment supprimer "${assetName}" ?`);
    if (!confirmed) return;

    const [err] = await safeCall(DeleteAsset(assetId));

    if (err) {
      eventBus.emit({
        type: 'error',
        message: `Erreur lors de la suppression: ${err.message}`
      });
      return;
    }

    eventBus.emit({
      type: 'success',
      message: `"${assetName}" supprimé avec succès`
    });

    // Reload assets
    loadAssets();
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
    return date.toLocaleDateString('fr-FR', { year: 'numeric', month: 'short', day: 'numeric' });
  }

  function getAssetTypeLabel(type: string): string {
    const found = assetTypes.find(t => t.value === type);
    return found?.label || type;
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
        <div>
          <h2 class="text-2xl font-bold">Gestion des documents</h2>
          {#if itemName}
            <p class="text-sm text-muted-foreground mt-1">{itemName}</p>
          {/if}
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
        <div class="p-6 space-y-6">
          <!-- Upload Section -->
          <div class="space-y-4">
            <h3 class="font-semibold flex items-center gap-2">
              <Upload class="w-5 h-5" />
              Ajouter un document
            </h3>

            <!-- Drag & Drop Zone -->
            <div
              class="border-2 border-dashed rounded-lg p-8 text-center transition {isDragging ? 'border-primary bg-primary/5' : 'border-muted-foreground/25'}"
              ondragenter={handleDragEnter}
              ondragover={handleDragOver}
              ondragleave={handleDragLeave}
              ondrop={handleDrop}
            >
              {#if selectedFile}
                <div class="space-y-4">
                  <div class="flex items-center justify-center gap-2 text-primary">
                    <Check class="w-8 h-8" />
                    <div class="text-left">
                      <p class="font-medium">{selectedFile.name}</p>
                      <p class="text-sm text-muted-foreground">{formatFileSize(selectedFile.size)}</p>
                    </div>
                  </div>
                  <button
                    onclick={clearSelection}
                    class="text-sm text-muted-foreground hover:text-foreground underline"
                  >
                    Choisir un autre fichier
                  </button>
                </div>
              {:else}
                <Upload class="w-12 h-12 text-muted-foreground mx-auto mb-4" />
                <p class="font-medium mb-2">Glissez-déposez un fichier ici</p>
                <p class="text-sm text-muted-foreground mb-4">ou</p>
                <button
                  onclick={() => fileInput?.click()}
                  class="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition"
                >
                  Sélectionner un fichier
                </button>
                <input
                  bind:this={fileInput}
                  type="file"
                  class="hidden"
                  onchange={handleFileInputChange}
                />
              {/if}
            </div>

            {#if selectedFile}
              <!-- Upload Form -->
              <div class="space-y-3 p-4 bg-secondary/30 rounded-lg">
                <div>
                  <label for="assetType" class="block text-sm font-medium mb-2">
                    Type de document
                  </label>
                  <select
                    id="assetType"
                    bind:value={assetType}
                    class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                  >
                    {#each assetTypes as type}
                      <option value={type.value}>{type.label}</option>
                    {/each}
                  </select>
                </div>

                <div>
                  <label for="assetName" class="block text-sm font-medium mb-2">
                    Nom du document
                  </label>
                  <input
                    id="assetName"
                    type="text"
                    bind:value={assetName}
                    placeholder="Ex: Manuel utilisateur"
                    class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                  />
                </div>

                <div class="flex gap-3">
                  <button
                    onclick={clearSelection}
                    class="flex-1 px-4 py-2 border rounded-lg hover:bg-secondary transition"
                    disabled={uploading}
                  >
                    Annuler
                  </button>
                  <button
                    onclick={handleUpload}
                    disabled={uploading || !assetName.trim()}
                    class="flex-1 flex items-center justify-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {#if uploading}
                      <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-current"></div>
                      Upload...
                    {:else}
                      <Upload class="w-4 h-4" />
                      Ajouter
                    {/if}
                  </button>
                </div>
              </div>

              <!-- Info message -->
              <div class="bg-blue-50 border border-blue-200 text-blue-800 px-4 py-3 rounded-lg text-sm">
                <p class="font-semibold">ℹ️ Note</p>
                <p class="mt-1">
                  L'upload de fichiers via l'interface graphique nécessite des modifications côté backend.
                  Pour le moment, utilisez la commande CLI:
                  <code class="bg-blue-100 px-2 py-1 rounded mt-2 block font-mono text-xs">
                    brique asset add {itemId} /chemin/vers/fichier -t {assetType} -n "{assetName}"
                  </code>
                </p>
              </div>
            {/if}
          </div>

          <!-- Assets List -->
          <div class="space-y-4">
            <h3 class="font-semibold flex items-center gap-2">
              <FileText class="w-5 h-5" />
              Documents existants ({assets.length})
            </h3>

            {#if loading}
              <div class="flex items-center justify-center py-8">
                <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
              </div>
            {:else if assets.length === 0}
              <div class="text-center py-8 text-muted-foreground">
                <FileText class="w-12 h-12 mx-auto mb-2 opacity-50" />
                <p class="text-sm">Aucun document associé</p>
              </div>
            {:else}
              <div class="space-y-2">
                {#each assets as asset}
                  <div class="flex items-center justify-between p-4 bg-secondary/30 rounded-lg hover:bg-secondary/50 transition">
                    <div class="flex-1 min-w-0">
                      <div class="flex items-center gap-2 mb-1">
                        <FileText class="w-4 h-4 text-primary flex-shrink-0" />
                        <p class="font-medium truncate">{asset.name}</p>
                      </div>
                      <div class="flex items-center gap-3 text-xs text-muted-foreground">
                        <span class="px-2 py-0.5 bg-background rounded">{getAssetTypeLabel(asset.type)}</span>
                        <span>{formatFileSize(asset.fileSize)}</span>
                        <span>{formatDate(asset.createdAt)}</span>
                      </div>
                    </div>
                    <button
                      onclick={() => handleDelete(asset.id, asset.name)}
                      class="ml-4 p-2 text-destructive hover:bg-destructive/10 rounded-lg transition flex-shrink-0"
                      aria-label="Supprimer"
                    >
                      <Trash2 class="w-4 h-4" />
                    </button>
                  </div>
                {/each}
              </div>
            {/if}
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-end gap-3 p-6 border-t bg-secondary/30">
        <button
          onclick={onClose}
          class="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition"
        >
          Fermer
        </button>
      </div>
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
