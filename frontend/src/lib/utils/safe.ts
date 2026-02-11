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

// Usage: const [err, items] = await safeCall(GetItems());
