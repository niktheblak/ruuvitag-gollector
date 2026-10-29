# ruuvitag-gollector

RuuviTags are an amazing way to collect environmental data. But sometimes, you want full control over how that data is stored, visualized, or processed. ruuvitag-gollector is a self-hosted solution that listens to BLE broadcasts and pushes sensor data to platforms like InfluxDB, PostgreSQL, AWS, GCP and more.

Whether you're monitoring a greenhouse, a guitar case, a terrarium, or just love collecting atmospheric data with Raspberry Pi; this project gives you full flexibility without needing a subscription. You will need to self-host or host in the cloud some the server stuff, such as a database and Grafana though. But for collecting and seeing their data, all you need is the cheapest possible Raspberry Pi and a Go compiler.

Supports the RAWv2 format (data format 5) emitted by RuuviTags with 2.x or later
firmware, and the older RAWv1 format (data format 3). RAWv1 tags do not report
TX power, movement counter or measurement number, and their battery voltage is
in millivolts rather than volts.

## Setup

The collector targets Linux and scans through the BlueZ D-Bus API. Ensure the
BlueZ service is installed, running, and has access to the Bluetooth adapter.

Compile and install the `ruuvitag-gollector` binary:

```bash
git clone https://github.com/niktheblak/ruuvitag-gollector.git
cd ruuvitag-gollector
task build
task install
```

Then copy the example configuration file `configs/example-config.toml` to `$HOME/.ruuvitag-gollector/config.toml` or `/etc/ruuvitag-gollector/config.toml` if you're installing the collector globally.
and fill your preferred configuration values. For reference of possible configuration
values, run `ruuvitag-gollector -h`.

At the very least you need to add the MAC addresses and human-readable names of your
RuuviTags into the config file under the `ruuvitags` key:

```toml
[ruuvitags]
"CC:CA:7E:52:CC:34" = "Backyard"
"FB:E1:B7:04:95:EE" = "Upstairs"
"E8:E0:C6:0B:B8:C5" = "Downstairs"
```

If you want to save data to InfluxDB (local or remote), add the following options to your config file:

```toml
[exporters.influxdb]
type = "influxdb"
addr = "http://localhost:8086"
org = "my-org"
bucket = "ruuvitag"
measurement = "ruuvitag"
token = "abc123"
```

`org` is required by InfluxDB 2.x. `measurement` defaults to `ruuvitag`. Each
measurement is written as one point with `mac` (upper-cased) and `name` as tags
and the remaining enabled columns as fields.

For a complete configuration example, see [example config](#complete-example-configuration).

The following exporters are supported for sending measurements:

- InfluxDB
- PostgreSQL (and TimescaleDB)
- Webhook, meaning a URL that accepts an HTTP POST request with the measurement as JSON in the request body. A non-2xx response is treated as a failed export.
- AWS DynamoDB
- AWS SQS (standard and FIFO queues; FIFO messages are grouped per tag)
- GCP Pub/Sub (`mac` and `name` are sent as message attributes, the rest as the JSON body)
- MQTT (one message per measurement on topic `ruuvitag-gollector/<name>/<mac without colons>`)

All exporters except the console exporter serialize measurements using the
column names from the `[columns]` table described below. The AWS exporters use
the configured static credentials when `access_key_id` and `secret_access_key`
are set and the default AWS credential chain otherwise.

See the command-line help for the arguments needed by each exporter:

```bash
ruuvitag-gollector -h
```

You can specify custom column names for the data fields ruuvitag-gollector sends out to match your database schema.
You can also exclude columns whose value you don't need. The `time` column and
at least one of `mac` and `name` must be present; unknown column keys are
rejected at startup. The default columns and their names are:

```toml
[columns]
time = "time"
mac = "mac"
name = "name"
temperature = "temperature"
humidity = "humidity"
pressure = "pressure"
acceleration_x = "acceleration_x"
acceleration_y = "acceleration_y"
acceleration_z = "acceleration_z"
movement_counter = "movement_counter"
measurement_number = "measurement_number"
dew_point = "dew_point"
battery_voltage = "battery_voltage"
tx_power = "tx_power"
wet_bulb = "wet_bulb"
```

The same list is printed by `ruuvitag-gollector print columns`.

### PostgreSQL and TimescaleDB schema

The PostgreSQL exporter does not create or alter tables. It runs a plain
`INSERT` into the configured `table` with one column per enabled entry in
`[columns]`, so the table must exist beforehand with matching column names.

A measurement row is identified by `(time, mac)`. The `name` column is the label
from the `[ruuvitags]` table and is written to every row, but it is not part of
the key: labels may be changed over the life of a tag, so queries should group
on `mac` and treat `name` as display text.

`scripts/timescaledb/create_table.sql` creates the table with the default
column names, a `(time, mac)` primary key, and a TimescaleDB hypertable
partitioned on `time`. `scripts/timescaledb/create_aggregates.sql` adds 5 minute
and 30 minute continuous aggregates grouped by `mac`. Both scripts can be
adapted for plain PostgreSQL by dropping the TimescaleDB calls.

## Running

Now you can try to run it manually. The process must be allowed to use BlueZ on
the system bus and to power on the selected adapter; running with `sudo` is the
simplest initial setup:

```bash
sudo ruuvitag-gollector scan
```

To collect values continuously, run:

```bash
sudo ruuvitag-gollector daemon
```

## Complete Example Configuration

```toml
interval = "0m"
device = "default"

[ruuvitags]
"CC:CA:7E:52:CC:34" = "Backyard"
"FB:E1:B7:04:95:EE" = "Upstairs"
"E8:E0:C6:0B:B8:C5" = "Downstairs"

[exporters]

[exporters.influxdb]
type = "influxdb"
addr = "https://eu-central-1-1.aws.cloud2.influxdata.com"
org = "my-org"
bucket = "ruuvitag"
measurement = "ruuvitag"
token = "abc123"
async = true
batch_size = 20
flush_interval = "5s"

[exporters.dynamodb]
type = "dynamodb"
table = "ruuvitag"
access_key_id = "MYAWSAKKESSKEY"
secret_access_key = "my+aws+secret+key"
region = "us-east-2"

[exporters.sqs]
type = "sqs"
queue.url = "https://us-east-2.queue.amazonaws.com/321667262165/measurements"
access_key_id = "MYAWSAKKESSKEY"
secret_access_key = "my+aws+secret+key"
region = "us-east-2"

[exporters.postgres]
type = "postgres"
host = "my-instance-name.eu-central-1.aws.neon.tech"
port = 5432
database = "ruuvitag"
username = "myorg@example.com"
password = "some_secret_password"
table = "ruuvitag"
sslmode = "require"

[exporters.http]
type = "http"
url = "https://my-api.herokuapp.com/receive"
token = "MyHerokuToken"
  
[exporters.mqtt]
type = "mqtt"
addr = "tcp://localhost:8883"
client_id = "ruuvitag-gollector"
username = "mqtt_user"
password = "my_secret_password"
ca_file = "root_ca.pem"
auto_reconnect = true
# maximum wait between reconnection attempts, in seconds or as a duration string such as "30s"; default 10 minutes
reconnect_interval = 30

[columns]
time = "time"
mac = "mac"
name = "name"
temperature = "temperature"
humidity = "humidity"
pressure = "pressure"
movement_counter = "movementCounter"
measurement_number = "measurementNumber"
```

`device` may be `default`, a BlueZ adapter name such as `hci0`, or the adapter's
Bluetooth address. `default` selects the first adapter reported by BlueZ.
