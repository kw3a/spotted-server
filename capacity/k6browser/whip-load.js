// capacity/k6browser/whip-load.js
// Réplica con k6 browser de la prueba WHIP (ver capacity/k6browser/whip-page.html).
// Cada VU abre un Chromium con dispositivos falsos (video sintético + tono,
// sin archivos) y publica un stream propio (vu<__VU>).
//
// k6 v2: API asíncrona y flags de Chromium por variable de entorno.
// Servir la página y ejecutar (solo escritura, no ejecutar aquí):
//
//   python3 -m http.server 8000 --directory capacity/k6browser
//   export K6_BROWSER_ARGS="--no-sandbox,--disable-dev-shm-usage,--use-fake-device-for-media-stream,--use-fake-ui-for-media-stream,--autoplay-policy=no-user-gesture-required"
//   k6 run --out json=results-k6-5.json capacity/k6browser/whip-load.js \
//     -e PAGE_URL=http://localhost:8000/whip-page.html \
//     -e WHIP_URL=https://rtc.menudencia.site -e USERS=5
//
// Contenido propio en lugar del patrón sintético: agregar a K6_BROWSER_ARGS
// --use-file-for-fake-video-capture=/ruta/video.y4m (el archivo debe estar
// en la máquina que ejecuta el navegador). Nota: probado sin éxito, el video
// resultante llega vacío; se deja documentado sin activar.
//
// Notas: los medios los genera Chromium (no k6); nack_count aquí cuenta
// mensajes (el navegador no expone paquetes NACKeados); DTLS se infiere
// igual que en Go (conectado menos ICE conectado); loss_ratio usa enviados
// como base porque el navegador no expone el rango de secuencia esperado.
import { browser } from 'k6/browser';
import { check, sleep } from 'k6';
import { Trend, Counter, Rate } from 'k6/metrics';

const USERS = parseInt(__ENV.USERS || '5', 10);
const PAGE_URL = __ENV.PAGE_URL || 'http://localhost:8000/whip-page.html';
const WHIP_URL = __ENV.WHIP_URL || 'http://localhost:8889';
const DURATION = __ENV.DURATION || '10';

export const options = {
  scenarios: {
    publishers: {
      executor: 'shared-iterations',
      vus: USERS,
      iterations: USERS,
      maxDuration: '3m',
      options: { browser: { type: 'chromium' } },
    },
  },
  thresholds: {
    whip_failed: ['rate<0.01'],
    signaling_latency: ['p(95)<3000'],
    ice_establishment: ['p(95)<5000'],
    video_loss_ratio: ['max<0.05'],
  },
};

const signalingLatency = new Trend('signaling_latency', true);
const iceEstablishment = new Trend('ice_establishment', true);
const dtlsHandshake = new Trend('dtls_handshake', true);
const firstRtp = new Trend('first_rtp', true);
const videoKbps = new Trend('video_kbps', true);
const videoLossRatio = new Trend('video_loss_ratio', true);
const videoRtt = new Trend('video_rtt', true);
const packetsSent = new Counter('packets_sent');
const packetsLost = new Counter('packets_lost');
const nackCount = new Counter('nack_count');
const failed = new Rate('whip_failed');

export default async function () {
  // STREAM fuerza el nombre (para varias instancias k6 en paralelo).
  const stream = __ENV.STREAM || `vu${__VU}`;
  // Arranque escalonado: evita que N Chromiums inicialicen el dispositivo
  // falso de video en el mismo instante (falla con "Could not start video source").
  // Ajustar con -e START_GAP=<segundos> (0 para ráfaga simultánea).
  const gap = parseFloat(__ENV.START_GAP || '2');
  if (gap > 0 && __VU > 1) sleep((__VU - 1) * gap);
  // Resolución de captura (ver ?width=&height=&fps= en la página).
  const geo = `&width=${__ENV.WIDTH || '1280'}&height=${__ENV.HEIGHT || '720'}&fps=${__ENV.FPS || '30'}&audio=${__ENV.AUDIO || '1'}`;
  const context = await browser.newContext();
  await context.grantPermissions(['camera', 'microphone']);
  const page = await context.newPage();
  let rep = null;
  try {
    const url = `${PAGE_URL}?whip=${encodeURIComponent(WHIP_URL)}&stream=${stream}&duration=${DURATION}${geo}`;
    await page.goto(url);
    const deadline = Date.now() + (parseInt(DURATION, 10) + 45) * 1000;
    for (;;) {
      const done = await page.evaluate(() => window.__done);
      if (done) break;
      if (Date.now() > deadline) break;
      sleep(1);
    }
    rep = await page.evaluate(() => window.__report || null);
  } finally {
    await page.close();
    await context.close();
  }

  const ok = check(rep, {
    'reporte completo': (r) => r !== null && r !== undefined,
    'sin error': (r) => r && !r.error,
    'http 201': (r) => r && r.signaling && (r.signaling.http_status === 201 || r.signaling.http_status === 200),
    'sin stalls': (r) => r && r.steady && (r.steady.stalls || []).length === 0,
  });
  failed.add(!ok, { stream });
  if (!rep) return;

  // Línea JSON por VU (parseable; el agregado sale con --out json).
  console.log(`REPORT ${stream} ` + JSON.stringify(rep));

  const tags = { stream };
  if (rep.signaling && rep.signaling.latency_ms != null) signalingLatency.add(rep.signaling.latency_ms, tags);
  if (rep.ice && rep.ice.establishment_ms != null) iceEstablishment.add(rep.ice.establishment_ms, tags);
  if (rep.dtls && rep.dtls.handshake_ms != null) dtlsHandshake.add(rep.dtls.handshake_ms, tags);
  if (rep.transport && rep.transport.first_rtp_ms != null) firstRtp.add(rep.transport.first_rtp_ms, tags);
  const v = (rep.transport && rep.transport.video) || {};
  const a = (rep.transport && rep.transport.audio) || {};
  packetsSent.add((v.packets_sent || 0) + (a.packets_sent || 0), tags);
  packetsLost.add((v.packets_lost || 0) + (a.packets_lost || 0), tags);
  if (rep.steady) {
    videoKbps.add(rep.steady.video_kbps || 0, tags);
    nackCount.add(rep.steady.nack_count || 0, tags);
  }
  if (v.loss_ratio != null) videoLossRatio.add(v.loss_ratio, tags);
  if (v.rtt_ms != null) videoRtt.add(v.rtt_ms, tags);
}

// Resumen agregado a archivo (las líneas REPORT llevan el detalle por VU).
export function handleSummary(data) {
  const get = (name, field) => {
    const m = data.metrics[name];
    if (!m) return null;
    return m.values ? m.values[field || 'avg'] : null;
  };
  return {
    'summary-k6.json': JSON.stringify(
      {
        vus: USERS,
        whip_failed_rate: get('whip_failed', 'rate'),
        signaling_latency_p95: get('signaling_latency', 'p(95)'),
        ice_establishment_p95: get('ice_establishment', 'p(95)'),
        video_kbps_avg: get('video_kbps'),
        video_loss_ratio_max: get('video_loss_ratio', 'max'),
        packets_sent: get('packets_sent', 'count'),
        packets_lost: get('packets_lost', 'count'),
      },
      null,
      '  ',
    ),
  };
}
