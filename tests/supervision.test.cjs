const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const vm = require('node:vm');
const path = require('node:path');

const readScript = name => fs.readFileSync(path.join(__dirname, '../internal/server/static', name), 'utf8');

function element() {
    const classes = new Set();
    return {
        clientWidth: 520, isConnected: true, innerHTML: '', textContent: '',
        classList: {
            add: value => classes.add(value), remove: value => classes.delete(value),
            contains: value => classes.has(value),
        },
    };
}

test('cursor updates the new chart after closing and reopening; resizing does not seek video', () => {
    const elements = new Map();
    const observers = [];
    let seeks = 0;
    const window = { seekRecording() { seeks++; } };
    const context = {
        window, document: { getElementById: id => elements.get(id) }, Intl,
        ResizeObserver: class {
            constructor(callback) { this.callback = callback; observers.push(this); }
            observe() {}
            disconnect() { this.disconnected = true; }
        },
    };
    vm.runInNewContext(readScript('keystroke-report.js'), context);
    const config = {
        data: { participationID: 'candidate', points: [{ startAt: 1000, smd: 1, strokeCount: 20 }] },
        containerId: 'chart', legendId: 'legend', noDataId: 'empty', selectionId: 'selection',
    };
    function mount() {
        const cursor = element();
        cursor.attributes = {};
        cursor.setAttribute = (key, value) => { cursor.attributes[key] = value; };
        const svg = { querySelector: () => cursor, querySelectorAll: () => [] };
        const chart = element();
        chart.querySelector = () => svg;
        for (const id of ['legend', 'empty', 'selection']) elements.set(id, element());
        elements.set('chart', chart);
        return { chart, cursor };
    }
    const first = mount();
    window.renderKeystrokeReport(config);
    window.setTelemetrySelection('candidate', 1010);
    const firstX = first.cursor.attributes.x1;
    window.disposeKeystrokeReport('candidate');
    assert.equal(observers[0].disconnected, true);
    const second = mount();
    window.renderKeystrokeReport(config);
    window.setTelemetrySelection('candidate', 1040);
    assert.ok(Number.isFinite(second.cursor.attributes.x1));
    assert.equal(second.cursor.classList.contains('hidden'), false);
    assert.equal(first.cursor.attributes.x1, firstX);
    second.chart.clientWidth = 360;
    observers[1].callback();
    assert.match(second.chart.innerHTML, /viewBox="0 0 328 318"/);
    assert.equal(seeks, 0);
});

test('closing during recording fetch aborts the old load before reopening', async () => {
    const elements = new Map(['recordings', 'video', 'status'].map(id => [id, element()]));
    const signals = [];
    let publications = 0;
    const window = { videoRecordsURL: 'http://localhost:9996', setTelemetryRecordingSegments() { publications++; } };
    vm.runInNewContext(readScript('applicant-playback.js'), {
        window, document: { getElementById: id => elements.get(id) }, AbortController, URL,
        console, fetch(url, { signal }) {
            signals.push(signal);
            return new Promise((resolve, reject) => signal.addEventListener('abort', () => reject(signal.reason)));
        },
    });
    window.setupTelemetryModal('candidate', 'recordings', 'video', 'status');
    window.disposeTelemetryRecording('candidate');
    window.setupTelemetryModal('candidate', 'recordings', 'video', 'status');
    assert.equal(signals[0].aborted, true);
    assert.equal(signals[1].aborted, false);
    window.disposeTelemetryRecording('candidate');
    await new Promise(resolve => setImmediate(resolve));
    assert.equal(publications, 0, 'Canceled requests must not overwrite the new chart');
});
