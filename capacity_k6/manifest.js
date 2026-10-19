export function validateManifest(manifest) {
  if (!manifest || manifest.version !== 4 || !Array.isArray(manifest.offers) || manifest.offers.length !== 10) {
    throw new Error('Las precondiciones requieren 10 ofertas de version 4');
  }
  const uuid = /^[0-9a-f]{8}-(?:[0-9a-f]{4}-){3}[0-9a-f]{12}$/i;
  const ids = new Set();
  const validateID = (id) => {
    if (typeof id !== 'string' || !uuid.test(id) || ids.has(id)) throw new Error('IDs de fixtures invalidos o duplicados');
    ids.add(id);
  };
  for (const offer of manifest.offers) {
    validateID(offer.offerID); validateID(offer.companyID); validateID(offer.quizID);
    if (!Array.isArray(offer.problems) || offer.problems.length !== 3) throw new Error('Cada quiz requiere 3 problemas');
    if (new Set(offer.problems.map((p) => p.languageID)).size !== 2) throw new Error('Cada quiz requiere C++ y Java');
    for (const p of offer.problems) {
      validateID(p.problemID);
      if (!['54', '62'].includes(p.languageID) || p.testCaseCount !== 5 || typeof p.source !== 'string' || !p.source.trim()) {
        throw new Error('Cada problema requiere C++ o Java, una solucion y 5 casos secretos');
      }
      if (!Array.isArray(p.solutions) || p.solutions.length !== 3) throw new Error('Cada problema requiere tres tipos de solucion');
      const kinds = new Set();
      for (const s of p.solutions) {
        const e = s && s.expected;
        if (!s || kinds.has(s.kind) || !['valid', 'partial', 'syntax_error'].includes(s.kind)
          || typeof s.source !== 'string' || !s.source.trim() || !e
          || !['accepted', 'wrongAnswer', 'errors', 'timeLimitExceeded'].every((k) => Number.isInteger(e[k]) && e[k] >= 0)
          || e.accepted + e.wrongAnswer + e.errors + e.timeLimitExceeded !== 5 || e.timeLimitExceeded !== 0) {
          throw new Error('Solucion o veredicto esperado invalido');
        }
        kinds.add(s.kind);
        if ((s.kind === 'valid' && (e.accepted !== 5 || s.source !== p.source))
          || (s.kind === 'partial' && !(e.accepted > 0 && e.accepted < 5 && e.wrongAnswer === 5 - e.accepted))
          || (s.kind === 'syntax_error' && e.errors !== 5)) throw new Error('Veredicto incompatible con el tipo de solucion');
      }
    }
  }
  return manifest.offers;
}
