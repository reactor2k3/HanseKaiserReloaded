// Typen des WebSocket-Protokolls, gespiegelt aus game/server/internal/server und engine.

/** Abschnitte eines Jahres (engine/game.go): Handel, Ereignis, Abschluss, fertig */
export type Phase = 'trade' | 'event' | 'feud' | 'done';

export interface ArmsSpec {
  /** Hauszeichen (Anfangsbuchstabe in Hausfarbe) oder hochgeladenes Familienwappen */
  kind: 'house' | 'upload';
  variant?: number;
  media: string;
}

export interface Settings {
  phaseSeconds: number;
  battleMode: 'tactical' | 'quick';
  mortality: Mortality;
  difficulty: Difficulty;
  mode: 'live' | 'async';
}

export type Difficulty = 'leicht' | 'mittel' | 'schwer';
export type Mortality = 'unsterblich' | 'lang' | 'normal' | 'kurz';

/** Felder je Stadt sind 0–9 indiziert (0 bleibt frei, wie im Original); Waren 1–6 */
export interface YearState {
  sea: number;
  prices: number[]; saltPrice: number; arrived: number[];
  kontorCost: number[]; storePrice: number[]; cannonCost: number[]; heuer: number[];
  buildPrice: number; shipValue: number; rigCost: number; hullCost: number;
  lost: number; delivered: number;
  event?: EventState;
  feud: boolean; dividend: number; interest: number; skipped: boolean; joined: boolean;
  overfull: boolean[];
}

export interface EventState {
  id: number; lines: string[]; ask?: 'jn' | 'name'; result?: string[]; delta: number; battle?: BattleResult;
}

export interface Chronik { year: number; text: string }

export interface Player {
  name: string; female: boolean; mark: number; index: number; rank: number; lifeLeft: number;
  generation: number; out: boolean;
  ships: number; underway: number; hull: number; rigging: number; orders: number;
  kontor: boolean[]; stores: number[]; cannons: number[]; sent: number[]; stock: number[];
  shares: number[]; freeShares: number;
  prison: number; married: boolean; child: string; childFem: boolean;
  indexHist: number[] | null; history: Chronik[] | null; prisoned: number; feuds: number; attacked: number;
  motto: string; arms: ArmsSpec;
  year: YearState;
}

export interface SelfView {
  seat: number; phase: Phase; deadline: number; player: Player; year: number; epoch: number;
  exchange: boolean; canFeud: boolean;
}

export interface PublicPlayer {
  seat: number; name: string; female: boolean; rank: number; index: number; mark: number; ships: number;
  kontore: number[] | null; generation: number; out: boolean; phase: Phase; bot: string;
  motto: string; arms: ArmsSpec; freeShares: number; indexHist: number[] | null;
}

export interface BattleResult {
  attacker: number; defender: number; city: number; mode: 'flotte' | 'kontor'; focus: string;
  attShips: number; defShips: number; cannons: number; stores: number;
  attLost: number; shipLost: number; canLost: number; stoLost: number; attLoot: number; defLoot: number;
}

/** Offene Beitrittsanfrage, nur für den Gastgeber sichtbar. */
export interface JoinRequest { id: string; name: string; female: boolean }

/** Eintrag der öffentlichen Spieleliste (GET /api/games). */
export interface GameListing {
  gameId: string; host: string; players: number; maxPlayers: number;
  started: boolean; year: number; mode: string; phaseSeconds: number; joinable: boolean;
}

export interface LobbyPlayer { name: string; female: boolean; online: boolean; bot: string }

export interface GameEvent { type: string; player: number; year: number; data?: any }

export type ServerMsg =
  | { t: 'welcome'; gameId: string; token: string; seat: number; host: boolean }
  | { t: 'lobby' | 'preview'; gameId: string; started: boolean; joinable: boolean; settings: Settings; players: LobbyPlayer[] }
  | { t: 'pending'; gameId: string; token: string; host: string }
  | { t: 'requests'; gameId: string; requests: JoinRequest[] }
  | { t: 'public'; year: number; over: boolean; winner: number; full: boolean; players: PublicPlayer[]; danish?: number }
  | { t: 'self'; self: SelfView }
  | { t: 'events' | 'history'; events: GameEvent[] }
  | { t: 'games'; games: GameStatus[] }
  | { t: 'error'; code: string; message: string };

/** Stand eines gemerkten Spiels für „Meine Spiele“ (Server: gameStatus). */
export interface GameStatus {
  gameId: string; found: boolean; valid: boolean; pending: boolean; started: boolean; over: boolean; year?: number;
  name?: string; female?: boolean; title?: number; alive: boolean; host: boolean;
  myTurn: boolean; deadline: number; winner?: string; players: string[] | null;
}

export interface Action {
  type: string;
  qty?: number; good?: number; city?: number; target?: number;
  mode?: string; stance?: string; name?: string;
  motto?: string; arms?: ArmsSpec;
}
