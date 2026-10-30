// Tests für Texte und Zustandsverarbeitung des Clients (ohne Browser).
import { describe, expect, it } from 'vitest';
import { describeEvent, num, title, zugzeit } from './format';
import { app, receive } from './state.svelte';
import type { PublicPlayer } from './types';

const player = (seat: number, name: string, female = false): PublicPlayer => ({
  seat, name, female, rank: 3, index: 120, mark: 10000, ships: 2, kontore: [1, 6], generation: 0, out: false,
  phase: 'trade', bot: '', motto: '', arms: { kind: 'house', media: '' }, freeShares: 10, indexHist: [],
});

describe('Texte', () => {
  it('formatiert Zahlen deutsch und Ränge je Anrede', () => {
    expect(num(14200)).toBe('14.200');
    expect(title(3, false)).toBe('Kaufmann');
    expect(title(3, true)).toBe('Kauffrau');
    expect(title(11, true)).toBe('Senatorin');
    expect(zugzeit(180)).toBe('3 Minuten');
  });

  it('beschreibt Gefechte aus Sicht beider Seiten', () => {
    const ps = [player(0, 'Micha'), player(1, 'Gertrud', true)];
    const d = { attacker: 0, defender: 1, city: 6, mode: 'kontor', shipLost: 1, canLost: 2, stoLost: 0, attLost: 3 };
    expect(describeEvent({ type: 'battle', player: -1, year: 1360, data: d }, ps, 0)).toContain('Sie greifen das Kontor von Kauffrau Gertrud in Ystadt an');
    expect(describeEvent({ type: 'battle', player: -1, year: 1360, data: { ...d, attacker: -1 } }, ps, 1)).toContain('Waldemars Flotte');
  });

  it('zeigt fremde Ankünfte nicht in der eigenen Chronik', () => {
    const ps = [player(0, 'Micha'), player(1, 'Gertrud', true)];
    const e = { type: 'arrivals', player: 1, year: 1357, data: { arrived: [0, 0, 0, 0, 0, 0, 17, 0, 0, 0], sea: 1, lost: 0, delivered: 0 } };
    expect(describeEvent(e, ps, 0)).toBe('');
    expect(describeEvent({ ...e, player: 0 }, ps, 0)).toBe('Bei tobender See trafen 17 Ballen Waren ein.');
  });
});

describe('Zustand', () => {
  it('mischt Teilübersichten in bestehende Spieler', () => {
    app.seat = 0;
    receive({ t: 'public', year: 1357, over: false, winner: -1, full: true, players: [player(0, 'Micha'), player(1, 'Gertrud', true)] });
    receive({ t: 'public', year: 1357, over: false, winner: -1, full: false, players: [{ ...player(1, 'Gertrud', true), index: 300 }] });
    expect(app.players[1].index).toBe(300);
    expect(app.players[0].name).toBe('Micha');
  });

  it('legt eigene Ankünfte als Hinweisfenster an', () => {
    app.notices.length = 0;
    receive({ t: 'events', events: [{ type: 'arrivals', player: 0, year: 1357, data: { arrived: [0, 0, 0, 0, 0, 0, 17, 0, 0, 0], sea: 2, lost: 0, delivered: 0 } }] });
    expect(app.notices[0]?.kind).toBe('arrivals');
  });
});
