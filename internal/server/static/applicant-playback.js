// Applicant Recording Playback - Button-triggered loading for offerAdmin page
// Adapted from playback.js for multiple applicants with on-demand loading
(function () {
    const BASE = 'http://localhost:9996';

    function buildGetUrl(path, start, duration) {
        const url = new URL('/get', BASE);
        url.searchParams.set('path', path);
        url.searchParams.set('start', start);
        url.searchParams.set('duration', String(duration));
        url.searchParams.set('format', 'mp4');
        return url.toString();
    }

    function renderList(items, path, container) {
        container.innerHTML = '';

        if (!Array.isArray(items) || items.length === 0) {
            const p = document.createElement('p');
            p.className = 'text-shark-300 text-center py-4';
            p.textContent = 'Grabaciones no disponibles';
            container.appendChild(p);
            return;
        }

        const list = document.createElement('div');
        list.className = 'grid grid-cols-1 md:grid-cols-2 gap-4 mt-4';

        items.forEach((it, idx) => {
            const card = document.createElement('div');
            card.className = 'border border-shark-700 bg-shark-800/50 rounded-lg p-4';

            const title = document.createElement('div');
            title.className = 'font-semibold text-shark-200 mb-3';
            title.textContent = `Grabación #${idx + 1} • ${it.start} • ${it.duration}s`;

            const video = document.createElement('video');
            video.controls = true;
            video.className = 'w-full rounded bg-black';
            const src = document.createElement('source');
            src.type = 'video/mp4';
            src.src = buildGetUrl(path, it.start, it.duration);
            video.appendChild(src);

            const link = document.createElement('a');
            link.href = src.src;
            link.target = '_blank';
            link.rel = 'noopener noreferrer';
            link.className = 'inline-block mt-2 text-blue-400 hover:text-blue-600 text-sm';
            link.textContent = 'Abrir en pestaña';

            card.appendChild(title);
            card.appendChild(video);
            card.appendChild(link);
            list.appendChild(card);
        });

        container.appendChild(list);
    }

    async function loadRecordings(participationID, container, button) {
        if (!participationID) {
            console.error('No participationID provided');
            return;
        }

        // Disable button and show loading state
        button.disabled = true;
        button.textContent = 'Cargando...';
        container.innerHTML = '<p class="text-shark-300 text-center py-4">Cargando grabaciones...</p>';

        try {
            const url = new URL('/list', BASE);
            url.searchParams.set('path', participationID);
            const resp = await fetch(url.toString());

            if (!resp.ok) {
                throw new Error(`HTTP ${resp.status}`);
            }

            const data = await resp.json();
            renderList(data, participationID, container);

            // Hide button after successful load
            button.style.display = 'none';
        } catch (err) {
            container.innerHTML = '';
            const p = document.createElement('p');
            p.className = 'text-shark-300 text-center py-4';
            p.textContent = 'Grabaciones no disponibles';
            container.appendChild(p);

            // Re-enable button on error
            button.disabled = false;
            button.textContent = 'Ver grabación';
            console.error('Error loading recordings:', err);
        }
    }

    // Initialize all recording sections
    function initializeRecordingSections() {
        const sections = document.querySelectorAll('[data-recording-section]');

        sections.forEach(section => {
            const participationID = section.dataset.participationId;
            const button = section.querySelector('[data-load-recordings]');
            const container = section.querySelector('[data-recordings-container]');

            if (button && container && participationID) {
                button.addEventListener('click', () => {
                    loadRecordings(participationID, container, button);
                });
            }
        });
    }

    // Initialize on DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initializeRecordingSections);
    } else {
        initializeRecordingSections();
    }

    // Expose for dynamic content (HTMX swaps, etc.)
    window.initializeRecordingSections = initializeRecordingSections;

})();
