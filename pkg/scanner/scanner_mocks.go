package scanner

import (
	"context"
	"encoding/binary"
	"sync"

	"github.com/niktheblak/ruuvitag-common/pkg/sensor"
)

type mockAdapterFactory struct {
	adapter BLEAdapter
}

func (m mockAdapterFactory) NewAdapter(string) (BLEAdapter, error) {
	return m.adapter, nil
}

type mockBLEAdapter struct {
	advertisements []Advertisement
	current        int
	mu             sync.Mutex
}

func NewMockBLEAdapter(advertisements ...mockAdvertisement) *mockBLEAdapter {
	result := &mockBLEAdapter{}
	for _, advertisement := range advertisements {
		result.advertisements = append(result.advertisements, advertisement.advertisement())
	}
	return result
}

func (m *mockBLEAdapter) Scan(ctx context.Context, handler AdvertisementHandler) error {
	m.mu.Lock()
	if m.current < len(m.advertisements) {
		advertisement := m.advertisements[m.current]
		m.current++
		m.mu.Unlock()
		handler(advertisement)
	} else {
		m.mu.Unlock()
	}
	<-ctx.Done()
	return ctx.Err()
}

func (m *mockBLEAdapter) Close() error { return nil }

type mockAdvertisement struct {
	manufacturerData []byte
	addr             string
}

func (m mockAdvertisement) advertisement() Advertisement {
	manufacturerData := make(map[uint16][]byte)
	if len(m.manufacturerData) >= 2 {
		companyID := binary.LittleEndian.Uint16(m.manufacturerData[:2])
		manufacturerData[companyID] = append([]byte(nil), m.manufacturerData[2:]...)
	}
	return Advertisement{
		Address:          m.addr,
		ManufacturerData: manufacturerData,
	}
}

type mockExporter struct {
	events []sensor.Data
}

func (m *mockExporter) Name() string {
	return "Mock"
}

func (m *mockExporter) Export(_ context.Context, data sensor.Data) error {
	m.events = append(m.events, data)
	return nil
}

func (m *mockExporter) Close() error {
	return nil
}
