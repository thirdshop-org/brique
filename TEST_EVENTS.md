# Test du Système d'Événements

## Préparation

```bash
# Build l'application
go build -o brique-gui .

# OU lancer en mode dev (si webkit2gtk-4.0 est installé)
wails dev

# En mode dev, accéder à: http://localhost:34115
```

## Tests à effectuer

### 1. Notification au démarrage ✅

**Résultat attendu :** Au lancement de l'application, une notification verte "Brique démarré" devrait apparaître en haut à droite.

### 2. Affichage des items existants ✅

**Fix appliqué :** Le frontend utilisait les bindings stub au lieu des vrais bindings Wails.

**Test :**
1. Ouvrir l'application
2. Les items existants dans la base de données devraient s'afficher

**Vérification avec CLI :**
```bash
./brique item list
```

Si vous voyez des items dans le CLI mais pas dans l'UI, vérifier que le frontend utilise bien :
```typescript
import { GetAllItems } from './lib/wails/wailsjs/go/main/App';
// et NON './lib/wails/go/main/App'
```

### 3. Création d'item avec notification

**Test :**
1. Cliquer sur "Ajouter" (bouton en haut à droite)
2. Remplir le formulaire (quand implémenté)
3. Valider

**Résultat attendu :**
- ✅ Notification verte : "Item créé - '[nom]' a été ajouté à l'inventaire"
- ✅ L'item apparaît dans la liste
- ✅ Auto-dismiss après 5 secondes

### 4. Mise à jour d'item

**Résultat attendu :**
- ✅ Notification verte : "Item mis à jour - '[nom]' a été modifié"

### 5. Suppression d'item

**Résultat attendu :**
- ✅ Notification verte : "Item supprimé - '[nom]' a été supprimé de l'inventaire"
- ✅ L'item disparaît de la liste

### 6. Erreur de chargement (test négatif)

**Test :**
1. Arrêter la base de données ou corrompre le fichier
2. Relancer l'application

**Résultat attendu :**
- ✅ Notification rouge : "Erreur de chargement - Impossible de charger les items"

### 7. Ajout d'asset avec progression

**Test :**
1. Ajouter un fichier à un item
2. Observer la barre de progression

**Résultat attendu :**
- ✅ Barre de progression en bas à droite : "Ajout de fichier"
- ✅ Affichage du nom du fichier
- ✅ Pourcentage et progression
- ✅ Disparaît automatiquement à la fin
- ✅ Notification verte : "Fichier ajouté - '[nom]' a été ajouté à l'item"

### 8. Types de notifications

**Test manuel dans la console développeur :**

```javascript
// Ouvrir la console (F12)
// Appeler directement les fonctions backend (si exposées)

// Success
window.go.main.App.CreateItem("Test", "Test", "Test", "Test", "", "");

// Error (avec ID invalide)
window.go.main.App.DeleteItem(99999);
```

### 9. Fermeture manuelle de notification

**Test :**
1. Déclencher une notification
2. Cliquer sur le bouton X

**Résultat attendu :**
- ✅ La notification disparaît immédiatement

### 10. Multiple notifications

**Test :**
1. Effectuer plusieurs actions rapidement
2. Observer le stack de notifications

**Résultat attendu :**
- ✅ Les notifications s'empilent verticalement
- ✅ Chaque notification a son propre timer
- ✅ Elles disparaissent dans l'ordre

## Dépendances manquantes

Si vous obtenez l'erreur `Package webkit2gtk-4.0 was not found` :

**Solution 1 : Installer webkit2gtk**
```bash
# Ubuntu/Debian
sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev

# Arch
sudo pacman -S webkit2gtk

# Fedora
sudo dnf install gtk3-devel webkit2gtk3-devel
```

**Solution 2 : Utiliser le mode navigateur**
```bash
wails dev
# Puis accéder à http://localhost:34115 dans votre navigateur
```

**Solution 3 : Build en mode production**
```bash
wails build
./build/bin/brique
```

## Logs pour debugging

Les événements sont loggés dans la console :

```bash
# Backend logs
# Visible dans le terminal où Wails tourne

# Frontend logs
# Visible dans la console développeur (F12 dans le navigateur)
```

## Résultat final attendu

✅ **Interface utilisateur :**
- Liste des items s'affiche correctement
- Notifications apparaissent pour toutes les actions CRUD
- Barres de progression pour les opérations longues
- Auto-dismiss après durée configurée
- Fermeture manuelle possible

✅ **Expérience utilisateur :**
- Feedback immédiat sur chaque action
- Messages clairs et contextuels
- Pas de latence visible
- Interface réactive et fluide

## Prochaines étapes suggérées

1. Implémenter le formulaire d'ajout/édition d'items
2. Ajouter la modale de détails
3. Implémenter le drag & drop pour les assets
4. Ajouter des sons pour les notifications critiques
5. Créer un dashboard avec statistiques
