from flask import Flask, jsonify, request
from flask_cors import CORS
import os
import sqlite3

app = Flask(__name__)
CORS(app)  # Enable CORS for frontend integration

# Database configuration
DB_TYPE = 'sqlite'  # Using SQLite exclusively

# SQLite connection
def get_db_connection():
    db_path = os.path.join(os.path.dirname(__file__), 'weather.db')
    conn = sqlite3.connect(db_path)
    conn.row_factory = sqlite3.Row
    return conn

# API endpoint for yearly weather data
@app.route('/api/weather/yearly', methods=['GET'])
def get_yearly_weather():
    conn = get_db_connection()
    
    try:
        cur = conn.cursor()
        cur.execute('SELECT * FROM weather_yearly ORDER BY year')
        data = [dict(row) for row in cur.fetchall()]
        cur.close()
        
        return jsonify(data)
    except Exception as e:
        return jsonify({'error': str(e)}), 500
    finally:
        conn.close()

# API endpoint for seasonal weather data
@app.route('/api/weather/seasonal', methods=['GET'])
def get_seasonal_weather():
    conn = get_db_connection()
    
    try:
        cur = conn.cursor()
        cur.execute('SELECT * FROM weather_seasonal ORDER BY season')
        data = [dict(row) for row in cur.fetchall()]
        cur.close()
        
        return jsonify(data)
    except Exception as e:
        return jsonify({'error': str(e)}), 500
    finally:
        conn.close()

# API endpoint for daily weather data
@app.route('/api/weather/daily', methods=['GET'])
def get_daily_weather():
    conn = get_db_connection()
    
    try:
        # Get date range from query parameters
        start_date = request.args.get('start_date')
        end_date = request.args.get('end_date')
        
        cur = conn.cursor()
        if start_date and end_date:
            cur.execute('SELECT * FROM weather_daily WHERE day BETWEEN ? AND ? ORDER BY day', (start_date, end_date))
        else:
            cur.execute('SELECT * FROM weather_daily ORDER BY day LIMIT 365')  # Limit to 1 year
        data = [dict(row) for row in cur.fetchall()]
        cur.close()
        
        return jsonify(data)
    except Exception as e:
        return jsonify({'error': str(e)}), 500
    finally:
        conn.close()

# API endpoint for weather comparison
@app.route('/api/weather/compare', methods=['GET'])
def compare_weather():
    conn = get_db_connection()
    
    try:
        # Get comparison parameters
        period = request.args.get('period', 'yearly')  # yearly, seasonal, daily
        year1 = request.args.get('year1')
        year2 = request.args.get('year2')
        
        if period == 'yearly' and year1 and year2:
            cur = conn.cursor()
            cur.execute('SELECT * FROM weather_yearly WHERE year IN (?, ?) ORDER BY year', (year1, year2))
            data = [dict(row) for row in cur.fetchall()]
            cur.close()
        else:
            return jsonify({'error': 'Invalid comparison parameters'}), 400
        
        return jsonify(data)
    except Exception as e:
        return jsonify({'error': str(e)}), 500
    finally:
        conn.close()

if __name__ == '__main__':
    app.run(debug=True, port=5001)