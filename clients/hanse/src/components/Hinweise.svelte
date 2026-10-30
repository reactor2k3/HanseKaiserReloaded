<script lang="ts">
  // Hinweisfenster der Reihe nach: Ankunft der Waren zu Jahresbeginn, Gefechte, Urkunde bei
  // Beförderung, Erbfolge. Mit Autopilot öffnen sie sich nicht (alles steht in der Chronik).
  import Modal from './Modal.svelte';
  import { app, autopilotAn } from '../lib/state.svelte';
  import { CITIES, GOODS, SEA, num, title } from '../lib/format';

  const n = $derived(app.notices[0]);
  let open = $state(false);
  $effect(() => {
    if (autopilotAn()) {
      app.notices.length = 0;
      open = false;
      return;
    }
    if (n && !open) open = true;
  });
  function weiter() {
    app.notices.shift();
  }
  const me = $derived(app.self?.player);
  const name = (seat: number) => (seat < 0 ? 'Waldemar IV.' : seat === app.seat ? 'Sie' : app.players[seat]?.name ?? '?');
  const titel = $derived(!n ? '' : n.kind === 'arrivals' ? `Anno ${n.year}: Eingetroffen` : n.kind === 'battle' ? 'Seegefecht'
    : n.kind === 'promotion' ? 'Urkunde' : 'Erbfolge');
</script>

{#if n}
  <Modal bind:open title={titel} variant="event" onclose={weiter}>
    {#if n.kind === 'arrivals'}
      <p>Es war <strong>{SEA[n.sea]}</strong> See.</p>
      <table class="tab">
        <tbody>
          {#each CITIES as c, i (i)}
            {#if i > 1 && n.arrived[i] > 0}
              <tr><td>{c.name}</td><td>{GOODS[c.good]}</td><td class="r num">{num(n.arrived[i])}</td></tr>
            {/if}
          {/each}
        </tbody>
      </table>
      {#if n.arrived.every((x) => !x)}<p>Keine Waren eingetroffen.</p>{/if}
      {#if n.lost}<p class="warn">{n.lost} Schiff(e) gingen durch Verschleiß verloren.</p>{/if}
      {#if n.delivered}<p>{n.delivered} neue(s) Schiff(e) liefen vom Stapel.</p>{/if}
    {:else if n.kind === 'battle'}
      {@const b = n.battle}
      <p><strong>{name(b.attacker)}</strong> gegen {b.mode === 'flotte' ? 'die Handelsflotte' : 'das Kontor'} von <strong>{name(b.defender)}</strong> in {CITIES[b.city].name}.</p>
      <table class="tab">
        <thead><tr><th></th><th class="r">Einsatz</th><th class="r">Verlust</th></tr></thead>
        <tbody>
          <tr><td>Angreifer: Schiffe</td><td class="r">{b.attShips}</td><td class="r">{b.attLost}</td></tr>
          <tr><td>Verteidiger: Schiffe</td><td class="r">{b.defShips}</td><td class="r">{b.shipLost}</td></tr>
          {#if b.mode === 'kontor'}
            <tr><td>Landkanonen</td><td class="r">{b.cannons}</td><td class="r">{b.canLost}</td></tr>
            <tr><td>Speicher</td><td class="r">{b.stores}</td><td class="r">{b.stoLost}</td></tr>
          {/if}
        </tbody>
      </table>
      {#if b.attacker === app.seat && b.attLoot}<p>Erbeutet: {num(b.attLoot)} Mark.</p>{/if}
      {#if b.defender === app.seat && b.defLoot}<p>Erbeutet: {num(b.defLoot)} Mark.</p>{/if}
    {:else if n.kind === 'promotion' && me}
      <div class="urkunde">
        <p class="display gross">Urkunde</p>
        <p>{title(n.rank, me.female)} {me.name}</p>
        <p>wird hiermit das Privileg zuteil, {n.body} {n.verb} zu werden.</p>
        <p class="gez">Gez. der Senat, im Jahre {n.year}</p>
      </div>
    {:else if n.kind === 'died'}
      <p>Der Herr des Hauses ist verstorben. <strong>{n.heir}</strong> führt die Geschäfte weiter.</p>
    {/if}
    {#snippet actions()}
      <button class="btn btn-primary" type="button" onclick={() => (open = false)}>Zur Kenntnis genommen</button>
    {/snippet}
  </Modal>
{/if}

<style>
  .tab { width: 100%; border-collapse: collapse; }
  .tab td, .tab th { padding: 3px 6px; border-bottom: 1px solid var(--line); }
  .tab th { font-family: var(--font-label); font-size: 13px; font-weight: 400; color: var(--sepia); text-align: left; }
  .r { text-align: right; }
  .warn { color: var(--warn); }
  .urkunde { text-align: center; }
  .urkunde p { margin: 6px 0; font-size: 18px; }
  .urkunde .gross { font-size: 34px; }
  .urkunde .gez { font-style: italic; color: var(--sepia); }
</style>
