package Helper

import "time"

// Helper function to calculate absolute duration
func AbsDuration(d time.Duration) time.Duration {
    if d < 0 {
        return -d
    }
    return d
}

// Helper function to calculate absolute difference between uint64 values
func AbsUint64(a, b uint64) uint64 {
    if a > b {
        return a - b
    }
    return b - a
}

// Helper function to find max time
func MaxTime(t1, t2, t3 time.Time) time.Time {
    if t1.After(t2) && t1.After(t3) {
        return t1
    } else if t2.After(t3) {
        return t2
    }
    return t3
}

// Helper function to find max of two numbers
func Max(a, b uint64) uint64 {
    if a > b {
        return a
    }
    return b
}

