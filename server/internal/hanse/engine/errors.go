package engine

// RuleError ist ein Regelverstoß, den der Client dem Spieler direkt anzeigen darf.
// Code ist maschinenlesbar, Message der deutsche Text.
type RuleError struct {
	Code    string
	Message string
}

func (e *RuleError) Error() string { return e.Message }

// Bekannte Regelverstöße; „Nicht möglich!“ ist die Sammelmeldung des Originals.
var (
	ErrInvalidAmount  = &RuleError{"invalid_amount", "Ungültige Menge."}
	ErrNotPossible    = &RuleError{"not_possible", "Nicht möglich!"}
	ErrNoKontor       = &RuleError{"no_kontor", "Dort haben Sie kein Kontor."}
	ErrBadGood        = &RuleError{"bad_good", "Diese Ware gibt es nicht."}
	ErrExchangeClosed = &RuleError{"exchange_closed", "Die Börse ist im ersten Jahr geschlossen."}
	ErrWrongPhase     = &RuleError{"wrong_phase", "Das geht in dieser Phase nicht."}
	ErrNoFeud         = &RuleError{"no_feud", "Händel sind gerade nicht möglich."}
	ErrNoShips        = &RuleError{"no_ships", "Keine Schiffe in der Nähe."}
)
