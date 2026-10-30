<script lang="ts">
  // Profil-Dialog: Wahlspruch und Familienwappen (Hauszeichen oder eigenes hochgeladenes Bild).
  // Der Wahlspruch wird erst nach einer echten Änderung speicherbar; Uploads übernimmt der Server sofort.
  import { untrack } from 'svelte';
  import Modal from './Modal.svelte';
  import Wappen from './Wappen.svelte';
  import { app } from '../lib/state.svelte';
  import { act, upload } from '../lib/net';
  import { setSkin, skin } from '../lib/skin.svelte';
  import { HAUSMARKEN, loadNachweise, type Bildnachweis } from '../lib/format';

  let { open = $bindable(false) }: { open?: boolean } = $props();

  const p = $derived(app.self!.player);
  let motto = $state('');
  let busy = $state(false);
  let nachweise = $state<Record<string, Bildnachweis>>({});
  $effect(() => {
    if (open) loadNachweise().then((n) => (nachweise = n));
  });
  const gewaehlt = $derived(p.arms.kind === 'house' ? p.arms.variant ?? 0 : -1);
  const nachweis = $derived(gewaehlt >= 0 ? nachweise[`hausmarken/${gewaehlt}`] : undefined);
  // Hausmarke sofort übernehmen (eine Auswahl ist eine echte Änderung)
  function waehle(v: number) {
    if (v !== gewaehlt) act({ type: 'profile', motto: p.motto, arms: { kind: 'house', variant: v, media: '' } });
  }

  // Nur beim Öffnen den gespeicherten Stand übernehmen (spätere Server-Updates überschreiben keine Eingaben)
  $effect(() => {
    if (open) untrack(() => (motto = p.motto));
  });
  const changed = $derived(motto.trim() !== p.motto);

  async function onFile(e: Event) {
    const input = e.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;
    busy = true;
    await upload('arms', file);
    busy = false;
    input.value = '';
  }
</script>

<Modal bind:open title="Profil und Wappen">
  <div class="cols">
    <Wappen arms={p.arms} seat={app.seat} name={p.name} size={72} label="Ihr Familienwappen" />
    <div class="rechts">
      <label class="feld"><input bind:value={motto} maxlength="40" placeholder="z. B. Gott segne den Handel" /><span class="feld-label">Wahlspruch</span></label>
      <button class="btn" type="button" disabled={!changed} onclick={() => act({ type: 'profile', motto: motto.trim() })}>Wahlspruch speichern</button>
      <div class="marken" role="radiogroup" aria-label="Hausmarke wählen">
        <span class="label">Hausmarke</span>
        <div class="raster">
          {#each Array.from({ length: HAUSMARKEN }, (_, i) => i) as v (v)}
            <button type="button" role="radio" aria-checked={gewaehlt === v} class:on={gewaehlt === v} onclick={() => waehle(v)}
              aria-label="Hausmarke {v + 1}">
              <img src="{import.meta.env.BASE_URL}heraldik/hausmarken/{v}.svg" alt="" width="40" height="48" />
            </button>
          {/each}
        </div>
        {#if nachweis}
          <span class="hint">Bild: <a href={nachweis.seite} target="_blank" rel="noopener">{nachweis.datei}</a> · {nachweis.urheber} · {nachweis.lizenz}</span>
        {/if}
      </div>
      <div class="upload">
        <span class="label">Eigenes Familienwappen</span>
        <!-- Natives Dateifeld unsichtbar, der Knopf im Spieldesign löst es aus (Tastatur über das Label) -->
        <label class="btn datei" class:busy>
          <input type="file" accept="image/png,image/jpeg,image/webp,image/svg+xml" disabled={busy} onchange={onFile} />
          {busy ? 'Wird hochgeladen …' : 'Wappen hochladen …'}
        </label>
        <span class="hint">PNG, JPEG, WebP oder SVG, höchstens 1 MB</span>
      </div>
      <p class="hint">Darstellung:
        <button type="button" class="skin" class:on={skin.value === 'barock'} onclick={() => setSkin('barock')}>Pergament</button>
        <button type="button" class="skin" class:on={skin.value === 'c64'} onclick={() => setSkin('c64')}>C64</button>
      </p>
    </div>
  </div>
  {#snippet actions()}
    <button class="btn btn-primary" type="button" onclick={() => (open = false)}>Fertig</button>
  {/snippet}
</Modal>

<style>
  .cols { display: flex; gap: 16px; align-items: flex-start; flex-wrap: wrap; }
  .rechts { flex: 1; min-width: 240px; display: flex; flex-direction: column; gap: 10px; }
  .marken { display: flex; flex-direction: column; gap: 4px; }
  .raster { display: grid; grid-template-columns: repeat(4, 52px); gap: 6px; }
  .raster button { width: 52px; height: 60px; padding: 4px; border: 1px solid var(--line-strong); background: var(--parchment); cursor: pointer; }
  .raster button.on { border: 2px solid var(--blood); background: var(--gold-light); }
  .raster button:focus-visible { outline: 3px solid var(--gold); outline-offset: 2px; }
  .raster img { width: 100%; height: 100%; object-fit: contain; }
  .upload { display: flex; flex-direction: column; gap: 4px; align-items: flex-start; }
  .datei { position: relative; cursor: pointer; }
  .datei.busy { opacity: .6; cursor: progress; }
  .datei input { position: absolute; width: 1px; height: 1px; opacity: 0; overflow: hidden; }
  .datei:focus-within { outline: 3px solid var(--gold); outline-offset: 2px; }
  .hint { font-size: 15px; }
  .hint { color: var(--sepia); margin: 0; display: flex; gap: 6px; align-items: center; }
  /* Darstellungswahl wie auf der Startseite: kleine Schaltflächen im Schriftbild der Beschriftungen */
  .skin { border: 1px solid var(--line-strong); background: var(--parchment); color: var(--ink); padding: 4px 10px; cursor: pointer; font-family: var(--font-label); font-size: 14px; letter-spacing: .06em; }
  .skin.on { background: var(--ink); color: var(--gold-light); border-color: var(--ink); }
</style>
