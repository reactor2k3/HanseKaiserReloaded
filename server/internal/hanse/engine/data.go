package engine

// Stammdaten des Spiels: Waren, Städte, Ränge und Grenzen.

// StartYear ist das erste Spieljahr; nach MaxYears Jahren endet jede Partie („große Seuche“).
const (
	StartYear = 1356
	MaxYears  = 255
)

// Waren: Index 1–6 wie im Original; 0 bleibt frei.
const (
	Salt = 1 + iota
	Furs
	Linen
	Train
	Honey
	Wool
	NumGoods = 6
)

// GoodNames sind die Anzeigenamen der Waren.
var GoodNames = [NumGoods + 1]string{"", "Salz", "Pelze", "Leinen", "Tran", "Honig", "Wolle"}

// City ist ein Hafen der Hanse: erzeugte Ware und Entfernung von Lübeck (Seemeilen-Maß des Originals).
type City struct {
	Name string `json:"name"`
	Good int    `json:"good"`
	Dist int    `json:"dist"`
}

// Lübeck ist Stadt 1 (Heimathafen); die Indizes 1–9 entsprechen dem Original.
const (
	Luebeck   = 1
	NumCities = 9
)

// Cities: die neun Hansestädte.
var Cities = [NumCities + 1]City{
	{},
	{"Lübeck", Salt, 1},
	{"Bergen", Furs, 170},
	{"Tönsberg", Train, 120},
	{"Warberg", Honey, 80},
	{"Malmö", Wool, 60},
	{"Ystadt", Wool, 50},
	{"Visby", Honey, 100},
	{"Riga", Linen, 140},
	{"Nowgorod", Furs, 200},
}

// MaxRank: Rang 12 bedeutet die Berufung in den Senat und damit den Sieg.
const MaxRank = 12

// Ränge für Männer und Frauen; Index 1–11.
var RankTitles = [2][MaxRank]string{
	{"", "Bürger", "Händler", "Kaufmann", "Großkaufmann", "Patrizier", "Bruder", "Junker", "Diplomat", "Schaffer", "Ratsherr", "Senator"},
	{"", "Bürgerin", "Händlerin", "Kauffrau", "Großkauffrau", "Patrizierin", "Schwester", "Edelfrau", "Diplomatin", "Schafferin", "Ratsfrau", "Senatorin"},
}

// Aufnahme-Urkunden: beim Erreichen von Rang r wird man in Guilds[r-1] aufgenommen.
var Guilds = [MaxRank]struct{ Body, Verb string }{
	{},
	{"in die Handelsrolle", "eingetragen"},
	{"in die Kaufleutekompanie", "aufgenommen"},
	{"in die Kaufmannsgilde", "aufgenommen"},
	{"in die Große Gilde", "eingelassen"},
	{"in den Klub der Reichen", "eingeladen"},
	{"in die Georgenbruderschaft", "eingeführt"},
	{"in die Junkergesellschaft", "aufgenommen"},
	{"in die Gewandschneidergilde", "eingeladen"},
	{"in die Zirkelgesellschaft", "eingeführt"},
	{"in den Rat", "berufen"},
	{"in den Senat", "berufen"},
}

// SeaNames: Seegang des Jahres, 1 = tobende … 4 = ruhige See.
var SeaNames = [5]string{"", "tobende", "stürmische", "bewegte", "ruhige"}

// Grenzen aus dem Original.
const (
	MaxStores      = 5   // Speicher je Stadt
	MaxCannons     = 20  // Landkanonen je Stadt
	StoreCapacity  = 900 // Lagerplatz je Speicher in Lübeck (Einheiten je Ware)
	CargoPerShip   = 100 // Salz je ausgesandtem Schiff
	StartLife      = 12  // Lebensjahre eines neuen Herrschers
	MaxGenerations = 3   // nach dem Tod der vierten Generation scheidet ein Haus aus (wie im Original)
	SeucheDeaths   = 2   // haben alle Häuser zweimal ihr Oberhaupt verloren, endet das Spiel („große Seuche“)
	TotalShares    = 100 // Anteile je Handelshaus
	OwnShares      = 90  // davon hält der Gründer
	DanishStart    = 6   // dänische Flotte zu Beginn
	DanishMax      = 16
	MaxFleet       = 20 // Schiffe je Seite im Gefecht
	PrisonCap      = 5000
)

// Title liefert die Anrede eines Ranges.
func Title(rank int, female bool) string {
	r := max(1, min(MaxRank-1, rank))
	if female {
		return RankTitles[1][r]
	}
	return RankTitles[0][r]
}
