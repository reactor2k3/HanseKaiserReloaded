<script lang="ts">
  // Aktionsspalte: Hauptmenü des Originals (Handel, Börse, Schiffe, Kontore), danach Ereignis und
  // Abschluss (Händel). Alle Prüfungen spiegeln engine/trade.go, damit Knöpfe nur bei Erfolg aktiv sind.
  import { app } from '../lib/state.svelte';
  import { act } from '../lib/net';
  import { CITIES, GOODS, MAX_CANNONS, MAX_STORES, CARGO, num, title } from '../lib/format';

  let { city = $bindable(6), onEvent }: { city?: number; onEvent: () => void } = $props();

  const self = $derived(app.self!);
  const p = $derived(self.player);
  const y = $derived(p.year);
  let tab = $state<'handel' | 'boerse' | 'schiffe' | 'kontore'>('handel');

  // Mengen-Eingaben je Bereich
  let saltQty = $state(100);
  let goodQty = $state<Record<number, number>>({});
  let shareQty = $state<Record<number, number>>({});
  let rigQty = $state(0), hullQty = $state(0), orderQty = $state(1), sellShipQty = $state(1);
  let sendQty = $state(1), storeQty = $state(1), cannonQty = $state(1);
  // Händel
  let target = $state(-1), feudCity = $state(0), mode = $state<'kontor' | 'flotte'>('kontor'), focus = $state('schiffe');

  const others = $derived(app.players.filter((o) => o.seat !== app.seat && !o.out));
  const rigMax = $derived(Math.min(100 - p.rigging, Math.floor(p.mark / Math.max(1, y.rigCost))));
  const hullMax = $derived(Math.min(100 - p.hull, Math.floor(p.mark / Math.max(1, y.hullCost))));
  const sendMax = $derived(city > 1 && p.kontor[city] ? Math.min(p.ships, Math.floor(p.stock[1] / CARGO), Math.floor(p.mark / Math.max(1, y.heuer[city]))) : 0);
  // Mögliche Händel-Ziele: Städte, in denen ich diesmal Schiffe habe und der Gegner ein Kontor
  const feudCities = $derived(target < 0 ? [] : CITIES.map((_, c) => c).filter((c) => c > 1 && p.sent[c] > 0 && app.players[target]?.kontore?.includes(c)));

  function go(type: string, extra: Record<string, unknown> = {}) {
    act({ type, ...extra });
  }
</script>

<section class="card panel" aria-labelledby="aktion-titel">
  {#if p.year.skipped}
    <h2 id="aktion-titel" class="display">Schuldturm</h2>
    <p>Sie sitzen dieses Jahr im Schuldturm. Ihre Geschäfte ruhen, bis das nächste Jahr beginnt.</p>
  {:else if self.phase === 'trade'}
    <h2 id="aktion-titel" class="display">Lübecker Kontor</h2>
    <div class="tabs" role="tablist" aria-label="Bereiche">
      <button role="tab" type="button" aria-selected={tab === 'handel'} class:active={tab === 'handel'} onclick={() => (tab = 'handel')}>Handel</button>
      <button role="tab" type="button" aria-selected={tab === 'boerse'} class:active={tab === 'boerse'} onclick={() => (tab = 'boerse')}>Börse</button>
      <button role="tab" type="button" aria-selected={tab === 'schiffe'} class:active={tab === 'schiffe'} onclick={() => (tab = 'schiffe')}>Schiffe</button>
      <button role="tab" type="button" aria-selected={tab === 'kontore'} class:active={tab === 'kontore'} onclick={() => (tab = 'kontore')}>Kontore</button>
    </div>

    <div class="inhalt">
      {#if tab === 'handel'}
        <table class="tab">
          <thead><tr><th>Ware</th><th class="r">Lager</th><th class="r">Preis</th><th>Menge</th><th></th></tr></thead>
          <tbody>
            {#each [2, 3, 4, 5, 6] as w (w)}
              <tr class:voll={y.overfull?.[w]}>
                <td>{GOODS[w]}</td>
                <td class="r num">{num(p.stock[w])}</td>
                <td class="r num">{y.prices[w]}</td>
                <td><input class="menge" type="number" min="0" max={p.stock[w]} bind:value={goodQty[w]} placeholder="alle" aria-label="Menge {GOODS[w]}" /></td>
                <td><button class="btn sm" type="button" disabled={!p.stock[w] || (goodQty[w] ?? 0) > p.stock[w]}
                  onclick={() => go('sellGood', { good: w, qty: goodQty[w] || 0 })}>Verkaufen</button></td>
              </tr>
            {/each}
          </tbody>
        </table>
        <button class="btn" type="button" disabled={[2, 3, 4, 5, 6].every((w) => !p.stock[w])} onclick={() => go('sellAll')}>Alles verkaufen</button>
        <div class="zeile">
          <span><strong>Salz</strong> · Lager {num(p.stock[1])} · Preis {y.saltPrice} Mark</span>
        </div>
        <div class="zeile">
          <label class="feld kompakt"><input type="number" min="1" bind:value={saltQty} placeholder=" " /><span class="feld-label">Salz</span></label>
          <button class="btn sm" type="button" disabled={!saltQty || saltQty * y.saltPrice > p.mark} onclick={() => go('buySalt', { qty: saltQty })}>Salz kaufen ({num(saltQty * y.saltPrice)} Mark)</button>
        </div>
        {#if p.ships * CARGO > p.stock[1]}
          <button class="btn sm" type="button" disabled={(p.ships * CARGO - p.stock[1]) * y.saltPrice > p.mark}
            onclick={() => go('buySalt', { qty: p.ships * CARGO - p.stock[1] })}>Ladung für alle {p.ships} Schiffe ({num(p.ships * CARGO - p.stock[1])})</button>
        {/if}
      {:else if tab === 'boerse'}
        {#if !self.exchange}
          <p>Die Börse ist im ersten Jahr geschlossen.</p>
        {:else}
          <p class="hint">Preis je Anteil = Index · 10. Dividende 30 % des Index je Anteil und Jahr.
            Vorjahr: Dividende {num(y.dividend)} · Zinsen {num(y.interest)} Mark.</p>
          <table class="tab">
            <thead><tr><th>Haus</th><th class="r">Index</th><th class="r">Preis</th><th class="r">frei</th><th class="r">Ihre</th><th>Menge</th><th></th></tr></thead>
            <tbody>
              {#each others as o (o.seat)}
                <tr>
                  <td>{o.name}</td>
                  <td class="r num">{o.index}</td>
                  <td class="r num">{num(o.index * 10)}</td>
                  <td class="r num">{o.freeShares}</td>
                  <td class="r num">{p.shares[o.seat] ?? 0}</td>
                  <td><input class="menge" type="number" min="1" bind:value={shareQty[o.seat]} aria-label="Anteile {o.name}" /></td>
                  <td class="knoepfe">
                    <button class="btn sm" type="button" disabled={!shareQty[o.seat] || shareQty[o.seat] > o.freeShares || shareQty[o.seat] * o.index * 10 > p.mark}
                      onclick={() => go('buyShares', { target: o.seat, qty: shareQty[o.seat] })}>Kauf</button>
                    <button class="btn sm" type="button" disabled={!shareQty[o.seat] || shareQty[o.seat] > (p.shares[o.seat] ?? 0)}
                      onclick={() => go('sellShares', { target: o.seat, qty: shareQty[o.seat] })}>Verk.</button>
                  </td>
                </tr>
              {:else}
                <tr><td colspan="7">Keine anderen Handelshäuser.</td></tr>
              {/each}
            </tbody>
          </table>
        {/if}
      {:else if tab === 'schiffe'}
        <dl class="werte">
          <dt>Im Hafen</dt><dd class="num">{p.ships}</dd>
          <dt>Unterwegs</dt><dd class="num">{p.underway}</dd>
          <dt>Bauaufträge</dt><dd class="num">{p.orders}</dd>
          <dt>Takelage</dt><dd class="num">{p.rigging} %</dd>
          <dt>Rumpf</dt><dd class="num">{p.hull} %</dd>
          <dt>Schiffswert</dt><dd class="num">{num(y.shipValue)} Mark</dd>
        </dl>
        <div class="zeile">
          <label class="feld kompakt"><input type="number" min="1" max={rigMax} bind:value={rigQty} placeholder=" " /><span class="feld-label">Takelage %</span></label>
          <button class="btn sm" type="button" disabled={!rigQty || rigQty > rigMax} onclick={() => go('rigging', { qty: rigQty })}>Takelage trimmen ({num(y.rigCost)} je %)</button>
          <button class="btn sm" type="button" disabled={rigMax <= 0} onclick={() => go('rigging', { qty: rigMax })}>Max</button>
        </div>
        <div class="zeile">
          <label class="feld kompakt"><input type="number" min="1" max={hullMax} bind:value={hullQty} placeholder=" " /><span class="feld-label">Rumpf %</span></label>
          <button class="btn sm" type="button" disabled={!hullQty || hullQty > hullMax} onclick={() => go('hull', { qty: hullQty })}>Rumpf ausbessern ({num(y.hullCost)} je %)</button>
          <button class="btn sm" type="button" disabled={hullMax <= 0} onclick={() => go('hull', { qty: hullMax })}>Max</button>
        </div>
        <div class="zeile">
          <label class="feld kompakt"><input type="number" min="1" bind:value={orderQty} placeholder=" " /><span class="feld-label">Neubauten</span></label>
          <button class="btn sm" type="button" disabled={!orderQty || orderQty * y.buildPrice > p.mark} onclick={() => go('orderShips', { qty: orderQty })}>Bauauftrag ({num(y.buildPrice)} je Schiff)</button>
        </div>
        <div class="zeile">
          <label class="feld kompakt"><input type="number" min="1" max={p.ships} bind:value={sellShipQty} placeholder=" " /><span class="feld-label">Schiffe</span></label>
          <button class="btn sm" type="button" disabled={!sellShipQty || sellShipQty > p.ships} onclick={() => go('sellShips', { qty: sellShipQty })}>Schiffe verkaufen</button>
        </div>
      {:else}
        <p class="hint">Stadt auf der Karte oder hier wählen. Schiffe fahren nur in Städte mit Kontor; jedes braucht {CARGO} Salz.</p>
        <div class="staedte" role="radiogroup" aria-label="Stadt">
          {#each CITIES as c, i (i)}
            {#if i > 0}
              <button type="button" role="radio" aria-checked={city === i} class:active={city === i} class:kontor={p.kontor[i]} onclick={() => (city = i)}>{c.name}</button>
            {/if}
          {/each}
        </div>
        <img class="stadtwappen" src="{import.meta.env.BASE_URL}heraldik/staedte/{city}.svg" alt="Wappen von {CITIES[city].name}" width="44" height="52" />
        <dl class="werte">
          <dt>{CITIES[city].name}</dt><dd>{p.kontor[city] ? 'Kontor' : 'kein Kontor'} · {GOODS[CITIES[city].good]}</dd>
          <dt>Speicher</dt><dd class="num">{p.stores[city]} / {MAX_STORES} · {num(y.storePrice[city])} Mark</dd>
          {#if city > 1}
            <dt>Landkanonen</dt><dd class="num">{p.cannons[city]} / {MAX_CANNONS} · {num(y.cannonCost[city])} Mark</dd>
            <dt>Heuer je Schiff</dt><dd class="num">{num(y.heuer[city])} Mark</dd>
            <dt>Diesmal ausgesandt</dt><dd class="num">{p.sent[city]}</dd>
          {/if}
        </dl>
        {#if city > 1 && !p.kontor[city]}
          <button class="btn" type="button" disabled={y.kontorCost[city] > p.mark} onclick={() => go('openKontor', { city })}>Kontor eröffnen ({num(y.kontorCost[city])} Mark)</button>
        {/if}
        {#if city > 1 && p.kontor[city]}
          <div class="zeile">
            <label class="feld kompakt"><input type="number" min="1" max={sendMax} bind:value={sendQty} placeholder=" " /><span class="feld-label">Schiffe</span></label>
            <button class="btn sm" type="button" disabled={!sendQty || sendQty > sendMax} onclick={() => go('sendShips', { city, qty: sendQty })}>Schiffe senden</button>
            <button class="btn sm" type="button" disabled={sendMax <= 0} onclick={() => go('sendShips', { city, qty: sendMax })}>Alle ({sendMax})</button>
          </div>
        {/if}
        {#if p.kontor[city]}
          <div class="zeile">
            <label class="feld kompakt"><input type="number" min="1" bind:value={storeQty} placeholder=" " /><span class="feld-label">Speicher</span></label>
            <button class="btn sm" type="button" disabled={!storeQty || p.stores[city] + storeQty > MAX_STORES || storeQty * y.storePrice[city] > p.mark}
              onclick={() => go('buyStores', { city, qty: storeQty })}>Speicher kaufen</button>
            <button class="btn sm" type="button" disabled={!storeQty || storeQty > p.stores[city]}
              onclick={() => go('sellStores', { city, qty: storeQty })}>verkaufen (75 %)</button>
          </div>
        {/if}
        {#if city > 1 && p.kontor[city]}
          <div class="zeile">
            <label class="feld kompakt"><input type="number" min="1" bind:value={cannonQty} placeholder=" " /><span class="feld-label">Kanonen</span></label>
            <button class="btn sm" type="button" disabled={!cannonQty || p.cannons[city] + cannonQty > MAX_CANNONS || cannonQty * y.cannonCost[city] > p.mark}
              onclick={() => go('buyCannons', { city, qty: cannonQty })}>Landkanonen kaufen</button>
          </div>
        {/if}
      {/if}
    </div>
    <footer>
      <button class="btn btn-primary" type="button" onclick={() => go('next')}>Geschäfte abschließen</button>
    </footer>
  {:else if self.phase === 'event'}
    <!-- Das Ereignis selbst steht im Ereignis-Fenster (Modal); hier nur der Weg zurück dorthin -->
    <h2 id="aktion-titel" class="display">Ereignis</h2>
    <p class="hint">Eine Entscheidung steht an.</p>
    <button class="btn btn-primary" type="button" onclick={onEvent}>Ereignis zeigen</button>
  {:else if self.phase === 'feud'}
    <h2 id="aktion-titel" class="display">Abschluss</h2>
    {#if y.event}
      <button class="linkbtn" type="button" onclick={onEvent}>Ereignis des Jahres noch einmal zeigen</button>
    {/if}
    {#if self.canFeud}
      <h3 class="label">Händel anzetteln</h3>
      <p class="hint">Angreifen können Sie dort, wo Ihre Schiffe dieses Jahr liegen und der Gegner ein Kontor hat.</p>
      <div class="staedte">
        {#each others as o (o.seat)}
          <button type="button" class:active={target === o.seat} onclick={() => { target = o.seat; feudCity = 0; }}>{o.name}</button>
        {/each}
      </div>
      {#if target >= 0}
        <div class="staedte">
          {#each feudCities as c (c)}
            <button type="button" class:active={feudCity === c} onclick={() => (feudCity = c)}>{CITIES[c].name} ({p.sent[c]} Schiffe)</button>
          {:else}
            <span class="hint">Keine gemeinsame Stadt mit eigenen Schiffen.</span>
          {/each}
        </div>
      {/if}
      {#if feudCity}
        <div class="staedte">
          <button type="button" class:active={mode === 'kontor'} onclick={() => (mode = 'kontor')}>Kontor</button>
          <button type="button" class:active={mode === 'flotte'} onclick={() => (mode = 'flotte')}>Handelsflotte</button>
        </div>
        {#if app.settings?.battleMode === 'tactical' && mode === 'kontor'}
          <div class="staedte" aria-label="Ziel">
            {#each [['schiffe', 'Schiffe zuerst'], ['kanonen', 'Kanonen zuerst'], ['speicher', 'Speicher zuerst']] as [v, l] (v)}
              <button type="button" class:active={focus === v} onclick={() => (focus = v)}>{l}</button>
            {/each}
          </div>
        {/if}
        <button class="btn" type="button" onclick={() => go('attack', { target, city: feudCity, mode, stance: focus })}>Angreifen</button>
      {/if}
    {/if}
    <footer>
      <button class="btn btn-primary" type="button" onclick={() => go('endTurn')}>Jahr beenden</button>
    </footer>
  {:else}
    <h2 id="aktion-titel" class="display">Jahr abgeschlossen</h2>
    <p>Sobald alle Häuser fertig sind, beginnt das nächste Jahr.</p>
    {@const warten = app.players.filter((o) => !o.out && o.phase !== 'done' && o.seat !== app.seat)}
    {#if warten.length}<p class="hint">Es fehlen noch: {warten.map((o) => `${title(o.rank, o.female)} ${o.name}`).join(', ')}.</p>{/if}
  {/if}
</section>

<style>
  .panel { display: flex; flex-direction: column; gap: 8px; min-height: 0; overflow: hidden; height: 100%; }
  .inhalt { flex: 1; min-height: 0; overflow: auto; display: flex; flex-direction: column; gap: 8px; }
  .tabs { display: flex; gap: 4px; }
  .tabs button, .staedte button {
    border: 1px solid var(--line-strong); background: var(--parchment); padding: 4px 10px; cursor: pointer;
    font-family: var(--font-label); font-size: 14px; letter-spacing: .06em; color: var(--ink);
  }
  .tabs button.active, .staedte button.active { background: var(--blood); color: var(--parchment); border-color: var(--blood); }
  .staedte { display: flex; flex-wrap: wrap; gap: 4px; }
  .staedte button.kontor { font-weight: 700; }
  .tab { width: 100%; border-collapse: collapse; font-size: 17px; }
  .tab th { font-family: var(--font-label); font-size: 13px; text-align: left; color: var(--sepia); font-weight: 400; }
  .tab td, .tab th { padding: 2px 4px; border-bottom: 1px solid var(--line); }
  .tab .r { text-align: right; }
  .tab tr.voll td:first-child::after { content: ' (überfüllt)'; color: var(--warn); font-size: 14px; }
  .knoepfe { display: flex; gap: 4px; }
  .menge { width: 80px; padding: 3px 6px; border: 1px solid var(--line-strong); background: var(--parchment); font: inherit; }
  .zeile { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
  .btn.sm { padding: 3px 10px; font-size: 15px; min-height: 30px; }
  .stadtwappen { float: right; width: 44px; height: 52px; object-fit: contain; margin: 0 0 4px 8px; }
  .werte { display: grid; grid-template-columns: auto 1fr; gap: 2px 12px; margin: 0; font-size: 17px; }
  .werte dt { font-family: var(--font-label); font-size: 14px; color: var(--sepia); }
  .werte dd { margin: 0; }
  .hint { color: var(--sepia); font-size: 16px; margin: 0; }
  .linkbtn { align-self: flex-start; border: none; background: none; padding: 0; color: var(--blood); font-family: var(--font-label); font-size: 15px; text-decoration: underline; cursor: pointer; }
  /* Abschluss-Knopf immer unten in der Spalte */
  footer { margin-top: auto; display: flex; justify-content: flex-end; padding-top: 6px; border-top: 1px solid var(--line); }
  h3 { margin: 6px 0 0; }
</style>
