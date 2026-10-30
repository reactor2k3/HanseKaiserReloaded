# HanseKaiserReloaded

Zwei C64-Klassiker als Mehrspieler-Browserspiele, neu gebaut und auf einem gemeinsamen Server:

- **Kaiser**: Ein Fürstentum im 18. Jahrhundert führen, mit Korn, Land, Steuern, Bauten und Heeren, bis zur Kaiserkrone. Ergänzt um Bausteine aus Kaiser II.
- **Hanse**: Ein Lübecker Handelshaus im 14. Jahrhundert führen, mit Salz, Koggen, Kontoren und Fehden, bis in den Senat.

Beide Spiele sind eigenständige Nachbauten mit eigenem Programmcode, eigener Oberfläche und eigenen
Klängen. Spielidee und Regelwerk der Originale dienten als Vorbild. Programmteile, Grafiken oder
Texte der Originale enthält das Projekt nicht.

## Spielen

```bash
docker compose up -d
```

Danach im Browser `http://<server>:8080/`. Die Startseite zeigt beide Spiele, die laufenden Partien
und die Ruhmeshallen. Jede Partie hat 1–6 Spieler (auch Computer-Gegner) und läuft live oder
asynchron. Auf dem Server laufen höchstens 10 Partien gleichzeitig, Kaiser und Hanse zusammen.

Das Image wird bei jedem Push auf `main` gebaut: `ghcr.io/reactor2k3/hansekaiserreloaded:latest`.

## Aufbau

| Pfad | Inhalt |
|---|---|
| `server/cmd/hkd` | gemeinsamer Go-Server: `/` Startseite, `/kaiser/` und `/hanse/` je ein Spiel |
| `server/internal/kaiser/{engine,server}` | Regeln und Spielräume von Kaiser |
| `server/internal/hanse/{engine,server}` | Regeln und Spielräume von Hanse |
| `server/internal/{store,media}` | gemeinsam: SQLite-Spielstände, hochgeladene Bilder |
| `clients/kaiser`, `clients/hanse` | Svelte-Clients, gebaut für `/kaiser/` bzw. `/hanse/` |
| `clients/lobby` | Startseite (HTML/CSS/JS ohne Build) |
| `data/kaiser`, `data/hanse` | Balance- und Weltdaten |

Jedes Spiel hat eine eigene Datenbank (`kaiser.db`, `hanse.db`), eine eigene Lobby und eine eigene
Ruhmeshalle. Gemeinsam sind Port, Startseite und die Grenze für laufende Partien (`-max-games`).

## Selbst bauen

Voraussetzungen: Go 1.27, Node 24.

```powershell
./build.ps1
server/hkd.exe -addr 127.0.0.1:8080 -data data -static dist -state spielstand
```

## Lizenz

Quellcode unter MIT (`LICENSE`). Wappen, Kartendaten und Schriften von Dritten behalten ihre
eigenen Lizenzen, siehe `DRITTANBIETER.md`.
