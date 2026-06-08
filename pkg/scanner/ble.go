package scanner

import (
	"context"

	"tinygo.org/x/bluetooth"
)

type Adapter interface {
	Enable() error
	Scan(func(*bluetooth.Adapter, bluetooth.ScanResult)) error
	StopScan() error
}

type BLEScanner interface {
	Scan(ctx context.Context, handler func(bluetooth.ScanResult)) error
}

type BluetoothScanner struct {
	Adapter Adapter
}

func (s *BluetoothScanner) Scan(ctx context.Context, handler func(sr bluetooth.ScanResult)) error {
	errCh := make(chan error, 1)
	go func() {
		err := s.Adapter.Scan(func(_ *bluetooth.Adapter, result bluetooth.ScanResult) {
			handler(result)
		})
		errCh <- err
	}()
	select {
	case <-ctx.Done():
		if err := s.Adapter.StopScan(); err != nil {
			return err
		}
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
