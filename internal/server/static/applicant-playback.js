// Applicant Recording Playback - custom MSE player for the MediaMTX playback server.
// Combines all recording segments into a single seekable timeline.
(function () {
  const BASE = window.videoRecordsURL || "";
  const states = new Map();
  const recordingSegments = new Map();
  const loads = new Map();
  const WINDOW_SECONDS = 30;
  const BUFFER_AHEAD_SECONDS = 90;
  const KEEP_BEHIND_SECONDS = 60;

  function listUrl(path) {
    const url = new URL('/list', BASE);
    url.searchParams.set('path', path);
    return url.toString();
  }

  function getUrl(path, start, duration) {
    const url = new URL('/get', BASE);
    url.searchParams.set('path', path);
    url.searchParams.set('start', start);
    url.searchParams.set('duration', String(duration));
    url.searchParams.set('format', 'fmp4');
    return url.toString();
  }

  function addSeconds(value, seconds) {
    if (seconds === 0) return value;
    const milliseconds = toEpochSeconds(value) * 1000;
    if (!Number.isFinite(milliseconds)) return String(value);
    const fraction = typeof value === 'string'
      ? value.match(/\.(\d+)(?=Z|[+-]\d{2}:\d{2}$)/)?.[1]
      : null;
    const wholeSecond = Math.floor(milliseconds / 1000) * 1000 + seconds * 1000;
    const utc = new Date(wholeSecond).toISOString();
    return fraction ? utc.replace(/\.000Z$/, `.${fraction}Z`) : utc;
  }

  // MediaMTX /list returns `start` as an ISO-8601 string (e.g. "2026-03-18T18:23:16-04:00").
  function toEpochSeconds(value) {
    const ms = typeof value === 'number' ? value * 1000 : Date.parse(value);
    return Number.isFinite(ms) ? ms / 1000 : NaN;
  }

  function formatLocalTime(value) {
    const ms = typeof value === 'number' ? value * 1000 : Date.parse(value);
    if (!Number.isFinite(ms)) return String(value);
    return new Intl.DateTimeFormat(undefined, {
      hour: '2-digit', minute: '2-digit', second: '2-digit'
    }).format(new Date(ms));
  }

  function setStatus(el, text) {
    if (!el) return;
    const spinner = el.querySelector?.('[data-status-spinner]');
    const message = el.querySelector?.('[data-status-message]');
    if (spinner && message) {
      const value = text || '';
      const loading = value.startsWith('Cargando');
      if (loading && message.textContent && !message.textContent.startsWith('Cargando')) return;
      spinner.style.display = loading ? 'inline-block' : 'none';
      message.textContent = loading ? '' : value;
      el.title = value;
      el.setAttribute('aria-label', value);
      return;
    }
    if (!text) {
      el.textContent = '';
      el.classList.add('hidden');
      return;
    }
    el.textContent = text;
    el.classList.remove('hidden');
  }

  function clearLoadingStatus(el) {
    if (!el) return;
    const message = el.querySelector?.('[data-status-message]');
    const current = message ? message.textContent : el.textContent;
    if (!current || current.startsWith('Cargando')) setStatus(el, '');
  }

  // --- MP4 (fragmented) helpers ---

  function readU32(b, i) {
    return ((b[i] << 24) | (b[i + 1] << 16) | (b[i + 2] << 8) | b[i + 3]) >>> 0;
  }

  function indexOfAscii(bytes, str) {
    const pat = [];
    for (let i = 0; i < str.length; i++) pat.push(str.charCodeAt(i));
    outer:
    for (let i = 0; i <= bytes.length - pat.length; i++) {
      for (let k = 0; k < pat.length; k++) {
        if (bytes[i + k] !== pat[k]) continue outer;
      }
      return i;
    }
    return -1;
  }

  function hex2(n) {
    return n.toString(16).padStart(2, '0').toUpperCase();
  }

  // Build an MSE codec string from the init segment (ftyp+moov).
  function codecString(init) {
    const j = indexOfAscii(init, 'avcC');
    let video = 'avc1.42C01F';
    if (j >= 0 && j + 7 < init.length) {
      video = 'avc1.' + hex2(init[j + 5]) + hex2(init[j + 6]) + hex2(init[j + 7]);
    }
    let audio = null;
    if (indexOfAscii(init, 'Opus') >= 0) audio = 'opus';
    else if (indexOfAscii(init, 'mp4a') >= 0) audio = 'mp4a.40.2';
    return 'video/mp4; codecs="' + video + (audio ? ',' + audio : '') + '"';
  }

  // Split an fMP4 into its init segment (ftyp+moov) and its media fragments.
  function splitInitAndMedia(bytes) {
    let i = 0;
    while (i + 8 <= bytes.length) {
      let size = readU32(bytes, i);
      const type = String.fromCharCode(bytes[i + 4], bytes[i + 5], bytes[i + 6], bytes[i + 7]);
      if (size === 1) size = readU32(bytes, i + 8) * 2 ** 32 + readU32(bytes, i + 12);
      if (size < 8) break;
      if (type === 'moov') {
        return { init: bytes.subarray(0, i + size), media: bytes.subarray(i + size) };
      }
      i += size;
    }
    return { init: bytes, media: bytes.subarray(bytes.length) };
  }

  function appendBuffer(sb, data, signal) {
    return new Promise((resolve, reject) => {
      if (signal.aborted) { reject(signal.reason); return; }
      if (!data || data.byteLength === 0) return resolve();
      const cleanup = () => {
        sb.removeEventListener('updateend', onEnd);
        sb.removeEventListener('error', onErr);
        signal.removeEventListener('abort', onAbort);
      };
      const onEnd = () => {
        cleanup();
        resolve();
      };
      const onErr = () => {
        cleanup();
        reject(new Error('SourceBuffer append failed'));
      };
      const onAbort = () => { cleanup(); reject(signal.reason); };
      sb.addEventListener('updateend', onEnd);
      sb.addEventListener('error', onErr);
      signal.addEventListener('abort', onAbort, { once: true });
      try {
        sb.appendBuffer(data);
      } catch (e) {
        onErr();
      }
    });
  }

  function sameBytes(a, b) {
    if (a.length !== b.length) return false;
    for (let i = 0; i < a.length; i++) if (a[i] !== b[i]) return false;
    return true;
  }

  async function fetchBytes(url, signal) {
    const resp = await fetch(url, { signal });
    if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
    return new Uint8Array(await resp.arrayBuffer());
  }

  // --- Timestamp <-> unified timeline mapping ---

  function segmentAt(st, ts) {
    for (const s of st.segments) {
      if (ts >= s.epoch && ts < s.epoch + s.duration) {
        return s;
      }
    }
    return null;
  }

  function timestampAtPlayerTime(st, playerTime) {
    const segment = st.segments.find(s => playerTime >= s.timelineStart && playerTime < s.timelineStart + s.duration);
    return segment ? segment.epoch + (playerTime - segment.timelineStart) : null;
  }

  function isPlayerTimeBuffered(video, playerTime) {
    for (let i = 0; i < video.buffered.length; i++) {
      if (playerTime >= video.buffered.start(i) - 0.25 && playerTime < video.buffered.end(i)) return true;
    }
    return false;
  }

  function windowKey(segmentIndex, windowIndex) {
    return `${segmentIndex}:${windowIndex}`;
  }

  function enqueueAppend(st, action) {
    const next = st.appendQueue.then(() => {
      if (st.signal.aborted) throw st.signal.reason;
      return action();
    });
    st.appendQueue = next.catch(() => {});
    return next;
  }

  async function ensureWindow(st, segmentIndex, windowIndex) {
    const key = windowKey(segmentIndex, windowIndex);
    if (st.loadedWindows.has(key)) return;
    if (st.pendingWindows.has(key)) return st.pendingWindows.get(key);
    const segment = st.segments[segmentIndex];
    const segmentOffset = windowIndex * WINDOW_SECONDS;
    const duration = Math.min(WINDOW_SECONDS, segment.duration - segmentOffset);
    if (duration <= 0) return;
    const request = (async () => {
      setStatus(st.status, 'Cargando tramo de video...');
      const start = addSeconds(segment.start, segmentOffset);
      const parts = splitInitAndMedia(await fetchBytes(getUrl(st.participationID, start, duration), st.signal));
      await enqueueAppend(st, async () => {
        const sb = st.sourceBuffer;
        if (!sameBytes(st.currentInit, parts.init)) {
          const nextMime = codecString(parts.init);
          try { sb.changeType(nextMime); }
          catch (err) { console.warn('SourceBuffer.changeType failed for recording window', err); }
          await appendBuffer(sb, parts.init, st.signal);
          st.currentInit = parts.init;
          st.mime = nextMime;
        }
        sb.timestampOffset = segment.timelineStart + segmentOffset;
        await appendBuffer(sb, parts.media, st.signal);
      });
      st.loadedWindows.add(key);
      clearLoadingStatus(st.status);
    })();
    st.pendingWindows.set(key, request);
    try { await request; }
    finally { st.pendingWindows.delete(key); }
  }

  async function ensureBuffered(st, timestamp) {
    const target = segmentAt(st, timestamp);
    if (!target) return;
    const targetIndex = st.segments.indexOf(target);
    const targetOffset = timestamp - target.epoch;
    const endTime = target.timelineStart + targetOffset + BUFFER_AHEAD_SECONDS;
    const jobs = [];
    for (let i = targetIndex; i < st.segments.length; i++) {
      const segment = st.segments[i];
      if (segment.timelineStart > endTime) break;
      const from = i === targetIndex ? targetOffset : 0;
      const to = Math.min(segment.duration, endTime - segment.timelineStart);
      if (to < from) continue;
      const firstWindow = Math.floor(from / WINDOW_SECONDS);
      const lastWindow = Math.floor(Math.max(from, to - 0.001) / WINDOW_SECONDS);
      for (let w = firstWindow; w <= lastWindow; w++) jobs.push([i, w]);
    }
    if (jobs.length === 0) return;
    await ensureWindow(st, jobs[0][0], jobs[0][1]);
    (async () => {
      try {
        for (const [i, w] of jobs.slice(1)) await ensureWindow(st, i, w);
      } catch (err) {
        if (!st.signal.aborted) {
          console.error('Error buffering ahead:', err);
          setStatus(st.status, 'No se pudo cargar la siguiente parte de la grabación');
        }
      }
    })();
  }

  function trimBuffer(st) {
    const sb = st.sourceBuffer;
    if (!sb || sb.updating || sb.buffered.length === 0) return;
    const cutoff = (st.video.currentTime || 0) - KEEP_BEHIND_SECONDS;
    if (cutoff <= 0 || cutoff <= (st.lastTrimmed || 0) + WINDOW_SECONDS) return;
    const end = Math.min(cutoff, sb.buffered.end(0));
    if (end <= 0) return;
    enqueueAppend(st, () => new Promise((resolve, reject) => {
      if (st.signal.aborted) { reject(st.signal.reason); return; }
      const onEnd = () => { cleanup(); resolve(); };
      const onError = () => { cleanup(); reject(new Error('SourceBuffer remove failed')); };
      const cleanup = () => { sb.removeEventListener('updateend', onEnd); sb.removeEventListener('error', onError); };
      sb.addEventListener('updateend', onEnd, { once: true });
      sb.addEventListener('error', onError, { once: true });
      st.lastTrimmed = end;
      try { sb.remove(0, end); } catch (err) { cleanup(); reject(err); }
    })).then(() => {
      for (let i = 0; i < st.segments.length; i++) {
        const segment = st.segments[i];
        const count = Math.ceil(segment.duration / WINDOW_SECONDS);
        for (let w = 0; w < count; w++) {
          if (segment.timelineStart + Math.min(segment.duration, (w + 1) * WINDOW_SECONDS) <= end)
            st.loadedWindows.delete(windowKey(i, w));
        }
      }
    }).catch(err => { if (!st.signal.aborted) console.warn('Unable to trim video buffer:', err); });
  }

  function publishSegments(participationID, segments) {
    recordingSegments.set(participationID, segments);
    if (window.setTelemetryRecordingSegments) window.setTelemetryRecordingSegments(participationID, segments);
  }

  function updatePosition(participationID, timestamp, video) {
    const position = document.getElementById(`telemetry-position-${participationID}`);
    if (!position || !Number.isFinite(timestamp)) return;
    const minutes = Math.floor((video.currentTime || 0) / 60);
    const seconds = Math.floor((video.currentTime || 0) % 60).toString().padStart(2, '0');
    position.textContent = `Hora local: ${formatLocalTime(timestamp)} · Video: ${minutes}:${seconds}`;
  }

  function seekVideo(video, t) {
    return new Promise((resolve) => {
      const doSeek = () => {
        try {
          video.currentTime = t;
          video.play().catch(() => {});
        } catch (e) { }
        resolve();
      };
      if (video.readyState >= 1) doSeek();
      else video.addEventListener('loadedmetadata', doSeek, { once: true });
    });
  }

  // --- Loading ---

  async function loadRecordings(participationID, container, video, status) {
    window.disposeTelemetryRecording(participationID);
    const controller = new AbortController();
    const { signal } = controller;
    loads.set(participationID, controller);
    if (!BASE) {
      setStatus(status, 'Grabaciones no disponibles (sin servidor de video configurado).');
      publishSegments(participationID, []);
      return;
    }

    const prev = states.get(participationID);
    if (prev && prev.objectUrl) URL.revokeObjectURL(prev.objectUrl);
    states.delete(participationID);
    if (container) container.innerHTML = '';
    video.classList.add('hidden');
    setStatus(status, 'Cargando grabación...');

    try {
      const resp = await fetch(listUrl(participationID), { signal });
      if (resp.status === 400 || resp.status === 404) {
        setStatus(status, 'Grabaciones no disponibles');
        publishSegments(participationID, []);
        return;
      }
      if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
      const items = await resp.json();
      if (!Array.isArray(items) || items.length === 0) {
        setStatus(status, 'Grabaciones no disponibles');
        publishSegments(participationID, []);
        return;
      }

      items.sort((a, b) => toEpochSeconds(a.start) - toEpochSeconds(b.start));
      const segments = items.map((it) => ({
        start: it.start,
        epoch: toEpochSeconds(it.start),
        duration: it.duration,
        timelineStart: 0,
      }));
      publishSegments(participationID, segments);

      const firstDuration = Math.min(WINDOW_SECONDS, segments[0].duration);
      const first = splitInitAndMedia(
        await fetchBytes(getUrl(participationID, segments[0].start, firstDuration), signal)
      );
      const mime = codecString(first.init);

      if (!window.MediaSource || !MediaSource.isTypeSupported(mime)) {
        console.warn('MSE unsupported for', mime, '- falling back to per-segment playback');
        setupFallback(participationID, segments, container, video, status);
        return;
      }

      const mediaSource = new MediaSource();
      const objectUrl = URL.createObjectURL(mediaSource);
      const st = { segments, mediaSource, sourceBuffer: null, objectUrl, video, signal,
        mime, currentInit: null, loadedWindows: new Set(), pendingWindows: new Map(),
        appendQueue: Promise.resolve(), status, participationID };
      states.set(participationID, st);
      video.src = objectUrl;

      await new Promise((resolve, reject) => {
        const onAbort = () => reject(signal.reason);
        signal.addEventListener('abort', onAbort, { once: true });
        mediaSource.addEventListener('sourceopen', () => {
          signal.removeEventListener('abort', onAbort);
          resolve();
        }, { once: true, signal });
      });
      const sb = mediaSource.addSourceBuffer(mime);
      st.sourceBuffer = sb;
      let offset = 0;
      for (const segment of segments) {
        segment.timelineStart = offset;
        offset += segment.duration;
      }
      try { mediaSource.duration = offset; } catch (e) { }
      st.currentInit = first.init;
      await enqueueAppend(st, async () => {
        await appendBuffer(sb, first.init, signal);
        sb.timestampOffset = segments[0].timelineStart;
        await appendBuffer(sb, first.media, signal);
      });
      st.loadedWindows.add(windowKey(0, 0));

      video.classList.remove('hidden');
      renderPlaybackSummary(container, segments);
      const ensureForCurrentTime = () => {
        const playerTime = video.currentTime || 0;
        const timestamp = timestampAtPlayerTime(st, playerTime);
        if (timestamp !== null) ensureBuffered(st, timestamp).then(() => {
          if (!signal.aborted && !isPlayerTimeBuffered(video, playerTime)) {
            video.pause();
            setStatus(status, 'Sin video en este punto');
          }
        }).catch(err => {
          if (!signal.aborted) { console.error('Error buffering recording window:', err); setStatus(status, 'No se pudo cargar esta parte de la grabación'); }
        });
        trimBuffer(st);
      };
      st.ensureForCurrentTime = ensureForCurrentTime;
      video.addEventListener('timeupdate', () => {
        ensureForCurrentTime();
        const timestamp = timestampAtPlayerTime(st, video.currentTime || 0);
        if (Number.isFinite(timestamp) && window.setTelemetrySelection) {
          window.setTelemetrySelection(participationID, timestamp);
          updatePosition(participationID, timestamp, video);
        }
      }, { signal });
      video.addEventListener('seeked', () => {
        ensureForCurrentTime();
        const timestamp = timestampAtPlayerTime(st, video.currentTime || 0);
        if (Number.isFinite(timestamp) && window.setTelemetrySelection) {
          window.setTelemetrySelection(participationID, timestamp);
          updatePosition(participationID, timestamp, video);
        }
      }, { signal });
      video.addEventListener('seeking', ensureForCurrentTime, { signal });
      const initialTimestamp = timestampAtPlayerTime(st, video.currentTime || 0);
      if (Number.isFinite(initialTimestamp)) window.setTelemetrySelection?.(participationID, initialTimestamp);
      updatePosition(participationID, initialTimestamp, video);
      setStatus(status, '');
      ensureBuffered(st, segments[0].epoch).catch(() => {});
    } catch (err) {
      if (signal.aborted) return;
      console.error('Error loading recordings:', err);
      setStatus(status, 'Grabaciones no disponibles');
      publishSegments(participationID, []);
    }
  }

  // Fallback for browsers without MSE support: one button per segment (native player).
  function setupFallback(participationID, segments, container, video, status) {
    setStatus(status, 'Reproductor combinado no soportado; mostrando grabaciones individuales.');
    states.set(participationID, { segments, fallback: true, video });
    publishSegments(participationID, segments);
    if (!container) return;
    container.innerHTML = '';
    segments.forEach((s, idx) => {
      const btn = document.createElement('button');
      btn.className = 'px-3 py-1 bg-shark-700 hover:bg-shark-600 border border-shark-600 text-shark-200 rounded text-xs font-medium transition-colors';
      btn.textContent = `Grabación #${idx + 1} • ${formatLocalTime(s.start)} • ${Math.round(s.duration)}s`;
      btn.onclick = () => {
        video.src = getUrl(participationID, s.start, s.duration);
        video.classList.remove('hidden');
        video.load();
      };
      container.appendChild(btn);
    });
    if (segments[0]) {
      video.src = getUrl(participationID, segments[0].start, segments[0].duration);
      video.classList.remove('hidden');
    }
  }

  // The player must concatenate media, so the shared timeline above is the source
  // of truth for wall-clock time and gaps. This summary makes that explicit.
  function renderPlaybackSummary(container, segments) {
    if (!container) return;
    container.innerHTML = '';
    const first = segments[0];
    const last = segments[segments.length - 1];
    const message = document.createElement('p');
    message.className = 'text-xs text-shark-400';
    message.textContent = `${segments.length} segmento(s), desde ${formatLocalTime(first.start)} hasta ${formatLocalTime(last.epoch + last.duration)}. Los huecos se muestran arriba.`;
    container.appendChild(message);
  }

  window.setupTelemetryModal = function (participationID, containerId, videoId, statusId) {
    const container = document.getElementById(containerId);
    const video = document.getElementById(videoId);
    const status = document.getElementById(statusId);
    if (container && video && status) {
      loadRecordings(participationID, container, video, status);
      const dialog = video.closest('dialog');
      const signal = loads.get(participationID)?.signal;
      dialog?.querySelectorAll('[data-video-step]').forEach(button => button.addEventListener('click', () => {
        if (!Number.isFinite(video.duration)) return;
        video.currentTime = Math.max(0, Math.min(video.duration, video.currentTime + Number(button.dataset.videoStep)));
      }, { signal }));
      dialog?.querySelector('[data-video-speed]')?.addEventListener('change', event => {
        video.playbackRate = Number(event.target.value);
      }, { signal });
    }
  };

  window.disposeTelemetryRecording = function (participationID) {
    loads.get(participationID)?.abort();
    loads.delete(participationID);
    const state = states.get(participationID);
    if (state) {
      state.video.pause();
      state.video.removeAttribute('src');
      state.video.load();
      if (state.objectUrl) URL.revokeObjectURL(state.objectUrl);
    }
    states.delete(participationID);
    recordingSegments.delete(participationID);
  };

  window.getTelemetryRecordingSegments = function (participationID) {
    return recordingSegments.get(participationID) || [];
  };

  // Seek only when the requested instant is actually recorded. Never snap a gap.
  window.seekRecording = async function (participationID, unixTs) {
    const st = states.get(participationID);
    if (!st || !st.video) return;
    const segment = segmentAt(st, unixTs);
    const status = document.getElementById(`telemetry-status-${participationID}`);
    if (!segment) {
      setStatus(status, 'No hay video en el instante seleccionado. La franja rayada de la línea de tiempo indica este hueco.');
      return;
    }
    setStatus(status, '');
    if (st.fallback) {
      st.video.src = getUrl(participationID, segment.start, segment.duration);
      st.video.load();
      st.video.addEventListener('loadedmetadata', () => {
        seekVideo(st.video, unixTs - segment.epoch);
      }, { once: true });
      return;
    }
    const targetTime = segment.timelineStart + (unixTs - segment.epoch);
    try {
      await ensureBuffered(st, unixTs);
      if (!isPlayerTimeBuffered(st.video, targetTime)) {
        st.video.pause();
        setStatus(status, 'Sin video en este punto');
        return;
      }
      await seekVideo(st.video, targetTime);
      setStatus(status, '');
    } catch (err) {
      if (!st.signal.aborted) setStatus(status, 'No se pudo cargar esta parte de la grabación');
    }
  };
})();
