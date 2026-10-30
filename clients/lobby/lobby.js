// Gemeinsame Startseite: lädt die öffentlichen Spielelisten und Ruhmeshallen beider Spielarten
// (/kaiser/api/… und /hanse/api/…) und zeigt sie zusammen an. Nur lesend, keine Eingaben.
// Alle Texte vom Server werden per textContent gesetzt (kein HTML aus Spielernamen).

const SPIELE = [
  { id: 'kaiser', name: 'Kaiser', titel: (t) => KAISER_TITEL[t] ?? '' },
  { id: 'hanse', name: 'Hanse', titel: (t) => HANSE_TITEL[t] ?? '' },
];
// Höchste erreichte Titel je Spielart (wie in den Clients, männliche Form)
const KAISER_TITEL = ['', 'Herr', 'Baron', 'Landgraf', 'Markgraf', 'Fürst', 'Herzog', 'Kurfürst', 'König', 'Kaiser'];
const HANSE_TITEL = ['', 'Bürger', 'Händler', 'Kaufmann', 'Großkaufmann', 'Patrizier', 'Bruder', 'Junker', 'Diplomat', 'Schaffer', 'Ratsherr', 'Senator'];

const listeEl = document.getElementById('spiele');
const grenzeEl = document.getElementById('spiele-grenze');

/** Kleines DOM-Hilfsmittel: Element mit Klasse und Text. */
function el(tag, cls, text) {
  const e = document.createElement(tag);
  if (cls) e.className = cls;
  if (text !== undefined) e.textContent = text;
  return e;
}

/** JSON holen; bei Fehlern null (die Seite bleibt benutzbar). */
async function holen(url) {
  try {
    const r = await fetch(url, { cache: 'no-store' });
    return r.ok ? await r.json() : null;
  } catch {
    return null;
  }
}

/** Laufende Spiele beider Spielarten; beitretbare zuerst, mit Link samt Spielcode. */
async function spieleLaden() {
  const [k, h, status] = await Promise.all([holen('/kaiser/api/games'), holen('/hanse/api/games'), holen('/kaiser/api/status')]);
  const alle = [
    ...(k?.games ?? []).map((g) => ({ ...g, spiel: SPIELE[0] })),
    ...(h?.games ?? []).map((g) => ({ ...g, spiel: SPIELE[1] })),
  ].sort((a, b) => Number(b.joinable) - Number(a.joinable));
  listeEl.replaceChildren();
  if (status) grenzeEl.textContent = `${status.games} von ${status.maxGames} Spielen belegt (Kaiser und Hanse zusammen).`;
  if (!k && !h) {
    listeEl.append(el('li', 'leer', 'Die Spielelisten sind gerade nicht erreichbar.'));
    return;
  }
  if (!alle.length) {
    listeEl.append(el('li', 'leer', 'Gerade läuft kein Spiel. Legen Sie eines an!'));
    return;
  }
  for (const g of alle) {
    const li = el('li', 'zeile');
    li.append(el('span', `marke ${g.spiel.id}`, g.spiel.name));
    const info = el('span', 'info');
    info.append(el('strong', '', `Spiel von ${g.host}`));
    const stand = g.started ? `Anno ${g.year}` : 'wartet auf Spieler';
    info.append(el('span', 'klein', `${g.players}/${g.maxPlayers} Spieler · ${stand} · ${g.mode === 'async' ? 'asynchron' : 'live'}`));
    li.append(info);
    if (g.joinable) {
      const a = el('a', 'knopf klein-knopf', 'Beitreten');
      a.href = `/${g.spiel.id}/?spiel=${encodeURIComponent(g.gameId)}`;
      li.append(a);
    } else {
      li.append(el('span', 'klein', 'läuft'));
    }
    listeEl.append(li);
  }
}

/** Bestenliste je Spielart (Siege, höchster Titel, Partien). */
async function hallenLaden() {
  for (const s of SPIELE) {
    const ziel = document.getElementById(`halle-${s.id}`);
    const daten = await holen(`/${s.id}/api/ruhmeshalle`);
    ziel.replaceChildren();
    const best = daten?.best ?? [];
    if (!best.length) {
      ziel.append(el('li', 'leer', daten ? 'Noch keine beendete Partie.' : 'Gerade nicht erreichbar.'));
      continue;
    }
    for (const b of best) {
      const li = el('li', 'zeile');
      li.append(el('strong', '', b.name));
      li.append(el('span', 'klein', `${s.titel(b.bestTitle)} · ${b.crowns} Sieg${b.crowns === 1 ? '' : 'e'} · ${b.games} Partie${b.games === 1 ? '' : 'n'}`));
      ziel.append(li);
    }
  }
}

spieleLaden();
hallenLaden();
// Spieleliste alle 20 Sekunden auffrischen, solange die Seite sichtbar ist
setInterval(() => {
  if (!document.hidden) spieleLaden();
}, 20000);
