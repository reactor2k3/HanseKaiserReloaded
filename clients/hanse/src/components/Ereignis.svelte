<script lang="ts">
  // Ereignis des Jahres als Modal: Meldung, ggf. Ja/Nein-Frage oder Name des Kindes, danach der Ausgang.
  // Öffnet sich einmal je Jahr und Ereignis; über „Ereignis zeigen“ in der Aktionsspalte erneut.
  // Mit Autopilot entscheidet der Computer, das Fenster bleibt zu.
  import Modal from './Modal.svelte';
  import { app, autopilotAn } from '../lib/state.svelte';
  import { act } from '../lib/net';
  import { num } from '../lib/format';

  let { open = $bindable(false) }: { open?: boolean } = $props();

  const e = $derived(app.self?.player.year.event);
  const frage = $derived(app.self?.phase === 'event' && !!e?.ask);
  let gezeigt = '';
  let kind = $state('');

  // Neues Ereignis (je Jahr und Art) einmal von selbst öffnen; Hinweisfenster des Jahresbeginns zuerst
  $effect(() => {
    if (!e) return;
    const key = `${app.year}-${e.id}`;
    // Mit Autopilot gilt das Ereignis als gesehen, damit es beim Abschalten nicht nachträglich aufgeht
    if (autopilotAn()) { gezeigt = key; return; }
    if (key === gezeigt || app.notices.length) return;
    gezeigt = key;
    open = true;
  });
  // Autopilot eingeschaltet: Fenster schließen
  $effect(() => {
    if (autopilotAn()) open = false;
  });

  function antwort(ja: boolean) {
    act({ type: 'decide', stance: ja ? 'ja' : 'nein', name: kind.trim() });
  }
</script>

{#if e}
  <Modal bind:open title={e.id === 0 ? 'Ein ruhiges Jahr' : 'Ereignis'} variant="event">
    <div class="urkunde">
      {#each e.lines as l, i (i)}<p class="zeile">{l}</p>{/each}
      {#if e.result?.length}
        <div class="ausgang">
          {#each e.result as l, i (i)}<p>{l}</p>{/each}
        </div>
      {/if}
      {#if !frage && e.delta}
        <p class="fazit" class:minus={e.delta < 0}>{e.delta > 0 ? '+' : '−'}{num(Math.abs(e.delta))} Mark</p>
      {/if}
      {#if frage && e.ask === 'name'}
        <label class="feld"><input bind:value={kind} maxlength="12" placeholder="z. B. Gesche" /><span class="feld-label">Name des Kindes</span></label>
      {/if}
    </div>
    {#snippet actions()}
      {#if frage && e.ask === 'jn'}
        <button class="btn" type="button" onclick={() => antwort(false)}>Nein</button>
        <button class="btn btn-primary" type="button" onclick={() => antwort(true)}>Ja</button>
      {:else if frage && e.ask === 'name'}
        <button class="btn btn-primary" type="button" onclick={() => antwort(true)}>Taufen</button>
      {:else}
        <button class="btn btn-primary" type="button" onclick={() => (open = false)}>Zur Kenntnis genommen</button>
      {/if}
    {/snippet}
  </Modal>
{/if}

<style>
  .urkunde { display: flex; flex-direction: column; gap: 8px; text-align: center; }
  .zeile { margin: 0; font-size: 19px; }
  .ausgang { border-top: 1px solid var(--line); padding-top: 8px; }
  .ausgang p { margin: 0; font-size: 18px; font-weight: 600; }
  .fazit { margin: 0; font-family: var(--font-label); font-size: 18px; color: var(--good); }
  .fazit.minus { color: var(--blood); }
</style>
