package engine

// Ereignisse des Jahres (Zufallszahl 1–30, Zuordnung wie im Original).
// Texte nach dem Original, in normaler Groß-/Kleinschreibung.

// EventState ist das Ereignis eines Hauses im laufenden Jahr.
type EventState struct {
	ID     int      `json:"id"`
	Lines  []string `json:"lines"`            // Meldung
	Ask    string   `json:"ask,omitempty"`    // "jn" = Ja/Nein-Frage, "name" = Name des Kindes
	Result []string `json:"result,omitempty"` // Ausgang nach der Antwort
	Delta  int      `json:"delta"`            // Mark-Änderung
	Battle any      `json:"battle,omitempty"` // Dänenangriff
}

// eventTable ordnet die Zufallszahl der Ereignisart zu (9 und 18 gehören zur jeweils vorherigen
// Gruppe, 28 bleibt ohne Ereignis).
var eventTable = [31]string{"",
	"erbe", "brand_luebeck", "gewinn", "krank", "veruntreuung", "bergen", "brand_luebeck", "gewinn",
	"veruntreuung", "fest", "audienz", "zoll_dk", "zoll_se", "unwetter", "witwen", "dom", "raub",
	"piraten", "brand", "wechsel", "schmuggel", "belagerung", "heirat", "kind", "gewinn", "kind", "kind",
	"", "daenen", "daenen",
}

// drawEvent zieht das Ereignis beim Wechsel in die Ereignisphase und wendet sofortige Folgen an.
func (g *Game) drawEvent(seat int) []Event {
	p := g.Seats[seat].Player
	r := 1 + g.Rng.IntN(30)
	e := &EventState{ID: r}
	p.Year.Event = e
	G, R := p.Year.Luck, p.Rank
	mark := func(d int) { e.Delta = d; p.Mark += d }
	switch eventTable[r] {
	case "erbe":
		if p.Stores[Luebeck] > 4 {
			return g.noEvent(seat)
		}
		p.Stores[Luebeck]++
		e.Lines = []string{"Sie haben einen Speicher", "in Lübeck geerbt."}
	case "brand_luebeck":
		if p.Stores[Luebeck] < 1 {
			return g.noEvent(seat)
		}
		p.Stores[Luebeck]--
		e.Lines = []string{"Ein Speicher ist in", "Lübeck abgebrannt."}
	case "gewinn":
		d := G * 200 * R
		mark(d)
		e.Lines = []string{"Sie gewannen " + itoa(d) + " Mark."}
	case "krank":
		e.Lines, e.Ask = []string{"Sie sind erkrankt.", "Wollen Sie einen Arzt?"}, "jn"
	case "veruntreuung":
		e.Delta = (G + 1) * 978
		e.Lines, e.Ask = []string{"Ein Schreiber veruntreute " + itoa(e.Delta) + " Mark.", "Wollen Sie einen Prozess?"}, "jn"
	case "bergen":
		e.Lines, e.Ask = []string{"Wollen Sie ein Schiff bergen?"}, "jn"
	case "fest":
		e.Lines, e.Ask = []string{"Geben Sie ein Wohltätigkeitsfest?"}, "jn"
	case "audienz":
		e.Lines, e.Ask = []string{"Bitten Sie um eine Audienz", "beim Kurfürsten von Sachsen?"}, "jn"
	case "zoll_dk":
		d := -R * g.Rng.IntN(1000)
		mark(d)
		e.Lines = []string{"Waldemar IV. von Dänemark", "fordert von Ihnen " + itoa(-d) + " Mark Zoll."}
	case "zoll_se":
		d := -R * g.Rng.IntN(800)
		mark(d)
		e.Lines = []string{"Der Schwedenkönig, Erich der XIII.,", "fordert von Ihnen " + itoa(-d) + " Mark Zoll."}
	case "unwetter":
		if p.Stock[Salt] == 0 {
			return g.noEvent(seat)
		}
		p.Stock[Salt] /= 2
		e.Lines = []string{"Sie verloren die Hälfte Ihres Salzes", "bei einem Unwetter."}
	case "witwen":
		e.Delta = -max(0, p.Index*20-G)
		e.Lines, e.Ask = []string{"Spenden Sie für das Witwenstift?"}, "jn"
	case "dom":
		e.Delta = -500
		e.Lines, e.Ask = []string{"Spenden Sie für den Dombau 500 Mark?"}, "jn"
	case "raub":
		d := -int(g.Rng.Float() * 1000 * float64(R))
		mark(d)
		e.Lines = []string{"Ihr Kontor in Lübeck wurde", "um " + itoa(-d) + " Mark beraubt."}
	case "piraten":
		c := 0
		for i := 2; i <= NumCities && c == 0; i++ {
			if p.Sent[i] > 0 {
				c = i
			}
		}
		if c == 0 {
			return g.noEvent(seat)
		}
		p.Sent[c]--
		p.Underway--
		e.Lines = []string{"Eines Ihrer Schiffe wurde", "von Seeräubern geentert."}
	case "brand":
		c := min(9, g.Rng.IntN(R))
		if c == 0 || p.Stores[c] == 0 || (c == Luebeck && p.Stores[c] < 2) {
			return g.noEvent(seat)
		}
		p.Stores[c]--
		e.Lines = []string{"Ein Speicher in " + Cities[c].Name, "brannte ab."}
	case "wechsel":
		mark(3000)
		e.Lines = []string{"Aus einem Wechselgeschäft", "erhalten Sie 3000 Mark."}
	case "schmuggel":
		e.Lines, e.Ask = []string{"Ihnen wird ein Schmuggelgeschäft", "angeboten. Nehmen Sie an?"}, "jn"
	case "belagerung":
		c := 1 + g.Rng.IntN(NumCities)
		p.Sent[c] = 0 // die Schiffe kehren ohne Ladung zurück
		e.Lines = []string{"Waldemar IV. belagert die Stadt", Cities[c].Name + "."}
	case "heirat":
		if p.Married {
			return g.noEvent(seat)
		}
		e.Lines, e.Ask = []string{"Wollen Sie in eine wohlhabende", "Familie einheiraten?"}, "jn"
	case "kind":
		if !p.Married || p.Child != "" {
			return g.noEvent(seat)
		}
		p.ChildFem = g.Rng.IntN(2) == 1
		kind := "einen Sohn"
		if p.ChildFem {
			kind = "eine Tochter"
		}
		e.Lines, e.Ask = []string{"Sie haben " + kind + " bekommen!", "Wie soll das Kind heißen?"}, "name"
	case "daenen":
		return g.danishAttack(seat)
	default:
		return g.noEvent(seat)
	}
	return []Event{{"ereignis", seat, g.Year, e}}
}

// noEvent: ruhiges Jahr ohne Ereignis.
func (g *Game) noEvent(seat int) []Event {
	e := &EventState{ID: 0, Lines: []string{"Ein ruhiges Jahr."}}
	g.Seats[seat].Player.Year.Event = e
	return []Event{{"ereignis", seat, g.Year, e}}
}

// luckless: „Pech“ des Originals (bei Glückswert über 4).
func luckless(p *Player) bool { return p.Year.Luck > 4 }

// decide beantwortet die Frage des Ereignisses.
func (g *Game) decide(seat int, yes bool, name string) ([]Event, error) {
	p := g.Seats[seat].Player
	e := p.Year.Event
	if e == nil || e.Ask == "" {
		return nil, ErrWrongPhase
	}
	G, R := p.Year.Luck, p.Rank
	mark := func(d int) { e.Delta = d; p.Mark += d }
	switch eventTable[e.ID] {
	case "krank":
		if yes {
			d := -(G + p.Index/10)
			mark(d)
			e.Result = []string{"Arztkosten: " + itoa(-d) + " Mark."}
		} else {
			d := -(G + p.Index) * 2
			mark(d)
			p.LifeLeft--
			e.Result = []string{"Hospital-Kosten " + itoa(-d) + " Mark.", "Die Krankheit kostet Sie ein Lebensjahr."}
		}
	case "veruntreuung", "bergen":
		amount := e.Delta
		if eventTable[e.ID] == "bergen" {
			amount = G + p.Index*20
		}
		e.Delta = 0
		switch {
		case !yes:
			e.Result = []string{"Sie verzichten."}
		case luckless(p):
			d := -amount / 5
			mark(d)
			e.Result = []string{"Sie hatten Pech – Kosten " + itoa(-d) + " Mark."}
		default:
			e.Result = []string{"Sie waren erfolgreich!"}
			if eventTable[e.ID] == "bergen" {
				p.Orders++
			} else {
				mark(amount)
			}
		}
	case "fest":
		if yes {
			mark(-400 * R)
			p.LifeLeft++
			e.Result = []string{"Kosten " + itoa(400*R) + " Mark."}
		}
	case "audienz":
		if yes {
			if luckless(p) {
				mark(-100 * R)
				e.Result = []string{"Der Kurfürst ist verstimmt.", "Reisekosten " + itoa(100*R) + " Mark."}
			} else {
				mark(g.Year * R)
				e.Result = []string{"Für Ihre Dienste " + itoa(g.Year*R) + " Mark."}
			}
		}
	case "witwen", "dom":
		d := e.Delta
		e.Delta = 0
		if yes {
			mark(d)
			p.LifeLeft++
			e.Result = []string{"Kosten " + itoa(-d) + " Mark."}
		}
	case "schmuggel":
		if yes {
			if !luckless(p) {
				// Im Original führt gerade ein niedriger Glückswert in den Schuldturm
				p.Prison = 1
				p.Prisoned++
				p.chronik(g.Year, "Schuldturm (Schmuggel)")
				e.Result = []string{"Sie hatten Pech, 1 Jahr Schuldturm."}
			} else {
				mark(10*g.Year - G*111)
				e.Result = []string{"Gewinn: " + itoa(e.Delta) + " Mark."}
			}
		}
	case "heirat":
		if yes {
			mark(2*R*g.Year + G*1511)
			p.Married = true
			p.chronik(g.Year, "Heirat")
			e.Result = []string{"Ihr Kapital steigt um " + itoa(e.Delta) + " Mark."}
		}
	case "kind":
		n, ok := CleanText(name, 12, false)
		if ok != TextOK {
			n = "Erbe"
			if p.ChildFem {
				n = "Erbin"
			}
		}
		p.Child = n
		p.chronik(g.Year, "Geburt von "+n)
		e.Result = []string{n + " ist Ihr Erbe."}
	}
	// Antwort festhalten, damit Chronik und Fenster nicht nur die Frage zeigen (auch bei Ablehnung
	// ohne Folgen und bei abgelaufener Zeit, die als „Nein“ zählt)
	if e.Ask == "jn" {
		antwort := "Ihre Antwort: Nein."
		if yes {
			antwort = "Ihre Antwort: Ja."
		}
		if len(e.Result) == 0 && !yes {
			e.Result = []string{"Sie lassen die Gelegenheit verstreichen."}
		}
		e.Result = append([]string{antwort}, e.Result...)
	}
	e.Ask = ""
	return []Event{{"ereignis", seat, g.Year, e}}, nil
}
