package engine

// Seegefechte: Händel zwischen Häusern und Angriffe der dänischen Flotte.
// Im Original stellt man die Schiffe auf dem Bildschirm auf, jede Einheit feuert zwei Schüsse mit
// zufälliger Weite, Treffer zählen über Bildschirmzeichen. Das Remake bildet das als Salve nach:
// Trefferwahrscheinlichkeit steigt mit der Zahl der Ziele (dichte Aufstellung = leichter zu treffen).
// Taktischer Modus: der Angreifer wählt das bevorzugte Ziel; schneller Modus: Schiffe zuerst.

// BattleResult beschreibt ein Gefecht für Chronik und Wiedergabe.
type BattleResult struct {
	Attacker int    `json:"attacker"` // Sitz, -1 = Dänen
	Defender int    `json:"defender"`
	City     int    `json:"city"`
	Mode     string `json:"mode"` // "flotte" oder "kontor"
	Focus    string `json:"focus"`
	AttShips int    `json:"attShips"`
	DefShips int    `json:"defShips"`
	Cannons  int    `json:"cannons"`
	Stores   int    `json:"stores"`
	AttLost  int    `json:"attLost"`
	ShipLost int    `json:"shipLost"`
	CanLost  int    `json:"canLost"`
	StoLost  int    `json:"stoLost"`
	AttLoot  int    `json:"attLoot"`
	DefLoot  int    `json:"defLoot"`
}

// hitChance: je mehr Ziele auf dem Wasser, desto eher trifft eine Kugel.
func hitChance(targets int) float64 {
	return min(0.5, 0.08+0.02*float64(targets))
}

// volley: n Einheiten feuern je zwei Schüsse auf einen Zielbestand; liefert die Trefferzahl.
func (g *Game) volley(shooters, targets int) int {
	hits := 0
	for i := 0; i < shooters*2 && hits < targets; i++ {
		if g.Rng.Float() < hitChance(targets-hits) {
			hits++
		}
	}
	return hits
}

// distribute verteilt Treffer auf Schiffe, Kanonen und Speicher; focus bekommt bevorzugt die Treffer.
func (g *Game) distribute(hits int, ships, cannons, stores int, focus string) (ls, lc, lw int) {
	for ; hits > 0; hits-- {
		pick := ""
		switch {
		case focus == "kanonen" && cannons-lc > 0 && g.Rng.Float() < 0.7:
			pick = "c"
		case focus == "speicher" && stores-lw > 0 && g.Rng.Float() < 0.7:
			pick = "w"
		case focus == "schiffe" && ships-ls > 0 && g.Rng.Float() < 0.7:
			pick = "s"
		}
		if pick == "" {
			// gleichmäßig nach Bestand
			rest := (ships - ls) + (cannons - lc) + (stores - lw)
			if rest <= 0 {
				break
			}
			r := g.Rng.IntN(rest)
			switch {
			case r < ships-ls:
				pick = "s"
			case r < ships-ls+cannons-lc:
				pick = "c"
			default:
				pick = "w"
			}
		}
		switch pick {
		case "s":
			ls++
		case "c":
			lc++
		case "w":
			lw++
		}
	}
	return
}

// canFeud: Händel gehen ab dem zweiten Jahr, einmal je Jahr und nur mit Mitspielern.
func (g *Game) canFeud(seat int) bool {
	s := g.Seats[seat]
	return s.Phase == PhaseFeud && !s.Player.Year.Feud && g.gameYear() >= 2 && len(g.Seats) > 1
}

// attack: Händel gegen ein anderes Haus in einer Stadt, in der es ein Kontor hat.
func (g *Game) attack(seat int, a Action) ([]Event, error) {
	if !g.canFeud(seat) {
		return nil, ErrNoFeud
	}
	t, c := a.Target, a.City
	if t < 0 || t >= len(g.Seats) || t == seat || g.Seats[t].Player.Out || c < 2 || c > NumCities {
		return nil, ErrNotPossible
	}
	p, o := g.Seats[seat].Player, g.Seats[t].Player
	if !o.Kontor[c] {
		return nil, ErrNoKontor
	}
	att := min(MaxFleet, p.Sent[c])
	if att == 0 {
		return nil, ErrNoShips
	}
	focus := "schiffe"
	if g.Settings.BattleMode == "tactical" && (a.Stance == "kanonen" || a.Stance == "speicher") {
		focus = a.Stance
	}
	b := BattleResult{Attacker: seat, Defender: t, City: c, Mode: a.Mode, Focus: focus, AttShips: att}
	G := p.Year.Luck
	if a.Mode == "flotte" {
		// Gegen die Handelsflotte: nur Schiffe kämpfen
		b.DefShips = min(MaxFleet, o.Sent[c])
		if b.DefShips == 0 {
			return nil, ErrNoShips
		}
		b.ShipLost = g.volley(att, b.DefShips)
		b.AttLost = g.volley(b.DefShips, att)
		b.AttLoot, b.DefLoot = b.ShipLost*20*G, b.AttLost*20*G
	} else {
		// Gegen das Kontor: Landkanonen, Speicher und dortige Schiffe
		b.Mode = "kontor"
		b.DefShips, b.Cannons, b.Stores = min(14, o.Sent[c]), o.Cannons[c], o.Stores[c]
		hits := g.volley(att, b.DefShips+b.Cannons+b.Stores)
		b.ShipLost, b.CanLost, b.StoLost = g.distribute(hits, b.DefShips, b.Cannons, b.Stores, focus)
		b.AttLost = g.volley(b.DefShips+b.Cannons, att)
		b.AttLoot, b.DefLoot = b.ShipLost*40*G, b.AttLost*40*G
	}
	// Verluste und Beute verbuchen
	p.Sent[c] -= b.AttLost
	p.Underway -= b.AttLost
	o.Sent[c] -= b.ShipLost
	o.Underway -= b.ShipLost
	o.Cannons[c] -= b.CanLost
	o.Stores[c] -= b.StoLost
	p.Mark += b.AttLoot
	o.Mark += b.DefLoot
	p.Year.Feud = true
	p.Feuds++
	o.Attacked++
	p.chronik(g.Year, "Händel mit "+o.Name+" in "+Cities[c].Name)
	o.chronik(g.Year, "Angriff von "+p.Name+" in "+Cities[c].Name)
	return []Event{{"battle", -1, g.Year, b}}, nil
}

// danishAttack: Waldemars Flotte greift ein Kontor an (Ereignis 29/30).
func (g *Game) danishAttack(seat int) []Event {
	p := g.Seats[seat].Player
	c := 1 + g.Rng.IntN(NumCities)
	e := &EventState{ID: 29}
	p.Year.Event = e
	if g.Danish <= 0 || c == Luebeck || !p.Kontor[c] {
		e.Lines = []string{"Der dänische König erwog einen Angriff –", "aber Ihre Majestät wurde durch",
			"diplomatische Intervention umgestimmt."}
		return []Event{{"ereignis", seat, g.Year, e}}
	}
	b := BattleResult{Attacker: -1, Defender: seat, City: c, Mode: "kontor", Focus: "schiffe",
		AttShips: min(DanishMax, g.Danish), DefShips: min(14, p.Sent[c]), Cannons: p.Cannons[c], Stores: p.Stores[c]}
	hits := g.volley(b.AttShips, b.DefShips+b.Cannons+b.Stores)
	b.ShipLost, b.CanLost, b.StoLost = g.distribute(hits, b.DefShips, b.Cannons, b.Stores, "schiffe")
	b.AttLost = g.volley(b.DefShips+b.Cannons, b.AttShips)
	b.DefLoot = b.AttLost * 40 * p.Year.Luck
	p.Sent[c] -= b.ShipLost
	p.Underway -= b.ShipLost
	p.Cannons[c] -= b.CanLost
	p.Stores[c] -= b.StoLost
	p.Mark += b.DefLoot
	p.Attacked++
	g.Danish -= b.AttLost
	e.Lines = []string{"Waldemar IV. greift Ihr Kontor", "in " + Cities[c].Name + " an."}
	e.Battle = b
	p.chronik(g.Year, "Dänenangriff auf "+Cities[c].Name)
	return []Event{{"ereignis", seat, g.Year, e}, {"battle", -1, g.Year, b}}
}
