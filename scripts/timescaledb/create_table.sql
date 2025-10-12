CREATE TABLE ruuvitag (
	time TIMESTAMPTZ NOT NULL,
	mac MACADDR NOT NULL,
	name TEXT,
	temperature DOUBLE PRECISION,
	humidity DOUBLE PRECISION,
	pressure DOUBLE PRECISION,
	acceleration_x INTEGER,
	acceleration_y INTEGER,
	acceleration_z INTEGER,
	movement_counter INTEGER,
	measurement_number INTEGER,
	dew_point DOUBLE PRECISION,
	wet_bulb DOUBLE PRECISION,
	battery_voltage DOUBLE PRECISION,
    tx_power INTEGER);

SELECT create_hypertable('ruuvitag', by_range('time'));
