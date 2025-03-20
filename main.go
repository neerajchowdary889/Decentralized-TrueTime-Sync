package main

import (
    "fmt"
    "math/rand"
    "sync"
    "time"
	"github.com/Decentralized-TrueTime-Sync/Node"
	"github.com/Decentralized-TrueTime-Sync/Events"
	"github.com/Decentralized-TrueTime-Sync/Metrics"
	"github.com/Decentralized-TrueTime-Sync/Helper"
)

// Simulate multiple blockchain nodes with P2P time synchronization
func main() {
    // Set random seed
    rand.Seed(time.Now().UnixNano())
    
    // Number of nodes
    numNodes := 100
    simulationTime := 10 * time.Second // Run for 10 seconds
    
    // Create nodes with different clock drifts and network latencies
    nodes := make([]*Node.Node, numNodes)
    initialClocks := make([]Node.HLC, numNodes)
    
    for i := 0; i < numNodes; i++ {
        // Random clock drift between -50ms and +50ms
        drift := time.Duration(rand.Intn(100) - 50) * time.Millisecond
        // Random network latency between 10ms and 100ms
        latency := time.Duration(rand.Intn(90) + 10) * time.Millisecond
        
        nodes[i] = Node.NewNode(i, drift, latency)
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
        go Events.ProcessEvents(node, &wg, done)
    }
    
    // Start gossip-based synchronization
    for _, node := range nodes {
        wg.Add(1)
        go Events.GossipSync(node, &wg, done)
    }
    
    // Generate events on each node
    for _, node := range nodes {
        wg.Add(1)
        go Events.GenerateLocalEvent(node, &wg, done)
    }
    
    // Run simulation for the specified duration
	fmt.Printf("Simulation started, running for %v...\n", simulationTime)
	fmt.Println("┌─────┬─────┬──────────────────────┬─────────────────────────────────────────────────┐")
	fmt.Println("│ SRC │ DST │ TIMESTAMP            │ VECTOR CLOCK                                    │")
	fmt.Println("├─────┼─────┼──────────────────────┼─────────────────────────────────────────────────")
    time.Sleep(simulationTime)    
    // Signal completion and wait for goroutines to finish
    close(done)
    wg.Wait()
	fmt.Println("└─────┴─────┴──────────────────────┴─────────────────────────────────────────────────┘")
    
    // Calculate and display metrics
    metrics := Metrics.CalculateMetrics(nodes, initialClocks)
    
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
            timeDiff := Helper.AbsDuration(nodes[i].HLC.PhysicalTime.Sub(nodes[j].HLC.PhysicalTime))
            logicalDiff := Helper.AbsUint64(nodes[i].HLC.LogicalCounter, nodes[j].HLC.LogicalCounter)
            
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