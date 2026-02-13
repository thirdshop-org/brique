# Gossip Grids - Document de conception

Date : 13 février 2026

## Vision

Le module "Gossip Grids" permet la synchronisation décentralisée des inventaires entre plusieurs instances de Brique. Il suit la philosophie "local-first" en permettant le partage de connaissances sans dépendre d'un serveur central.

## Cas d'usage

1. **Atelier de réparation communautaire** : Plusieurs réparateurs partagent leurs documentations
2. **Maison multi-utilisateurs** : Synchronisation entre les machines de la famille
3. **Mode Sneakernet** : Synchronisation via clé USB en l'absence de réseau
4. **Réseau de repair cafés** : Partage de documentation entre différents lieux

## Architecture générale

```
┌─────────────────────────────────────────────────────────────┐
│                     Brique Instance A                        │
│                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ Backpack     │  │ Gossip       │  │ HTTP Server  │      │
│  │ Service      │◄─┤ Service      │◄─┤ (sync API)   │      │
│  └──────────────┘  └──────────────┘  └──────┬───────┘      │
│                           │                   │              │
│                           │                   │              │
│                    ┌──────▼───────┐           │              │
│                    │ Peer Manager │           │              │
│                    │ (discovery)  │           │              │
│                    └──────────────┘           │              │
└────────────────────────────────────────────────┼─────────────┘
                                                 │
                                          Network│(LAN/Internet)
                                                 │
┌────────────────────────────────────────────────┼─────────────┐
│                     Brique Instance B          │              │
│                                                │              │
│  ┌──────────────┐  ┌──────────────┐  ┌───────▼──────┐      │
│  │ Backpack     │  │ Gossip       │  │ HTTP Server  │      │
│  │ Service      │◄─┤ Service      │◄─┤ (sync API)   │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Composants

### 1. GossipService (core/services/gossip_service.go)

Responsable de la logique de synchronisation.

**Méthodes:**
- `DiscoverPeers() ([]Peer, error)` : Découvre les pairs sur le réseau local
- `AddPeer(peer *Peer) error` : Ajoute un pair manuellement
- `SyncWithPeer(peerID string) (*SyncResult, error)` : Synchronise avec un pair
- `GetPeers() ([]Peer, error)` : Liste tous les pairs connus
- `RemovePeer(peerID string) error` : Supprime un pair
- `GetSyncHistory() ([]SyncLog, error)` : Historique des synchronisations

**Structure de données:**
```go
type Peer struct {
    ID          string    // UUID unique
    Name        string    // Nom de l'instance
    Address     string    // IP:Port
    LastSeen    time.Time
    LastSync    time.Time
    Status      PeerStatus // online, offline, syncing
    IsTrusted   bool      // Pair approuvé par l'utilisateur
}

type SyncResult struct {
    ItemsReceived int
    ItemsSent     int
    Conflicts     int
    Duration      time.Duration
}

type SyncLog struct {
    ID        int64
    PeerID    string
    Timestamp time.Time
    Result    SyncResult
    Error     string
}
```

### 2. Découverte de pairs

#### Option A: mDNS/Bonjour (recommandé)
- Utilise le package `github.com/hashicorp/mdns`
- Broadcast automatique sur le LAN
- Pas de configuration nécessaire
- Fonctionne sur tous les OS

#### Option B: Broadcast UDP
- Simple mais moins fiable
- Nécessite configuration du pare-feu
- Pas de support multicast sur tous les réseaux

**Implémentation retenue: mDNS**

### 3. Protocole de synchronisation

#### API HTTP REST

**Endpoints:**

```
GET  /api/v1/gossip/info
POST /api/v1/gossip/sync
GET  /api/v1/gossip/items/changes?since=<timestamp>
POST /api/v1/gossip/items/batch
```

**Flux de synchronisation:**

```
Instance A                          Instance B
    |                                   |
    |─── GET /gossip/info ──────────>  |
    |<── {instanceID, lastSync} ────   |
    |                                   |
    |─── GET /items/changes?since ──>  |
    |<── [changedItems] ────────────   |
    |                                   |
    |─── POST /items/batch ──────────> |
    |<── {received: 5} ──────────────  |
    |                                   |
```

#### Résolution de conflits

**Stratégie: Last-Write-Wins (LWW)**

- Chaque modification stocke un timestamp (UpdatedAt)
- En cas de conflit, la version la plus récente gagne
- Simple et efficace pour ce cas d'usage

**Alternative future: Vector Clocks**

Pour une résolution plus sophistiquée si nécessaire.

### 4. Schéma de base de données (migrations)

```sql
-- Table des pairs connus
CREATE TABLE peers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    last_seen TIMESTAMP,
    last_sync TIMESTAMP,
    is_trusted BOOLEAN DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table d'historique de sync
CREATE TABLE sync_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    peer_id TEXT NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    items_received INTEGER,
    items_sent INTEGER,
    conflicts INTEGER,
    duration_ms INTEGER,
    error TEXT,
    FOREIGN KEY (peer_id) REFERENCES peers(id)
);

-- Ajouter un champ à items pour tracker l'origine
ALTER TABLE items ADD COLUMN origin_peer_id TEXT;
ALTER TABLE items ADD COLUMN sync_version INTEGER DEFAULT 1;
```

### 5. Sécurité et confiance

#### Modèle de confiance

- **Par défaut: Découverte uniquement** (pas de sync automatique)
- **Sync manuelle** : L'utilisateur décide avec qui synchroniser
- **Pairs approuvés** : Liste de pairs de confiance pour sync automatique

#### Authentification (Phase 2)

- Token partagé (PSK - Pre-Shared Key)
- Échange de tokens via QR code
- TLS pour le chiffrement du transport

#### Validation des données

- Vérification des hash SHA256 des assets
- Validation des champs requis
- Rejet des données malformées

### 6. Interface utilisateur

#### Nouvel onglet "Synchronisation"

**Sections:**

1. **Pairs découverts**
   - Liste des pairs actifs sur le LAN
   - Bouton "Synchroniser" pour chaque pair
   - Indicateur de statut (online/offline/syncing)

2. **Pairs approuvés**
   - Liste des pairs de confiance
   - Toggle pour activer la sync automatique
   - Bouton pour supprimer un pair

3. **Historique de synchronisation**
   - Timeline des dernières sync
   - Détails: items échangés, durée, erreurs
   - Filtre par pair

4. **Configuration**
   - Nom de l'instance
   - Port d'écoute
   - Activer/désactiver la découverte
   - Activer/désactiver la sync automatique

## Phases d'implémentation

### Phase 1: Base (MVP)
- [x] Service GossipService
- [x] Découverte mDNS sur LAN
- [x] API HTTP de synchronisation
- [x] Sync manuelle (bouton UI)
- [x] Last-Write-Wins pour conflits
- [x] UI basique avec liste de pairs

### Phase 2: Améliorations
- [ ] Authentification par token
- [ ] Chiffrement TLS
- [ ] Sync automatique périodique
- [ ] Synchronisation des assets (pas seulement métadonnées)
- [ ] Résolution de conflits avancée
- [ ] Statistiques détaillées

### Phase 3: Modes avancés
- [ ] Mode Sneakernet (export/import via USB)
- [ ] Synchronisation Internet (relay server optionnel)
- [ ] Sync sélective (filtres par catégorie)
- [ ] Versioning et rollback

## Limitations connues

1. **Assets non synchronisés** : Seules les métadonnées sont partagées (pour l'instant)
2. **Pas de chiffrement** : Communication en clair (phase 1)
3. **Pas d'authentification** : Confiance implicite sur le LAN
4. **Pas de sync automatique** : L'utilisateur doit déclencher manuellement
5. **Conflits simplistes** : Last-Write-Wins peut perdre des modifications

## Alternatives considérées

### CRDT (Conflict-free Replicated Data Types)
- **Avantages** : Résolution automatique des conflits
- **Inconvénients** : Complexité élevée, overhead de stockage
- **Décision** : Non retenu pour la phase 1

### Blockchain/Merkle DAG
- **Avantages** : Traçabilité complète, immuabilité
- **Inconvénients** : Overkill pour ce cas d'usage
- **Décision** : Non nécessaire

### Syncthing intégré
- **Avantages** : Solution mature et éprouvée
- **Inconvénients** : Dépendance externe, moins de contrôle
- **Décision** : Peut être une intégration future

## Références

- [Gossip Protocols](https://en.wikipedia.org/wiki/Gossip_protocol)
- [CRDT](https://crdt.tech/)
- [mDNS RFC 6762](https://tools.ietf.org/html/rfc6762)
- [Local-First Software](https://www.inkandswitch.com/local-first/)

## Prochaines étapes

1. ✅ Créer ce document de design
2. Implémenter GossipService
3. Ajouter la découverte mDNS
4. Créer l'API HTTP de sync
5. Développer l'UI de gestion des pairs
6. Tests d'intégration avec 2+ instances
