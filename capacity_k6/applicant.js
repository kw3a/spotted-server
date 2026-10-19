import http from 'k6/http';
import exec from 'k6/execution';
import { Trend, Counter } from 'k6/metrics';
import {
  BASE, tag, think, branch, randint, pick, register, login, logout, verify,
  redirectIs, header, journeysCompleted, authenticatedPage,
} from './shared.js';
import { OFFERS, SEARCH_TERMS, readingOffers } from './data.js';
import { attemptsFor, finishedVerdict, verdictMatches } from './solutions.js';
import {
  MODE, RUN_ID, P_FILTER, P_APPLY, P_FINISH, ENABLE_KEYSTROKES,
  RESULTS_TIMEOUT, SUBMISSIONS_PER_PROBLEM,
} from './config.js';

export const judgeE2E = new Trend('judge_e2e', true);
export const submissionsVerified = new Counter('submissions_verified');
export const problemsSolved = new Counter('problems_solved');
// Una identidad independiente por VU, renovada al consumir sus diez quizzes.
let user = null;
let usedQuizzes = 0;
let generation = 0;

function session() {
  if (!user || usedQuizzes === OFFERS.length) {
    const nick = `ck_${RUN_ID}_${exec.vu.idInTest}_${generation++}`;
    if (nick.length > 32) exec.test.abort('Demasiadas generaciones de cuentas para el formato de nick');
    user = { nick, password: `cap_${RUN_ID}`, name: `Capacity run ${RUN_ID}` };
    usedQuizzes = 0;
    if (!register(user)) { user = null; return false; }
    return true;
  }
  return login(user.nick, user.password);
}
function keystrokes(quizID) {
  const r = http.post(`${BASE}/keystrokes`, {
    quizID, strokeAmount: 20, udMean: 90, udStdDev: 15,
    du1Mean: 120, du1StdDev: 20, du2Mean: 210, du2StdDev: 30,
    ddMean: 130, ddStdDev: 20, uuMean: 140, uuStdDev: 20,
  }, { ...tag('keystrokes'), timeout: '10s' });
  return verify(r, 'keystrokes');
}
function writingPause(quizID) {
  let remaining = 60 + Math.random() * 60;
  if (!ENABLE_KEYSTROKES) { think(remaining, remaining); return true; }
  while (remaining > 0) {
    const interval = Math.min(5, remaining);
    think(interval, interval);
    if (!keystrokes(quizID)) return false;
    remaining -= interval;
  }
  return true;
}
function solve(offer, problem) {
  const info = http.get(`${BASE}/problems?problemID=${problem.problemID}`, { ...tag('problem_read'), timeout: '10s' });
  if (!verify(info, 'problem_read', (r) => r.body.includes(problem.title))) return false;
  const examples = http.get(`${BASE}/examples?problemID=${problem.problemID}`, { ...tag('examples'), timeout: '10s' });
  if (!verify(examples, 'examples', (r) => (r.body.match(/ENTRADA/g) || []).length === 2 && (r.body.match(/SALIDA/g) || []).length === 2)) return false;
  let bestAccepted = 0;
  for (const solution of attemptsFor(problem, SUBMISSIONS_PER_PROBLEM)) {
    const tags = { solution_type: solution.kind, language: problem.languageID === '54' ? 'cpp' : 'java' };
    if (!writingPause(offer.quizID)) return false;
    const started = Date.now();
    const submitted = http.post(`${BASE}/submissions`, {
      quizID: offer.quizID, problemID: problem.problemID,
      languageID: problem.languageID, src: solution.source,
    }, { ...tag('submit', tags), timeout: '65s' });
    const match = (submitted.body || '').match(/sse-connect="\/results\/([0-9a-f]{8}-(?:[0-9a-f]{4}-){3}[0-9a-f]{12})"/i);
    if (!verify(submitted, 'submit', () => Boolean(match), 200, tags)) return false;
    const results = http.get(`${BASE}/results/${match[1]}`, {
      ...tag('results', tags), headers: { Accept: 'text/event-stream' }, timeout: RESULTS_TIMEOUT,
    });
    judgeE2E.add(Date.now() - started, tags);
    if (!verify(results, 'results', (r) => header(r, 'Content-Type').includes('text/event-stream')
      && verdictMatches(finishedVerdict(r.body), solution.expected, problem.testCaseCount), 200, tags)) return false;
    bestAccepted = Math.max(bestAccepted, solution.expected.accepted);
    const score = http.get(`${BASE}/score?problemID=${problem.problemID}`, { ...tag('score', tags), timeout: '10s' });
    // /score muestra la mejor submission, no necesariamente el ultimo intento.
    if (!verify(score, 'score', (r) => new RegExp(`\\b${bestAccepted}\\s*/\\s*${problem.testCaseCount}\\b`).test(r.body), 200, tags)) return false;
    submissionsVerified.add(1, tags);
  }
  problemsSolved.add(1);
  return true;
}
export function runApplicant() {
  if (!session()) { think(2, 5); return; }
  let loggedOut = false;
  try {
    think(2, 5);
    const home = http.get(`${BASE}/`, { ...tag('search'), timeout: '10s' });
    if (!verify(home, 'search', (r) => authenticatedPage(r) && r.body.includes('id="search-field"'))) return;
    think(2, 4);
    if (branch(P_FILTER)) {
      const n = randint(1, 3);
      for (let i = 0; i < n; i++) {
        const r = http.get(`${BASE}/?q=${encodeURIComponent(pick(SEARCH_TERMS))}`, { ...tag('search'), timeout: '10s' });
        if (!verify(r, 'search', (x) => authenticatedPage(x) && x.body.includes('id="search-field"'))) return;
        think(2, 4);
      }
    }
    const offer = OFFERS[(exec.vu.idInTest - 1 + usedQuizzes) % OFFERS.length];
    for (const o of readingOffers(offer)) {
      const r = http.get(`${BASE}/preamble/${o.offerID}`, { ...tag('offer_read'), timeout: '10s' });
      if (!verify(r, 'offer_read', (x) => authenticatedPage(x) && x.body.includes('id="confirm-modal"'))) return;
      think(10, 30);
    }
    if (MODE === 'full' && branch(P_APPLY)) {
      // Consumir antes de participar evita reintentar una pareja ya insertada.
      usedQuizzes++;
      const part = http.post(`${BASE}/participate`, { quizID: offer.quizID }, { ...tag('participate'), timeout: '10s' });
      if (!verify(part, 'participate', (r) => redirectIs(r, `/quizes/${offer.quizID}`))) return;
      const page = http.get(`${BASE}/quizes/${offer.quizID}`, { ...tag('quiz_page'), timeout: '10s' });
      if (!verify(page, 'quiz_page', (r) => r.body.includes(`data-quiz_id="${offer.quizID}"`)
        && offer.problems.every((p) => r.body.includes(`data-problem_id="${p.problemID}"`)))) return;
      think(5, 10);
      for (const problem of offer.problems) if (!solve(offer, problem)) return;
      if (branch(P_FINISH)) {
        const r = http.post(`${BASE}/end`, { quizID: offer.quizID }, { ...tag('quiz_end'), timeout: '10s' });
        if (!verify(r, 'quiz_end', (x) => redirectIs(x, `/preamble/${offer.offerID}`))) return;
      }
    }
    think(2, 5);
    loggedOut = true;
    if (logout()) journeysCompleted.add(1);
  } finally {
    if (!loggedOut) logout();
  }
}
