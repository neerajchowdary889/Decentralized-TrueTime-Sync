package Events

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
    "github.com/Decentralized-TrueTime-Sync/Node"
	"github.com/Decentralized-TrueTime-Sync/VectorClockFormat"
)

// ProcessEvents handles incoming events
func ProcessEvents(n *Node.Node, wg *sync.WaitGroup, done chan struct{}) {
    defer wg.Done()
    
    for {
        select {
        case event := <-n.EventChannel:
            n.Mutex.Lock()
            // Update local clock based on event timestamp
            syncedHLC := n.HLC.MergeTimestamps(event.Timestamp)
            n.HLC = &syncedHLC
            // Store the event with the updated timestamp
            event.Timestamp = *n.HLC
            n.Events = append(n.Events, event)
            n.Mutex.Unlock()
            
        case <-done:
            return
        }
    }
}

// GenerateLocalEvent creates a new local event
func GenerateLocalEvent(n *Node.Node, wg *sync.WaitGroup, done chan struct{}) {
    defer wg.Done()
    
    for i := 0; i < 5; i++ { // Generate 5 events per node
        select {
        case <-done:
            return
        default:
            n.Mutex.Lock()
            timestamp := n.HLC.GenerateTimestamp()
            eventData := fmt.Sprintf("Event from Node %d (#%d)", n.ID, i)
            event := Node.Event{
                NodeID:    n.ID,
                Timestamp: timestamp,
                Data:      eventData,
            }
            n.Events = append(n.Events, event)
            
            // Broadcast event to peers with artificial network latency
            go func(e Node.Event) {
                for _, peer := range n.Peers {
                    time.Sleep(n.NetworkLatency) // Simulate network latency
                    peer.EventChannel <- e
                }
            }(event)
            
            n.Mutex.Unlock()
            
            // Random delay between events
            time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
        }
    }
}

// GossipSync periodically syncs with random peers
func GossipSync(n *Node.Node, wg *sync.WaitGroup, done chan struct{}) {
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
            
            n.Mutex.Lock()
            randomPeer.Mutex.Lock()
            
            // Merge clocks
            syncedHLC := n.HLC.MergeTimestamps(*randomPeer.HLC)
            n.HLC = &syncedHLC
            
            // Update peer's clock too for bidirectional sync
            peerSyncedHLC := randomPeer.HLC.MergeTimestamps(syncedHLC)
            randomPeer.HLC = &peerSyncedHLC
            
            n.SyncCount++
            randomPeer.SyncCount++

            fmt.Printf("│ %-3d │ %-3d │ %s │ %s │\n", 
                n.ID, 
                randomPeer.ID, 
                fmt.Sprintf("%02d:%02d:%02d.%03d LC:%-3d", 
                    n.HLC.PhysicalTime.Hour(),
                    n.HLC.PhysicalTime.Minute(),
                    n.HLC.PhysicalTime.Second(),
                    n.HLC.PhysicalTime.Nanosecond()/1000000,
                    n.HLC.LogicalCounter),
                VectorClockFormat.VectorClockFormat(*n.HLC, n.ID))
            
            randomPeer.Mutex.Unlock()
            n.Mutex.Unlock()
        }
    }
}


