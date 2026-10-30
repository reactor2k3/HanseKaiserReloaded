<script lang="ts">
  // Tooltip-Karte zu einem Handelshaus (Hover/Fokus auf die Kreise in der Kopfleiste):
  // Wappen, Rang, Wahlspruch, Kennzahlen, Kontore und Stand im laufenden Jahr.
  import type { PublicPlayer } from '../lib/types';
  import { CITIES, PHASES, num, strategyLabel, title } from '../lib/format';
  import Wappen from './Wappen.svelte';

  let { p, me = false, id }: { p: PublicPlayer; me?: boolean; id: string } = $props();

  const phase = $derived(PHASES.find((x) => x.id === p.phase)?.label ?? '');
  const kontore = $derived((p.kontore ?? []).filter((c) => c > 1).map((c) => CITIES[c].name));
</script>

<span class="tip" role="tooltip" {id}>
  <span class="head">
    <Wappen arms={p.arms} seat={p.seat} name={p.name} size={40} label="Wappen von {p.name}" />
    <span class="who">
      <span class="rank">{title(p.rank, p.female)}{p.generation ? ` · ${p.generation + 1}. Generation` : ''}</span>
      <span class="name">{p.name}{me ? ' (Sie)' : ''}</span>
      {#if p.motto}<span class="motto">„{p.motto}“</span>{/if}
    </span>
  </span>
  <span class="facts">
    <span><small>Index</small><b class="num">{num(p.index)}</b></span>
    <span><small>Mark</small><b class="num" class:neg={p.mark < 0}>{num(p.mark)}</b></span>
    <span><small>Schiffe</small><b class="num">{p.ships}</b></span>
    <span><small>Anteile frei</small><b class="num">{p.freeShares}</b></span>
  </span>
  <span class="kontore">Kontore: Lübeck{kontore.length ? `, ${kontore.join(', ')}` : ''}</span>
  <span class="state">
    {#if p.out}✝ ausgestorben
    {:else if p.phase === 'done'}<span class="ok">✓</span> Jahr abgeschlossen
    {:else}<span class="busy">◆</span> am Zug: {phase}{/if}
    {#if p.bot}<span class="bot">Computer · {strategyLabel(p.bot)}</span>{/if}
  </span>
</span>

<style>
  /* Pergamentkarte mit Goldfaden und kleinem Zeiger nach oben; sichtbar über .players li:hover/:focus-within */
  .tip {
    position: absolute; z-index: 300; top: calc(100% + 12px); right: -8px; width: 290px;
    display: flex; flex-direction: column; gap: 8px; padding: 10px 12px;
    background: linear-gradient(180deg, #fffaf0, #f3e6c8); color: var(--ink);
    border: 1.5px solid #5a4630; border-radius: 4px;
    box-shadow: inset 0 0 0 3px #fffaf0, inset 0 0 0 4px #b08a3e99, 0 10px 24px #00000066;
    font-family: var(--font-body); font-size: 17px; font-weight: 400; text-align: left; line-height: 1.3;
    pointer-events: none; opacity: 0; transform: translateY(-4px); visibility: hidden;
    transition: opacity .14s ease, transform .14s ease, visibility 0s linear .14s;
  }
  .tip::before {
    content: ''; position: absolute; top: -8px; right: 18px; width: 14px; height: 14px;
    background: #fffaf0; border-left: 1.5px solid #5a4630; border-top: 1.5px solid #5a4630; transform: rotate(45deg);
  }
  .head { display: flex; gap: 10px; align-items: center; }
  .who { display: flex; flex-direction: column; min-width: 0; }
  .rank { font-family: var(--font-label); font-size: 12px; letter-spacing: .1em; text-transform: uppercase; color: var(--sepia); }
  .name { font-family: var(--font-display); font-size: 24px; line-height: 1.05; }
  .motto { font-style: italic; color: var(--sepia); font-size: 16px; }
  .facts { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 4px; padding: 6px 0; border-top: 1px solid #c9a95c88; border-bottom: 1px solid #c9a95c88; }
  .facts span { display: flex; flex-direction: column; align-items: center; }
  .facts small { font-family: var(--font-label); font-size: 11px; letter-spacing: .08em; text-transform: uppercase; color: var(--sepia); }
  .facts b { font-size: 18px; font-weight: 600; }
  .neg { color: var(--blood); }
  .kontore { font-size: 15.5px; color: var(--sepia); }
  .state { display: flex; flex-wrap: wrap; align-items: center; gap: 4px 10px; font-size: 16px; }
  .ok { color: var(--good); font-weight: 700; }
  .busy { color: var(--blood); font-size: 11.5px; }
  .bot { margin-left: auto; font-family: var(--font-label); font-size: 12px; letter-spacing: .06em; color: var(--sepia); }
</style>
