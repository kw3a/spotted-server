// Quiz Recording - Auto-start recording for quiz sessions
// Adapted from recording.js with auto-start, no preview, and ParticipationID integration
(function () {
    // ParticipationID will be set globally by the template
    const sessionID = window.participationID;

    const BASE_URL = 'http://localhost:8889/';

    let pc = null;
    let localStream = null;
    let whipResourceLocation = null;

    // Original streams (to stop at the end)
    let screenStream = null;
    let camStream = null;

    let isRecording = false;

    function log(...args) {
        console.log('[Quiz Recording]', ...args);
    }

    function getEl(id) {
        return document.getElementById(id);
    }

    function setStateEl(id, state) {
        const el = getEl(id);
        if (!el) return;
        // Reset classes
        el.className = 'w-3 h-3 rounded-full mt-1';

        let title = state;
        let colorClass = 'bg-gray-400'; // Default/unknown

        if (state === 'not-available') {
            title = 'No disponible';
            colorClass = 'bg-gray-500';
        } else if (state === 'not-allowed') {
            title = 'No permitido';
            colorClass = 'bg-red-500';
        } else if (state === 'working') {
            title = 'Activo';
            colorClass = 'bg-green-500';
        }

        el.title = title;
        el.classList.add(colorClass);
        el.textContent = ''; // Clear text
    }

    function preferH264(sdp) {
        const lines = sdp.split("\r\n");
        const mLineIndex = lines.findIndex(l => l.startsWith("m=video"));
        if (mLineIndex === -1) return sdp;
        const h264pt = lines
            .filter(l => l.startsWith("a=rtpmap") && l.toLowerCase().includes("h264"))
            .map(l => l.split(" ")[0].split(":")[1])[0];
        if (!h264pt) return sdp;
        const parts = lines[mLineIndex].split(" ");
        const newMLine = parts.slice(0, 3).join(" ") + " " + h264pt + " " + parts.slice(3).filter(p => p !== h264pt).join(" ");
        lines[mLineIndex] = newMLine;
        return lines.join("\r\n");
    }

    // --- Permissions & availability helpers ---
    async function checkDevicesAndPermissions() {
        try {
            if (navigator.permissions) {
                try {
                    const camPerm = await navigator.permissions.query({ name: 'camera' });
                    updateStateFromPermission('cam', camPerm.state);
                    camPerm.onchange = () => updateStateFromPermission('cam', camPerm.state);
                } catch (e) {
                    updateStateFromPermission('cam', 'unknown');
                }
                try {
                    const micPerm = await navigator.permissions.query({ name: 'microphone' });
                    updateStateFromPermission('mic', micPerm.state);
                    micPerm.onchange = () => updateStateFromPermission('mic', micPerm.state);
                } catch (e) {
                    updateStateFromPermission('mic', 'unknown');
                }
            } else {
                updateStateFromPermission('cam', 'unknown');
                updateStateFromPermission('mic', 'unknown');
            }
        } catch (e) {
            console.warn('Permission query failed', e);
        }

        if (navigator.mediaDevices && typeof navigator.mediaDevices.getDisplayMedia === 'function') {
            setStateEl('screenState', 'not-allowed');
        } else {
            setStateEl('screenState', 'not-available');
        }

        try {
            if (navigator.mediaDevices && navigator.mediaDevices.enumerateDevices) {
                const devices = await navigator.mediaDevices.enumerateDevices();
                const hasCam = devices.some(d => d.kind === 'videoinput');
                const hasMic = devices.some(d => d.kind === 'audioinput');
                if (!hasCam) setStateEl('camState', 'not-available');
                if (!hasMic) setStateEl('micState', 'not-available');
            }
        } catch { }
    }

    function updateStateFromPermission(which, permState) {
        if (which === 'cam') {
            if (permState === 'granted') setStateEl('camState', 'working');
            else if (permState === 'denied') setStateEl('camState', 'not-allowed');
            else setStateEl('camState', 'not-available');
        } else if (which === 'mic') {
            if (permState === 'granted') setStateEl('micState', 'working');
            else if (permState === 'denied') setStateEl('micState', 'not-allowed');
            else setStateEl('micState', 'not-available');
        }
    }

    async function startMixedStream() {
        // Request camera and microphone
        camStream = await navigator.mediaDevices.getUserMedia({
            video: { width: { ideal: 640 }, height: { ideal: 360 } },
            audio: true
        });
        const camTrack = camStream.getVideoTracks()[0];
        const micTrack = camStream.getAudioTracks()[0];

        // Update status
        setStateEl('camState', camTrack ? 'working' : 'not-available');
        setStateEl('micState', micTrack ? 'working' : 'not-available');

        // Request screen share
        screenStream = await navigator.mediaDevices.getDisplayMedia({
            video: { width: { ideal: 1280 }, height: { ideal: 720 }, frameRate: { ideal: 30 } },
            audio: false
        });

        setStateEl('screenState', 'working');

        // Create canvas for mixing screen + camera
        const canvas = document.createElement('canvas');
        canvas.width = 1280;
        canvas.height = 720;
        const ctx = canvas.getContext('2d');

        const screenVideo = document.createElement('video');
        screenVideo.srcObject = screenStream;
        screenVideo.play();

        const camVideo = document.createElement('video');
        camVideo.srcObject = new MediaStream([camTrack]);
        camVideo.play();

        function draw() {
            // Draw screen
            ctx.drawImage(screenVideo, 0, 0, canvas.width, canvas.height);
            // Draw camera overlay (bottom-right corner)
            const camW = 320, camH = 180;
            ctx.drawImage(camVideo, canvas.width - camW - 20, 20, camW, camH);
            requestAnimationFrame(draw);
        }
        draw();

        const mixedStream = canvas.captureStream(30);
        if (micTrack) mixedStream.addTrack(micTrack);
        return mixedStream;
    }

    async function startRecording() {
        if (isRecording) {
            log('Recording already in progress');
            return;
        }

        if (!sessionID) {
            log('Error: No participationID available');
            return;
        }

        log('Starting recording for participation:', sessionID);

        try {
            localStream = await startMixedStream();
            isRecording = true;
        } catch (err) {
            log('Error obtaining media:', err);
            setStateEl('micState', 'not-allowed');
            setStateEl('camState', 'not-allowed');
            setStateEl('screenState', 'not-allowed');
            return;
        }

        pc = new RTCPeerConnection({ iceServers: [] });
        localStream.getTracks().forEach(track => pc.addTrack(track, localStream));

        pc.addEventListener('icecandidate', e => {
            if (e.candidate) log('ICE candidate local:', e.candidate.candidate);
            else log('ICE gathering finished.');
        });

        try {
            const offer = await pc.createOffer();
            await pc.setLocalDescription(offer);
            let h264Sdp = preferH264(offer.sdp);

            const path = sessionID.trim().replace(/^\/+|\/+$/g, '');
            const publishUrl = BASE_URL + encodeURIComponent(path) + '/whip';
            log('POST SDP offer to', publishUrl);

            const resp = await fetch(publishUrl, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/sdp',
                    'Accept': 'application/sdp'
                },
                body: h264Sdp
            });

            if (!resp.ok) {
                const text = await resp.text().catch(() => '<no body>');
                throw new Error(`HTTP ${resp.status} ${resp.statusText} - ${text}`);
            }

            const answerSDP = await resp.text();
            whipResourceLocation = resp.headers.get('Location') || null;
            log('Response OK. Location:', whipResourceLocation);

            await pc.setRemoteDescription({ type: 'answer', sdp: answerSDP });
            log('RemoteDescription set. Recording in progress.');
            return true;
        } catch (err) {
            log('Error publishing stream:', err);
            pc?.close();
            pc = null;
            localStream?.getTracks().forEach(t => t.stop());
            localStream = null;
            stopAllDevices();
            isRecording = false;
            // Throw so the UI knows
            throw err;
        }
    }

    async function stopRecording() {
        if (!isRecording) {
            log('No recording in progress');
            return;
        }

        log('Stopping recording...');

        try {
            if (whipResourceLocation) {
                let url = whipResourceLocation;
                if (!/^https?:\/\//i.test(url)) {
                    const base = new URL(BASE_URL + sessionID.trim().replace(/^\/+|\/+$/g, '') + '/whip');
                    url = new URL(url, base).toString();
                }
                log('Sending DELETE to', url);
                await fetch(url, { method: 'DELETE' }).catch(e => log('DELETE error (ignore):', e));
                whipResourceLocation = null;
            }

            pc?.getSenders().forEach(s => {
                try { s.track?.stop(); } catch (e) { }
            });
            pc?.close();
            pc = null;

            localStream?.getTracks().forEach(t => t.stop());
            localStream = null;

            stopAllDevices();
            log('Recording stopped.');
        } finally {
            isRecording = false;
        }
    }

    function stopAllDevices() {
        if (screenStream) {
            screenStream.getTracks().forEach(t => t.stop());
            screenStream = null;
        }
        if (camStream) {
            camStream.getTracks().forEach(t => t.stop());
            camStream = null;
        }
    }

    // Initialization logic
    function init() {
        log('Initializing recording logic...');

        const startBtn = document.getElementById('btn-start-recording');
        const statusIcons = document.getElementById('recording-status-icons');

        if (startBtn) {
            log('Start button found in DOM');

            // Clean up: Clone to remove old listeners
            const newBtn = startBtn.cloneNode(true);

            // Validate: If parentNode is missing, the button is detached -> unexpected
            if (!startBtn.parentNode) {
                log('Warning: Start button parent missing');
                return;
            }

            startBtn.parentNode.replaceChild(newBtn, startBtn);
            log('Start button listener attached');

            newBtn.addEventListener('click', async () => {
                log('Button clicked');
                newBtn.disabled = true;
                newBtn.textContent = 'Iniciando...';

                // NOW we check permissions/devices
                await checkDevicesAndPermissions();

                try {
                    const success = await startRecording();
                    log('startRecording returned', success);

                    if (success) {
                        // Toggle UI: Hide button, Show icons
                        newBtn.style.display = 'none';
                        const currentStatusIcons = document.getElementById('recording-status-icons');
                        if (currentStatusIcons) {
                            log('Showing status icons');
                            currentStatusIcons.classList.remove('hidden');
                        } else {
                            console.error('Error: recording-status-icons not found in DOM');
                            alert('Error UI: Iconos no encontrados');
                        }
                    } else {
                        newBtn.disabled = false;
                        newBtn.textContent = 'Reintentar';
                        // If startRecording returned false (was handled internally) but we want to know why.
                        // Ideally startRecording should throw.
                        // For now, let's assume if it returns false, an error was logged.
                        // We can't alert here easily without changing startRecording.
                    }
                } catch (err) {
                    console.error("Failed to start recording:", err);
                    alert('Error al iniciar grabación: ' + err.message);
                    newBtn.disabled = false;
                    newBtn.textContent = 'Reintentar';
                }
            });
        } else {
            console.error('Fatal: Start recording button not found');
        }
    }

    // Since this script is placed at the end of the body (inserted by HTMX),
    // it executes immediately when the snippet is parsed.
    // We can just call init().
    init();

    // Stop recording on page unload (quiz finish, browser close, etc.)
    window.addEventListener('beforeunload', () => {
        if (pc) pc.close();
        if (localStream) localStream.getTracks().forEach(t => t.stop());
        stopAllDevices();
    });

    // Expose stop function globally for manual quiz finish
    window.stopQuizRecording = stopRecording;

})();
