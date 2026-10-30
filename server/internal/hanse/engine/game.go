package engine

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Phase ist der Abschnitt eines Jahres, den jedes Handelshaus gleichzeitig durchläuft.
// Das Original spielte die Häuser nacheinander; im Remake handeln alle zugleich, das Jahr endet,
// wenn alle fertig sind (wie bei KaiserReloaded).
type Phase string

const (
	PhaseTrade Phase = "trade" // Handel, Börse, Schiffe, Kontore (Hauptmenü des Originals)
	PhaseEvent Phase = "event" // Ereignis des Jahres, ggf. mit Ja/Nein-Frage
	PhaseFeud  Phase = "feud"  // Abschluss: Händel (Angriff) oder Jahr beenden
	PhaseDone  Phase = "done"  // wartet auf die anderen
)

// Seat verbindet ein Handelshaus mit seinem Fortschritt im laufenden Jahr.
type Seat struct {
	Player     *Player `json:"player"`
	Phase      Phase   `json:"phase"`
	PhaseStart int64   `json:"phaseStart"` // Unix-Sekunden
}

// Event ist eine Meldung für Chronik und Hinweise. Player = -1 heißt „an alle“.
type Event struct {
	Type   string `json:"type"`
	Player int    `json:"player"`
	Year   int    `json:"year"`
	Data   any    `json:"data,omitempty"`
}

// Game ist ein laufendes Spiel. Nicht nebenläufig sicher; der Server ruft es aus genau einer
// Goroutine je Spiel auf.
type Game struct {
	Settings Settings       `json:"settings"`
	Balance  *Balance       `json:"balance"`
	Year     int            `json:"year"`
	Seats    []*Seat        `json:"seats"`
	Over     bool           `json:"over"`
	Winner   int            `json:"winner"` // Sitz des Senators, sonst -1
	Danish   int            `json:"danish"` // Stärke der dänischen Flotte
	Bots     map[int]string `json:"bots"`   // Sitz → Spielweise (Computer oder Autopilot)
	CPU      map[int]bool   `json:"cpu"`    // Computer-Häuser (ohne Autopilot menschlicher Spieler)
	Log      []Event        `json:"log"`
	Rng      *Rng           `json:"-"`
}

// MaxSeats begrenzt die Handelshäuser je Partie (Original: 1–6; Server-Flag -max-players).
var MaxSeats = 6

// NewGame legt ein Spiel ohne Spieler an.
func NewGame(s Settings, b *Balance, seed uint64) *Game {
	return &Game{Settings: s, Balance: b, Year: StartYear, Rng: NewRng(seed), Winner: -1, Danish: DanishStart, CPU: map[int]bool{}}
}

// AddPlayer gründet ein Handelshaus vor Spielbeginn.
func (g *Game) AddPlayer(name string, female bool) (int, error) {
	if len(g.Seats) >= MaxSeats {
		return 0, &RuleError{"room_full", "Alle Plätze im Rat sind vergeben."}
	}
	p := newPlayer(name, female)
	p.LifeLeft += g.Settings.lifeOffset()
	p.Arms.Variant = len(g.Seats) % MaxArmsVariants // jedes Haus beginnt mit einer anderen Hausmarke
	g.Seats = append(g.Seats, &Seat{Player: p, Phase: PhaseDone})
	g.syncShares()
	return len(g.Seats) - 1, nil
}

// JoinRunning nimmt ein Haus ins laufende Spiel auf; es steigt mit dem nächsten Jahr ein.
func (g *Game) JoinRunning(name string, female bool) (int, error) {
	seat, err := g.AddPlayer(name, female)
	if err == nil {
		g.Seats[seat].Player.Year.Joined = true
	}
	return seat, err
}

// Promoted meldet, ob schon ein Haus über den ersten Rang aufgestiegen ist. Ab dann ist kein
// Beitritt mehr möglich.
func (g *Game) Promoted() bool {
	for _, s := range g.Seats {
		if s.Player.Rank > 1 {
			return true
		}
	}
	return false
}

// RemoveSeat entfernt ein Haus vor Spielbeginn; spätere Sitze rücken auf.
func (g *Game) RemoveSeat(seat int) {
	g.Seats = append(g.Seats[:seat], g.Seats[seat+1:]...)
	for _, s := range g.Seats {
		p := s.Player
		if seat < len(p.Shares) {
			p.Shares = append(p.Shares[:seat], p.Shares[seat+1:]...)
		}
	}
	shift := func(k int) int {
		if k > seat {
			return k - 1
		}
		return k
	}
	bots, cpu := map[int]string{}, map[int]bool{}
	for k, v := range g.Bots {
		if k != seat {
			bots[shift(k)] = v
		}
	}
	for k, v := range g.CPU {
		if k != seat {
			cpu[shift(k)] = v
		}
	}
	g.Bots, g.CPU = bots, cpu
}

// syncShares bringt die Anteilslisten auf die Zahl der Häuser; der Gründer hält OwnShares.
func (g *Game) syncShares() {
	for i, s := range g.Seats {
		p := s.Player
		for len(p.Shares) < len(g.Seats) {
			p.Shares = append(p.Shares, 0)
		}
		if p.Shares[i] == 0 && p.FreeShares == TotalShares-OwnShares {
			p.Shares[i] = OwnShares
		}
	}
}

// Start beginnt das erste Jahr.
func (g *Game) Start(now int64) []Event {
	return g.startYear(now)
}

// startYear: Rückkehr der Schiffe, Verschleiß, Ankünfte, Preise und Index je Haus.
func (g *Game) startYear(now int64) []Event {
	var ev []Event
	for i, s := range g.Seats {
		p := s.Player
		s.Phase, s.PhaseStart = PhaseDone, now
		if p.Out {
			continue
		}
		p.Year = YearState{Luck: g.rollLuck(), Sea: 1 + g.Rng.IntN(4)}
		p.Ships += p.Underway
		p.Underway = 0
		g.cityPrices(p)
		if p.Prison > 0 {
			// Ein Jahr im Schuldturm: keine Geschäfte, nur die Preise laufen weiter
			p.Prison--
			p.Year.Skipped = true
			g.shipPrices(p)
			g.marketPrices(p)
			ev = append(ev, Event{"prison", i, g.Year, nil})
			continue
		}
		g.wearAndArrivals(p)
		g.marketPrices(p)
		g.updateIndex(i)
		s.Phase = PhaseTrade
		ev = append(ev, Event{"arrivals", i, g.Year, map[string]any{
			"arrived": p.Year.Arrived, "sea": p.Year.Sea, "lost": p.Year.Lost, "delivered": p.Year.Delivered,
		}})
	}
	if g.allDone() && !g.Over {
		ev = append(ev, g.endYear(now)...)
	}
	return ev
}

// Action ist ein Spielzug vom Client. Nur die zur Art passenden Felder zählen.
type Action struct {
	Type   string    `json:"type"`
	Qty    int       `json:"qty,omitempty"`
	Good   int       `json:"good,omitempty"`
	City   int       `json:"city,omitempty"`
	Target int       `json:"target,omitempty"` // Sitz (Börse, Händel)
	Mode   string    `json:"mode,omitempty"`   // Händel: "flotte" oder "kontor"
	Stance string    `json:"stance,omitempty"` // Ereignis: "ja"/"nein"; Gefecht: Zielwahl
	Name   string    `json:"name,omitempty"`   // Name des Kindes
	Motto  string    `json:"motto,omitempty"`
	Arms   *ArmsSpec `json:"arms,omitempty"`
}

// Apply führt eine Aktion aus.
func (g *Game) Apply(seat int, a Action, now int64) ([]Event, error) {
	if g.Over {
		return nil, &RuleError{"game_over", "Das Spiel ist beendet."}
	}
	if seat < 0 || seat >= len(g.Seats) {
		return nil, &RuleError{"bad_seat", "Unbekannter Platz."}
	}
	s := g.Seats[seat]
	p := s.Player
	if a.Type == "profile" {
		return nil, setProfile(p, a.Motto, a.Arms)
	}
	if p.Out {
		return nil, &RuleError{"out", "Ihr Haus ist ausgestorben."}
	}
	if a.Qty < 0 {
		return nil, ErrInvalidAmount
	}
	var ev []Event
	var err error
	switch s.Phase {
	case PhaseTrade:
		if a.Type == "next" {
			g.advance(s, PhaseEvent, now)
			ev = g.drawEvent(seat)
			if p.Year.Event == nil || p.Year.Event.Ask == "" {
				g.advance(s, PhaseFeud, now)
			}
			break
		}
		ev, err = g.trade(seat, a)
	case PhaseEvent:
		if a.Type != "decide" {
			return nil, ErrWrongPhase
		}
		ev, err = g.decide(seat, a.Stance == "ja", a.Name)
		if err == nil {
			g.advance(s, PhaseFeud, now)
		}
	case PhaseFeud:
		switch a.Type {
		case "attack":
			ev, err = g.attack(seat, a)
		case "endTurn":
			g.advance(s, PhaseDone, now)
		default:
			return nil, ErrWrongPhase
		}
	default:
		return nil, ErrWrongPhase
	}
	if err != nil {
		return nil, err
	}
	if g.allDone() {
		ev = append(ev, g.endYear(now)...)
	}
	return ev, nil
}

// advance wechselt die Phase.
func (g *Game) advance(s *Seat, next Phase, now int64) {
	s.Phase, s.PhaseStart = next, now
}

// Deadline liefert das Ende der laufenden Phase (Unix-Sekunden) oder 0 ohne Limit.
func (g *Game) Deadline(seat int) int64 {
	s := g.Seats[seat]
	if g.Settings.PhaseSeconds == 0 || s.Phase == PhaseDone {
		return 0
	}
	return s.PhaseStart + int64(g.Settings.PhaseSeconds)
}

// Timeouts beendet abgelaufene Phasen automatisch, damit niemand das Spiel aufhält.
func (g *Game) Timeouts(now int64) []Event {
	if g.Settings.PhaseSeconds == 0 || g.Over {
		return nil
	}
	var ev []Event
	for i, s := range g.Seats {
		if s.Phase == PhaseDone || now-s.PhaseStart < int64(g.Settings.PhaseSeconds) {
			continue
		}
		ev = append(ev, Event{"timeout", i, g.Year, s.Phase})
		var a Action
		switch s.Phase {
		case PhaseTrade:
			a = Action{Type: "next"}
		case PhaseEvent:
			a = Action{Type: "decide", Stance: "nein"}
		case PhaseFeud:
			a = Action{Type: "endTurn"}
		}
		more, _ := g.Apply(i, a, now)
		ev = append(ev, more...)
	}
	return ev
}

func (g *Game) allDone() bool {
	for _, s := range g.Seats {
		if s.Phase != PhaseDone {
			return false
		}
	}
	return true
}

// endYear: Lager kappen, Beförderung, Alter und Tod, Schuldturm, Zinsen/Dividenden;
// danach beginnt das nächste Jahr.
func (g *Game) endYear(now int64) []Event {
	var ev []Event
	for i, s := range g.Seats {
		p := s.Player
		// Neu Beigetretene und Häuser im Schuldturm altern in diesem Jahr nicht
		if p.Out || p.Year.Skipped || p.Year.Joined {
			continue
		}
		capStock(p)
		// Beförderung: Index > R·100 + R·G
		if p.Rank < MaxRank && p.Index > p.Rank*100+p.Rank*p.Year.Luck {
			p.Rank++
			p.LifeLeft += 10 - p.Rank/2
			g.Danish = min(DanishMax, g.Danish+1)
			gd := Guilds[min(p.Rank, MaxRank-1)]
			p.chronik(g.Year, "Aufnahme "+gd.Body)
			ev = append(ev, Event{"promotion", i, g.Year, map[string]any{"rank": p.Rank, "body": gd.Body, "verb": gd.Verb}})
			if p.Rank >= MaxRank {
				g.Over, g.Winner = true, i
				ev = append(ev, Event{"gameOver", -1, g.Year, map[string]any{"winner": i}})
			}
		}
		// Alter: ein Lebensjahr weniger, bei 0 Tod und Erbfolge
		if g.Settings.Mortality != "unsterblich" {
			p.LifeLeft--
		}
		if p.LifeLeft <= 0 {
			ev = append(ev, g.die(i)...)
		}
		// Schulden: Schuldturm, man arbeitet bis 5000 Mark ab
		if !p.Out && p.Mark < 0 && p.Prison == 0 {
			work := min(PrisonCap, -p.Mark)
			p.Mark += work
			p.Prison = 1
			p.Prisoned++
			p.chronik(g.Year, "Schuldturm")
			ev = append(ev, Event{"schuldturm", i, g.Year, map[string]any{"work": work}})
		}
	}
	g.dividends()
	if !g.Over {
		// Original: Das Spiel endet, sobald jedes Haus zweimal sein Oberhaupt verloren hat
		mourning := 0
		for _, s := range g.Seats {
			if s.Player.Out || s.Player.Generation >= SeucheDeaths {
				mourning++
			}
		}
		g.Year++
		if mourning == len(g.Seats) || g.Year >= StartYear+MaxYears {
			g.Over = true
			ev = append(ev, Event{"gameOver", -1, g.Year, map[string]any{"winner": -1, "seuche": true}})
		}
	}
	if !g.Over {
		ev = append(ev, Event{"newYear", -1, g.Year, nil})
		ev = append(ev, g.startYear(now)...)
	}
	return ev
}

// die: Tod des Herrschers, Erbfolge über das Kind oder einen benannten Erben.
func (g *Game) die(seat int) []Event {
	p := g.Seats[seat].Player
	ev := []Event{{"died", seat, g.Year, map[string]any{"name": p.Name}}}
	p.chronik(g.Year, "Tod von "+p.Name)
	p.Generation++
	if p.Generation > MaxGenerations {
		p.Out = true
		g.Seats[seat].Phase = PhaseDone
		ev = append(ev, Event{"extinct", seat, g.Year, nil})
		return ev
	}
	if p.Child != "" {
		p.Name, p.Female = p.Child, p.ChildFem
	} else {
		// Ohne Kind übernimmt ein Verwandter; die übrigen Erben werden ausgezahlt
		payout := p.Index * (g.epoch() + 3)
		p.Mark -= payout
		p.Name = p.Name + " d. J."
		ev = append(ev, Event{"payout", seat, g.Year, map[string]any{"cost": payout}})
	}
	p.Child, p.Married = "", false
	p.LifeLeft = StartLife + g.Settings.lifeOffset()
	ev = append(ev, Event{"heir", seat, g.Year, map[string]any{"name": p.Name}})
	return ev
}

// dividends: Zinsen auf eigene freie Anteile, Dividenden aus Beteiligungen.
func (g *Game) dividends() {
	for _, s := range g.Seats {
		p := s.Player
		if p.Out {
			continue
		}
		p.Year.Interest = int(float64(p.Index*p.FreeShares) * 0.3)
		p.Mark -= p.Year.Interest
	}
	for i, s := range g.Seats {
		p := s.Player
		if p.Out {
			continue
		}
		for j, n := range p.Shares {
			if j == i || n == 0 || j >= len(g.Seats) {
				continue
			}
			o := g.Seats[j].Player
			d := int(0.3 * float64(o.Index) * float64(n))
			p.Mark += d
			p.Year.Dividend += d
			o.Mark -= d
			o.Year.Interest += d
		}
	}
}

// PublicPlayer ist die für alle sichtbare Übersicht eines Hauses.
type PublicPlayer struct {
	Seat       int      `json:"seat"`
	Name       string   `json:"name"`
	Female     bool     `json:"female"`
	Rank       int      `json:"rank"`
	Index      int      `json:"index"`
	Mark       int      `json:"mark"`
	Ships      int      `json:"ships"`
	Kontore    []int    `json:"kontore"`
	Generation int      `json:"generation"`
	Out        bool     `json:"out"`
	Phase      Phase    `json:"phase"`
	Bot        string   `json:"bot"`
	Motto      string   `json:"motto"`
	Arms       ArmsSpec `json:"arms"`
	FreeShares int      `json:"freeShares"`
	IndexHist  []int    `json:"indexHist"`
}

// Public liefert die Übersicht aller Häuser.
func (g *Game) Public() []PublicPlayer {
	out := make([]PublicPlayer, len(g.Seats))
	for i, s := range g.Seats {
		p := s.Player
		var k []int
		for c := 1; c <= NumCities; c++ {
			if p.Kontor[c] {
				k = append(k, c)
			}
		}
		out[i] = PublicPlayer{Seat: i, Name: p.Name, Female: p.Female, Rank: p.Rank, Index: p.Index, Mark: p.Mark,
			Ships: p.TotalShips(), Kontore: k, Generation: p.Generation, Out: p.Out, Phase: s.Phase, Bot: g.Bots[i],
			Motto: p.Motto, Arms: p.Arms, FreeShares: p.FreeShares, IndexHist: p.IndexHist}
	}
	return out
}

// SelfView ist die vollständige Sicht eines Hauses auf sich selbst.
type SelfView struct {
	Seat     int     `json:"seat"`
	Phase    Phase   `json:"phase"`
	Deadline int64   `json:"deadline"`
	Player   *Player `json:"player"`
	Year     int     `json:"year"`
	Epoch    int     `json:"epoch"`
	Exchange bool    `json:"exchange"` // Börse geöffnet (ab dem 2. Jahr)
	CanFeud  bool    `json:"canFeud"`
}

// Self liefert die eigene Sicht.
func (g *Game) Self(seat int) SelfView {
	s := g.Seats[seat]
	return SelfView{Seat: seat, Phase: s.Phase, Deadline: g.Deadline(seat), Player: s.Player, Year: g.Year,
		Epoch: g.epoch(), Exchange: g.gameYear() > 1, CanFeud: g.canFeud(seat)}
}

// RestoreGame stellt ein gespeichertes Spiel wieder her.
func RestoreGame(data []byte) (*Game, error) {
	var aux struct {
		Game
		RngState []byte `json:"rng"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return nil, fmt.Errorf("snapshot lesen: %w", err)
	}
	g := aux.Game
	g.Rng = &Rng{}
	if err := g.Rng.UnmarshalBinary(aux.RngState); err != nil {
		return nil, fmt.Errorf("zufallszustand: %w", err)
	}
	if g.Balance == nil {
		g.Balance = DefaultBalance()
	}
	return &g, nil
}

// Snapshot serialisiert das Spiel samt Zufallszustand.
func (g *Game) Snapshot() ([]byte, error) {
	rs, err := g.Rng.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		*Game
		RngState []byte `json:"rng"`
	}{g, rs})
}

// itoa kürzt strconv.Itoa für Meldungstexte.
func itoa(n int) string { return strconv.Itoa(n) }
