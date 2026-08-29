package scanner

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/niktheblak/ruuvitag-gollector/pkg/sensor"
)

type fakeBlueZBus struct {
	objects managedObjects
	signals chan *dbus.Signal
	ctx     context.Context
	cancel  context.CancelFunc

	mu          sync.Mutex
	powered     []bool
	filter      map[string]dbus.Variant
	startErr    error
	stopErr     error
	startCalls  int
	stopCalls   int
	startCalled chan struct{}
	startOnce   sync.Once
}

func newFakeBlueZBus(powered bool) *fakeBlueZBus {
	ctx, cancel := context.WithCancel(context.Background())
	path := dbus.ObjectPath("/org/bluez/hci0")
	return &fakeBlueZBus{
		objects: managedObjects{
			path: {
				adapterInterface: {
					"Address": dbus.MakeVariant("00:11:22:33:44:55"),
					"Powered": dbus.MakeVariant(powered),
				},
			},
		},
		signals:     make(chan *dbus.Signal, 16),
		ctx:         ctx,
		cancel:      cancel,
		startCalled: make(chan struct{}),
	}
}

func (b *fakeBlueZBus) ManagedObjects(context.Context) (managedObjects, error) {
	return b.objects, nil
}

func (b *fakeBlueZBus) SetPowered(_ context.Context, _ dbus.ObjectPath, powered bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.powered = append(b.powered, powered)
	return nil
}

func (b *fakeBlueZBus) SetDiscoveryFilter(_ context.Context, _ dbus.ObjectPath, filter map[string]dbus.Variant) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.filter = filter
	return nil
}

func (b *fakeBlueZBus) StartDiscovery(context.Context, dbus.ObjectPath) error {
	b.mu.Lock()
	b.startCalls++
	err := b.startErr
	b.mu.Unlock()
	b.startOnce.Do(func() { close(b.startCalled) })
	return err
}

func (b *fakeBlueZBus) StopDiscovery(context.Context, dbus.ObjectPath) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.stopCalls++
	return b.stopErr
}

func (b *fakeBlueZBus) Signals() <-chan *dbus.Signal { return b.signals }
func (b *fakeBlueZBus) Context() context.Context     { return b.ctx }
func (b *fakeBlueZBus) Close() error {
	b.cancel()
	return nil
}

func TestSelectBlueZAdapter(t *testing.T) {
	t.Parallel()
	bus := newFakeBlueZBus(true)
	secondPath := dbus.ObjectPath("/org/bluez/hci1")
	bus.objects[secondPath] = map[string]map[string]dbus.Variant{
		adapterInterface: {
			"Address": dbus.MakeVariant("AA:BB:CC:DD:EE:FF"),
			"Powered": dbus.MakeVariant(true),
		},
	}

	path, _, err := selectBlueZAdapter(bus.objects, "default")
	require.NoError(t, err)
	assert.Equal(t, dbus.ObjectPath("/org/bluez/hci0"), path)

	path, _, err = selectBlueZAdapter(bus.objects, "hci1")
	require.NoError(t, err)
	assert.Equal(t, secondPath, path)

	path, _, err = selectBlueZAdapter(bus.objects, "aa:bb:cc:dd:ee:ff")
	require.NoError(t, err)
	assert.Equal(t, secondPath, path)
}

func TestNewBlueZAdapterPowersOnAdapter(t *testing.T) {
	t.Parallel()
	bus := newFakeBlueZBus(false)
	adapter, err := newBlueZAdapter(bus, "default")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, adapter.Close()) })
	assert.Equal(t, []bool{true}, bus.powered)
}

func TestBlueZScanForwardsDuplicateAdvertisements(t *testing.T) {
	bus := newFakeBlueZBus(true)
	adapter, err := newBlueZAdapter(bus, "default")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, adapter.Close()) })

	ctx, cancel := context.WithCancel(context.Background())
	results := make(chan Advertisement, 2)
	done := make(chan error, 1)
	go func() {
		done <- adapter.Scan(ctx, func(advertisement Advertisement) {
			results <- advertisement
		})
	}()
	<-bus.startCalled

	payload := make([]byte, 14)
	payload[0] = 5
	manufacturerData := dbus.MakeVariant(map[uint16]dbus.Variant{
		sensor.RuuviManufacturerID: dbus.MakeVariant(payload),
	})
	devicePath := dbus.ObjectPath("/org/bluez/hci0/dev_CC_CA_7E_52_CC_34")
	bus.signals <- &dbus.Signal{
		Path: dbus.ObjectPath("/"),
		Name: interfacesAddedSignal,
		Body: []any{
			devicePath,
			map[string]map[string]dbus.Variant{
				deviceInterface: {
					"Address":          dbus.MakeVariant("CC:CA:7E:52:CC:34"),
					"ManufacturerData": manufacturerData,
				},
			},
		},
	}
	bus.signals <- &dbus.Signal{
		Path: devicePath,
		Name: propertiesChangedSignal,
		Body: []any{
			deviceInterface,
			map[string]dbus.Variant{"ManufacturerData": manufacturerData},
			[]string{},
		},
	}

	first := <-results
	second := <-results
	assert.Equal(t, "cc:ca:7e:52:cc:34", first.Address)
	assert.Equal(t, first, second)
	assert.Equal(t, []byte{0x99, 0x04, 0x05}, first.RawManufacturerData(sensor.RuuviManufacturerID)[:3])

	bus.mu.Lock()
	duplicateData := bus.filter["DuplicateData"]
	transport := bus.filter["Transport"]
	bus.mu.Unlock()
	assert.Equal(t, true, duplicateData.Value())
	assert.Equal(t, "le", transport.Value())

	cancel()
	require.ErrorIs(t, <-done, context.Canceled)
	bus.mu.Lock()
	assert.Equal(t, 1, bus.startCalls)
	assert.Equal(t, 1, bus.stopCalls)
	bus.mu.Unlock()
}

func TestBlueZScanReturnsStartFailureWithoutStopping(t *testing.T) {
	t.Parallel()
	bus := newFakeBlueZBus(true)
	bus.startErr = errors.New("start failed")
	adapter, err := newBlueZAdapter(bus, "default")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, adapter.Close()) })

	err = adapter.Scan(context.Background(), func(Advertisement) {})
	require.ErrorContains(t, err, "start failed")
	bus.mu.Lock()
	assert.Equal(t, 1, bus.startCalls)
	assert.Zero(t, bus.stopCalls)
	bus.mu.Unlock()
}

func TestBlueZScanReturnsErrorWhenBlueZDisappears(t *testing.T) {
	t.Parallel()
	bus := newFakeBlueZBus(true)
	adapter, err := newBlueZAdapter(bus, "default")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, adapter.Close()) })

	done := make(chan error, 1)
	go func() {
		done <- adapter.Scan(context.Background(), func(Advertisement) {})
	}()
	<-bus.startCalled
	bus.signals <- &dbus.Signal{
		Path: dbus.ObjectPath("/org/freedesktop/DBus"),
		Name: nameOwnerChangedSignal,
		Body: []any{blueZService, ":1.10", ""},
	}
	require.ErrorContains(t, <-done, "owner changed")
}
