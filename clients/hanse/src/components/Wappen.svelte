<script lang="ts">
  // Familienwappen eines Hauses: Hausmarke aus der Auswahl (Wikimedia Commons, Bildnachweis in der
  // Hilfe, scripts/heraldik/commons_hanse.py) oder hochgeladene Datei (vom Server gesäubert).
  // Bewusst keine eigenen Zeichnungen (Micha-Vorgabe). Nur <img>, nie eingebettetes SVG.
  import type { ArmsSpec } from '../lib/types';
  import { HAUSMARKEN } from '../lib/format';

  let { arms, seat, size = 56, label = 'Wappen' }:
    { arms?: ArmsSpec; seat: number; name?: string; size?: number; label?: string } = $props();

  const src = $derived(arms?.kind === 'upload' && arms.media ? `${import.meta.env.BASE_URL}media/${arms.media}`
    : `${import.meta.env.BASE_URL}heraldik/hausmarken/${(arms?.variant ?? seat) % HAUSMARKEN}.svg`);
</script>

<img {src} alt={label} width={size} height={size * 1.2} class="wappen" />

<style>
  .wappen { display: block; flex-shrink: 0; object-fit: contain; }
</style>
