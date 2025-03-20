package Node

import (
	"time"
	"github.com/Decentralized-TrueTime-Sync/Helper"
)

// NewHLC initializes a new Hybrid Logical Clock with a default uncertainty bound
func NewHLC(clockDrift time.Duration) *HLC {
    // Add artificial clock drift to simulate real-world conditions
    driftedTime := time.Now().Add(clockDrift)
    return &HLC{
        PhysicalTime:   driftedTime,
        LogicalCounter: 0,
        Uncertainty:    5 * time.Millisecond,
    }
}

// GenerateTimestamp generates a new timestamp based on local time
func (hlc *HLC) GenerateTimestamp() HLC {
    hlc.Mutex.Lock()
    defer hlc.Mutex.Unlock()

    now := time.Now()

    if now.After(hlc.PhysicalTime) {
        hlc.PhysicalTime = now
        hlc.LogicalCounter = 0
    } else {
        hlc.LogicalCounter++
    }

    return HLC{
        PhysicalTime:   hlc.PhysicalTime,
        LogicalCounter: hlc.LogicalCounter,
        Uncertainty:    hlc.Uncertainty,
    }
}

// MergeTimestamps synchronizes with another node
func (hlc *HLC) MergeTimestamps(peerHLC HLC) HLC {
    hlc.Mutex.Lock()
    defer hlc.Mutex.Unlock()

    now := time.Now()
    maxTime := Helper.MaxTime(now, peerHLC.PhysicalTime, hlc.PhysicalTime)
    logicalCounter := max(hlc.LogicalCounter, peerHLC.LogicalCounter)

    if maxTime.Equal(hlc.PhysicalTime) && maxTime.Equal(peerHLC.PhysicalTime) {
        logicalCounter = Helper.Max(hlc.LogicalCounter, peerHLC.LogicalCounter) + 1
    } else if maxTime.Equal(hlc.PhysicalTime) {
        logicalCounter = hlc.LogicalCounter + 1
    } else if maxTime.Equal(peerHLC.PhysicalTime) {
        logicalCounter = peerHLC.LogicalCounter + 1
    } else {
        logicalCounter = Helper.Max(hlc.LogicalCounter, peerHLC.LogicalCounter) + 1
    }

    return HLC{
        PhysicalTime:   maxTime,
        LogicalCounter: logicalCounter,
        Uncertainty:    hlc.Uncertainty,
    }
}

// ConnectPeers establishes peer connections
func (n *Node) ConnectPeers(nodes []*Node) {
    for _, peer := range nodes {
        if peer.ID != n.ID {
            n.Peers = append(n.Peers, peer)
        }
    }
}
