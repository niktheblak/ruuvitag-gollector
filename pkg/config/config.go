package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

type RuuviTag struct {
	ID   string `toml:"id"`
	Name string `toml:"name"`
}

type Config struct {
	ReportingInterval Duration   `toml:"reporting_interval"`
	RuuviTags         []RuuviTag `toml:"ruuvitag"`
}

func (c Config) Validate() error {
	if c.ReportingInterval.Duration == 0 {
		return fmt.Errorf("reporting interval must be set")
	}
	return nil
}

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

func Read(fileName string) (cfg Config, err error) {
	var blob []byte
	if filepath.IsAbs(fileName) {
		blob, err = ioutil.ReadFile(fileName)
		if err != nil {
			return
		}
		_, err = toml.Decode(string(blob), &cfg)
		return
	}

	// Try to read the file from the current directory
	blob, err = ioutil.ReadFile(fileName)
	if err == nil {
		_, err = toml.Decode(string(blob), &cfg)
		return
	}

	// Try to read the file from the user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	filePath := path.Join(home, fileName)
	blob, err = ioutil.ReadFile(filePath)
	if err == nil {
		_, err = toml.Decode(string(blob), &cfg)
		return
	}

	// Try to read the file from the /config directory
	filePath = path.Join("/config", fileName)
	blob, err = ioutil.ReadFile(filePath)
	if err == nil {
		_, err = toml.Decode(string(blob), &cfg)
		return
	}

	err = fmt.Errorf("configuration file %s not found", fileName)
	return
}
