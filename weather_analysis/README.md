# Beijing Weather Analysis (1976-2026)

A comprehensive weather analysis website for Beijing's 50-year weather trends (1976-2026), featuring interactive visualizations with yearly, seasonal, and daily comparisons.

## Table of Contents
- [Features](#features)
- [Project Structure](#project-structure)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Running the Services](#running-the-services)
- [Stopping the Services](#stopping-the-services)
- [Verification](#verification)
- [API Endpoints](#api-endpoints)
- [Data Storage](#data-storage)
- [Deployment](#deployment)

## Features

### Backend
- Python/Flask API server with RESTful endpoints
- SQLite database with optimized schema and materialized views
- PostgreSQL with TimescaleDB support for production scalability
- Data ingestion script for generating realistic weather data

### Frontend
- React + Recharts for interactive data visualizations
- Multiple view modes: yearly trends, seasonal patterns, daily variations
- Year-to-year weather comparison functionality
- Responsive design with mobile-friendly layout
- Sample data fallback for demonstration

## Project Structure

```
weather_analysis/
├── backend/             # Python backend with Flask API
│   ├── app.py           # Main API server
│   ├── database.py      # Database initialization
│   ├── ingest_data.py   # Data ingestion script
│   ├── weather.db       # SQLite database
│   └── requirements.txt # Python dependencies
├── frontend/            # React frontend
│   ├── src/             # React source code
│   │   ├── App.jsx      # Main component with visualizations
│   │   ├── main.jsx     # Entry point
│   │   └── index.css    # Styling
│   ├── package.json     # npm dependencies
│   └── vite.config.js   # Vite configuration
├── docker-compose.yml   # PostgreSQL with TimescaleDB setup
└── README.md            # This documentation
```

## Prerequisites

### System Requirements
- Python 3.8+ (with pip)
- Node.js 16+ (with npm)
- Docker (optional, for PostgreSQL setup)
- **Recommended:** Virtual environment tool (Conda or Python venv)

### Software Dependencies

#### Backend
- Flask 2.0.1+
- Flask-CORS 3.0.10+
- psycopg2-binary 2.9.3+ (for PostgreSQL, optional)
- python-dotenv 0.19.0+
- **Built-in:** SQLite (included with Python - no separate installation needed)

#### Frontend
- React 18.2.0+
- React DOM 18.2.0+
- Recharts 2.10.3+
- Vite 5.0.8+

## Installation

### 1. Clone the Repository
```bash
git clone <repository-url>
cd weather_analysis
```

### 2. Set Up Backend Environment

#### Option 1: Conda Environment (Recommended for Data Science Users)
```bash
# Create conda environment
conda create -n weather-analysis python=3.10

# Activate environment
conda activate weather-analysis

# Install dependencies
cd backend
pip install -r requirements.txt
```

#### Option 2: Python Virtual Environment (Standard Python)
```bash
# Create virtual environment
python3 -m venv venv

# Activate environment (macOS/Linux)
source venv/bin/activate

# Activate environment (Windows)
venv\Scripts\activate

# Install dependencies
cd backend
pip install -r requirements.txt
```

### 3. Install Frontend Dependencies
```bash
cd ../frontend
npm install
```

**Note:** The frontend doesn't require a separate virtual environment - npm automatically creates and manages dependencies in the `node_modules` directory.

### 4. Initialize Database
```bash
cd ../backend
python -c "from database import init_database; init_database()"
python ingest_data.py  # Generate sample weather data
```

## Running the Services

### Start Backend API Server
```bash
cd backend
python app.py
```
**Output:**
```
* Running on http://127.0.0.1:5001
Press CTRL+C to quit
```

### Start Frontend Development Server
```bash
cd frontend
npm run dev
```
**Output:**
```
VITE v5.0.8  ready in 300 ms
➜  Local:   http://localhost:5173/
```

### Start PostgreSQL (Optional, for Production Only)
```bash
cd weather_analysis
docker compose up -d
```

**Note:** PostgreSQL is **not required** for normal operation. The project uses SQLite by default, which is already set up and ready to use.

## Stopping the Services

### Stop Backend Server
- Press `Ctrl+C` (or `Cmd+C` on Mac) in the backend terminal

### Stop Frontend Server
- Press `Ctrl+C` (or `Cmd+C` on Mac) in the frontend terminal

### Stop PostgreSQL (If Running)
```bash
cd weather_analysis
docker compose down
```

**Note:** Only run this if you started PostgreSQL for production use. The default SQLite database doesn't require any stopping.

## Verification

### Verify Backend API
```bash
curl http://localhost:5001/api/weather/yearly
```
**Expected:** JSON response with weather trend data

### Verify Frontend
1. Open `http://localhost:5173` in your browser
2. You should see the "Beijing Weather Analysis" dashboard with interactive charts

### Verify Database Connection
```bash
# Check if SQLite database exists
ls -la backend/weather.db

# Check if PostgreSQL is running
docker ps | grep weather_analysis_db
```

### Check Running Ports
```bash
# Backend port
lsof -i :5001

# Frontend port
lsof -i :5173

# PostgreSQL port (if running)
lsof -i :5432
```

## API Endpoints

### Backend API Base URL: `http://localhost:5001/api`

| Endpoint               | Method | Description                         |
|------------------------|--------|-------------------------------------|
| `/weather/yearly`      | GET    | 50-year annual weather trends       |
| `/weather/seasonal`    | GET    | Seasonal weather patterns           |
| `/weather/daily`       | GET    | Daily weather data (last year)      |
| `/weather/compare`     | GET    | Year-to-year weather comparisons    |

### Example Request: Compare Years
```bash
curl "http://localhost:5001/api/weather/compare?period=yearly&year1=2020&year2=2021"
```

## Data Storage

### Default Database: SQLite
- **Location:** `backend/weather.db`
- **Status:** **Currently active** - This is what the project uses by default
- **Features:** Local file-based storage, optimized with indexes and materialized views
- **Capacity:** Contains 18,628 weather records (1976-2026)
- **Use Case:** Development, testing, and small-scale deployments

### Optional Database: PostgreSQL with TimescaleDB
- **Configuration:** `docker-compose.yml`
- **Status:** **Not running by default** - Optional for production
- **Features:** Time-series optimization, production scalability, better performance with large datasets
- **Connection:** `postgresql://weather_user:weather_password@localhost:5432/weather_db`
- **Use Case:** Production environments with larger datasets or higher traffic

### Database Configuration

#### Using SQLite (Default)
- No additional setup required - SQLite is built into Python
- Database file is automatically created during initialization

#### Using PostgreSQL (Optional)
1. Start PostgreSQL:
   ```bash
   docker compose up -d
   ```

2. Update backend configuration:
   ```python
   # In backend/app.py
   DB_TYPE = 'postgres'  # Change from 'sqlite'
   ```

3. Restart backend:
   ```bash
   cd backend && python app.py
   ```

## Deployment

### Development Environment
1. Follow the [Installation](#installation) steps
2. Run both backend and frontend services as described in [Running the Services](#running-the-services)

### Production Environment

#### Option 1: SQLite (Lightweight)
1. Deploy backend with Gunicorn:
   ```bash
   cd backend
   gunicorn -w 4 -b 0.0.0.0:5001 app:app
   ```

2. Build and deploy frontend:
   ```bash
   cd frontend
   npm run build
   # Serve the `dist` directory with a web server like Nginx
   ```

#### Option 2: PostgreSQL with TimescaleDB (Scalable)
1. Start PostgreSQL:
   ```bash
   docker compose up -d
   ```

2. Update backend configuration to use PostgreSQL:
   ```python
   # In backend/app.py
   DB_TYPE = 'postgres'
   ```

3. Deploy backend and frontend as described above

### Environment Variables

#### Backend
- `DB_TYPE`: Database type (`sqlite` or `postgres`)
- `FLASK_ENV`: Environment (`development` or `production`)

#### Frontend
- `VITE_API_URL`: Backend API URL (default: `/api`)

## Troubleshooting

### Common Issues

1. **Port Conflicts**
   - If port 5001 is occupied: Modify `backend/app.py` to use a different port
   - If port 5173 is occupied: Run `npm run dev -- --port 5174`

2. **npm Permission Errors**
   ```bash
   sudo npm cache clean --force
   sudo chown -R $USER ~/.npm
   ```

3. **Database Connection Errors**
   - Verify SQLite database exists: `backend/weather.db`
   - Check PostgreSQL container status: `docker compose ps`

4. **API Endpoint Errors**
   - Verify backend is running: `curl http://localhost:5001/api/weather/yearly`
   - Check backend logs for error messages

## License

MIT License - see the [LICENSE](LICENSE) file for details.
