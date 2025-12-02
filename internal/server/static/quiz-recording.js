// Quiz Recording - Auto-start recording for quiz sessions
// Adapted from recording.js with auto-start, no preview, and ParticipationID integration
(function () {
    // Status elements (to be defined in HTML)
    const micStateEl = document.getElementById('micState');
    const camStateEl = document.getElementById('camState');
    const screenStateEl = document.getElementById('screenState');

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

    function setStateEl(el, state) {
        if (!el) return;
        el.classList.remove('state-not-available', 'state-not-allowed', 'state-working');
        if (state === 'not-available') {
            el.textContent = 'no disponible';
            el.classList.add('state-not-available');
        } else if (state === 'not-allowed') {
            el.textContent = 'no permitido';
            el.classList.add('state-not-allowed');
        } else if (state === 'working') {
            el.textContent = 'activo';
            el.classList.add('state-working');
        } else {
            el.textContent = state;
        }
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
            setStateEl(screenStateEl, 'not-allowed');
        } else {
            setStateEl(screenStateEl, 'not-available');
        }

        try {
            if (navigator.mediaDevices && navigator.mediaDevices.enumerateDevices) {
                const devices = await navigator.mediaDevices.enumerateDevices();
                const hasCam = devices.some(d => d.kind === 'videoinput');
                const hasMic = devices.some(d => d.kind === 'audioinput');
                if (!hasCam) setStateEl(camStateEl, 'not-available');
                if (!hasMic) setStateEl(micStateEl, 'not-available');
            }
        } catch { }
    }

    function updateStateFromPermission(which, permState) {
        if (which === 'cam') {
            if (permState === 'granted') setStateEl(camStateEl, 'working');
            else if (permState === 'denied') setStateEl(camStateEl, 'not-allowed');
            else setStateEl(camStateEl, 'not-available');
        } else if (which === 'mic') {
            if (permState === 'granted') setStateEl(micStateEl, 'working');
            else if (permState === 'denied') setStateEl(micStateEl, 'not-allowed');
            else setStateEl(micStateEl, 'not-available');
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
        setStateEl(camStateEl, camTrack ? 'working' : 'not-available');
        setStateEl(micStateEl, micTrack ? 'working' : 'not-available');

        // Request screen share
        screenStream = await navigator.mediaDevices.getDisplayMedia({
            video: { width: { ideal: 1280 }, height: { ideal: 720 }, frameRate: { ideal: 30 } },
            audio: false
        });

        setStateEl(screenStateEl, 'working');

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
            setStateEl(micStateEl, 'not-allowed');
            setStateEl(camStateEl, 'not-allowed');
            setStateEl(screenStateEl, 'not-allowed');
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
        } catch (err) {
            log('Error publishing stream:', err);
            pc?.close();
            pc = null;
            localStream?.getTracks().forEach(t => t.stop());
            localStream = null;
            stopAllDevices();
            isRecording = false;
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

    // Auto-start on page load
    window.addEventListener('load', async () => {
        await checkDevicesAndPermissions();
        // Small delay to ensure UI is ready
        setTimeout(async () => {
            await startRecording();
        }, 500);
    });

    // Stop recording on page unload (quiz finish, browser close, etc.)
    window.addEventListener('beforeunload', () => {
        if (pc) pc.close();
        if (localStream) localStream.getTracks().forEach(t => t.stop());
        stopAllDevices();
    });

    // Expose stop function globally for manual quiz finish
    window.stopQuizRecording = stopRecording;

})();
