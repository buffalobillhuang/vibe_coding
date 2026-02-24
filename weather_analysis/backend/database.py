import sqlite3
import os

# Create database connection
db_path = os.path.join(os.path.dirname(__file__), 'weather.db')

def get_db_connection():
    conn = sqlite3.connect(db_path)
    conn.row_factory = sqlite3.Row
    return conn

# Initialize database schema
def init_database():
    conn = get_db_connection()
    cursor = conn.cursor()
    
    # Create weather table
    cursor.execute('''
    CREATE TABLE IF NOT EXISTS weather (
        time TEXT NOT NULL,
        city TEXT DEFAULT 'Beijing',
        temperature REAL,
        precipitation REAL,
        humidity REAL,
        wind_speed REAL,
        pressure REAL
    )
    ''')
    
    # Create indexes for faster queries
    cursor.execute('CREATE INDEX IF NOT EXISTS idx_weather_city_time ON weather(city, time)')
    
    # Create materialized views (as tables in SQLite)
    cursor.execute('''
    CREATE TABLE IF NOT EXISTS weather_daily (
        day TEXT,
        city TEXT,
        avg_temp REAL,
        total_precipitation REAL,
        avg_humidity REAL,
        avg_wind_speed REAL,
        avg_pressure REAL,
        PRIMARY KEY (day, city)
    )
    ''')
    
    cursor.execute('''
    CREATE TABLE IF NOT EXISTS weather_seasonal (
        season TEXT,
        city TEXT,
        avg_temp REAL,
        total_precipitation REAL,
        avg_humidity REAL,
        PRIMARY KEY (season, city)
    )
    ''')
    
    cursor.execute('''
    CREATE TABLE IF NOT EXISTS weather_yearly (
        year TEXT,
        city TEXT,
        avg_temp REAL,
        total_precipitation REAL,
        avg_humidity REAL,
        PRIMARY KEY (year, city)
    )
    ''')
    
    conn.commit()
    conn.close()

# Refresh materialized views
def refresh_materialized_views():
    conn = get_db_connection()
    cursor = conn.cursor()
    
    # Refresh daily view
    cursor.execute('DELETE FROM weather_daily')
    cursor.execute('''
    INSERT INTO weather_daily
    SELECT 
        date(time) AS day,
        city,
        AVG(temperature) AS avg_temp,
        SUM(precipitation) AS total_precipitation,
        AVG(humidity) AS avg_humidity,
        AVG(wind_speed) AS avg_wind_speed,
        AVG(pressure) AS avg_pressure
    FROM weather
    GROUP BY day, city
    ''')
    
    # Refresh seasonal view (simplified: 3-month periods)
    cursor.execute('DELETE FROM weather_seasonal')
    cursor.execute('''
    INSERT INTO weather_seasonal
    SELECT 
        strftime('%Y-%m', time, '-1 month', 'start of month') || '-01' AS season,
        city,
        AVG(temperature) AS avg_temp,
        SUM(precipitation) AS total_precipitation,
        AVG(humidity) AS avg_humidity
    FROM weather
    GROUP BY season, city
    ''')
    
    # Refresh yearly view
    cursor.execute('DELETE FROM weather_yearly')
    cursor.execute('''
    INSERT INTO weather_yearly
    SELECT 
        strftime('%Y', time) AS year,
        city,
        AVG(temperature) AS avg_temp,
        SUM(precipitation) AS total_precipitation,
        AVG(humidity) AS avg_humidity
    FROM weather
    GROUP BY year, city
    ''')
    
    conn.commit()
    conn.close()

# Initialize database on module load
init_database()