import http from 'k6/http';
import exec from 'k6/execution';
import { runApplicant } from './applicant.js';
import { OFFERS } from './data.js';
import { BASE, tag, verify, setPhase } from './shared.js';
import { MODE, SMOKE, MAX_VUS, RAMP, HOLD, RAMP_DOWN, GRACE, ENABLE_KEYSTROKES, RESULTS_P95_MS, RUN_ID } from './config.js';

http.setResponseCallback(http.expectedStatuses(200, 201));
const scenario = SMOKE ? {
  executor: 'per-vu-iterations', exec: 'applicant', vus: MAX_VUS, iterations: 1,
  maxDuration: GRACE, gracefulStop: GRACE,
} : {
  executor: 'ramping-vus', exec: 'applicant', startVUs: 0,
  stages: [{ duration: RAMP, target: MAX_VUS }, { duration: HOLD, target: MAX_VUS }, { duration: RAMP_DOWN, target: 0 }],
  gracefulRampDown: GRACE, gracefulStop: GRACE,
};
const thresholds = {
  'http_req_failed{phase:load}': ['rate<0.05'],
  'functional_errors{phase:load}': ['rate<0.01'],
  journeys_completed: ['count>0'],
};
const latency = (action, ms, required = true) => {
  thresholds[`http_req_duration{action:${action},phase:load}`] = [`p(95)<${ms}`];
  thresholds[`functional_errors{action:${action},phase:load}`] = ['rate<0.01'];
  if (required) thresholds[`http_reqs{action:${action},phase:load}`] = ['count>0'];
};
['register', 'logout', 'search', 'offer_read'].forEach((a) => latency(a, 1500));
// En smoke cada VU se autentica con el registro; login aparece al repetir visitas.
latency('login', 1500, false);
if (MODE === 'full') {
  ['participate', 'quiz_page', 'quiz_end', 'score', 'problem_read', 'examples'].forEach((a) => latency(a, 1500));
  latency('submit', 8000);
  latency('results', RESULTS_P95_MS);
  thresholds.judge_e2e = [`p(95)<${RESULTS_P95_MS}`];
  thresholds.problems_solved = ['count>0'];
  for (const kind of ['valid', 'partial', 'syntax_error']) {
    thresholds[`submissions_verified{solution_type:${kind}}`] = ['count>0'];
    thresholds[`functional_errors{action:results,solution_type:${kind},phase:load}`] = ['rate<0.01'];
    thresholds[`judge_e2e{solution_type:${kind}}`] = [`p(95)<${RESULTS_P95_MS}`];
  }
  for (const language of ['cpp', 'java']) {
    thresholds[`submissions_verified{language:${language}}`] = ['count>0'];
    thresholds[`functional_errors{action:results,language:${language},phase:load}`] = ['rate<0.01'];
    thresholds[`judge_e2e{language:${language}}`] = [`p(95)<${RESULTS_P95_MS}`];
  }
  if (ENABLE_KEYSTROKES) latency('keystrokes', 1500);
}
export const options = {
  setupTimeout: '2m', scenarios: { applicant: scenario }, thresholds,
  summaryTrendStats: ['avg', 'med', 'p(90)', 'p(95)', 'p(99)', 'max'],
};
export function setup() {
  setPhase('preflight');
  console.log(`run=${RUN_ID}; mode=${MODE}; VUs=${MAX_VUS}; 10 quizzes precargados`);
  // Lecturas anonimas: el preflight no registra ni autentica usuarios base.
  for (const o of OFFERS) {
    const page = http.get(`${BASE}/preamble/${o.offerID}`, { ...tag('offer_read'), timeout: '10s' });
    if (!verify(page, 'offer_read', (r) => r.body.includes(o.companyID) && r.body.includes('id="confirm-modal"'))) {
      exec.test.abort('Preflight: oferta o quiz no disponible; verificar las precondiciones ya creadas y el servidor');
    }
  }
}
export function applicant() { runApplicant(); }
