package scanner

import (
	"context"

	"github.com/go-ble/ble"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

type mockDevice struct {
}

func (m mockDevice) AddService(svc *ble.Service) error {
	return nil
}

func (m mockDevice) RemoveAllServices() error {
	return nil
}

func (m mockDevice) SetServices(svcs []*ble.Service) error {
	return nil
}

func (m mockDevice) Stop() error {
	return nil
}

func (m mockDevice) Advertise(ctx context.Context, adv ble.Advertisement) error {
	return nil
}

func (m mockDevice) AdvertiseNameAndServices(ctx context.Context, name string, uuids ...ble.UUID) error {
	return nil
}

func (m mockDevice) AdvertiseMfgData(ctx context.Context, id uint16, b []byte) error {
	return nil
}

func (m mockDevice) AdvertiseServiceData16(ctx context.Context, id uint16, b []byte) error {
	return nil
}

func (m mockDevice) AdvertiseIBeaconData(ctx context.Context, b []byte) error {
	return nil
}

func (m mockDevice) AdvertiseIBeacon(ctx context.Context, u ble.UUID, major, minor uint16, pwr int8) error {
	return nil
}

func (m mockDevice) Scan(ctx context.Context, allowDup bool, h ble.AdvHandler) error {
	return nil
}

func (m mockDevice) Dial(ctx context.Context, a ble.Addr) (ble.Client, error) {
	return nil, nil
}

type mockDeviceCreator struct {
	device ble.Device
}

func (m mockDeviceCreator) NewDevice(impl string) (ble.Device, error) {
	return m.device, nil
}

type mockBLEScanner struct {
	manufacturerData []byte
}

func (m mockBLEScanner) Scan(ctx context.Context, allowDup bool, h ble.AdvHandler, f ble.AdvFilter) error {
	h(mockAdvertisement{
		localName:        "RuuviTag",
		addr:             "CC:CA:7E:52:CC:34",
		manufacturerData: m.manufacturerData,
	})
	return nil
}

type mockAdvertisement struct {
	localName        string
	manufacturerData []byte
	addr             string
}

func (m mockAdvertisement) Addr() ble.Addr {
	return ble.NewAddr(m.addr)
}

func (m mockAdvertisement) LocalName() string {
	return m.localName
}

func (m mockAdvertisement) ManufacturerData() []byte {
	return m.manufacturerData
}

func (m mockAdvertisement) ServiceData() []ble.ServiceData {
	return nil
}

func (m mockAdvertisement) Services() []ble.UUID {
	return nil
}

func (m mockAdvertisement) OverflowService() []ble.UUID {
	return nil
}

func (m mockAdvertisement) TxPowerLevel() int {
	return 1
}

func (m mockAdvertisement) Connectable() bool {
	return false
}

func (m mockAdvertisement) SolicitedService() []ble.UUID {
	return nil
}

func (m mockAdvertisement) RSSI() int {
	return 0
}

func (m mockAdvertisement) Address() ble.Addr {
	return ble.NewAddr(m.addr)
}

type mockExporter struct {
	events []sensor.Data
}

func (m *mockExporter) Name() string {
	return "Mock"
}

func (m *mockExporter) Export(ctx context.Context, data sensor.Data) error {
	m.events = append(m.events, data)
	return nil
}

func (m *mockExporter) Close() error {
	return nil
}
