const parameters = globalThis.__CAPACITY_CONFIG;
if (!parameters) throw new Error('Falta configuracion del lanzador: usar ./run.sh run con flags');
export function numberParam(name, fallback, min, max, integer = false) {
  const raw = String(parameters[name] === undefined ? fallback : parameters[name]);
  const value = Number(raw);
  if (raw.trim() === '' || !Number.isFinite(value) || value < min || value > max || (integer && !Number.isInteger(value))) {
    throw new Error(`${name} debe ser ${integer ? 'un entero' : 'un numero'} entre ${min} y ${max}`);
  }
  return value;
}
function flag(name, fallback = false) {
  const value = parameters[name] === undefined ? fallback : parameters[name];
  if (typeof value !== 'boolean') throw new Error(`${name} debe ser booleano`);
  return value;
}
function duration(name, fallback) {
  const value = parameters[name] || fallback;
  if (!/^(?:\d+(?:\.\d+)?(?:ms|s|m|h))+$/.test(value) || !/[1-9]/.test(value)) {
    throw new Error(`${name} debe ser una duracion positiva de k6`);
  }
  return value;
}
export const MODE = parameters.TEST_MODE || 'full';
if (!['browse', 'full'].includes(MODE)) throw new Error('--mode debe ser browse o full');
export const SMOKE = flag('SMOKE');
export const MAX_VUS = numberParam('MAX_VUS', 110, 1, 100000, true);
export const RAMP = duration('RAMP', '2m');
export const HOLD = duration('HOLD', '5m');
export const RAMP_DOWN = duration('RAMP_DOWN', '30s');
export const GRACE = duration('GRACE', MODE === 'full' ? '12m' : '30s');
export const THINK_SCALE = numberParam('THINK_SCALE', 1, 0, 100);
export const P_FILTER = numberParam('P_FILTER', 0.6, 0, 1);
export const P_APPLY = numberParam('P_APPLY', 1, 0, 1);
export const P_FINISH = numberParam('P_FINISH', 1, 0, 1);
export const SUBMISSIONS_PER_PROBLEM = numberParam('SUBMISSIONS_PER_PROBLEM', 3, 3, 30, true);
export const ENABLE_RESULTS = flag('ENABLE_RESULTS', MODE === 'full');
export const ENABLE_KEYSTROKES = flag('ENABLE_KEYSTROKES', MODE === 'full');
export const RESULTS_TIMEOUT = duration('RESULTS_TIMEOUT', '65s');
export const RESULTS_P95_MS = numberParam('RESULTS_P95_MS', 60000, 1, 3600000);
export const BASE = (parameters.BASE_URL || '').replace(/\/$/, '');
if (!/^https:\/\/(?:\[[0-9a-f:]+\]|[^/?#:@]+)(?::\d+)?$/i.test(BASE)) {
  throw new Error('BASE_URL es obligatoria y debe apuntar por HTTPS al servidor externo');
}
const hostname = BASE.match(/^https:\/\/(\[[^\]]+\]|[^:]+)(?::\d+)?$/i)[1].toLowerCase().replace(/\.$/, '');
if (hostname === 'localhost' || hostname.endsWith('.localhost') || /^127(?:\.|$)/.test(hostname)
  || ['0.0.0.0', '[::]', '[::1]'].includes(hostname)) {
  throw new Error('Las pruebas de capacidad requieren un servidor externo, no loopback');
}
export const RUN_ID = parameters.RUN_ID || '';
if (!/^[a-f0-9]{16}$/.test(RUN_ID)) throw new Error('RUN_ID invalido: ejecutar con ./run.sh run -- flags-de-k6');
if (MODE === 'full' && (P_APPLY === 0 || P_FINISH === 0 || !ENABLE_RESULTS)) {
  throw new Error('full requiere participacion, finalizacion y --enable-results=true para comprobar las soluciones');
}
if (MODE === 'browse' && (ENABLE_RESULTS || ENABLE_KEYSTROKES)) throw new Error('SSE y pulsaciones requieren --mode full');

export const FIXTURES_FILE = parameters.FIXTURES_FILE || '';
