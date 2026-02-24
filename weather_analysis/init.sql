-- Create weather table
CREATE TABLE weather (
    time TIMESTAMPTZ NOT NULL,
    city TEXT DEFAULT 'Beijing',
    temperature FLOAT,
    precipitation FLOAT,
    humidity FLOAT,
    wind_speed FLOAT,
    pressure FLOAT
);

-- Convert to TimescaleDB hypertable
SELECT create_hypertable('weather', 'time');

-- Create indexes for faster queries
CREATE INDEX idx_weather_city_time ON weather(city, time);

-- Create materialized views for common aggregations
CREATE MATERIALIZED VIEW weather_daily AS
SELECT 
    time_bucket('1 day', time) AS day,
    city,
    AVG(temperature) AS avg_temp,
    SUM(precipitation) AS total_precipitation,
    AVG(humidity) AS avg_humidity,
    AVG(wind_speed) AS avg_wind_speed,
    AVG(pressure) AS avg_pressure
FROM weather
GROUP BY day, city;

CREATE MATERIALIZED VIEW weather_seasonal AS
SELECT 
    time_bucket('3 months', time) AS season,
    city,
    AVG(temperature) AS avg_temp,
    SUM(precipitation) AS total_precipitation,
    AVG(humidity) AS avg_humidity
FROM weather
GROUP BY season, city;

CREATE MATERIALIZED VIEW weather_yearly AS
SELECT 
    time_bucket('1 year', time) AS year,
    city,
    AVG(temperature) AS avg_temp,
    SUM(precipitation) AS total_precipitation,
    AVG(humidity) AS avg_humidity
FROM weather
GROUP BY year, city;

-- Create indexes on materialized views
CREATE INDEX idx_weather_daily_day ON weather_daily(day);
CREATE INDEX idx_weather_seasonal_season ON weather_seasonal(season);
CREATE INDEX idx_weather_yearly_year ON weather_yearly(year);

-- Refresh materialized views
REFRESH MATERIALIZED VIEW weather_daily;
REFRESH MATERIALIZED VIEW weather_seasonal;
REFRESH MATERIALIZED VIEW weather_yearly;