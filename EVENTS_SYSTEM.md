# Système d'Événements - Bus de Communication Frontend ↔ Backend

## Vue d'ensemble

Le système d'événements permet au backend Go d'envoyer des notifications et des mises à jour de progression en temps réel au frontend Svelte, sans que le frontend ait besoin de faire du polling.

## Architecture

### Backend (Go)

**Fichier : `events.go`**

Le `EventEmitter` fournit des méthodes pour émettre des événements vers le frontend via Wails runtime:

```go
type EventEmitter struct {
    ctx context.Context
}

// Types d'événements
- notification: Notifications success/error/info/warning
- progress: Barres de progression pour opérations longues
- progress:complete: Signal de fin d'opération
```

**Utilisation dans les handlers :**

```go
// Notification simple
a.events.Success("Item créé", "L'item a été ajouté")
a.events.Error("Erreur", "Quelque chose s'est mal passé")

// Progression
progressID := "upload-123"
a.events.EmitProgress(ProgressData{
    ID: progressID,
    Operation: "Upload fichier",
    Current: 50,
    Total: 100,
    Filename: "manual.pdf",
})
a.events.EmitProgressComplete(progressID)
```

### Frontend (Svelte)

**Fichier : `frontend/src/lib/stores/events.svelte.ts`**

Store Svelte 5 avec runes qui gère les événements du backend:

```typescript
import { eventBus } from './lib/stores/events.svelte';

// Notifications
eventBus.notifications // Array<Notification>
eventBus.progressEvents // Map<string, ProgressEvent>

// Méthodes
eventBus.success("Titre", "Message")
eventBus.error("Titre", "Message")
eventBus.info("Titre", "Message")
eventBus.warning("Titre", "Message")
```

**Composants UI :**

1. **NotificationToast.svelte** : Affiche les notifications en haut à droite
   - Auto-dismiss configurable
   - Icônes par type (success/error/warning/info)
   - Animation slide-in
   - Bouton de fermeture manuelle

2. **ProgressBar.svelte** : Affiche les barres de progression en bas à droite
   - Pourcentage et valeurs current/total
   - Nom du fichier en cours
   - Animation smooth

## Événements Disponibles

### Notifications

Émis par le backend pour informer l'utilisateur :

| Type | Durée par défaut | Icône | Couleur |
|------|------------------|-------|---------|
| success | 5s | ✓ | Vert |
| error | 8s | ✗ | Rouge |
| warning | 6s | ⚠ | Jaune |
| info | 5s | ℹ | Bleu |

### Progression

Émis pour les opérations longues (upload, sync, etc.) :

```typescript
{
  id: string,          // Identifiant unique
  operation: string,   // Nom de l'opération
  current: number,     // Progression actuelle
  total: number,       // Total
  filename?: string    // Fichier en cours (optionnel)
}
```

## Intégration dans l'Application

### App.svelte

```svelte
<script>
  import NotificationToast from './lib/components/NotificationToast.svelte';
  import ProgressBar from './lib/components/ProgressBar.svelte';
  import { eventBus } from './lib/stores/events.svelte';

  onMount(() => {
    // Cleanup on unmount
    return () => {
      eventBus.destroy();
    };
  });
</script>

<div>
  <!-- App content -->
</div>

<!-- Global UI Components -->
<NotificationToast />
<ProgressBar />
```

### Handlers Go

Tous les handlers ont été mis à jour pour émettre des notifications :

- `CreateItem` : Notification de succès/erreur
- `UpdateItem` : Notification de mise à jour
- `DeleteItem` : Notification de suppression
- `AddAsset` : Progression + notification
- `DeleteAsset` : Notification de suppression

## Cas d'Usage

### 1. Upload de fichier avec progression

```go
// Backend
progressID := fmt.Sprintf("upload-%d", itemID)
a.events.EmitProgress(ProgressData{
    ID: progressID,
    Operation: "Upload fichier",
    Current: 0,
    Total: fileSize,
    Filename: filename,
})

// ... upload logic ...

a.events.EmitProgress(ProgressData{
    ID: progressID,
    Operation: "Upload fichier",
    Current: bytesUploaded,
    Total: fileSize,
    Filename: filename,
})

a.events.EmitProgressComplete(progressID)
a.events.Success("Fichier ajouté", filename)
```

### 2. Notifications simples

```go
// Backend
a.events.Success("Opération réussie", "Les données ont été sauvegardées")
a.events.Error("Erreur réseau", "Impossible de se connecter au serveur")
a.events.Warning("Attention", "L'espace disque est faible")
a.events.Info("Information", "Synchronisation terminée")
```

### 3. Écoute côté frontend (pour actions custom)

```typescript
// Frontend
import { EventsOn } from './lib/wails/wailsjs/runtime/runtime';

EventsOn('custom-event', (data) => {
  console.log('Event reçu:', data);
});
```

## Avantages

✅ **Temps réel** : Pas de polling, communication instantanée
✅ **UX améliorée** : Feedback immédiat pour l'utilisateur
✅ **Réutilisable** : Store Svelte 5 réactif
✅ **Type-safe** : Interfaces TypeScript et Go
✅ **Extensible** : Facile d'ajouter de nouveaux types d'événements

## Prochaines Améliorations

- [ ] Historique des notifications (avec limite)
- [ ] Notifications persistantes (ne se ferment pas automatiquement)
- [ ] Actions dans les notifications (boutons)
- [ ] Sons pour les notifications critiques
- [ ] Regroupement de notifications similaires
- [ ] Progression P2P pour la synchronisation Gossip Grids
