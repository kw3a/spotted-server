// Applicant Recording Playback - shared loader for the telemetry modal
(function () {
  const BASE = window.videoBrokerURL || "";

  function buildGetUrl(path, start, duration) {
    const url = new URL('/get', BASE);
    url.searchParams.set('path', path);
    url.searchParams.set('start', start);
    url.searchParams.set('duration', String(duration));
    url.searchParams.set('format', 'mp4');
    return url.toString();
  }

  function listUrl(path) {
    const url = new URL('/list', BASE);
    url.searchParams.set('path', path);
    return url.toString();
  }

  async function loadRecordings(participationID, container, video, status) {
    if (!participationID) return;
    if (!BASE) {
      if (status) {
        status.textContent = 'Grabaciones no disponibles (sin servidor de video configurado).';
        status.classList.remove('hidden');
      }
      return;
    }

    try {
      const resp = await fetch(listUrl(participationID));
      if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
      const items = await resp.json();

      if (!Array.isArray(items) || items.length === 0) {
        if (status) {
          status.textContent = 'Grabaciones no disponibles';
          status.classList.remove('hidden');
        }
        return;
      }

      container.innerHTML = '';
      items.forEach((it, idx) => {
        const btn = document.createElement('button');
        btn.className = 'px-3 py-1 bg-shark-700 hover:bg-shark-600 border border-shark-600 text-shark-200 rounded text-xs font-medium transition-colors';
        btn.textContent = `Grabación #${idx + 1} • ${it.start} • ${it.duration}s`;
        btn.onclick = () => loadVideo(participationID, it, video, status);
        container.appendChild(btn);
      });

      // Load first recording by default
      loadVideo(participationID, items[0], video, status);
    } catch (err) {
      if (status) {
        status.textContent = 'Grabaciones no disponibles';
        status.classList.remove('hidden');
      }
      console.error('Error loading recordings:', err);
    }
  }

  function loadVideo(participationID, it, video, status) {
    video.src = buildGetUrl(participationID, it.start, it.duration);
    video.classList.remove('hidden');
    video.load();
  }

  // Find the recording segment covering unixTs (seconds) and seek the video to that offset
  window.seekRecording = async function (participationID, unixTs, container, video, status) {
    if (!video) return;
    if (!BASE) return;

    try {
      const resp = await fetch(listUrl(participationID));
      if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
      const items = await resp.json();

      if (!Array.isArray(items) || items.length === 0) {
        status.textContent = 'Sin grabación para esta ventana.';
        status.classList.remove('hidden');
        return;
      }

      let target = null;
      let offset = 0;
      for (const it of items) {
        const end = it.start + it.duration;
        if (unixTs >= it.start && unixTs < end) {
          target = it;
          offset = unixTs - it.start;
          break;
        }
      }

      if (!target) {
        status.textContent = 'Sin grabación para esta ventana.';
        status.classList.remove('hidden');
        return;
      }

      // Activate the matching recording button if present
      const btns = container.querySelectorAll('button');
      btns.forEach((b, i) => {
        if (b.textContent.startsWith(`Grabación #${items.indexOf(target) + 1}`)) {
          b.classList.add('bg-shark-600');
        } else {
          b.classList.remove('bg-shark-600');
        }
      });

      status.classList.add('hidden');
      video.src = buildGetUrl(participationID, target.start, target.duration);
      video.classList.remove('hidden');
      video.load();
      video.addEventListener('loadedmetadata', () => {
        video.currentTime = Math.max(0, offset);
        video.play().catch(() => { });
      }, { once: true });
    } catch (err) {
      status.textContent = 'Sin grabación para esta ventana.';
      status.classList.remove('hidden');
      console.error('Error seeking recording:', err);
    }
  };

  window.setupTelemetryModal = function (participationID, containerId, videoId, statusId) {
    const container = document.getElementById(containerId);
    const video = document.getElementById(videoId);
    const status = document.getElementById(statusId);
    if (container && video && status) {
      loadRecordings(participationID, container, video, status);
    }
  };
})();
