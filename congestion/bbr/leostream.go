package bbr

import (
	"fmt"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/congestion"
)

// Set installs a BBR variant on conn:
//   - bbr: BBRv1 as in Chromium;
//   - bbr-skiprtt: skips PROBE_RTT while the sender is app-limited at an RTT
//     near min_rtt, so video and interactive traffic don't stall every 10 s;
//   - bbr-leo: bbr-skiprtt, and PROBE_RTT at half the BDP rather than 4
//     packets, so a sender that fills the link keeps sending at half rate.
func Set(conn *quic.Conn, variant string) error {
	var skip, half bool
	switch variant {
	case "bbr":
	case "bbr-skiprtt":
		skip = true
	case "bbr-leo":
		skip, half = true, true
	default:
		return fmt.Errorf("unknown BBR variant %q", variant)
	}
	conn.SetCongestionControl(func(size congestion.ByteCount) congestion.CongestionControl {
		s := NewBbrSender(DefaultClock{}, size, ProfileStandard)
		s.SkipSimilarProbeRtt = skip
		s.ProbeRttHalfBdp = half
		return s
	})
	return nil
}
