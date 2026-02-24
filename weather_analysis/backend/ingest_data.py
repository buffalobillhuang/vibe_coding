import sqlite3
import os
import random
from datetime import datetime, timedelta

# Database configuration
DB_TYPE = 'sqlite'  # Force SQLite for now
db_path = os.path.join(os.path.dirname(__file__), 'weather.db')

# Get database connection
def get_db_connection():
    # SQLite connection
    return sqlite3.connect(db_path)

# Generate sample weather data
def generate_sample_data(start_year=1976, end_year=2026):
    data = []
    start_date = datetime(start_year, 1, 1)
    end_date = datetime(end_year, 12, 31)
    current_date = start_date
    
    # Base temperature with trend (global warming)
    base_temp = 12.0
    temp_increase_per_year = 0.03
    
    while current_date <= end_date:
        # Calculate year effect (global warming)
        year_factor = (current_date.year - start_year) * temp_increase_per_year
        
        # Seasonal variation
        month = current_date.month
        if month in [12, 1, 2]:  # Winter
            seasonal_factor = -5.0
        elif month in [3, 4, 5]:  # Spring
            seasonal_factor = 5.0
        elif month in [6, 7, 8]:  # Summer
            seasonal_factor = 15.0
        else:  # Fall
            seasonal_factor = 5.0
        
        # Daily variation
        daily_variation = random.uniform(-2.0, 2.0)
        
        # Calculate temperature
        temperature = base_temp + year_factor + seasonal_factor + daily_variation
        
        # Precipitation (higher in summer)
        if month in [6, 7, 8]:
            precipitation = random.uniform(0, 10)
        else:
            precipitation = random.uniform(0, 2)
        
        # Humidity (higher in summer)
        humidity = random.uniform(40, 90)
        
        # Wind speed
        wind_speed = random.uniform(0, 10)
        
        # Pressure
        pressure = random.uniform(990, 1030)
        
        data.append({
            'time': current_date.strftime('%Y-%m-%d %H:%M:%S'),
            'city': 'Beijing',
            'temperature': round(temperature, 1),
            'precipitation': round(precipitation, 1),
            'humidity': round(humidity, 1),
            'wind_speed': round(wind_speed, 1),
            'pressure': round(pressure, 1)
        })
        
        # Move to next day
        current_date += timedelta(days=1)
    
    return data

# Ingest data into database
def ingest_data(data):
    conn = get_db_connection()
    cursor = conn.cursor()
    
    try:
        # Insert weather data
        if DB_TYPE == 'postgres':
            insert_query = '''
            INSERT INTO weather (time, city, temperature, precipitation, humidity, wind_speed, pressure)
            VALUES (%s, %s, %s, %s, %s, %s, %s)
            '''
        else:
            insert_query = '''
            INSERT INTO weather (time, city, temperature, precipitation, humidity, wind_speed, pressure)
            VALUES (?, ?, ?, ?, ?, ?, ?)
            '''
        
        # Batch insert
        batch_size = 1000
        for i in range(0, len(data), batch_size):
            batch = data[i:i+batch_size]
            if DB_TYPE == 'postgres':
                cursor.executemany(insert_query, [(d['time'], d['city'], d['temperature'], 
                                                 d['precipitation'], d['humidity'], d['wind_speed'], d['pressure']) for d in batch])
            else:
                cursor.executemany(insert_query, [(d['time'], d['city'], d['temperature'], 
                                                 d['precipitation'], d['humidity'], d['wind_speed'], d['pressure']) for d in batch])
            conn.commit()
            print(f"Inserted {min(i+batch_size, len(data))}/{len(data)} records")
        
        # Refresh materialized views
        if DB_TYPE == 'postgres':
            cursor.execute('REFRESH MATERIALIZED VIEW weather_daily')
            cursor.execute('REFRESH MATERIALIZED VIEW weather_seasonal')
            cursor.execute('REFRESH MATERIALIZED VIEW weather_yearly')
        else:
            # For SQLite, recreate the views
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
        print("Data ingestion completed successfully!")
        
    except Exception as e:
        print(f"Error during ingestion: {e}")
        conn.rollback()
    finally:
        conn.close()

# Main function
if __name__ == '__main__':
    print("Generating sample weather data for Beijing (1976-2026)...")
    sample_data = generate_sample_data()
    print(f"Generated {len(sample_data)} records")
    
    print("Ingesting data into database...")
    ingest_data(sample_data)
    print("Process completed!")