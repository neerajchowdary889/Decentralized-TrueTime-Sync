package Metrics

import (
	"github.com/Decentralized-TrueTime-Sync/Node"
)

// CalculateMetrics collects metrics from all nodes
func CalculateMetrics(nodes []*Node.Node, initialClocks []Node.HLC) []Node.NodeMetrics {
    metrics := make([]Node.NodeMetrics, len(nodes))
    
    for i, node := range nodes {
        node.Mutex.Lock()
        metrics[i] = Node.NodeMetrics{
            NodeID:       node.ID,
            SyncCount:    node.SyncCount,
            EventCount:   len(node.Events),
            InitialClock: initialClocks[i],
            FinalClock:   *node.HLC,
        }
        node.Mutex.Unlock()
    }
    
    return metrics
}