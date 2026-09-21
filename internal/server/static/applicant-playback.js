// Applicant Recording Playback - custom MSE player for the MediaMTX playback server.
// Combines all recording segments into a single seekable timeline.
(function () {
  const BASE = window.videoRecordsURL || "";
  const states = new Map();
  const recordingSegments = new Map();

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
    if (!text) {
      el.textContent = '';
      el.classList.add('hidden');
      return;
    }
    el.textContent = text;
    el.classList.remove('hidden');
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

  function appendBuffer(sb, data) {
    return new Promise((resolve, reject) => {
      if (!data || data.byteLength === 0) return resolve();
      const onEnd = () => {
        sb.removeEventListener('updateend', onEnd);
        sb.removeEventListener('error', onErr);
        resolve();
      };
      const onErr = () => {
        sb.removeEventListener('updateend', onEnd);
        sb.removeEventListener('error', onErr);
        reject(new Error('SourceBuffer append failed'));
      };
      sb.addEventListener('updateend', onEnd);
      sb.addEventListener('error', onErr);
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

  async function fetchBytes(url) {
    const resp = await fetch(url);
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

  function publishSegments(participationID, segments) {
    recordingSegments.set(participationID, segments);
    if (window.setTelemetryRecordingSegments) window.setTelemetryRecordingSegments(participationID, segments);
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
    if (!BASE) {
      setStatus(status, 'Grabaciones no disponibles (sin servidor de video configurado).');
      return;
    }

    const prev = states.get(participationID);
    if (prev && prev.objectUrl) URL.revokeObjectURL(prev.objectUrl);
    states.delete(participationID);
    if (container) container.innerHTML = '';
    video.classList.add('hidden');
    setStatus(status, 'Cargando grabación...');

    try {
      const resp = await fetch(listUrl(participationID));
      if (resp.status === 400 || resp.status === 404) {
        setStatus(status, 'Grabaciones no disponibles');
        return;
      }
      if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
      const items = await resp.json();
      if (!Array.isArray(items) || items.length === 0) {
        setStatus(status, 'Grabaciones no disponibles');
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

      const first = splitInitAndMedia(
        await fetchBytes(getUrl(participationID, segments[0].start, segments[0].duration))
      );
      const mime = codecString(first.init);

      if (!window.MediaSource || !MediaSource.isTypeSupported(mime)) {
        console.warn('MSE unsupported for', mime, '- falling back to per-segment playback');
        setupFallback(participationID, segments, container, video, status);
        return;
      }

      const mediaSource = new MediaSource();
      const objectUrl = URL.createObjectURL(mediaSource);
      const st = { segments, mediaSource, sourceBuffer: null, objectUrl, video };
      states.set(participationID, st);
      video.src = objectUrl;

      await new Promise((resolve) => mediaSource.addEventListener('sourceopen', resolve, { once: true }));
      const sb = mediaSource.addSourceBuffer(mime);
      st.sourceBuffer = sb;
      await appendBuffer(sb, first.init);

      let offset = 0;
      let currentInit = first.init;
      for (let k = 0; k < segments.length; k++) {
        let segInit, segMedia;
        if (k === 0) {
          segInit = first.init;
          segMedia = first.media;
        } else {
          const seg = splitInitAndMedia(
            await fetchBytes(getUrl(participationID, segments[k].start, segments[k].duration))
          );
          segInit = seg.init;
          segMedia = seg.media;
        }

        // Recording segments come from independent publish sessions and may carry a
        // different codec config (SPS/PPS). MSE can't mix configs in one SourceBuffer,
        // so switch type and re-append the init segment when it changes.
        if (k > 0 && !sameBytes(currentInit, segInit)) {
          try {
            sb.changeType(codecString(segInit));
          } catch (e) {
            console.warn('SourceBuffer.changeType failed for segment', k, e);
          }
          await appendBuffer(sb, segInit);
          currentInit = segInit;
        }

        segments[k].timelineStart = offset;
        sb.timestampOffset = offset;
        await appendBuffer(sb, segMedia);

        if (sb.buffered.length > 0) offset = sb.buffered.end(sb.buffered.length - 1);
        setStatus(status, `Cargando grabación... ${k + 1}/${segments.length}`);
      }

      try { mediaSource.duration = offset; } catch (e) { }
      if (mediaSource.readyState === 'open') mediaSource.endOfStream();

      video.classList.remove('hidden');
      renderPlaybackSummary(container, segments);
      video.addEventListener('timeupdate', () => {
        const timestamp = timestampAtPlayerTime(st, video.currentTime || 0);
        if (Number.isFinite(timestamp) && window.setTelemetrySelection) {
          window.setTelemetrySelection(participationID, timestamp);
        }
      });
      video.addEventListener('seeked', () => {
        const timestamp = timestampAtPlayerTime(st, video.currentTime || 0);
        if (Number.isFinite(timestamp) && window.setTelemetrySelection) {
          window.setTelemetrySelection(participationID, timestamp);
        }
      });
      setStatus(status, '');
    } catch (err) {
      console.error('Error loading recordings:', err);
      setStatus(status, 'Grabaciones no disponibles');
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
    }
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
    await seekVideo(st.video, segment.timelineStart + (unixTs - segment.epoch));
  };
})();
