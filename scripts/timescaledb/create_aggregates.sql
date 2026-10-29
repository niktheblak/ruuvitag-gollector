-- Continuous aggregates over the ruuvitag hypertable (see create_table.sql).
--
-- Buckets are grouped by mac, the stable identity of a tag. The name label
-- can change over time, so grouping by it would split one tag's history at
-- every rename. The most recent name in each bucket is carried along for
-- display purposes only.

CREATE MATERIALIZED VIEW ruuvitag_aggregated_5min
WITH (timescaledb.continuous) AS
    SELECT
        time_bucket('5min', time) AS bucket,
        mac,
        LAST(name, time) AS name,
        AVG(temperature) AS temperature,
        AVG(humidity) AS humidity,
        AVG(pressure) AS pressure,
        LAST(movement_counter, time) AS movement_counter,
        LAST(battery_voltage, time) AS battery_voltage,
        LAST(measurement_number, time) AS measurement_number,
        AVG(dew_point) AS dew_point,
        AVG(wet_bulb) AS wet_bulb
    FROM ruuvitag
    GROUP BY bucket, mac;

SELECT add_continuous_aggregate_policy('ruuvitag_aggregated_5min',
  start_offset => NULL,
  end_offset => INTERVAL '5min',
  schedule_interval => INTERVAL '5min');

CREATE MATERIALIZED VIEW ruuvitag_aggregated_30min
WITH (timescaledb.continuous) AS
    SELECT
        time_bucket('30min', time) AS bucket,
        mac,
        LAST(name, time) AS name,
        AVG(temperature) AS temperature,
        AVG(humidity) AS humidity,
        AVG(pressure) AS pressure,
        LAST(movement_counter, time) AS movement_counter,
        LAST(battery_voltage, time) AS battery_voltage,
        LAST(measurement_number, time) AS measurement_number,
        AVG(dew_point) AS dew_point,
        AVG(wet_bulb) AS wet_bulb
    FROM ruuvitag
    GROUP BY bucket, mac;

SELECT add_continuous_aggregate_policy('ruuvitag_aggregated_30min',
  start_offset => NULL,
  end_offset => INTERVAL '30min',
  schedule_interval => INTERVAL '3h');
