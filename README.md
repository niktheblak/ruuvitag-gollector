# ruuvitag-gollector

ruuvitag-gollector collects Bluetooth sensor data from RuuviTags and exports it to time-series databases and cloud platforms.

Supports the RAWv2 format emitted by RuuviTags with 2.x firmware.

## Setup

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
[expoters.influxdb]
type = "influxdb"
addr = "http://localhost:8086"
bucket = "ruuvitag"
measurement = "ruuvitag"
token = "abc123"
```

For a complete configuration example, see [example config](#complete-example-configuration).

The following exporters are supported for sending measurements:

- InfluxDB
- PostgreSQL (and TimescaleDB)
- Webhook, meaning a URL that accepts an HTTP POST request with the measurement as JSON in the request body
- AWS DynamoDB
- AWS SQS
- GCP Pub/Sub
- MQTT

See the command-line help for the arguments needed by each exporter:

```bash
ruuvitag-gollector -h
```

You can specify custom column names for the data fields ruuvitag-gollector sends out to match your database schema.
You can also exclude columns whose value you don't need. The default columns and their names are:

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
```

## Running

Now you can try to run it manually (you typically need to run as root to allow the collector
process access to Bluetooth hardware):

```bash
sudo ruuvitag-gollector collect
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
