<script lang="ts">
  // Das Lübecker Kontorhaus als feiner Holzstich (Skill hanse-design): schlanker Treppengiebel mit
  // Fialen, Backsteinfassade als zarte Schraffur, je Ware ein Geschoss mit Spitzbogen-Nischen.
  // Der Lagerstand erscheint als Reihe kleiner Warenbilder (Sack, Bündel, Tuchrolle, Fass, Topf,
  // Ballen): bis zu acht je Geschoss, gemessen an der Kapazität (900 je Lübecker Speicher).
  // Überfüllte Waren werden rot gerahmt; sie verderben zum Jahresende.
  import { app } from '../lib/state.svelte';
  import { GOODS, STORE_CAPACITY, num } from '../lib/format';

  const p = $derived(app.self!.player);
  const cap = $derived(Math.max(1, STORE_CAPACITY * p.stores[1]));
  const SLOTS = 8;
  // linke Kante der beiden Arkaden je Geschoss (je 52 breit, vier Warenbilder)
  const ARKADEN = [6, 62];
  // Geschoss je Ware, Salz zuoberst wie im Original
  const X = 104, BW = 120, TOP = 70, RH = 24;
  const stueck = (w: number) => Math.min(SLOTS, Math.ceil((p.stock[w] / cap) * SLOTS));
  // Treppengiebel: fünf Stufen je Seite, Fialen auf den Stufen
  const STUFE = 11, HOCH = 10;
  // Oberkante des Erdgeschosses unter den sechs Warengeschossen
  const eg = TOP + 6 * RH;
  const giebel = (() => {
    const pts: string[] = [`M${X} ${TOP}`];
    for (let i = 0; i < 5; i++) pts.push(`V${TOP - (i + 1) * HOCH}H${X + (i + 1) * STUFE}`);
    pts.push(`V${TOP - 6 * HOCH}H${X + BW - 5 * STUFE}`);
    for (let i = 4; i >= 0; i--) pts.push(`V${TOP - (i + 1) * HOCH}H${X + BW - i * STUFE}`);
    return pts.join('') + `V${TOP}Z`;
  })();
  const fialen = [0, 1, 2, 3, 4].flatMap((i) => [X + (i + 1) * STUFE - 2, X + BW - (i + 1) * STUFE + 2].map((fx) => ({ x: fx, y: TOP - (i + 1) * HOCH })));
</script>

<figure class="lager card" aria-label="Lager im Lübecker Kontor">
  <svg viewBox="0 0 240 246" role="img" aria-label="Lagerstand je Ware">
    <defs>
      <!-- Backstein als zarte waagerechte Schraffur mit versetzten Stoßfugen -->
      <pattern id="lg-stein" width="12" height="6" patternUnits="userSpaceOnUse">
        <path d="M0 6H12M0 3H12M3 0V3M9 3V6" stroke="#7a5f3f" stroke-width="0.35" fill="none" />
      </pattern>
      <!-- Schattenseite: Schraffur 45° -->
      <pattern id="lg-schatten" width="3" height="3" patternUnits="userSpaceOnUse" patternTransform="rotate(45)">
        <path d="M0 0V3" stroke="#241a12" stroke-width="0.45" />
      </pattern>
      <!-- Warenbilder 9×10, Linienzeichnung -->
      <g id="lg-1"><path d="M1.5 9.5Q0.5 4 2.5 2L3.5 1H5.5L6.5 2Q8.5 4 7.5 9.5Z" /><path d="M3 2.6H6" /></g>
      <g id="lg-2"><path d="M0.5 7Q2 3 4.5 3.5Q7 3 8.5 7Q7 9.5 4.5 9Q2 9.5 0.5 7Z" /><path d="M2 6.5Q4.5 5 7 6.5" /></g>
      <g id="lg-3"><ellipse cx="2" cy="6" rx="1.6" ry="3.5" /><path d="M2 2.5H8M2 9.5H8" /><ellipse cx="8" cy="6" rx="1.6" ry="3.5" /></g>
      <g id="lg-4"><path d="M1.5 1.5Q0.5 5.5 1.5 9.5H7.5Q8.5 5.5 7.5 1.5Z" /><path d="M1 3.6H8M1 7.4H8" /></g>
      <g id="lg-5"><path d="M2 3.5Q0.5 6 1.5 9.5H7.5Q8.5 6 7 3.5Z" /><path d="M2 3.5H7M3 3.5V1.8H6V3.5" /></g>
      <g id="lg-6"><rect x="0.8" y="2.5" width="7.4" height="7" rx="1" /><path d="M0.8 6H8.2M4.5 2.5V9.5" /></g>
    </defs>

    <!-- Kranbalken mit Seil und Haken vor dem Giebel -->
    <path d="M{X + BW / 2} {TOP - 6 * HOCH} V{TOP - 6 * HOCH - 9} H{X + BW / 2 + 24}" class="linie" />
    <path d="M{X + BW / 2 + 21} {TOP - 6 * HOCH - 9} V{TOP - 6 * HOCH + 6}" class="linie fein" />
    <path d="M{X + BW / 2 + 19} {TOP - 6 * HOCH + 6}q2 3 4 0" class="linie fein" />

    <!-- Giebel: Backsteinfläche, Schattenseite rechts schraffiert, Fialen -->
    <path d={giebel} class="wand" />
    <path d={giebel} fill="url(#lg-stein)" />
    <path d="M{X + BW * 0.62} {TOP}V{TOP - 6 * HOCH}H{X + BW - 5 * STUFE}V{TOP - 5 * HOCH}H{X + BW - 4 * STUFE}V{TOP - 4 * HOCH}H{X + BW - 3 * STUFE}V{TOP - 3 * HOCH}H{X + BW - 2 * STUFE}V{TOP - 2 * HOCH}H{X + BW - STUFE}V{TOP - HOCH}H{X + BW}V{TOP}Z"
      fill="url(#lg-schatten)" opacity="0.5" />
    <path d={giebel} class="umriss" />
    {#each fialen as f, i (i)}
      <path d="M{f.x} {f.y}V{f.y - 6}M{f.x - 1.4} {f.y - 4.5}H{f.x + 1.4}" class="linie fein" />
    {/each}
    <!-- Luken im Giebel: je Lübecker Speicher eine geöffnete (goldene) Luke mit Spitzbogen -->
    {#each [0, 1, 2, 3, 4] as i (i)}
      {@const lx = X + 22 + i * 17}
      {@const ly = TOP - (i === 2 ? 46 : i % 2 ? 34 : 22)}
      <path d="M{lx} {ly + 12}V{ly + 4}Q{lx + 3.5} {ly - 1} {lx + 7} {ly + 4}V{ly + 12}Z" class="luke" class:offen={i < p.stores[1]} />
    {/each}

    <!-- Geschosse -->
    {#each [1, 2, 3, 4, 5, 6] as w, r (w)}
      {@const yy = TOP + r * RH}
      {@const n = stueck(w)}
      {@const voll = p.stock[w] > cap}
      <rect x={X} y={yy} width={BW} height={RH} class="wand" />
      <rect x={X} y={yy} width={BW} height={RH} fill="url(#lg-stein)" />
      <rect x={X + BW * 0.62} y={yy} width={BW * 0.38} height={RH} fill="url(#lg-schatten)" opacity="0.35" />
      <!-- Zwei offene Spitzbogen-Arkaden mit dunklem Innenraum; die Ware liegt hell darin -->
      {#each ARKADEN as a (a)}
        <path d="M{X + a} {yy + RH - 2}V{yy + 11}C{X + a} {yy + 6} {X + a + 20} {yy + 3} {X + a + 26} {yy + 1}C{X + a + 32} {yy + 3} {X + a + 52} {yy + 6} {X + a + 52} {yy + 11}V{yy + RH - 2}Z"
          class="nische" class:voll />
      {/each}
      {#each Array.from({ length: n }, (_, k) => k) as k (k)}
        <use href="#lg-{w}" transform="translate({X + ARKADEN[k < 4 ? 0 : 1] + 3 + (k % 4) * 12.3} {yy + 6.5}) scale(1.3)" class="ware" class:voll />
      {/each}
      <path d="M{X + 4} {yy + RH - 2}H{X + BW - 4}" class="linie" />
      <path d="M{X} {yy + RH}H{X + BW}" class="linie fein" />
      <!-- Beschriftung links: Ware in Kapitälchen und Menge -->
      <text x="6" y={yy + 16} class="ware-name">{GOODS[w]}</text>
      <text x={X - 8} y={yy + 16} class="menge" class:voll text-anchor="end">{num(p.stock[w])}</text>
    {/each}
    <!-- Erdgeschoss mit Spitzbogentor und Stufe -->
    <rect x={X} y={eg} width={BW} height="22" class="wand" />
    <rect x={X} y={eg} width={BW} height="22" fill="url(#lg-stein)" />
    <path d="M{X + BW / 2 - 11} {eg + 22}V{eg + 11}Q{X + BW / 2} {eg - 1} {X + BW / 2 + 11} {eg + 11}V{eg + 22}" class="tor" />
    <path d="M{X} {TOP}V{eg + 22}H{X + BW}V{TOP}" class="umriss" />
    <path d="M90 {eg + 22}H236" class="linie" />
    <path d="M94 {eg + 25}H232" class="linie fein" />
    <text x="6" y={eg + 15} class="kapazitaet">fasst {num(cap)} je Ware</text>
  </svg>
</figure>

<style>
  .lager { margin: 0; padding: 6px 8px; }
  svg { display: block; width: 100%; height: auto; }
  /* Linienzeichnung in Tinte auf warmem Backstein-Pergament; keine Flächen in Schwarz */
  .wand { fill: #e6c9a6; stroke: none; }
  .umriss { fill: none; stroke: #241a12; stroke-width: 1.1; stroke-linejoin: round; }
  .linie { fill: none; stroke: #241a12; stroke-width: 1; stroke-linecap: round; stroke-linejoin: round; }
  .linie.fein { stroke-width: 0.6; }
  .nische { fill: #5a4630; stroke: #241a12; stroke-width: 0.8; }
  .nische.voll { stroke: #942f22; stroke-width: 1.2; }
  .ware { fill: #f3e7c9; stroke: #241a12; stroke-width: 0.55; stroke-linejoin: round; }
  .ware.voll { fill: #e8b9a8; stroke: #942f22; }
  .luke { fill: #4a3826; stroke: #241a12; stroke-width: 0.7; }
  .luke.offen { fill: #d9b45c; }
  .tor { fill: #4a3826; stroke: #241a12; stroke-width: 1; }
  .ware-name { font-family: var(--font-body); font-size: 13px; font-variant: small-caps; font-weight: 600; fill: #241a12; }
  .menge { font-family: var(--font-body); font-size: 12.5px; fill: #241a12; font-variant-numeric: tabular-nums; }
  .menge.voll { fill: #942f22; font-weight: 700; }
  .kapazitaet { font-family: var(--font-body); font-size: 11px; font-style: italic; fill: #7a5f3f; }
</style>
