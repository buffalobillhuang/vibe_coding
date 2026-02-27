# World Museums Backend

FastAPI backend service for fetching museum information from Baidu Baike, Bilibili, and Xiaohongshu.

## Setup

```bash
# Create virtual environment
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Copy .env.example to .env and configure if needed
cp .env.example .env

# Run development server with auto-reload
uvicorn app.main:app --reload --host 0.0.0.0 --port 8000

# Production mode
uvicorn app.main:app --host 0.0.0.0 --port 8000 --workers 4
```

## API Endpoints

### `GET /api/museums/{city}`
Fetch museums for a specific city.

**Response:**
```json
{
  "city": "Beijing",
  "museums": [
    {
      "id": "national_museum_china",
      "name": "National Museum of China",
      "location": "Dongcheng District, Beijing",
      "introduction": "...",
      "images": ["url1", "url2"],
      "videos": [{"title": "...", "embedUrl": "..."}],
      "sources": {"baike": "...", "bilibili": "..."}
    }
  ]
}
```

### `GET /health`
Health check endpoint.

## Scraping Services

The backend includes scrapers for:

1. **Baidu Baike** (`app/services/baidu_baike.py`)
   - Extracts museum introductions
   - Uses BeautifulSoup4 for HTML parsing

2. **Bilibili** (`app/services/bilibili.py`)
   - Searches for museum-related videos
   - Generates embed URLs from video IDs

3. **Xiaohongshu** (`app/services/xiaohongshu.py`)
   - Currently uses fallback images (requires authentication)

## Caching

Results are cached with 24-hour TTL to reduce API calls:
- Development: In-memory LRU cache
- Production: Redis (optional, configure in `.env`)

## Testing

```bash
# Run all tests
pytest tests/ -v

# Run specific test file
pytest tests/test_museums.py -v
```

## Environment Variables

See `.env.example` for available configuration options.
