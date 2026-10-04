// Tap RTP/RTCP: interceptor mínimo que cuenta paquetes/bytes salientes por
// SSRC y registra los Sender Reports propios para calcular RTT (LSR/DLSR).
// GetStats() de Pion no expone outbound-rtp en un PC sendonly, así que se
// mide aquí, en el path real de los paquetes.
package main

import (
	"strings"
	"time"

	"github.com/pion/interceptor"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
)

type tapFactory struct{ m *Metrics }

func (f *tapFactory) NewInterceptor(_ string) (interceptor.Interceptor, error) {
	return &tap{m: f.m}, nil
}

type tap struct{ m *Metrics }

func (t *tap) BindLocalStream(info *interceptor.StreamInfo, writer interceptor.RTPWriter) interceptor.RTPWriter {
	kind := kindFromMime(info.MimeType)
	t.m.bindSSRC(info.SSRC, kind)
	if info.SSRCRetransmission != 0 {
		t.m.bindSSRC(info.SSRCRetransmission, kind)
	}
	return interceptor.RTPWriterFunc(func(h *rtp.Header, p []byte, a interceptor.Attributes) (int, error) {
		t.m.countRTP(h.SSRC, h.MarshalSize()+len(p))
		return writer.Write(h, p, a)
	})
}

func (t *tap) BindRemoteStream(_ *interceptor.StreamInfo, reader interceptor.RTPReader) interceptor.RTPReader {
	return reader
}

func (t *tap) BindRTCPReader(reader interceptor.RTCPReader) interceptor.RTCPReader { return reader }

// BindRTCPWriter ve los Sender Reports que genera el report-interceptor.
// El tap debe registrarse ANTES que el resto de interceptores para que los
// SR pasen por aquí (el report-interceptor escribe en el writer que recibe).
func (t *tap) BindRTCPWriter(writer interceptor.RTCPWriter) interceptor.RTCPWriter {
	return interceptor.RTCPWriterFunc(func(pkts []rtcp.Packet, a interceptor.Attributes) (int, error) {
		now := time.Now()
		for _, pkt := range pkts {
			if sr, ok := pkt.(*rtcp.SenderReport); ok {
				// Middle-32 del NTP (RFC 3550 §6.4.1), matchea el LSR de los RR.
				t.m.trackSR(sr.SSRC, uint32((sr.NTPTime>>16)&0xFFFFFFFF), now) // #nosec G115 -- truncado intencional al middle-32
			}
		}
		return writer.Write(pkts, a)
	})
}

func (t *tap) UnbindLocalStream(_ *interceptor.StreamInfo)  {}
func (t *tap) UnbindRemoteStream(_ *interceptor.StreamInfo) {}
func (t *tap) Close() error                                 { return nil }

func kindFromMime(mt string) string {
	mt = strings.ToLower(mt)
	if strings.HasPrefix(mt, "video") {
		return "video"
	}
	if strings.HasPrefix(mt, "audio") {
		return "audio"
	}
	return ""
}
