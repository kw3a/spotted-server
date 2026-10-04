// capacity: publica 720p/H264 + Opus contra un endpoint WHIP con Pion y
// mide las 5 etapas (señalización, ICE, DTLS, transporte, steady state).
//
// Uso (1 usuario, JSON a stdout):
//
//	go run . -whip http://192.168.1.3:8889 -stream test \
//	  -video testdata/video.h264 -audio testdata/audio.ogg
//
// Uso (N usuarios concurrentes, JSON a archivo):
//
//	go run . -whip http://192.168.1.3:8889 -users 5 \
//	  -video testdata/video.h264 -audio testdata/audio.ogg
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/bits"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pion/interceptor"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
)

func whipPublish(client *http.Client, publishURL, offer string) (answer, location string, status int, err error) {
	req, err := http.NewRequest(http.MethodPost, publishURL,
		strings.NewReader(offer))
	if err != nil {
		return "", "", 0, err
	}
	req.Header.Set("Content-Type", "application/sdp")
	req.Header.Set("Accept", "application/sdp")
	resp, err := client.Do(req)
	if err != nil {
		return "", "", 0, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", "", resp.StatusCode,
			fmt.Errorf("whip rechazado: http %d: %s", resp.StatusCode, snippet(string(body)))
	}
	loc := resp.Header.Get("Location")
	if loc != "" && !strings.HasPrefix(loc, "http") {
		if base, berr := url.Parse(publishURL); berr == nil {
			if ref, rerr := url.Parse(loc); rerr == nil {
				loc = base.ResolveReference(ref).String()
			}
		}
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return "", loc, resp.StatusCode, fmt.Errorf("respuesta sdp vacía (http %d)", resp.StatusCode)
	}
	return string(body), loc, resp.StatusCode, nil
}

func snippet(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 300 {
		return s[:300] + "…"
	}
	return s
}

func whipDelete(client *http.Client, location string) {
	if location == "" {
		return
	}
	req, _ := http.NewRequest(http.MethodDelete, location, nil)
	if resp, err := client.Do(req); err == nil && resp != nil {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
		_ = resp.Body.Close()
	}
}

// readRTCP cuenta NACK/PLI/FIR/RR que el servidor devuelve al sender.
func readRTCP(sender *webrtc.RTPSender, m *Metrics, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		pkts, _, err := sender.ReadRTCP()
		if err != nil {
			return // pc cerrado o sender detenido
		}
		for _, pkt := range pkts {
			switch v := pkt.(type) {
			case *rtcp.TransportLayerNack:
				var n uint64
				for _, pair := range v.Nacks {
					n++ // el propio PacketID
					for i, want := 0, bits.OnesCount16(uint16(pair.LostPackets)); i < want; i++ {
						n++
					}
				}
				m.addNACK(n)
			case *rtcp.PictureLossIndication:
				m.addPLI(time.Now())
			case *rtcp.FullIntraRequest:
				m.addFIR()
			case *rtcp.ReceiverReport:
				m.addRR()
				now := time.Now()
				for _, rep := range v.Reports {
					m.addRRBlock(rep.SSRC, rep.TotalLost, rep.LastSequenceNumber,
						rep.Jitter, rep.LastSenderReport, rep.Delay, now)
				}
			}
		}
	}
}

// runConfig lleva lo parseado una vez (los frames se comparten solo-lectura).
type runConfig struct {
	whipBase   string
	duration   time.Duration
	iceTimeout time.Duration
	statsEvery time.Duration
	stun       string
	dumpSDP    string
	frames     [][][]byte
	opusPkts   []opusPacket
}

// runPublisher ejecuta una publicación completa contra un stream.
// exitCode: 0 ok, 1 fallo de red/media, 2 fallo de setup.
func runPublisher(cfg runConfig, stream string, progress func(string, ...any)) (Report, int) {
	logf := func(format string, args ...any) {
		if progress != nil {
			progress("["+stream+"] "+format, args...)
		}
	}
	fail := func(errStr string, code int) (Report, int) {
		return Report{Error: errStr}, code
	}

	m := newMetrics()
	publishURL := strings.TrimRight(cfg.whipBase, "/") + "/" + url.PathEscape(stream) + "/whip"

	me := &webrtc.MediaEngine{}
	if err := me.RegisterDefaultCodecs(); err != nil {
		return fail(err.Error(), 2)
	}
	ir := &interceptor.Registry{}
	// El tap va PRIMERO para ver los SR que genera el report-interceptor
	// (así se calcula RTT vía LSR/DLSR) y contar RTP tal cual sale al socket.
	ir.Add(&tapFactory{m: m})
	if err := webrtc.RegisterDefaultInterceptors(me, ir); err != nil {
		return fail(err.Error(), 2)
	}
	api := webrtc.NewAPI(webrtc.WithMediaEngine(me), webrtc.WithInterceptorRegistry(ir))

	pcCfg := webrtc.Configuration{}
	if cfg.stun != "" {
		pcCfg.ICEServers = []webrtc.ICEServer{{URLs: []string{cfg.stun}}}
	}
	pc, err := api.NewPeerConnection(pcCfg)
	if err != nil {
		return fail(err.Error(), 2)
	}
	defer pc.Close()

	pcConnected := make(chan struct{})
	iceConnected := make(chan struct{})
	var pcOnce, iceOnce sync.Once
	pc.OnICEConnectionStateChange(func(s webrtc.ICEConnectionState) {
		m.onICE(s)
		if s == webrtc.ICEConnectionStateConnected || s == webrtc.ICEConnectionStateCompleted {
			iceOnce.Do(func() { close(iceConnected) })
		}
	})
	pc.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		m.onPC(s, time.Now())
		if s == webrtc.PeerConnectionStateConnected {
			pcOnce.Do(func() { close(pcConnected) })
		}
	})

	videoTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264, ClockRate: 90000}, "video", "capacity")
	if err != nil {
		return fail(err.Error(), 2)
	}
	audioTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 2}, "audio", "capacity")
	if err != nil {
		return fail(err.Error(), 2)
	}
	videoSender, err := pc.AddTrack(videoTrack)
	if err != nil {
		return fail(err.Error(), 2)
	}
	audioSender, err := pc.AddTrack(audioTrack)
	if err != nil {
		return fail(err.Error(), 2)
	}

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		return fail(err.Error(), 2)
	}
	if err := pc.SetLocalDescription(offer); err != nil {
		return fail(err.Error(), 2)
	}
	<-webrtc.GatheringCompletePromise(pc)
	if cfg.dumpSDP != "" {
		_ = os.WriteFile(cfg.dumpSDP+"-offer.sdp", []byte(pc.LocalDescription().SDP), 0o600)
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	tSig := time.Now()
	answerSDP, location, status, sigErr := whipPublish(httpClient, publishURL, pc.LocalDescription().SDP)
	sigLat := time.Since(tSig)
	if sigErr != nil {
		rejected := status >= 400
		m.setSignaling(sigLat, status, "", rejected, sigErr.Error())
		return m.report(publishURL, 0, sigErr.Error()), 1
	}
	m.setSignaling(sigLat, status, "", false, "")
	if cfg.dumpSDP != "" {
		_ = os.WriteFile(cfg.dumpSDP+"-answer.sdp", []byte(answerSDP), 0o600)
	}

	if err := pc.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer, SDP: answerSDP,
	}); err != nil {
		m.setSignaling(sigLat, status, "", true, "answer inválida: "+err.Error())
		return m.report(publishURL, 0, err.Error()), 1
	}
	m.setT0(time.Now())

	var wg sync.WaitGroup
	wg.Add(2)
	go readRTCP(videoSender, m, &wg)
	go readRTCP(audioSender, m, &wg)
	statsDone := make(chan struct{})
	go m.sampleLoop(cfg.statsEvery, statsDone)

	// ICE: conectado o timeout/fallo.
	select {
	case <-iceConnected:
	case <-time.After(cfg.iceTimeout):
		m.setICETimeout()
	}
	if m.iceFailed || m.iceTimeout {
		cleanup(httpClient, location, pc, statsDone, &wg)
		return m.report(publishURL, 0, "ice no establecido"), 1
	}
	// DTLS: PC connected poco después de ICE.
	select {
	case <-pcConnected:
	case <-time.After(5 * time.Second):
		m.onPC(webrtc.PeerConnectionStateFailed, time.Now())
		cleanup(httpClient, location, pc, statsDone, &wg)
		return m.report(publishURL, 0, "dtls handshake timeout"), 1
	}
	if m.pcFailed {
		cleanup(httpClient, location, pc, statsDone, &wg)
		return m.report(publishURL, 0, "dtls falló"), 1
	}
	logf("conectado, transmitiendo %.0fs", cfg.duration.Seconds())

	// SSRCs negociados (el tap ya los mapea por StreamInfo; esto es respaldo).
	for _, s := range []struct {
		sender *webrtc.RTPSender
		kind   string
	}{{videoSender, "video"}, {audioSender, "audio"}} {
		for _, enc := range s.sender.GetParameters().Encodings {
			if enc.SSRC != 0 {
				m.bindSSRC(uint32(enc.SSRC), s.kind)
			}
		}
	}

	// Steady state: transmite durante duration.
	ctx, cancel := context.WithTimeout(context.Background(), cfg.duration)
	var sendWG sync.WaitGroup
	sendWG.Add(2)
	var sendErr error
	var sendMu sync.Mutex
	go func() {
		defer sendWG.Done()
		if err := sendVideoLoop(ctx, videoTrack, cfg.frames, m); err != nil {
			sendMu.Lock()
			sendErr = err
			sendMu.Unlock()
		}
	}()
	go func() {
		defer sendWG.Done()
		if err := sendAudioLoop(ctx, audioTrack, cfg.opusPkts, m); err != nil {
			sendMu.Lock()
			sendErr = err
			sendMu.Unlock()
		}
	}()
	sendWG.Wait()
	cancel()
	time.Sleep(300 * time.Millisecond) // deja llegar RTCP/stats finales

	errStr := ""
	if sendErr != nil {
		errStr = "envío: " + sendErr.Error()
	}
	cleanup(httpClient, location, pc, statsDone, &wg)
	rep := m.report(publishURL, cfg.duration.Seconds(), errStr)
	logf("fin kbps=%.0f perdidos=%d err=%q",
		rep.Steady.TotalKbps, rep.Transport.Video.PacketsLost+rep.Transport.Audio.PacketsLost, errStr)
	if errStr != "" {
		return rep, 1
	}
	return rep, 0
}

type multiSummary struct {
	Succeeded      int      `json:"succeeded"`
	Failed         int      `json:"failed"`
	AvgSignalingMs *float64 `json:"avg_signaling_ms"`
	AvgICEestMs    *float64 `json:"avg_ice_establishment_ms"`
	TotalKbps      float64  `json:"total_kbps"`
	PacketsSent    uint64   `json:"packets_sent"`
	PacketsLost    uint64   `json:"packets_lost"`
}

type multiReport struct {
	TargetBase string       `json:"target_base"`
	StartedAt  string       `json:"started_at"`
	Users      int          `json:"users"`
	Summary    multiSummary `json:"summary"`
	Results    []Report     `json:"results"`
}

func summarize(results []Report) multiSummary {
	s := multiSummary{}
	var sigSum, iceSum float64
	var sigN, iceN int
	for _, r := range results {
		if r.Error != "" {
			s.Failed++
			continue
		}
		s.Succeeded++
		sigSum += r.Signaling.LatencyMs
		sigN++
		if r.ICE.EstablishmentMs != nil {
			iceSum += *r.ICE.EstablishmentMs
			iceN++
		}
		s.TotalKbps += r.Steady.TotalKbps
		s.PacketsSent += r.Transport.Video.PacketsSent + r.Transport.Audio.PacketsSent
		s.PacketsLost += r.Transport.Video.PacketsLost + r.Transport.Audio.PacketsLost
	}
	if sigN > 0 {
		v := sigSum / float64(sigN)
		s.AvgSignalingMs = &v
	}
	if iceN > 0 {
		v := iceSum / float64(iceN)
		s.AvgICEestMs = &v
	}
	return s
}

func main() {
	whipBase := flag.String("whip", "http://localhost:8889", "base WHIP")
	stream := flag.String("stream", "test", "stream_id (modo 1 usuario)")
	users := flag.Int("users", 1, "usuarios concurrentes (streams vu1..vuN)")
	out := flag.String("out", "", "archivo JSON de resultados (modo multiusuario)")
	videoPath := flag.String("video", "capacity/testdata/video.h264", "h264 annex-b 720p")
	audioPath := flag.String("audio", "capacity/testdata/audio.ogg", "opus en ogg")
	duration := flag.Duration("duration", 10*time.Second, "segundos de transmisión")
	iceTimeout := flag.Duration("ice-timeout", 8*time.Second, "timeout de checks ICE")
	statsEvery := flag.Duration("stats-interval", 500*time.Millisecond, "muestreo de contadores")
	stun := flag.String("stun", "", "stun url opcional")
	dumpSDP := flag.String("dump-sdp", "", "prefijo de archivos para guardar offer/answer SDP (debug)")
	flag.Parse()

	die := func(code int, format string, args ...any) {
		fmt.Fprintf(os.Stderr, "capacity: "+format+"\n", args...)
		os.Exit(code)
	}

	// Archivos primero: fallar antes de tocar la red.
	frames, err := parseH264File(*videoPath)
	if err != nil {
		die(2, "video: %s", err.Error())
	}
	opusPkts, err := parseOpusOgg(*audioPath)
	if err != nil {
		die(2, "audio: %s", err.Error())
	}
	cfg := runConfig{
		whipBase: *whipBase, duration: *duration, iceTimeout: *iceTimeout,
		statsEvery: *statsEvery, stun: *stun, frames: frames, opusPkts: opusPkts,
		dumpSDP: *dumpSDP,
	}

	// Modo 1 usuario: JSON a stdout (comportamiento original).
	if *users <= 1 && *out == "" {
		rep, code := runPublisher(cfg, *stream, nil)
		printReport(rep)
		os.Exit(code)
	}

	// Modo multiusuario: ráfaga concurrente, JSON a archivo.
	n := *users
	if n < 1 {
		n = 1
	}
	if *out == "" {
		*out = "results-" + time.Now().Format("20060102-150405") + ".json"
	}
	progress := func(format string, args ...any) {
		fmt.Fprintf(os.Stderr, "capacity: "+format+"\n", args...)
	}
	results := make([]Report, n)
	var wg sync.WaitGroup
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			rep, _ := runPublisher(cfg, fmt.Sprintf("vu%d", i+1), progress)
			results[i] = rep
		}(i)
	}
	wg.Wait()

	mr := multiReport{
		TargetBase: strings.TrimRight(*whipBase, "/"),
		StartedAt:  time.Now().Format(time.RFC3339),
		Users:      n,
		Summary:    summarize(results),
		Results:    results,
	}
	raw, _ := json.MarshalIndent(mr, "", "  ")
	if err := os.WriteFile(*out, raw, 0o600); err != nil {
		die(1, "escribiendo %s: %s", *out, err.Error())
	}
	progress("resultados en %s (%d/%d ok)", *out, mr.Summary.Succeeded, n)
}

func cleanup(c *http.Client, location string, pc *webrtc.PeerConnection, statsDone chan struct{}, wg *sync.WaitGroup) {
	whipDelete(c, location)
	close(statsDone)
	_ = pc.Close()
	wg.Wait()
}

func printReport(rep Report) {
	out, _ := json.MarshalIndent(rep, "", "  ")
	fmt.Println(string(out))
}
