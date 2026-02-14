# Brique Server - Docker Deployment

Ce guide décrit comment déployer Brique en mode serveur API avec Docker.

## 🚀 Démarrage rapide

### Avec Docker Compose

```bash
docker-compose up -d
```

### Avec Docker uniquement

```bash
# Build
docker build -t brique-server .

# Run
docker run -d \
  --name brique \
  -p 8080:8080 \
  -v brique-data:/var/lib/brique \
  -e BRIQUE_INSTANCE_NAME="Mon Brique" \
  brique-server
```

## 📡 API Endpoints

Une fois déployé, l'API est accessible sur `http://localhost:8080`

### Health Check
- `GET /health` - Vérifie l'état du serveur

### Items (Inventaire)
- `GET /api/v1/items` - Liste tous les items
- `POST /api/v1/items` - Crée un nouvel item
- `GET /api/v1/items/{id}` - Récupère un item
- `PUT /api/v1/items/{id}` - Met à jour un item
- `DELETE /api/v1/items/{id}` - Supprime un item

### Assets (Documentation)
- `GET /api/v1/items/{id}/assets` - Liste les assets d'un item
- `DELETE /api/v1/assets/{id}` - Supprime un asset

### Gossip (Synchronisation P2P)
- `GET /api/v1/gossip/info` - Informations sur l'instance
- `GET /api/v1/gossip/changes?since={timestamp}` - Changements depuis une date
- `GET /api/v1/gossip/peers` - Liste des pairs découverts
- `POST /api/v1/gossip/peers` - Ajoute un pair manuellement
- `PUT /api/v1/gossip/peers/{id}` - Met à jour la confiance d'un pair
- `DELETE /api/v1/gossip/peers/{id}` - Supprime un pair
- `POST /api/v1/gossip/sync/{peer_id}` - Synchronise avec un pair

## 🔧 Variables d'environnement

| Variable | Description | Défaut |
|----------|-------------|--------|
| `BRIQUE_DATA_DIR` | Répertoire des données | `/var/lib/brique` |
| `BRIQUE_PORT` | Port HTTP | `8080` |
| `BRIQUE_INSTANCE_NAME` | Nom de l'instance | `Brique-Server` |

## 💾 Volumes

- `/var/lib/brique` - Contient la base de données SQLite et les fichiers assets

## 📊 Déploiement sur Dokploy

1. Connectez votre repository Git à Dokploy
2. Dokploy détectera automatiquement le Dockerfile
3. Configurez les variables d'environnement si nécessaire
4. Configurez un volume persistant pour `/var/lib/brique`
5. Déployez !

## 🔒 Sécurité

⚠️ **Important** : Cette version n'inclut pas d'authentification. Pour un déploiement en production :

- Ajoutez un reverse proxy (Traefik, Nginx) avec HTTPS
- Configurez l'authentification (OAuth2, JWT, etc.)
- Limitez l'accès réseau via firewall
- Utilisez des secrets pour les configurations sensibles

## 📝 Exemples d'utilisation de l'API

### Créer un item

```bash
curl -X POST http://localhost:8080/api/v1/items \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Machine à laver",
    "category": "Électroménager",
    "brand": "Samsung",
    "model": "WW90T",
    "notes": "Achetée en 2023"
  }'
```

### Lister tous les items

```bash
curl http://localhost:8080/api/v1/items
```

### Obtenir les informations de l'instance

```bash
curl http://localhost:8080/api/v1/gossip/info
```

### Ajouter un pair manuellement (pour sync inter-VPS)

```bash
curl -X POST http://localhost:8080/api/v1/gossip/peers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Brique Production",
    "address": "vps.example.com:8080",
    "is_trusted": true
  }'
```

### Synchroniser avec un pair

```bash
# Récupérer l'ID du pair
curl http://localhost:8080/api/v1/gossip/peers

# Lancer la synchronisation
curl -X POST http://localhost:8080/api/v1/gossip/sync/{peer_id}
```

## 🐛 Debug

Voir les logs :
```bash
docker-compose logs -f brique
```

Accéder au conteneur :
```bash
docker-compose exec brique sh
```

Vérifier la base de données :
```bash
docker-compose exec brique sqlite3 /var/lib/brique/brique.db ".tables"
```
