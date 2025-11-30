CREATE TABLE IF NOT EXISTS weather_states (
    state_id UUID PRIMARY KEY,
    cell_id UUID NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    state_type VARCHAR(20) NOT NULL, -- clear, cloudy, rain, snow, storm
    temperature FLOAT NOT NULL, -- °C
    precipitation FLOAT NOT NULL, -- mm/day
    wind_direction FLOAT NOT NULL, -- degrees
    wind_speed FLOAT NOT NULL, -- m/s
    humidity FLOAT NOT NULL, -- 0-100%
    visibility FLOAT NOT NULL, -- km
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_weather_cell_time ON weather_states (cell_id, timestamp DESC);
CREATE INDEX idx_weather_timestamp ON weather_states (timestamp);

-- Weather history (aggregated daily)
CREATE TABLE IF NOT EXISTS weather_history (
    cell_id UUID NOT NULL,
    date DATE NOT NULL,
    avg_temperature FLOAT,
    total_precipitation FLOAT,
    avg_wind_speed FLOAT,
    PRIMARY KEY (cell_id, date)
);

CREATE INDEX idx_weather_history_date ON weather_history (date);

-- Extreme weather events
CREATE TABLE IF NOT EXISTS extreme_weather_events (
    event_id UUID PRIMARY KEY,
    event_type VARCHAR(20) NOT NULL, -- hurricane, blizzard, drought, heatwave
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    duration_hours INT NOT NULL,
    intensity FLOAT NOT NULL, -- 0-1 scale
    affected_cells UUID[] NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_extreme_events_time ON extreme_weather_events (start_time);
