import { SharedArray } from 'k6/data';
import { validateManifest } from './manifest.js';
import { FIXTURES_FILE } from './config.js';
import { randint } from './shared.js';

export const OFFERS = new SharedArray('capacity-preconditions', () => {
  if (!FIXTURES_FILE) throw new Error('Falta FIXTURES_FILE; generar precondiciones y usar el lanzador');
  return validateManifest(JSON.parse(open(FIXTURES_FILE)));
});
export const SEARCH_TERMS = ['algoritmos', 'subsecuencia', 'mochila', 'camino', 'capacidad'];
export function readsCount() {
  const r = Math.random();
  if (r < 0.25) return 1;
  if (r < 0.55) return 2;
  if (r < 0.8) return 3;
  if (r < 0.92) return 4;
  return 5;
}
export function readingOffers(requiredOffer) {
  const remaining = OFFERS.filter((o) => o.offerID !== requiredOffer.offerID);
  const selected = [requiredOffer];
  const n = readsCount();
  while (selected.length < n) selected.push(remaining.splice(randint(0, remaining.length - 1), 1)[0]);
  return selected;
}
