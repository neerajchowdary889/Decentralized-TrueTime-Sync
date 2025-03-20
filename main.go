package main

import (
    "fmt"
    "math/rand"
    "sync"
    "time"
)

// HLC represents a Hybrid Logical Clock with Decentralized TrueTime (DTT)
type HLC struct {
    PhysicalTime   time.Time
    LogicalCounter uint64
    Uncertainty    time.Duration
    mutex          sync.Mutex
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
    mutex         sync.Mutex
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

// String method for better HLC output
func (hlc HLC) String() string {
    return fmt.Sprintf("Time: %s, LC: %d, Uncertainty: %v", 
        hlc.PhysicalTime.Format("15:04:05.000"), 
        hlc.LogicalCounter, 
        hlc.Uncertainty)
}

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
    hlc.mutex.Lock()
    defer hlc.mutex.Unlock()

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
    hlc.mutex.Lock()
    defer hlc.mutex.Unlock()

    now := time.Now()
    maxTime := maxTime(now, peerHLC.PhysicalTime, hlc.PhysicalTime)
    logicalCounter := max(hlc.LogicalCounter, peerHLC.LogicalCounter)

    if maxTime.Equal(hlc.PhysicalTime) && maxTime.Equal(peerHLC.PhysicalTime) {
        logicalCounter = max(hlc.LogicalCounter, peerHLC.LogicalCounter) + 1
    } else if maxTime.Equal(hlc.PhysicalTime) {
        logicalCounter = hlc.LogicalCounter + 1
    } else if maxTime.Equal(peerHLC.PhysicalTime) {
        logicalCounter = peerHLC.LogicalCounter + 1
    } else {
        logicalCounter = max(hlc.LogicalCounter, peerHLC.LogicalCounter) + 1
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

// ProcessEvents handles incoming events
func (n *Node) ProcessEvents(wg *sync.WaitGroup, done chan struct{}) {
    defer wg.Done()
    
    for {
        select {
        case event := <-n.EventChannel:
            n.mutex.Lock()
            // Update local clock based on event timestamp
            syncedHLC := n.HLC.MergeTimestamps(event.Timestamp)
            n.HLC = &syncedHLC
            // Store the event with the updated timestamp
            event.Timestamp = *n.HLC
            n.Events = append(n.Events, event)
            n.mutex.Unlock()
            
        case <-done:
            return
        }
    }
}

// GenerateLocalEvent creates a new local event
func (n *Node) GenerateLocalEvent(wg *sync.WaitGroup, done chan struct{}) {
    defer wg.Done()
    
    for i := 0; i < 5; i++ { // Generate 5 events per node
        select {
        case <-done:
            return
        default:
            n.mutex.Lock()
            timestamp := n.HLC.GenerateTimestamp()
            eventData := fmt.Sprintf("Event from Node %d (#%d)", n.ID, i)
            event := Event{
                NodeID:    n.ID,
                Timestamp: timestamp,
                Data:      eventData,
            }
            n.Events = append(n.Events, event)
            
            // Broadcast event to peers with artificial network latency
            go func(e Event) {
                for _, peer := range n.Peers {
                    time.Sleep(n.NetworkLatency) // Simulate network latency
                    peer.EventChannel <- e
                }
            }(event)
            
            n.mutex.Unlock()
            
            // Random delay between events
            time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
        }
    }
}

// GossipSync periodically syncs with random peers
func (n *Node) GossipSync(wg *sync.WaitGroup, done chan struct{}) {
    defer wg.Done()
    
    for {
        select {
        case <-done:
            return
        default:
            time.Sleep(time.Duration(rand.Intn(300)+200) * time.Millisecond)
            
            if len(n.Peers) == 0 {
                continue
            }
            
            // Choose a random peer to sync with
            randomPeer := n.Peers[rand.Intn(len(n.Peers))]
            
            // Simulate network latency
            time.Sleep(n.NetworkLatency)
            
            n.mutex.Lock()
            randomPeer.mutex.Lock()
            
            // Merge clocks
            syncedHLC := n.HLC.MergeTimestamps(*randomPeer.HLC)
            n.HLC = &syncedHLC
            
            // Update peer's clock too for bidirectional sync
            peerSyncedHLC := randomPeer.HLC.MergeTimestamps(syncedHLC)
            randomPeer.HLC = &peerSyncedHLC
            
            n.SyncCount++
            randomPeer.SyncCount++
            
            // Print both standard format and Vector Clock format
            fmt.Printf("Node %d synced with Node %d: %v | Vector: %s\n", 
                n.ID, 
                randomPeer.ID, 
                n.HLC, 
                n.HLC.VectorClockFormat(n.ID))
            
            randomPeer.mutex.Unlock()
            n.mutex.Unlock()
        }
    }
}

// Helper function to find max time
func maxTime(t1, t2, t3 time.Time) time.Time {
    if t1.After(t2) && t1.After(t3) {
        return t1
    } else if t2.After(t3) {
        return t2
    }
    return t3
}

// Helper function to find max of two numbers
func max(a, b uint64) uint64 {
    if a > b {
        return a
    }
    return b
}

// CalculateMetrics collects metrics from all nodes
func calculateMetrics(nodes []*Node, initialClocks []HLC) []NodeMetrics {
    metrics := make([]NodeMetrics, len(nodes))
    
    for i, node := range nodes {
        node.mutex.Lock()
        metrics[i] = NodeMetrics{
            NodeID:       node.ID,
            SyncCount:    node.SyncCount,
            EventCount:   len(node.Events),
            InitialClock: initialClocks[i],
            FinalClock:   *node.HLC,
        }
        node.mutex.Unlock()
    }
    
    return metrics
}

// VectorClockFormat returns the timestamp in Vector Clock style format
func (hlc HLC) VectorClockFormat(nodeID int) string {
    return fmt.Sprintf("N%d@[%s:%d:%dms]",
        nodeID,
        hlc.PhysicalTime.Format(time.RFC3339Nano),
        hlc.LogicalCounter,
        hlc.Uncertainty/time.Millisecond)
}

// Simulate multiple blockchain nodes with P2P time synchronization
func main() {
    // Set random seed
    rand.Seed(time.Now().UnixNano())
    
    // Number of nodes
    numNodes := 100
    simulationTime := 10 * time.Second // Run for 10 seconds
    
    // Create nodes with different clock drifts and network latencies
    nodes := make([]*Node, numNodes)
    initialClocks := make([]HLC, numNodes)
    
    for i := 0; i < numNodes; i++ {
        // Random clock drift between -50ms and +50ms
        drift := time.Duration(rand.Intn(100) - 50) * time.Millisecond
        // Random network latency between 10ms and 100ms
        latency := time.Duration(rand.Intn(90) + 10) * time.Millisecond
        
        nodes[i] = NewNode(i, drift, latency)
        initialClocks[i] = *nodes[i].HLC
        fmt.Printf("Node %d created with drift %v and latency %v\n", i, drift, latency)
    }
    
    // Establish peer connections (fully connected topology)
    for _, node := range nodes {
        node.ConnectPeers(nodes)
    }
    
    // Channel to signal done
    done := make(chan struct{})
    var wg sync.WaitGroup
    
    // Start event processing for each node
    for _, node := range nodes {
        wg.Add(1)
        go node.ProcessEvents(&wg, done)
    }
    
    // Start gossip-based synchronization
    for _, node := range nodes {
        wg.Add(1)
        go node.GossipSync(&wg, done)
    }
    
    // Generate events on each node
    for _, node := range nodes {
        wg.Add(1)
        go node.GenerateLocalEvent(&wg, done)
    }
    
    // Run simulation for the specified duration
    fmt.Printf("Simulation started, running for %v...\n", simulationTime)
    time.Sleep(simulationTime)
    
    // Signal completion and wait for goroutines to finish
    close(done)
    wg.Wait()
    
    // Calculate and display metrics
    metrics := calculateMetrics(nodes, initialClocks)
    
    fmt.Println("\n==== Simulation Results ====")
    for _, m := range metrics {
        fmt.Printf("\nNode %d:\n", m.NodeID)
        fmt.Printf("  Sync Count: %d\n", m.SyncCount)
        fmt.Printf("  Event Count: %d\n", m.EventCount)
        fmt.Printf("  Initial Clock: %v\n", m.InitialClock)
        fmt.Printf("  Final Clock: %v\n", m.FinalClock)
    }
    
    // Check clock convergence
    fmt.Println("\n==== Clock Convergence Analysis ====")
    maxTimeDiff := time.Duration(0)
    maxLogicalDiff := uint64(0)
    
    for i := 0; i < len(nodes); i++ {
        for j := i + 1; j < len(nodes); j++ {
            timeDiff := absDuration(nodes[i].HLC.PhysicalTime.Sub(nodes[j].HLC.PhysicalTime))
            logicalDiff := absUint64(nodes[i].HLC.LogicalCounter, nodes[j].HLC.LogicalCounter)
            
            if timeDiff > maxTimeDiff {
                maxTimeDiff = timeDiff
            }
            if logicalDiff > maxLogicalDiff {
                maxLogicalDiff = logicalDiff
            }
        }
    }
    
    fmt.Printf("Maximum Physical Time Difference: %v\n", maxTimeDiff)
    fmt.Printf("Maximum Logical Counter Difference: %d\n", maxLogicalDiff)
    
    if maxTimeDiff <= 10*time.Millisecond && maxLogicalDiff <= 2 {
        fmt.Println("✅ Clocks have successfully converged!")
    } else {
        fmt.Println("⚠️ Clocks have not fully converged.")
    }
}

// Helper function to calculate absolute duration
func absDuration(d time.Duration) time.Duration {
    if d < 0 {
        return -d
    }
    return d
}

// Helper function to calculate absolute difference between uint64 values
func absUint64(a, b uint64) uint64 {
    if a > b {
        return a - b
    }
    return b - a
}