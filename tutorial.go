package main

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"time"
)

// --- Structures Tutoriel (RFC-0001) ---

type Step struct {
	Index       int               `json:"index"`
	ImageHash   string            `json:"img_hash"`
	TextContent map[string]string `json:"text_content"` // ex: "en": "Open case", "fr": "Ouvrir..."
}

type ToolReference struct {
	ProductID string `json:"product_id"` // Lien vers un ProductObject
	Quantity  int    `json:"qty"`
}

type TutorialObject struct {
	SchemaVersion uint64            `json:"schema_v"`
	ID            string            `json:"id"`        // UUID ou Hash
	TargetProduct string            `json:"target_id"` // ID du produit réparé
	Title         map[string]string `json:"title"`     // Titre multilingue
	Tools         []ToolReference   `json:"tools"`
	Steps         []Step            `json:"steps"`
	Hash          string            `json:"hash"`        // Le hash de CET objet (son ID unique)
	ParentHash    string            `json:"parent_hash"` // Le hash de la version précédente (vide si c'est la v1)
	Generation    uint64            `json:"generation"`  // v1=1, v2=2... (pour trier facilement)

	// Metadonnées pour le CRDT (Last-Write-Wins simplifié)
	CRDTMeta struct {
		UpdatedAt    int64  `json:"updated_at"`
		AuthorDevice string `json:"author_device"`
	} `json:"crdt_meta"`

	Signature []byte `json:"signature"`
}

// --- Logique Tutoriel ---

// CreateTutorial initialise un tuto vide lié à un produit
func CreateTutorial(author *Device, productID string, defaultLang, title string) *TutorialObject {
	tuto := &TutorialObject{
		SchemaVersion: 1,
		ID:            fmt.Sprintf("tuto_%d_%s", time.Now().Unix(), author.ID[:6]), // ID simple pour le PoC
		TargetProduct: productID,
		Title:         map[string]string{defaultLang: title},
		Tools:         []ToolReference{},
		Steps:         []Step{},
	}
	tuto.UpdateMeta(author)
	return tuto
}

// AddStep ajoute une étape (en anglais par exemple)
func (t *TutorialObject) AddStep(imgHash string, lang, text string) {
	newStep := Step{
		Index:       len(t.Steps) + 1,
		ImageHash:   imgHash,
		TextContent: map[string]string{lang: text},
	}
	t.Steps = append(t.Steps, newStep)
}

// TranslateStep permet à un AUTRE device d'ajouter une traduction
func (t *TutorialObject) TranslateStep(stepIndex int, lang, text string, author *Device) {
	if stepIndex > 0 && stepIndex <= len(t.Steps) {
		t.Steps[stepIndex-1].TextContent[lang] = text
		// L'auteur change car c'est une modification
		t.UpdateMeta(author)
		// Il faudra re-signer l'objet après ça !
	}
}

// UpdateMeta met à jour timestamp et auteur
func (t *TutorialObject) UpdateMeta(author *Device) {
	t.CRDTMeta.UpdatedAt = time.Now().Unix()
	t.CRDTMeta.AuthorDevice = author.ID
}

// SignTutorial sérialise et signe (Même logique que Product)
func (t *TutorialObject) SignTutorial(signer *Device) {
	// Copie temporaire sans signature pour calcul du hash
	temp := *t
	temp.Signature = nil

	payload, _ := json.Marshal(temp)
	t.Signature = ed25519.Sign(signer.PrivateKey, payload)
}
