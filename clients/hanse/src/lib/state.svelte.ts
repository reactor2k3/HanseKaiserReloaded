// Zentraler Client-Zustand mit Svelte-5-Runes. Nachrichten vom Server werden
// hineingemischt statt ersetzt, damit nur geänderte Stellen neu gezeichnet werden.
import type { BattleResult, EventState, GameEvent, GameStatus, JoinRequest, LobbyPlayer, PublicPlayer, SelfView, ServerMsg, Settings } from './types';
import { neueAnfrage } from './benachrichtigung';
import { describeEvent } from './format';
import { sfxBell, sfxChime, sfxError, sfxFanfare, sfxCannon } from './audio.svelte';

export interface Toast { id: number; kind: 'info' | 'warn' | 'error'; text: string }
export interface ChronicleEntry { id: number; year: number; text: string }
/** Hinweisfenster der Reihe nach: Ankunft, Ereignis, Gefecht, Urkunde, Tod */
export type Notice =
  | { kind: 'arrivals'; year: number; arrived: number[]; sea: number; lost: number; delivered: number }
  | { kind: 'battle'; battle: BattleResult }
  | { kind: 'promotion'; rank: number; body: string; verb: string; year: number }
  | { kind: 'died'; name: string; heir: string };

export const app = $state({
  connection: 'connecting' as 'connecting' | 'open' | 'closed',
  lastAction: '', // zuletzt geschickte Aktion (Fehlermeldung „geht in dieser Phase nicht“)
  gameId: '',
  seat: -1,
  host: false,
  started: false,
  settings: null as Settings | null,
  lobby: [] as LobbyPlayer[],
  preview: null as { gameId: string; started: boolean; joinable: boolean; settings: Settings; players: LobbyPlayer[] } | null,
  pendingJoin: null as { gameId: string; host: string } | null, // eigene Beitrittsanfrage wartet auf den Gastgeber
  joinRequests: [] as JoinRequest[], // als Gastgeber: offene Anfragen anderer
  year: 0,
  over: false,
  winner: -1,
  danish: 6,
  players: [] as PublicPlayer[],
  self: null as SelfView | null,
  chronicle: [] as ChronicleEntry[],
  notices: [] as Notice[],
  toasts: [] as Toast[],
  // „Meine Spiele“ auf der Startseite; null = noch nicht abgefragt
  myGames: null as GameStatus[] | null,
});

let nextId = 1;

/** Spielt der Computer gerade für das eigene Haus? Dann öffnen sich keine Hinweisfenster. */
export const autopilotAn = () => !!app.players[app.seat]?.bot;

/** Kurzer Hinweis oben rechts, verschwindet nach einigen Sekunden von selbst. */
export function toast(kind: Toast['kind'], text: string, ms = 5000) {
  const id = nextId++;
  app.toasts.push({ id, kind, text });
  setTimeout(() => {
    const i = app.toasts.findIndex((t) => t.id === id);
    if (i >= 0) app.toasts.splice(i, 1);
  }, ms);
}

/** Überträgt Felder rekursiv in ein bestehendes Objekt, damit Proxys und Bindungen erhalten bleiben. */
function merge(target: any, src: any) {
  for (const [k, v] of Object.entries(src)) {
    if (v && typeof v === 'object' && !Array.isArray(v) && target[k] && typeof target[k] === 'object') {
      merge(target[k], v);
    } else if (target[k] !== v) {
      target[k] = v;
    }
  }
}

/** Verarbeitet eine Server-Nachricht. */
export function receive(msg: ServerMsg) {
  switch (msg.t) {
    case 'welcome':
      app.pendingJoin = null;
      app.gameId = msg.gameId;
      app.seat = msg.seat;
      app.host = msg.host;
      return msg;
    case 'preview':
      app.preview = { gameId: msg.gameId, started: msg.started, joinable: msg.joinable, settings: msg.settings, players: msg.players };
      return;
    case 'lobby':
      app.gameId = msg.gameId;
      app.started = msg.started;
      app.settings = msg.settings;
      app.lobby = msg.players;
      return;
    case 'public':
      app.year = msg.year;
      app.over = msg.over;
      app.winner = msg.winner;
      if (msg.danish !== undefined) app.danish = msg.danish;
      if (msg.full) app.players = msg.players;
      else for (const p of msg.players) {
        if (app.players[p.seat]) merge(app.players[p.seat], p);
        else app.players[p.seat] = p;
      }
      return;
    case 'self':
      if (app.self) {
        merge(app.self, msg.self);
        // Jahreswerte und Wappen vollständig übernehmen: erledigte Ereignisse dürfen nicht stehen bleiben
        app.self.player.year = msg.self.player.year;
        app.self.player.arms = msg.self.player.arms;
      } else app.self = msg.self;
      return;
    case 'events':
      for (const e of msg.events) addEvent(e);
      return;
    case 'history':
      // Wiedereinstieg: Chronik neu aufbauen, ohne Hinweisfenster erneut zu zeigen
      app.chronicle = [];
      for (const e of msg.events ?? []) addEvent(e, true);
      return;
    case 'pending':
      app.pendingJoin = { gameId: msg.gameId, host: msg.host };
      return;
    case 'requests':
      // Neue Anfrage: Ton und (bei verstecktem Tab) Browser-Meldung für den Gastgeber
      if (msg.requests.some((q) => !app.joinRequests.some((o) => o.id === q.id))) {
        sfxChime();
        neueAnfrage(msg.requests[msg.requests.length - 1].name);
      }
      app.joinRequests = msg.requests;
      return;
    case 'games':
      app.myGames = msg.games;
      return;
    case 'error':
      if (!['bad_token', 'no_game'].includes(msg.code)) sfxError();
      if (['denied', 'not_joinable', 'banned', 'kicked'].includes(msg.code)) app.pendingJoin = null;
      // Phasenfehler entstehen meist, wenn die Zugzeit abläuft, während man noch klickt: Aktion nennen
      if (msg.code === 'wrong_phase') {
        const was = ACTION_LABEL[app.lastAction] ?? app.lastAction;
        toast('error', `„${was}“ geht jetzt nicht mehr: Ihre Phase ist schon weiter (Zugzeit abgelaufen?).`);
        return;
      }
      toast('error', msg.message);
      return;
  }
}

/** Anzeigenamen der Aktionen für Fehlermeldungen */
const ACTION_LABEL: Record<string, string> = {
  sellGood: 'Ware verkaufen', sellAll: 'Alles verkaufen', buySalt: 'Salz kaufen', buyShares: 'Anteile kaufen',
  sellShares: 'Anteile verkaufen', rigging: 'Takelage trimmen', hull: 'Rumpf ausbessern', orderShips: 'Bauauftrag',
  sellShips: 'Schiffe verkaufen', sendShips: 'Schiffe senden', buyStores: 'Speicher kaufen', sellStores: 'Speicher verkaufen',
  buyCannons: 'Landkanonen', openKontor: 'Kontor eröffnen', next: 'Weiter', decide: 'Entscheidung', attack: 'Händel',
  endTurn: 'Jahr beenden', profile: 'Profil',
};

/** Ereignisse landen in der Chronik; eigene wichtige zusätzlich als Hinweisfenster. */
function addEvent(e: GameEvent, historic = false) {
  const text = describeEvent(e, app.players, app.seat);
  if (text) {
    app.chronicle.unshift({ id: nextId++, year: e.year, text });
    if (app.chronicle.length > 150) app.chronicle.length = 150;
  }
  if (historic) return;
  const mine = e.player === app.seat;
  const d = e.data ?? {};
  if (e.type === 'newYear') sfxChime();
  if (mine && e.type === 'arrivals') hinweis({ kind: 'arrivals', year: e.year, ...d });
  if (e.type === 'battle' && (d.attacker === app.seat || d.defender === app.seat)) {
    sfxCannon();
    hinweis({ kind: 'battle', battle: d as BattleResult });
  }
  if (mine && e.type === 'promotion') {
    sfxFanfare();
    hinweis({ kind: 'promotion', year: e.year, rank: d.rank, body: d.body, verb: d.verb });
  }
  if (mine && e.type === 'died') sfxBell();
  if (mine && e.type === 'heir') hinweis({ kind: 'died', name: '', heir: d.name });
  if (e.type === 'promotion' && !mine && text) toast('info', text, 6000);
  if (mine && ['schuldturm', 'timeout', 'extinct'].includes(e.type) && text) toast('warn', text, 8000);
}

/** Ereignis der eigenen Ereignisphase (aus der eigenen Sicht) */
export const myEvent = (): EventState | undefined => app.self?.player.year.event;

/** Hinweisfenster einreihen – nicht mit Autopilot: dort steht alles nur in der Chronik, und beim
 *  Abschalten sollen keine gesammelten Fenster auf einmal aufgehen. */
function hinweis(n: Notice) {
  if (!autopilotAn()) app.notices.push(n);
}
