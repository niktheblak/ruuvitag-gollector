package mqtt

import (
	"fmt"
	"log/slog"
	"time"
)

const DefaultClientID = "ruuvitag-gollector"

type Config struct {
	Addr              string
	ClientId          string
	Username          string
	Password          string
	CaFile            string
	AutoReconnect     bool
	ReconnectInterval time.Duration
	Columns           map[string]string
	Logger            *slog.Logger
}

func Validate(cfg Config) error {
	if cfg.Addr == "" {
		return fmt.Errorf("MQTT broker address must be specified")
	}
	if len(cfg.Columns) == 0 {
		return fmt.Errorf("columns must be non-empty")
	}
	return nil
}
