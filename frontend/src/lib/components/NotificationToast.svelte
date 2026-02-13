<script lang="ts">
  import { eventBus, type Notification } from '../stores/events.svelte';
  import { CheckCircle, XCircle, Info, AlertTriangle, X } from 'lucide-svelte';

  function getIcon(type: Notification['type']) {
    switch (type) {
      case 'success': return CheckCircle;
      case 'error': return XCircle;
      case 'warning': return AlertTriangle;
      case 'info': return Info;
    }
  }

  function getColorClasses(type: Notification['type']) {
    switch (type) {
      case 'success': return 'bg-green-50 border-green-200 text-green-900 dark:bg-green-900/10 dark:border-green-800 dark:text-green-100';
      case 'error': return 'bg-red-50 border-red-200 text-red-900 dark:bg-red-900/10 dark:border-red-800 dark:text-red-100';
      case 'warning': return 'bg-yellow-50 border-yellow-200 text-yellow-900 dark:bg-yellow-900/10 dark:border-yellow-800 dark:text-yellow-100';
      case 'info': return 'bg-blue-50 border-blue-200 text-blue-900 dark:bg-blue-900/10 dark:border-blue-800 dark:text-blue-100';
    }
  }

  function getIconColorClass(type: Notification['type']) {
    switch (type) {
      case 'success': return 'text-green-600 dark:text-green-400';
      case 'error': return 'text-red-600 dark:text-red-400';
      case 'warning': return 'text-yellow-600 dark:text-yellow-400';
      case 'info': return 'text-blue-600 dark:text-blue-400';
    }
  }

  function dismiss(id: string) {
    eventBus.removeNotification(id);
  }
</script>

<!-- Toast Container -->
<div class="fixed top-4 right-4 z-50 flex flex-col gap-3 w-96 max-w-[calc(100vw-2rem)]">
  {#each eventBus.notifications as notification (notification.id)}
    <div
      class="border rounded-lg shadow-lg p-4 flex items-start gap-3 transition-all duration-300 ease-in-out animate-slide-in {getColorClasses(notification.type)}"
    >
      <!-- Icon -->
      <svelte:component
        this={getIcon(notification.type)}
        class="w-5 h-5 flex-shrink-0 mt-0.5 {getIconColorClass(notification.type)}"
      />

      <!-- Content -->
      <div class="flex-1 min-w-0">
        <p class="font-semibold text-sm">{notification.title}</p>
        {#if notification.message}
          <p class="text-sm mt-1 opacity-90">{notification.message}</p>
        {/if}
      </div>

      <!-- Dismiss button -->
      <button
        onclick={() => dismiss(notification.id)}
        class="flex-shrink-0 p-1 rounded hover:bg-black/5 dark:hover:bg-white/5 transition"
        aria-label="Dismiss"
      >
        <X class="w-4 h-4" />
      </button>
    </div>
  {/each}
</div>

<style>
  @keyframes slide-in {
    from {
      transform: translateX(100%);
      opacity: 0;
    }
    to {
      transform: translateX(0);
      opacity: 1;
    }
  }

  .animate-slide-in {
    animation: slide-in 0.3s ease-out;
  }
</style>
