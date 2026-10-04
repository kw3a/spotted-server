// Lectura de archivos locales y envío continuo (sin generar nada).
package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
)

const videoFPS = 30

var frameDur = time.Second / videoFPS

// --- H264 Annex B (stdlib, sin depender de h264reader) ---

func nalType(nal []byte) byte { return nal[0] & 0x1F }

// splitAnnexB corta por start codes 0x000001 / 0x00000001.
func splitAnnexB(buf []byte) [][]byte {
	var nals [][]byte
	start := -1
	for i := 0; i+3 <= len(buf); i++ {
		n := 0
		if buf[i] == 0 && buf[i+1] == 0 {
			if buf[i+2] == 1 {
				n = 3
			} else if i+3 < len(buf) && buf[i+2] == 0 && buf[i+3] == 1 {
				n = 4
			}
		}
		if n == 0 {
			continue
		}
		if start >= 0 && i > start {
			nals = append(nals, buf[start:i])
		}
		start = i + n
		i += n - 1
	}
	if start >= 0 && start < len(buf) {
		nals = append(nals, buf[start:])
	}
	return nals
}

// groupAUs agrupa NALUs en unidades de acceso (1 sample = 1 frame).
// Con AUDs (tipo 9) el corte es exacto; sin AUDs, corta antes de cada
// slice (1-5) si el grupo ya trae slice.
func groupAUs(nals [][]byte) [][][]byte {
	hasAUD := false
	for _, n := range nals {
		if len(n) > 0 && nalType(n) == 9 {
			hasAUD = true
			break
		}
	}
	var aus [][][]byte
	var cur [][]byte
	flush := func() {
		if len(cur) > 0 {
			aus = append(aus, cur)
			cur = nil
		}
	}
	isSlice := func(t byte) bool { return t >= 1 && t <= 5 }
	for _, n := range nals {
		if len(n) == 0 {
			continue
		}
		t := nalType(n)
		if hasAUD {
			if t == 9 {
				flush()
			}
			cur = append(cur, n)
			continue
		}
		if isSlice(t) {
			for _, c := range cur {
				if isSlice(nalType(c)) {
					flush()
					break
				}
			}
		}
		cur = append(cur, n)
	}
	flush()
	return aus
}

func parseH264File(path string) ([][][]byte, error) {
	raw, err := os.ReadFile(path) // #nosec G304 -- harness de prueba, la ruta viene de flag a propósito
	if err != nil {
		return nil, err
	}
	aus := groupAUs(splitAnnexB(raw))
	if len(aus) == 0 {
		return nil, fmt.Errorf("sin NALUs en %s", path)
	}
	return aus, nil
}

// --- Opus en Ogg (parser mínimo stdlib: páginas -> paquetes) ---

type opusPacket struct {
	data []byte
	dur  time.Duration
}

func parseOpusOgg(path string) ([]opusPacket, error) {
	raw, err := os.ReadFile(path) // #nosec G304 -- harness de prueba, la ruta viene de flag a propósito
	if err != nil {
		return nil, err
	}
	var (
		p          int
		serial     uint32
		haveSerial bool
		partial    []byte
		prevGran   uint64
		pktIdx     int
		out        []opusPacket
	)
	for p+27 <= len(raw) {
		pageStart := p
		if string(raw[p:p+4]) != "OggS" {
			return nil, fmt.Errorf("página ogg inválida en offset %d", p)
		}
		granule := binary.LittleEndian.Uint64(raw[p+6 : p+14])
		nSeg := int(raw[p+26])
		if p+27+nSeg > len(raw) {
			return nil, errors.New("tabla de segmentos truncada")
		}
		segTable := raw[p+27 : p+27+nSeg]
		payloadLen := 0
		for _, s := range segTable {
			payloadLen += int(s)
		}
		head := p + 27 + nSeg
		if head+payloadLen > len(raw) {
			return nil, errors.New("payload ogg truncado")
		}
		payload := raw[head : head+payloadLen]
		p = head + payloadLen

		if !haveSerial {
			serial = binary.LittleEndian.Uint32(raw[pageStart+14 : pageStart+18])
			haveSerial = true
		}
		// Una sola secuencia: ignora páginas de otro serial.
		if binary.LittleEndian.Uint32(raw[pageStart+14:pageStart+18]) != serial {
			continue
		}

		const noGranule = ^uint64(0)
		var done [][]byte
		off := 0
		for _, s := range segTable {
			seg := payload[off : off+int(s)]
			off += int(s)
			partial = append(partial, seg...)
			if s < 255 {
				done = append(done, partial)
				partial = nil
			}
		}
		audioPkts := 0
		for _, d := range done {
			if pktIdx < 2 { // OpusHead + OpusTags
				pktIdx++
				continue
			}
			cp := append([]byte(nil), d...)
			out = append(out, opusPacket{data: cp})
			audioPkts++
			pktIdx++
		}
		if audioPkts > 0 {
			dur := 20 * time.Millisecond
			if granule != noGranule && granule > prevGran && prevGran != 0 {
				// Granule en muestras @48kHz; clamp a (0,10s] para que
				// la conversión a time.Duration no pueda desbordar.
				secs := float64(granule-prevGran) / float64(audioPkts) / 48000.0
				if secs > 0 && secs <= 10 {
					dur = time.Duration(secs * float64(time.Second))
				}
			} else if granule != noGranule {
				prevGran = granule
			}
			for i := len(out) - audioPkts; i < len(out); i++ {
				out[i].dur = dur
			}
			if granule != noGranule {
				prevGran = granule
			}
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("sin paquetes opus en %s", path)
	}
	return out, nil
}

// --- loops de envío ---

func sendVideoLoop(ctx context.Context, track *webrtc.TrackLocalStaticSample, aus [][][]byte, m *Metrics) error {
	t := time.NewTicker(frameDur)
	defer t.Stop()
	i := 0
	for {
		select {
		case <-ctx.Done():
			return nil
		case now := <-t.C:
			// 1 sample = 1 frame: el payloader de Pion corta por start
			// codes, arma STAP-A con SPS+PPS y descarta AUDs. Un WriteSample
			// por NAL rompía los timestamps (33ms por NAL en vez de por frame).
			frame := joinAU(aus[i%len(aus)])
			if err := track.WriteSample(media.Sample{Data: frame, Duration: frameDur}); err != nil {
				if ctx.Err() != nil {
					return nil
				}
				return err
			}
			m.markFirstRTP(now)
			i++
		}
	}
}

// joinAU une las NALUs de un access unit con start codes Annex-B.
func joinAU(nals [][]byte) []byte {
	var frame []byte
	for _, n := range nals {
		frame = append(frame, 0x00, 0x00, 0x00, 0x01)
		frame = append(frame, n...)
	}
	return frame
}

func sendAudioLoop(ctx context.Context, track *webrtc.TrackLocalStaticSample, pkts []opusPacket, m *Metrics) error {
	i := 0
	for {
		pkt := pkts[i%len(pkts)]
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(pkt.dur):
		}
		if err := track.WriteSample(media.Sample{Data: pkt.data, Duration: pkt.dur}); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		m.markFirstRTP(time.Now())
		i++
	}
}
