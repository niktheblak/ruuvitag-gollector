package scanner

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"
	"time"

	"github.com/niktheblak/ruuvitag-gollector/pkg/config"
	"github.com/niktheblak/ruuvitag-gollector/pkg/exporter"
	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
	"github.com/paypal/gatt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDevice struct {
	manufacturerData []byte
	handler          func(p gatt.Peripheral, a *gatt.Advertisement, rssi int)
}

func (m *mockDevice) NewDevice() (gatt.Device, error) {
	return m, nil
}

func (m *mockDevice) Init(stateChanged func(gatt.Device, gatt.State)) error {
	stateChanged(m, gatt.StatePoweredOn)
	return nil
}

func (m *mockDevice) Advertise(a *gatt.AdvPacket) error {
	return nil
}

func (m *mockDevice) AdvertiseNameAndServices(name string, ss []gatt.UUID) error {
	return nil
}

func (m *mockDevice) AdvertiseIBeaconData(b []byte) error {
	return nil
}

func (m *mockDevice) AdvertiseIBeacon(u gatt.UUID, major, minor uint16, pwr int8) error {
	return nil
}

func (m *mockDevice) StopAdvertising() error {
	return nil
}

func (m *mockDevice) RemoveAllServices() error {
	return nil
}

func (m *mockDevice) AddService(s *gatt.Service) error {
	return nil
}

func (m *mockDevice) SetServices(ss []*gatt.Service) error {
	return nil
}

func (m *mockDevice) Scan(ss []gatt.UUID, dup bool) {
	p := &mockPeripheral{
		id:   "03ee30e3-1f90-46d7-b822-5d6570f4c5be",
		name: "RuuviTag",
	}
	a := &gatt.Advertisement{
		ManufacturerData: m.manufacturerData,
	}
	m.handler(p, a, 0)
}

func (m *mockDevice) StopScanning() {}

func (m *mockDevice) Connect(p gatt.Peripheral) {}

func (m *mockDevice) CancelConnection(p gatt.Peripheral) {}

func (m *mockDevice) Handle(h ...gatt.Handler) {}

func (m *mockDevice) Option(o ...gatt.Option) error {
	return nil
}

func (m *mockDevice) HandlePeripheralDiscovered(d gatt.Device, handler func(p gatt.Peripheral, a *gatt.Advertisement, rssi int)) {
	m.handler = handler
}

type mockPeripheral struct {
	id   string
	name string
}

func (m *mockPeripheral) Device() gatt.Device {
	return nil
}

func (m *mockPeripheral) ID() string {
	return m.id
}

func (m *mockPeripheral) Name() string {
	return m.name
}

func (m *mockPeripheral) Services() []*gatt.Service {
	return nil
}

func (m *mockPeripheral) DiscoverServices(s []gatt.UUID) ([]*gatt.Service, error) {
	return nil, nil
}

func (m *mockPeripheral) DiscoverIncludedServices(ss []gatt.UUID, s *gatt.Service) ([]*gatt.Service, error) {
	return nil, nil
}

func (m *mockPeripheral) DiscoverCharacteristics(c []gatt.UUID, s *gatt.Service) ([]*gatt.Characteristic, error) {
	return nil, nil
}

func (m *mockPeripheral) DiscoverDescriptors(d []gatt.UUID, c *gatt.Characteristic) ([]*gatt.Descriptor, error) {
	return nil, nil
}

func (m *mockPeripheral) ReadCharacteristic(c *gatt.Characteristic) ([]byte, error) {
	return nil, nil
}

func (m *mockPeripheral) ReadLongCharacteristic(c *gatt.Characteristic) ([]byte, error) {
	return nil, nil
}

func (m *mockPeripheral) ReadDescriptor(d *gatt.Descriptor) ([]byte, error) {
	return nil, nil
}

func (m *mockPeripheral) WriteCharacteristic(c *gatt.Characteristic, b []byte, noRsp bool) error {
	return nil
}

func (m *mockPeripheral) WriteDescriptor(d *gatt.Descriptor, b []byte) error {
	return nil
}

func (m *mockPeripheral) SetNotifyValue(c *gatt.Characteristic, f func(*gatt.Characteristic, []byte, error)) error {
	return nil
}

func (m *mockPeripheral) SetIndicateValue(c *gatt.Characteristic, f func(*gatt.Characteristic, []byte, error)) error {
	return nil
}

func (m *mockPeripheral) ReadRSSI() int {
	return 0
}

func (m *mockPeripheral) SetMTU(mtu uint16) error {
	return nil
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

var testData = sensor.DataFormat3{
	ManufacturerID:      0x9904,
	DataFormat:          3,
	Humidity:            120,
	Temperature:         55,
	TemperatureFraction: 0,
	Pressure:            1000,
	AccelerationX:       0,
	AccelerationY:       0,
	AccelerationZ:       0,
	BatteryVoltageMv:    500,
}

func TestScanner(t *testing.T) {
	cfg := config.Config{}
	scn, err := New(cfg)
	require.NoError(t, err)
	exp := new(mockExporter)
	scn.Exporters = []exporter.Exporter{exp}
	device := new(mockDevice)
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, testData)
	require.NoError(t, err)
	device.manufacturerData = buf.Bytes()
	scn.deviceCreator = device
	scn.peripheralDiscoverer = device
	ctx := context.Background()
	err = scn.Start(ctx)
	require.NoError(t, err)
	// Wait a bit for messages to appear in the measurements channel
	time.Sleep(100 * time.Millisecond)
	scn.Stop()
	assert.NotEmpty(t, exp.events)
	e := exp.events[0]
	assert.Equal(t, "", e.Name)
	assert.Equal(t, "03ee30e3-1f90-46d7-b822-5d6570f4c5be", e.DeviceID)
	assert.Equal(t, 55.0, e.Temperature)
	assert.Equal(t, 60.0, e.Humidity)
	assert.Equal(t, 51000, e.Pressure)
	assert.Equal(t, 500, e.Battery)
}
