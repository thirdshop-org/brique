<script lang="ts">
  import { onMount } from 'svelte';
  import { Network, RefreshCw, Check, X, Clock, Wifi, WifiOff, AlertCircle, Plus } from 'lucide-svelte';
  import { safeCall } from '../utils/safe';
  import { GetPeers, SyncWithPeer, SetPeerTrusted, RemovePeer, GetSyncHistory, AddPeer } from '../wails/wailsjs/go/main/App';
  import { eventBus } from '../stores/events.svelte';

  interface Peer {
    id: string;
    name: string;
    address: string;
    lastSeen: string;
    lastSync: string;
    isTrusted: boolean;
    status: string;
  }

  interface SyncLog {
    id: number;
    peerName: string;
    timestamp: string;
    itemsReceived: number;
    itemsSent: number;
    conflicts: number;
    durationMs: number;
    error: string;
  }

  let peers = $state<Peer[]>([]);
  let syncLogs = $state<SyncLog[]>([]);
  let loading = $state(true);
  let syncing = $state<Record<string, boolean>>({});
  let showAddPeerModal = $state(false);
  let newPeer = $state({ name: '', address: '', isTrusted: true });

  onMount(() => {
    loadData();

    // Refresh peers every 10 seconds
    const interval = setInterval(loadPeers, 10000);
    return () => clearInterval(interval);
  });

  async function loadData() {
    loading = true;
    await Promise.all([loadPeers(), loadHistory()]);
    loading = false;
  }

  async function loadPeers() {
    const [err, data] = await safeCall(GetPeers());

    if (err) {
      eventBus.error(`Erreur: ${err.message}`);
      return;
    }

    peers = data || [];
  }

  async function loadHistory() {
    const [err, data] = await safeCall(GetSyncHistory(20));

    if (err) {
      return; // Silent fail for history
    }

    syncLogs = data || [];
  }

  async function handleSync(peer: Peer) {
    syncing[peer.id] = true;

    const [err, result] = await safeCall(SyncWithPeer(peer.id));

    syncing[peer.id] = false;

    if (err) {
      // Error already handled by backend
      return;
    }

    // Reload data after sync
    loadData();
  }

  async function handleToggleTrust(peer: Peer) {
    const [err] = await safeCall(SetPeerTrusted(peer.id, !peer.isTrusted));

    if (err) {
      // Error already handled by backend
      return;
    }

    // Reload peers
    loadPeers();
  }

  async function handleRemove(peer: Peer) {
    const confirmed = confirm(`Voulez-vous vraiment supprimer le pair "${peer.name}" ?`);
    if (!confirmed) return;

    const [err] = await safeCall(RemovePeer(peer.id));

    if (err) {
      // Error already handled by backend
      return;
    }

    // Reload peers
    loadPeers();
  }

  function getStatusColor(status: string) {
    switch (status) {
      case 'online': return 'text-green-600';
      case 'syncing': return 'text-blue-600';
      case 'offline': return 'text-gray-400';
      default: return 'text-gray-400';
    }
  }

  function getStatusIcon(status: string) {
    switch (status) {
      case 'online': return Wifi;
      case 'syncing': return RefreshCw;
      case 'offline': return WifiOff;
      default: return WifiOff;
    }
  }

  function formatDuration(ms: number): string {
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(1)}s`;
  }

  async function handleAddPeer() {
    if (!newPeer.name || !newPeer.address) {
      eventBus.error('Nom et adresse requis');
      return;
    }

    const [err] = await safeCall(AddPeer(newPeer.name, newPeer.address, newPeer.isTrusted));

    if (err) {
      // Error already handled by backend
      return;
    }

    // Reset form and close modal
    newPeer = { name: '', address: '', isTrusted: true };
    showAddPeerModal = false;

    // Reload peers
    loadPeers();
  }
</script>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-3">
      <div class="p-2 bg-primary/10 rounded-lg">
        <Network class="w-6 h-6 text-primary" />
      </div>
      <div>
        <h2 class="text-2xl font-bold">Synchronisation</h2>
        <p class="text-sm text-muted-foreground">Gestion des pairs et synchronisation P2P</p>
      </div>
    </div>

    <div class="flex gap-2">
      <button
        onclick={() => showAddPeerModal = true}
        class="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition"
        title="Ajouter un pair"
      >
        <Plus class="w-4 h-4" />
        Ajouter un pair
      </button>

      <button
        onclick={loadData}
        class="flex items-center gap-2 px-4 py-2 border rounded-lg hover:bg-secondary transition"
        title="Actualiser"
      >
        <RefreshCw class="w-4 h-4" />
        Actualiser
      </button>
    </div>
  </div>

  {#if loading}
    <div class="flex items-center justify-center py-16">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
    </div>
  {:else}
    <!-- Peers Section -->
    <div class="space-y-4">
      <h3 class="font-semibold text-lg">Pairs découverts ({peers.length})</h3>

      {#if peers.length === 0}
        <div class="bg-card border rounded-lg p-8 text-center">
          <Network class="w-12 h-12 text-muted-foreground mx-auto mb-4 opacity-50" />
          <p class="text-muted-foreground">Aucun pair découvert sur le réseau local</p>
          <p class="text-sm text-muted-foreground mt-2">
            Assurez-vous qu'une autre instance de Brique est active sur votre réseau
          </p>
        </div>
      {:else}
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          {#each peers as peer (peer.id)}
            <div class="bg-card border rounded-lg p-5">
              <!-- Header -->
              <div class="flex items-start justify-between mb-4">
                <div class="flex items-center gap-3">
                  <div class="p-2 bg-secondary rounded-lg">
                    {#if peer.status === 'online'}
                      <Wifi class="{getStatusColor(peer.status)} w-5 h-5" />
                    {:else}
                      <WifiOff class="{getStatusColor(peer.status)} w-5 h-5" />
                    {/if}
                  </div>
                  <div>
                    <h4 class="font-semibold">{peer.name}</h4>
                    <p class="text-xs text-muted-foreground">{peer.address}</p>
                  </div>
                </div>

                {#if peer.isTrusted}
                  <span class="px-2 py-1 bg-green-100 text-green-800 text-xs rounded-full">
                    Approuvé
                  </span>
                {/if}
              </div>

              <!-- Info -->
              <div class="space-y-2 mb-4 text-sm">
                {#if peer.lastSeen}
                  <div class="flex items-center gap-2 text-muted-foreground">
                    <Clock class="w-4 h-4" />
                    <span>Vu: {peer.lastSeen}</span>
                  </div>
                {/if}
                {#if peer.lastSync}
                  <div class="flex items-center gap-2 text-muted-foreground">
                    <RefreshCw class="w-4 h-4" />
                    <span>Dernière sync: {peer.lastSync}</span>
                  </div>
                {/if}
              </div>

              <!-- Actions -->
              <div class="flex gap-2">
                <button
                  onclick={() => handleSync(peer)}
                  disabled={syncing[peer.id] || peer.status !== 'online'}
                  class="flex-1 flex items-center justify-center gap-2 px-3 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {#if syncing[peer.id]}
                    <RefreshCw class="w-4 h-4 animate-spin" />
                    Sync...
                  {:else}
                    <RefreshCw class="w-4 h-4" />
                    Synchroniser
                  {/if}
                </button>

                <button
                  onclick={() => handleToggleTrust(peer)}
                  class="px-3 py-2 border rounded-lg hover:bg-secondary transition"
                  title={peer.isTrusted ? 'Révoquer la confiance' : 'Approuver'}
                >
                  {#if peer.isTrusted}
                    <X class="w-4 h-4" />
                  {:else}
                    <Check class="w-4 h-4" />
                  {/if}
                </button>

                <button
                  onclick={() => handleRemove(peer)}
                  class="px-3 py-2 text-destructive border border-destructive rounded-lg hover:bg-destructive/10 transition"
                  title="Supprimer"
                >
                  <X class="w-4 h-4" />
                </button>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>

    <!-- Sync History -->
    <div class="space-y-4">
      <h3 class="font-semibold text-lg">Historique de synchronisation</h3>

      {#if syncLogs.length === 0}
        <div class="bg-card border rounded-lg p-8 text-center">
          <RefreshCw class="w-12 h-12 text-muted-foreground mx-auto mb-4 opacity-50" />
          <p class="text-muted-foreground">Aucune synchronisation effectuée</p>
        </div>
      {:else}
        <div class="bg-card border rounded-lg divide-y">
          {#each syncLogs as log (log.id)}
            <div class="p-4 hover:bg-secondary/30 transition">
              <div class="flex items-center justify-between mb-2">
                <div class="flex items-center gap-2">
                  {#if log.error}
                    <AlertCircle class="w-4 h-4 text-destructive" />
                  {:else}
                    <Check class="w-4 h-4 text-green-600" />
                  {/if}
                  <span class="font-medium">{log.peerName || 'Unknown'}</span>
                </div>
                <span class="text-xs text-muted-foreground">{log.timestamp}</span>
              </div>

              <div class="flex items-center gap-4 text-sm text-muted-foreground">
                <span>↓ {log.itemsReceived} reçus</span>
                <span>↑ {log.itemsSent} envoyés</span>
                {#if log.conflicts > 0}
                  <span class="text-yellow-600">⚠ {log.conflicts} conflits</span>
                {/if}
                <span>{formatDuration(log.durationMs)}</span>
              </div>

              {#if log.error}
                <p class="text-sm text-destructive mt-2">{log.error}</p>
              {/if}
            </div>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>

<!-- Add Peer Modal -->
{#if showAddPeerModal}
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onclick={() => showAddPeerModal = false}>
    <div class="bg-card border rounded-lg p-6 w-full max-w-md" onclick={(e) => e.stopPropagation()}>
      <h3 class="text-xl font-bold mb-4">Ajouter un pair manuellement</h3>

      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium mb-1">Nom du pair</label>
          <input
            type="text"
            bind:value={newPeer.name}
            placeholder="ex: Brique Production"
            class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
          />
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">Adresse</label>
          <input
            type="text"
            bind:value={newPeer.address}
            placeholder="ex: https://brique.example.com"
            class="w-full px-3 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-primary"
          />
          <p class="text-xs text-muted-foreground mt-1">
            Entrez l'URL complète (ex: https://thirdshop-brique-xxx.traefik.me)<br/>
            ou juste le host:port (ex: 192.168.1.10:8080 pour HTTP local)
          </p>
        </div>

        <div class="flex items-center gap-2">
          <input
            type="checkbox"
            id="trusted"
            bind:checked={newPeer.isTrusted}
            class="w-4 h-4"
          />
          <label for="trusted" class="text-sm">Pair de confiance</label>
        </div>
      </div>

      <div class="flex gap-2 mt-6">
        <button
          onclick={handleAddPeer}
          class="flex-1 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition"
        >
          Ajouter
        </button>
        <button
          onclick={() => showAddPeerModal = false}
          class="px-4 py-2 border rounded-lg hover:bg-secondary transition"
        >
          Annuler
        </button>
      </div>
    </div>
  </div>
{/if}
