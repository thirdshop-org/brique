# BRIQUE

L'infrastructure de résilience pour la réparation et l'entraide locale.

## Philosophie

Brique est conçu sur un postulat simple : **Internet n'est pas éternel, mais nos objets le sont.**

- **0% Cloud** : Aucune donnée n'est stockée sur un serveur distant centralisé
- **Local-First** : L'application est totalement fonctionnelle sans connexion réseau
- **Anti-Obsolescence** : Prolonger la durée de vie des objets en sécurisant les connaissances nécessaires à leur maintenance

## Structure du Projet

```
/brique
├── /cmd
│   ├── /brique-ui       # GUI avec Wails (à venir)
│   └── /brique-cli      # CLI Headless (implémenté)
├── /core                # Domaine métier
│   ├── /db              # Repository pattern avec sqlc
│   ├── /services        # Logique métier
│   └── /models          # Structs Go partagées
├── /frontend            # Svelte + Shadcn (à venir)
├── /migrations          # Migrations SQL avec goose
└── /pkg                 # Utils techniques
```

## Technologies

- **Backend** : Go 1.21+
- **Base de données** : SQLite (modernc.org/sqlite - pure Go)
- **Migrations** : goose
- **Génération SQL** : sqlc (type-safe)
- **Logging** : slog (standard library)
- **CLI** : cobra
- **Configuration** : viper

## Installation

### Prérequis

- Go 1.21 ou supérieur
- sqlc (pour la génération de code)

```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

### Build

```bash
# Clone le projet
cd brique

# Installer les dépendances
go mod download

# Générer le code sqlc
sqlc generate

# Build l'application CLI
go build -o brique ./cmd/brique-cli
```

## Utilisation

### CLI

**Gestion des Items:**

```bash
# Ajouter un item (mode interactif)
./brique item add

# Lister tous les items
./brique item list

# Voir les détails d'un item
./brique item get <id>

# Modifier un item
./brique item update <id>

# Supprimer un item
./brique item delete <id>

# Rechercher des items
./brique item search <query>
```

**Gestion des Assets (fichiers):**

```bash
# Ajouter un fichier à un item
./brique asset add <item-id> <file> --type manual --name "User Manual"

# Types supportés: manual, service_manual, exploded_view, stl, firmware, driver, schematic, other

# Lister les assets d'un item
./brique asset list <item-id>

# Supprimer un asset
./brique asset delete <asset-id>
```

**Exemple complet:**

```bash
# 1. Créer un item
./brique item add
# → Suivre les prompts interactifs

# 2. Ajouter des fichiers
./brique asset add 1 ~/Downloads/manual.pdf -t manual
./brique asset add 1 ~/Downloads/service.pdf -t service_manual

# 3. Voir le résultat
./brique item get 1
# → Affiche la santé documentaire: 🟢 Secured
```

### Stockage des données

Les données sont stockées dans :
- **Linux** : `~/.config/brique/`
- **Windows** : `%APPDATA%\Brique\`
- **macOS** : `~/Library/Application Support/Brique/`

Structure :
```
~/.config/brique/
├── brique.db           # Base de données SQLite
└── assets/            # Fichiers stockés (PDFs, STLs, etc.)
    └── item_<id>/     # Un dossier par item
```

## Module : Le Sac à Dos (Backpack)

Le premier module implémenté est le "Sac à Dos", qui permet de :

### Fonctionnalités actuelles (CLI complète)

**Items:**
- ✅ CRUD complet (Create, Read, Update, Delete)
- ✅ Recherche par nom, marque ou catégorie
- ✅ Vue détaillée avec santé documentaire

**Assets:**
- ✅ Ajout de fichiers avec type et nom personnalisé
- ✅ Listing détaillé avec tailles et hash
- ✅ Suppression sécurisée (DB + fichier physique)

**Santé documentaire:**
- ✅ 🟢 Secured : manuel + manuel de service présents
- ✅ 🟡 Partial : quelques fichiers présents
- ✅ 🔴 Incomplete : aucun fichier

### Champs d'un Item

- **Identité** : Nom, Catégorie, Marque, Modèle
- **Traçabilité** : Numéro de série, Date d'achat
- **Média** : Photo de l'objet
- **Notes** : Zone de texte libre

### Types d'Assets supportés

- Manuels utilisateurs (PDF)
- Manuels de service / Schémas techniques
- Vues éclatées (Exploded Views)
- Fichiers de fabrication (STL pour impression 3D)
- Firmwares / Drivers
- Autres

### Santé Documentaire

Chaque item affiche un statut de complétude :
- **Incomplete** : Aucun fichier stocké
- **Partial** : Quelques fichiers présents
- **Secured** : Manuel utilisateur ET manuel de service présents

## Développement

### Ajouter une migration

```bash
# Créer une nouvelle migration
cd migrations
# Créer 00003_nom_migration.sql avec:
-- +goose Up
-- +goose StatementBegin
-- Votre SQL ici
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Rollback SQL ici
-- +goose StatementEnd
```

### Ajouter des requêtes SQL

1. Créer/modifier un fichier dans `core/db/queries/`
2. Ajouter vos requêtes avec annotations sqlc :
```sql
-- name: GetSomething :one
SELECT * FROM table WHERE id = ?;
```
3. Régénérer le code : `sqlc generate`

## Progression

- ✅ **Étape 1** : Infrastructure + Module "Sac à Dos" (backend)
- ✅ **Étape 2** : CLI complète avec toutes les commandes
- 🚧 **Étape 3** : Interface graphique (Wails + Svelte)
- ⏳ **Étape 4** : Fonctionnalités avancées (QR codes, export/import)
- ⏳ **Étape 5** : Mode "Gossip Grids" (synchronisation P2P)

## Prochaines étapes

**Étape 3 : Interface Graphique**
- [ ] Initialiser le projet Wails
- [ ] Frontend Svelte avec Shadcn
- [ ] Écrans : Dashboard, détails, formulaires
- [ ] Drag & drop pour les assets
- [ ] Pattern "Safe Fetch" (tuple return)

**Fonctionnalités avancées:**
- [ ] Génération d'étiquettes QR Code
- [ ] Import/Export de données (JSON, CSV)
- [ ] Backup automatique
- [ ] Statistiques et rapports
- [ ] Mode headless pour Raspberry Pi

## License

À définir

## Contribution

Ce projet est en développement actif. Les contributions sont les bienvenues !
