package scanner

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	blueZService           = "org.bluez"
	adapterInterface       = "org.bluez.Adapter1"
	deviceInterface        = "org.bluez.Device1"
	objectManagerInterface = "org.freedesktop.DBus.ObjectManager"
	propertiesInterface    = "org.freedesktop.DBus.Properties"
	dbusInterface          = "org.freedesktop.DBus"

	interfacesAddedSignal   = objectManagerInterface + ".InterfacesAdded"
	interfacesRemovedSignal = objectManagerInterface + ".InterfacesRemoved"
	propertiesChangedSignal = propertiesInterface + ".PropertiesChanged"
	nameOwnerChangedSignal  = dbusInterface + ".NameOwnerChanged"
)

type managedObjects map[dbus.ObjectPath]map[string]map[string]dbus.Variant

// blueZBus keeps D-Bus mechanics behind a small interface so the adapter state
// machine and signal decoder can be tested without BlueZ or Bluetooth hardware.
type blueZBus interface {
	ManagedObjects(context.Context) (managedObjects, error)
	SetPowered(context.Context, dbus.ObjectPath, bool) error
	SetDiscoveryFilter(context.Context, dbus.ObjectPath, map[string]dbus.Variant) error
	StartDiscovery(context.Context, dbus.ObjectPath) error
	StopDiscovery(context.Context, dbus.ObjectPath) error
	Signals() <-chan *dbus.Signal
	Context() context.Context
	Close() error
}

type systemBlueZBus struct {
	conn    *dbus.Conn
	signals chan *dbus.Signal
}

func connectSystemBlueZBus() (_ blueZBus, err error) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return nil, fmt.Errorf("connect to system D-Bus: %w", err)
	}
	bus := &systemBlueZBus{
		conn:    conn,
		signals: make(chan *dbus.Signal, 1024),
	}
	conn.Signal(bus.signals)
	defer func() {
		if err != nil {
			conn.RemoveSignal(bus.signals)
			_ = conn.Close()
		}
	}()

	matches := [][]dbus.MatchOption{
		{
			dbus.WithMatchSender(blueZService),
			dbus.WithMatchInterface(objectManagerInterface),
			dbus.WithMatchMember("InterfacesAdded"),
			dbus.WithMatchObjectPath(dbus.ObjectPath("/")),
		},
		{
			dbus.WithMatchSender(blueZService),
			dbus.WithMatchInterface(objectManagerInterface),
			dbus.WithMatchMember("InterfacesRemoved"),
			dbus.WithMatchObjectPath(dbus.ObjectPath("/")),
		},
		{
			dbus.WithMatchSender(blueZService),
			dbus.WithMatchInterface(propertiesInterface),
			dbus.WithMatchMember("PropertiesChanged"),
			dbus.WithMatchPathNamespace(dbus.ObjectPath("/org/bluez")),
		},
		{
			dbus.WithMatchSender(dbusInterface),
			dbus.WithMatchInterface(dbusInterface),
			dbus.WithMatchMember("NameOwnerChanged"),
			dbus.WithMatchObjectPath(dbus.ObjectPath("/org/freedesktop/DBus")),
			dbus.WithMatchArg(0, blueZService),
		},
	}
	matchCtx, matchCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer matchCancel()
	for _, match := range matches {
		if err = conn.AddMatchSignalContext(matchCtx, match...); err != nil {
			return nil, fmt.Errorf("subscribe to BlueZ signals: %w", err)
		}
	}
	return bus, nil
}

func (b *systemBlueZBus) ManagedObjects(ctx context.Context) (objects managedObjects, err error) {
	err = b.conn.Object(blueZService, dbus.ObjectPath("/")).
		CallWithContext(ctx, objectManagerInterface+".GetManagedObjects", 0).
		Store(&objects)
	return
}

func (b *systemBlueZBus) SetPowered(ctx context.Context, path dbus.ObjectPath, powered bool) error {
	return b.conn.Object(blueZService, path).
		CallWithContext(ctx, propertiesInterface+".Set", 0, adapterInterface, "Powered", dbus.MakeVariant(powered)).Err
}

func (b *systemBlueZBus) SetDiscoveryFilter(ctx context.Context, path dbus.ObjectPath, filter map[string]dbus.Variant) error {
	return b.conn.Object(blueZService, path).
		CallWithContext(ctx, adapterInterface+".SetDiscoveryFilter", 0, filter).Err
}

func (b *systemBlueZBus) StartDiscovery(ctx context.Context, path dbus.ObjectPath) error {
	return b.conn.Object(blueZService, path).
		CallWithContext(ctx, adapterInterface+".StartDiscovery", 0).Err
}

func (b *systemBlueZBus) StopDiscovery(ctx context.Context, path dbus.ObjectPath) error {
	return b.conn.Object(blueZService, path).
		CallWithContext(ctx, adapterInterface+".StopDiscovery", 0).Err
}

func (b *systemBlueZBus) Signals() <-chan *dbus.Signal { return b.signals }
func (b *systemBlueZBus) Context() context.Context     { return b.conn.Context() }

func (b *systemBlueZBus) Close() error {
	b.conn.RemoveSignal(b.signals)
	return b.conn.Close()
}

type BlueZAdapterFactory struct {
	connect func() (blueZBus, error)
}

func (f *BlueZAdapterFactory) NewAdapter(name string) (BLEAdapter, error) {
	connect := f.connect
	if connect == nil {
		connect = connectSystemBlueZBus
	}
	bus, err := connect()
	if err != nil {
		return nil, err
	}
	adapter, err := newBlueZAdapter(bus, name)
	if err != nil {
		_ = bus.Close()
		return nil, err
	}
	return adapter, nil
}

type blueZDevice struct {
	address string
}

type blueZAdapter struct {
	bus     blueZBus
	path    dbus.ObjectPath
	devices map[dbus.ObjectPath]blueZDevice

	mu         sync.Mutex
	closed     bool
	scanning   bool
	scanCancel context.CancelFunc
	scanDone   chan struct{}
}

func newBlueZAdapter(bus blueZBus, name string) (*blueZAdapter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	objects, err := bus.ManagedObjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("enumerate BlueZ objects: %w", err)
	}
	path, properties, err := selectBlueZAdapter(objects, name)
	if err != nil {
		return nil, err
	}
	if powered, ok, err := boolProperty(properties, "Powered"); err != nil {
		return nil, err
	} else if !ok || !powered {
		if err := bus.SetPowered(ctx, path, true); err != nil {
			return nil, fmt.Errorf("power on Bluetooth adapter %s: %w", path, err)
		}
	}

	adapter := &blueZAdapter{
		bus:     bus,
		path:    path,
		devices: make(map[dbus.ObjectPath]blueZDevice),
	}
	for objectPath, interfaces := range objects {
		if !adapter.owns(objectPath) {
			continue
		}
		if properties, ok := interfaces[deviceInterface]; ok {
			adapter.updateDevice(objectPath, properties)
		}
	}
	return adapter, nil
}

func selectBlueZAdapter(objects managedObjects, name string) (dbus.ObjectPath, map[string]dbus.Variant, error) {
	adapters := make([]dbus.ObjectPath, 0)
	for path, interfaces := range objects {
		if _, ok := interfaces[adapterInterface]; ok {
			adapters = append(adapters, path)
		}
	}
	slices.Sort(adapters)
	if len(adapters) == 0 {
		return "", nil, fmt.Errorf("no BlueZ Bluetooth adapters found")
	}

	selector := strings.TrimSpace(name)
	if selector == "" || selector == "default" {
		path := adapters[0]
		return path, objects[path][adapterInterface], nil
	}
	if !strings.HasPrefix(selector, "/") && strings.HasPrefix(selector, "hci") {
		selector = "/org/bluez/" + selector
	}
	for _, path := range adapters {
		properties := objects[path][adapterInterface]
		if string(path) == selector || strings.EqualFold(stringProperty(properties, "Address"), selector) {
			return path, properties, nil
		}
	}
	return "", nil, fmt.Errorf("BlueZ Bluetooth adapter %q not found", name)
}

func boolProperty(properties map[string]dbus.Variant, name string) (bool, bool, error) {
	value, ok := properties[name]
	if !ok {
		return false, false, nil
	}
	result, ok := value.Value().(bool)
	if !ok {
		return false, true, fmt.Errorf("BlueZ property %s has type %T, want bool", name, value.Value())
	}
	return result, true, nil
}

func stringProperty(properties map[string]dbus.Variant, name string) string {
	value, ok := properties[name]
	if !ok {
		return ""
	}
	result, _ := value.Value().(string)
	return result
}

//nolint:gocognit // Unfortunately BlueZ handler functions tend to be long and complex
func (a *blueZAdapter) Scan(ctx context.Context, handler AdvertisementHandler) (err error) {
	if handler == nil {
		return fmt.Errorf("Bluetooth advertisement handler must be specified")
	}
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return fmt.Errorf("Bluetooth adapter is closed")
	}
	if a.scanning {
		a.mu.Unlock()
		return fmt.Errorf("Bluetooth discovery is already active")
	}
	scanCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	a.scanning = true
	a.scanCancel = cancel
	a.scanDone = done
	a.mu.Unlock()

	started := false
	defer func() {
		if started {
			stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
			stopErr := a.bus.StopDiscovery(stopCtx, a.path)
			stopCancel()
			if stopErr != nil {
				stopErr = fmt.Errorf("stop Bluetooth discovery: %w", stopErr)
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					err = stopErr
				} else {
					err = errors.Join(err, stopErr)
				}
			}
		}
		cancel()
		a.mu.Lock()
		a.scanning = false
		a.scanCancel = nil
		a.scanDone = nil
		close(done)
		a.mu.Unlock()
	}()

	a.drainSignals()
	discoveryFilter := map[string]dbus.Variant{
		"Transport":     dbus.MakeVariant("le"),
		"DuplicateData": dbus.MakeVariant(true),
	}
	if err := a.bus.SetDiscoveryFilter(scanCtx, a.path, discoveryFilter); err != nil {
		return fmt.Errorf("set Bluetooth discovery filter: %w", err)
	}
	if err := a.bus.StartDiscovery(scanCtx, a.path); err != nil {
		return fmt.Errorf("start Bluetooth discovery: %w", err)
	}
	started = true

	for {
		select {
		case <-scanCtx.Done():
			return scanCtx.Err()
		case <-a.bus.Context().Done():
			if cause := context.Cause(a.bus.Context()); cause != nil {
				return fmt.Errorf("system D-Bus connection closed: %w", cause)
			}
			return fmt.Errorf("system D-Bus connection closed")
		case signal, ok := <-a.bus.Signals():
			if !ok {
				return fmt.Errorf("BlueZ signal stream closed")
			}
			advertisement, signalErr := a.handleSignal(signal)
			if signalErr != nil {
				return signalErr
			}
			if advertisement != nil {
				handler(*advertisement)
			}
		}
	}
}

func (a *blueZAdapter) drainSignals() {
	for {
		select {
		case _, ok := <-a.bus.Signals():
			if !ok {
				return
			}
		default:
			return
		}
	}
}

func (a *blueZAdapter) handleSignal(signal *dbus.Signal) (*Advertisement, error) {
	if signal == nil {
		return nil, nil
	}
	switch signal.Name {
	case nameOwnerChangedSignal:
		return nil, a.handleNameOwnerChanged(signal)
	case interfacesRemovedSignal:
		return nil, a.handleInterfacesRemoved(signal)
	case interfacesAddedSignal:
		return a.handleInterfacesAdded(signal)
	case propertiesChangedSignal:
		return a.handlePropertiesChanged(signal)
	}
	return nil, nil
}

func (a *blueZAdapter) handleNameOwnerChanged(signal *dbus.Signal) error {
	if len(signal.Body) != 3 {
		return fmt.Errorf("invalid NameOwnerChanged signal body")
	}
	name, nameOK := signal.Body[0].(string)
	oldOwner, oldOK := signal.Body[1].(string)
	newOwner, newOK := signal.Body[2].(string)
	if !nameOK || !oldOK || !newOK {
		return fmt.Errorf("invalid NameOwnerChanged signal types")
	}
	if name == blueZService && oldOwner != newOwner {
		return fmt.Errorf("BlueZ D-Bus service owner changed")
	}
	return nil
}

func (a *blueZAdapter) handleInterfacesRemoved(signal *dbus.Signal) error {
	if len(signal.Body) != 2 {
		return fmt.Errorf("invalid InterfacesRemoved signal body")
	}
	path, pathOK := signal.Body[0].(dbus.ObjectPath)
	interfaces, interfacesOK := signal.Body[1].([]string)
	if !pathOK || !interfacesOK {
		return fmt.Errorf("invalid InterfacesRemoved signal types")
	}
	if path == a.path && slices.Contains(interfaces, adapterInterface) {
		return fmt.Errorf("Bluetooth adapter %s was removed", a.path)
	}
	if slices.Contains(interfaces, deviceInterface) {
		delete(a.devices, path)
	}
	return nil
}

func (a *blueZAdapter) handleInterfacesAdded(signal *dbus.Signal) (*Advertisement, error) {
	if len(signal.Body) != 2 {
		return nil, fmt.Errorf("invalid InterfacesAdded signal body")
	}
	path, pathOK := signal.Body[0].(dbus.ObjectPath)
	interfaces, interfacesOK := signal.Body[1].(map[string]map[string]dbus.Variant)
	if !pathOK || !interfacesOK {
		return nil, fmt.Errorf("invalid InterfacesAdded signal types")
	}
	if !a.owns(path) {
		return nil, nil
	}
	properties, ok := interfaces[deviceInterface]
	if !ok {
		return nil, nil
	}
	a.updateDevice(path, properties)
	return a.advertisement(path, properties)
}

func (a *blueZAdapter) handlePropertiesChanged(signal *dbus.Signal) (*Advertisement, error) {
	if len(signal.Body) != 3 {
		return nil, fmt.Errorf("invalid PropertiesChanged signal body")
	}
	iface, ifaceOK := signal.Body[0].(string)
	changed, changedOK := signal.Body[1].(map[string]dbus.Variant)
	if !ifaceOK || !changedOK {
		return nil, fmt.Errorf("invalid PropertiesChanged signal types")
	}
	if signal.Path == a.path && iface == adapterInterface {
		return nil, a.handleAdapterPropertiesChanged(changed)
	}
	if iface != deviceInterface || !a.owns(signal.Path) {
		return nil, nil
	}
	a.updateDevice(signal.Path, changed)
	return a.advertisement(signal.Path, changed)
}

func (a *blueZAdapter) handleAdapterPropertiesChanged(changed map[string]dbus.Variant) error {
	powered, present, err := boolProperty(changed, "Powered")
	if err != nil {
		return err
	}
	if present && !powered {
		return fmt.Errorf("Bluetooth adapter %s was powered off", a.path)
	}

	discovering, present, err := boolProperty(changed, "Discovering")
	if err != nil {
		return err
	}
	if present && !discovering {
		return fmt.Errorf("Bluetooth adapter %s stopped discovering", a.path)
	}
	return nil
}

func (a *blueZAdapter) owns(path dbus.ObjectPath) bool {
	return strings.HasPrefix(string(path), string(a.path)+"/")
}

func (a *blueZAdapter) updateDevice(path dbus.ObjectPath, properties map[string]dbus.Variant) {
	device := a.devices[path]
	if address := stringProperty(properties, "Address"); address != "" {
		device.address = NormalizeAddress(address)
	}
	a.devices[path] = device
}

func (a *blueZAdapter) advertisement(path dbus.ObjectPath, properties map[string]dbus.Variant) (*Advertisement, error) {
	value, ok := properties["ManufacturerData"]
	if !ok {
		return nil, nil
	}
	manufacturerData, err := decodeManufacturerData(value)
	if err != nil {
		return nil, fmt.Errorf("decode manufacturer data for %s: %w", path, err)
	}
	device := a.devices[path]
	if device.address == "" {
		return nil, fmt.Errorf("BlueZ device %s has manufacturer data but no address", path)
	}
	return &Advertisement{
		Address:          device.address,
		ManufacturerData: manufacturerData,
	}, nil
}

func decodeManufacturerData(value dbus.Variant) (map[uint16][]byte, error) {
	entries, ok := value.Value().(map[uint16]dbus.Variant)
	if !ok {
		return nil, fmt.Errorf("property has type %T, want map[uint16]dbus.Variant", value.Value())
	}
	result := make(map[uint16][]byte, len(entries))
	for companyID, entry := range entries {
		payload, ok := entry.Value().([]byte)
		if !ok {
			return nil, fmt.Errorf("company %04x data has type %T, want []byte", companyID, entry.Value())
		}
		result[companyID] = slices.Clone(payload)
	}
	return result, nil
}

func (a *blueZAdapter) Close() error {
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return nil
	}
	a.closed = true
	cancel := a.scanCancel
	done := a.scanDone
	if cancel != nil {
		cancel()
	}
	a.mu.Unlock()
	if done != nil {
		<-done
	}
	return a.bus.Close()
}
