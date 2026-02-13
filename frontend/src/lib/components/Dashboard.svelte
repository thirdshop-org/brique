<script lang="ts">
  import { onMount } from 'svelte';
  import { Package, TrendingUp, AlertCircle, CheckCircle, AlertTriangle, BarChart3, Download, FileJson, FileSpreadsheet, Upload, HardDrive } from 'lucide-svelte';
  import { safeCall } from '../utils/safe';
  import { GetAllItems, ExportToJSON, ExportToCSV, ImportFromJSON, CreateBackup } from '../wails/wailsjs/go/main/App';
  import { main } from '../wails/wailsjs/go/models';
  import { eventBus } from '../stores/events.svelte';

  let items = $state<main.ItemDTO[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);

  onMount(() => {
    loadData();
  });

  async function loadData() {
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

  // Calculate statistics
  const stats = $derived(() => {
    const total = items.length;

    // Count by health (we need to calculate this based on asset count)
    // For now, we'll use mock data since ItemDTO doesn't include health
    // In a real implementation, we would need to fetch ItemWithAssetsDTO for each item
    const byHealth = {
      secured: 0,
      partial: 0,
      incomplete: total // Assume all incomplete for now
    };

    // Count by category
    const byCategory: Record<string, number> = {};
    items.forEach(item => {
      byCategory[item.category] = (byCategory[item.category] || 0) + 1;
    });

    // Sort categories by count
    const sortedCategories = Object.entries(byCategory)
      .sort(([, a], [, b]) => b - a)
      .slice(0, 5); // Top 5 categories

    // Count by brand
    const byBrand: Record<string, number> = {};
    items.forEach(item => {
      byBrand[item.brand] = (byBrand[item.brand] || 0) + 1;
    });

    const topBrands = Object.entries(byBrand)
      .sort(([, a], [, b]) => b - a)
      .slice(0, 5);

    return {
      total,
      byHealth,
      byCategory: sortedCategories,
      topBrands,
      healthPercentages: {
        secured: total > 0 ? Math.round((byHealth.secured / total) * 100) : 0,
        partial: total > 0 ? Math.round((byHealth.partial / total) * 100) : 0,
        incomplete: total > 0 ? Math.round((byHealth.incomplete / total) * 100) : 0,
      }
    };
  });

  function getMaxCount(items: [string, number][]): number {
    return Math.max(...items.map(([, count]) => count), 1);
  }

  let exporting = $state(false);
  let importing = $state(false);
  let backing = $state(false);

  async function handleExportJSON() {
    exporting = true;
    const [err] = await safeCall(ExportToJSON());
    exporting = false;

    if (err) {
      // Error already handled by backend events
      return;
    }
  }

  async function handleExportCSV() {
    exporting = true;
    const [err] = await safeCall(ExportToCSV());
    exporting = false;

    if (err) {
      // Error already handled by backend events
      return;
    }
  }

  async function handleImportJSON() {
    importing = true;
    const [err] = await safeCall(ImportFromJSON());
    importing = false;

    if (err) {
      // Error already handled by backend events
      return;
    }

    // Reload data after import
    loadData();
  }

  async function handleCreateBackup() {
    backing = true;
    const [err] = await safeCall(CreateBackup());
    backing = false;

    if (err) {
      // Error already handled by backend events
      return;
    }
  }
</script>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-3">
      <div class="p-2 bg-primary/10 rounded-lg">
        <BarChart3 class="w-6 h-6 text-primary" />
      </div>
      <div>
        <h2 class="text-2xl font-bold">Tableau de bord</h2>
        <p class="text-sm text-muted-foreground">Vue d'ensemble de votre inventaire</p>
      </div>
    </div>

    <!-- Export/Import/Backup Actions -->
    <div class="flex gap-2">
      <button
        onclick={handleCreateBackup}
        disabled={backing || importing || exporting || loading}
        class="flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition disabled:opacity-50 disabled:cursor-not-allowed"
        title="Créer un backup"
      >
        <HardDrive class="w-4 h-4" />
        <span class="hidden lg:inline">Backup</span>
      </button>
      <div class="w-px bg-border"></div>
      <button
        onclick={handleImportJSON}
        disabled={importing || exporting || loading || backing}
        class="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition disabled:opacity-50 disabled:cursor-not-allowed"
        title="Importer depuis JSON"
      >
        <Upload class="w-4 h-4" />
        <span class="hidden sm:inline">Importer</span>
      </button>
      <button
        onclick={handleExportJSON}
        disabled={exporting || loading || importing || backing}
        class="flex items-center gap-2 px-4 py-2 border rounded-lg hover:bg-secondary transition disabled:opacity-50 disabled:cursor-not-allowed"
        title="Exporter en JSON"
      >
        <FileJson class="w-4 h-4" />
        <span class="hidden sm:inline">JSON</span>
      </button>
      <button
        onclick={handleExportCSV}
        disabled={exporting || loading || importing || backing}
        class="flex items-center gap-2 px-4 py-2 border rounded-lg hover:bg-secondary transition disabled:opacity-50 disabled:cursor-not-allowed"
        title="Exporter en CSV"
      >
        <FileSpreadsheet class="w-4 h-4" />
        <span class="hidden sm:inline">CSV</span>
      </button>
    </div>
  </div>

  {#if loading}
    <div class="flex items-center justify-center py-16">
      <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
    </div>
  {:else if error}
    <div class="bg-destructive/10 text-destructive px-4 py-3 rounded-lg">
      <p class="font-semibold">Erreur</p>
      <p class="text-sm">{error}</p>
    </div>
  {:else}
    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <!-- Total Items -->
      <div class="bg-card border rounded-lg p-5">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm font-medium text-muted-foreground">Total d'objets</span>
          <Package class="w-5 h-5 text-primary" />
        </div>
        <p class="text-3xl font-bold">{stats().total}</p>
        <p class="text-xs text-muted-foreground mt-1">Dans l'inventaire</p>
      </div>

      <!-- Secured Items -->
      <div class="bg-card border rounded-lg p-5">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm font-medium text-muted-foreground">🟢 Sécurisés</span>
          <CheckCircle class="w-5 h-5 text-green-600" />
        </div>
        <p class="text-3xl font-bold">{stats().byHealth.secured}</p>
        <p class="text-xs text-muted-foreground mt-1">{stats().healthPercentages.secured}% du total</p>
      </div>

      <!-- Partial Items -->
      <div class="bg-card border rounded-lg p-5">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm font-medium text-muted-foreground">🟡 Partiels</span>
          <AlertTriangle class="w-5 h-5 text-yellow-600" />
        </div>
        <p class="text-3xl font-bold">{stats().byHealth.partial}</p>
        <p class="text-xs text-muted-foreground mt-1">{stats().healthPercentages.partial}% du total</p>
      </div>

      <!-- Incomplete Items -->
      <div class="bg-card border rounded-lg p-5">
        <div class="flex items-center justify-between mb-2">
          <span class="text-sm font-medium text-muted-foreground">🔴 Incomplets</span>
          <AlertCircle class="w-5 h-5 text-red-600" />
        </div>
        <p class="text-3xl font-bold">{stats().byHealth.incomplete}</p>
        <p class="text-xs text-muted-foreground mt-1">{stats().healthPercentages.incomplete}% du total</p>
      </div>
    </div>

    <!-- Health Progress Bar -->
    <div class="bg-card border rounded-lg p-5">
      <h3 class="font-semibold mb-4 flex items-center gap-2">
        <TrendingUp class="w-5 h-5" />
        Santé documentaire globale
      </h3>
      <div class="space-y-3">
        <div class="h-8 bg-secondary rounded-lg overflow-hidden flex">
          {#if stats().byHealth.secured > 0}
            <div
              class="bg-green-500 flex items-center justify-center text-white text-xs font-medium"
              style="width: {stats().healthPercentages.secured}%"
            >
              {#if stats().healthPercentages.secured > 10}
                {stats().healthPercentages.secured}%
              {/if}
            </div>
          {/if}
          {#if stats().byHealth.partial > 0}
            <div
              class="bg-yellow-500 flex items-center justify-center text-white text-xs font-medium"
              style="width: {stats().healthPercentages.partial}%"
            >
              {#if stats().healthPercentages.partial > 10}
                {stats().healthPercentages.partial}%
              {/if}
            </div>
          {/if}
          {#if stats().byHealth.incomplete > 0}
            <div
              class="bg-red-500 flex items-center justify-center text-white text-xs font-medium"
              style="width: {stats().healthPercentages.incomplete}%"
            >
              {#if stats().healthPercentages.incomplete > 10}
                {stats().healthPercentages.incomplete}%
              {/if}
            </div>
          {/if}
        </div>
        <div class="flex items-center justify-between text-sm">
          <div class="flex items-center gap-2">
            <div class="w-3 h-3 bg-green-500 rounded"></div>
            <span>Sécurisé ({stats().byHealth.secured})</span>
          </div>
          <div class="flex items-center gap-2">
            <div class="w-3 h-3 bg-yellow-500 rounded"></div>
            <span>Partiel ({stats().byHealth.partial})</span>
          </div>
          <div class="flex items-center gap-2">
            <div class="w-3 h-3 bg-red-500 rounded"></div>
            <span>Incomplet ({stats().byHealth.incomplete})</span>
          </div>
        </div>
      </div>

      {#if stats().total > 0}
        <div class="mt-4 p-3 bg-secondary/30 rounded-lg">
          <p class="text-sm">
            <span class="font-medium">
              {stats().healthPercentages.secured + stats().healthPercentages.partial}%
            </span>
            de votre inventaire possède au moins un document.
            Continuez à ajouter de la documentation pour améliorer la résilience de vos objets!
          </p>
        </div>
      {/if}
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Categories Chart -->
      <div class="bg-card border rounded-lg p-5">
        <h3 class="font-semibold mb-4">Top catégories</h3>
        {#if stats().byCategory.length === 0}
          <p class="text-sm text-muted-foreground text-center py-8">Aucune donnée</p>
        {:else}
          <div class="space-y-3">
            {#each stats().byCategory as [category, count]}
              <div>
                <div class="flex items-center justify-between text-sm mb-1">
                  <span class="font-medium truncate">{category}</span>
                  <span class="text-muted-foreground ml-2">{count}</span>
                </div>
                <div class="h-2 bg-secondary rounded-full overflow-hidden">
                  <div
                    class="h-full bg-primary rounded-full transition-all duration-500"
                    style="width: {(count / getMaxCount(stats().byCategory)) * 100}%"
                  ></div>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Brands Chart -->
      <div class="bg-card border rounded-lg p-5">
        <h3 class="font-semibold mb-4">Top marques</h3>
        {#if stats().topBrands.length === 0}
          <p class="text-sm text-muted-foreground text-center py-8">Aucune donnée</p>
        {:else}
          <div class="space-y-3">
            {#each stats().topBrands as [brand, count]}
              <div>
                <div class="flex items-center justify-between text-sm mb-1">
                  <span class="font-medium truncate">{brand}</span>
                  <span class="text-muted-foreground ml-2">{count}</span>
                </div>
                <div class="h-2 bg-secondary rounded-full overflow-hidden">
                  <div
                    class="h-full bg-primary rounded-full transition-all duration-500"
                    style="width: {(count / getMaxCount(stats().topBrands)) * 100}%"
                  ></div>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </div>

    <!-- Info Box -->
    {#if stats().total === 0}
      <div class="bg-blue-50 border border-blue-200 text-blue-800 px-4 py-3 rounded-lg">
        <p class="font-semibold">👋 Bienvenue dans Brique!</p>
        <p class="text-sm mt-1">
          Commencez par ajouter votre premier objet pour voir vos statistiques ici.
          Brique vous aidera à préserver la documentation de vos objets pour les réparer plus facilement.
        </p>
      </div>
    {/if}
  {/if}
</div>
