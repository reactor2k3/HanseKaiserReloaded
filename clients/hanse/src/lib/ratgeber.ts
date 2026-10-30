// Ratgeber („Der Ältermann rät“): leitet aus dem eigenen Stand ab, was jetzt zu tun ist.
// Der Ältermann stand einem Hansekontor bzw. einer Fahrergesellschaft vor und beriet die Kaufleute.
// Reine Funktion ohne Zustand; Schwellen spiegeln die Regeln in engine/economy.go und trade.go.
import type { SelfView } from './types';
import { CARGO, CITIES, GOODS, STORE_CAPACITY, num } from './format';

export interface Tip {
  /** muss = sonst droht Schaden, rat = lohnt sich, info = was der Abschnitt verlangt */
  level: 'muss' | 'rat' | 'info';
  text: string;
}

// Was jeder Abschnitt des Jahres vom Kaufmann will (erste Zeile des Ratgebers)
const PHASE_TODO: Record<string, string> = {
  trade: 'Waren verkaufen, Salz kaufen, Schiffe pflegen und aussenden, dann die Geschäfte abschließen.',
  event: 'Entscheiden Sie über das Ereignis des Jahres.',
  feud: 'Zetteln Sie Händel mit einem Mitspieler an oder beenden Sie das Jahr.',
  done: 'Ihr Jahr ist abgeschlossen. Warten Sie auf die übrigen Häuser.',
};

/** Ratschläge für den aktuellen Abschnitt, wichtigste zuerst. */
export function advise(self: SelfView): Tip[] {
  const p = self.player;
  const y = p.year;
  const tips: Tip[] = [];
  const todo = PHASE_TODO[self.phase];
  if (todo) tips.push({ level: 'info', text: todo });
  if (self.phase !== 'trade') return tips;

  // Überfüllte Waren werden zum Jahresende auf die Lübecker Kapazität gekappt (capStock)
  const cap = STORE_CAPACITY * p.stores[1];
  for (let w = 1; w <= 6; w++) {
    if (p.stock[w] > cap) {
      tips.push({ level: 'muss', text: `Ihr Lager fasst nur ${num(cap)} ${GOODS[w]}, Sie haben ${num(p.stock[w])}. Verkaufen Sie oder kaufen Sie einen Speicher in Lübeck, sonst verdirbt der Rest.` });
    }
  }
  if (p.mark < 0) tips.push({ level: 'muss', text: 'Ihr Konto ist im Minus und kostet Zinsen. Verkaufen Sie Waren, bevor Sie Neues kaufen.' });

  // Schiffe: Zustand, Ladung, Aussendung
  if (p.ships + p.underway > 0 && (p.hull < 40 || p.rigging < 40)) {
    tips.push({ level: 'muss', text: `Ihre Flotte ist in schlechtem Zustand (Rumpf ${p.hull} %, Takelage ${p.rigging} %). Lassen Sie sie unter „Schiffe“ ausbessern, sonst sinken Schiffe.` });
  }
  const kontore = CITIES.map((_, i) => i).filter((i) => i > 1 && p.kontor[i]);
  if (p.ships > 0 && !kontore.length) {
    const billig = CITIES.map((_, i) => i).filter((i) => i > 1).sort((a, b) => y.kontorCost[a] - y.kontorCost[b])[0];
    if (billig && y.kontorCost[billig] <= p.mark) {
      tips.push({ level: 'rat', text: `Ohne Kontor fahren Ihre Schiffe nirgends hin. In ${CITIES[billig].name} kostet eines ${num(y.kontorCost[billig])} Mark.` });
    }
  }
  if (p.ships > 0 && kontore.length) {
    const salz = p.ships * CARGO;
    if (p.stock[1] < salz) {
      tips.push({ level: 'rat', text: `Für ${p.ships} Schiff${p.ships === 1 ? '' : 'e'} brauchen Sie ${num(salz)} Salz als Ladung, im Lager sind ${num(p.stock[1])}.` });
    } else {
      tips.push({ level: 'rat', text: `${p.ships} Schiff${p.ships === 1 ? ' liegt' : 'e liegen'} beladen im Hafen. Senden Sie sie unter „Kontore“ nach ${kontore.map((i) => CITIES[i].name).join(' oder ')}.` });
    }
  }
  // Wachstum: Neubau lohnt, wenn die Kasse mehrere Schiffe trägt
  if (p.orders === 0 && p.mark > 3 * y.buildPrice) {
    tips.push({ level: 'rat', text: `Die Kasse ist gut gefüllt. Ein neues Schiff kostet ${num(y.buildPrice)} Mark und bringt mehr Ware heim.` });
  }
  return tips;
}
