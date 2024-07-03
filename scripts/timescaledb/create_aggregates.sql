CREATE MATERIALIZED VIEW ruuvitag_aggregated_5min
WITH (timescaledb.continuous) AS
    SELECT
        time_bucket('5min', time) AS bucket,
        name,
        AVG(temperature) as "temperature",
        AVG(humidity) as "humidity",
        AVG(pressure) as "pressure",
        LAST(movement_counter, time) as "movement_counter",
        LAST(battery_voltage, time) as "battery_voltage",
        LAST(measurement_number, time) as "measurement_number",
        AVG(dew_point) as "dew_point"
    FROM ruuvitag
    GROUP BY bucket, name;

SELECT add_continuous_aggregate_policy('ruuvitag_aggregated_5min',
  start_offset => NULL,
  end_offset => INTERVAL '5min',
  schedule_interval => INTERVAL '5min');

CREATE MATERIALIZED VIEW ruuvitag_aggregated_30min
WITH (timescaledb.continuous) AS
    SELECT
        time_bucket('30min', time) AS bucket,
        name,
        AVG(temperature) as "temperature",
        AVG(humidity) as "humidity",
        AVG(pressure) as "pressure",
        LAST(movement_counter, time) as "movement_counter",
        LAST(battery_voltage, time) as "battery_voltage",
        LAST(measurement_number, time) as "measurement_number",
        AVG(dew_point) as "dew_point"
    FROM ruuvitag
    GROUP BY bucket, name;

SELECT add_continuous_aggregate_policy('ruuvitag_aggregated_30min',
  start_offset => NULL,
  end_offset => INTERVAL '30min',
  schedule_interval => INTERVAL '3h');
