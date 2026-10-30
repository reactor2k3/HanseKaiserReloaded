package engine

// Player ist ein Handelshaus.
type Player struct {
	Name       string `json:"name"`
	Female     bool   `json:"female"`
	Mark       int    `json:"mark"`       // Barvermögen
	Index      int    `json:"index"`      // Ansehen/Wertungsmaß
	Rank       int    `json:"rank"`       // 1–11, 12 = Senat (Sieg)
	LifeLeft   int    `json:"lifeLeft"`   // verbleibende Lebensjahre
	Generation int    `json:"generation"` // 0 = Gründer; über MaxGenerations ausgeschieden
	Out        bool   `json:"out"`        // Haus ist ausgestorben

	Ships    int `json:"ships"`    // Schiffe im Lübecker Hafen
	Underway int `json:"underway"` // ausgesandt, kehren zu Jahresbeginn zurück
	Hull     int `json:"hull"`     // Rumpf in % (gemeinsam für die Flotte)
	Rigging  int `json:"rigging"`  // Takelage in %
	Orders   int `json:"orders"`   // Bauaufträge, Lieferung zu Jahresbeginn

	Kontor  [NumCities + 1]bool `json:"kontor"`
	Stores  [NumCities + 1]int  `json:"stores"`  // Speicher je Stadt
	Cannons [NumCities + 1]int  `json:"cannons"` // Landkanonen je Stadt
	Sent    [NumCities + 1]int  `json:"sent"`    // in diesem Jahr ausgesandte Schiffe
	Stock   [NumGoods + 1]int   `json:"stock"`   // Lager in Lübeck

	Shares     []int `json:"shares"`     // gehaltene Anteile je Handelshaus (Sitz)
	FreeShares int   `json:"freeShares"` // eigene Anteile, die noch am Markt sind

	Prison   int    `json:"prison"` // verbleibende Schuldturm-Jahre
	Married  bool   `json:"married"`
	Child    string `json:"child"` // Name des Kindes (Erbe), leer = keins
	ChildFem bool   `json:"childFem"`

	IndexHist []int     `json:"indexHist"` // Index der letzten Jahre (Chronik-Grafik)
	History   []Chronik `json:"history"`   // Lebenslauf des Hauses
	Prisoned  int       `json:"prisoned"`  // Zähler „mal im Schuldturm“
	Feuds     int       `json:"feuds"`     // „mal Händel angezettelt“
	Attacked  int       `json:"attacked"`  // „mal angegriffen worden“

	Motto string   `json:"motto"`
	Arms  ArmsSpec `json:"arms"`

	Year YearState `json:"year"`
}

// Chronik ist ein Eintrag im Lebenslauf eines Hauses.
type Chronik struct {
	Year int    `json:"year"`
	Text string `json:"text"`
}

// YearState sind die Werte des laufenden Jahres (Preise, Ankünfte, Glück).
type YearState struct {
	Luck       int                   `json:"-"`         // Glückswert 1–10, je Jahr neu gewürfelt
	Sea        int                   `json:"sea"`       // Seegang 1–4
	Prices     [NumGoods + 1]int     `json:"prices"`    // Verkaufspreise 2–6
	SaltPrice  int                   `json:"saltPrice"` // Einkaufspreis Salz
	Arrived    [NumCities + 1]int    `json:"arrived"`   // eingetroffene Waren je Stadt
	KontorCost [NumCities + 1]int    `json:"kontorCost"`
	StorePrice [NumCities + 1]int    `json:"storePrice"`
	CannonCost [NumCities + 1]int    `json:"cannonCost"`
	Heuer      [NumCities + 1]int    `json:"heuer"` // Heuer je Schiff in die Stadt
	BuildPrice int                   `json:"buildPrice"`
	ShipValue  int                   `json:"shipValue"`
	RigCost    int                   `json:"rigCost"`  // 1 % Takelage
	HullCost   int                   `json:"hullCost"` // 1 % Rumpf
	Lost       int                   `json:"lost"`     // durch Verschleiß verlorene Schiffe
	Delivered  int                   `json:"delivered"`
	Event      *EventState           `json:"event,omitempty"`
	Feud       bool                  `json:"feud"` // Händel in diesem Jahr schon angezettelt
	Dividend   int                   `json:"dividend"`
	Interest   int                   `json:"interest"`
	Skipped    bool                  `json:"skipped"` // Jahr im Schuldturm
	Joined     bool                  `json:"joined"`  // mitten im Jahr beigetreten, steigt nächstes Jahr ein
	BotDone    map[string]bool       `json:"-"`
	Overfull   [NumGoods + 1]bool    `json:"overfull"`
	Wear       [NumCities + 1][2]int `json:"-"`
}

// newPlayer legt ein Handelshaus mit der Startausstattung des Originals an.
func newPlayer(name string, female bool) *Player {
	p := &Player{
		Name: name, Female: female, Mark: 10000, Rank: 1, LifeLeft: StartLife,
		Ships: 1, Hull: 100, Rigging: 100, FreeShares: TotalShares - OwnShares,
		Arms: ArmsSpec{Kind: "house"},
	}
	p.Kontor[Luebeck], p.Kontor[6] = true, true
	p.Stores[Luebeck], p.Stores[6] = 1, 1
	p.Stock[Salt] = 200
	// Wie im Original liefert zu Beginn ein Schiff aus Ystadt, ohne im Hafen zu fehlen
	p.Sent[6] = 1
	return p
}

// Alive meldet ein Haus, das noch mitspielt.
func (p *Player) Alive() bool { return !p.Out }

// Loads ist die Zahl der Schiffsladungen Salz (je 100).
func (p *Player) Loads() int { return p.Stock[Salt] / CargoPerShip }

// TotalShips zählt alle Schiffe (im Hafen und unterwegs).
func (p *Player) TotalShips() int { return p.Ships + p.Underway }

func (p *Player) chronik(year int, text string) {
	p.History = append(p.History, Chronik{year, text})
}
