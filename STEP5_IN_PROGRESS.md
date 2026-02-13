# Étape 5 : Module "Gossip Grids" - EN COURS

Date de début : 13 février 2026

## Vue d'ensemble

Le module "Gossip Grids" permettra la synchronisation décentralisée des inventaires entre plusieurs instances de Brique, suivant la philosophie "local-first" du projet.

## Travail complété

### ✅ 1. Document de conception

**Fichier:** `GOSSIP_GRIDS_DESIGN.md`

- Architecture complète du module
- Cas d'usage détaillés
- Protocole de synchronisation
- Stratégie de résolution de conflits (Last-Write-Wins)
- Structure des données
- Modèle de confiance et sécurité
- Plan d'implémentation en 3 phases

### ✅ 2. Migrations de base de données

**Créées:**
- `00003_create_peers_table.sql` : Table des pairs découverts
- `00004_create_sync_tables.sql` : Table sync_logs + colonnes de tracking

**Schéma ajouté:**

```sql
-- Peers (instances Brique distantes)
CREATE TABLE peers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    address TEXT NOT NULL,        -- IP:Port
    last_seen TIMESTAMP,
    last_sync TIMESTAMP,
    is_trusted BOOLEAN DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Sync logs (historique de synchronisation)
CREATE TABLE sync_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    peer_id TEXT NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    items_received INTEGER DEFAULT 0,
    items_sent INTEGER DEFAULT 0,
    conflicts INTEGER DEFAULT 0,
    duration_ms INTEGER DEFAULT 0,
    error TEXT,
    FOREIGN KEY (peer_id) REFERENCES peers(id)
);

-- Tracking sur les items
ALTER TABLE items ADD COLUMN origin_peer_id TEXT;
ALTER TABLE items ADD COLUMN sync_version INTEGER DEFAULT 1;
```

### ✅ 3. Modèles de données

**Fichier:** `core/models/peer.go`

**Types créés:**
- `PeerStatus` : enum (online, offline, syncing)
- `Peer` : représente une instance Brique distante
- `SyncResult` : résultat d'une synchronisation
- `SyncLog` : entrée d'historique de sync
- `ChangeSet` : ensemble de changements à synchroniser
- `SyncInfo` : informations d'instance

### ✅ 4. Requêtes SQL (sqlc)

**Fichiers créés:**
- `core/db/queries/peers.sql` : CRUD complet sur les pairs
- `core/db/queries/sync_logs.sql` : Gestion des logs de sync
- Ajout dans `core/db/queries/items.sql` : GetItemsModifiedSince, CountItems

**Requêtes générées:**
- 10 requêtes pour la gestion des peers
- 5 requêtes pour les sync logs
- 2 nouvelles requêtes pour les items

### ✅ 5. Service GossipService

**Fichier:** `core/services/gossip_service.go` (~260 lignes)

**Fonctionnalités implémentées:**

**Gestion des pairs:**
- `GetInstanceInfo()` : Informations sur l'instance locale
- `AddPeer()` : Ajouter un pair découvert
- `GetPeers()` : Liste tous les pairs
- `GetTrustedPeers()` : Liste les pairs approuvés
- `UpdatePeerLastSeen()` : Mise à jour du heartbeat
- `UpdatePeerLastSync()` : Mise à jour après sync
- `SetPeerTrust()` : Approuver/révoquer un pair
- `RemovePeer()` : Supprimer un pair

**Synchronisation:**
- `GetChanges()` : Récupère les items modifiés depuis un timestamp
- `LogSync()` : Enregistre un événement de sync
- `GetSyncHistory()` : Historique de sync par pair
- `GetRecentSyncHistory()` : Dernières sync globales

**Helpers:**
- `dbPeerToModel()` : Conversion DB → modèle
- `dbSyncLogToModel()` : Conversion DB → modèle
- `dbItemToModel()` : Conversion DB → modèle

**Pattern:**
- Suit le même modèle que BackpackService
- Utilise sqlc pour la sécurité des types
- Gestion propre des types sql.Null*
- Context-aware pour cancellation

## Travail restant

### 🚧 Tâche #12: Découverte de pairs (LAN)

**À implémenter:**
- Intégration mDNS avec `github.com/hashicorp/mdns`
- Broadcast automatique de l'instance sur le LAN
- Écoute des annonces d'autres instances
- Mise à jour automatique de la table `peers`
- Gestion du heartbeat (keep-alive)

**Fichiers à créer:**
- `core/services/discovery_service.go`
- Méthodes: `StartDiscovery()`, `StopDiscovery()`, `AnnounceInstance()`

### 🚧 Tâche #13: Protocole de synchronisation

**À implémenter:**
- API HTTP REST dans `main.go` ou handler dédié
- Endpoints:
  - `GET /api/v1/gossip/info` : Informations d'instance
  - `POST /api/v1/gossip/sync` : Déclencher une sync
  - `GET /api/v1/gossip/items/changes?since=<ts>` : Changements
  - `POST /api/v1/gossip/items/batch` : Recevoir des items

**Logique de synchronisation:**
1. Comparer les timestamps
2. Récupérer les changements depuis dernier sync
3. Envoyer les items modifiés
4. Recevoir et appliquer les changements distants
5. Résoudre les conflits (Last-Write-Wins)
6. Logger le résultat

**Fichiers à créer:**
- `gossip_handlers.go` : Handlers HTTP
- Méthodes dans `GossipService`: `SyncWithPeer()`, `ApplyChanges()`, `ResolveConflicts()`

### 🚧 Tâche #14: Interface utilisateur

**À créer:**

**Composant: SyncView.svelte**
- Onglet "Synchronisation" dans la navigation
- Sections:
  - Liste des pairs découverts (avec statut online/offline)
  - Bouton "Synchroniser" par pair
  - Liste des pairs approuvés
  - Historique de synchronisation (timeline)
  - Configuration (nom d'instance, activer/désactiver)

**Composants auxiliaires:**
- `PeerCard.svelte` : Carte représentant un pair
- `SyncHistoryItem.svelte` : Entrée d'historique
- `SyncProgress.svelte` : Barre de progression pendant sync

**Handlers Wails à exposer:**
- `GetPeers() ([]PeerDTO, error)`
- `SyncWithPeer(peerID string) (*SyncResultDTO, error)`
- `SetPeerTrust(peerID string, trusted bool) error`
- `GetSyncHistory(limit int) ([]SyncLogDTO, error)`

**Modifications à App.svelte:**
- Ajouter l'onglet "Synchronisation" dans la navigation
- Router vers SyncView

## Dépendances à ajouter

```bash
# Pour la découverte mDNS
go get github.com/hashicorp/mdns
```

## Prochaines étapes

1. ✅ ~~Concevoir l'architecture~~
2. ✅ ~~Créer les migrations et modèles~~
3. ✅ ~~Implémenter GossipService~~
4. [ ] Implémenter la découverte mDNS
5. [ ] Créer l'API HTTP de synchronisation
6. [ ] Développer l'UI de gestion des pairs
7. [ ] Tests d'intégration (2+ instances locales)

## Tests suggérés

### Test de synchronisation locale

1. **Setup:**
   - Lancer 2 instances de Brique sur des ports différents
   - Instance A: port 8080
   - Instance B: port 8081

2. **Scénario:**
   - Ajouter des items dans l'instance A
   - Découvrir le pair B depuis A
   - Déclencher une synchronisation
   - Vérifier que les items apparaissent dans B
   - Modifier un item dans B
   - Re-synchroniser
   - Vérifier la résolution de conflit

3. **Validation:**
   - Logs de sync corrects dans les deux instances
   - Pas de duplication d'items
   - Résolution de conflits LWW fonctionne
   - Historique de sync précis

## Limitations actuelles

- ✅ Service de base créé, mais non connecté
- ❌ Pas de découverte automatique
- ❌ Pas de protocole de sync implémenté
- ❌ Pas d'UI
- ❌ Pas de tests
- ⚠️ Seulement les métadonnées des items (pas les assets)

## Améliorations futures (Phase 2+)

- Authentification par token (PSK)
- Chiffrement TLS
- Synchronisation automatique périodique
- Synchronisation des assets (pas seulement métadonnées)
- Résolution de conflits avancée (vector clocks)
- Mode Sneakernet (export/import via USB)
- Relay server pour synchronisation Internet
- Synchronisation sélective (filtres)
- Statistiques détaillées de synchronisation

## Fichiers créés

```
core/
├── models/
│   └── peer.go                    (60 lignes)
├── services/
│   └── gossip_service.go          (260 lignes)
└── db/
    ├── queries/
    │   ├── peers.sql              (11 requêtes)
    │   ├── sync_logs.sql          (5 requêtes)
    │   └── items.sql              (2 requêtes ajoutées)
    ├── peers.sql.go               (généré par sqlc)
    ├── sync_logs.sql.go           (généré par sqlc)
    └── models.go                  (types Peer, SyncLog ajoutés)

migrations/
├── 00003_create_peers_table.sql
└── 00004_create_sync_tables.sql

docs/
└── GOSSIP_GRIDS_DESIGN.md         (270 lignes)
```

## Statistiques

- **Lignes de code Go ajoutées:** ~320 lignes
- **Requêtes SQL créées:** 18
- **Migrations:** 2
- **Modèles de données:** 6 types
- **Documentation:** 1 document de design complet

## Conclusion

Le module Gossip Grids a été initié avec succès. La fondation est solide:
- Architecture bien définie
- Base de données prête
- Service de base implémenté et testé (compilation)

Les prochaines étapes (découverte, protocole, UI) nécessitent encore du développement mais le travail le plus architectural est fait. Le module suit les mêmes patterns que le reste de l'application et s'intègre proprement.

**État:** 🟡 En cours (fondations complètes, implémentation à continuer)
