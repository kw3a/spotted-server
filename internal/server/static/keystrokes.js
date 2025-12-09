let isRunning = false;
let buffer = [];
let keyDownTimes = {};
let lastEventDownTime = null;
let lastEventUpTime = null;
let interval = null;
let timerInterval = null;
let secondsElapsed = 0;

const startBtn = document.getElementById('startBtn');
const startTelemetryBtn = document.getElementById('startTelemetryBtn');
const stopBtn = document.getElementById('stopBtn');
const sessionInput = document.getElementById('sessionid');
const timerProgress = document.getElementById('timerProgress');
const timerLabel = document.getElementById('timerLabel');

function startTelemetry() {
    const userID = sessionInput.value.trim();
    if (!userID) { alert("Enter session ID"); return; }

    if (isRunning) return; // Already running

    isRunning = true;
    buffer = [];
    keyDownTimes = {};
    lastEventDownTime = null;
    lastEventUpTime = null;
    secondsElapsed = 0;

    updateTimerUI();

    // Send every 1 minute (60000 ms)
    interval = setInterval(() => {
        sendBatch(userID);
        secondsElapsed = 0; // Reset timer
        updateTimerUI();
    }, 60000);

    // Update timer every second
    timerInterval = setInterval(() => {
        secondsElapsed++;
        updateTimerUI();
    }, 1000);

    // Update UI state
    startBtn.disabled = true;
    startTelemetryBtn.disabled = true;
    stopBtn.disabled = false;
    sessionInput.disabled = true;
}

function stopTelemetry() {
    isRunning = false;
    clearInterval(interval);
    clearInterval(timerInterval);
    interval = null;
    timerInterval = null;
    secondsElapsed = 0;
    updateTimerUI();

    startBtn.disabled = false;
    startTelemetryBtn.disabled = false;
    stopBtn.disabled = true;
    sessionInput.disabled = false;
}

function updateTimerUI() {
    if (!timerProgress || !timerLabel) return;
    const percentage = Math.min((secondsElapsed / 60) * 100, 100);
    timerProgress.style.width = `${percentage}%`;
    timerLabel.textContent = `${secondsElapsed}s`;
}

// "Start Recording + Telemetry" is handled by recording.js for media, 
// but we also need to start telemetry.
// recording.js attaches its own listener. We attach ours here.
startBtn.addEventListener('click', () => {
    startTelemetry();
});

// "Start Telemetry Only"
startTelemetryBtn.addEventListener('click', () => {
    startTelemetry();
});

stopBtn.addEventListener('click', () => {
    stopTelemetry();
    // recording.js also listens to this to stop media
});

document.getElementById("view").onclick = () => {
    const userID = sessionInput.value.trim();
    if (!userID) { alert("Enter session ID"); return; }

    fetch(`http://localhost:8080/profile/${userID}`)
        .then(r => r.json())
        .then(data => {
            document.getElementById("result").textContent = JSON.stringify(data, null, 2);
        })
        .catch(() => {
            document.getElementById("result").textContent = "Profile not found";
        });
};

document.addEventListener("keydown", (e) => {
    if (!isRunning) return;
    if (!keyDownTimes[e.code]) {
        keyDownTimes[e.code] = performance.now();
    }
});

document.addEventListener("keyup", (e) => {
    if (!isRunning) return;
    const downTime = keyDownTimes[e.code];
    if (!downTime) return;

    const upTime = performance.now();

    // Calculate raw features (rounded to integer)
    const ud = Math.round(upTime - downTime); // Dwell
    let du1 = 0; // Flight
    let dd = 0;
    let uu = 0;
    let du2 = 0;

    if (lastEventDownTime !== null) {
        dd = Math.round(downTime - lastEventDownTime);
        du2 = Math.round(upTime - lastEventDownTime);
    }
    if (lastEventUpTime !== null) {
        du1 = Math.round(downTime - lastEventUpTime);
        uu = Math.round(upTime - lastEventUpTime);
    }

    lastEventDownTime = downTime;
    lastEventUpTime = upTime;

    buffer.push({
        ud: ud,
        du1: du1,
        dd: dd,
        uu: uu,
        du2: du2
    });
    delete keyDownTimes[e.code];
});

function calculateStats(values) {
    // Filter out 0s (assuming 0 means "not applicable" for the first keystroke of the window)
    const validValues = values.filter(v => v !== 0); // Strict inequality if we allow negative? No, time diffs.
    // Actually, Flight Time (DU1) can be negative if overlap? 
    // "Flight time is the time between the release of a key and the press of the next".
    // If next key pressed BEFORE release of previous, it is negative (overlap).
    // So we should NOT filter out negative values, only "0" if it means "missing".
    // But wait, if we initialize to 0, and it IS 0 (very rare), we might filter it.
    // Let's assume 0 is missing for DD/UU/DU2/DU1 on first key.
    // But UD is never 0 (unless instant).

    // Better approach: use null or undefined for missing, but we used 0.
    // Let's stick to filtering 0 for now as it's the initialization value for "no previous key".

    if (validValues.length === 0) {
        return { mean: 0, std_dev: 0 };
    }

    const sum = validValues.reduce((a, b) => a + b, 0);
    const mean = sum / validValues.length;

    const variance = validValues.reduce((a, b) => a + Math.pow(b - mean, 2), 0) / validValues.length;
    const stdDev = Math.sqrt(variance);

    return {
        mean: Math.round(mean), // Round mean to integer
        std_dev: Math.round(stdDev) // Round stdDev to integer
    };
}

function sendBatch(userID) {
    if (!isRunning) {
        return;
    }

    try {
        // Aggregate
        const stats = {
            ud: calculateStats(buffer.map(e => e.ud)),
            du1: calculateStats(buffer.map(e => e.du1)),
            dd: calculateStats(buffer.map(e => e.dd)),
            uu: calculateStats(buffer.map(e => e.uu)),
            du2: calculateStats(buffer.map(e => e.du2)),
            stroke_count: buffer.length  // Number of keystrokes in this window
        };

        fetch("http://localhost:8080/telemetry", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                user_id: userID,
                stats: stats
            })
        }).catch(err => console.error("Error sending telemetry:", err));

    } catch (err) {
        console.error("Error in sendBatch aggregation:", err);
    } finally {
        // Reset buffer and state for independence, ensuring it happens even if aggregation fails
        buffer = [];
        lastEventDownTime = null;
        lastEventUpTime = null;
    }
}

// SMD Calculation Functions

document.getElementById("calculateSMD").onclick = () => {
    const userID = sessionInput.value.trim();
    if (!userID) { alert("Enter session ID"); return; }

    fetch(`http://localhost:8080/profile/${userID}`)
        .then(r => r.json())
        .then(data => {
            if (!Array.isArray(data) || data.length === 0) {
                document.getElementById("smdResults").innerHTML = "<p>No hay datos disponibles</p>";
                return;
            }

            if (data.length < 2) {
                document.getElementById("smdResults").innerHTML = "<p>Se necesitan al menos 2 ventanas para calcular SMD</p>";
                return;
            }

            // Find profile: window with most keystrokes
            let profileIndex = 0;
            let maxStrokes = data[0].stroke_count || 0;

            for (let i = 1; i < data.length; i++) {
                const strokes = data[i].stroke_count || 0;
                if (strokes > maxStrokes) {
                    maxStrokes = strokes;
                    profileIndex = i;
                }
            }

            const profile = data[profileIndex];

            // Calculate SMD for each window (except profile)
            const results = [];

            for (let i = 0; i < data.length; i++) {
                const strokeCount = data[i].stroke_count || 0;
                const isInactive = strokeCount < 10;

                if (i === profileIndex) {
                    results.push({
                        windowIndex: i,
                        smd: 0,
                        isProfile: true,
                        strokeCount: strokeCount,
                        isInactive: isInactive
                    });
                    continue;
                }

                // Skip SMD calculation for inactive windows
                if (isInactive) {
                    results.push({
                        windowIndex: i,
                        smd: 0,
                        isProfile: false,
                        strokeCount: strokeCount,
                        isInactive: true
                    });
                    continue;
                }

                const sample = data[i];
                const smd = calculateSMD(profile, sample);

                results.push({
                    windowIndex: i,
                    smd: smd,
                    isProfile: false,
                    strokeCount: strokeCount,
                    isInactive: false
                });
            }

            // Display results
            displaySMDResults(results, profileIndex);
        })
        .catch(err => {
            document.getElementById("smdResults").innerHTML = "<p>Error al obtener perfil</p>";
            console.error(err);
        });
};

function calculateSMD(profile, sample) {
    // Features: UD, DU1, DU2, DD, UU (N = 5)
    const features = ['ud', 'du1', 'du2', 'dd', 'uu'];
    let sum = 0;
    let validFeatures = 0;

    for (const feature of features) {
        const profileMean = profile[feature]?.mean || 0;
        const profileStdDev = profile[feature]?.std_dev || 0;
        const sampleMean = sample[feature]?.mean || 0;

        // Skip if std dev is 0 (avoid division by zero)
        if (profileStdDev === 0) {
            continue;
        }

        // Calculate scaled Manhattan distance for this feature
        const distance = Math.abs(profileMean - sampleMean) / profileStdDev;
        sum += distance;
        validFeatures++;
    }

    // Return average distance (or 0 if no valid features)
    return validFeatures > 0 ? sum / validFeatures : 0;
}


function displaySMDResults(results, profileIndex) {
    const smdResultsDiv = document.getElementById("smdResults");

    // SVG dimensions
    const width = 800;
    const height = 300;
    const padding = { top: 20, right: 20, bottom: 40, left: 50 };
    const chartWidth = width - padding.left - padding.right;
    const chartHeight = height - padding.top - padding.bottom;

    // Y-axis: 0 to 5
    const maxY = 5;
    const minY = 0;

    // Create SVG
    let html = `<svg width="${width}" height="${height}" style="background: white; border: 1px solid #ddd; border-radius: 6px;">`;

    // Draw Y-axis
    html += `<line x1="${padding.left}" y1="${padding.top}" x2="${padding.left}" y2="${height - padding.bottom}" stroke="#333" stroke-width="2"/>`;

    // Draw X-axis
    html += `<line x1="${padding.left}" y1="${height - padding.bottom}" x2="${width - padding.right}" y2="${height - padding.bottom}" stroke="#333" stroke-width="2"/>`;

    // Y-axis labels and grid lines
    for (let i = 0; i <= 5; i++) {
        const y = padding.top + (chartHeight * (5 - i) / 5);
        html += `<text x="${padding.left - 10}" y="${y + 5}" text-anchor="end" font-size="12" fill="#666">${i}</text>`;
        html += `<line x1="${padding.left}" y1="${y}" x2="${width - padding.right}" y2="${y}" stroke="#eee" stroke-width="1"/>`;
    }

    // Threshold line at 1.2 (legit vs suspicious)
    const thresholdY = padding.top + (chartHeight * (1 - 1.2 / maxY));
    html += `<line x1="${padding.left}" y1="${thresholdY}" x2="${width - padding.right}" y2="${thresholdY}" stroke="#ff9800" stroke-width="2" stroke-dasharray="5,5"/>`;
    html += `<text x="${width - padding.right + 5}" y="${thresholdY + 5}" font-size="11" fill="#ff9800" font-weight="bold">1.2</text>`;

    // Threshold labels
    html += `<text x="${padding.left + 10}" y="${padding.top + 15}" font-size="10" fill="#d32f2f" font-weight="bold">Sospechoso</text>`;
    html += `<text x="${padding.left + 10}" y="${height - padding.bottom - 10}" font-size="10" fill="#388e3c" font-weight="bold">Legítimo</text>`;


    // Y-axis label
    html += `<text x="15" y="${height / 2}" text-anchor="middle" font-size="12" fill="#333" transform="rotate(-90, 15, ${height / 2})">SMD</text>`;

    // Calculate X positions for each window
    const windowCount = results.length;
    const xStep = chartWidth / (windowCount > 1 ? windowCount - 1 : 1);

    // Draw line and points
    let points = [];

    for (let i = 0; i < results.length; i++) {
        const result = results[i];
        const x = padding.left + (i * xStep);
        const isInactive = result.strokeCount < 10;

        // Cap SMD at 5
        const smdValue = Math.min(result.smd, 5);
        const y = padding.top + (chartHeight * (1 - smdValue / maxY));

        points.push({ x, y, result, isInactive, index: i });
    }

    // Draw line segments (break at inactive windows)
    let segmentStart = null;
    for (let i = 0; i < points.length; i++) {
        const point = points[i];

        if (!point.isInactive) {
            if (segmentStart === null) {
                segmentStart = i;
            }
        } else {
            // Draw segment if we have one
            if (segmentStart !== null && i - segmentStart > 1) {
                let pathData = '';
                for (let j = segmentStart; j < i; j++) {
                    const p = points[j];
                    if (pathData === '') {
                        pathData = `M ${p.x} ${p.y}`;
                    } else {
                        pathData += ` L ${p.x} ${p.y}`;
                    }
                }
                html += `<path d="${pathData}" fill="none" stroke="#007bff" stroke-width="2"/>`;
            }
            segmentStart = null;
        }
    }

    // Draw final segment if exists
    if (segmentStart !== null && points.length - segmentStart > 1) {
        let pathData = '';
        for (let j = segmentStart; j < points.length; j++) {
            const p = points[j];
            if (!p.isInactive) {
                if (pathData === '') {
                    pathData = `M ${p.x} ${p.y}`;
                } else {
                    pathData += ` L ${p.x} ${p.y}`;
                }
            }
        }
        if (pathData !== '') {
            html += `<path d="${pathData}" fill="none" stroke="#007bff" stroke-width="2"/>`;
        }
    }

    // Draw points
    for (const point of points) {
        const { x, y, result, isInactive, index } = point;

        if (isInactive) {
            // Inactive time: larger gray X marker with background
            const xSize = 8;
            const yPos = height - padding.bottom - 15;

            // Group for hover
            html += `<g>`;
            html += `<title>Ventana ${index + 1}: ${result.strokeCount} teclas (Inactivo)</title>`;

            // Background circle
            html += `<circle cx="${x}" cy="${yPos}" r="12" fill="#f0f0f0" stroke="#999" stroke-width="1"/>`;

            // X marker
            html += `<line x1="${x - xSize}" y1="${yPos - xSize}" x2="${x + xSize}" y2="${yPos + xSize}" stroke="#666" stroke-width="2.5"/>`;
            html += `<line x1="${x - xSize}" y1="${yPos + xSize}" x2="${x + xSize}" y2="${yPos - xSize}" stroke="#666" stroke-width="2.5"/>`;
            html += `</g>`;
        } else if (result.isProfile) {
            // Profile: green square
            html += `<g>`;
            html += `<title>Ventana ${index + 1}: ${result.strokeCount} teclas (Perfil) - SMD: ${result.smd.toFixed(2)}</title>`;
            html += `<rect x="${x - 6}" y="${y - 6}" width="12" height="12" fill="#28a745" stroke="#1e7e34" stroke-width="2"/>`;
            html += `</g>`;
        } else {
            // Regular point: color based on threshold
            const isSuspicious = result.smd > 1.2;
            const color = isSuspicious ? '#d32f2f' : '#388e3c';
            const strokeColor = isSuspicious ? '#b71c1c' : '#2e7d32';
            const status = isSuspicious ? 'Sospechoso' : 'Legítimo';

            html += `<g>`;
            html += `<title>Ventana ${index + 1}: ${result.strokeCount} teclas (${status}) - SMD: ${result.smd.toFixed(2)}</title>`;
            html += `<circle cx="${x}" cy="${y}" r="5" fill="${color}" stroke="${strokeColor}" stroke-width="2"/>`;
            html += `</g>`;
        }

        // X-axis label
        html += `<text x="${x}" y="${height - padding.bottom + 20}" text-anchor="middle" font-size="11" fill="#666">W${index + 1}</text>`;

        // Hover info (simplified - just show value on top of point if not inactive)
        if (!isInactive) {
            html += `<text x="${x}" y="${y - 10}" text-anchor="middle" font-size="10" fill="#333">${result.smd.toFixed(2)}</text>`;
        }
    }

    html += '</svg>';

    // Legend
    html += '<div style="margin-top: 1rem; font-size: 0.85rem;">';
    html += '<div style="display: flex; gap: 1.5rem; flex-wrap: wrap;">';
    html += '<div><span style="display: inline-block; width: 12px; height: 12px; background: #28a745; border: 2px solid #1e7e34; margin-right: 0.3rem;"></span> Perfil</div>';
    html += '<div><span style="display: inline-block; width: 12px; height: 12px; background: #388e3c; border: 2px solid #2e7d32; border-radius: 50%; margin-right: 0.3rem;"></span> Legítimo (≤1.2)</div>';
    html += '<div><span style="display: inline-block; width: 12px; height: 12px; background: #d32f2f; border: 2px solid #b71c1c; border-radius: 50%; margin-right: 0.3rem;"></span> Sospechoso (&gt;1.2)</div>';
    html += '<div><span style="display: inline-block; margin-right: 0.3rem;">✕</span> Inactivo (&lt;10 teclas)</div>';
    html += '</div>';
    html += `<p style="margin-top: 0.5rem; color: #666;">Perfil: Ventana ${profileIndex + 1} (${results[profileIndex].strokeCount} teclas)</p>`;
    html += '</div>';

    smdResultsDiv.innerHTML = html;
}
