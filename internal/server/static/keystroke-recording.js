(function () {
    const SEND_INTERVAL_MS = 180000;
    const MAX_STROKES_PER_WINDOW = 100;
    const LONG_PAUSE_MS = 1000;
    const STORAGE_PREFIX = "spotted-keystrokes:";
    let buffer = [];
    let keyDownTimes = new Map();
    let lastDown = null;
    let lastUp = null;
    let interval = null;
    let flushing = false;

    const getQuizID = () => window.quizID;
    const median = values => {
        if (!values.length) return 0;
        const sorted = [...values].sort((a, b) => a - b);
        const middle = Math.floor(sorted.length / 2);
        return sorted.length % 2 ? sorted[middle] : Math.round((sorted[middle - 1] + sorted[middle]) / 2);
    };
    const stats = values => {
        const valid = values.filter(value => Number.isFinite(value) && value > 0);
        if (!valid.length) return { mean: 0, std: 0, median: 0, mad: 0 };
        const mean = valid.reduce((sum, value) => sum + value, 0) / valid.length;
        const variance = valid.reduce((sum, value) => sum + (value - mean) ** 2, 0) / valid.length;
        const center = median(valid);
        return {
            mean: Math.round(mean), std: Math.round(Math.sqrt(variance)),
            median: center, mad: median(valid.map(value => Math.abs(value - center)))
        };
    };
    const storageKey = quizID => STORAGE_PREFIX + quizID;
    function loadQueue(quizID) {
        try { return JSON.parse(localStorage.getItem(storageKey(quizID)) || "[]"); }
        catch (_) { return []; }
    }
    function saveQueue(quizID, queue) {
        try { localStorage.setItem(storageKey(quizID), JSON.stringify(queue)); }
        catch (error) { console.error("Could not persist keystroke telemetry queue:", error); }
    }
    function toFormData(payload) {
        const form = new FormData();
        Object.entries(payload).forEach(([key, value]) => form.append(key, String(value)));
        return form;
    }
    async function drainQueue() {
        const quizID = getQuizID();
        if (!quizID || flushing || !navigator.onLine) return;
        flushing = true;
        try {
            let queue = loadQueue(quizID);
            while (queue.length && navigator.onLine) {
                const response = await fetch("/keystrokes", { method: "POST", body: toFormData(queue[0]) });
                if (!response.ok) break;
                queue.shift();
                saveQueue(quizID, queue);
            }
        } catch (error) {
            console.error("Error sending keystroke telemetry:", error);
        } finally {
            flushing = false;
        }
    }
    function enqueue(payload) {
        const quizID = getQuizID();
        if (!quizID) return;
        const queue = loadQueue(quizID);
        queue.push(payload);
        saveQueue(quizID, queue);
        drainQueue();
    }
    function closeWindow() {
        if (!buffer.length) return;
        const events = buffer;
        const startMs = events[0].epoch;
        const endMs = events[events.length - 1].epochEnd;
        const names = ["ud", "du1", "du2", "dd", "uu"];
        const calculated = Object.fromEntries(names.map(name => [name, stats(events.map(event => event[name]))]));
        const validTransitions = events.filter(event => event.hadPrevious).length;
        const longPauses = events.filter(event => event.longPause).length;
        const payload = {
            quizID: getQuizID(), windowId: crypto.randomUUID(), featureVersion: 1,
            windowStartMs: startMs, windowEndMs: endMs,
            strokeAmount: events.length,
            longPausePermille: validTransitions ? Math.round(1000 * longPauses / validTransitions) : 0
        };
        names.forEach(name => {
            payload[name + "Mean"] = calculated[name].mean;
            payload[name + "StdDev"] = calculated[name].std;
            payload[name + "Median"] = calculated[name].median;
            payload[name + "Mad"] = calculated[name].mad;
        });
        buffer = [];
        lastDown = null;
        lastUp = null;
        enqueue(payload);
    }
    function resetFocusState() {
        keyDownTimes.clear();
        lastDown = null;
        lastUp = null;
    }
    document.addEventListener("keydown", event => {
        if (!keyDownTimes.has(event.code)) keyDownTimes.set(event.code, { perf: performance.now(), epoch: Date.now() });
    });
    document.addEventListener("keyup", event => {
        const down = keyDownTimes.get(event.code);
        if (!down) return;
        const upPerf = performance.now();
        const ud = Math.max(1, Math.round(upPerf - down.perf));
        const du1 = lastUp === null ? 0 : Math.round(down.perf - lastUp);
        const dd = lastDown === null ? 0 : Math.round(down.perf - lastDown);
        const du2 = lastDown === null ? 0 : Math.round(upPerf - lastDown);
        const uu = lastUp === null ? 0 : Math.round(upPerf - lastUp);
        const hadPrevious = lastUp !== null;
        const longPause = du1 > LONG_PAUSE_MS;
        buffer.push({ ud, du1: longPause ? 0 : du1, dd: longPause ? 0 : dd, du2: longPause ? 0 : du2, uu: longPause ? 0 : uu, longPause, hadPrevious, epoch: down.epoch, epochEnd: Date.now() });
        lastDown = down.perf;
        lastUp = upPerf;
        keyDownTimes.delete(event.code);
        if (buffer.length >= MAX_STROKES_PER_WINDOW) closeWindow();
    });
    window.addEventListener("blur", resetFocusState);
    document.addEventListener("visibilitychange", () => { if (document.hidden) resetFocusState(); });
    window.addEventListener("online", drainQueue);
    interval = setInterval(closeWindow, SEND_INTERVAL_MS);
    window.addEventListener("pagehide", () => {
        if (interval) clearInterval(interval);
        closeWindow();
        const quizID = getQuizID();
        const queued = quizID && loadQueue(quizID);
        if (queued && queued.length && navigator.sendBeacon) navigator.sendBeacon("/keystrokes", toFormData(queued[0]));
    });
    drainQueue();
})();
