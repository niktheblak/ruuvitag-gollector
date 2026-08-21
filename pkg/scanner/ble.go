package scanner

import (
	"context"
	"encoding/binary"
	"strings"
)

// Advertisement contains the portion of a Bluetooth LE advertisement used by
// the collector. ManufacturerData is keyed by the Bluetooth company identifier.
type Advertisement struct {
	Address          string
	ManufacturerData map[uint16][]byte
}

// RawManufacturerData returns a manufacturer-specific advertising field in
// its on-air form: a little-endian company identifier followed by its payload.
func (a Advertisement) RawManufacturerData(id uint16) []byte {
	payload, ok := a.ManufacturerData[id]
	if !ok {
		return nil
	}
	data := make([]byte, 2+len(payload))
	binary.LittleEndian.PutUint16(data, id)
	copy(data[2:], payload)
	return data
}

type AdvertisementHandler func(Advertisement)
type AdvertisementFilter func(Advertisement) bool

// BLEAdapter owns a Bluetooth adapter and its discovery lifecycle.
type BLEAdapter interface {
	Scan(context.Context, AdvertisementHandler) error
	Close() error
}

type AdapterFactory interface {
	NewAdapter(name string) (BLEAdapter, error)
}

// NormalizeAddress ensures configured and observed addresses are compared
// case-insensitively.
func NormalizeAddress(address string) string {
	return strings.ToLower(address)
}
