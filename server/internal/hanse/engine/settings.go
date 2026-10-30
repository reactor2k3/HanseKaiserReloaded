package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Balance enthält die wenigen einstellbaren Rahmenwerte (data/balance.json); die Spielformeln
// stehen bewusst im Code.
type Balance struct {
	PhaseSeconds int `json:"phaseSeconds"` // Vorgabe der Zugzeit in der Lobby
}

// DefaultBalance liefert die eingebauten Vorgaben.
func DefaultBalance() *Balance { return &Balance{PhaseSeconds: 180} }

// LoadBalance liest data/balance.json; fehlende Felder behalten die Vorgaben.
func LoadBalance(path string) (*Balance, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("balance lesen: %w", err)
	}
	b := DefaultBalance()
	if err := json.Unmarshal(raw, b); err != nil {
		return nil, fmt.Errorf("balance parsen: %w", err)
	}
	return b, nil
}

// Settings sind die Einstellungen, die der Gastgeber pro Spiel in der Lobby wählt.
type Settings struct {
	// PhaseSeconds ist die Zugzeit je Phase; 0 = ohne Zeitlimit.
	PhaseSeconds int `json:"phaseSeconds"`
	// BattleMode: "tactical" (Zielwahl im Gefecht) oder "quick" (automatisch).
	BattleMode string `json:"battleMode"`
	// Mortality: Lebenserwartung – "unsterblich", "lang", "normal" (Original), "kurz".
	Mortality string `json:"mortality,omitempty"`
	// Difficulty: Spielstärke der Computer-Häuser – "leicht", "mittel", "schwer".
	Difficulty string `json:"difficulty,omitempty"`
	// Mode: "live" (kurze Zugzeiten, ruht ohne Spieler) oder "async" (Stunden bis Tage).
	Mode string `json:"mode,omitempty"`
}

// Async meldet ein asynchrones Spiel.
func (s Settings) Async() bool { return s.Mode == "async" }

// lifeSpans: zusätzliche Lebensjahre je Einstellung (Original: 12 Jahre, „normal“).
var lifeSpans = map[string]int{"unsterblich": 0, "lang": 6, "normal": 0, "kurz": -4}

// lifeOffset liefert die zusätzlichen Lebensjahre eines neuen Herrschers.
func (s Settings) lifeOffset() int { return lifeSpans[s.Mortality] }

// Zugzeit-Auswahl (Sekunden je Phase); 0 = unbegrenzt. Asynchron: 1 h bis 3 Tage.
var (
	PhaseSecondsOptions = []int{0, 120, 180, 300, 600}
	AsyncSecondsOptions = []int{3600, 21600, 43200, 86400, 259200}
)

// DefaultSettings liefert die Lobby-Vorgaben.
func DefaultSettings(b *Balance) Settings {
	return Settings{PhaseSeconds: b.PhaseSeconds, BattleMode: "tactical", Mortality: "normal", Difficulty: "mittel", Mode: "live"}
}

// Validate prüft die Lobby-Eingaben.
func (s Settings) Validate() error {
	if s.Mode != "" && s.Mode != "live" && s.Mode != "async" {
		return &RuleError{"invalid_mode", "Ungültiger Spielmodus."}
	}
	options := PhaseSecondsOptions
	if s.Async() {
		options = AsyncSecondsOptions
	}
	ok := false
	for _, v := range options {
		ok = ok || s.PhaseSeconds == v
	}
	if !ok {
		return &RuleError{"invalid_phase_seconds", "Ungültige Zugzeit."}
	}
	if s.BattleMode != "tactical" && s.BattleMode != "quick" {
		return &RuleError{"invalid_battle_mode", "Ungültiger Gefechtsmodus."}
	}
	if _, ok := botSkills[s.Difficulty]; s.Difficulty != "" && !ok {
		return &RuleError{"invalid_difficulty", "Ungültiger Schwierigkeitsgrad."}
	}
	if _, ok := lifeSpans[s.Mortality]; s.Mortality != "" && !ok {
		return &RuleError{"invalid_mortality", "Ungültige Lebenserwartung."}
	}
	return nil
}

// ArmsSpec beschreibt das Familienwappen: Hausmarke aus der Auswahl (client/public/heraldik/hausmarken) oder Upload.
type ArmsSpec struct {
	Kind    string `json:"kind"`    // house | upload
	Variant int    `json:"variant"` // Hausmarke: Nummer der Vorlage
	Media   string `json:"media"`   // Upload: Dateiname unter /media/
}

// MaxArmsVariants: Zahl der Hausmarken (Dateien hausmarken/0–7.svg im Client).
const MaxArmsVariants = 8

var mediaName = regexp.MustCompile(`^[0-9a-f]{32}\.(png|svg)$`)

// ErrBadProfile meldet ungültige Profilangaben.
var ErrBadProfile = &RuleError{"bad_profile", "Ungültige Profilangaben."}

// MaxMotto: Länge des Wahlspruchs.
const MaxMotto = 40

// ValidMedia prüft einen Dateinamen aus dem Upload (Zufalls-ID + Endung).
func ValidMedia(name string) bool { return mediaName.MatchString(name) }

// Validate prüft ein Wappen gegen die erlaubten Werte.
func (a ArmsSpec) Validate() bool {
	switch a.Kind {
	case "house":
		return a.Variant >= 0 && a.Variant < MaxArmsVariants
	case "upload":
		return ValidMedia(a.Media)
	}
	return false
}

// setProfile ändert Wahlspruch und Wappen (in jeder Phase erlaubt).
func setProfile(p *Player, motto string, arms *ArmsSpec) error {
	m, problem := CleanText(motto, MaxMotto, true)
	if problem != TextOK {
		return &RuleError{"bad_motto", TextMessage(problem, "Der Wahlspruch", MaxMotto)}
	}
	if arms != nil {
		if !arms.Validate() {
			return ErrBadProfile
		}
		p.Arms = *arms
	}
	p.Motto = m
	return nil
}

// SetUploadedArms setzt ein hochgeladenes Wappen.
func (g *Game) SetUploadedArms(seat int, media string) error {
	if seat < 0 || seat >= len(g.Seats) || !ValidMedia(media) || strings.Contains(media, "/") {
		return ErrBadProfile
	}
	g.Seats[seat].Player.Arms = ArmsSpec{Kind: "upload", Media: media}
	return nil
}
