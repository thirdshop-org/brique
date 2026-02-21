package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"
)

// --- 1. Structures de données (Conformes RFC-0001) ---

// Identity représente une identité chargée en mémoire (Human ou Device)
type Identity struct {
	Type       string             // "Human" ou "Device"
	PublicKey  ed25519.PublicKey  // 32 bytes
	PrivateKey ed25519.PrivateKey // 64 bytes (inclut la pub)
	ID         []byte             // SHA256(PublicKey)
}

// DelegationCertificate représente le lien de confiance
// RFC: Le Root (Human) signe le Device pour l'autoriser à agir
type DelegationCertificate struct {
	SchemaVersion uint64
	IssuerID      []byte   // Human ID
	SubjectID     []byte   // Device ID
	Permissions   []string // ex: "sign_product"
	ValidUntil    uint64   // Unix Timestamp
	Signature     []byte   // Ed25519 Signature
}

// --- 2. Fonctions de Génération (Cryptographie) ---

// GenerateIdentity crée une nouvelle paire de clés Ed25519 et dérive l'ID
func GenerateIdentity(idType string) (*Identity, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	// RFC Section 3.2: id = SHA256(pubkey)
	hash := sha256.Sum256(pub)
	id := hash[:]

	return &Identity{
		Type:       idType,
		PublicKey:  pub,
		PrivateKey: priv,
		ID:         id,
	}, nil
}

// CreateDelegation crée un certificat signé par le Root pour le Device
func CreateDelegation(root *Identity, device *Identity, duration time.Duration) (*DelegationCertificate, error) {
	expiry := uint64(time.Now().Add(duration).Unix())
	perms := []string{"sign_product", "sign_tutorial", "gossip_write"}

	cert := &DelegationCertificate{
		SchemaVersion: 1,
		IssuerID:      root.ID,
		SubjectID:     device.ID,
		Permissions:   perms,
		ValidUntil:    expiry,
	}

	// RFC Section 17 Annexe A: Signer le contenu canonique
	// NOTE: Dans une vraie implémentation, utiliser un encodeur CBOR Canonique (RFC 7049)
	// Ici, on simule une sérialisation déterministe simple pour le PoC.
	payload := serializeCertForSigning(cert)

	// Signature Ed25519
	signature := ed25519.Sign(root.PrivateKey, payload)
	cert.Signature = signature

	return cert, nil
}

// VerifyDelegation vérifie qu'un certificat est valide
func VerifyDelegation(rootPub ed25519.PublicKey, cert *DelegationCertificate) bool {
	// 1. Vérifier l'expiration
	if uint64(time.Now().Unix()) > cert.ValidUntil {
		fmt.Println("❌ Certificat expiré")
		return false
	}

	// 2. Reconstruire le payload (sans la signature)
	payload := serializeCertForSigning(cert)

	// 3. Vérifier la signature Ed25519
	return ed25519.Verify(rootPub, payload, cert.Signature)
}

// --- Helper: Sérialisation sommaire (Simule CBOR Canonical) ---
func serializeCertForSigning(cert *DelegationCertificate) []byte {
	// Ordre strict des champs pour le hash
	// SchemaVersion + IssuerID + SubjectID + ValidUntil + Permissions...
	var data []byte

	// Schema (uint64 Big Endian)
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, cert.SchemaVersion)
	data = append(data, b...)

	data = append(data, cert.IssuerID...)
	data = append(data, cert.SubjectID...)

	// Time (uint64 Big Endian)
	binary.BigEndian.PutUint64(b, cert.ValidUntil)
	data = append(data, b...)

	// Permissions (simples strings concaténées pour le PoC)
	for _, p := range cert.Permissions {
		data = append(data, []byte(p)...)
	}

	return data
}

// --- 3. Main (Scénario POC) ---

func NewIdentity() {
	fmt.Println("🧱 BRIQUE - Identity System POC (RFC-0001)")
	fmt.Println("------------------------------------------")

	// 1. Création de l'Identité Humaine (Root)
	// Cette clé doit être stockée offline (ex: Paper Key ou Hardware Wallet)
	human, _ := GenerateIdentity("Human")
	fmt.Printf("👤 HUMAN IDENTITY (Root)\n")
	fmt.Printf("   ID (Hash): %s\n", hex.EncodeToString(human.ID))
	fmt.Printf("   Pub Key:   %s...\n", hex.EncodeToString(human.PublicKey[:8]))
	fmt.Println()

	// 2. Création de l'Identité Device (Machine)
	// Cette clé vit sur le téléphone/serveur
	device, _ := GenerateIdentity("Device")
	fmt.Printf("📱 DEVICE IDENTITY (Node)\n")
	fmt.Printf("   ID (Hash): %s\n", hex.EncodeToString(device.ID))
	fmt.Printf("   Pub Key:   %s...\n", hex.EncodeToString(device.PublicKey[:8]))
	fmt.Println()

	// 3. Délégation
	// L'humain autorise ce device pour 30 jours
	fmt.Println("🔐 CREATION CERTIFICAT DELEGATION...")
	cert, _ := CreateDelegation(human, device, 30*24*time.Hour)

	fmt.Printf("   Issuer:    %s (Human)\n", hex.EncodeToString(cert.IssuerID))
	fmt.Printf("   Subject:   %s (Device)\n", hex.EncodeToString(cert.SubjectID))
	fmt.Printf("   Expires:   %d\n", cert.ValidUntil)
	fmt.Printf("   Signature: %s...\n", hex.EncodeToString(cert.Signature[:16]))
	fmt.Println()

	// 4. Vérification (Ce que font les autres nœuds du réseau)
	fmt.Println("🔍 VERIFICATION DU CERTIFICAT PAR UN TIERS...")
	isValid := VerifyDelegation(human.PublicKey, cert)

	if isValid {
		fmt.Println("✅ SUCCÈS : Le device est légitimement autorisé par l'humain.")
		fmt.Println("   Ce device peut maintenant signer des ProductObjects.")
	} else {
		fmt.Println("❌ ÉCHEC : Signature invalide.")
	}
}
