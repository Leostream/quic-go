package ackhandler

import (
	"github.com/quic-go/quic-go/congestion"
	cgInternal "github.com/quic-go/quic-go/internal/congestion"
	"github.com/quic-go/quic-go/internal/monotime"
	"github.com/quic-go/quic-go/internal/protocol"
	"github.com/quic-go/quic-go/internal/utils"
)

var (
	_ cgInternal.SendAlgorithmWithDebugInfos = &ccAdapter{}
	_ cgInternal.SendAlgorithmEx             = &ccAdapterEx{}
)

// ccAdapter runs a public CongestionControl as quic-go's SendAlgorithm. It
// keeps the connection stats that quic-go's own sender would otherwise keep.
type ccAdapter struct {
	cc        congestion.CongestionControl
	connStats *utils.ConnectionStats
}

type ccAdapterEx struct {
	ccAdapter
	ex congestion.CongestionControlEx
}

func newCCAdapter(cc congestion.CongestionControl, connStats *utils.ConnectionStats) cgInternal.SendAlgorithmWithDebugInfos {
	a := ccAdapter{cc: cc, connStats: connStats}
	a.storeWindow()
	if ex, ok := cc.(congestion.CongestionControlEx); ok {
		return &ccAdapterEx{a, ex}
	}
	return &a
}

func (a *ccAdapter) storeWindow() {
	a.connStats.CongestionWindow.Store(uint64(a.cc.GetCongestionWindow()))
}

func (a *ccAdapter) TimeUntilSend(bytesInFlight protocol.ByteCount) monotime.Time {
	return a.cc.TimeUntilSend(congestion.ByteCount(bytesInFlight))
}

func (a *ccAdapter) HasPacingBudget(now monotime.Time) bool {
	return a.cc.HasPacingBudget(now)
}

func (a *ccAdapter) OnPacketSent(sentTime monotime.Time, bytesInFlight protocol.ByteCount, packetNumber protocol.PacketNumber, bytes protocol.ByteCount, isRetransmittable bool) {
	a.cc.OnPacketSent(sentTime, congestion.ByteCount(bytesInFlight), congestion.PacketNumber(packetNumber), congestion.ByteCount(bytes), isRetransmittable)
}

func (a *ccAdapter) CanSend(bytesInFlight protocol.ByteCount) bool {
	return a.cc.CanSend(congestion.ByteCount(bytesInFlight))
}

func (a *ccAdapter) MaybeExitSlowStart() {
	a.cc.MaybeExitSlowStart()
}

func (a *ccAdapter) OnPacketAcked(number protocol.PacketNumber, ackedBytes protocol.ByteCount, priorInFlight protocol.ByteCount, eventTime monotime.Time) {
	a.cc.OnPacketAcked(congestion.PacketNumber(number), congestion.ByteCount(ackedBytes), congestion.ByteCount(priorInFlight), eventTime)
	a.storeWindow()
}

// OnCongestionEvent counts the loss as quic-go's own sender does.
func (a *ccAdapter) OnCongestionEvent(number protocol.PacketNumber, lostBytes protocol.ByteCount, priorInFlight protocol.ByteCount) {
	a.connStats.PacketsLost.Add(1)
	a.connStats.BytesLost.Add(uint64(lostBytes))
	a.cc.OnCongestionEvent(congestion.PacketNumber(number), congestion.ByteCount(lostBytes), congestion.ByteCount(priorInFlight))
	a.storeWindow()
}

// OnSpuriousLoss is a no-op: the public interface has no hook for undoing a
// congestion event after a loss turns out to be spurious.
func (a *ccAdapter) OnSpuriousLoss(protocol.PacketNumber) {}

func (a *ccAdapter) OnRetransmissionTimeout(packetsRetransmitted bool) {
	a.cc.OnRetransmissionTimeout(packetsRetransmitted)
	a.storeWindow()
}

func (a *ccAdapter) SetMaxDatagramSize(size protocol.ByteCount) {
	a.cc.SetMaxDatagramSize(congestion.ByteCount(size))
	a.storeWindow()
}

func (a *ccAdapter) InSlowStart() bool {
	return a.cc.InSlowStart()
}

func (a *ccAdapter) InRecovery() bool {
	return a.cc.InRecovery()
}

func (a *ccAdapter) GetCongestionWindow() protocol.ByteCount {
	return protocol.ByteCount(a.cc.GetCongestionWindow())
}

func (a *ccAdapterEx) OnCongestionEventEx(priorInFlight protocol.ByteCount, eventTime monotime.Time, ackedPackets []congestion.AckedPacketInfo, lostPackets []congestion.LostPacketInfo) {
	a.ex.OnCongestionEventEx(congestion.ByteCount(priorInFlight), eventTime, ackedPackets, lostPackets)
	a.storeWindow()
}
