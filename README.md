# ruuvitag-gollector

Collects data from RuuviTag sensors to InfluxDB and other databases.

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
[influxdb]
enabled = true
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
console = false

[ruuvitags]
"CC:CA:7E:52:CC:34" = "Backyard"
"FB:E1:B7:04:95:EE" = "Upstairs"
"E8:E0:C6:0B:B8:C5" = "Downstairs"

[influxdb]
enabled = true
addr = "https://eu-central-1-1.aws.cloud2.influxdata.com"
bucket = "ruuvitag"
measurement = "ruuvitag"
token = "abc123"
async = true
batch_size = 20
flush_interval = "5s"

[aws]
access_key_id = "MYAWSAKKESSKEY"
secret_access_key = "my+aws+secret+key"
region = "us-east-2"

[aws.dynamodb]
enabled = true
table = "ruuvitag"
[aws.sqs]
enabled = true
queue.url = "https://us-east-2.queue.amazonaws.com/321667262165/measurements"

[postgres]
enabled = true
host = "my-instance-name.eu-central-1.aws.neon.tech"
port = 5432
database = "ruuvitag"
username = "myorg@example.com"
password = "some_secret_password"
table = "ruuvitag"
sslmode = "require"

[postgres.column]
time = "time"

[http]
enabled = true
url = "https://my-api.herokuapp.com/receive"
token = "MyHerokuToken"
  
[mqtt]
enabled = true
addr = "tcp://localhost:8883"
client_id = "ruuvitag-gollector"
username = "mqtt_user"
password = "my_secret_password"
ca_file = "root_ca.pem"
auto_reconnect = true
reconnect_interval = 30
```
