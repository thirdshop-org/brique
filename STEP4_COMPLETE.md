# Étape 4 : Fonctionnalités avancées - COMPLÉTÉE

Date : 13 février 2026

## Vue d'ensemble

Les fonctionnalités avancées ont été implémentées pour enrichir l'expérience utilisateur et sécuriser les données. Les utilisateurs peuvent maintenant générer des QR codes pour leurs objets, exporter/importer leur inventaire, et créer des backups de leurs données.

## Fonctionnalités implémentées

### 1. Génération de QR Codes

**Backend (Go):**
- Handler `GenerateQRCode(itemID int64) (string, error)`
- Utilise la bibliothèque `github.com/skip2/go-qrcode`
- Génère un QR code 256x256 pixels avec correction d'erreur moyenne
- Encode les données de l'item en JSON (ID, nom, marque, modèle, numéro de série)
- Retourne l'image en base64 pour affichage direct dans le frontend

**Frontend (Svelte):**
- Composant `QRCodeModal.svelte` (180 lignes)
- Affichage du QR code dans un modal élégant
- Bouton de téléchargement pour sauvegarder le QR code en PNG
- Message d'information sur l'utilisation du QR code
- Intégré dans `ItemDetailModal` avec un bouton "QR Code"

**Structure des données QR:**
```json
{
  "id": 123,
  "name": "Machine à laver",
  "brand": "Bosch",
  "model": "WAE28210FF",
  "serialNumber": "SN123456",
  "type": "brique-item"
}
```

**Cas d'usage:**
- Imprimer le QR code et le coller sur l'objet physique
- Scanner le QR code avec l'application mobile Brique (à venir)
- Accès rapide à la documentation de l'objet

### 2. Export de données

**Formats supportés:**
- **JSON** : Export complet avec métadonnées
- **CSV** : Export tabulaire pour Excel/LibreOffice

**Handler Go `ExportToJSON()`:**
- Récupère tous les items avec leurs assets
- Génère une structure `ExportData` complète :
  ```json
  {
    "exportDate": "2026-02-13T11:30:00Z",
    "version": "1.0",
    "items": [...],
    "stats": {
      "totalItems": 42
    }
  }
  ```
- Utilise `runtime.SaveFileDialog` pour choisir l'emplacement
- Nom de fichier par défaut: `brique-export-2026-02-13.json`
- Notification de succès avec nombre d'items exportés

**Handler Go `ExportToCSV()`:**
- Export tabulaire simple
- Colonnes: ID, Nom, Catégorie, Marque, Modèle, Numéro de série, Date d'achat, Notes, Date de création
- Encodage UTF-8
- Compatible avec Excel, LibreOffice, Google Sheets
- Nom de fichier par défaut: `brique-export-2026-02-13.csv`

**Frontend:**
- Boutons d'export dans le Dashboard
- Icônes distinctes (FileJson, FileSpreadsheet)
- États de chargement pendant l'export
- Notifications de succès/erreur

**Avantages:**
- Portabilité des données
- Analyse dans des outils externes
- Archivage long terme
- Migration vers d'autres systèmes

### 3. Import de données

**Handler Go `ImportFromJSON()`:**
- Lecture d'un fichier JSON exporté depuis Brique
- Utilise `runtime.OpenFileDialog` pour sélectionner le fichier
- Parse la structure `ExportData`
- Validation des données
- Détection des doublons (par numéro de série)
- Skip des items existants
- Compteurs: items importés vs ignorés
- Notifications détaillées (succès/avertissement)

**Stratégie de déduplication:**
- Recherche par numéro de série avant import
- Si le numéro de série existe déjà, l'item est ignoré
- Compteur des items ignorés dans la notification

**Limitations:**
- Les assets ne sont PAS importés (chemins de fichiers locaux)
- Seules les métadonnées des items sont importées
- Les utilisateurs doivent ré-ajouter les assets manuellement via la CLI ou l'UI

**Frontend:**
- Bouton "Importer" dans le Dashboard
- Dialog de sélection de fichier natif
- Rechargement automatique des données après import
- États de chargement

**Cas d'usage:**
- Restaurer des données après une réinstallation
- Migrer entre machines
- Fusionner plusieurs inventaires
- Tester avec des données d'exemple

### 4. Backup automatique

**Handler Go `CreateBackup()`:**
- Création d'un dossier de backup horodaté
- Structure: `backups/backup_2026-02-13_11-30-00/`
- Contenu du backup:
  - `brique.db` : copie de la base de données SQLite
  - `assets/` : copie complète du dossier assets
  - `metadata.json` : métadonnées (timestamp, version)

**Processus de backup:**
1. Création du dossier `~/.config/brique/backups/`
2. Création d'un sous-dossier avec timestamp
3. Copie de la base de données SQLite
4. Copie récursive du dossier assets
5. Génération d'un fichier metadata.json
6. Notification de succès avec chemin du backup

**Intégration avec le système d'événements:**
- Barre de progression pendant le backup
- Étapes: "Copie de la base de données" (30%), "Copie des assets" (60%)
- Notification de succès avec chemin complet
- Gestion des erreurs avec notifications

**Frontend:**
- Bouton "Backup" vert dans le Dashboard
- Icône HardDrive
- État de chargement pendant le backup
- Barre de progression en temps réel

**Fonctions helper ajoutées:**
- `copyFile(src, dst string) error` : copie un fichier
- `copyDir(src, dst string) error` : copie récursive de dossier

**Cas d'usage:**
- Sauvegardes avant modifications importantes
- Protection contre la perte de données
- Archivage périodique
- Restauration en cas de problème

**Améliorations futures:**
- Backup automatique périodique (quotidien, hebdomadaire)
- Rétention automatique (garder les N derniers backups)
- Compression des backups (tar.gz)
- Restauration depuis l'UI
- Backup vers cloud (optionnel)

## Modifications de fichiers

### Backend (Go)

**app_handlers.go** (extensions majeures):
- Ajout des imports: `encoding/base64`, `encoding/csv`, `encoding/json`, `io`, `os`, `path/filepath`, `time`, `runtime`
- Ajout du type `QRCodeData`
- Ajout du type `ExportData`
- Handler `GenerateQRCode(itemID int64) (string, error)` - 40 lignes
- Handler `ExportToJSON() error` - 90 lignes
- Handler `ImportFromJSON() error` - 70 lignes
- Handler `ExportToCSV() error` - 80 lignes
- Handler `CreateBackup() error` - 100 lignes
- Helper `copyFile(src, dst string) error` - 15 lignes
- Helper `copyDir(src, dst string) error` - 40 lignes
- **Total ajouté:** ~435 lignes

**go.mod**:
- Ajout de `github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e`

### Frontend (Svelte)

**Nouveau composant:**
- `QRCodeModal.svelte` (180 lignes)
  - Modal avec affichage du QR code
  - Téléchargement en PNG
  - Animations et design cohérent

**Modifications:**
- `App.svelte`:
  - Import de `QRCodeModal`
  - États pour gérer le QR modal
  - Fonctions `openQRCodeModal` et `closeQRCodeModal`
  - Intégration du QRCodeModal dans les modals
  - Passage du callback `onGenerateQR` à ItemDetailModal

- `ItemDetailModal.svelte`:
  - Import de l'icône `QrCode`
  - Ajout du prop `onGenerateQR`
  - Bouton "QR Code" dans le footer

- `Dashboard.svelte`:
  - Import des fonctions: `ExportToJSON`, `ExportToCSV`, `ImportFromJSON`, `CreateBackup`
  - Import des icônes: `FileJson`, `FileSpreadsheet`, `Upload`, `HardDrive`
  - Import de `eventBus` pour les notifications
  - États: `exporting`, `importing`, `backing`
  - Fonctions: `handleExportJSON`, `handleExportCSV`, `handleImportJSON`, `handleCreateBackup`
  - Section "Export/Import/Backup Actions" dans le header
  - 4 boutons: Backup (vert), Importer, JSON, CSV

**Bindings Wails générés:**
- `GenerateQRCode(arg1:number):Promise<string>`
- `ExportToJSON():Promise<void>`
- `ExportToCSV():Promise<void>`
- `ImportFromJSON():Promise<void>`
- `CreateBackup():Promise<void>`

## Statistiques

- **Lignes de code Go ajoutées:** ~435 lignes
- **Nouveaux composants Svelte:** 1 (QRCodeModal)
- **Composants modifiés:** 3 (App, ItemDetailModal, Dashboard)
- **Nouvelles dépendances Go:** 1 (go-qrcode)
- **Handlers exposés:** 5
- **Taille du bundle frontend:**
  - JS: 114 KB (34.3 KB gzippé)
  - CSS: 21.5 KB (4.8 KB gzippé)

## Tests manuels suggérés

### Test QR Code
1. Ouvrir le détail d'un item
2. Cliquer sur "QR Code"
3. Vérifier l'affichage du QR code
4. Télécharger le QR code
5. Scanner avec un lecteur QR (vérifier les données JSON)

### Test Export JSON
1. Aller dans le Dashboard
2. Cliquer sur "JSON"
3. Choisir un emplacement de sauvegarde
4. Vérifier le fichier généré (JSON valide, avec tous les items)
5. Vérifier la notification de succès

### Test Export CSV
1. Aller dans le Dashboard
2. Cliquer sur "CSV"
3. Choisir un emplacement de sauvegarde
4. Ouvrir le fichier dans Excel/LibreOffice
5. Vérifier que toutes les colonnes sont présentes
6. Vérifier l'encodage UTF-8 (caractères spéciaux)

### Test Import JSON
1. Créer un export JSON
2. Supprimer quelques items (optionnel)
3. Cliquer sur "Importer"
4. Sélectionner le fichier JSON exporté
5. Vérifier la notification (X items importés, Y ignorés)
6. Vérifier que les items sont apparus dans la liste

### Test Backup
1. Aller dans le Dashboard
2. Cliquer sur "Backup"
3. Attendre la fin (barre de progression)
4. Vérifier la notification avec le chemin du backup
5. Explorer `~/.config/brique/backups/backup_YYYY-MM-DD_HH-MM-SS/`
6. Vérifier la présence de: `brique.db`, `assets/`, `metadata.json`
7. Vérifier la taille des fichiers (non vides)

## Améliorations futures possibles

### QR Codes
- [ ] Personnalisation de la taille du QR code
- [ ] Ajout d'un logo Brique au centre du QR code
- [ ] Impression directe depuis l'application
- [ ] Génération batch de QR codes (tous les items)
- [ ] QR codes avec URL vers une webapp (si hébergé)

### Export/Import
- [ ] Export sélectif (filtrer par catégorie, marque)
- [ ] Import CSV
- [ ] Mapping de colonnes lors de l'import CSV
- [ ] Preview des données avant import
- [ ] Gestion des conflits (choisir: remplacer, ignorer, fusionner)
- [ ] Export des assets (dans un ZIP)
- [ ] Import des assets depuis un ZIP

### Backup
- [ ] Planification automatique (quotidien, hebdomadaire)
- [ ] Rétention automatique (garder les N derniers)
- [ ] Compression (tar.gz, zip)
- [ ] Interface de restauration depuis backup
- [ ] Liste des backups disponibles dans l'UI
- [ ] Backup incrémental (seulement les changements)
- [ ] Backup vers le cloud (Nextcloud, Syncthing, etc.)
- [ ] Vérification de l'intégrité des backups

### Général
- [ ] Planificateur de tâches (backups automatiques)
- [ ] Synchronisation avec d'autres instances Brique (P2P)
- [ ] API REST pour intégrations externes
- [ ] Webhooks (notification de changements)

## Conclusion

L'étape 4 est maintenant complète avec:
- ✅ Génération de QR codes pour identification rapide des objets
- ✅ Export de données en JSON et CSV pour portabilité
- ✅ Import de données depuis JSON pour restauration
- ✅ Système de backup manuel pour sécurité des données

Ces fonctionnalités positionnent Brique comme une solution robuste et professionnelle pour la gestion d'inventaire local-first, avec une attention particulière à la sécurité et la portabilité des données.

**Prochaine étape:** Module "Gossip Grids" pour la synchronisation P2P (Étape 5).
