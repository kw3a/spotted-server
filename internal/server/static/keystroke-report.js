// Telemetry and recording coverage share one wall-clock timeline.
(function () {
    const WINDOW_SECONDS = 3 * 60;
    const CUSUM_THRESHOLD = 4;
    const windowEnd = point => point.endAt > point.startAt ? point.endAt : point.startAt + WINDOW_SECONDS;
    const states = new Map();
    // ponytail: participaciones cuya disponibilidad de video ya se resolvió (haya o no segmentos).
    const segmentsResolved = new Set();

    function formatTime(timestamp, seconds) {
        if (!Number.isFinite(timestamp)) return 'Hora no disponible';
        return new Intl.DateTimeFormat(undefined, {
            hour: '2-digit', minute: '2-digit', ...(seconds ? { second: '2-digit' } : {})
        }).format(new Date(timestamp * 1000));
    }

    function selectedMessage(point, segments) {
        if (!point) return 'Sin ventana de telemetría en este instante.';
        const start = point.startAt;
        const end = windowEnd(point);
        const range = `${formatTime(start, true)}–${formatTime(end, true)}`;
        const overlaps = segments.filter(segment => segment.epoch < end && segment.epoch + segment.duration > start);
        let coveredUntil = start;
        for (const segment of overlaps) {
            if (segment.epoch > coveredUntil) break;
            coveredUntil = Math.max(coveredUntil, segment.epoch + segment.duration);
        }
        const coverage = !overlaps.length ? 'Sin video' : coveredUntil >= end ? 'Video completo' : 'Video parcial';
        const metric = point.isBaseline ? 'Referencia provisional' : point.isInsufficient ? 'Datos insuficientes' : `distancia ${point.score.toFixed(2)} · CUSUM ${point.cusum.toFixed(2)}${point.isSuspicious ? ' · cambio sospechoso' : ''}`;
        return `${range} · ${point.strokeCount} teclas · ${metric} · ${coverage}`;
    }

    function setSelection(participationID, timestamp, source) {
        const state = states.get(participationID);
        if (!state || !Number.isFinite(timestamp)) return;
        state.selectedAt = timestamp;
        const { segments, selection, svg, xFor } = state;
        const point = state.results.find(point => timestamp >= point.startAt && timestamp < windowEnd(point));
        if (selection) selection.textContent = selectedMessage(point, segments);
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
                block.setAttribute('stroke-width', selected ? state.strokeWidth : '0');
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
                block.setAttribute('stroke-width', selected ? state.strokeWidth : '0');
            }
        });
        if (source === 'chart' && window.seekRecording) window.seekRecording(participationID, timestamp);
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
            if (noDataMsg) { noDataMsg.classList.remove('hidden'); noDataMsg.classList.add('flex'); }
            container.innerHTML = '';
            if (legendContainer) legendContainer.classList.add('hidden');
            return;
        }
        if (noDataMsg) { noDataMsg.classList.add('hidden'); noDataMsg.classList.remove('flex'); }
        if (legendContainer) legendContainer.classList.remove('hidden');

        const participationID = data.participationID;
        const existing = states.get(participationID);
        const knownSegments = existing?.segments || window.getTelemetryRecordingSegments?.(participationID) || [];
        const segments = knownSegments.filter(s => Number.isFinite(s.epoch) && Number.isFinite(s.duration));
        const pointsStart = Math.min(...results.map(r => r.startAt));
        const pointsEnd = Math.max(...results.map(windowEnd));
        const t0 = Math.min(pointsStart, ...(segments.map(s => s.epoch)));
        const t1 = Math.max(pointsEnd, ...(segments.map(s => s.epoch + s.duration)));
        const timeSpan = Math.max(t1 - t0, WINDOW_SECONDS);

        // These are SVG coordinates, derived from the space allocated by Tailwind.
        // The SVG has no intrinsic pixel width or height and cannot size its parent.
        const width = Math.max(1, container.clientWidth);
        const height = Math.max(1, container.clientHeight);
        const padding = { top: height * 0.07, right: width * 0.05, left: width * 0.10 };
        const chartWidth = width - padding.left - padding.right;
        const plotBottom = height * 0.57;
        const telemetryY = height * 0.69;
        const coverageY = height * 0.85;
        const trackHeight = height * 0.06;
        const strokeWidth = Math.min(width, height) * 0.004;
        const xFor = t => padding.left + ((t - t0) / timeSpan) * chartWidth;
        const yFor = value => padding.top + ((plotBottom - padding.top) * (1 - Math.min(value, 8) / 8));
        // ponytail: grid fijo de 5 min alineado a época; bloques y segmentos van en tiempo pared preciso, sin snap.
        const gridStep = 300;
        const firstTick = Math.ceil(t0 / gridStep) * gridStep;
        const tickCount = Math.max(1, Math.floor((t1 - firstTick) / gridStep) + 1);
        // Estimate capacity in relative font units rather than pixel breakpoints.
        const fontSize = parseFloat(getComputedStyle(container).fontSize);
        const labelWidth = formatTime(t0).length * fontSize * 0.6;
        const labelEvery = Math.max(1, Math.ceil(labelWidth * tickCount / chartWidth));
        const selectedAt = existing?.selectedAt;
        const gaps = results.slice(1).map((point, index) => xFor(point.startAt) - xFor(results[index].startAt)).filter(gap => gap > 0);
        const pointRadius = Math.min(height * 0.018, ...gaps.map(gap => gap * 0.35));
        const patternSize = height * 0.035;
        const patternID = `no-video-${participationID}`;
        let html = `<svg class="telemetry-timeline block h-full w-full" viewBox="0 0 ${width} ${height}" role="group" aria-label="Telemetría y cobertura de video en una línea de tiempo compartida">`;
        html += `<defs><pattern id="${patternID}" width="${patternSize}" height="${patternSize}" patternUnits="userSpaceOnUse" patternTransform="rotate(45)"><rect width="${patternSize}" height="${patternSize}" fill="#f1f5f9"/><line x1="0" y1="0" x2="0" y2="${patternSize}" stroke="#cbd5e1" stroke-width="${patternSize * 0.35}"/></pattern></defs>`;
        html += `<line x1="${padding.left}" y1="${plotBottom}" x2="${width - padding.right}" y2="${plotBottom}" stroke="#374151" stroke-width="${strokeWidth}"/>`;
        html += `<line x1="${padding.left}" y1="${padding.top}" x2="${padding.left}" y2="${plotBottom}" stroke="#374151" stroke-width="${strokeWidth}"/>`;
        for (let i = 0; i <= 8; i += 2) {
            const y = yFor(i);
            html += `<line x1="${padding.left}" y1="${y}" x2="${width - padding.right}" y2="${y}" stroke="#e5e7eb"/>`;
            html += `<text x="${padding.left - width * 0.02}" y="${y}" dominant-baseline="middle" text-anchor="end" class="text-xs" fill="#6b7280">${i}</text>`;
        }
        const thresholdY = yFor(CUSUM_THRESHOLD);
        html += `<line x1="${padding.left}" y1="${thresholdY}" x2="${width - padding.right}" y2="${thresholdY}" stroke="#f59e0b" stroke-width="${strokeWidth}" stroke-dasharray="${strokeWidth * 4} ${strokeWidth * 3}"/>`;
        html += `<text x="${width - padding.right + width * 0.01}" y="${thresholdY}" dominant-baseline="middle" class="text-xs" fill="#b45309">4</text>`;
        for (let n = 0, tick = firstTick; tick <= t1; tick += gridStep, n++) {
            const x = xFor(tick);
            html += `<line x1="${x}" y1="${padding.top}" x2="${x}" y2="${coverageY + trackHeight}" stroke="#e5e7eb"/>`;
            if (n % labelEvery === 0) html += `<text x="${x}" y="${height * 0.97}" text-anchor="middle" class="text-xs" fill="#6b7280">${formatTime(tick)}</text>`;
        }

        let path = '';
        let pathColor = '';
        let previousPoint = null;
        const flushPath = () => {
            if (path) html += `<path d="${path}" fill="none" stroke="${pathColor}" stroke-width="${strokeWidth}"/>`;
            path = '';
            pathColor = '';
        };
        results.forEach(point => {
            if (point.isInsufficient || point.hasGapBefore) {
                flushPath();
                previousPoint = null;
                return;
            }
            const x = xFor((point.startAt + windowEnd(point)) / 2);
            const y = yFor(point.cusum || 0);
            const color = point.isSuspicious ? '#dc2626' : '#2563eb';
            if (!path) {
                path = `M ${x} ${y}`;
                pathColor = color;
            } else if (pathColor === color) {
                path += ` L ${x} ${y}`;
            } else {
                const previousX = xFor((previousPoint.startAt + windowEnd(previousPoint)) / 2);
                const previousY = yFor(previousPoint.cusum || 0);
                flushPath();
                path = `M ${previousX} ${previousY} L ${x} ${y}`;
                pathColor = color;
            }
            previousPoint = point;
        });
        flushPath();
        results.forEach((point, index) => {
            const x = xFor(point.startAt);
            const y = yFor(point.cusum || 0);
            const end = windowEnd(point);
            const blockWidth = Math.max(0, xFor(end) - x);
            const status = point.isInsufficient ? 'Datos insuficientes' : point.isBaseline ? 'Referencia provisional' : point.isSuspicious ? 'Cambio sospechoso' : 'Sin cambio acumulado';
            const color = point.isInsufficient ? '#94a3b8' : point.isBaseline ? '#60a5fa' : point.isSuspicious ? '#dc2626' : '#16a34a';
            html += `<g class="ks-window cursor-pointer" data-start="${point.startAt}" data-end="${end}" tabindex="0" role="button" aria-label="Ventana ${index + 1}, ${formatTime(point.startAt, true)}, ${status}">`;
            html += `<title>${formatTime(point.startAt, true)}–${formatTime(end, true)} · ${point.strokeCount} teclas · ${status}</title>`;
            html += `<rect x="${x}" y="${padding.top}" width="${blockWidth}" height="${telemetryY + trackHeight - padding.top}" fill="transparent"/>`;
            html += `<rect x="${x}" y="${telemetryY}" width="${blockWidth}" height="${trackHeight}" fill="${color}" fill-opacity=".8"/>`;
            if (!point.isInsufficient) html += `<circle cx="${xFor((point.startAt + end) / 2)}" cy="${y}" r="${pointRadius}" fill="${color}" stroke="#fff" stroke-width="${strokeWidth}"/>`;
            html += '</g>';
        });
        html += `<text x="${padding.left}" y="${telemetryY - height * 0.025}" class="text-xs font-semibold" fill="#374151">Telemetría · CUSUM</text>`;
        html += `<text x="${padding.left}" y="${coverageY - height * 0.025}" class="text-xs font-semibold" fill="#374151">Cobertura de video</text>`;
        html += `<rect x="${padding.left}" y="${coverageY}" width="${chartWidth}" height="${trackHeight}" fill="url(#${patternID})" stroke="#94a3b8"/>`;
        segments.forEach((segment, index) => {
            const x = xFor(segment.epoch);
            const segmentWidth = Math.max(0, xFor(segment.epoch + segment.duration) - x);
            html += `<g class="video-segment cursor-pointer" data-start="${segment.epoch}" data-end="${segment.epoch + segment.duration}" tabindex="0" role="button" aria-label="Video ${index + 1}, desde ${formatTime(segment.epoch, true)}">`;
            html += `<title>Video ${index + 1}: ${formatTime(segment.epoch, true)}–${formatTime(segment.epoch + segment.duration, true)}</title>`;
            html += `<rect x="${x}" y="${coverageY}" width="${segmentWidth}" height="${trackHeight}" fill="#2563eb"/>`;
            html += '</g>';
        });
        if (!segments.length) html += `<text x="${padding.left + chartWidth / 2}" y="${coverageY + trackHeight / 2}" dominant-baseline="middle" text-anchor="middle" class="text-xs" fill="#64748b">${segmentsResolved.has(participationID) ? 'Sin grabación disponible.' : 'Cargando disponibilidad de video…'}</text>`;
        const cursorX = Number.isFinite(selectedAt) ? xFor(selectedAt) : -width;
        html += `<line class="timeline-cursor${Number.isFinite(selectedAt) ? '' : ' hidden'}" x1="${cursorX}" x2="${cursorX}" y1="${padding.top}" y2="${coverageY + trackHeight}" stroke="#0f172a" stroke-width="${strokeWidth}" stroke-dasharray="${strokeWidth * 3} ${strokeWidth * 2}" pointer-events="none"/>`;
        html += '</svg>';
        container.innerHTML = html;

        const svg = container.querySelector('svg');
        const state = { config, segments, results, selectedAt, selection, svg, xFor, strokeWidth, observer: existing?.observer };
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
    }

    window.renderKeystrokeReport = config => {
        render(config);
        const state = states.get(config.data.participationID);
        const container = document.getElementById(config.containerId);
        if (!state || !container || state.observer) return;
        let previousWidth = container.clientWidth;
        let previousHeight = container.clientHeight;
        state.observer = new ResizeObserver(() => {
            if ((container.clientWidth === previousWidth && container.clientHeight === previousHeight) || !container.isConnected) return;
            previousWidth = container.clientWidth;
            previousHeight = container.clientHeight;
            render(config);
        });
        state.observer.observe(container);
    };
    window.disposeKeystrokeReport = participationID => {
        states.get(participationID)?.observer?.disconnect();
        states.delete(participationID);
        segmentsResolved.delete(participationID);
    };
    window.setTelemetryRecordingSegments = (participationID, segments) => {
        segmentsResolved.add(participationID);
        const state = states.get(participationID);
        if (!state) return;
        state.segments = segments;
        render(state.config);
    };
    window.setTelemetrySelection = (participationID, timestamp) => setSelection(participationID, timestamp, 'video');
    window.stepTelemetryWindow = (participationID, direction) => {
        const state = states.get(participationID);
        if (!state) return;
        const current = state.selectedAt;
        const index = state.results.findIndex(point => current >= point.startAt && current < windowEnd(point));
        const target = direction === 'deviation'
            ? state.results.find(point => point.startAt > (current ?? -Infinity) && point.isSuspicious)
            : state.results[index < 0 ? 0 : index + direction];
        if (target) setSelection(participationID, target.startAt, 'chart');
        else if (state.selection) state.selection.textContent = direction === 'deviation' ? 'No hay más ventanas fuera del umbral.' : 'No hay más ventanas en esta dirección.';
    };
})();
