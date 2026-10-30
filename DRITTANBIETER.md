# Bestandteile von Dritten

Der Quellcode von HanseKaiserReloaded steht unter der MIT-Lizenz (`LICENSE`). Die folgenden Bestandteile
stammen von Dritten und behalten ihre eigenen Lizenzen.

| Bestandteil | Ort | Herkunft | Lizenz |
|---|---|---|---|
| Landeswappen (28 Varianten) | `clients/kaiser/public/heraldik/laender/` | Wikimedia Commons, Urheber je Datei in `LIZENZEN.md` | CC BY-SA 3.0/4.0, CC BY 3.0, CC0, gemeinfrei (je Datei) |
| Reichskarte (Grenzen, Flüsse) | `clients/kaiser/src/assets/map.json` | [Natural Earth](https://www.naturalearthdata.com), vereinfacht | gemeinfrei |
| Schriften Cinzel, EB Garamond, UnifrakturMaguntia, Press Start 2P | per Google Fonts eingebunden | Google Fonts | SIL Open Font License 1.1 |

Eigene Bestandteile (MIT): Spielcode, Oberfläche, Favicon, Krone und Fanfaren der Krönungsurkunde,
Klänge und Musik (im Browser erzeugt, `clients/*/src/lib/audio.svelte.ts`).

## Vorbilder

- **Kaiser** (C64, 1984): Spielidee und Regelwerk dienten als Vorbild. Es werden keine Programmteile,
  Grafiken oder Texte des Originals verwendet; Regeln sind in eigenen Worten beschrieben.
- **Kaiser II** (Michael Eisermann, Oliver Schablitzki, 1987, als Public Domain veröffentlicht):
  Ereignisse, Innenpolitik, Bündnisse, Finanzen sowie Flotten und Kolonien wurden als Spielbausteine
  übernommen und frei nacherzählt.

## Hanse

- **Hanse** (C64): Spielidee und Regelwerk dienten als Vorbild. Es werden keine Programmteile, Grafiken
  oder Texte des Originals verwendet; das Spiel ist ein eigenständiger Nachbau.
- **Hausmarken und Stadtwappen** (`clients/hanse/public/heraldik/`): Wikimedia Commons, je Bild mit
  Urheber und Lizenz in `heraldik/LIZENZEN.md` und `heraldik/lizenzen.json`, im Spiel unter Hilfe → Bildnachweis.
- **Küstenlinien der Ostseekarte**: Natural Earth (gemeinfrei), umgerechnet mit `scripts/map/build_kueste.py` im Projekt HanseReloaded.
