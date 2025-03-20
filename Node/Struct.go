package Node

import (
	"sync"
	"time"
)

// HLC represents a Hybrid Logical Clock with Decentralized TrueTime (DTT)
type HLC struct {
    PhysicalTime   time.Time
    LogicalCounter uint64
    Uncertainty    time.Duration
    Mutex          sync.Mutex
}

// Event represents a simulated event in the distributed system
type Event struct {
    NodeID    int
    Timestamp HLC
    Data      string
}

// P2P Gossip-Based Synchronization
type Node struct {
    ID            int
    HLC           *HLC
    Peers         []*Node
    Events        []Event
    SyncCount     int
    Mutex         sync.Mutex
    EventChannel  chan Event
    NetworkLatency time.Duration // Simulated network latency
}

// NodeMetrics tracks statistics about node operations
type NodeMetrics struct {
    NodeID       int
    SyncCount    int
    EventCount   int
    InitialClock HLC
    FinalClock   HLC
}

// NewNode creates a new blockchain node with a specific drift
func NewNode(id int, clockDrift time.Duration, latency time.Duration) *Node {
    return &Node{
        ID:            id,
        HLC:           NewHLC(clockDrift),
        Peers:         []*Node{},
        Events:        []Event{},
        SyncCount:     0,
        NetworkLatency: latency,
        EventChannel:  make(chan Event, 100),
    }
}
