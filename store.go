package main

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
)

// LocalStore simule le disque dur du téléphone
type LocalStore struct {
	Identities map[string]ed25519.PublicKey // Cache des clés publiques connues (Trust Store)
	Products   map[string]*ProductObject
	Tutorials  map[string]*TutorialObject
	// Map : TutoID -> Liste des Hashs des versions disponibles
	TutorialIndex map[string][]string

	// Map : HashVersion -> L'objet complet
	TutorialBlobs map[string]*TutorialObject
}

func NewStore() *LocalStore {
	return &LocalStore{
		Identities: make(map[string]ed25519.PublicKey),
		Products:   make(map[string]*ProductObject),
		Tutorials:  make(map[string]*TutorialObject),
	}
}

// RegisterIdentity simule la découverte d'un pair (QR Code scan, etc)
func (s *LocalStore) RegisterIdentity(idStr string, pubKey ed25519.PublicKey) {
	s.Identities[idStr] = pubKey
}

// IngestProduct reçoit un produit du réseau, vérifie la signature et stocke
func (s *LocalStore) IngestProduct(p *ProductObject) error {
	// 1. Trouver la clé publique de l'auteur
	authorPub, known := s.Identities[p.CRDTMeta.AuthorDevice]
	if !known {
		return fmt.Errorf("auteur inconnu: %s", p.CRDTMeta.AuthorDevice)
	}

	// 2. Vérifier la signature
	// Note: On réutilise la logique de product.go (PrepareCanonicalBytes non réécrit ici par souci de brièveté, voir code précédent)
	// Dans ce PoC simplifié, on suppose que VerifyProductSignature existe et fonctionne
	if !VerifyProductSignature(p, authorPub) {
		return fmt.Errorf("signature produit invalide")
	}

	// 3. Stocker (Last Write Wins)
	existing, exists := s.Products[p.ProductID]
	if !exists || p.CRDTMeta.UpdatedAt > existing.CRDTMeta.UpdatedAt {
		s.Products[p.ProductID] = p
		fmt.Printf("💾 [STORE] Produit sauvegardé/mis à jour : %s\n", p.Data.Name)
	}
	return nil
}

// IngestTutorial fait pareil pour les tutos
func (s *LocalStore) IngestTutorial(t *TutorialObject) error {
	authorPub, known := s.Identities[t.CRDTMeta.AuthorDevice]
	if !known {
		return fmt.Errorf("auteur tuto inconnu")
	}

	// Vérification signature manuelle pour le PoC
	temp := *t
	temp.Signature = nil
	payload, _ := json.Marshal(temp)

	if !ed25519.Verify(authorPub, payload, t.Signature) {
		return fmt.Errorf("signature tuto invalide")
	}

	s.Tutorials[t.ID] = t
	fmt.Printf("💾 [STORE] Tutoriel sauvegardé : %v\n", t.Title)
	return nil
}

// Fonction pour récupérer la "Meilleure" version à afficher
func (s *LocalStore) GetBestTutorialVersion(tutoID string, trustedKeys map[string]bool) *TutorialObject {
	versionsHashes := s.TutorialIndex[tutoID]
	if len(versionsHashes) == 0 {
		return nil
	}

	var bestVersion *TutorialObject

	for _, h := range versionsHashes {
		candidate := s.TutorialBlobs[h]

		if bestVersion == nil {
			bestVersion = candidate
			continue
		}

		// CRITÈRE 1 : La Confiance (Est-ce que je connais l'auteur ?)
		isCandidateTrusted := trustedKeys[candidate.CRDTMeta.AuthorDevice]
		isBestTrusted := trustedKeys[bestVersion.CRDTMeta.AuthorDevice]

		if isCandidateTrusted && !isBestTrusted {
			bestVersion = candidate
			continue
		}
		if !isCandidateTrusted && isBestTrusted {
			continue // On garde la version de confiance
		}

		// CRITÈRE 2 : La Génération (La version la plus avancée)
		if candidate.Generation > bestVersion.Generation {
			bestVersion = candidate
			continue
		}

		// CRITÈRE 3 : Le Tie-Break (En cas d'égalité parfaite, on prend le hash le plus grand pour être déterministe)
		if candidate.Generation == bestVersion.Generation {
			if candidate.Hash > bestVersion.Hash {
				bestVersion = candidate
			}
		}
	}

	return bestVersion
}
