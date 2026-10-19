import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Counter } from 'k6/metrics';
import { BASE, THINK_SCALE } from './config.js';

export { BASE };
export const functionalErrors = new Rate('functional_errors');
let phase = 'load';
export function setPhase(value) { phase = value; }
export const journeysCompleted = new Counter('journeys_completed');

export function think(minS, maxS) {
  sleep((minS + Math.random() * (maxS - minS)) * THINK_SCALE);
}
export function branch(p) { return Math.random() < p; }
export function randint(min, max) { return min + Math.floor(Math.random() * (max - min + 1)); }
export function pick(arr) { return arr[Math.floor(Math.random() * arr.length)]; }

// name evita una serie nueva por cada UUID en los resultados exportados.
export function tag(action, extraTags = {}) {
  return { tags: { action, name: action, phase, ...extraTags }, redirects: 0 };
}
export function header(response, name) {
  const key = Object.keys(response.headers || {}).find((k) => k.toLowerCase() === name.toLowerCase());
  return key ? response.headers[key] : '';
}
export function verify(response, action, predicate = () => true, status = 200, extraTags = {}) {
  const ok = check(response, { [`${action}_ok`]: (r) => r.status === status && !r.error && predicate(r) }, { action, ...extraTags });
  functionalErrors.add(!ok, { action, phase, ...extraTags });
  return ok;
}
export function authenticatedPage(response) {
  return typeof response.body === 'string' && response.body.includes('hx-post="/logout"');
}
export function redirectIs(response, path) { return header(response, 'HX-Redirect') === path; }
export function redirectID(response, prefix) {
  const path = header(response, 'HX-Redirect');
  const id = path.startsWith(prefix) ? path.slice(prefix.length) : '';
  return /^[0-9a-f]{8}-(?:[0-9a-f]{4}-){3}[0-9a-f]{12}$/i.test(id) ? id : '';
}
// Registro y login comparten validacion de la sesion; no se imprimen tokens.
function sessionCookies(r) {
  const cookies = r.cookies || {};
  const hasCookie = (name) => cookies[name] && cookies[name].some((c) => c.value);
  if (!hasCookie('access_token') || !hasCookie('refresh_token')) return false;
  return ['access_token', 'refresh_token'].every((name) => {
    const values = http.cookieJar().cookiesForURL(BASE)[name];
    return values && values.some((value) => value);
  });
}
export function register(user) {
  const r = http.post(`${BASE}/register`, user, { ...tag('register'), timeout: '15s' });
  return verify(r, 'register', (x) => Boolean(redirectID(x, '/profile/')) && sessionCookies(x), 201);
}
export function login(nick, password) {
  const r = http.post(`${BASE}/login`, { nick, password }, { ...tag('login'), timeout: '10s' });
  return verify(r, 'login', (x) => redirectIs(x, '/') && sessionCookies(x));
}
export function logout() {
  const r = http.post(`${BASE}/logout`, {}, { ...tag('logout'), timeout: '10s' });
  // handlerProfile.go configura el logout hacia '/', no '/login'.
  return verify(r, 'logout', (x) => redirectIs(x, '/'));
}
