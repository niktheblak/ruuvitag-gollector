# ruuvitag-gollector

Collects data from RuuviTag sensors to InfluxDB and other databases.

Supports the RAWv2 format emitted by RuuviTags with 2.x firmware.

## Setup

Compile and install the `ruuvitag-gollector` binary:

```bash
git clone https://github.com/niktheblak/ruuvitag-gollector.git
cd ruuvitag-gollector
make build
make install
```

Then copy the example configuration file `example-config.yaml` to `$HOME/ruuvitag-gollector.yaml`
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

For a complete configuration example, see [example config](#complete-example-configuration).

The following exporters are supported for sending measurements:

- InfluxDB
- PostgreSQL
- Webhook, meaning a URL that accepts an HTTP POST request with the measurement as JSON in the request body
- AWS DynamoDB
- AWS SQS
- GCP Pub/Sub

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

```yaml
interval: 0m

ruuvitags:
  "CC:CA:7E:52:CC:34": Backyard
  "FB:E1:B7:04:95:EE": Upstairs
  "E8:E0:C6:0B:B8:C5": Downstairs

influxdb:
  enabled: true
  host: http://localhost:8086
  database: ruuvitag
  measurement: ruuvitag_measurement

aws:
  access_key_id: MYAWSAKKESSKEY
  secret_access_key: "my+aws+secret+key"
  region: us-east-2
  dynamodb:
    enabled: true
    table: ruuvitag
  sqs:
    enabled: true
    queue.url: "https://us-east-2.queue.amazonaws.com/321667262165/measurements"

postgres:
  enabled: true
  conn: "postgres://postgres:mysecretpassword@postgres/postgres?sslmode=disable"
  table: measurements

http:
  enabled: true
  url: https://my-api.herokuapp.com/receive
  token: MyHerokuToken
```
