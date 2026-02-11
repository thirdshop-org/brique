# Guide de Démarrage Rapide - BRIQUE

## 🎉 L'Étape 1 est terminée !

Le module "Sac à Dos" (inventaire personnel) est complètement opérationnel avec :
- Base de données SQLite avec migrations automatiques
- Service métier complet avec gestion des items et assets
- CLI fonctionnelle
- Tests unitaires (6/6 passent ✅)

## 🚀 Utilisation immédiate

### 1. Build l'application

```bash
go build -o brique ./cmd/brique-cli
```

### 2. Commandes disponibles

```bash
# Lister tous les items
./brique item list

# Ajouter un item (mode interactif)
./brique item add
```

### 3. Exemple d'utilisation

```bash
$ ./brique item add
Name: Lave-Linge Cuisine
Category: Gros Électroménager
Brand: Brandt
Model: WTC1234

Item created successfully with ID: 1

$ ./brique item list

Inventory (1 items):

ID: 1
  Name: Lave-Linge Cuisine
  Category: Gros Électroménager
  Brand: Brandt
  Model: WTC1234
```

## 📦 Où sont stockées les données ?

- **Linux** : `~/.config/brique/`
- **Windows** : `%APPDATA%\Brique\`
- **macOS** : `~/Library/Application Support/Brique/`

Structure :
```
~/.config/brique/
├── brique.db           # Base de données SQLite
└── assets/            # Fichiers (PDFs, STLs, etc.)
    └── item_1/        # Un dossier par item
```

## 🧪 Lancer les tests

```bash
go test ./core/services/... -v
```

Tous les tests devraient passer :
- ✅ TestCreateItem
- ✅ TestGetAllItems
- ✅ TestSearchItems
- ✅ TestUpdateItem
- ✅ TestAddAsset
- ✅ TestDocumentationHealth

## 📚 Architecture technique

### Stack

- **Go 1.21+** : Backend
- **SQLite** : Base de données (modernc.org/sqlite - pure Go)
- **sqlc** : Génération de code type-safe
- **goose** : Migrations
- **cobra** : CLI
- **viper** : Configuration

### Philosophie du code

1. **Type-safe** : sqlc génère du code Go à partir de SQL
2. **Testable** : Architecture en couches (models, db, services, cmd)
3. **Offline-First** : Tout fonctionne sans Internet
4. **Graceful** : Context pour l'annulation, shutdown propre
5. **Robuste** : Gestion d'erreurs, logging structuré

### Structure des packages

```
core/          # Logique métier (agnostique de l'interface)
  models/      # Structures de données partagées
  db/          # Accès base de données (sqlc generated)
  services/    # Logique métier (Backpack, etc.)
cmd/           # Points d'entrée (CLI, UI)
pkg/           # Utilitaires techniques
migrations/    # Migrations SQL versionnées
```

## 🔨 Développement

### Ajouter une migration

```bash
# Créer un fichier dans migrations/
# Format : 00003_nom_descriptif.sql

-- +goose Up
-- +goose StatementBegin
CREATE TABLE ma_table (...);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE ma_table;
-- +goose StatementEnd
```

Les migrations s'exécutent automatiquement au démarrage.

### Ajouter des requêtes SQL

1. Modifier/créer un fichier dans `core/db/queries/`
2. Ajouter vos requêtes avec annotations sqlc :

```sql
-- name: GetSomething :one
SELECT * FROM items WHERE id = ?;

-- name: ListSomething :many
SELECT * FROM items ORDER BY name;

-- name: UpdateSomething :exec
UPDATE items SET name = ? WHERE id = ?;
```

3. Régénérer le code :
```bash
sqlc generate
```

Le code Go type-safe est généré automatiquement !

### Ajouter un service

Les services sont dans `core/services/` et contiennent la logique métier.

Exemple pattern :

```go
type MonService struct {
    queries *db.Queries
}

func NewMonService(queries *db.Queries) *MonService {
    return &MonService{queries: queries}
}

func (s *MonService) FaireQuelqueChose(ctx context.Context) error {
    // Utiliser s.queries pour accéder à la DB
    return nil
}
```

### Ajouter une commande CLI

Dans `cmd/brique-cli/main.go` :

```go
maCmd := &cobra.Command{
    Use:   "ma-commande",
    Short: "Description",
    RunE:  runMaCommande,
}

rootCmd.AddCommand(maCmd)
```

## 🎯 Prochaines étapes suggérées

### Étape 2A : Compléter la CLI

Ajouter les commandes manquantes :
- `item get <id>` : voir un item en détail
- `item update <id>` : modifier un item
- `item delete <id>` : supprimer un item
- `item search <query>` : rechercher
- `asset add <item-id> <file>` : ajouter un fichier
- `asset list <item-id>` : lister les fichiers d'un item

### Étape 2B : Interface graphique (Wails)

1. Installer Wails : https://wails.io
2. Initialiser le frontend Svelte
3. Implémenter les écrans :
   - Liste des items (grille/tableau)
   - Détail d'un item
   - Formulaire ajout/édition
   - Gestion des assets avec drag & drop

### Étape 3 : Fonctionnalités avancées

- Génération de QR codes pour les étiquettes
- Export/Import de données
- Backup automatique
- Statistiques

### Étape 4 : "Gossip Grids" (P2P)

- Synchronisation locale (LAN)
- Mode Sneakernet (USB)
- Synchronisation Internet

## 📖 Documentation

- `README.md` : Vue d'ensemble du projet
- `STATUS.md` : État détaillé de l'implémentation
- `PROJECT.md` : Vision et philosophie
- `REQUIRED.md` : Spécifications techniques
- `FIRST_STEP.md` : Cahier des charges du module Sac à Dos

## 💡 Conseils

1. **Commits fréquents** : Git init + commits réguliers
2. **Tests d'abord** : Écrire les tests avant le code
3. **Logs structurés** : Utiliser slog avec des attributs
4. **Contexts partout** : Pour l'annulation et les timeouts
5. **Erreurs wrappées** : fmt.Errorf("context: %w", err)

## 🐛 Debugging

### Voir les logs en détail

Modifier le niveau de log dans `cmd/brique-cli/main.go` :

```go
logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,  // Debug au lieu de Info
}))
```

### Inspecter la base de données

```bash
sqlite3 ~/.config/brique/brique.db

sqlite> .tables
sqlite> SELECT * FROM items;
sqlite> .schema items
```

### Supprimer toutes les données (reset)

```bash
rm -rf ~/.config/brique/
```

## 🎊 Félicitations !

Vous avez maintenant une base solide pour Brique. Le code est :
- ✅ Type-safe (sqlc)
- ✅ Testé (6 tests unitaires)
- ✅ Documenté
- ✅ Structuré (architecture en couches)
- ✅ Robuste (gestion d'erreurs, logging)
- ✅ Offline-first (tout fonctionne localement)

**Bon développement ! 🚀**
