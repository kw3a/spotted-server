// Verifica las reglas del runner sin HTTP, MySQL ni Judge0.
import test from 'node:test';
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';

async function moduleFile(name, suffix = '') {
  const source = await readFile(new URL(`../${name}`, import.meta.url), 'utf8');
  return import(`data:text/javascript;base64,${Buffer.from(source + suffix).toString('base64')}`);
}
const { attemptsFor, finishedVerdict, verdictMatches } = await moduleFile('solutions.js');
const { validateManifest } = await moduleFile('manifest.js');

function fixtures() {
  let counter = 1;
  const id = () => `00000000-0000-4000-8000-${String(counter++).padStart(12, '0')}`;
  return { version: 4, offers: Array.from({ length: 10 }, (_, quiz) => ({
    offerID: id(), companyID: id(), quizID: id(),
    problems: Array.from({ length: 3 }, (_, index) => ({
      problemID: id(), languageID: (quiz + index) % 2 === 0 ? '54' : '62',
      source: 'valid code', testCaseCount: 5,
      solutions: [
        { kind: 'valid', source: 'valid code', expected: { accepted: 5, wrongAnswer: 0, errors: 0, timeLimitExceeded: 0 } },
        { kind: 'partial', source: 'partial code', expected: { accepted: 2, wrongAnswer: 3, errors: 0, timeLimitExceeded: 0 } },
        { kind: 'syntax_error', source: 'invalid code', expected: { accepted: 0, wrongAnswer: 0, errors: 5, timeLimitExceeded: 0 } },
      ],
    })),
  })) };
}

test('cada quiz mezcla C++ y Java y exige todas las variantes', () => {
  const manifest = fixtures();
  assert.equal(validateManifest(manifest).length, 10);
  const wrongLanguage = fixtures();
  wrongLanguage.offers[0].problems[0].languageID = '71';
  assert.throws(() => validateManifest(wrongLanguage));
  const oneLanguage = fixtures();
  oneLanguage.offers[0].problems.forEach((p) => { p.languageID = '54'; });
  assert.throws(() => validateManifest(oneLanguage));
  const missingVariant = fixtures();
  missingVariant.offers[0].problems[0].solutions.pop();
  assert.throws(() => validateManifest(missingVariant));
  const oldManifest = fixtures();
  oldManifest.version = 3;
  assert.throws(() => validateManifest(oldManifest));
});

test('los intentos incluyen rechazo de compilacion, parcial y aceptacion final', () => {
  const problem = fixtures().offers[0].problems[0];
  assert.deepEqual(attemptsFor(problem, 3).map((s) => s.kind), ['syntax_error', 'partial', 'valid']);
  assert.equal(attemptsFor(problem, 5).at(-1).kind, 'valid');
});

test('los errores de compilacion esperados son un resultado correcto del juez', () => {
  const expected = fixtures().offers[0].problems[0].solutions[2].expected;
  assert.equal(verdictMatches(finishedVerdict('event: finished\ndata: <p>En cola: 0, Error: 5</p>\n\n'), expected, 5), true);
  assert.equal(verdictMatches(finishedVerdict('event: finished\ndata: <p>En cola: 1, Error: 4</p>\n\n'), expected, 5), false);
  assert.equal(verdictMatches(finishedVerdict('event: finished\ndata: <p>En cola: 0, Aceptado: 5</p>\n\n'), expected, 5), false);
});

test('el veredicto parcial debe tener exactamente los aceptados y rechazados previstos', () => {
  const expected = fixtures().offers[0].problems[0].solutions[1].expected;
  assert.equal(verdictMatches(finishedVerdict('event: finished\ndata: <p>En cola: 0, Aceptado: 2, Respuesta equivocada: 3</p>\n\n'), expected, 5), true);
  assert.equal(verdictMatches(finishedVerdict('event: finished\ndata: <p>En cola: 0, Aceptado: 2, Error: 3</p>\n\n'), expected, 5), false);
  assert.equal(verdictMatches(finishedVerdict('event: update\ndata: <p>En cola: 0, Aceptado: 2, Respuesta equivocada: 3</p>\n\n'), expected, 5), false);
});

test('la carga requiere una URL externa por HTTPS', async () => {
  const cases = [undefined, 'http://192.168.1.50:42069', 'https://localhost', 'https://127.0.0.1',
    'https://[::1]', 'https://0.0.0.0', 'https://localhost.', 'https://127.1'];
  for (let i = 0; i < cases.length; i++) {
    globalThis.__CAPACITY_CONFIG = { RUN_ID: '0123456789abcdef', BASE_URL: cases[i] };
    await assert.rejects(moduleFile('config.js', `\n// invalid ${i}`));
  }
  globalThis.__CAPACITY_CONFIG = { RUN_ID: '0123456789abcdef', BASE_URL: 'https://192.168.1.50:443/' };
  assert.equal((await moduleFile('config.js', '\n// external')).BASE, 'https://192.168.1.50:443');
});
