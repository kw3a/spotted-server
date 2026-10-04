// Métricas por etapa de una publicación WHIP con Pion (sin navegador).
package main

import (
	"sync"
	"time"

	"github.com/pion/webrtc/v4"
)

// signaling: latencia HTTP, errores y rechazo de SDP.
type signalingMetrics struct {
	LatencyMs   float64 `json:"latency_ms"`
	HTTPStatus  int     `json:"http_status"`
	HTTPError   string  `json:"http_error,omitempty"`
	SDPRejected bool    `json:"sdp_rejected"`
	SDPDetail   string  `json:"sdp_detail,omitempty"`
}

type iceMetrics struct {
	EstablishmentMs *float64 `json:"establishment_ms"`
	Failed          bool     `json:"failed"`
	Timeout         bool     `json:"connectivity_check_timeout"`
	FinalState      string   `json:"final_state"`
}

type dtlsMetrics struct {
	HandshakeMs *float64 `json:"handshake_ms"`
	Failed      bool     `json:"failed"`
	Detail      string   `json:"detail,omitempty"`
}

type kindTransport struct {
	PacketsSent uint64   `json:"packets_sent"`
	PacketsLost uint64   `json:"packets_lost"`
	LossRatio   float64  `json:"loss_ratio"`
	RTTMs       *float64 `json:"rtt_ms"`
	JitterMs    *float64 `json:"jitter_ms"`
}

type transportMetrics struct {
	FirstRTPMs          *float64      `json:"first_rtp_ms"`
	FirstRTPAfterDTLSMs *float64      `json:"first_rtp_after_dtls_ms"`
	Video               kindTransport `json:"video"`
	Audio               kindTransport `json:"audio"`
	RRReceived          uint64        `json:"rr_received"`
}

type stallEvent struct {
	StartMs    float64 `json:"start_ms"`
	DurationMs float64 `json:"duration_ms"`
}

type steadyMetrics struct {
	DurationS   float64      `json:"duration_s"`
	VideoKbps   float64      `json:"video_kbps"`
	AudioKbps   float64      `json:"audio_kbps"`
	TotalKbps   float64      `json:"total_kbps"`
	NACKCount   uint64       `json:"nack_count"`
	NACKRate    float64      `json:"nack_rate_per_s"`
	PLICount    uint64       `json:"pli_count"`
	RepeatedPLI uint64       `json:"repeated_pli"`
	FIRCount    uint64       `json:"fir_count"`
	Stalls      []stallEvent `json:"stalls"`
}

// Report es el JSON final con las 5 etapas.
type Report struct {
	Target    string           `json:"target"`
	Signaling signalingMetrics `json:"signaling"`
	ICE       iceMetrics       `json:"ice"`
	DTLS      dtlsMetrics      `json:"dtls"`
	Transport transportMetrics `json:"transport"`
	Steady    steadyMetrics    `json:"steady"`
	Error     string           `json:"error,omitempty"`
}

type statsSample struct {
	t       time.Time
	bytes   uint64
	packets uint64
}

// Metrics acumula eventos con mutex; t0 = instante de SetRemoteDescription.
type Metrics struct {
	mu sync.Mutex
	t0 time.Time // referencia: answer aplicada

	sig signalingMetrics

	iceCheckingAt  time.Time
	iceConnectedAt time.Time
	iceFailed      bool
	iceTimeout     bool
	iceFinal       string

	pcConnectedAt time.Time
	pcFailed      bool
	pcFailedDesc  string

	firstRTPAt time.Time

	// RTCP en vivo (por sender).
	nackMsgs  uint64
	nackPkts  uint64
	pli       uint64
	pliRepeat uint64
	fir       uint64
	rr        uint64
	lastPLIAt time.Time
	steadyT0  time.Time // inicio de ventana steady (primer RTP)

	// Salida real (tap RTP) por SSRC + estado RR por SSRC.
	sentPkts map[uint32]uint64
	sentBits map[uint32]uint64
	ssrcKind map[uint32]string
	srMarks  map[uint32][]srMark // últimos SR propios por SSRC (para RTT)
	rrState  map[uint32]*rrSource

	samples []statsSample
}

// srMark registra cuándo se envió un Sender Report (middle-32 del NTP).
type srMark struct {
	lsr uint32
	at  time.Time
}

// rrSource acumula bloques RR de un SSRC (RFC 3550 §A.3).
type rrSource struct {
	baseSeq uint32 // primer extended-highest-seq visto
	maxSeq  uint32 // mayor extended-highest-seq visto
	lost    uint64
	jitter  uint32
	updated time.Time
	rttSec  float64
	hasRTT  bool
}

func newMetrics() *Metrics {
	return &Metrics{
		sentPkts: map[uint32]uint64{},
		sentBits: map[uint32]uint64{},
		ssrcKind: map[uint32]string{},
		srMarks:  map[uint32][]srMark{},
		rrState:  map[uint32]*rrSource{},
	}
}

func ms(d time.Duration) float64 { return float64(d.Nanoseconds()) / 1e6 }

// --- señalización ---

func (m *Metrics) setSignaling(lat time.Duration, status int, httpErr string, rejected bool, detail string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sig = signalingMetrics{LatencyMs: ms(lat), HTTPStatus: status, HTTPError: httpErr, SDPRejected: rejected, SDPDetail: detail}
}

// --- ICE / DTLS ---

func (m *Metrics) setT0(t time.Time) { m.mu.Lock(); m.t0 = t; m.mu.Unlock() }

func (m *Metrics) onICE(s webrtc.ICEConnectionState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.iceFinal = s.String()
	switch s {
	case webrtc.ICEConnectionStateChecking:
		if m.iceCheckingAt.IsZero() {
			m.iceCheckingAt = time.Now()
		}
	case webrtc.ICEConnectionStateConnected, webrtc.ICEConnectionStateCompleted:
		if m.iceConnectedAt.IsZero() {
			m.iceConnectedAt = time.Now()
		}
	case webrtc.ICEConnectionStateFailed:
		m.iceFailed = true
	}
}

func (m *Metrics) setICETimeout() { m.mu.Lock(); m.iceTimeout = true; m.mu.Unlock() }

func (m *Metrics) onPC(s webrtc.PeerConnectionState, t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s == webrtc.PeerConnectionStateConnected && m.pcConnectedAt.IsZero() {
		m.pcConnectedAt = t
	}
	if s == webrtc.PeerConnectionStateFailed {
		m.pcFailed = true
		m.pcFailedDesc = "peerconnection failed"
	}
}

// --- media / RTCP ---

func (m *Metrics) markFirstRTP(t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.firstRTPAt.IsZero() {
		m.firstRTPAt = t
		m.steadyT0 = t
	}
}

func (m *Metrics) addNACK(nPkts uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nackMsgs++
	m.nackPkts += nPkts
}

func (m *Metrics) addPLI(t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// ponytail: repetido = PLI con <500ms desde el anterior.
	if !m.lastPLIAt.IsZero() && t.Sub(m.lastPLIAt) < 500*time.Millisecond {
		m.pliRepeat++
	}
	m.lastPLIAt = t
	m.pli++
}

func (m *Metrics) addFIR() { m.mu.Lock(); m.fir++; m.mu.Unlock() }
func (m *Metrics) addRR()  { m.mu.Lock(); m.rr++; m.mu.Unlock() }

// --- tap RTP/RTCP ---

func (m *Metrics) bindSSRC(ssrc uint32, kind string) {
	if kind == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ssrcKind[ssrc] = kind
}

func (m *Metrics) countRTP(ssrc uint32, sizeBytes int) {
	if sizeBytes < 0 {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sentPkts[ssrc]++
	m.sentBits[ssrc] += uint64(sizeBytes) // #nosec G115 -- acotado a no-negativo arriba
}

func (m *Metrics) trackSR(ssrc, lsr uint32, at time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	marks := append(m.srMarks[ssrc], srMark{lsr: lsr, at: at})
	if len(marks) > 16 {
		marks = marks[len(marks)-16:]
	}
	m.srMarks[ssrc] = marks
}

// addRRBlock procesa un bloque de Receiver Report (RFC 3550 §6.4.1):
// pérdida acumulada, jitter y RTT vía LSR/DLSR contra nuestros SR.
func (m *Metrics) addRRBlock(ssrc uint32, totalLost, extSeq, jitter, lsr, dlsr uint32, arrival time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	st := m.rrState[ssrc]
	if st == nil {
		st = &rrSource{baseSeq: extSeq, maxSeq: extSeq}
		m.rrState[ssrc] = st
	}
	if extSeq < st.baseSeq {
		st.baseSeq = extSeq
	}
	if extSeq > st.maxSeq {
		st.maxSeq = extSeq
	}
	st.lost = uint64(totalLost & 0xFFFFFF)
	st.jitter = jitter
	st.updated = arrival
	if lsr != 0 {
		for _, mk := range m.srMarks[ssrc] {
			if mk.lsr == lsr {
				if rtt := arrival.Sub(mk.at).Seconds() - float64(dlsr)/65536.0; rtt >= 0 {
					st.rttSec = rtt
					st.hasRTT = true
				}
				break
			}
		}
	}
}

// sampleLoop muestrea los contadores del tap cada interval hasta cerrar done.
func (m *Metrics) sampleLoop(interval time.Duration, done <-chan struct{}) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-done:
			return
		case now := <-t.C:
			m.mu.Lock()
			var b, p uint64
			for _, v := range m.sentBits {
				b += v
			}
			for _, v := range m.sentPkts {
				p += v
			}
			m.samples = append(m.samples, statsSample{t: now, bytes: b, packets: p})
			m.mu.Unlock()
		}
	}
}

// --- reporte ---

func kindReport(m *Metrics, kind string) kindTransport {
	out := kindTransport{}
	var expected uint64
	var lastJitter time.Time
	var lastRTT time.Time
	for ssrc, k := range m.ssrcKind {
		if k != kind {
			continue
		}
		out.PacketsSent += m.sentPkts[ssrc]
		st := m.rrState[ssrc]
		if st == nil {
			continue
		}
		out.PacketsLost += st.lost
		if st.maxSeq >= st.baseSeq {
			expected += uint64(st.maxSeq-st.baseSeq) + 1
		}
		if st.updated.After(lastJitter) {
			lastJitter = st.updated
			v := float64(st.jitter) * 1000 / float64(clockRate(kind))
			out.JitterMs = &v
		}
		if st.hasRTT && st.updated.After(lastRTT) {
			lastRTT = st.updated
			v := st.rttSec * 1000
			out.RTTMs = &v
		}
	}
	if expected > 0 {
		out.LossRatio = float64(out.PacketsLost) / float64(expected)
		if out.LossRatio > 1 {
			out.LossRatio = 1
		}
	}
	return out
}

func clockRate(kind string) uint32 {
	if kind == "video" {
		return 90000
	}
	return 48000
}

func (m *Metrics) report(target string, steadyS float64, errStr string) Report {
	m.mu.Lock()
	defer m.mu.Unlock()
	r := Report{Target: target, Error: errStr}
	r.Signaling = m.sig

	if !m.iceConnectedAt.IsZero() && !m.t0.IsZero() {
		v := ms(m.iceConnectedAt.Sub(m.t0))
		r.ICE.EstablishmentMs = &v
	}
	r.ICE.Failed = m.iceFailed
	r.ICE.Timeout = m.iceTimeout
	r.ICE.FinalState = m.iceFinal

	if !m.pcConnectedAt.IsZero() {
		base := m.iceConnectedAt
		if base.IsZero() {
			base = m.t0
		}
		v := ms(m.pcConnectedAt.Sub(base))
		r.DTLS.HandshakeMs = &v
	}
	r.DTLS.Failed = m.pcFailed
	r.DTLS.Detail = m.pcFailedDesc

	if !m.firstRTPAt.IsZero() && !m.t0.IsZero() {
		v := ms(m.firstRTPAt.Sub(m.t0))
		r.Transport.FirstRTPMs = &v
	}
	if !m.firstRTPAt.IsZero() && !m.pcConnectedAt.IsZero() {
		v := ms(m.firstRTPAt.Sub(m.pcConnectedAt))
		if v < 0 {
			v = 0
		}
		r.Transport.FirstRTPAfterDTLSMs = &v
	}
	r.Transport.Video = kindReport(m, "video")
	r.Transport.Audio = kindReport(m, "audio")
	r.Transport.RRReceived = m.rr

	// Steady: bitrate sin el primer segundo (rampa), stalls = huecos >=1s sin bytes.
	r.Steady.DurationS = steadyS
	winA, winB := statsSample{}, statsSample{}
	hasWin := false
	if len(m.samples) >= 2 {
		start := 0
		cutoff := m.samples[0].t.Add(time.Second)
		for i, s := range m.samples {
			if !s.t.Before(cutoff) {
				start = i
				break
			}
			start = i
		}
		a, b := m.samples[start], m.samples[len(m.samples)-1]
		winA, winB, hasWin = a, b, true
		if dt := b.t.Sub(a.t).Seconds(); dt > 0 {
			r.Steady.TotalKbps = float64(b.bytes-a.bytes) * 8 / dt / 1000
		}
		var lastProg time.Time
		var stallOpen time.Time
		for i, s := range m.samples {
			if i == 0 || s.bytes > m.samples[i-1].bytes {
				if !stallOpen.IsZero() {
					r.Steady.Stalls = append(r.Steady.Stalls, stallEvent{
						StartMs:    ms(stallOpen.Sub(m.t0)),
						DurationMs: ms(s.t.Sub(stallOpen)),
					})
					stallOpen = time.Time{}
				}
				lastProg = s.t
				continue
			}
			if lastProg.IsZero() {
				lastProg = m.samples[0].t
			}
			if s.t.Sub(lastProg) >= time.Second && stallOpen.IsZero() {
				stallOpen = lastProg
			}
		}
		if !stallOpen.IsZero() {
			last := m.samples[len(m.samples)-1].t
			r.Steady.Stalls = append(r.Steady.Stalls, stallEvent{
				StartMs:    ms(stallOpen.Sub(m.t0)),
				DurationMs: ms(last.Sub(stallOpen)),
			})
		}
	}
	if r.Steady.Stalls == nil {
		r.Steady.Stalls = []stallEvent{}
	}
	r.Steady.VideoKbps, r.Steady.AudioKbps = m.kindBitrates(winA, winB, hasWin)
	r.Steady.NACKCount = m.nackPkts
	if steadyS > 0 {
		r.Steady.NACKRate = float64(m.nackPkts) / steadyS
	}
	r.Steady.PLICount = m.pli
	r.Steady.RepeatedPLI = m.pliRepeat
	r.Steady.FIRCount = m.fir
	return r
}

// kindBitrates reparte el bitrate de la ventana [a,b] según bytes por kind.
func (m *Metrics) kindBitrates(a, b statsSample, ok bool) (video, audio float64) {
	var vBytes, aBytes uint64
	for ssrc, k := range m.ssrcKind {
		switch k {
		case "video":
			vBytes += m.sentBits[ssrc]
		case "audio":
			aBytes += m.sentBits[ssrc]
		}
	}
	total := vBytes + aBytes
	if !ok || total == 0 {
		return 0, 0
	}
	dt := b.t.Sub(a.t).Seconds()
	if dt <= 0 {
		return 0, 0
	}
	kbps := float64(b.bytes-a.bytes) * 8 / dt / 1000
	video = kbps * float64(vBytes) / float64(total)
	audio = kbps - video
	return video, audio
}
