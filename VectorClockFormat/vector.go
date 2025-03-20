package VectorClockFormat

import (
	"fmt"
	"time"
	"github.com/Decentralized-TrueTime-Sync/Node"
)


// VectorClockFormat returns the timestamp in Vector Clock style format
func VectorClockFormat(hlc Node.HLC, nodeID int) string {
    return fmt.Sprintf("N%d@[%s:%d:%dms]",
        nodeID,
        hlc.PhysicalTime.Format(time.RFC3339Nano),
        hlc.LogicalCounter,
        hlc.Uncertainty/time.Millisecond)
}