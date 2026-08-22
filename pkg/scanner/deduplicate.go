package scanner

import "sync"

// measurementDeduplicator tracks the latest data format 5 measurement number
// received from each peripheral. Measurement numbers are compared for equality
// rather than ordering so uint16 rollover is treated as a new measurement.
type measurementDeduplicator struct {
	mu     sync.Mutex
	latest map[string]uint16
}

func (d *measurementDeduplicator) IsDuplicate(address string, measurementNumber uint16) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.latest == nil {
		d.latest = make(map[string]uint16)
	}
	address = NormalizeAddress(address)
	previous, seen := d.latest[address]
	if seen && previous == measurementNumber {
		return true
	}
	d.latest[address] = measurementNumber
	return false
}
