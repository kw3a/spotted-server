// Sorting and navigation use the data already returned with the page.
const highlighter = import('https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/es/highlight.min.js')
  .then(module => module.default).catch(() => null);
const initializedRoots = new WeakSet();

function initializeReview() {
  const root = document.getElementById('applicant-review');
  if (!root || initializedRoots.has(root)) return;
  initializedRoots.add(root);
  const list = root.querySelector('#candidate-list');
  const candidates = [...root.querySelectorAll('[data-candidate]')];
  const panels = [...root.querySelectorAll('[data-review]')];
  const sort = root.querySelector('#candidate-sort');
  const feedback = root.querySelector('#review-feedback');
  const reference = root.querySelector('#test-reference');
  const params = new URL(location.href).searchParams;
  const storageKey = `applicant-review:${root.dataset.offer}`;
  let saved = {};
  try { saved = JSON.parse(sessionStorage.getItem(storageKey) || '{}'); } catch { /* Storage is optional. */ }
  if (!saved || typeof saved !== 'object' || Array.isArray(saved)) saved = {};
  let selected = params.get('candidate') || candidates[0]?.dataset.candidate;
  let tab = params.get('tab') === 'code' ? 'code' : 'results';
  let problem = params.get('problem');
  let ordered = [];
  sort.value = ['score', 'name'].includes(params.get('sort')) ? params.get('sort') : 'original';

  function remember() {
    const url = new URL(location.href);
    url.searchParams.delete('q');
    url.searchParams.delete('filter');
    const values = { candidate: selected, tab, problem, sort: sort.value };
    for (const [key, value] of Object.entries(values)) {
      if (!value || value === 'original' || (key === 'tab' && value === 'results')) url.searchParams.delete(key);
      else url.searchParams.set(key, value);
    }
    history.replaceState(history.state, '', url);
    try { sessionStorage.setItem(storageKey, JSON.stringify(saved)); } catch { /* Navigation still works without storage. */ }
  }

  function updateContent(panel) {
    const selector = panel.querySelector('[data-problem-select]');
    if (![...selector.options].some(option => option.value === problem)) problem = selector.options[0]?.value || '';
    selector.value = problem;
    panel.querySelectorAll('[data-problem]').forEach(item => item.classList.toggle('hidden', item.dataset.problem !== problem));
    panel.querySelectorAll('[data-view]').forEach(view => view.classList.toggle('hidden', view.dataset.view !== tab));
    panel.querySelectorAll('[data-tab]').forEach(button => {
      const active = button.dataset.tab === tab;
      button.setAttribute('aria-selected', String(active));
      button.tabIndex = active ? 0 : -1;
    });
    if (tab === 'code') {
      const code = [...panel.querySelectorAll('[data-source]')].find(item => item.closest('[data-problem]').dataset.problem === problem);
      if (code && !code.dataset.prepared) {
        code.dataset.prepared = 'true';
        const numbers = code.closest('[data-problem]').querySelector('[data-line-numbers]');
        numbers.textContent = code.textContent.split('\n').map((_, index) => index + 1).join('\n');
        highlighter.then(hljs => { if (hljs && code.isConnected) hljs.highlightElement(code); });
      }
    }
  }

  function selectCandidate(id) {
    const button = ordered.find(candidate => candidate.dataset.candidate === id) || ordered[0];
    selected = button?.dataset.candidate || '';
    candidates.forEach(candidate => candidate.setAttribute('aria-pressed', String(candidate.dataset.candidate === selected)));
    panels.forEach(panel => {
      const active = panel.dataset.review === selected;
      panel.classList.toggle('hidden', !active);
      if (!active) return;
      updateContent(panel);
      const index = ordered.indexOf(button);
      panel.querySelector('[data-previous]').disabled = index <= 0;
      panel.querySelector('[data-next]').disabled = index >= ordered.length - 1;
      panel.querySelector('[data-review-scroll]').scrollTop = saved[selected] || 0;
    });
    root.querySelector('#review-empty').classList.toggle('hidden', Boolean(button));
    remember();
  }

  function applySort() {
    ordered = [...candidates];
    ordered.sort((a, b) => {
      if (sort.value === 'score') return Number(b.dataset.score) - Number(a.dataset.score) || Number(a.dataset.order) - Number(b.dataset.order);
      if (sort.value === 'name') return a.dataset.name.localeCompare(b.dataset.name);
      return Number(a.dataset.order) - Number(b.dataset.order);
    });
    ordered.forEach(candidate => list.appendChild(candidate));
    root.querySelector('#candidate-count').textContent = String(candidates.length);
    selectCandidate(selected);
  }

  function moveCandidate(delta) {
    const index = ordered.findIndex(candidate => candidate.dataset.candidate === selected);
    const next = ordered[index + delta];
    if (!next) return;
    selectCandidate(next.dataset.candidate);
    next.scrollIntoView({ block: 'nearest' });
  }

  root.addEventListener('click', async event => {
    const button = event.target.closest('button');
    if (!button) return;
    if (button.dataset.candidate) selectCandidate(button.dataset.candidate);
    if (button.hasAttribute('data-previous')) moveCandidate(-1);
    if (button.hasAttribute('data-next')) moveCandidate(1);
    if (button.dataset.tab) {
      tab = button.dataset.tab;
      const panel = panels.find(item => item.dataset.review === selected);
      updateContent(panel);
      remember();
    }
    if (button.hasAttribute('data-open-test')) reference.showModal();
    if (button.hasAttribute('data-close-test')) reference.close();
    if (button.hasAttribute('data-copy-code')) {
      const code = button.closest('[data-problem]').querySelector('[data-source]');
      try {
        await navigator.clipboard.writeText(code.textContent);
        button.textContent = 'Copiado';
        feedback.textContent = 'Código copiado al portapapeles.';
        setTimeout(() => { button.textContent = 'Copiar código'; }, 1500);
      } catch { feedback.textContent = 'No se pudo copiar. Selecciona el código para copiarlo manualmente.'; }
    }
  });
  reference.addEventListener('click', event => { if (event.target === reference) reference.close(); });
  sort.addEventListener('change', applySort);
  root.addEventListener('change', event => {
    if (event.target.matches('[data-problem-select]')) {
      problem = event.target.value;
      updateContent(event.target.closest('[data-review]'));
      remember();
    }
  });
  panels.forEach(panel => panel.querySelector('[data-review-scroll]').addEventListener('scroll', event => {
    saved[panel.dataset.review] = event.target.scrollTop;
    try { sessionStorage.setItem(storageKey, JSON.stringify(saved)); } catch { /* Optional persistence. */ }
  }, { passive: true }));
  list.addEventListener('scroll', () => {
    saved.list = list.scrollTop;
    try { sessionStorage.setItem(storageKey, JSON.stringify(saved)); } catch { /* Optional persistence. */ }
  }, { passive: true });
  root.addEventListener('keydown', event => {
    if (reference.open || document.querySelector('dialog[open]')) return;
    if (event.altKey && ['ArrowLeft', 'ArrowRight'].includes(event.key)) {
      event.preventDefault(); moveCandidate(event.key === 'ArrowRight' ? 1 : -1);
    }
    if (event.target.matches('[data-tab]') && ['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) {
      event.preventDefault();
      const buttons = [...event.target.closest('[role=tablist]').querySelectorAll('[data-tab]')];
      const next = event.key === 'Home' ? buttons[0] : event.key === 'End' ? buttons[buttons.length - 1] : buttons[(buttons.indexOf(event.target) + (event.key === 'ArrowRight' ? 1 : buttons.length - 1)) % buttons.length];
      next.click(); next.focus();
    }
  });
  applySort();
  list.scrollTop = saved.list || 0;
}

initializeReview();
document.addEventListener('htmx:load', initializeReview);
document.addEventListener('htmx:historyRestore', initializeReview);
document.addEventListener('htmx:afterSwap', event => {
  const trigger = event.detail.requestConfig?.elt;
  if (!trigger?.matches('[data-supervision]')) return;
  const dialog = document.querySelector('dialog[id^="report-modal-"][open]');
  if (dialog) dialog.querySelector('h2').textContent = `Supervisión · ${trigger.dataset.supervision}`;
});
