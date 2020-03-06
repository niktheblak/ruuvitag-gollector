CREATE TABLE IF NOT EXISTS measurements (
    id SERIAL PRIMARY KEY,
    mac TEXT NOT NULL,
    name TEXT NOT NULL,
    ts TIMESTAMP NOT NULL,
    temperature REAL NOT NULL,
    humidity REAL NOT NULL,
    pressure REAL NOT NULL,
    acceleration_x INTEGER NOT NULL,
    acceleration_y INTEGER NOT NULL,
    acceleration_z INTEGER NOT NULL,
    movement_counter INTEGER NOT NULL,
    battery INTEGER NOT NULL
);