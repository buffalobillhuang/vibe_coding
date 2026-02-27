# World Museums - China Map Web Application

A desktop-first web application displaying an interactive China map with museum information sourced from Baidu Baike, Bilibili, and Xiaohongshu.

## Architecture

```
world_museums/
├── frontend/              # Next.js → Vercel deployment
│   ├── src/
│   │   ├── components/    # React components (ChinaMap, MuseumCard)
│   │   ├── services/      # API clients and TypeScript types
│   │   └── pages/         # Next.js page routes
│   └── public/            # Static assets (GeoJSON)
│
├── backend/               # FastAPI → Vercel/Railway deployment
│   ├── app/
│   │   ├── routes/        # API endpoints (/api/museums/{city})
│   │   ├── services/      # Scrapers for Baidu Baike, Bilibili, Xiaohongshu
│   │   ├── models/        # Pydantic data schemas
│   │   └── utils/         # Caching and configuration utilities
│   └── data/              # Mock museum data
```

## Quick Start

### Prerequisites
- Node.js 18+ for frontend
- Python 3.11+ for backend

### Frontend Setup

```bash
cd frontend
npm install
npm run dev
# Opens at http://localhost:3000
```

### Backend Setup

```bash
cd backend
python -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate
pip install -r requirements.txt
uvicorn app.main:app --reload
# Opens at http://localhost:8000
```

## Features

- **Interactive China Map**: Click on cities to view museums (Leaflet.js)
- **Museum Details Panel**: Shows introduction, images, and videos for each museum
- **Multi-source Data**: Aggregates from Baidu Baike (intros), Bilibili (videos), Xiaohongshu (images)
- **Caching Layer**: Reduces API calls with 24-hour TTL caching

## Tech Stack

| Component | Technology |
|-----------|------------|
| Frontend | Next.js 14, TypeScript, Tailwind CSS |
| Map | Leaflet + React-Leaflet |
| Backend | FastAPI (Python), Async I/O |
| Scraping | BeautifulSoup4, requests, aiohttp |
| Caching | LRU cache / Redis |

## Project Structure

See `AGENTS.md` for detailed code style guidelines and commands.  
See `PLAN.md` for implementation timeline and architecture details.

## API Documentation

Backend API docs available at: `http://localhost:8000/docs` (Swagger UI)

## Deployment

- **Frontend**: Deploy to Vercel (automatic from GitHub)
- **Backend**: Deploy to Vercel serverless or Railway/Render

## License

MIT
