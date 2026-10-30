package engine

import "math"

// Computer-Häuser: entscheiden Schritt für Schritt über normale Aktionen (Apply), damit Zuschauer
// jeden Zug sehen. Jede vorgeschlagene Aktion ist vorher geprüft und ändert den Zustand, sonst
// beendet der Bot die Phase (keine Endlosschleifen).

// BotStrategy beschreibt das Verhalten eines Computer-Hauses.
type BotStrategy struct {
	Name   string
	Feuds  bool    // zettelt Händel an
	Spend  float64 // Anteil der Kasse, der investiert werden darf
	Repair int     // repariert unter diesem Zustand
}

// BotStrategies sind die wählbaren Spielweisen.
var BotStrategies = map[string]BotStrategy{
	"kaufmann": {Name: "Kaufmann", Spend: 0.6, Repair: 70},
	"kaper":    {Name: "Kaperfahrer", Spend: 0.5, Repair: 80, Feuds: true},
}

// botSkill: Spielstärke je Schwierigkeitsgrad (Anteil ausgelassener Gelegenheiten, Planung).
type botSkill struct {
	Idle    float64 // Wahrscheinlichkeit, eine Investition auszulassen
	MaxKont int     // höchstens so viele Kontore
}

var botSkills = map[string]botSkill{
	"leicht": {Idle: 0.5, MaxKont: 3},
	"mittel": {Idle: 0.25, MaxKont: 5},
	"schwer": {Idle: 0, MaxKont: 8},
}

func (g *Game) skill() botSkill {
	if s, ok := botSkills[g.Settings.Difficulty]; ok {
		return s
	}
	return botSkills["mittel"]
}

// BotAction liefert die nächste Aktion eines Computer-Hauses; ok = false: nichts zu tun.
func (g *Game) BotAction(seat int, strategy string) (Action, bool) {
	st, found := BotStrategies[strategy]
	if !found {
		st = BotStrategies["kaufmann"]
	}
	s := g.Seats[seat]
	switch s.Phase {
	case PhaseTrade:
		return g.botTrade(seat, st), true
	case PhaseEvent:
		return g.botDecide(seat, st), true
	case PhaseFeud:
		return g.botFeud(seat, st), true
	}
	return Action{}, false
}

// BotFallback beendet die Phase, falls die gewünschte Aktion abgelehnt wurde.
func (g *Game) BotFallback(seat int) Action {
	switch g.Seats[seat].Phase {
	case PhaseEvent:
		return Action{Type: "decide", Stance: "nein"}
	case PhaseFeud:
		return Action{Type: "endTurn"}
	}
	return Action{Type: "next"}
}

// once merkt sich je Jahr, welche Entscheidung schon gefallen ist (z. B. „Investition ausgelassen“).
func once(p *Player, key string) bool {
	if p.Year.BotDone == nil {
		p.Year.BotDone = map[string]bool{}
	}
	if p.Year.BotDone[key] {
		return false
	}
	p.Year.BotDone[key] = true
	return true
}

// botTrade: verkaufen, reparieren, Salz laden, Schiffe gewinnbringend verteilen, dann investieren.
// Ankünfte wachsen mit (Speicher+Schiffe)·√(Schiffe·Speicher); der Bot rechnet deshalb mit dem
// erwarteten Zuwachs je Schiff bzw. Speicher (Formel aus economy.go bei mittlerem Seegang).
func (g *Game) botTrade(seat int, st BotStrategy) Action {
	p := g.Seats[seat].Player
	y := &p.Year
	sk := g.skill()
	idle := func(key string) bool { return once(p, key) && g.Rng.Float() < sk.Idle }

	for w := 2; w <= NumGoods; w++ {
		if p.Stock[w] > 0 {
			return Action{Type: "sellAll"}
		}
	}
	if p.TotalShips() > 0 {
		budget := int(float64(p.Mark) * st.Spend)
		if p.Rigging < st.Repair && budget > y.RigCost {
			return Action{Type: "rigging", Qty: min(100-p.Rigging, budget/y.RigCost/2)}
		}
		if p.Hull < st.Repair && budget > y.HullCost {
			return Action{Type: "hull", Qty: min(100-p.Hull, budget/y.HullCost/2)}
		}
	}
	// Ohne Speicher in Lübeck verdirbt am Jahresende jede Ware: sofort einen kaufen
	if p.Stores[Luebeck] == 0 && p.Mark > y.StorePrice[Luebeck] {
		return Action{Type: "buyStores", City: Luebeck, Qty: 1}
	}
	// Salz für jede Ausfahrt (100 je Schiff), notfalls nur so viel, wie die Kasse samt Heuer hergibt
	if need := p.Ships*CargoPerShip - p.Stock[Salt]; need > 0 {
		perShip := CargoPerShip*y.SaltPrice + y.Heuer[2]
		ships := min(need/CargoPerShip+1, max(0, p.Mark)/max(1, perShip))
		if q := min(need, ships*CargoPerShip); q > 0 && once(p, "salt-"+itoa(p.Stock[Salt])) {
			return Action{Type: "buySalt", Qty: q}
		}
	}
	// Schiffe einzeln in die Stadt mit dem größten Zugewinn schicken
	if p.Ships > 0 && p.Loads() > 0 {
		best, gain := 0, 0.0
		for c := 2; c <= NumCities; c++ {
			if !p.Kontor[c] || p.Mark < y.Heuer[c] {
				continue
			}
			v := g.botValue(p, c, p.Sent[c]+1, p.Stores[c]) - g.botValue(p, c, p.Sent[c], p.Stores[c]) - float64(y.Heuer[c])
			if best == 0 || v > gain {
				best, gain = c, v
			}
		}
		if best > 0 && gain > 0 {
			return Action{Type: "sendShips", City: best, Qty: 1}
		}
	}
	// Rücklage: Heuer und Ladung für die ganze Flotte im nächsten Jahr, dazu Zinsen auf freie Anteile
	spare := p.Mark - 1000*p.Rank - p.Ships*(CargoPerShip*y.SaltPrice+y.Heuer[4]) - p.Index*p.FreeShares/3
	if spare <= 0 {
		return Action{Type: "next"}
	}
	// Investitionen nach Amortisation (Jahre bis der Preis wieder eingespielt ist)
	type inv struct {
		a       Action
		payback float64
		cost    int
	}
	var opts []inv
	for c := 2; c <= NumCities; c++ {
		if p.Kontor[c] && p.Stores[c] < MaxStores {
			n := max(p.Sent[c], 1)
			gain := g.botValue(p, c, n, p.Stores[c]+1) - g.botValue(p, c, n, p.Stores[c])
			if gain > 0 {
				opts = append(opts, inv{Action{Type: "buyStores", City: c, Qty: 1}, float64(y.StorePrice[c]) / gain, y.StorePrice[c]})
			}
		}
		if !p.Kontor[c] && !idle("kontor") {
			// Neues Kontor mit einem Speicher und zwei Schiffen
			gain := g.botValue(p, c, 2, 1) - 2*float64(y.Heuer[c])
			if gain > 0 {
				cost := y.KontorCost[c] + y.StorePrice[c]
				opts = append(opts, inv{Action{Type: "openKontor", City: c}, float64(cost) / gain, cost})
			}
		}
	}
	// Ein weiteres Schiff: Zugewinn in der besten Stadt abzüglich Heuer und Ladung
	shipGain := 0.0
	for c := 2; c <= NumCities; c++ {
		if p.Kontor[c] {
			v := g.botValue(p, c, p.Sent[c]+1, p.Stores[c]) - g.botValue(p, c, p.Sent[c], p.Stores[c]) -
				float64(y.Heuer[c]+CargoPerShip*y.SaltPrice)
			shipGain = max(shipGain, v)
		}
	}
	if shipGain > 0 && !idle("order") {
		opts = append(opts, inv{Action{Type: "orderShips", Qty: 1}, float64(y.BuildPrice) / shipGain, y.BuildPrice})
	}
	// Lübecker Lager muss die Ankünfte fassen (900 je Speicher und Ware)
	for w := 2; w <= NumGoods; w++ {
		if p.Year.Overfull[w] && p.Stores[Luebeck] < MaxStores {
			opts = append(opts, inv{Action{Type: "buyStores", City: Luebeck, Qty: 1}, 0, y.StorePrice[Luebeck]})
		}
	}
	// Kaperfahrer und bedrohte Häuser stellen Landkanonen auf
	if st.Feuds || p.Attacked > 0 {
		for c := 2; c <= NumCities; c++ {
			if p.Kontor[c] && p.Stores[c] > 0 && p.Cannons[c] < 4+p.Rank {
				opts = append(opts, inv{Action{Type: "buyCannons", City: c, Qty: 2}, 3, y.CannonCost[c] * 2})
			}
		}
	}
	best := -1
	for i, o := range opts {
		if o.cost <= spare && o.payback < 6 && (best < 0 || o.payback < opts[best].payback) {
			best = i
		}
	}
	if best >= 0 {
		return opts[best].a
	}
	return Action{Type: "next"}
}

// botValue schätzt den Erlös der Ankünfte aus Stadt c mit n Schiffen und st Speichern
// (Formel aus wearAndArrivals bei Seegang 2,5 und aktuellem Zustand, bewertet zum Tagespreis).
func (g *Game) botValue(p *Player, c, n, st int) float64 {
	S, R := 2.5, float64(p.Rank)
	cond := float64(p.Hull+p.Rigging) * 0.005
	if p.TotalShips() == 0 {
		cond = 1 // Neubauten kommen mit vollem Zustand
	}
	fn, fs, d := float64(n), float64(st), float64(Cities[c].Dist)
	w := (math.Floor(85-R) / d) * (fs + fn) * S * cond * math.Sqrt(fn*fs) * 2
	w += (d*0.2 + S*math.Sqrt(fn+5+fs)) * math.Sqrt(fn*fs) * S * 0.7 * cond
	w += d / (18 + 2*R) * S * fn
	price := float64(p.Year.Prices[Cities[c].Good])
	if price == 0 {
		price = 15
	}
	return w * price
}

// botDecide beantwortet Ereignisfragen vorsichtig, Kaperfahrer gehen auch Risiken ein.
func (g *Game) botDecide(seat int, st BotStrategy) Action {
	p := g.Seats[seat].Player
	e := p.Year.Event
	if e == nil {
		return Action{Type: "decide", Stance: "nein"}
	}
	yes := false
	switch eventTable[e.ID] {
	case "krank", "veruntreuung", "audienz", "heirat":
		yes = true
	case "bergen":
		yes = p.Mark > (p.Year.Luck+p.Index*20)*2
	case "fest":
		yes = p.Mark > 400*p.Rank*8
	case "witwen", "dom":
		yes = p.Mark > -e.Delta*8
	case "schmuggel":
		yes = st.Feuds
	case "kind":
		return Action{Type: "decide", Stance: "ja", Name: BotNames[(seat+p.Generation+1)%len(BotNames)].Name}
	}
	stance := "nein"
	if yes {
		stance = "ja"
	}
	return Action{Type: "decide", Stance: stance}
}

// botFeud: Kaperfahrer greifen ein schwach verteidigtes Kontor an, wo eigene Schiffe liegen.
func (g *Game) botFeud(seat int, st BotStrategy) Action {
	p := g.Seats[seat].Player
	if !st.Feuds || !g.canFeud(seat) {
		return Action{Type: "endTurn"}
	}
	for c := 2; c <= NumCities; c++ {
		if p.Sent[c] < 2 {
			continue
		}
		for t, s := range g.Seats {
			o := s.Player
			if t == seat || o.Out || !o.Kontor[c] {
				continue
			}
			if o.Cannons[c]+o.Sent[c] < p.Sent[c] {
				return Action{Type: "attack", Target: t, City: c, Mode: "kontor", Stance: "speicher"}
			}
		}
	}
	return Action{Type: "endTurn"}
}

// Namen für Computer-Häuser (Lübecker Kaufmannsfamilien nachempfunden, frei gewählt).
var BotNames = []struct {
	Name   string
	Female bool
}{
	{"Hinrich", false}, {"Gertrud", true}, {"Johann", false}, {"Mechthild", true},
	{"Bertram", false}, {"Margarethe", true}, {"Godeke", false}, {"Alheyd", true},
	{"Tidemann", false}, {"Wobbeke", true}, {"Hermen", false}, {"Gesche", true},
}

// AddBot gründet ein Computer-Haus.
func (g *Game) AddBot(strategy string) (int, error) {
	if _, ok := BotStrategies[strategy]; !ok {
		return 0, &RuleError{"bad_strategy", "Unbekannte Spielweise."}
	}
	n := BotNames[len(g.Seats)%len(BotNames)]
	seat, err := g.AddPlayer(n.Name, n.Female)
	if err != nil {
		return 0, err
	}
	g.SetBot(seat, strategy)
	if g.CPU == nil {
		g.CPU = map[int]bool{}
	}
	g.CPU[seat] = true
	return seat, nil
}

// SetBot schaltet den Autopiloten eines Sitzes ein (strategy) oder aus (leer).
func (g *Game) SetBot(seat int, strategy string) {
	if g.Bots == nil {
		g.Bots = map[int]string{}
	}
	if strategy == "" {
		delete(g.Bots, seat)
		return
	}
	g.Bots[seat] = strategy
}
