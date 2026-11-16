package bbr

import (
	"testing"
	"time"

	"github.com/quic-go/quic-go/monotime"
)

func TestSkipSimilarProbeRtt(t *testing.T) {
	for _, tc := range []struct {
		skip, appLimited bool
		sample           time.Duration
		expired          bool
	}{
		{false, true, 82 * time.Millisecond, true},
		{true, true, 82 * time.Millisecond, false},
		{true, false, 82 * time.Millisecond, true},
		{true, true, 95 * time.Millisecond, true},
	} {
		b := NewBbrSender(DefaultClock{}, 1200, ProfileStandard)
		b.SkipSimilarProbeRtt = tc.skip
		t0 := monotime.Now()
		b.maybeUpdateMinRtt(t0, 80*time.Millisecond)
		b.appLimitedSinceLastProbeRtt = tc.appLimited
		b.maybeUpdateMinRtt(t0.Add(5*time.Second), tc.sample)
		if got := b.maybeUpdateMinRtt(t0.Add(11*time.Second), tc.sample); got != tc.expired {
			t.Errorf("%+v: expired %v", tc, got)
		}
		if !tc.expired && b.minRtt != 80*time.Millisecond {
			t.Errorf("%+v: min_rtt %v", tc, b.minRtt)
		}
	}
}

func TestProbeRttHalfBdp(t *testing.T) {
	b := NewBbrSender(DefaultClock{}, 1200, ProfileStandard)
	b.minRtt = 40 * time.Millisecond
	b.maxBandwidth.Update(Bandwidth(100e6), b.roundTripCount) // 100 Mbit/s: BDP 500 kB
	if got := b.probeRttCongestionWindow(); got != b.minCongestionWindow {
		t.Errorf("default: %d, want %d", got, b.minCongestionWindow)
	}
	b.ProbeRttHalfBdp = true
	if got := b.probeRttCongestionWindow(); got != 250_000 {
		t.Errorf("half BDP: %d, want 250000", got)
	}
	b.mode = bbrModeProbeRtt
	b.congestionWindow = 100_000 // e.g. after an RTO
	if got := b.GetCongestionWindow(); got != 100_000 {
		t.Errorf("PROBE_RTT raised the window to %d", got)
	}
}
