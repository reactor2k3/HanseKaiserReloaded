// Texte und Zahlenformate für die Oberfläche. Städte, Waren und Ränge spiegeln engine/data.go.
import type { GameEvent, PublicPlayer } from './types';

/** Städte 1–9 (Index 0 frei): Name, Ware, Entfernung von Lübeck, Lage auf der Karte (Längen-/Breitengrad) */
export const CITIES = [
  { name: '', good: 0, dist: 0, lon: 0, lat: 0 },
  { name: 'Lübeck', good: 1, dist: 1, lon: 10.69, lat: 53.87 },
  { name: 'Bergen', good: 2, dist: 170, lon: 5.32, lat: 60.39 },
  { name: 'Tönsberg', good: 4, dist: 120, lon: 10.41, lat: 59.27 },
  { name: 'Warberg', good: 5, dist: 80, lon: 12.25, lat: 57.11 },
  { name: 'Malmö', good: 6, dist: 60, lon: 13.0, lat: 55.6 },
  { name: 'Ystadt', good: 6, dist: 50, lon: 13.82, lat: 55.43 },
  { name: 'Visby', good: 5, dist: 100, lon: 18.29, lat: 57.64 },
  { name: 'Riga', good: 3, dist: 140, lon: 24.11, lat: 56.95 },
  { name: 'Nowgorod', good: 2, dist: 200, lon: 31.27, lat: 58.52 },
];
/** Seewege ab Lübeck (Länge, Breite) entlang der Küsten: Fehmarnbelt, Øresund, Kattegat, Skagerrak, südlich an
 *  Rügen und Bornholm vorbei, Irbenstraße, Finnischer Meerbusen. Mit Natural-Earth-Küsten auf Landfreiheit geprüft. */
export const SEEWEGE: [number, number][][] = [
  [],
  [],
  [[10.9, 54.05], [11.4, 54.4], [12.2, 54.65], [12.75, 55.1], [12.75, 55.7], [12.62, 56.05], [12.0, 56.6], [11.0, 57.6], [10.75, 57.98], [9.5, 58.0], [7.5, 57.75], [5.3, 58.6], [4.85, 59.4], [5.05, 60.2]],
  [[10.9, 54.05], [11.4, 54.4], [12.2, 54.65], [12.75, 55.1], [12.75, 55.7], [12.62, 56.05], [12.0, 56.6], [11.0, 57.6], [10.75, 58.5], [10.5, 59.1]],
  [[10.9, 54.05], [11.4, 54.4], [12.2, 54.65], [12.75, 55.1], [12.75, 55.7], [12.62, 56.05], [12.1, 56.7], [12.15, 57.05]],
  [[10.9, 54.05], [11.4, 54.4], [12.2, 54.65], [12.75, 55.1], [12.85, 55.5]],
  [[10.9, 54.05], [11.4, 54.4], [12.4, 54.6], [13.4, 55.15], [13.8, 55.38]],
  [[10.9, 54.05], [11.4, 54.4], [12.4, 54.6], [13.4, 54.88], [14.5, 54.85], [15.6, 54.95], [16.5, 56.0], [17.6, 57.0], [18.15, 57.6]],
  [[10.9, 54.05], [11.4, 54.4], [12.4, 54.6], [13.4, 54.88], [14.5, 54.85], [15.6, 54.95], [19.0, 55.9], [20.8, 57.3], [22.2, 57.85], [22.8, 57.88], [23.4, 57.4], [24.0, 57.05]],
  [[10.9, 54.05], [11.4, 54.4], [12.4, 54.6], [13.4, 54.88], [14.5, 54.85], [15.6, 54.95], [19.0, 56.2], [20.5, 58.6], [22.5, 59.4], [26.0, 60.0], [28.5, 59.9]],
];
/** Nowgorod liegt im Binnenland: Newa, Ladogasee und Wolchow ab dem Ende des Seewegs */
export const FLUSSWEG_NOWGOROD: [number, number][] = [[28.5, 59.9], [30.3, 59.95], [31.1, 60.0], [32.3, 60.1], [32.3, 59.9], [32.0, 59.3], [31.5, 58.85], [31.27, 58.52]];
export const GOODS = ['', 'Salz', 'Pelze', 'Leinen', 'Tran', 'Honig', 'Wolle'];
export const SEA = ['', 'tobende', 'stürmische', 'bewegte', 'ruhige'];
/** Seegang im Dativ („bei tobender See“) */
export const SEA_DAT = ['', 'tobender', 'stürmischer', 'bewegter', 'ruhiger'];
/** Zahl der Hausmarken (public/heraldik/hausmarken/0–7.svg, engine MaxArmsVariants) */
export const HAUSMARKEN = 8;
/** Bildnachweise der Hausmarken und Stadtwappen (public/heraldik/lizenzen.json) */
export interface Bildnachweis { datei: string; seite: string; urheber: string; lizenz: string; lizenzUrl: string }
let nachweise: Promise<Record<string, Bildnachweis>> | null = null;
export function loadNachweise(): Promise<Record<string, Bildnachweis>> {
  nachweise ??= fetch(`${import.meta.env.BASE_URL}heraldik/lizenzen.json`).then((r) => (r.ok ? r.json() : {})).catch(() => ({}));
  return nachweise;
}
export const MAX_STORES = 5, MAX_CANNONS = 20, STORE_CAPACITY = 900, CARGO = 100, MAX_RANK = 12;

const RANKS_M = ['', 'Bürger', 'Händler', 'Kaufmann', 'Großkaufmann', 'Patrizier', 'Bruder', 'Junker', 'Diplomat', 'Schaffer', 'Ratsherr', 'Senator', 'Senator'];
const RANKS_F = ['', 'Bürgerin', 'Händlerin', 'Kauffrau', 'Großkauffrau', 'Patrizierin', 'Schwester', 'Edelfrau', 'Diplomatin', 'Schafferin', 'Ratsfrau', 'Senatorin', 'Senatorin'];
/** Aufnahme-Urkunden je Rang (engine Guilds) */
export const GUILDS = ['', 'in die Handelsrolle', 'in die Kaufleutekompanie', 'in die Kaufmannsgilde', 'in die Große Gilde',
  'in den Klub der Reichen', 'in die Georgenbruderschaft', 'in die Junkergesellschaft', 'in die Gewandschneidergilde',
  'in die Zirkelgesellschaft', 'in den Rat', 'in den Senat'];

export const PHASES: { id: string; label: string }[] = [
  { id: 'trade', label: 'Handel' },
  { id: 'event', label: 'Ereignis' },
  { id: 'feud', label: 'Abschluss' },
  { id: 'done', label: 'Fertig' },
];

/** Hausfarben je Sitz (gedeckte Töne, gut unterscheidbar auf Pergament) */
export const HOUSE_COLORS = ['#7a1f1f', '#1f3a5f', '#3e6b2a', '#8a5a12', '#5b3a73', '#2f6b6b'];
export const houseColor = (seat: number) => HOUSE_COLORS[seat % HOUSE_COLORS.length];

const nf = new Intl.NumberFormat('de-DE');
export const num = (n: number) => nf.format(Math.round(n));

/** Anrede eines Ranges */
export function title(rank: number, female: boolean) {
  return (female ? RANKS_F : RANKS_M)[Math.max(1, Math.min(MAX_RANK - 1, rank))] ?? '';
}

/** Lebenserwartung in der Lobby wie engine/settings.go */
export const MORTALITY: { value: 'unsterblich' | 'lang' | 'normal' | 'kurz'; label: string; info: string }[] = [
  { value: 'unsterblich', label: 'Unsterblich', info: 'kein natürlicher Tod' },
  { value: 'lang', label: 'Langes Leben', info: '+6 Jahre' },
  { value: 'normal', label: 'Normal', info: 'Original, 12 Jahre' },
  { value: 'kurz', label: 'Kurzes Leben', info: '−4 Jahre' },
];

/** Spielstärke der Computer-Häuser (engine/bot.go botSkills) */
export const DIFFICULTY: { value: 'leicht' | 'mittel' | 'schwer'; label: string }[] = [
  { value: 'leicht', label: 'Leicht' },
  { value: 'mittel', label: 'Mittel' },
  { value: 'schwer', label: 'Schwer' },
];

/** Spielweisen der Computer-Häuser */
export const STRATEGIES: { value: string; label: string }[] = [
  { value: 'kaufmann', label: 'Kaufmann' },
  { value: 'kaper', label: 'Kaperfahrer' },
];
export const strategyLabel = (s: string) => STRATEGIES.find((x) => x.value === s)?.label ?? s;

/** Zugzeiten je Phase (Sekunden) wie engine/settings.go: live und asynchron */
export const LIVE_SECONDS = [0, 120, 180, 300, 600];
export const ASYNC_SECONDS = [3600, 21600, 43200, 86400, 259200];

/** Lesbare Zugzeit: Sekunden, Minuten, Stunden oder Tage */
export function zugzeit(s: number): string {
  if (!s) return 'unbegrenzt';
  if (s < 60) return `${s} Sekunden`;
  if (s < 3600) return `${s / 60} Minuten`;
  if (s < 86400) return `${s / 3600} Stunde${s === 3600 ? '' : 'n'}`;
  return `${s / 86400} Tag${s === 86400 ? '' : 'e'}`;
}

/** Restzeit für Uhr und „Meine Spiele“: m:ss, ab einer Stunde „5 h 07 min“, ab einem Tag „2 T 3 h“ */
export function restzeit(s: number): string {
  s = Math.max(0, Math.floor(s));
  if (s < 3600) return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`;
  if (s < 86400) return `${Math.floor(s / 3600)} h ${String(Math.floor((s % 3600) / 60)).padStart(2, '0')} min`;
  return `${Math.floor(s / 86400)} T ${Math.floor((s % 86400) / 3600)} h`;
}

const who = (players: PublicPlayer[], seat: number, me: number) => {
  if (seat === me) return 'Sie';
  const p = players[seat];
  return p ? `${title(p.rank, p.female)} ${p.name}` : 'Unbekannt';
};

/** Chroniktext zu einem Ereignis; leer = nicht in die Chronik. */
export function describeEvent(e: GameEvent, players: PublicPlayer[], me: number): string {
  const d = e.data ?? {};
  const w = who(players, e.player, me);
  const mine = e.player === me;
  switch (e.type) {
    case 'newYear': return `Anno ${e.year} beginnt.`;
    case 'arrivals': {
      if (!mine) return '';
      const sum = (d.arrived as number[]).reduce((s, x) => s + x, 0);
      return `Bei ${SEA_DAT[d.sea]} See trafen ${num(sum)} Ballen Waren ein` + (d.lost ? `, ${d.lost} Schiff(e) gingen verloren` : '') +
        (d.delivered ? `, ${d.delivered} neue(s) Schiff(e) vom Stapel` : '') + '.';
    }
    case 'ereignis': return mine ? (d.lines as string[]).join(' ') + (d.result ? ' ' + (d.result as string[]).join(' ') : '') : '';
    case 'kontor': return `${w} ${mine ? 'eröffnen' : 'eröffnet'} ein Kontor in ${CITIES[d.city]?.name}.`;
    case 'promotion': return `${w} ${mine ? 'werden' : 'wird'} ${d.body} ${d.verb}.`;
    case 'died': return `${mine ? 'Ihr Herr des Hauses' : w} ist verstorben.`;
    case 'heir': return `${d.name} übernimmt das Haus.`;
    case 'payout': return mine ? `Die übrigen Erben werden mit ${num(d.cost)} Mark ausgezahlt.` : '';
    case 'extinct': return `Das Haus ${players[e.player]?.name ?? ''} ist ausgestorben.`;
    case 'schuldturm': return `${w} ${mine ? 'müssen' : 'muss'} ein Jahr in den Schuldturm.`;
    case 'prison': return mine ? 'Sie sitzen im Schuldturm – dieses Jahr ruhen Ihre Geschäfte.' : '';
    case 'timeout': return mine ? 'Die Zeit ist abgelaufen – der Zug wurde automatisch fortgesetzt.' : '';
    case 'battle': {
      const att = d.attacker < 0 ? 'Waldemars Flotte' : who(players, d.attacker, me);
      const def = who(players, d.defender, me);
      const ort = CITIES[d.city]?.name ?? '';
      return d.mode === 'flotte'
        ? `Seegefecht vor ${ort}: ${att} gegen die Flotte von ${def} – ${d.shipLost} gegen ${d.attLost} versenkte Schiffe.`
        : `${att} ${d.attacker === me ? 'greifen' : 'greift'} das Kontor von ${def} in ${ort} an: ${d.shipLost} Schiffe, ${d.canLost} Kanonen, ${d.stoLost} Speicher zerstört, ${d.attLost} Angreifer versenkt.`;
    }
    case 'gameOver':
      if (d.seuche) return 'Bei einer großen Seuche verstarben alle Kaufleute. Das Spiel ist zu Ende.';
      return d.winner >= 0 ? `${who(players, d.winner, me)} ${d.winner === me ? 'werden' : 'wird'} in den Senat berufen – das Spiel ist entschieden!` : 'Das Spiel ist zu Ende.';
  }
  return '';
}
