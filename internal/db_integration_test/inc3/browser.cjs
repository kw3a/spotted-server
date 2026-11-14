// Runs the application's unmodified recording/player scripts against native MediaMTX.
const assert = require('node:assert/strict');
const fs = require('node:fs/promises');
const path = require('node:path');
const os = require('node:os');
const net = require('node:net');
const dgram = require('node:dgram');
const http = require('node:http');
const { spawn, execFileSync } = require('node:child_process');
const { randomUUID } = require('node:crypto');

const [scenario, binary, root] = process.argv.slice(2);
const pause = ms => new Promise(resolve => setTimeout(resolve, ms));
const diagnostics = [];
let shutdown = async () => {};
let interrupted = false;

async function playwright() {
  if (process.env.PLAYWRIGHT_MODULE_PATH) return require(process.env.PLAYWRIGHT_MODULE_PATH);
  for (const module of ['playwright', 'playwright-core']) {
    try { return require(module); } catch (err) { if (err.code !== 'MODULE_NOT_FOUND') throw err; }
  }
  // Reuse an installed playwright-go driver instead of downloading another runtime.
  const cache = path.join(os.homedir(), '.cache/ms-playwright-go');
  const versions = (await fs.readdir(cache).catch(() => [])).sort((a,b) => b.localeCompare(a, undefined, { numeric: true }));
  for (const version of versions) {
    try { return require(path.join(cache, version, 'package')); } catch (err) { if (err.code !== 'MODULE_NOT_FOUND') throw err; }
  }
  throw new Error('Playwright is required: provide PLAYWRIGHT_MODULE_PATH or install playwright-core locally. No runtime is downloaded by this suite.');
}

async function eventually(fn, label, timeout = 20000) {
  const deadline = Date.now() + timeout;
  let last;
  do {
    if (interrupted) throw new Error('Test interrupted');
    try { const value = await fn(); if (value) return value; } catch (err) { last = err; }
    await pause(100);
  } while (Date.now() < deadline);
  throw new Error(`Timed out: ${label}${last ? ` (${last.message})` : ''}`);
}

async function tcpPort() {
  const server = net.createServer();
  await new Promise((resolve, reject) => server.once('error', reject).listen(0, '127.0.0.1', resolve));
  const port = server.address().port;
  await new Promise(resolve => server.close(resolve));
  return port;
}
async function udpPort() {
  const socket = dgram.createSocket('udp4');
  await new Promise((resolve,reject) => socket.once('error',reject).bind(0,'127.0.0.1',resolve));
  const port = socket.address().port;socket.close();return port;
}

async function nativeServer() {
  const directory = await fs.mkdtemp(path.join(os.tmpdir(), 'spotted-mediamtx-integration-'));
  const [api, publish, playback, udp] = await Promise.all([tcpPort(),tcpPort(),tcpPort(),udpPort()]);
  const config = path.join(directory, 'mediamtx.yml');
  await fs.writeFile(config, `logLevel: info
logDestinations: [stdout]
api: true
apiAddress: 127.0.0.1:${api}
rtsp: false
rtmp: false
hls: false
srt: false
webrtc: true
webrtcAddress: 127.0.0.1:${publish}
webrtcAllowOrigins: ['*']
webrtcLocalUDPAddress: 127.0.0.1:${udp}
webrtcLocalTCPAddress: ''
webrtcIPsFromInterfaces: false
webrtcAdditionalHosts: ['127.0.0.1']
playback: true
playbackAddress: 127.0.0.1:${playback}
playbackAllowOrigins: ['*']
pathDefaults:
  record: true
  recordPath: ${JSON.stringify(path.join(directory,'recordings/%path/%Y-%m-%d_%H-%M-%S-%f'))}
  recordFormat: fmp4
  recordPartDuration: 200ms
  recordSegmentDuration: 2s
  recordDeleteAfter: 0s
paths:
  all_others:
`);
  const child = spawn(binary, [config], { cwd: directory, stdio: ['ignore','pipe','pipe'] });
  child.stdout.on('data', data => diagnostics.push(data.toString()));
  child.stderr.on('data', data => diagnostics.push(data.toString()));
  const exited = new Promise(resolve => {child.once('error',err => resolve(err));child.once('exit',(code,signal) => resolve({code,signal}));});
  let stopped = false;
  const service = {
    directory, api:`http://127.0.0.1:${api}`, publish:`http://127.0.0.1:${publish}`, playback:`http://127.0.0.1:${playback}`,
    async stop() {
      if (stopped) return;
      stopped=true;
      if (child.exitCode === null && child.signalCode === null) child.kill('SIGINT');
      const timer=setTimeout(() => child.kill('SIGKILL'),8000);
      try { await exited; } finally {clearTimeout(timer);}
    },
    async close() {try {await this.stop();} finally {await fs.rm(directory,{recursive:true,force:true});}},
    async active(id) {
      const response=await fetch(`${this.api}/v3/paths/get/${id}`,{signal:AbortSignal.timeout(2000)});
      return response.ok ? response.json() : null;
    },
  };
  try {
    await eventually(async () => {
      assert.equal(child.exitCode,null,'MediaMTX exited during startup');
      const response=await fetch(`${service.api}/v3/paths/list`,{signal:AbortSignal.timeout(1000)});
      return response.ok;
    },'MediaMTX API readiness');
  } catch(err) {await service.close();throw err;}
  return service;
}

async function fixtureHTTP(service) {
  const scripts = {
    '/quiz-recording.js':await fs.readFile(path.join(root,'internal/server/static/quiz-recording.js')),
    '/applicant-playback.js':await fs.readFile(path.join(root,'internal/server/static/applicant-playback.js')),
  };
  const server=http.createServer((request,response) => {
    const url=new URL(request.url,'http://localhost');
    if (scripts[url.pathname]) {response.writeHead(200,{'Content-Type':'application/javascript; charset=utf-8'});response.end(scripts[url.pathname]);return;}
    const id=url.searchParams.get('id');
    if (!id || !/^[a-f0-9-]{36}$/.test(id)) {response.writeHead(400);response.end();return;}
    response.writeHead(200,{'Content-Type':'text/html; charset=utf-8'});
    // Minimal DOM contract for the real scripts; no WHIP, MSE or playback logic
    // is implemented in the fixture page itself.
    if (url.pathname==='/record') {
      response.end(`<!doctype html><style>.hidden{display:none}</style>
<button id="btn-start-recording"><span>Grabar</span></button><div id="recording-lock"></div>
<div id="recording-status-icons" class="hidden"><i id="camState"></i><i id="micState"></i><i id="screenState"></i></div>
<script>window.participationID=${JSON.stringify(id)};window.videoPublishURL=${JSON.stringify(service.publish)};</script>
<script src="/quiz-recording.js"></script>`);
    } else if (url.pathname==='/playback') {
      response.end(`<!doctype html><style>.hidden{display:none}</style>
<div id="segments"></div><video id="video" muted controls class="hidden"></video><div id="status"></div>
<script>window.videoRecordsURL=${JSON.stringify(service.playback)};</script>
<script src="/applicant-playback.js"></script>
<script>setupTelemetryModal(${JSON.stringify(id)},'segments','video','status');</script>`);
    } else response.end();
  });
  await new Promise((resolve,reject) => server.once('error',reject).listen(0,'127.0.0.1',resolve));
  return {url:`http://127.0.0.1:${server.address().port}`,close:() => new Promise(resolve => server.close(resolve))};
}

async function instrument(page, color) {
  page.on('console',message => {if (message.type()==='error') diagnostics.push(`browser: ${message.text()}`);});
  page.on('pageerror',error => diagnostics.push(`pageerror: ${error.message}`));
  const alerts=[];
  page.on('dialog',async dialog => {alerts.push(dialog.message());await dialog.dismiss();});
  await page.addInitScript(({color}) => {
    window.__mediaTracks=[];window.__peerConnections=[];
    const NativePC=window.RTCPeerConnection;
    window.RTCPeerConnection=class extends NativePC {constructor(...args){super(...args);window.__peerConnections.push(this);}};
    const getUserMedia=navigator.mediaDevices.getUserMedia.bind(navigator.mediaDevices);
    navigator.mediaDevices.getUserMedia=async constraints => {
      const stream=await getUserMedia(constraints);
      window.__mediaTracks.push(...stream.getTracks());return stream;
    };
    // Replace the physical screen source only. The app still composites it,
    // adds the browser's synthetic camera/microphone, and negotiates real WebRTC.
    navigator.mediaDevices.getDisplayMedia=async () => {
      const canvas=document.createElement('canvas');canvas.width=1280;canvas.height=720;
      const context=canvas.getContext('2d');
      function draw(){context.fillStyle=color;context.fillRect(0,0,1280,720);context.fillStyle='white';context.fillText(String(performance.now()),20,20);requestAnimationFrame(draw);}
      draw();const stream=canvas.captureStream(30);window.__mediaTracks.push(...stream.getTracks());return stream;
    };
  },{color});
  return alerts;
}

async function recording(browser,fixture,service,color='red') {
  const context=await browser.newContext({permissions:['camera','microphone']});
  const page=await context.newPage();
  const alerts=await instrument(page,color);
  const id=randomUUID();
  const responses=[];
  page.on('response',response => {if(response.url().startsWith(service.publish)) responses.push({method:response.request().method(),status:response.status(),url:response.url()});});
  await page.goto(`${fixture.url}/record?id=${id}`);
  return {id,page,context,alerts,responses};
}

async function start(record,service) {
  await record.page.click('#btn-start-recording');
  await eventually(() => record.page.evaluate(() => window.__peerConnections.some(pc => pc.connectionState==='connected')),'WebRTC connected');
  const active=await eventually(async () => {
    const state=await service.active(record.id);
    return state?.ready && state.bytesReceived>0 && state;
  },'MediaMTX receiving real media');
  assert.ok(active.tracks.some(track => /H264/.test(track)),`expected H264, received ${active.tracks}`);
  assert.ok(active.tracks.some(track => /Opus/.test(track)),`expected Opus, received ${active.tracks}`);
  assert.ok(record.responses.some(response => response.method==='POST' && response.status===201 && response.url.endsWith(`/${record.id}/whip`)));
  assert.equal(await record.page.locator('#recording-lock').count(),0);
  return active;
}

async function stop(record,service) {
  await record.page.evaluate(() => window.stopQuizRecording());
  assert.ok(record.responses.some(response => response.method==='DELETE' && response.status>=200 && response.status<300),'WHIP DELETE succeeded');
  assert.ok(await record.page.evaluate(() => window.__peerConnections.every(pc => pc.connectionState==='closed')));
  assert.ok(await record.page.evaluate(() => window.__mediaTracks.length>=3 && window.__mediaTracks.every(track => track.readyState==='ended')));
  await eventually(async () => !(await service.active(record.id))?.ready,'publisher removed');
}

async function capture(record,service) {
  await start(record,service);
  const began=Date.now();
  await eventually(async () => {
    const outbound=await record.page.evaluate(async () => {
      const stats=await window.__peerConnections[0].getStats();
      return [...stats.values()].filter(stat => stat.type==='outbound-rtp').map(stat => ({kind:stat.kind,bytes:stat.bytesSent,frames:stat.framesEncoded||0}));
    });
    return Date.now()-began>=5000 && outbound.some(s => s.kind==='video' && s.frames>=60 && s.bytes>0) && outbound.some(s => s.kind==='audio' && s.bytes>0);
  },'at least five seconds of encoded audio/video');
  await stop(record,service);
  await eventually(async () => {
    const response=await fetch(`${service.playback}/list?path=${record.id}`,{signal:AbortSignal.timeout(2000)});
    if(!response.ok)return false;
    const list=await response.json();return Array.isArray(list) && list.length>0 && list.every(item => item.duration>0);
  },'recording segments available');
}

async function playback(browser,fixture,id,color) {
  const page=await browser.newPage();
  const requests=[];
  page.on('response',response => {if(new URL(response.url()).pathname==='/get')requests.push(response);});
  await page.goto(`${fixture.url}/playback?id=${id}`);
  await eventually(() => page.evaluate(() => {
    const video=document.getElementById('video');return !video.classList.contains('hidden') && video.readyState>=2 && video.videoWidth>0 && video.duration>0;
  }),'player decoded recorded video');
  await page.evaluate(() => document.getElementById('video').play());
  await eventually(() => page.evaluate(() => document.getElementById('video').currentTime>0.4),'recorded video advancing');
  assert.ok(requests.length>0,'real recording downloaded');
  for(const response of requests){assert.equal(response.status(),200);assert.equal(new URL(response.url()).searchParams.get('path'),id);}
  if(color){
    const pixel=await page.evaluate(() => {
      const canvas=document.createElement('canvas');canvas.width=1;canvas.height=1;const context=canvas.getContext('2d');
      const video=document.getElementById('video');
      context.drawImage(video,Math.floor(video.videoWidth/4),Math.floor(video.videoHeight/2),1,1,0,0,1,1);return [...context.getImageData(0,0,1,1).data];
    });
    if(color==='red')assert.ok(pixel[0]>150 && pixel[2]<100,`expected red participation, pixel=${pixel}`);
    else assert.ok(pixel[2]>150 && pixel[0]<100,`expected blue participation, pixel=${pixel}`);
  }
  return page;
}

async function main() {
  assert.equal(execFileSync(binary,['--version'],{encoding:'utf8'}).trim(),'v1.17.0');
  const pw=await playwright();
  const service=await nativeServer();
  let fixture,browser;
  let closing;
  shutdown = () => closing ||= (async () => {
    try {if(browser)await browser.close();} finally {
      try {if(fixture)await fixture.close();} finally {await service.close();}
    }
  })();
  try {
    fixture=await fixtureHTTP(service);
    browser=await pw.chromium.launch({executablePath:process.env.CHROME_BINARY || '/usr/bin/google-chrome',headless:true,args:[
      '--no-sandbox','--use-fake-device-for-media-stream','--use-fake-ui-for-media-stream','--autoplay-policy=no-user-gesture-required',
    ]});
    if(scenario==='empty'){
      const page=await browser.newPage();const id=randomUUID();await page.goto(`${fixture.url}/playback?id=${id}`);
      await eventually(async () => (await page.locator('#status').textContent())==='Grabaciones no disponibles','empty recording state');
      assert.ok(await page.locator('#video').evaluate(video => video.classList.contains('hidden') && !video.getAttribute('src')));
      assert.deepEqual(await page.evaluate(id => window.getTelemetryRecordingSegments(id),id),[]);
    } else if(scenario==='unavailable'){
      await service.stop();
      const record=await recording(browser,fixture,service);
      await record.page.click('#btn-start-recording');
      await eventually(() => record.alerts.some(message => message.includes('Error al iniciar grabación')),'visible publishing failure');
      assert.equal(await record.page.locator('#btn-start-recording').textContent(),'Reintentar');
      assert.ok(await record.page.locator('#btn-start-recording').isEnabled());
      assert.ok(await record.page.evaluate(() => window.__peerConnections.length===1 && window.__peerConnections[0].connectionState==='closed'));
      assert.ok(await record.page.evaluate(() => window.__mediaTracks.length>=3 && window.__mediaTracks.every(track => track.readyState==='ended')));
    } else if(scenario==='isolation'){
      const sessions=[];
      for(const color of ['red','blue']){
        const record=await recording(browser,fixture,service,color);await capture(record,service);
        sessions.push({id:record.id,color});await record.context.close();
      }
      // Read both after the second recording exists: a service returning the
      // latest recording for every ID must fail the red-session assertion.
      for(const session of sessions){
        const page=await playback(browser,fixture,session.id,session.color);await page.close();
      }
    } else {
      const record=await recording(browser,fixture,service);
      if(scenario==='publish'){
        const before=await start(record,service);
        await eventually(async () => (await service.active(record.id))?.bytesReceived>before.bytesReceived,'received byte counter advancing');
        assert.deepEqual(record.alerts,[]);
        await stop(record,service);
      } else if(scenario==='stop'){
        await start(record,service);await stop(record,service);
      } else if(scenario==='playback'){
        await capture(record,service);await playback(browser,fixture,record.id,'red');
      } else throw new Error(`Unknown scenario: ${scenario}`);
    }
    console.log(`PASS ${scenario}: real application scripts and native MediaMTX v1.17.0`);
  } finally {
    await shutdown();
  }
}

// Go's test deadline and Ctrl-C close the native process and browser as well.
for(const signal of ['SIGINT','SIGTERM'])process.once(signal,async () => {
  interrupted = true;
  try {await shutdown();} finally {process.exitCode=130;}
});
main().catch(error => {
  console.error(error.stack || error);
  console.error(diagnostics.slice(-60).join(''));
  process.exitCode=1;
});
