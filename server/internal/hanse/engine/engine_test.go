package engine

import "testing"

func newTestGame(t *testing.T, players int) *Game {
	t.Helper()
	g := NewGame(DefaultSettings(DefaultBalance()), DefaultBalance(), 42)
	for i := 0; i < players; i++ {
		if _, err := g.AddPlayer("Haus"+itoa(i), i%2 == 1); err != nil {
			t.Fatal(err)
		}
	}
	return g
}

func mustApply(t *testing.T, g *Game, seat int, a Action) []Event {
	t.Helper()
	ev, err := g.Apply(seat, a, 0)
	if err != nil {
		t.Fatalf("%+v: %v", a, err)
	}
	return ev
}

// TestStart: Startausstattung des Originals und erste Ankunft aus Ystadt.
func TestStart(t *testing.T) {
	g := newTestGame(t, 2)
	g.Start(0)
	p := g.Seats[0].Player
	if p.Mark != 10000 || p.Ships != 1 || !p.Kontor[Luebeck] || !p.Kontor[6] || p.Shares[0] != OwnShares {
		t.Fatalf("Startausstattung falsch: %+v", p)
	}
	if p.Year.Arrived[6] <= 0 || p.Stock[Wool] != p.Year.Arrived[6] {
		t.Errorf("aus Ystadt sollte Wolle kommen: %v", p.Year.Arrived)
	}
	if g.Seats[0].Phase != PhaseTrade || p.Index <= 0 {
		t.Errorf("Phase %s, Index %d", g.Seats[0].Phase, p.Index)
	}
}

// TestTrade: Verkaufen, Salz, Aussenden, Kontor – und die Grenzen des Originals.
func TestTrade(t *testing.T) {
	g := newTestGame(t, 1)
	g.Start(0)
	p := g.Seats[0].Player
	wool := p.Stock[Wool]
	mark := p.Mark
	mustApply(t, g, 0, Action{Type: "sellAll"})
	if p.Stock[Wool] != 0 || p.Mark != mark+wool*p.Year.Prices[Wool] {
		t.Error("Verkauf falsch verbucht")
	}
	if _, err := g.Apply(0, Action{Type: "sendShips", City: 2, Qty: 1}, 0); err != ErrNoKontor {
		t.Errorf("ohne Kontor aussenden: %v", err)
	}
	mustApply(t, g, 0, Action{Type: "sendShips", City: 6, Qty: 1})
	if p.Ships != 0 || p.Underway != 1 || p.Sent[6] != 1 || p.Stock[Salt] != 100 {
		t.Errorf("Aussenden falsch: %+v", p)
	}
	if _, err := g.Apply(0, Action{Type: "buyStores", City: 6, Qty: 5}, 0); err != ErrNotPossible {
		t.Error("mehr als 5 Speicher je Stadt erlaubt")
	}
	if _, err := g.Apply(0, Action{Type: "buyShares", Target: 0, Qty: 1}, 0); err != ErrExchangeClosed {
		t.Errorf("Börse im ersten Jahr: %v", err)
	}
}

// TestYearCycle: Ein Jahr durchspielen, Ereignis beantworten, Rückkehr der Schiffe.
func TestYearCycle(t *testing.T) {
	g := newTestGame(t, 1)
	g.Start(0)
	p := g.Seats[0].Player
	mustApply(t, g, 0, Action{Type: "sendShips", City: 6, Qty: 1})
	mustApply(t, g, 0, Action{Type: "next"})
	if g.Seats[0].Phase == PhaseEvent {
		mustApply(t, g, 0, Action{Type: "decide", Stance: "nein", Name: "Klaus"})
	}
	if g.Seats[0].Phase != PhaseFeud {
		t.Fatalf("nach Ereignis Phase %s", g.Seats[0].Phase)
	}
	mustApply(t, g, 0, Action{Type: "endTurn"})
	if g.Year != StartYear+1 || g.Seats[0].Phase != PhaseTrade {
		t.Fatalf("neues Jahr nicht begonnen: %d %s", g.Year, g.Seats[0].Phase)
	}
	if p.Ships+p.Underway < 1 || p.LifeLeft != StartLife-1 {
		t.Errorf("Schiff/Leben: %d %d %d", p.Ships, p.Underway, p.LifeLeft)
	}
}

// TestPromotionAndDeath: Beförderung über die Indexschwelle, Tod mit Erbfolge.
func TestPromotionAndDeath(t *testing.T) {
	g := newTestGame(t, 1)
	g.Start(0)
	p := g.Seats[0].Player
	p.Index, p.Year.Luck = 200, 1
	g.endYear(0)
	if p.Rank != 2 || p.LifeLeft != StartLife+10-1-1 {
		t.Errorf("Index 200 > 101 sollte befördern: Rang %d, Leben %d", p.Rank, p.LifeLeft)
	}
	p.Index, p.LifeLeft = 0, 1
	p.Child, p.ChildFem = "Klara", true
	g.endYear(0)
	if p.Name != "Klara" || !p.Female || p.Generation != 1 || p.LifeLeft != StartLife {
		t.Errorf("Erbfolge falsch: %+v", p)
	}
}

// TestBattle: Händel gegen ein Kontor verbucht Verluste beider Seiten.
func TestBattle(t *testing.T) {
	g := newTestGame(t, 2)
	g.Year = StartYear + 1
	g.Start(0)
	a, d := g.Seats[0].Player, g.Seats[1].Player
	a.Sent[6], a.Underway = 10, 10
	d.Sent[6], d.Underway = 2, 2
	d.Cannons[6] = 3
	g.Seats[0].Phase = PhaseFeud
	ev := mustApply(t, g, 0, Action{Type: "attack", Target: 1, City: 6, Mode: "kontor"})
	b := ev[0].Data.(BattleResult)
	if b.AttShips != 10 || b.DefShips != 2 || b.Cannons != 3 {
		t.Fatalf("Aufstellung falsch: %+v", b)
	}
	if a.Sent[6] != 10-b.AttLost || d.Cannons[6] != 3-b.CanLost || !a.Year.Feud {
		t.Errorf("Verluste nicht verbucht: %+v", b)
	}
	if _, err := g.Apply(0, Action{Type: "attack", Target: 1, City: 6}, 0); err != ErrNoFeud {
		t.Error("zweiter Händel im selben Jahr erlaubt")
	}
}

// TestBotsPlay: Vier Computer-Häuser spielen 80 Jahre; Bestände bleiben gültig, jemand steigt auf.
func TestBotsPlay(t *testing.T) {
	g := NewGame(DefaultSettings(DefaultBalance()), DefaultBalance(), 7)
	for _, s := range []string{"kaufmann", "kaper", "kaufmann", "kaper"} {
		if _, err := g.AddBot(s); err != nil {
			t.Fatal(err)
		}
	}
	g.Start(0)
	maxRank := 0
	for step := 0; step < 200000 && !g.Over && g.Year < StartYear+80; step++ {
		for i := range g.Seats {
			a, ok := g.BotAction(i, g.Bots[i])
			if !ok {
				continue
			}
			if _, err := g.Apply(i, a, 0); err != nil {
				if _, err := g.Apply(i, g.BotFallback(i), 0); err != nil {
					t.Fatalf("Fallback abgelehnt: %v", err)
				}
			}
		}
		for _, s := range g.Seats {
			p := s.Player
			maxRank = max(maxRank, p.Rank)
			if p.Ships < 0 || p.Underway < 0 || p.Stock[Salt] < 0 {
				t.Fatalf("negativer Bestand: %+v", p)
			}
			for c := 1; c <= NumCities; c++ {
				if p.Stores[c] < 0 || p.Stores[c] > MaxStores || p.Cannons[c] < 0 || p.Sent[c] < 0 {
					t.Fatalf("Stadt %d ungültig: %+v", c, p)
				}
			}
		}
	}
	for i, s := range g.Seats {
		p := s.Player
		t.Logf("%d %s Rang %d Index %d Mark %d Schiffe %d Gen %d", i, p.Name, p.Rank, p.Index, p.Mark, p.TotalShips(), p.Generation)
	}
	t.Logf("Jahr %d, über: %v, Sieger %d", g.Year, g.Over, g.Winner)
	if maxRank < 3 {
		t.Errorf("in 80 Jahren kaum Aufstieg (höchster Rang %d)", maxRank)
	}
}

// TestSnapshot: Spielstand übersteht Speichern und Laden.
func TestSnapshot(t *testing.T) {
	g := newTestGame(t, 2)
	g.Start(0)
	raw, err := g.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	h, err := RestoreGame(raw)
	if err != nil {
		t.Fatal(err)
	}
	if h.Seats[1].Player.Mark != g.Seats[1].Player.Mark || h.Rng.Float() != g.Rng.Float() {
		t.Error("Snapshot unvollständig")
	}
}
