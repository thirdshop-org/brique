# Guide de Démarrage - Interface Graphique Brique

## 🎨 L'Interface Graphique est prête !

Brique dispose maintenant d'une interface graphique moderne construite avec Wails, Svelte et Tailwind CSS.

## 🚀 Démarrage Rapide

### Prérequis

- Go 1.21+
- Node.js 18+
- Wails v2.11+

```bash
# Vérifier Wails
wails version

# Installer si nécessaire
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Mode Développement

```bash
# Lancer l'application en mode dev
wails dev
```

Cette commande va :
1. Compiler le backend Go
2. Démarrer le serveur Vite (frontend)
3. Lancer l'application avec hot-reload

L'application s'ouvrira dans une fenêtre native avec :
- ✅ Hot-reload du frontend (Vite)
- ✅ Hot-reload du backend (Wails)
- ✅ DevTools accessible (F12)

### Build Production

```bash
# Build l'application complète
wails build

# L'exécutable se trouve dans
./build/bin/brique        # Linux/macOS
./build/bin/brique.exe    # Windows
```

Options de build :
```bash
# Build en mode debug
wails build -debug

# Build pour une plateforme spécifique
wails build -platform darwin/amd64
wails build -platform windows/amd64

# Build avec optimisations
wails build -clean -upx
```

## 📱 Fonctionnalités Actuelles

### ✅ Implémenté

**Liste des Items**
- Affichage en grille responsive
- Cartes d'items avec :
  - Nom, marque, modèle
  - Catégorie
  - Numéro de série
  - Notes (tronquées)
  - Santé documentaire (🟢🟡🔴)

**Recherche en Temps Réel**
- Filtrage instantané
- Recherche sur nom, marque, catégorie
- Compteur de résultats

**États UI**
- Loading avec spinner
- Empty state avec message
- Error state avec détails
- Success avec grille d'items

### 🚧 À Venir

- Détail d'un item (modal)
- Formulaire d'ajout d'item
- Formulaire d'édition
- Dialog de confirmation de suppression
- Gestion des assets (liste, ajout drag & drop)
- Dashboard avec statistiques
- Dark mode toggle
- Export/Import de données

## 🎨 Design

### Thème

- **Palette** : Shadcn Slate
- **Radius** : 0.25rem (aspect "brique")
- **Police** : Inter (system-ui fallback)
- **Icons** : Lucide-svelte
- **Dark mode** : Supporté (à activer dans UI)

### Responsive

- **Mobile** : 1 colonne
- **Tablet** : 2 colonnes
- **Desktop** : 3 colonnes

## 🧪 Tester l'Application

### Avec des données existantes

Si vous avez déjà des items de l'étape 2 :

```bash
# Lancer l'UI
wails dev

# Les items devraient apparaître automatiquement
```

### Sans données

```bash
# Ajouter des items avec la CLI
./brique item add

# Puis lancer l'UI
wails dev
```

Ou utilisez le script de test :

```bash
./test_complete.sh

# Puis
wails dev
```

## 🔧 Développement

### Structure

```
/
├── main.go                    # Point d'entrée Wails
├── app_handlers.go            # API backend
├── wails.json                 # Config Wails
└── frontend/
    ├── src/
    │   ├── App.svelte         # Composant principal
    │   ├── main.ts            # Entry point
    │   └── lib/
    │       ├── components/    # Composants Svelte
    │       └── utils/         # Helpers (safe.ts)
    └── dist/                  # Build output
```

### Modifier le Frontend

1. Les fichiers Svelte sont dans `frontend/src/`
2. Les modifications sont hot-reloadées automatiquement
3. Les bindings Wails sont dans `frontend/src/lib/wails/`

### Ajouter une Méthode Backend

1. Ajouter la méthode dans `app_handlers.go` :
```go
func (a *App) MaNouvelleFonction(param string) (string, error) {
    // ...
    return result, nil
}
```

2. Les bindings TypeScript seront générés automatiquement
3. Utiliser dans le frontend :
```typescript
import { MaNouvelleFonction } from './lib/wails/go/main/App';
const [err, data] = await safeCall(MaNouvelleFonction("param"));
```

### Pattern Safe Fetch

**Toujours** utiliser le wrapper `safeCall()` :

```typescript
import { safeCall } from './lib/utils/safe';
import { GetAllItems } from './lib/wails/go/main/App';

// ✅ Bon
const [err, items] = await safeCall(GetAllItems());
if (err) {
  console.error(err);
  return;
}
// items est disponible

// ❌ Mauvais
const items = await GetAllItems(); // Pas de gestion d'erreur
```

## 🐛 Debugging

### DevTools

Appuyez sur `F12` pour ouvrir les DevTools Chrome dans l'application.

### Logs Backend

Les logs Go apparaissent dans le terminal où vous avez lancé `wails dev`.

### Logs Frontend

Les logs JavaScript/console apparaissent dans les DevTools.

### Problèmes Courants

**L'application ne démarre pas**
```bash
# Vérifier que le frontend build
cd frontend && npm run build

# Nettoyer et rebuild
wails build -clean
```

**Les bindings ne sont pas générés**
```bash
# Générer manuellement
wails generate module
```

**Hot-reload ne fonctionne pas**
- Relancer `wails dev`
- Vérifier que le port 5173 est libre

## 📊 Performance

### Build Size

- **Frontend** : 48KB JS + 11KB CSS (gzipped)
- **Backend** : ~15MB (Go binary)
- **Total** : ~15MB par plateforme

### Startup Time

- **Dev mode** : ~3-5 secondes
- **Production** : ~1-2 secondes

### Memory Usage

- **Idle** : ~50MB
- **Active** : ~100MB avec 100 items

## 🎯 Prochaines Fonctionnalités

### Prioritaires

1. **Détail d'Item** : Modal avec toutes les infos + assets
2. **Ajout d'Item** : Formulaire avec validation
3. **Édition** : Modifier un item existant
4. **Assets** : Drag & drop pour ajouter des fichiers

### Nice to Have

- Dashboard avec statistiques
- Export/Import (JSON, CSV)
- Dark mode toggle
- Keyboard shortcuts
- Animations de transition
- Toast notifications

## 📚 Ressources

- **Wails Docs** : https://wails.io/docs/intro
- **Svelte Tutorial** : https://learn.svelte.dev
- **Tailwind CSS** : https://tailwindcss.com/docs
- **Lucide Icons** : https://lucide.dev/icons

## 💡 Tips

### 1. Rechargement Rapide

En dev mode, les modifications Svelte sont appliquées instantanément. Les modifications Go nécessitent une recompilation (~1s).

### 2. Build Optimisé

Utilisez `-upx` pour compresser le binaire :
```bash
wails build -clean -upx
```

### 3. Distribution

Pour distribuer l'app :
- **Windows** : `build/bin/brique.exe` (installer avec NSIS)
- **macOS** : `build/bin/brique.app` (bundle)
- **Linux** : `build/bin/brique` (AppImage ou deb/rpm)

## 🎊 Conclusion

L'interface graphique transforme Brique en une application moderne et accessible !

**Commandes essentielles :**
```bash
# Développement
wails dev

# Build
wails build

# Nettoyer
wails build -clean
```

**Bon développement !** 🚀
