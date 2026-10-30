// Command hkd ist der gemeinsame Spielserver von HanseKaiserReloaded: Kaiser unter /kaiser/,
// Hanse unter /hanse/, die gemeinsame Startseite (Spielauswahl, Spielelisten, Ruhmeshallen) unter /.
// Jedes Spiel behält seinen eigenen Hub und seine eigene Datenbank; gemeinsam sind der Port, die
// Grenze für laufende Spiele (MaxGames gilt für beide zusammen) und das Begrenzen je IP.
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	hengine "hansekaiserreloaded/server/internal/hanse/engine"
	hserver "hansekaiserreloaded/server/internal/hanse/server"
	kengine "hansekaiserreloaded/server/internal/kaiser/engine"
	kserver "hansekaiserreloaded/server/internal/kaiser/server"
	"hansekaiserreloaded/server/internal/media"
	"hansekaiserreloaded/server/internal/store"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "Adresse des HTTP-Servers")
	data := flag.String("data", "../data", "Ordner mit kaiser/ (balance.json, world.json) und hanse/ (balance.json)")
	static := flag.String("static", "../dist", "Ordner mit den gebauten Clients lobby/, kaiser/ und hanse/ (leer = keiner)")
	state := flag.String("state", ".", "Ordner für Spielstände (kaiser.db, hanse.db) und hochgeladene Bilder (media/)")
	origins := flag.String("origins", "localhost:5173,127.0.0.1:5173", "erlaubte fremde Origins (Komma-Liste, Dev-Server)")
	// Grenzen gegen Überlastung und Missbrauch; MaxGames zählt Kaiser und Hanse zusammen
	klim, hlim := kserver.DefaultLimits(), hserver.DefaultLimits()
	flag.IntVar(&klim.MaxGames, "max-games", klim.MaxGames, "höchstens so viele gleichzeitig laufende Spiele (beide Spielarten zusammen)")
	flag.IntVar(&kengine.MaxSeats, "max-players", kengine.MaxSeats, "höchstens so viele Spieler je Spiel")
	flag.IntVar(&klim.MaxConnsPerIP, "max-conns-ip", klim.MaxConnsPerIP, "gleichzeitige Verbindungen je IP und Spielart")
	flag.IntVar(&klim.MaxConns, "max-conns", klim.MaxConns, "gleichzeitige Verbindungen insgesamt je Spielart")
	flag.BoolVar(&klim.TrustProxy, "trust-proxy", false, "Client-IP aus X-Forwarded-For (nur hinter eigenem Reverse-Proxy)")
	flag.Parse()
	hengine.MaxSeats = kengine.MaxSeats
	hlim.MaxGames, hlim.MaxConnsPerIP, hlim.MaxConns, hlim.TrustProxy = klim.MaxGames, klim.MaxConnsPerIP, klim.MaxConns, klim.TrustProxy

	fail := func(what string, err error) {
		slog.Error(what, "err", err)
		os.Exit(1)
	}
	allowed := splitList(*origins)
	mstore, err := media.NewStore(filepath.Join(*state, "media"))
	if err != nil {
		fail("media", err)
	}

	// Kaiser
	kb, err := kengine.LoadBalance(filepath.Join(*data, "kaiser", "balance.json"))
	if err != nil {
		fail("kaiser balance", err)
	}
	world, err := kengine.LoadWorld(filepath.Join(*data, "kaiser", "world.json"))
	if err != nil {
		fail("kaiser world", err)
	}
	kst, err := store.Open(filepath.Join(*state, "kaiser.db"))
	if err != nil {
		fail("kaiser datenbank", err)
	}
	defer kst.Close()
	khub, err := kserver.NewHub(kb, world, allowed, kst)
	if err != nil {
		fail("kaiser spiele laden", err)
	}

	// Hanse
	hb, err := hengine.LoadBalance(filepath.Join(*data, "hanse", "balance.json"))
	if err != nil {
		fail("hanse balance", err)
	}
	hst, err := store.Open(filepath.Join(*state, "hanse.db"))
	if err != nil {
		fail("hanse datenbank", err)
	}
	defer hst.Close()
	hhub, err := hserver.NewHub(hb, allowed, hst)
	if err != nil {
		fail("hanse spiele laden", err)
	}

	// Gemeinsame Spielegrenze: jeder Hub zählt die laufenden Spiele des anderen mit
	klim.Others, hlim.Others = hhub.ActiveGames, khub.ActiveGames
	khub.SetLimits(klim)
	hhub.SetLimits(hlim)

	// Je Spielart ein eigener Mux mit den bisherigen Pfaden (/ws, /api/…, /media/), unter /kaiser/
	// bzw. /hanse/ eingehängt. Bilder liegen in einem gemeinsamen Media-Ordner (zufällige Namen).
	kmux := http.NewServeMux()
	kmux.Handle("/ws", khub)
	kmux.Handle("/upload", khub.UploadHandler(mstore))
	kmux.Handle("/media/", mstore)
	kmux.Handle("/api/ruhmeshalle", kserver.HallHandler(kst))
	kmux.Handle("/api/status", khub.StatusHandler())
	kmux.Handle("/api/games", khub.ListHandler())
	serveClient(kmux, filepath.Join(*static, "kaiser"), *static != "")

	hmux := http.NewServeMux()
	hmux.Handle("/ws", hhub)
	hmux.Handle("/upload", hhub.UploadHandler(mstore))
	hmux.Handle("/media/", mstore)
	hmux.Handle("/api/ruhmeshalle", hserver.HallHandler(hst))
	hmux.Handle("/api/status", hhub.StatusHandler())
	hmux.Handle("/api/games", hhub.ListHandler())
	serveClient(hmux, filepath.Join(*static, "hanse"), *static != "")

	// Startseite (Spielauswahl) unter /; Web() setzt Header, gzip und Cache-Regeln je Teil mit den
	// Pfaden nach dem Abschneiden des Präfixes, damit /ws und /assets/ richtig erkannt werden.
	root := http.NewServeMux()
	root.Handle("/kaiser/", http.StripPrefix("/kaiser", kserver.Web(kmux)))
	root.Handle("/hanse/", http.StripPrefix("/hanse", hserver.Web(hmux)))
	root.Handle("/kaiser", http.RedirectHandler("/kaiser/", http.StatusMovedPermanently))
	root.Handle("/hanse", http.RedirectHandler("/hanse/", http.StatusMovedPermanently))
	// Lebenszeichen für Docker-Healthcheck und Reverse-Proxy
	root.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })
	lobby := http.NewServeMux()
	serveClient(lobby, filepath.Join(*static, "lobby"), *static != "")
	root.Handle("/", kserver.Web(lobby))

	// HTTP-Anfragen je IP begrenzen (20/s, Vorrat 200 für das erste Laden mit allen Wappen);
	// Zeitlimits gegen langsame Angriffe, kleine Header
	handler := kserver.LimitHTTP(root, kserver.NewRateLimiter(20, 200), klim.TrustProxy)
	srv := &http.Server{Addr: *addr, Handler: handler, ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout: 2 * time.Minute, MaxHeaderBytes: 16 << 10}

	// Sauber beenden bei SIGTERM (docker stop) und Strg+C; danach schließt main die Datenbanken.
	stop, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	go func() {
		<-stop.Done()
		ctx, done := context.WithTimeout(context.Background(), 10*time.Second)
		defer done()
		slog.Info("hkd wird beendet")
		srv.Shutdown(ctx)
	}()

	slog.Info("hkd läuft", "addr", *addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fail("server", err)
	}
}

// serveClient hängt einen gebauten Client als Dateiserver unter / ein, sofern der Ordner existiert.
func serveClient(mux *http.ServeMux, dir string, enabled bool) {
	if !enabled {
		return
	}
	if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
		mux.Handle("/", http.FileServer(http.Dir(dir)))
	}
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
