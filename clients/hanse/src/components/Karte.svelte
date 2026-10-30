<script lang="ts">
  // Seekarte der Ostsee: neun Hansestädte nach echter Lage, Seewege entlang der Küsten (SEEWEGE),
  // Kontore aller Häuser als farbige Wimpel, eigene ausgesandte Schiffe als Zahl.
  // Küsten aus Natural Earth (gemeinfrei, scripts/map/build_kueste.py), gleiche Projektion wie hier.
  // Zoom: Mausrad oder Knöpfe, Verschieben mit gedrückter Maus; Beschriftungen bleiben gleich groß.
  import { app } from '../lib/state.svelte';
  import { LAND } from '../lib/kueste';
  import { CITIES, FLUSSWEG_NOWGOROD, GOODS, SEEWEGE, houseColor } from '../lib/format';

  let { selected = $bindable(0) }: { selected?: number } = $props();

  // Plattkarte mit Breitenkorrektur um 57° N
  const W = 800, H = 430, LON0 = 3.5, LON1 = 33, LAT0 = 52.8, LAT1 = 61.2;
  const K = Math.cos((57 * Math.PI) / 180);
  const sx = W / ((LON1 - LON0) * K);
  const x = (lon: number) => (lon - LON0) * K * sx;
  const y = (lat: number) => H - ((lat - LAT0) / (LAT1 - LAT0)) * H;
  const pfad = (pts: [number, number][]) => pts.map(([lo, la], i) => `${i ? 'L' : 'M'}${x(lo).toFixed(1)} ${y(la).toFixed(1)}`).join('');

  // Seeweg je Stadt: Lübeck → Wegpunkte → Stadt (Nowgorod: letzter Abschnitt über die Flüsse)
  const wege = CITIES.map((c, i) => (i < 2 ? '' : pfad([[CITIES[1].lon, CITIES[1].lat], ...SEEWEGE[i], ...(i === 9 ? [] : [[c.lon, c.lat] as [number, number]])])));
  const fluss = pfad(FLUSSWEG_NOWGOROD);

  const me = $derived(app.self?.player);
  const owners = $derived(CITIES.map((_, c) => app.players.filter((p) => !p.out && p.kontore?.includes(c)).map((p) => p.seat)));

  // Fahrende Koggen je Haus in Hausfarbe, auf den Wegen zu den Kontoren; eigene auch zu ausgesandten Zielen
  // (sent); fremde Häuser nur auf den Wegen zu ihren Kontoren, denn die Fahrtziele der anderen
  // bleiben geheim (Fehde). Bei „Bewegung reduzieren“ fahren keine Schiffe.
  const still = typeof matchMedia === 'function' && matchMedia('(prefers-reduced-motion: reduce)').matches;
  const fahrten = $derived(still ? [] : app.players.filter((p) => !p.out && p.ships > 0).flatMap((p) => {
    // Alle Häuser auf den Wegen zu ihren Kontoren; eigene zusätzlich dorthin, wohin gerade gesandt
    const ziele = [...new Set([...(p.kontore ?? []),
      ...(p.seat === app.seat && me ? CITIES.map((_, c) => c).filter((c) => me.sent[c] > 0) : [])])].filter((c) => c > 1);
    return ziele.map((c, k) => ({ seat: p.seat, city: c, dauer: 16 + c * 2, versatz: (p.seat * 2.7 + k * 5.3) % 16 }));
  }));

  // Ausschnitt (viewBox); s = Vergrößerung, Beschriftungen werden mit 1/s skaliert
  let view = $state({ x: 0, y: 0, w: W, h: H });
  const s = $derived(W / view.w);
  let svg: SVGSVGElement;

  function clamp(v: { x: number; y: number; w: number; h: number }) {
    const w = Math.min(W, Math.max(W / 6, v.w));
    const h = (w / W) * H;
    return { w, h, x: Math.min(W - w, Math.max(0, v.x)), y: Math.min(H - h, Math.max(0, v.y)) };
  }
  // Zoomen um einen Punkt (Kartenkoordinaten), Faktor > 1 = näher heran
  function zoom(f: number, cx = view.x + view.w / 2, cy = view.y + view.h / 2) {
    const w = view.w / f, h = view.h / f;
    view = clamp({ x: cx - ((cx - view.x) / view.w) * w, y: cy - ((cy - view.y) / view.h) * h, w, h });
  }
  function punkt(e: { clientX: number; clientY: number }) {
    const r = svg.getBoundingClientRect();
    // viewBox wird mit „meet“ eingepasst: Maßstab und Rand aus dem kleineren Seitenverhältnis
    const k = Math.min(r.width / view.w, r.height / view.h);
    const ox = (r.width - view.w * k) / 2, oy = (r.height - view.h * k) / 2;
    return { x: view.x + (e.clientX - r.left - ox) / k, y: view.y + (e.clientY - r.top - oy) / k, k };
  }
  function onWheel(e: WheelEvent) {
    e.preventDefault();
    const p = punkt(e);
    zoom(e.deltaY < 0 ? 1.25 : 0.8, p.x, p.y);
  }
  // Ziehen zum Verschieben; ein kurzer Klick bleibt ein Klick (Stadt wählen)
  let drag = $state<{ x: number; y: number; vx: number; vy: number; k: number; moved: boolean } | null>(null);
  function onDown(e: PointerEvent) {
    const p = punkt(e);
    drag = { x: e.clientX, y: e.clientY, vx: view.x, vy: view.y, k: p.k, moved: false };
  }
  function onMove(e: PointerEvent) {
    if (!drag) return;
    const dx = e.clientX - drag.x, dy = e.clientY - drag.y;
    if (!drag.moved && Math.abs(dx) + Math.abs(dy) < 4) return;
    if (!drag.moved) svg.setPointerCapture(e.pointerId);
    drag.moved = true;
    view = clamp({ ...view, x: drag.vx - dx / drag.k, y: drag.vy - dy / drag.k });
  }
  function onUp() {
    drag = null;
  }
</script>

<div class="rahmen">
  <svg bind:this={svg} viewBox="{view.x} {view.y} {view.w} {view.h}" role="img" aria-label="Karte der Ostsee mit den Hansestädten"
    class="karte" class:zieht={!!drag?.moved} onwheel={onWheel} onpointerdown={onDown} onpointermove={onMove} onpointerup={onUp} onpointercancel={onUp}>
    <defs>
      <!-- Holzstich: See mit waagerechten Schraffen, Land mit schräger Schraffur (Skill hanse-design) -->
      <pattern id="see" width="12" height="5" patternUnits="userSpaceOnUse">
        <path d="M0 2.5H12" stroke="#5f7a70" stroke-width="0.5" opacity="0.55" />
      </pattern>
      <pattern id="landschraffur" width="5" height="5" patternUnits="userSpaceOnUse" patternTransform="rotate(45)">
        <path d="M0 0V5" stroke="#7a5f3f" stroke-width="0.5" opacity="0.35" />
      </pattern>
    </defs>
    <rect width={W} height={H} fill="var(--sea, #cfd8c9)" />
    <rect width={W} height={H} fill="url(#see)" />
    <!-- Landflächen mit Küstenschatten -->
    <path d={LAND} class="land-schatten" style:stroke-width="{6 / s}" />
    <path d={LAND} class="land" style:stroke-width="{0.9 / s}" />
    <path d={LAND} fill="url(#landschraffur)" fill-rule="evenodd" />

    <!-- Seewege ab Lübeck; eigene Kontore kräftig, übrige gestrichelt -->
    {#each wege as d, i (i)}
      {#if d}
        <path {d} id="weg-{i}" class="route" class:mine={me?.kontor[i]} style:stroke-width="{(me?.kontor[i] ? 2.2 : 1.2) / s}" style:stroke-dasharray="{0.1 / s} {(me?.kontor[i] ? 4.5 : 5) / s}" />
      {/if}
    {/each}
    <path d={fluss} class="route fluss" class:mine={me?.kontor[9]} style:stroke-width="{1.5 / s}" style:stroke-dasharray="{1.5 / s} {3 / s}" />

    <!-- Koggen pendeln Lübeck ↔ Ziel (SMIL animateMotion entlang des Seewegs); Größe bleibt beim Zoomen gleich -->
    {#each fahrten as f (`${f.seat}-${f.city}`)}
      <g class="kogge" aria-hidden="true">
        <g transform="scale({1.6 / s})">
          <path d="M-7 0H7L4.5 3.5H-4.5Z" class="rumpf" />
          <path d="M0 0V-12" class="mast" />
          <rect x="-4" y="-11" width="8" height="7" fill={houseColor(f.seat)} class="segel" />
        </g>
        <animateMotion dur="{f.dauer}s" begin="-{f.versatz}s" repeatCount="indefinite" rotate="0"
          keyPoints="0;1;1;0;0" keyTimes="0;0.45;0.5;0.95;1" calcMode="linear">
          <mpath href="#weg-{f.city}" />
        </animateMotion>
      </g>
    {/each}

    {#each CITIES as c, i (i)}
      {#if i > 0}
        <g class="stadt" class:sel={selected === i} transform="translate({x(c.lon)} {y(c.lat)}) scale({1 / s})"
          role="button" tabindex="0" aria-label="{c.name}, {GOODS[c.good]}"
          onclick={() => (selected = i)} onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && (selected = i)}>
          <!-- Stadt als kleiner Holzstich-Turm mit Spitzdach; Lübeck größer -->
          <g transform="scale({i === 1 ? 1.25 : 1})">
            <path d="M-5 6V-4L0 -11L5 -4V6Z" class="turm" />
            <path d="M-5 -4H5M-1.5 6V2Q0 0 1.5 2V6" class="turm-linie" />
          </g>
          <!-- Stadtwappen bzw. -siegel (Wikimedia Commons, Bildnachweis in der Hilfe) -->
          <image href="{import.meta.env.BASE_URL}heraldik/staedte/{i}.svg" x="-30" y="-11" width="16" height="19" preserveAspectRatio="xMidYMid meet" />
          <text y="-15" class="name">{c.name}</text>
          <text y="22" class="ware">{GOODS[c.good]}</text>
          {#each owners[i] as seat, k (seat)}
            <rect x={10 + k * 7} y="-6" width="6" height="12" fill={houseColor(seat)} stroke="#2b2118" stroke-width="0.6" />
          {/each}
          {#if me && me.sent[i] > 0}
            <text y="34" class="schiffe" text-anchor="middle">⛵ {me.sent[i]}</text>
          {/if}
        </g>
      {/if}
    {/each}

    <!-- Windrose (bleibt unten rechts im Ausschnitt) -->
    <g transform="translate({view.x + view.w - 50 / s} {view.y + view.h - 50 / s}) scale({1 / s})" class="rose" aria-hidden="true">
      <circle r="24" />
      <path d="M0 -30L5 0L0 30L-5 0Z" />
      <path d="M-30 0L0 5L30 0L0 -5Z" opacity="0.6" />
      <text y="-34" text-anchor="middle">N</text>
    </g>
  </svg>
  <div class="zoom" role="group" aria-label="Kartenausschnitt">
    <button type="button" aria-label="Heranzoomen" onclick={() => zoom(1.5)}>+</button>
    <button type="button" aria-label="Herauszoomen" onclick={() => zoom(1 / 1.5)}>−</button>
    <button type="button" aria-label="Ganze Karte" onclick={() => (view = { x: 0, y: 0, w: W, h: H })}>⤢</button>
  </div>
</div>

<style>
  .rahmen { position: relative; width: 100%; height: 100%; }
  .karte { width: 100%; height: 100%; display: block; touch-action: none; cursor: grab; user-select: none; }
  .karte.zieht { cursor: grabbing; }
  .land { fill: #e9d8b3; stroke: #241a12; fill-rule: evenodd; }
  .land-schatten { fill: none; stroke: #9aab9f; opacity: 0.7; }
  .route { fill: none; stroke: #241a12; opacity: 0.55; stroke-linecap: round; stroke-linejoin: round; }
  .route.mine { stroke: var(--blood); opacity: 0.95; }
  .kogge { pointer-events: none; }
  .rumpf { fill: #7a5f3f; stroke: var(--ink); stroke-width: 0.9; stroke-linejoin: round; }
  .mast { stroke: var(--ink); stroke-width: 1; }
  .segel { stroke: var(--ink); stroke-width: 0.8; }
  .stadt { cursor: pointer; outline: none; }
  .turm { fill: var(--parchment); stroke: var(--ink); stroke-width: 1.4; stroke-linejoin: round; }
  .turm-linie { fill: none; stroke: var(--ink); stroke-width: 1; }
  .stadt.sel .turm, .stadt:focus-visible .turm { fill: var(--gold-light); stroke: var(--blood); stroke-width: 2; }
  .name { font-family: var(--font-display); font-size: 16px; font-weight: 700; text-anchor: middle; fill: var(--ink); paint-order: stroke; stroke: #f3e7c9; stroke-width: 3px; }
  .ware { font-family: var(--font-body); font-size: 11px; text-anchor: middle; fill: var(--sepia); paint-order: stroke; stroke: #f3e7c9; stroke-width: 3px; }
  .schiffe { font-family: var(--font-body); font-size: 12px; fill: var(--prussian); font-weight: 600; }
  .rose circle { fill: none; stroke: var(--sepia); }
  .rose path { fill: var(--sepia); }
  .rose text { font-family: var(--font-label); font-size: 11px; fill: var(--sepia); }
  .zoom { position: absolute; top: 6px; right: 6px; display: flex; flex-direction: column; gap: 4px; }
  .zoom button {
    width: 30px; height: 30px; border: 1px solid var(--line-strong); background: var(--parchment); color: var(--ink);
    font-size: 18px; line-height: 1; cursor: pointer; box-shadow: 0 1px 3px #2b211844;
  }
  .zoom button:hover { background: var(--gold-light); }
  .zoom button:focus-visible { outline: 3px solid var(--gold); outline-offset: 1px; }
</style>
