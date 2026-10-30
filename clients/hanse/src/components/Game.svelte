<script lang="ts">
  // Hauptbildschirm: Kopfleiste mit Jahresabschnitten, Ostseekarte, Aktionsspalte, eigenes Haus
  // mit Rangliste und Chronik. Wird einmal aufgebaut; Werte aktualisieren sich über den Zustand.
  import { untrack } from 'svelte';
  import { app, autopilotAn } from '../lib/state.svelte';
  import Ratgeber from './Ratgeber.svelte';
  import { CITIES, GOODS, MAX_RANK, PHASES, houseColor, num, restzeit, strategyLabel, STRATEGIES, title } from '../lib/format';
  import Karte from './Karte.svelte';
  import Aktionen from './Aktionen.svelte';
  import Hinweise from './Hinweise.svelte';
  import Ereignis from './Ereignis.svelte';
  import SpielerKarte from './SpielerKarte.svelte';
  import Lager from './Lager.svelte';
  import ProfileDialog from './ProfileDialog.svelte';
  import Wappen from './Wappen.svelte';
  import Modal from './Modal.svelte';
  import HelpDialog from './HelpDialog.svelte';
  import Auswahl from './Auswahl.svelte';
  import { leaveGame, setAutopilot } from '../lib/net';
  import { cycleSound, setMusicAllowed, sound } from '../lib/audio.svelte';
  import { erlaubnisAnfragen, kannFragen, sieSindDran } from '../lib/benachrichtigung';

  const self = $derived(app.self!);
  const p = $derived(self.player);
  const phaseIndex = $derived(PHASES.findIndex((ph) => ph.id === self.phase));
  let profileOpen = $state(false);
  let helpOpen = $state(false);
  let leaveOpen = $state(false);
  let city = $state(6);
  let eventOpen = $state(false);
  let now = $state(Math.floor(Date.now() / 1000));
  let fragen = $state(kannFragen());

  // Musik nur, solange die Partie läuft
  $effect(() => {
    setMusicAllowed(!app.over);
    return () => setMusicAllowed(false);
  });

  // Asynchrones Spiel: melden, wenn eine neue eigene Phase beginnt
  let letztePhase = '';
  $effect(() => {
    const ph = self.phase;
    if (app.settings?.mode === 'async' && letztePhase === 'done' && ph !== 'done') {
      sieSindDran(app.year, PHASES.find((x) => x.id === ph)?.label ?? '');
    }
    letztePhase = ph;
  });

  // Sekundentakt nur für die Zugzeit-Anzeige
  $effect(() => {
    const t = setInterval(() => (now = Math.floor(Date.now() / 1000)), 1000);
    return () => clearInterval(t);
  });

  const remaining = $derived(self.deadline ? Math.max(0, self.deadline - now) : null);
  const clock = $derived(remaining === null ? '' : restzeit(remaining));
  const me = $derived(app.players[app.seat]);
  // Rangliste nach Index
  const ranking = $derived([...app.players].sort((a, b) => Number(a.out) - Number(b.out) || b.index - a.index));
  const kapazitaet = $derived(900 * p.stores[1]);
  let rangOpen = $state(false);

  // Rat des Ältermanns automatisch zu Beginn jedes Abschnitts (Handel, Ereignis, Abschluss), nach
  // den Hinweis- und Ereignisfenstern. Abschaltbar im Rat-Fenster; Wahl gilt für diesen Browser.
  const RAT_KEY = 'hanse.ratAuto';
  function loadRatAuto(): boolean {
    try {
      return localStorage.getItem(RAT_KEY) !== 'aus';
    } catch {
      return true;
    }
  }
  let ratAuto = $state(loadRatAuto());
  function setRatAuto(on: boolean) {
    ratAuto = on;
    try {
      localStorage.setItem(RAT_KEY, on ? 'an' : 'aus');
    } catch { /* privater Modus: Wahl gilt nur bis zum Neuladen */ }
  }
  let adviceOpen = $state(false);
  const RAT_PHASEN = ['trade', 'event', 'feud'];
  let ratSchritt = untrack(() => `${app.year}/${self.phase}`); // Wiedereinstieg öffnet den Rat nicht
  let ratFaellig = $state(false);
  $effect(() => {
    const schritt = `${app.year}/${self.phase}`;
    if (schritt === ratSchritt) return;
    ratSchritt = schritt;
    ratFaellig = ratAuto && !app.over && !p.out && !autopilotAn() && RAT_PHASEN.includes(self.phase);
  });
  // Warten, bis alle anderen Fenster (<dialog>) zu sind; Ereignisse treffen kurz nach dem Wechsel ein
  $effect(() => {
    if (!ratFaellig) return;
    let t: ReturnType<typeof setInterval> | undefined;
    const start = setTimeout(() => {
      t = setInterval(() => {
        if (autopilotAn()) {
          ratFaellig = false;
        } else if (!document.querySelector('dialog[open]')) {
          ratFaellig = false;
          adviceOpen = true;
        }
      }, 400);
    }, 1200);
    return () => {
      clearTimeout(start);
      clearInterval(t);
    };
  });
  // Autopilot an: Rat-Fenster schließen
  $effect(() => {
    if (autopilotAn()) adviceOpen = false;
  });
  // Fortschritt zwischen aktueller und nächster Rangschwelle (Rang · 101 Index)
  const rangAb = $derived((p.rank - 1) * 101);
  const rangBis = $derived(p.rank * 101);
  const fortschritt = $derived(Math.max(0, Math.min(100, ((p.index - rangAb) / Math.max(1, rangBis - rangAb)) * 100)));
</script>

<div class="screen">
  <header class="top">
    <div class="brand"><span class="display logo">Hanse</span><span class="label">Reloaded</span></div>
    <div class="year"><span class="label">Anno</span><span class="num">{app.year}</span></div>
    <nav aria-label="Abschnitte des Jahres">
      <ol>
        {#each PHASES.slice(0, 3) as ph, i}
          <li class:done={i < phaseIndex} class:current={i === phaseIndex} aria-current={i === phaseIndex ? 'step' : undefined}>{ph.label}</li>
        {/each}
      </ol>
    </nav>
    <div class="autopilot" class:on={me?.bot}>
      <span class="label" aria-hidden="true">Autopilot</span>
      <Auswahl dark label="Autopilot: Computer spielt für mich" value={me?.bot ?? ''} onchange={(v) => setAutopilot(v)}
        options={[{ value: '', label: 'Aus – ich spiele selbst' }, ...STRATEGIES]} />
    </div>
    {#if clock}<div class="clock num" class:urgent={remaining !== null && remaining < 15} aria-label="Verbleibende Zugzeit">{clock}</div>{/if}
    <!-- Handelshäuser als Kreise in Hausfarbe; Hover oder Fokus zeigt den Steckbrief -->
    <ul class="players" aria-label="Mitspieler">
      {#each app.players as x (x.seat)}
        <li style:background={houseColor(x.seat)} class:ready={x.phase === 'done'} class:me={x.seat === app.seat} class:dead={x.out}>
          <button class="chip" type="button" aria-label="{title(x.rank, x.female)} {x.name}" aria-describedby="spieler-{x.seat}">
            <span class="initiale" aria-hidden="true">{x.name.charAt(0).toUpperCase()}</span>
          </button>
          <SpielerKarte p={x} me={x.seat === app.seat} id="spieler-{x.seat}" />
        </li>
      {/each}
    </ul>
    <div class="tools">
      <button class="tool" type="button" aria-label="Rangliste" title="Rangliste" onclick={() => (rangOpen = true)}>
        <svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linejoin="round"><path d="M3 18h18M4 16l-1-9 5 4 4-6 4 6 5-4-1 9z" /></svg>
      </button>
      <button class="tool" type="button" onclick={cycleSound}
        aria-label="Ton: {sound.level === 'off' ? 'aus' : sound.level === 'sfx' ? 'nur Geräusche' : 'Geräusche und Musik'} – klicken zum Umschalten"
        title={sound.level === 'off' ? 'Ton aus' : sound.level === 'sfx' ? 'Geräusche an' : 'Geräusche und Musik an'}>
        <svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M11 5L6 9H2v6h4l5 4z" fill="currentColor" />
          {#if sound.level === 'off'}<path d="M22 9l-6 6M16 9l6 6" />
          {:else}<path d="M15.5 8.5a5 5 0 0 1 0 7" />{#if sound.level === 'all'}<path d="M19 5a10 10 0 0 1 0 14" />{/if}{/if}
        </svg>
      </button>
      <button class="tool" type="button" aria-label="Hilfe" title="Hilfe" onclick={() => (helpOpen = true)}>?</button>
      <button class="tool" type="button" aria-label="Spiel verlassen oder neues Spiel" title="Spiel verlassen / Neues Spiel" onclick={() => (leaveOpen = true)}>
        <svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4M10 17l-5-5 5-5M5 12h12"/></svg>
      </button>
    </div>
  </header>

  <main>
    <section class="card map" aria-label="Ostseekarte">
      <h1 class="display">Die Ostsee</h1>
      <div class="map-box"><Karte bind:selected={city} /></div>
    </section>

    <!-- Mitte: eigenes Haus mit deutlichem Rang über dem Aktionsbereich -->
    <div class="mitte">
      <section class="card ruler" aria-label="Ihr Handelshaus">
        <Wappen arms={p.arms} seat={app.seat} name={p.name} size={60} label="Ihr Familienwappen" />
        <div class="who">
          <span class="display rang">{title(p.rank, p.female)}</span>
          <span class="name">{p.name}{p.generation ? ` · ${p.generation + 1}. Generation` : ''}</span>
          {#if p.motto}<span class="motto">„{p.motto}“</span>{/if}
          <button class="edit" type="button" onclick={() => (profileOpen = true)}>Profil &amp; Wappen bearbeiten</button>
        </div>
        <!-- Rangstufe und Weg zum nächsten Rang (Schwelle wie im Original: Rang · 101 Index) -->
        <div class="stufe">
          <span class="label">Rang {p.rank} von {MAX_RANK - 1}</span>
          {#if p.rank < MAX_RANK - 1}
            <span class="leiste" role="progressbar" aria-label="Weg zum nächsten Rang" aria-valuemin={rangAb} aria-valuemax={rangBis} aria-valuenow={p.index}>
              <span style:width="{fortschritt}%"></span>
            </span>
            <span class="u">nächster: {title(p.rank + 1, p.female)} ab {num(rangBis)}</span>
          {/if}
        </div>
      </section>
      <Aktionen bind:city onEvent={() => (eventOpen = true)} />
    </div>

    <aside>
      <section class="stats" aria-label="Kennzahlen">
        <div class="card"><span class="label">Mark</span><span class="v num" class:neg={p.mark < 0}>{num(p.mark)}</span><span class="u">Barvermögen</span></div>
        <div class="card"><span class="label">Index</span><span class="v num">{num(p.index)}</span><span class="u">nächster Rang ab {p.rank * 100 + p.rank}</span></div>
        <div class="card"><span class="label">Lebensjahre</span><span class="v num" class:neg={p.lifeLeft <= 2}>{p.lifeLeft}</span><span class="u">{p.child ? `Erbe: ${p.child}` : p.married ? 'verheiratet' : 'kein Erbe'}</span></div>
        <div class="card"><span class="label">Schiffe</span><span class="v num">{p.ships + p.underway}</span><span class="u">{p.ships} im Hafen</span></div>
        <div class="card"><span class="label">Zustand</span><span class="v num" class:neg={p.hull < 40 || p.rigging < 40}>{Math.round((p.hull + p.rigging) / 2)} %</span><span class="u">Rumpf {p.hull} · Tak. {p.rigging}</span></div>
        <div class="card"><span class="label">Speicher</span><span class="v num">{p.stores.reduce((s, x) => s + x, 0)}</span><span class="u">Lübeck fasst {num(kapazitaet)}</span></div>
      </section>
      <Lager />
      <section class="card chronicle" aria-label="Chronik">
        <h2 class="display">Chronik</h2>
        <ul>
          {#each app.chronicle as c (c.id)}
            <li><span class="label">{c.year}</span><span>{c.text}</span></li>
          {/each}
        </ul>
      </section>
    </aside>
  </main>

  <footer>
    {#if app.over}
      <p><strong>Das Spiel ist beendet.</strong> {app.winner >= 0 ? `${app.players[app.winner]?.name} wurde in den Senat berufen!` : ''}</p>
      <button class="btn btn-primary" type="button" onclick={leaveGame}>Neues Spiel</button>
    {:else}
      <!-- Fußleiste: wichtigster Rat des Ältermanns, alle Ratschläge im Modal -->
      <div class="next"><Ratgeber compact /></div>
      <p class="hint">Dänische Flotte: {app.danish} Schiffe · Kontore in {CITIES.filter((_, i) => i > 0 && p.kontor[i]).map((c) => c.name).join(', ')}</p>
      <button class="btn" type="button" onclick={() => (adviceOpen = true)}>Rat des Ältermanns</button>
      {#if app.settings?.mode === 'async' && fragen}
        <button class="btn" type="button" title="Browser-Benachrichtigung, wenn Sie wieder am Zug sind"
          onclick={async () => { await erlaubnisAnfragen(); fragen = kannFragen(); }}>Benachrichtigen</button>
      {/if}
    {/if}
  </footer>
</div>

<Hinweise />
<Ereignis bind:open={eventOpen} />
<ProfileDialog bind:open={profileOpen} />
<HelpDialog bind:open={helpOpen} />

<Modal bind:open={adviceOpen} title="Der Ältermann rät">
  <Ratgeber all />
  {#snippet actions()}
    <label class="rat-auto"><input type="checkbox" checked={ratAuto} onchange={(e) => setRatAuto(e.currentTarget.checked)} />
      Bei jedem Abschnitt automatisch zeigen</label>
    <button class="btn btn-primary" type="button" onclick={() => (adviceOpen = false)}>Verstanden</button>
  {/snippet}
</Modal>

<!-- Rangliste aller Häuser nach Index, geöffnet über den Kronen-Knopf in der Kopfleiste -->
<Modal bind:open={rangOpen} title="Rangliste">
  <ol class="rangliste">
    {#each ranking as x, i (x.seat)}
      <li class:me={x.seat === app.seat} class:out={x.out} class:ready={x.phase === 'done'}>
        <span class="platz num">{i + 1}.</span>
        <Wappen arms={x.arms} seat={x.seat} size={30} label="" />
        <span class="farbe" style:background={houseColor(x.seat)}></span>
        <span class="n">
          <span class="rang-klein">{title(x.rank, x.female)} · Rang {x.rank}</span>
          <span>{x.name}{x.seat === app.seat ? ' (Sie)' : ''}{x.bot ? ` · ${strategyLabel(x.bot)}` : ''}</span>
        </span>
        <span class="num idx">{num(x.index)}</span>
      </li>
    {/each}
  </ol>
</Modal>

<Modal bind:open={leaveOpen} title="Spiel verlassen?" variant="confirm">
  <p>Sie kehren zur Startseite zurück und können ein neues Spiel anlegen oder einem anderen beitreten.</p>
  {#if !app.over}
    <p class="hint">Ihr Platz in dieser Partie bleibt bestehen. Ohne Autopilot läuft Ihre Zugzeit ab und der Zug wird automatisch fortgesetzt; mit Autopilot spielt der Computer für Sie weiter.</p>
  {/if}
  {#snippet actions()}
    <button class="btn" type="button" onclick={() => (leaveOpen = false)}>Bleiben</button>
    {#if !app.over && !me?.bot}
      <button class="btn" type="button" onclick={() => { setAutopilot('kaufmann'); setTimeout(leaveGame, 300); }}>Autopilot an und verlassen</button>
    {/if}
    <button class="btn btn-primary" type="button" onclick={leaveGame}>Verlassen</button>
  {/snippet}
</Modal>

<style>
  /* Ganzer Hauptbildschirm passt in eine Fensterhöhe; nur Chronik und Aktionen scrollen. */
  .screen { height: 100vh; height: 100dvh; display: flex; flex-direction: column; overflow: hidden; }
  .top { background: var(--ink); color: var(--parchment); border-bottom: 3px solid var(--gold); display: flex; flex-wrap: wrap; align-items: center; gap: 12px 28px; padding: 8px 12px; position: relative; z-index: 20; }
  .brand { display: flex; align-items: baseline; gap: 8px; }
  .logo { font-size: 34px; color: var(--gold-light); }
  .brand .label, .year .label { color: #cdbb94; }
  .year { display: flex; flex-direction: column; line-height: 1; font-family: var(--font-label); font-size: 24px; font-weight: 600; }
  nav { flex: 1; }
  ol { list-style: none; margin: 0; padding: 0; }
  nav ol { display: flex; flex-wrap: wrap; gap: 6px; font-family: var(--font-label); font-size: 15px; }
  nav ol li { padding: 8px 12px; border-radius: 4px; color: #a8987c; }
  nav ol li.done { color: #cdbb94; }
  nav ol li.done::before { content: '✓ '; }
  nav ol li.current { background: var(--gold); color: var(--ink); font-weight: 700; }
  .autopilot { display: flex; flex-direction: column; gap: 2px; min-width: 210px; }
  .autopilot .label { color: #cdbb94; font-size: 13px; }
  .autopilot.on :global(button[role=combobox]) { background: var(--gold); color: var(--ink); }
  .tools { display: flex; gap: 6px; }
  .tool { width: 36px; height: 36px; border-radius: 50%; border: 1px solid var(--gold); background: var(--ink); color: var(--gold-light); font-family: var(--font-label); font-size: 18px; font-weight: 700; cursor: pointer; display: flex; align-items: center; justify-content: center; }
  .tool:hover { background: var(--gold); color: var(--ink); }
  .tool:focus-visible { outline: 3px solid var(--gold); outline-offset: 2px; }
  .clock { font-family: var(--font-label); font-size: 20px; color: var(--gold-light); }
  .clock.urgent { color: #ff9b8a; }
  .players { list-style: none; margin: 0; padding: 0; display: flex; gap: 6px; }
  .players li { position: relative; width: 34px; height: 34px; border-radius: 50%; display: flex; align-items: center; justify-content: center; color: var(--gold-light); font-family: var(--font-label); font-weight: 700; box-shadow: 0 0 0 2px var(--ink); transition: transform .12s ease; }
  .players li.ready { box-shadow: 0 0 0 2px #6f8f4e; }
  .players li.me { box-shadow: 0 0 0 2px var(--gold-light); }
  .players li.dead { opacity: .4; }
  .players .chip { all: unset; width: 100%; height: 100%; display: flex; align-items: center; justify-content: center; border-radius: 50%; cursor: help; }
  /* Hausmarke klein im Kreis der Hausfarbe, auf Pergamentgrund */
  /* schlichter Kreis in Hausfarbe mit Anfangsbuchstaben; das Wappen steht im Steckbrief */
  .players .initiale { font-family: var(--font-display); font-size: 19px; line-height: 1; color: #f3e7c9; }
  .players li:hover, .players li:focus-within { transform: scale(1.12); box-shadow: 0 0 0 2px var(--gold-light), 0 0 10px #e9c45a88; }
  .players li:hover :global(.tip), .players li:focus-within :global(.tip) { opacity: 1; transform: none; visibility: visible; transition-delay: 0s; }
  .hint { color: var(--sepia); margin: 0; }

  /* Drei Spalten: Karte | Aktionen | eigenes Haus, Rangliste und Chronik */
  main { flex: 1; min-height: 0; display: grid; grid-template-columns: minmax(0, 1fr) clamp(380px, 32vw, 520px) clamp(290px, 22vw, 360px); gap: 2px; padding: 2px; }
  @media (max-width: 1250px) { main { grid-template-columns: minmax(0, 1fr) 380px 290px; } }
  .map { min-height: 0; min-width: 0; display: flex; flex-direction: column; gap: 8px; }
  .map h1 { font-size: clamp(22px, 2.2vw, 30px); }
  .map-box { flex: 1; min-height: 0; }
  aside { min-height: 0; display: flex; flex-direction: column; gap: 2px; }
  aside > section:not(.chronicle) { flex-shrink: 0; }
  aside :global(.card) { padding: 2px 6px; }
  /* Mittlere Spalte: Haus oben fest, Aktionsbereich füllt den Rest */
  .mitte { min-height: 0; min-width: 0; display: flex; flex-direction: column; gap: 2px; }
  .mitte > :global(section:last-child) { flex: 1; min-height: 0; }
  .ruler { display: flex; flex-wrap: wrap; align-items: center; gap: 6px 14px; padding: 10px 14px; flex-shrink: 0; }
  .ruler .who { flex: 1; min-width: 140px; display: flex; flex-direction: column; gap: 2px; }
  .ruler .rang { font-size: clamp(28px, 2.6vw, 38px); line-height: 1; color: var(--blood); }
  .ruler .name { font-size: 19px; font-weight: 600; }
  .stufe { display: flex; flex-direction: column; gap: 3px; min-width: 150px; }
  .leiste { display: block; height: 8px; border: 1px solid var(--ink); background: var(--parchment); }
  .leiste span { display: block; height: 100%; background: var(--gold); }
  .ruler .motto { font-style: italic; color: var(--sepia); }
  .ruler .edit { align-self: flex-start; min-height: 32px; padding: 0; border: none; background: none; color: var(--blood); font-family: var(--font-label); font-size: 14px; font-weight: 600; letter-spacing: .04em; text-decoration: underline; text-underline-offset: 3px; cursor: pointer; }
  .stats { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 2px; }
  .stats .card { padding: 2px 4px; display: flex; flex-direction: column; }
  .stats .label { font-size: 11.5px; letter-spacing: .06em; }
  .v { font-size: 22px; font-weight: 600; }
  .v.neg { color: var(--blood); }
  .u { font-size: 14px; color: var(--sepia); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .chronicle h2 { font-size: 22px; margin-bottom: 4px; }
  /* Rangliste im Modal */
  .rangliste { display: flex; flex-direction: column; gap: 4px; min-width: min(460px, 80vw); font-size: 18px; }
  .rangliste li { display: flex; align-items: center; gap: 8px; padding: 4px 0; border-bottom: 1px solid var(--line); }
  .rangliste .platz { width: 26px; text-align: right; font-weight: 700; }
  .rangliste .n { display: flex; flex-direction: column; line-height: 1.2; }
  .rangliste .rang-klein { font-family: var(--font-label); font-size: 13px; letter-spacing: .08em; text-transform: uppercase; color: var(--sepia); }
  .rangliste .idx { font-size: 20px; font-weight: 600; }
  .rangliste li.me { font-weight: 700; }
  .rangliste li.out { opacity: .45; text-decoration: line-through; }
  .rangliste li.ready .n::after { content: ' ✓'; color: var(--good); }
  .rangliste .farbe { width: 10px; height: 14px; border: 1px solid var(--ink); flex-shrink: 0; }
  .rangliste .n { flex: 1; min-width: 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .chronicle { flex: 1; min-height: 120px; overflow-y: auto; overscroll-behavior: contain; }
  .chronicle ul { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: 6px; font-size: 17px; line-height: 1.3; }
  .chronicle li { display: flex; gap: 10px; }
  .chronicle .label { padding-top: 3px; }
  footer { display: flex; flex-wrap: wrap; justify-content: space-between; align-items: center; gap: 12px; padding: 6px 12px; border-top: 1px solid var(--line); background: var(--ground-dark); }
  footer p { margin: 0; }
  .next { flex: 1; min-width: 0; }
  /* Schalter im Rat-Fenster links, „Verstanden“ bleibt rechts */
  .rat-auto { margin-right: auto; display: flex; align-items: center; gap: 6px; font-size: 16px; }
  @media (max-width: 640px) {
    .stats { grid-template-columns: repeat(2, minmax(0, 1fr)); }
    .top { gap: 6px 12px; padding: 6px 8px; }
    nav { flex: 1 1 100%; order: 3; }
    nav ol { flex-wrap: nowrap; gap: 2px; font-size: 13px; }
    nav ol li { padding: 4px 6px; }
    .autopilot { min-width: 0; flex: 1 1 150px; }
    .autopilot .label { display: none; }
  }
  /* Schmale Fenster: untereinander, der Hauptbereich scrollt */
  @media (max-width: 900px) {
    main { display: flex; flex-direction: column; overflow-y: auto; overscroll-behavior: contain; }
    main > :global(*) { flex: none; }
    .map-box { height: 45vh; flex: none; }
    .chronicle { max-height: 50vh; }
  }
</style>
