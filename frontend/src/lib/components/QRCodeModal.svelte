<script lang="ts">
  import { X, Download, QrCode } from 'lucide-svelte';
  import { safeCall } from '../utils/safe';
  import { GenerateQRCode } from '../wails/wailsjs/go/main/App';
  import { eventBus } from '../stores/events.svelte';

  interface Props {
    itemId: number | null;
    itemName?: string;
    onClose: () => void;
  }

  let { itemId, itemName, onClose }: Props = $props();

  let qrCodeBase64 = $state<string | null>(null);
  let loading = $state(false);
  let error = $state<string | null>(null);

  // Generate QR code when modal opens
  $effect(() => {
    if (itemId !== null) {
      generateQRCode();
    }
  });

  async function generateQRCode() {
    if (itemId === null) return;

    loading = true;
    error = null;

    const [err, data] = await safeCall(GenerateQRCode(itemId));

    if (err) {
      error = err.message;
      eventBus.emit({
        type: 'error',
        message: `Erreur lors de la génération: ${err.message}`
      });
      loading = false;
      return;
    }

    qrCodeBase64 = data;
    loading = false;
  }

  function downloadQRCode() {
    if (!qrCodeBase64) return;

    // Create a download link
    const link = document.createElement('a');
    link.href = `data:image/png;base64,${qrCodeBase64}`;
    link.download = `qrcode-${itemName || itemId}.png`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);

    eventBus.emit({
      type: 'success',
      message: 'QR Code téléchargé'
    });
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      onClose();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

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
      class="bg-card border rounded-lg shadow-xl max-w-md w-full overflow-hidden animate-slideUp"
      onclick={(e) => e.stopPropagation()}
      role="dialog"
      aria-modal="true"
    >
      <!-- Header -->
      <div class="flex items-center justify-between p-6 border-b">
        <div class="flex items-center gap-3">
          <div class="p-2 bg-primary/10 rounded-lg">
            <QrCode class="w-6 h-6 text-primary" />
          </div>
          <div>
            <h2 class="text-xl font-bold">QR Code</h2>
            {#if itemName}
              <p class="text-sm text-muted-foreground">{itemName}</p>
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
      <div class="p-6">
        {#if loading}
          <div class="flex flex-col items-center justify-center py-12">
            <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mb-4"></div>
            <p class="text-sm text-muted-foreground">Génération du QR Code...</p>
          </div>
        {:else if error}
          <div class="bg-destructive/10 text-destructive px-4 py-3 rounded-lg">
            <p class="font-semibold">Erreur</p>
            <p class="text-sm">{error}</p>
          </div>
        {:else if qrCodeBase64}
          <div class="space-y-4">
            <!-- QR Code Image -->
            <div class="flex justify-center p-4 bg-white rounded-lg">
              <img
                src={`data:image/png;base64,${qrCodeBase64}`}
                alt="QR Code"
                class="w-64 h-64"
              />
            </div>

            <!-- Info -->
            <div class="bg-blue-50 border border-blue-200 text-blue-800 px-4 py-3 rounded-lg text-sm">
              <p class="font-semibold mb-1">ℹ️ À propos de ce QR Code</p>
              <p>
                Ce QR Code contient les informations de l'objet (ID, nom, marque, modèle, numéro de série).
                Scannez-le avec l'application mobile Brique (à venir) pour accéder rapidement à la documentation.
              </p>
            </div>

            <!-- Actions -->
            <div class="flex gap-3">
              <button
                onclick={downloadQRCode}
                class="flex-1 flex items-center justify-center gap-2 px-4 py-3 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition"
              >
                <Download class="w-5 h-5" />
                Télécharger
              </button>
            </div>

            <!-- Print hint -->
            <p class="text-xs text-center text-muted-foreground">
              Imprimez ce QR Code et collez-le sur votre objet pour un accès rapide à sa documentation
            </p>
          </div>
        {/if}
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-end p-4 border-t bg-secondary/30">
        <button
          onclick={onClose}
          class="px-4 py-2 hover:bg-secondary rounded-lg transition"
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
