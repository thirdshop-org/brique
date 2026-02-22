package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// --- 1. Structures de données (RFC-0001: Product Model) ---

// ProductData contient les infos "métier" du produit
// Dans la vraie vie, c'est ce qui est sérialisé en CBOR.
type ProductData struct {
	Name         string            `json:"name"`
	Manufacturer string            `json:"manufacturer"`
	Specs        map[string]string `json:"specs,omitempty"`
}

// CRDTMeta gère la concurrence et l'historique
type CRDTMeta struct {
	UpdatedAt    int64  `json:"updated_at"`    // Unix Timestamp
	AuthorDevice string `json:"author_device"` // Hex ID du device qui a écrit
}

// ProductObject est l'enveloppe signée qui circule sur le réseau
type ProductObject struct {
	SchemaVersion uint64              `json:"schema_v"`
	ProductID     string              `json:"product_id"` // Hash déterministe
	Category      string              `json:"category"`
	Data          ProductData         `json:"data"`
	Resources     map[string][]string `json:"resources"` // Images, Manuels (Hashs)
	CRDTMeta      CRDTMeta            `json:"crdt_meta"`
	Signature     []byte              `json:"signature"` // Signature Ed25519
}

// --- 2. Simulation Identité (Reprise de l'étape précédente) ---
type Device struct {
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
	ID         string // Hex string pour l'affichage
}

func NewMockDevice() *Device {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	hash := sha256.Sum256(pub)
	return &Device{
		PublicKey:  pub,
		PrivateKey: priv,
		ID:         hex.EncodeToString(hash[:]),
	}
}

// --- 3. Logique Produit ---

// GenerateDeterministicProductID crée un ID unique basé sur les invariants du produit.
// Si deux personnes créent la "Game Boy" de "Nintendo", elles doivent trouver le même ID.
func GenerateDeterministicProductID(category, manufacturer, name string) string {
	// On normalise les chaînes (minuscules, trim) pour éviter les doublons
	rawID := fmt.Sprintf("%s:%s:%s", category, manufacturer, name)

	// Hash SHA256
	h := sha256.Sum256([]byte(rawID))
	return hex.EncodeToString(h[:])
}

// PrepareCanonicalBytes transforme l'objet en bytes pour la signature.
// IMPORTANT : L'ordre des champs doit être STRICTEMENT le même partout.
// En production, on utilise CBOR Canonical (RFC 7049).
// Ici, on utilise JSON.Marshal qui trie les clés par ordre alphabétique (Go spec).
func (p *ProductObject) PrepareCanonicalBytes() []byte {
	// On crée une copie temporaire SANS la signature (car on ne peut pas signer la signature)
	temp := struct {
		Schema    uint64              `json:"schema_v"`
		ID        string              `json:"product_id"`
		Cat       string              `json:"category"`
		Data      ProductData         `json:"data"`
		Resources map[string][]string `json:"resources"`
		Meta      CRDTMeta            `json:"crdt_meta"`
	}{
		Schema:    p.SchemaVersion,
		ID:        p.ProductID,
		Cat:       p.Category,
		Data:      p.Data,
		Resources: p.Resources,
		Meta:      p.CRDTMeta,
	}

	bytes, _ := json.Marshal(temp)
	return bytes
}

// SignProduct signe le produit avec la clé privée du device
func (d *Device) SignProduct(p *ProductObject) {
	// 1. Mettre à jour les métadonnées auteur
	p.CRDTMeta.AuthorDevice = d.ID
	p.CRDTMeta.UpdatedAt = time.Now().Unix()

	// 2. Sérialiser
	payload := p.PrepareCanonicalBytes()

	// 3. Signer
	p.Signature = ed25519.Sign(d.PrivateKey, payload)
}

// VerifyProductSignature vérifie l'intégrité et l'auteur
func VerifyProductSignature(p *ProductObject, pubKey ed25519.PublicKey) bool {
	payload := p.PrepareCanonicalBytes()
	return ed25519.Verify(pubKey, payload, p.Signature)
}

// --- 4. Main execution ---

func NewProduct() {
	fmt.Println("📦 BRIQUE - Product Management POC")
	fmt.Println("----------------------------------")

	// 1. Initialisation du Device (Alice)
	aliceDevice := NewMockDevice()
	fmt.Printf("📱 Device actif: %s...\n\n", aliceDevice.ID[:8])

	// 2. Création d'un produit
	fmt.Println("🛠️  Création de la fiche produit...")
	prod := &ProductObject{
		SchemaVersion: 1,
		Category:      "electronics",
		Data: ProductData{
			Name:         "RetroGame Boy",
			Manufacturer: "Nintendo",
			Specs:        map[string]string{"cpu": "Z80", "screen": "LCD"},
		},
		Resources: map[string][]string{
			"images": {"hash_img_front_view", "hash_img_pcb"},
		},
	}

	// 3. Calcul de l'ID Déterministe
	// Cela garantit que si Bob crée la même fiche, il aura le même ProductID
	prod.ProductID = GenerateDeterministicProductID(
		prod.Category,
		prod.Data.Manufacturer,
		prod.Data.Name,
	)
	fmt.Printf("🔑 Product ID calculé: %s\n", prod.ProductID)

	// 4. Signature par le Device
	fmt.Println("✍️  Signature de la fiche par le Device...")
	aliceDevice.SignProduct(prod)
	fmt.Printf("📝 Signature générée: %x...\n\n", prod.Signature[:16])

	// --- Simulation Réseau ---
	// Imaginons que ce JSON soit envoyé à Bob
	jsonPayload, _ := json.MarshalIndent(prod, "", "  ")
	fmt.Println("📡 [RESEAU] Envoi du payload JSON :")
	fmt.Println(string(jsonPayload))

	// 5. Réception et Vérification par Bob
	fmt.Println("\n🔍 [BOB] Réception et vérification...")

	// Bob utilise la clé publique d'Alice (qu'il a reçue via le système d'identité précédent)
	isValid := VerifyProductSignature(prod, aliceDevice.PublicKey)

	if isValid {
		fmt.Println("✅ SUCCÈS: La fiche produit est authentique et non modifiée.")
		fmt.Printf("   Auteur: %s\n", prod.CRDTMeta.AuthorDevice)
		fmt.Printf("   Date:   %s\n", time.Unix(prod.CRDTMeta.UpdatedAt, 0))
	} else {
		fmt.Println("❌ ERREUR: Signature invalide ! Données corrompues ou faussaire.")
	}
}
