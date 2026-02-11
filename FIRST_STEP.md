Module : Le Sac à Dos (The Backpack)

Description Générale
Le "Sac à Dos" est le module central de gestion d'inventaire personnel. Il agit comme un "Double Numérique" (Digital Twin) des possessions physiques de l'utilisateur. Son but est de centraliser, organiser et sécuriser localement toutes les données nécessaires à la maintenance et à la réparation des objets possédés.

Objectif Principal
Garantir l'autonomie totale de l'utilisateur : si Internet disparaît demain, l'utilisateur doit pouvoir réparer n'importe quel objet présent dans son "Sac à Dos" uniquement avec les données déjà stockées sur sa machine.
Fonctionnalités Requises (Features Scope)
1. Gestion de l'Inventaire (Input)

    Création de Fiche Objet : L'utilisateur doit pouvoir ajouter un objet physique dans la base.
    Champs de Données :
        Identité : Nom (ex: "Lave-Linge Cuisine"), Catégorie (ex: Gros Électroménager), Marque, Modèle précis.
        Traçabilité : Numéro de Série, Date d'achat (optionnel).
        Média : Photo de l'objet (upload local).
        Notes : Zone de texte libre pour l'historique de panne ou astuces perso.

2. Gestion des "Assets" de Réparation (Data Acquisition)

Le système associe des fichiers locaux à chaque objet de l'inventaire.

    Types d'Assets supportés :
        Manuels utilisateurs (PDF).
        Manuels de service / Schémas techniques (PDF, Images).
        Vues éclatées (Exploded Views).
        Fichiers de fabrication (STL pour impression 3D de pièces détachées).
        Firmwares / Drivers (Binaires).
    Statut de complétude : Chaque objet doit afficher un indicateur visuel de "Santé Documentaire" (ex: "Incomplet" si aucun manuel n'est stocké, "Sécurisé" si les schémas techniques sont présents localement).

3. Organisation et Recherche (UX)

    Visualisation : Affichage sous forme de grille (cartes visuelles) ou de liste technique.
    Recherche Rapide : Filtrage instantané par nom, marque ou catégorie.
    Navigation Hiérarchique : Possibilité de voir la liste des fichiers associés directement depuis la vue globale de l'objet sans navigation complexe.

4. Le Lien Physique (Physical Link)

    Génération d'Étiquettes : Le module doit pouvoir générer une étiquette imprimable (QR Code ou ID lisible) pour chaque objet.
    Scan-to-Open : L'objectif est qu'en scannant ce code (futur usage mobile), on ouvre directement la fiche du "Sac à Dos" correspondante.

5. Philosophie "Offline-First"

    Souveraineté des Données : Une fois un fichier (Asset) ajouté au Sac à Dos, il ne doit plus jamais dépendre d'une URL externe. Il est copié et stocké dans le système de fichiers local.
    Portabilité : Le Sac à Dos doit être conçu comme un dossier portable. Si on déplace le dossier de l'application, les liens entre les objets et leurs fichiers PDF/STL ne doivent pas être rompus.
