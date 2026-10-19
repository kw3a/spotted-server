// El ultimo intento resuelve el problema; los previos simulan correcciones.
export function attemptsFor(problem, count) {
  const byKind = Object.fromEntries(problem.solutions.map((s) => [s.kind, s]));
  const attempts = [];
  for (let i = 0; i < count - 1; i++) attempts.push(byKind[i % 2 === 0 ? 'syntax_error' : 'partial']);
  attempts.push(byKind.valid);
  return attempts;
}

export function finishedVerdict(body) {
  const blocks = String(body || '').replace(/\r\n/g, '\n').split('\n\n');
  const finished = blocks.filter((b) => /^event: finished$/m.test(b));
  if (finished.length !== 1) return null;
  const text = finished[0];
  const pending = text.match(/En cola:\s*(\d+)\b/);
  if (!pending) return null;
  const count = (label) => {
    const match = text.match(new RegExp(`${label}:\\s*(\\d+)\\b`));
    return match ? Number(match[1]) : 0;
  };
  return {
    pending: Number(pending[1]), accepted: count('Aceptado'),
    wrongAnswer: count('Respuesta equivocada'), errors: count('Error'),
    timeLimitExceeded: count('Tiempo límite excedido'),
  };
}
export function verdictMatches(actual, expected, total) {
  return actual !== null && actual.pending === 0
    && ['accepted', 'wrongAnswer', 'errors', 'timeLimitExceeded'].every((k) => actual[k] === expected[k])
    && actual.accepted + actual.wrongAnswer + actual.errors + actual.timeLimitExceeded === total;
}
