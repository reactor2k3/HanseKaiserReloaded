package engine

// Hauptmenü des Originals: Handel, Börse, Schiffe, Kontore.
// Alle Prüfungen folgen den „NICHT MÖGLICH!“-Bedingungen des Originals.

// trade führt eine Aktion der Handelsphase aus.
func (g *Game) trade(seat int, a Action) ([]Event, error) {
	p := g.Seats[seat].Player
	y := &p.Year
	n := a.Qty
	switch a.Type {
	case "sellGood":
		// Ware 2–6 zum Tagespreis verkaufen; Qty 0 = alles
		if a.Good < 2 || a.Good > NumGoods {
			return nil, ErrBadGood
		}
		if n == 0 {
			n = p.Stock[a.Good]
		}
		if n > p.Stock[a.Good] {
			return nil, ErrNotPossible
		}
		p.Stock[a.Good] -= n
		p.Mark += n * y.Prices[a.Good]
	case "sellAll":
		for w := 2; w <= NumGoods; w++ {
			p.Mark += p.Stock[w] * y.Prices[w]
			p.Stock[w] = 0
		}
	case "buySalt":
		if n == 0 || p.Mark < n*y.SaltPrice {
			return nil, ErrNotPossible
		}
		p.Mark -= n * y.SaltPrice
		p.Stock[Salt] += n
	case "buyShares", "sellShares":
		return nil, g.exchange(seat, a)
	case "rigging":
		// Takelage trimmen, je Prozentpunkt RigCost, höchstens 100 %
		if n == 0 || p.Rigging+n > 100 || p.Mark < n*y.RigCost {
			return nil, ErrNotPossible
		}
		p.Mark -= n * y.RigCost
		p.Rigging += n
		g.shipPrices(p)
	case "hull":
		if n == 0 || p.Hull+n > 100 || p.Mark < n*y.HullCost {
			return nil, ErrNotPossible
		}
		p.Mark -= n * y.HullCost
		p.Hull += n
		g.shipPrices(p)
	case "orderShips":
		if n == 0 || p.Mark < n*y.BuildPrice {
			return nil, ErrNotPossible
		}
		p.Mark -= n * y.BuildPrice
		p.Orders += n
	case "sellShips":
		if n == 0 || n > p.Ships {
			return nil, ErrNotPossible
		}
		p.Mark += n * y.ShipValue
		p.Ships -= n
		g.shipPrices(p)
	case "sendShips":
		// Nur in Städte mit Kontor (nicht Lübeck); je Schiff Heuer und eine Ladung Salz
		c := a.City
		if c < 2 || c > NumCities || !p.Kontor[c] {
			return nil, ErrNoKontor
		}
		if n == 0 || n > p.Loads() || n > p.Ships || p.Mark < n*y.Heuer[c] {
			return nil, ErrNotPossible
		}
		p.Mark -= n * y.Heuer[c]
		p.Ships -= n
		p.Stock[Salt] -= n * CargoPerShip
		p.Sent[c] += n
		p.Underway += n
	case "buyStores":
		c := a.City
		if !validCity(c) || !p.Kontor[c] {
			return nil, ErrNoKontor
		}
		if n == 0 || p.Stores[c]+n > MaxStores || p.Mark < n*y.StorePrice[c] {
			return nil, ErrNotPossible
		}
		p.Mark -= n * y.StorePrice[c]
		p.Stores[c] += n
	case "sellStores":
		c := a.City
		if !validCity(c) || n == 0 || n > p.Stores[c] {
			return nil, ErrNotPossible
		}
		p.Mark += int(float64(n*y.StorePrice[c]) * 0.75)
		p.Stores[c] -= n
	case "buyCannons":
		c := a.City
		if c < 2 || c > NumCities || !p.Kontor[c] {
			return nil, ErrNoKontor
		}
		if n == 0 || p.Cannons[c]+n > MaxCannons || p.Mark < n*y.CannonCost[c] {
			return nil, ErrNotPossible
		}
		p.Mark -= n * y.CannonCost[c]
		p.Cannons[c] += n
	case "openKontor":
		c := a.City
		if c < 2 || c > NumCities || p.Kontor[c] || p.Mark < y.KontorCost[c] {
			return nil, ErrNotPossible
		}
		p.Mark -= y.KontorCost[c]
		p.Kontor[c] = true
		p.chronik(g.Year, "Kontor eröffnet in "+Cities[c].Name)
		return []Event{{"kontor", seat, g.Year, map[string]any{"city": c}}}, nil
	default:
		return nil, ErrWrongPhase
	}
	return nil, nil
}

func validCity(c int) bool { return c >= 1 && c <= NumCities }

// exchange: Anteile anderer Häuser kaufen oder verkaufen, Preis = Index·10 je Anteil.
// Wie im Original fließt der Kaufpreis an niemanden (der Markt nimmt ihn auf).
func (g *Game) exchange(seat int, a Action) error {
	if g.gameYear() < 2 {
		return ErrExchangeClosed
	}
	t := a.Target
	if t < 0 || t >= len(g.Seats) || t == seat || g.Seats[t].Player.Out {
		return ErrNotPossible
	}
	p, o := g.Seats[seat].Player, g.Seats[t].Player
	n := a.Qty
	price := n * o.Index * 10
	if a.Type == "buyShares" {
		if n == 0 || n > o.FreeShares || p.Mark < price {
			return ErrNotPossible
		}
		p.Mark -= price
		o.FreeShares -= n
		p.Shares[t] += n
		return nil
	}
	if n == 0 || n > p.Shares[t] {
		return ErrNotPossible
	}
	p.Mark += price
	p.Shares[t] -= n
	o.FreeShares += n
	return nil
}
