-- Measurement table for ruuvitag-gollector's PostgreSQL / TimescaleDB exporter.
--
-- The exporter (pkg/exporter/postgres/postgres.go) never creates or alters
-- tables; it only runs INSERT INTO <table>(<columns>) VALUES (...) for the
-- columns enabled in the [columns] config. The column names below are the
-- defaults from ruuvitag-common's sensor.DefaultColumns, in insertion order.
-- Rename columns here and in [columns] together.
--
-- Key: (time, mac). A tag is identified by its MAC address. The name column
-- is a human-readable label that may change over the life of a tag, so it is
-- not part of the key; group and join on mac, and treat name as display text.
--
-- Zero means "not available" for battery_voltage, tx_power, dew_point and
-- wet_bulb. Data format 3 (RAWv1) tags always report 0 for tx_power,
-- movement_counter and measurement_number.

CREATE TABLE ruuvitag (
    time               TIMESTAMPTZ      NOT NULL,  -- collector receive time
    mac                MACADDR          NOT NULL,  -- tag address, sent lowercase
    name               TEXT,                       -- label from [ruuvitags] config
    temperature        DOUBLE PRECISION,           -- °C
    humidity           DOUBLE PRECISION,           -- % relative humidity
    pressure           DOUBLE PRECISION,           -- hPa
    acceleration_x     INTEGER,                    -- milli-g
    acceleration_y     INTEGER,                    -- milli-g
    acceleration_z     INTEGER,                    -- milli-g
    movement_counter   INTEGER,                    -- 0..254, wraps
    measurement_number INTEGER,                    -- 0..65534, wraps
    dew_point          DOUBLE PRECISION,           -- °C, computed by collector
    battery_voltage    DOUBLE PRECISION,           -- V (data format 5), mV (data format 3)
    tx_power           INTEGER,                    -- dBm
    wet_bulb           DOUBLE PRECISION,           -- °C, computed by collector
    PRIMARY KEY (time, mac)
);

-- TimescaleDB requires the partitioning column in every unique index, which
-- the (time, mac) primary key satisfies.
SELECT create_hypertable('ruuvitag', by_range('time'));
