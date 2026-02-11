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

```bash
# Lister les items dans l'inventaire
./brique item list

# Ajouter un item (mode interactif)
./brique item add
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

### Fonctionnalités actuelles

- ✅ Créer un item dans l'inventaire
- ✅ Lister tous les items
- ✅ Rechercher des items
- ✅ Ajouter des assets (fichiers) à un item
- ✅ Calculer la "santé documentaire" d'un item

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

## Prochaines étapes

- [ ] Interface graphique (Wails + Svelte)
- [ ] Mode "Gossip Grids" (synchronisation P2P)
- [ ] Génération d'étiquettes QR Code
- [ ] Import/Export de données
- [ ] Mode headless pour Raspberry Pi

## License

À définir

## Contribution

Ce projet est en développement actif. Les contributions sont les bienvenues !
