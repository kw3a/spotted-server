// Telemetry and recording coverage share one wall-clock timeline.
(function () {
    const WINDOW_SECONDS = 60;
    const states = new Map();

    function formatTime(timestamp, seconds) {
        if (!Number.isFinite(timestamp)) return 'Hora no disponible';
        return new Intl.DateTimeFormat(undefined, {
            hour: '2-digit', minute: '2-digit', ...(seconds ? { second: '2-digit' } : {})
        }).format(new Date(timestamp * 1000));
    }

    function selectedMessage(start, end, coverage) {
        const range = `${formatTime(start, true)}–${formatTime(end, true)}`;
        if (!coverage) return `${range} · Sin video para este intervalo.`;
        return `${range} · Video disponible desde ${formatTime(coverage.epoch, true)}.`;
    }

    function setSelection(participationID, timestamp, source) {
        const state = states.get(participationID);
        if (!state || !Number.isFinite(timestamp)) return;
        state.selectedAt = timestamp;
        const { segments, selection, svg, xFor } = state;
        const coverage = segments.find(s => timestamp >= s.epoch && timestamp < s.epoch + s.duration);
        if (selection) selection.textContent = selectedMessage(timestamp, timestamp + WINDOW_SECONDS, coverage);
        if (!svg) return;
        const cursor = svg.querySelector('.timeline-cursor');
        if (cursor) {
            const x = xFor(timestamp);
            cursor.setAttribute('x1', x);
            cursor.setAttribute('x2', x);
            cursor.classList.remove('hidden');
        }
        svg.querySelectorAll('.ks-window').forEach(el => {
            const start = Number(el.dataset.start);
            const end = Number(el.dataset.end);
            const selected = timestamp >= start && timestamp < end;
            el.classList.toggle('is-selected', selected);
            const block = el.querySelector('rect');
            if (block) {
                block.setAttribute('stroke', selected ? '#0f172a' : 'none');
                block.setAttribute('stroke-width', selected ? '2' : '0');
            }
        });
        svg.querySelectorAll('.video-segment').forEach(el => {
            const start = Number(el.dataset.start);
            const end = Number(el.dataset.end);
            const selected = timestamp >= start && timestamp < end;
            el.classList.toggle('is-selected', selected);
            const block = el.querySelector('rect');
            if (block) {
                block.setAttribute('stroke', selected ? '#0f172a' : 'none');
                block.setAttribute('stroke-width', selected ? '2' : '0');
            }
        });
        if (source !== 'video' && window.seekRecording) window.seekRecording(participationID, timestamp);
    }

    function render(config) {
        const { data, containerId, legendId, noDataId, selectionId } = config;
        const container = document.getElementById(containerId);
        const legendContainer = document.getElementById(legendId);
        const noDataMsg = document.getElementById(noDataId);
        const selection = document.getElementById(selectionId);
        if (!container) return;

        const results = [...(data?.points || [])].sort((a, b) => (a.startAt || 0) - (b.startAt || 0));
        if (!results.length) {
            if (noDataMsg) noDataMsg.classList.remove('hidden');
            container.innerHTML = '';
            if (legendContainer) { legendContainer.innerHTML = ''; legendContainer.classList.add('hidden'); }
            return;
        }
        if (noDataMsg) noDataMsg.classList.add('hidden');

        const participationID = data.participationID;
        const existing = states.get(participationID);
        const knownSegments = existing?.segments || window.getTelemetryRecordingSegments?.(participationID) || [];
        const segments = knownSegments.filter(s => Number.isFinite(s.epoch) && Number.isFinite(s.duration));
        const pointsStart = Math.min(...results.map(r => r.startAt));
        const pointsEnd = Math.max(...results.map(r => r.startAt + WINDOW_SECONDS));
        const t0 = Math.min(pointsStart, ...(segments.map(s => s.epoch)));
        const t1 = Math.max(pointsEnd, ...(segments.map(s => s.epoch + s.duration)));
        const timeSpan = Math.max(t1 - t0, WINDOW_SECONDS);

        const width = Math.max(360, Math.min(1100, (container.clientWidth || 1000) - 32));
        const height = 318;
        const chartWidth = width - 54 - 24;
        const twoRows = chartWidth < 560;
        const padding = { top: twoRows ? 56 : 44, right: 24, bottom: 38, left: 54 };
        const plotBottom = 176;
        const telemetryY = 204;
        const coverageY = 246;
        const trackHeight = 22;
        const xFor = t => padding.left + ((t - t0) / timeSpan) * chartWidth;
        const yFor = value => padding.top + ((plotBottom - padding.top) * (1 - Math.min(value, 5) / 5));
        // ponytail: grid fijo de 5 min alineado a época; bloques y segmentos van en tiempo pared preciso, sin snap.
        const gridStep = 300;
        const firstTick = Math.ceil(t0 / gridStep) * gridStep;
        const tickCount = Math.max(1, Math.floor((t1 - firstTick) / gridStep) + 1);
        const labelEvery = chartWidth / tickCount < 24 ? 4 : chartWidth / tickCount < 48 ? 2 : 1;
        const selectedAt = existing?.selectedAt;

        let html = `<svg class="telemetry-timeline" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}" role="img" aria-label="Telemetría y cobertura de video en una línea de tiempo compartida" style="background:#fff;border:1px solid #ddd;border-radius:6px">`;
        html += `<defs><pattern id="no-video" width="8" height="8" patternUnits="userSpaceOnUse" patternTransform="rotate(45)"><rect width="8" height="8" fill="#f1f5f9"/><line x1="0" y1="0" x2="0" y2="8" stroke="#cbd5e1" stroke-width="3"/></pattern></defs>`;
        html += `<text x="${padding.left}" y="14" font-size="12" fill="#374151" font-weight="600">Dinámica de pulsaciones (SMD)</text>`;
        // ponytail: leyenda dentro del gráfico, arriba; una fila, dos si no cabe.
        const legendItems = [
            { t: 'Legítimo / perfil', c: '#16a34a', s: 'circle' },
            { t: 'Sospechoso (>1.2)', c: '#dc2626', s: 'circle' },
            { t: 'Inactivo', c: '#94a3b8', s: 'square' },
            { t: 'Sin video', c: '', s: 'hatch' },
        ];
        const legendRows = twoRows ? [legendItems.slice(0, 2), legendItems.slice(2)] : [legendItems];
        let legendY = 32;
        legendRows.forEach(row => {
            let lx = padding.left;
            row.forEach(it => {
                if (it.s === 'circle') html += `<circle cx="${lx + 5}" cy="${legendY - 4}" r="5" fill="${it.c}"/>`;
                else if (it.s === 'square') html += `<rect x="${lx}" y="${legendY - 9}" width="10" height="10" fill="${it.c}"/>`;
                else html += `<rect x="${lx}" y="${legendY - 9}" width="14" height="10" fill="url(#no-video)" stroke="#94a3b8"/>`;
                const tx = lx + (it.s === 'hatch' ? 18 : 14);
                html += `<text x="${tx}" y="${legendY}" font-size="11" fill="#374151">${it.t}</text>`;
                lx = tx + it.t.length * 5.8 + 16;
            });
            legendY += 14;
        });
        html += `<line x1="${padding.left}" y1="${plotBottom}" x2="${width - padding.right}" y2="${plotBottom}" stroke="#374151" stroke-width="1.5"/>`;
        html += `<line x1="${padding.left}" y1="${padding.top}" x2="${padding.left}" y2="${plotBottom}" stroke="#374151" stroke-width="1.5"/>`;
        for (let i = 0; i <= 5; i++) {
            const y = yFor(i);
            html += `<line x1="${padding.left}" y1="${y}" x2="${width - padding.right}" y2="${y}" stroke="#e5e7eb"/>`;
            html += `<text x="${padding.left - 9}" y="${y + 4}" text-anchor="end" font-size="11" fill="#6b7280">${i}</text>`;
        }
        const thresholdY = yFor(1.2);
        html += `<line x1="${padding.left}" y1="${thresholdY}" x2="${width - padding.right}" y2="${thresholdY}" stroke="#f59e0b" stroke-width="1.5" stroke-dasharray="5 4"/>`;
        html += `<text x="${width - padding.right + 4}" y="${thresholdY + 4}" font-size="10" fill="#b45309">1.2</text>`;
        for (let n = 0, tick = firstTick; tick <= t1; tick += gridStep, n++) {
            const x = xFor(tick);
            html += `<line x1="${x}" y1="${padding.top}" x2="${x}" y2="${coverageY + trackHeight}" stroke="#e5e7eb"/>`;
            if (n % labelEvery === 0) html += `<text x="${x}" y="${coverageY + trackHeight + 20}" text-anchor="middle" font-size="10" fill="#6b7280">${formatTime(tick)}</text>`;
        }

        let path = '';
        results.forEach((point, index) => {
            if (point.isInactive) {
                if (path) html += `<path d="${path}" fill="none" stroke="#2563eb" stroke-width="2"/>`;
                path = '';
                return;
            }
            const command = `${path ? 'L' : 'M'} ${xFor(point.startAt + WINDOW_SECONDS / 2)} ${yFor(point.smd || 0)}`;
            path += `${path ? ' ' : ''}${command}`;
            if (index === results.length - 1) html += `<path d="${path}" fill="none" stroke="#2563eb" stroke-width="2"/>`;
        });
        results.forEach((point, index) => {
            const x = xFor(point.startAt);
            const y = yFor(point.smd || 0);
            const end = point.startAt + WINDOW_SECONDS;
            const blockWidth = Math.max(3, xFor(end) - x);
            const status = point.isInactive ? 'Inactivo' : point.isProfile ? 'Perfil' : point.smd > 1.2 ? 'Sospechoso' : 'Legítimo';
            const color = point.isInactive ? '#94a3b8' : point.isProfile ? '#16a34a' : point.smd > 1.2 ? '#dc2626' : '#16a34a';
            html += `<g class="ks-window" data-start="${point.startAt}" data-end="${end}" tabindex="0" role="button" aria-label="Ventana ${index + 1}, ${formatTime(point.startAt, true)}, ${status}">`;
            html += `<title>${formatTime(point.startAt, true)}–${formatTime(end, true)} · ${point.strokeCount} teclas · ${status}</title>`;
            html += `<rect x="${x}" y="${telemetryY}" width="${blockWidth}" height="20" rx="2" fill="${color}" fill-opacity=".8"/>`;
            if (point.isInactive) html += `<text x="${x + blockWidth / 2}" y="${telemetryY + 15}" text-anchor="middle" font-size="12" fill="#fff">×</text>`;
            if (!point.isInactive) html += `<circle cx="${xFor(point.startAt + WINDOW_SECONDS / 2)}" cy="${y}" r="5" fill="${color}" stroke="#fff" stroke-width="2"/>`;
            html += '</g>';
        });
        html += `<text x="${padding.left}" y="${telemetryY - 7}" font-size="11" fill="#374151" font-weight="600">Telemetría · ventanas de 1 min</text>`;
        html += `<text x="${padding.left}" y="${coverageY - 7}" font-size="11" fill="#374151" font-weight="600">Cobertura de video</text>`;
        html += `<rect x="${padding.left}" y="${coverageY}" width="${chartWidth}" height="${trackHeight}" fill="url(#no-video)" stroke="#94a3b8" rx="3"/>`;
        segments.forEach((segment, index) => {
            const x = xFor(segment.epoch);
            const segmentWidth = Math.max(2, xFor(segment.epoch + segment.duration) - x);
            html += `<g class="video-segment" data-start="${segment.epoch}" data-end="${segment.epoch + segment.duration}" tabindex="0" role="button" aria-label="Video ${index + 1}, desde ${formatTime(segment.epoch, true)}">`;
            html += `<title>Video ${index + 1}: ${formatTime(segment.epoch, true)}–${formatTime(segment.epoch + segment.duration, true)}</title>`;
            html += `<rect x="${x}" y="${coverageY}" width="${segmentWidth}" height="${trackHeight}" fill="#2563eb" rx="2"/>`;
            html += '</g>';
        });
        if (!segments.length) html += `<text x="${padding.left + chartWidth / 2}" y="${coverageY + 16}" text-anchor="middle" font-size="11" fill="#64748b">Cargando disponibilidad de video…</text>`;
        const cursorX = Number.isFinite(selectedAt) ? xFor(selectedAt) : -20;
        html += `<line class="timeline-cursor${Number.isFinite(selectedAt) ? '' : ' hidden'}" x1="${cursorX}" x2="${cursorX}" y1="${padding.top}" y2="${coverageY + trackHeight}" stroke="#0f172a" stroke-width="2" stroke-dasharray="4 3" pointer-events="none"/>`;
        html += '</svg>';
        container.innerHTML = html;

        const svg = container.querySelector('svg');
        const state = { config, segments, selectedAt, selection, svg, xFor };
        states.set(participationID, state);
        const activate = event => setSelection(participationID, Number(event.currentTarget.dataset.start), 'chart');
        svg.querySelectorAll('.ks-window, .video-segment').forEach(el => {
            el.addEventListener('click', activate);
            el.addEventListener('keydown', event => {
                if (event.key === 'Enter' || event.key === ' ') {
                    event.preventDefault();
                    activate(event);
                }
            });
        });
        if (Number.isFinite(selectedAt)) setSelection(participationID, selectedAt, 'render');
        if (legendContainer) { legendContainer.innerHTML = ''; legendContainer.classList.add('hidden'); }
    }

    window.renderKeystrokeReport = config => render(config);
    window.setTelemetryRecordingSegments = (participationID, segments) => {
        const state = states.get(participationID);
        if (!state) return;
        state.segments = segments;
        render(state.config);
    };
    window.setTelemetrySelection = (participationID, timestamp) => setSelection(participationID, timestamp, 'video');
})();
