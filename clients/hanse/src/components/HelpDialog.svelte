<script lang="ts">
  // Spielanleitung: Ablauf eines Jahres, Handel, Schiffe, Kontore, Börse, Ereignisse, Gefechte, Ränge.
  // Werte spiegeln server/internal/hanse/engine (data.go, economy.go).
  import Modal from './Modal.svelte';
  import { CITIES, GOODS, GUILDS, loadNachweise, title, type Bildnachweis } from '../lib/format';

  let { open = $bindable(false) }: { open?: boolean } = $props();
  let nachweise = $state<Record<string, Bildnachweis>>({});
  $effect(() => {
    if (open) loadNachweise().then((n) => (nachweise = n));
  });
</script>

<Modal bind:open title="Wie spielt man Hanse?" wide>
  <div class="hilfe">
    <section>
      <h3 class="label">Ziel</h3>
      <p>Anno 1356 gründen Sie ein Handelshaus in Lübeck. Mit Salz, Schiffen und Kontoren in den Hansestädten
        mehren Sie Vermögen und Ansehen (<em>Index</em>). Steigt der Index über Rang · 100 plus etwas Glück, werden Sie
        befördert – vom {title(1, false)} bis zum {title(11, false)}. Wer in den Senat berufen wird, gewinnt.
        Jede Beförderung schenkt Lebensjahre; ohne Aufstieg stirbt der Herr des Hauses nach etwa zwölf Jahren.</p>
    </section>
    <section>
      <h3 class="label">Ablauf eines Jahres</h3>
      <ol>
        <li><strong>Eingetroffen:</strong> Die ausgesandten Schiffe kehren mit Waren zurück; Seegang und Verschleiß wirken.</li>
        <li><strong>Handel:</strong> Waren verkaufen, Salz kaufen, Börse, Schiffe und Kontore – beliebig oft.</li>
        <li><strong>Ereignis:</strong> Glück oder Pech des Jahres, manchmal mit einer Ja/Nein-Frage.</li>
        <li><strong>Abschluss:</strong> Händel mit einem Mitspieler anzetteln oder das Jahr beenden.</li>
      </ol>
      <p>Alle Häuser spielen gleichzeitig; das Jahr endet, wenn alle fertig sind.</p>
    </section>
    <section>
      <h3 class="label">Salz und Schiffe</h3>
      <p>Jedes Schiff fährt mit einer Ladung von 100 Salz in eine Stadt, in der Sie ein Kontor haben, und zahlt dort Heuer.
        Im nächsten Jahr bringt es die Ware der Stadt nach Lübeck. Je mehr Schiffe und Speicher in einer Stadt, desto mehr
        Ware – Speicher und Schiffe verstärken sich gegenseitig. Rumpf und Takelage nutzen sich mit Entfernung und Seegang ab;
        unter 0 % sinkt ein Schiff.</p>
      <table>
        <thead><tr><th>Stadt</th><th>Ware</th><th>Entfernung</th></tr></thead>
        <tbody>
          {#each CITIES as c, i (i)}{#if i > 0}<tr><td>{c.name}</td><td>{GOODS[c.good]}</td><td class="num">{c.dist}</td></tr>{/if}{/each}
        </tbody>
      </table>
    </section>
    <section>
      <h3 class="label">Lager, Börse, Schulden</h3>
      <p>Lübeck fasst je Speicher 900 Einheiten jeder Ware; der Rest verdirbt am Jahresende („überfüllt“).
        An der Börse kaufen Sie Anteile anderer Häuser (Preis Index · 10); jedes Jahr fließen 30 % des Index je Anteil als
        Dividende. Wer am Jahresende Schulden hat, muss ein Jahr in den Schuldturm.</p>
    </section>
    <section>
      <h3 class="label">Händel und Dänen</h3>
      <p>Wo Ihre Schiffe liegen und ein Mitspieler ein Kontor hat, können Sie seine Handelsflotte oder sein Kontor
        angreifen. Landkanonen schützen Kontore. Waldemar IV. von Dänemark greift gelegentlich selbst an; seine Flotte
        wächst mit jeder Beförderung.</p>
    </section>
    <section>
      <h3 class="label">Ränge</h3>
      <ol class="raenge">
        {#each GUILDS as g, r (r)}{#if r > 0}<li><strong>{title(r, false)}</strong> – Aufnahme {g}</li>{/if}{/each}
      </ol>
    </section>
    <section>
      <h3 class="label">Bildnachweis</h3>
      <p class="quelle">Hausmarken und Stadtwappen von Wikimedia Commons, unverändert, unter ihren eigenen Lizenzen.
        Küstenlinien: Natural Earth (gemeinfrei).</p>
      <ul class="nachweise">
        {#each Object.entries(nachweise) as [k, n] (k)}
          <li>{k.startsWith('staedte/') ? CITIES[Number(k.slice(8))]?.name : `Hausmarke ${Number(k.slice(11)) + 1}`}:
            <a href={n.seite} target="_blank" rel="noopener">{n.datei}</a> · {n.urheber} ·
            {#if n.lizenzUrl}<a href={n.lizenzUrl} target="_blank" rel="noopener">{n.lizenz}</a>{:else}{n.lizenz}{/if}</li>
        {/each}
      </ul>
    </section>
    <p class="quelle">Nachbau des C64-Spiels „Hanse“ als Open-Source-Remake, mit eigenem Programmcode.</p>
  </div>
</Modal>

<style>
  .hilfe { display: flex; flex-direction: column; gap: 10px; max-height: 70vh; overflow-y: auto; }
  section p, li { line-height: 1.4; }
  table { border-collapse: collapse; }
  td, th { padding: 2px 10px 2px 0; text-align: left; }
  th { font-family: var(--font-label); font-size: 13px; font-weight: 400; color: var(--sepia); }
  .quelle { font-size: 15px; color: var(--sepia); }
  .nachweise { font-size: 15px; padding-left: 18px; margin: 4px 0 0; }
</style>
