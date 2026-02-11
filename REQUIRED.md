1. Structure du Projet (Raffinée)

On ajoute des dossiers pour la gestion des migrations et la config.

/brique
├── /cmd
│   ├── /brique-ui       # Main entry point Wails (GUI)
│   └── /brique-cli      # Main entry point CLI (Headless)
├── /core                # DOMAINE (Agnostique : ne connait ni Wails ni CLI)
│   ├── /db              # Repository pattern (SQLite)
│   ├── /services        # Logique métier (Sac à dos manager)
│   └── /models          # Structs Go partagées
├── /frontend            # Svelte + Shadcn
│   ├── /src/lib/wails   # Tes wrappers TS "Tuple Return"
├── /migrations          # Les fichiers .sql pour l'évolution du schéma
├── /pkg                 # Utils techniques (Logger, FS helper)
└── wails.json

2. Backend Requirements (Go & SQLite)
A. Bibliothèque SQL & Migrations

Pour garder l'esprit "Brique" (robuste et performant), je te conseille d'éviter un ORM lourd (comme GORM) qui cache trop de magie.

    Requis : Utiliser sqlc.
        Pourquoi ? sqlc génère du code Go type-safe à partir de tes requêtes SQL brutes. C’est ultra rapide et tu contrôles ton SQL.
    Requis : Outil de migration : goose.
        Le binaire doit pouvoir jouer les migrations au démarrage (auto-migrate au launch) pour que l'utilisateur n'ait rien à configurer.

B. Configuration Manager (Viper)

L'application doit savoir où stocker sa brique (ses fichiers) selon l'OS (Windows %APPDATA%, Linux ~/.config ou /var/lib en headless).

    Requis : Le core doit accepter un Config struct injecté au démarrage pour définir les chemins de stockage (Database path, Downloads folder).

C. Graceful Shutdown & Context

    Requis : Tout appel long (téléchargement, grosse requête DB) doit accepter un context.Context.
    Si l'utilisateur ferme l'app ou fait Ctrl+C en mode headless, ça ne doit pas corrompre la base de données.

3. Frontend Requirements (Svelte & TS)
A. Le Pattern "Safe Fetch" (Tuple Return)

Puisque Wails génère des promesses qui peuvent "fail", tu dois créer un helper générique pour respecter ton exigence [error, data].

Requirement Code (à mettre dans /frontend/src/lib/utils/safe.ts) :

// Type générique pour le retour tuple
type SafeResult<T> = Promise<[Error, null] | [null, T]>;

// Wrapper pour les appels Wails
export async function safeCall<T>(promise: Promise<T>): SafeResult<T> {
    try {
        const data = await promise;
        return [null, data];
    } catch (err) {
        // Normalisation de l'erreur Wails (souvent une string wrapper)
        const error = err instanceof Error ? err : new Error(String(err));
        return [error, null];
    }
}

Usage : const [err, items] = await safeCall(GetItems());
B. UI/UX Shadcn "Custom"

    Requis : Utiliser Lucide-Svelte pour les icônes (léger, vectoriel).
    Requis : Thème Shadcn configuré en "Slate" ou "Gray" (pas de bleu par défaut), avec radius: 0 ou 0.25rem pour un aspect carré/brique.
    Requis : Police "Inter" ou "JetBrains Mono" pour les données techniques (numéros de série, hash).

4. Communication Core <-> UI (Events)

Wails permet d'appeler des fonctions Go depuis JS, mais le Core doit aussi pouvoir "pousser" de l'info (ex: progression téléchargement).

    Requis : Un système de "Bus d'événements" dans le /app.
        Le /core émet un événement Go standard (ex: channel ou callback).
        Le /app écoute et transmute ça en Wails.EventsEmit.
        Pourquoi ? Pour que le core reste pur et testable sans wails.

5. Logging & Debugging (Crucial pour le Headless)

Si l'app tourne sur un Raspberry Pi sans écran, comment savoir pourquoi elle plante ?

    Requis : Logging structuré (Slog).
    Sortie :
        En UI : Logs écrits dans un fichier debug.log.
        En Headless : Logs colorés dans la console (StdOut).

6. Résumé des packages Go

Voici ta "Shopping List" pour le go.mod :

    DB Driver : modernc.org/sqlite (Version pure Go, évite les problèmes de compilation CGO chiants sous Windows/Cross-compile).
    Migrations : github.com/pressly/goose/v3.
    Logs : log/slog (Standard Go 1.21+).
    CLI : github.com/spf13/cobra (Le standard pour gérer les flags --headless proprement).
