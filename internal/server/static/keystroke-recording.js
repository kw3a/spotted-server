(function () {
    let buffer = [];
    let keyDownTimes = {};
    let lastEventDownTime = null;
    let lastEventUpTime = null;
    let interval = null;
    const SEND_INTERVAL_MS = 60000; // 1 minute

    // Get Quiz ID globally set in HTML
    const getQuizID = () => window.quizID;

    function calculateStats(values) {
        // Filter out 0s if any (assuming 0 might be invalid, consistent with reference)
        const validValues = values.filter(v => v !== 0);

        if (validValues.length === 0) {
            return { mean: 0, std_dev: 0 };
        }

        const sum = validValues.reduce((a, b) => a + b, 0);
        const mean = sum / validValues.length;

        const variance = validValues.reduce((a, b) => a + Math.pow(b - mean, 2), 0) / validValues.length;
        const stdDev = Math.sqrt(variance);

        return {
            mean: Math.round(mean),
            std_dev: Math.round(stdDev)
        };
    }

    function sendBatch() {
        const quizID = getQuizID();
        if (!quizID) {
            console.warn("Keystroke telemetry: No quizID found.");
            return;
        }

        if (buffer.length === 0) {
            return; // Nothing to send
        }

        try {
            // Aggregate stats
            const stats = {
                ud: calculateStats(buffer.map(e => e.ud)),
                du1: calculateStats(buffer.map(e => e.du1)),
                dd: calculateStats(buffer.map(e => e.dd)),
                uu: calculateStats(buffer.map(e => e.uu)),
                du2: calculateStats(buffer.map(e => e.du2)),
                stroke_count: buffer.length
            };

            const formData = new FormData();
            formData.append("quizID", quizID);
            formData.append("strokeAmount", stats.stroke_count);

            formData.append("udMean", stats.ud.mean);
            formData.append("udStdDev", stats.ud.std_dev);

            formData.append("du1Mean", stats.du1.mean);
            formData.append("du1StdDev", stats.du1.std_dev);

            formData.append("du2Mean", stats.du2.mean);
            formData.append("du2StdDev", stats.du2.std_dev);

            formData.append("ddMean", stats.dd.mean);
            formData.append("ddStdDev", stats.dd.std_dev);

            formData.append("uuMean", stats.uu.mean);
            formData.append("uuStdDev", stats.uu.std_dev);

            fetch("/keystrokes", {
                method: "POST",
                body: formData
            }).catch(err => console.error("Error sending keystroke telemetry:", err));

        } catch (err) {
            console.error("Error in keystroke aggregation:", err);
        } finally {
            // Reset buffer
            buffer = [];
            // We don't reset lastEventDownTime/UpTime to maintain continuity across batches?
            // Reference implementation resets them. We should probably reset to avoid huge gaps if user goes idle?
            // "Reset buffer and state for independence" -> reference does reset times.
            lastEventDownTime = null;
            lastEventUpTime = null;
        }
    }

    // Event Listeners
    document.addEventListener("keydown", (e) => {
        if (!keyDownTimes[e.code]) {
            keyDownTimes[e.code] = performance.now();
        }
    });

    document.addEventListener("keyup", (e) => {
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

    // Auto-start
    interval = setInterval(sendBatch, SEND_INTERVAL_MS);

    // Initial check/log
    console.log("Keystroke telemetry started.");

    // Cleanup on unload
    window.addEventListener("beforeunload", () => {
        if (interval) clearInterval(interval);
        // Try to send last batch? optional, fetch might fail on unload. 
        // Navigator.sendBeacon could be used but our endpoint expects form data and maybe auth headers.
        // For now, just stop.
    });

})();
