# BRIQUE - Étape 3 : Interface Graphique (En cours)

Date : 11 février 2026

## 🎯 Objectif Étape 3

Créer une interface graphique moderne avec Wails + Svelte + Shadcn pour remplacer/compléter la CLI.

## ✅ Infrastructure Complétée

### Wails Setup

**Configuration (wails.json):**
```json
{
  "name": "Brique",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "wailsjsdir": "./frontend/src/lib/wails",
  "assetdir": "./frontend/dist"
}
```

**Point d'entrée (main.go):**
- Hook `startup` : initialise database + service
- Hook `shutdown` : ferme proprement la database
- Hook `domReady` : appelé quand le DOM est prêt
- Hook `beforeClose` : peut empêcher la fermeture

**Handlers (app_handlers.go):**
- `GetAllItems()` : liste tous les items
- `GetItem(id)` : détails d'un item
- `GetItemWithAssets(id)` : item + assets + santé
- `CreateItem(...)` : créer un item
- `UpdateItem(...)` : modifier un item
- `DeleteItem(id)` : supprimer un item
- `SearchItems(query)` : rechercher
- `GetAssets(itemID)` : liste des assets
- `AddAsset(...)` : ajouter un asset
- `DeleteAsset(id)` : supprimer un asset

**DTOs:**
- `ItemDTO` : données d'un item pour le frontend
- `AssetDTO` : données d'un asset pour le frontend
- `ItemWithAssetsDTO` : item + assets + santé

### Frontend Svelte + TypeScript

**Stack:**
- Svelte 5 (avec runes)
- TypeScript
- Vite (bundler)
- Tailwind CSS
- Lucide-svelte (icons)

**Configuration:**
- `vite.config.ts` : bundler Vite
- `tsconfig.json` : TypeScript
- `tailwind.config.js` : Tailwind avec thème Shadcn
- `postcss.config.js` : PostCSS + Autoprefixer
- `svelte.config.js` : Svelte preprocessing

**Thème Shadcn:**
- Variables CSS pour colors, border-radius, etc.
- Support dark mode (`.dark` class)
- Palette Slate/Gray comme spécifié
- Radius: 0.25rem (aspect "brique")

**Pattern "Safe Fetch" (REQUIRED.md):**

Fichier : `frontend/src/lib/utils/safe.ts`

```typescript
type SafeResult<T> = Promise<[Error, null] | [null, T]>;

export async function safeCall<T>(promise: Promise<T>): SafeResult<T> {
    try {
        const data = await promise;
        return [null, data];
    } catch (err) {
        const error = err instanceof Error ? err : new Error(String(err));
        return [error, null];
    }
}
```

Usage :
```typescript
const [err, items] = await safeCall(GetAllItems());
if (err) {
  // Handle error
}
// Use items
```

## ✅ Composants Implémentés

### App.svelte (Composant Principal)

**Fonctionnalités:**
- Header avec logo et titre
- Barre de recherche en temps réel
- Bouton "Ajouter"
- Grille d'items responsive (1/2/3 colonnes)
- États : loading, error, empty, success
- Compteur d'items
- Filtrage local côté client

**Utilisation du Safe Fetch:**
```typescript
const [err, data] = await safeCall(GetAllItems());
if (err) {
  error = err.message;
  return;
}
items = data || [];
```

**Svelte 5 Runes:**
- `$state` pour les états réactifs
- `$derived` pour les valeurs calculées (filteredItems)

### ItemCard.svelte (Carte d'Item)

**Affichage:**
- Icon Package
- Nom de l'item
- Marque + Modèle
- Catégorie (badge)
- Numéro de série (si présent)
- Notes (tronquées à 2 lignes)
- Badge de santé documentaire

**Santé Documentaire:**
- 🟢 Sécurisé (vert) : manuel + manuel de service
- 🟡 Partiel (jaune) : quelques fichiers
- 🔴 Incomplet (rouge) : aucun fichier

**Hover:**
- Effet shadow au survol
- Cursor pointer
- Transition smooth

## 📊 Structure des Fichiers

```
/
├── main.go                      # Point d'entrée Wails
├── app_handlers.go              # Handlers exposés au frontend
├── wails.json                   # Configuration Wails
└── frontend/
    ├── package.json
    ├── vite.config.ts
    ├── tsconfig.json
    ├── tailwind.config.js
    ├── postcss.config.js
    ├── svelte.config.js
    ├── index.html
    ├── src/
    │   ├── main.ts              # Point d'entrée
    │   ├── App.svelte           # Composant principal
    │   ├── app.css              # Styles globaux + Tailwind
    │   └── lib/
    │       ├── utils/
    │       │   └── safe.ts      # Pattern Safe Fetch
    │       ├── components/
    │       │   └── ItemCard.svelte
    │       └── wails/
    │           └── go/main/
    │               └── App.js   # Bindings générés
    └── dist/                    # Build output
```

## 🎨 Design System

### Couleurs

Basé sur Shadcn avec palette Slate :
- **Primary** : noir/blanc selon le mode
- **Secondary** : gris clair/foncé
- **Muted** : gris très clair
- **Destructive** : rouge pour les actions dangereuses
- **Border** : gris clair pour les bordures

### Typographie

- Police : Inter (ou system-ui en fallback)
- Pas de police mono sauf pour données techniques

### Espacement

- Padding : 4, 6, 8px
- Gap : 4, 6px
- Radius : 0.25rem (4px) - aspect carré/brique

### Composants Shadcn

Pas encore implémentés mais prévus :
- Button
- Input
- Dialog
- Select
- Card (déjà stylé manuellement)
- Badge (déjà stylé manuellement)

## 🚀 Démarrage

### Développement

```bash
# Démarrer l'app en mode dev
wails dev

# Ou séparément :
# Terminal 1 - Frontend
cd frontend && npm run dev

# Terminal 2 - Backend
go run .
```

### Build Production

```bash
# Build l'app complète
wails build

# Ou manuellement :
cd frontend && npm run build
go build -o brique-ui .
```

## 📝 Prochaines étapes

### À implémenter

- [ ] Écran de détail d'un item (modal ou page)
- [ ] Formulaire d'ajout d'item
- [ ] Formulaire d'édition d'item
- [ ] Dialog de confirmation pour la suppression
- [ ] Gestion des assets :
  - [ ] Liste des assets d'un item
  - [ ] Ajout d'asset (drag & drop)
  - [ ] Aperçu de fichier
  - [ ] Suppression d'asset
- [ ] Recherche avancée (filtres)
- [ ] Tri des items (date, nom, santé)
- [ ] Vue grille / liste toggle
- [ ] Dashboard avec statistiques
- [ ] Dark mode toggle
- [ ] Paramètres de l'app
- [ ] Export/Import de données

### Composants à créer

- [ ] ItemDetail.svelte (détail complet)
- [ ] ItemForm.svelte (add/edit)
- [ ] AssetList.svelte (liste des assets)
- [ ] AssetCard.svelte (carte d'asset)
- [ ] FileUpload.svelte (drag & drop)
- [ ] ConfirmDialog.svelte (confirmation)
- [ ] SearchBar.svelte (recherche avancée)
- [ ] Stat Card.svelte (statistiques)
- [ ] ThemeToggle.svelte (dark mode)

### Améliorations UX

- [ ] Animations de transition
- [ ] Toast notifications
- [ ] Loading skeletons
- [ ] Infinite scroll / pagination
- [ ] Keyboard shortcuts
- [ ] Tooltips
- [ ] Empty states améliorés
- [ ] Error boundaries

## 🧪 Tests

### À tester

- [ ] Création d'item via l'UI
- [ ] Modification d'item via l'UI
- [ ] Suppression d'item via l'UI
- [ ] Recherche en temps réel
- [ ] Ajout d'asset via drag & drop
- [ ] Calcul de la santé documentaire
- [ ] Responsive design (mobile/tablet/desktop)
- [ ] Dark mode
- [ ] Gestion des erreurs
- [ ] Performance avec beaucoup d'items (100+)

## 💡 Notes Techniques

### Bindings Wails

Les bindings TypeScript sont générés automatiquement par Wails depuis les méthodes Go exportées de l'App struct.

Fichier généré : `frontend/src/lib/wails/go/main/App.js`

Chaque méthode Go devient une fonction TypeScript async qui retourne une Promise.

### Context dans Wails

Le context passé à `startup()` est stocké dans `a.ctx` et utilisé pour tous les appels au backpack service. Cela permet l'annulation et les timeouts.

### Event System

Wails fournit un système d'événements pour la communication bidirectionnelle :
- `runtime.EventsEmit()` : émettre depuis Go
- `runtime.EventsOn()` : écouter depuis JS

Utile pour :
- Progression de téléchargement
- Notifications
- Updates en temps réel

### File Picker

Wails fournit des dialogs natifs :
- `runtime.OpenFileDialog()` : sélection de fichier
- `runtime.SaveFileDialog()` : sauvegarde de fichier
- `runtime.SelectDirectoryDialog()` : sélection de dossier

## 📚 Documentation

### Ressources

- Wails : https://wails.io
- Svelte 5 : https://svelte.dev
- Shadcn : https://www.shadcn-svelte.com
- Tailwind : https://tailwindcss.com
- Lucide : https://lucide.dev

### Patterns à suivre

1. **Safe Fetch** : toujours utiliser `safeCall()` pour les appels Wails
2. **Svelte Runes** : utiliser `$state`, `$derived`, `$effect`
3. **TypeScript** : typer toutes les interfaces
4. **Composants** : petits, réutilisables, bien nommés
5. **Styles** : Tailwind classes, pas de CSS inline

## ✨ Conclusion (provisoire)

L'infrastructure de base de l'UI est en place ! Le projet est maintenant capable de :
- Afficher la liste des items
- Rechercher en temps réel
- Afficher la santé documentaire

Les fondations sont solides pour ajouter toutes les autres fonctionnalités.

**L'étape 3 continue...** 🚀
