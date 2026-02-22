package main

import (
	"fmt"
)

func main() {
	fmt.Println("🚀 BRIQUE PROTOCOL - FULL SCENARIO")
	fmt.Println("==================================")

	// --- ACTE 1 : SETUP ---
	// Alice initialise son téléphone
	aliceDevice := NewMockDevice()
	// Bob initialise son téléphone
	bobDevice := NewMockDevice()

	// Le "Réseau" (Store) de Bob.
	// Au début, Bob ne connait qu'Alice (imaginons qu'il a scanné son QR Code)
	bobsStore := NewStore()
	bobsStore.RegisterIdentity(aliceDevice.ID, aliceDevice.PublicKey)
	// Bob se connait lui-même
	bobsStore.RegisterIdentity(bobDevice.ID, bobDevice.PublicKey)

	// Le Store d'Alice doit connaitre Bob pour accepter ses trads plus tard
	alicesStore := NewStore()
	alicesStore.RegisterIdentity(bobDevice.ID, bobDevice.PublicKey)

	// --- ACTE 2 : ALICE CRÉE ---
	fmt.Println("\n👩‍🔧 [ALICE] Crée une fiche produit...")

	// 1. Produit : Game Boy
	gbProd := &ProductObject{
		SchemaVersion: 1,
		Category:      "console",
		Data:          ProductData{Name: "Game Boy", Manufacturer: "Nintendo"},
		Resources:     map[string][]string{"images": {"img_gb_front"}},
	}
	gbProd.ProductID = GenerateDeterministicProductID("console", "Nintendo", "Game Boy")
	aliceDevice.SignProduct(gbProd)

	// 2. Produit : Tournevis (Outil)
	toolProd := &ProductObject{
		SchemaVersion: 1,
		Category:      "tool",
		Data:          ProductData{Name: "Screwdriver Tri-wing", Manufacturer: "iFixit"},
	}
	toolProd.ProductID = GenerateDeterministicProductID("tool", "iFixit", "Screwdriver Tri-wing")
	aliceDevice.SignProduct(toolProd)

	// 3. Tutoriel : Remplacer batterie
	fmt.Println("👩‍🔧 [ALICE] Rédige un tutoriel en Anglais...")
	tuto := CreateTutorial(aliceDevice, gbProd.ProductID, "en", "Replace Battery")

	// Ajout outil
	tuto.Tools = append(tuto.Tools, ToolReference{ProductID: toolProd.ProductID, Quantity: 1})

	// Ajout étapes
	tuto.AddStep("hash_img_step1", "en", "Unscrew the 6 screws on the back.")
	tuto.AddStep("hash_img_step2", "en", "Remove the back cover gently.")

	// Signature du tuto par Alice
	tuto.SignTutorial(aliceDevice)

	// --- ACTE 3 : SYNCHRONISATION (Alice -> Bob) ---
	fmt.Println("\n📡 [RESEAU] Alice envoie les données à Bob...")

	// Bob reçoit les produits
	err1 := bobsStore.IngestProduct(gbProd)
	err2 := bobsStore.IngestProduct(toolProd)
	if err1 != nil || err2 != nil {
		fmt.Println("❌ Erreur sync produits:", err1, err2)
	}

	// Bob reçoit le tuto
	err3 := bobsStore.IngestTutorial(tuto)
	if err3 == nil {
		fmt.Println("✅ Bob a bien reçu le tuto complet.")
	}

	// --- ACTE 4 : COLLABORATION (Bob traduit) ---
	fmt.Println("\n👨‍🔧 [BOB] Ajoute une traduction Française...")

	// Bob récupère le tuto de son store local
	tutoCopy := bobsStore.Tutorials[tuto.ID]

	// Il ajoute les trads
	tutoCopy.Title["fr"] = "Remplacer la batterie"
	tutoCopy.TranslateStep(1, "fr", "Dévissez les 6 vis à l'arrière.", bobDevice)
	tutoCopy.TranslateStep(2, "fr", "Retirez le couvercle doucement.", bobDevice)

	// IMPORTANT : Bob doit re-signer le tuto car il l'a modifié !
	// Dans un vrai CRDT, on signerait juste le "delta", mais ici on signe l'objet entier (LWW)
	tutoCopy.SignTutorial(bobDevice)

	fmt.Println("✍️  [BOB] A signé la nouvelle version (FR+EN).")

	// --- ACTE 5 : RETOUR (Bob -> Alice) ---
	fmt.Println("\n📡 [RESEAU] Bob renvoie la mise à jour à Alice...")

	// Alice reçoit la mise à jour
	// Elle vérifie la signature de Bob (elle connait sa clé publique)
	err4 := alicesStore.IngestTutorial(tutoCopy)

	if err4 == nil {
		finalTuto := alicesStore.Tutorials[tuto.ID]
		fmt.Println("✅ SUCCÈS FINAL : Alice voit maintenant le tuto en 2 langues !")
		fmt.Printf("   Titre (EN): %s\n", finalTuto.Title["en"])
		fmt.Printf("   Titre (FR): %s\n", finalTuto.Title["fr"])
		fmt.Printf("   Dernier Auteur: %s (C'est Bob !)\n", finalTuto.CRDTMeta.AuthorDevice[:8])
	} else {
		fmt.Println("❌ Alice a rejeté la mise à jour:", err4)
	}
}
