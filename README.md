# ruuvitag-gollector

Collects data from RuuviTag sensors to InfluxDB and other databases.

Supports the RAWv2 format emitted by RuuviTags with 2.x firmware.

## Setup

Install the `ruuvitag-gollector` binary into your local `$HOME/go/bin` with:

```bash
go install
```

Then copy the example configuration file `example-config.yaml` to `$HOME/.ruuvitag-gollector.yaml`
and fill your preferred configuration values. For reference of possible configuration
values, run `ruuvitag-gollector -h`.

At the very least you need to add the MAC addresses and human-readable names of your
RuuviTags into the config file under the `ruuvitags` key:

```yaml
ruuvitags:
  "CC:CA:7E:52:CC:34": Backyard
  "FB:E1:B7:04:95:EE": Upstairs
  "E8:E0:C6:0B:B8:C5": Downstairs
```

If you want to save data to InfluxDB (local or remote), add the following options to your config file:

```yaml
influxdb:
  enabled: true
  addr: http://localhost:8086
  database: ruuvitag
  measurement: ruuvitag
  username: root
  password: root
```

For other storage options, see the command help

```bash
ruuvitag-gollector -h
```

Now you can try to run it manually (you typically need to run as root to allow the collector
process access to Bluetooth hardware):

```bash
sudo ruuvitag-gollector collect
```

To collect values continuously, run:

```bash
sudo ruuvitag-gollector daemon
```
