package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/BurntSushi/toml"
)

type RuuviTag struct {
	MAC  string `toml:"mac"`
	Name string `toml:"name"`
}

type Config struct {
	ReportingInterval Duration   `toml:"reporting_interval"`
	RuuviTag          []RuuviTag `toml:"ruuvitag"`
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

func ReadConfig(fileName string) (cfg Config, err error) {
	var blob []byte
	cfgFile := os.Getenv("RUUVITAG_CONFIG_FILE")
	if cfgFile != "" {
		blob, err = ioutil.ReadFile(cfgFile)
		if err != nil {
			return
		}
		_, err = toml.Decode(string(blob), &cfg)
		return
	}
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
	filePath = fileName
	blob, err = ioutil.ReadFile(filePath)
	if err == nil {
		_, err = toml.Decode(string(blob), &cfg)
		return
	}
	filePath = path.Join("configs", fileName)
	blob, err = ioutil.ReadFile(filePath)
	if err == nil {
		_, err = toml.Decode(string(blob), &cfg)
		return
	}
	err = fmt.Errorf("configuration file %s not found", fileName)
	return
}
