package engine

import "math"

// Wirtschaftsformeln, nachgebaut nach dem Spielverhalten des Vorbilds.
// J = Spieljahr ab 1, E = Epoche INT(J/3 + 1,5), G = Glück 1–10, R = Rang, S = Seegang 1–4.

// gameYear liefert J (1 im ersten Jahr).
func (g *Game) gameYear() int { return g.Year - StartYear + 1 }

// epoch liefert E.
func (g *Game) epoch() int { return int(float64(g.gameYear())/3 + 1.5) }

// rollLuck zieht den Glückswert 1–10. Hier ist er ein fairer Würfel je Spieler
// und Jahr.
func (g *Game) rollLuck() int { return 1 + g.Rng.IntN(10) }

// cityPrices setzt Kontor-, Speicher- und Kanonenpreise sowie die Heuer je Stadt.
func (g *Game) cityPrices(p *Player) {
	y := &p.Year
	G, E, J, R := y.Luck, g.epoch(), g.gameYear(), p.Rank
	for c := 1; c <= NumCities; c++ {
		d := Cities[c].Dist
		y.KontorCost[c] = max(500, G*E*10+20*d+2000-J*10)
		y.StorePrice[c] = max(200, G*20+d*20+E*250+1000-J*10)
		y.CannonCost[c] = max(20, G*10+d+E*10-J*3)
		y.Heuer[c] = d*2 + 40*R
	}
}

// shipPrices setzt Bauauftrag, Schiffswert und Reparaturpreise.
func (g *Game) shipPrices(p *Player) {
	y := &p.Year
	y.BuildPrice = y.Luck*100 + 1000*p.Rank + 200*g.epoch()
	cond := float64(p.Hull+p.Rigging) * 0.005
	y.ShipValue = int(float64(y.BuildPrice) * cond)
	total := p.TotalShips()
	for c := 2; c <= NumCities; c++ {
		total += p.Sent[c]
	}
	total = max(1, total)
	y.RigCost = max(1, int(float64(y.BuildPrice)*0.002*float64(total)))
	y.HullCost = max(1, int(float64(y.BuildPrice)/600*float64(total)))
}

// wearAndArrivals: Verschleiß der ausgesandten Schiffe, Lieferung der Neubauten und Ankunft der
// Waren aus den Kontoren (Schritte 5–8). Danach sind die Aussendungen verbraucht.
func (g *Game) wearAndArrivals(p *Player) {
	y := &p.Year
	S := float64(y.Sea)
	hull, rig := float64(p.Hull), float64(p.Rigging)
	for c := 2; c <= NumCities; c++ {
		n := float64(p.Sent[c])
		if n == 0 {
			continue
		}
		d := float64(Cities[c].Dist)
		hull -= n * d * (5 - S) / (n*100 + 1)
		rig -= n * d * (6 - S) / (n*100 + 1)
	}
	p.Hull, p.Rigging = int(hull), int(rig)
	// Unter 0 % geht je Bauteil ein Schiff verloren, der Rest steht bei 10 %
	if p.Hull < 0 {
		p.Hull = 10
		p.loseShip()
		y.Lost++
	}
	if p.Rigging < 0 {
		p.Rigging = 10
		p.loseShip()
		y.Lost++
	}
	// Neubauten: Zustand 100 %, mit dem Bestand gewichtet gemittelt
	if p.Orders > 0 {
		have := p.Ships
		if have+p.Orders > 0 {
			p.Hull = (p.Hull*have + 100*p.Orders) / (have + p.Orders)
			p.Rigging = (p.Rigging*have + 100*p.Orders) / (have + p.Orders)
		}
		p.Ships += p.Orders
		y.Delivered = p.Orders
		p.Orders = 0
	}
	if p.Ships <= 0 && p.Underway <= 0 {
		p.Ships, p.Hull, p.Rigging = 0, 0, 0
	}
	g.shipPrices(p)
	// Ankunft der Waren je Stadt (Formel des Originals)
	cond := float64(p.Hull+p.Rigging) * 0.005
	R := float64(p.Rank)
	for c := 2; c <= NumCities; c++ {
		n, st := float64(p.Sent[c]), float64(p.Stores[c])
		d := float64(Cities[c].Dist)
		w := (math.Floor(85-R) / d) * (st + n) * S * cond * math.Sqrt(n*st) * 2
		w += (d*0.2 + S*math.Sqrt(n+5+st)) * math.Sqrt(n*st) * S * 0.7 * cond
		w += d / (18 + 2*R) * S * n
		y.Arrived[c] = int(w)
		p.Stock[Cities[c].Good] += y.Arrived[c]
	}
	p.Sent = [NumCities + 1]int{}
}

// loseShip nimmt ein Schiff aus dem Hafen (oder von den Ausgesandten, falls der Hafen leer ist).
func (p *Player) loseShip() {
	if p.Ships > 0 {
		p.Ships--
		return
	}
	for c := 2; c <= NumCities; c++ {
		if p.Sent[c] > 0 {
			p.Sent[c]--
			return
		}
	}
}

// marketPrices würfelt die Verkaufspreise der Waren 2–6 und den Salzpreis (Schritt 9).
func (g *Game) marketPrices(p *Player) {
	y := &p.Year
	E := float64(g.epoch())
	for w := 2; w <= NumGoods; w++ {
		fw := float64(w)
		y.Prices[w] = max(1, int(g.Rng.Float()*(15-fw*1.5)+24-fw*2+E/fw))
	}
	y.SaltPrice = max(1, int(g.Rng.Float()*(float64(p.Rank)+3/float64(y.Sea))+1+E*0.4))
}

// wealth berechnet das Vermögen für den Index (Schritt 10).
func (g *Game) wealth(seat int) int {
	p := g.Seats[seat].Player
	y := &p.Year
	v := p.Mark + (p.Ships+p.Underway)*y.ShipValue
	for c := 1; c <= NumCities; c++ {
		v += p.Stores[c]*y.StorePrice[c] - p.Stores[c]*p.Rank*100
	}
	for i, n := range p.Shares {
		if i < len(g.Seats) && i != seat {
			v += n * g.Seats[i].Player.Index * 3
		}
	}
	for w := 2; w <= NumGoods; w++ {
		v += p.Stock[w] * y.Prices[w]
	}
	return v
}

// updateIndex berechnet den neuen Index, gedämpft durch den alten (Index-Formel des Originals).
func (g *Game) updateIndex(seat int) {
	p := g.Seats[seat].Player
	v := g.wealth(seat) + p.LifeLeft*500
	idx := v / (500 + p.Index)
	p.Index = max(0, min(3175, idx))
	p.IndexHist = append(p.IndexHist, p.Index)
	if len(p.IndexHist) > 9 {
		p.IndexHist = p.IndexHist[len(p.IndexHist)-9:]
	}
}

// capStock schneidet Waren über der Speicherkapazität in Lübeck ab („überfüllt“).
func capStock(p *Player) {
	limit := StoreCapacity * p.Stores[Luebeck]
	for w := 1; w <= NumGoods; w++ {
		p.Year.Overfull[w] = p.Stock[w] > limit
		if p.Stock[w] > limit {
			p.Stock[w] = limit
		}
	}
}
