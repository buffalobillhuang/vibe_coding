# World Museums - Agentic Coding Guidelines

## Project Overview
A desktop-first web application displaying an interactive China map with museum information sources from Baidu Baike, Bilibili, and Xiaohongshu.

---

## Build/Lint/Test Commands

### Frontend (Next.js)
```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Development mode with hot reload
npm run dev

# Production build
npm run build

# Start production server
npm start

# Linting
npm run lint

# Type checking
npx tsc --noEmit

# Run a single test file
npm test -- specific-test.test.tsx

# Run tests in watch mode (if tests exist)
npm test -- --watch

# Run tests for specific component
npm test -- MuseumCard.test.tsx
```

### Backend (FastAPI + Python)
```bash
# Navigate to backend directory
cd backend

# Create virtual environment
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Development mode with auto-reload
uvicorn app.main:app --reload --host 0.0.0.0 --port 8000

# Production mode
uvicorn app.main:app --host 0.0.0.0 --port 8000 --workers 4

# Run Python linter (flake8)
flake8 app/

# Type checking with mypy
mypy app/

# Run a single test file
pytest tests/test_museums.py -v

# Run all tests
pytest tests/ -v

# Test specific function
pytest tests/test_baidu_baike.py::test_scrape_intro -v
```

---

## Code Style Guidelines

### Frontend (TypeScript + React)

#### Imports
```typescript
// 1. Third-party libraries (sorted alphabetically)
import axios from 'axios';
import { useEffect, useState } from 'react';
import { MapContainer, TileLayer, GeoJSON } from 'react-leaflet';

// 2. Internal imports (relative paths)
import { MuseumCard } from '@/components/MuseumCard';
import { API_ENDPOINTS } from '@/services/api';

// 3. Styles last
import styles from './ChinaMap.module.css';

// Use absolute imports with @ alias configured in tsconfig.json
```

#### Formatting
- **Prettier**: 2-space indentation, single quotes, trailing commas
- **File structure**: Exports first, then imports within each component
- **Component size**: Keep components under 200 lines when possible
- **Line length**: Maximum 100 characters per line

#### Types & Interfaces
```typescript
// Use explicit types for all function parameters and return values
interface Museum {
  id: string;
  name: string;
  location: string;
  introduction: string;
  images: string[];
  videos: Video[];
  sources: Sources;
}

interface Video {
  title: string;
  embedUrl: string;
}

interface Sources {
  baike: string;
  bilibili?: string;
}

// Use type aliases for complex types
type MuseumList = Museum[];

// Prefer readonly for immutable data
const config: readonly string[] = ['value1', 'value2'];
```

#### Naming Conventions
- **Components**: PascalCase (`MuseumCard`, `ChinaMap`)
- **Functions/variables**: camelCase (`fetchMuseums`, `selectedCity`)
- **Constants**: UPPER_SNAKE_CASE (`API_ENDPOINTS`, `MAP_CENTER`)
- **Files**: Match component name (PascalCase) for components, lowercase with dashes for utilities

#### Error Handling
```typescript
// Frontend: Try-catch with user-friendly error messages
const fetchMuseums = async (city: string): Promise<void> => {
  try {
    const response = await axios.get(API_ENDPOINTS.museums(city));
    setMuseums(response.data.museums);
  } catch (error) {
    if (axios.isAxiosError(error)) {
      setError(`Failed to load museums: ${error.message}`);
    } else {
      setError('An unexpected error occurred');
    }
  }
};

// Show loading state during async operations
const [loading, setLoading] = useState(false);
setLoading(true);
// ... operation
setLoading(false);
```

#### React Best Practices
- Use functional components with hooks
- Memoize expensive computations with `useMemo` and `useCallback`
- Lazy load images with `loading="lazy"` attribute
- Implement proper cleanup in useEffect for subscriptions/timers

---

### Backend (Python + FastAPI)

#### Imports
```python
# 1. Standard library (alphabetically sorted)
import json
import logging
from functools import lru_cache
from typing import Optional

# 2. Third-party libraries (alphabetically sorted)
import aiohttp
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

# 3. Local imports (no leading dots for root, single dot for same package)
from app.models.museum import Museum
from app.utils.cache import cache_service

# Group with blank lines between sections
```

#### Formatting
- **Black**: Default formatting (line length = 88 characters)
- **Isort**: Alphabetical imports, grouped by standard lib/third-party/local
- **Type hints**: Required for all function parameters and return values

#### Type Hints & Pydantic Models
```python
from pydantic import BaseModel, Field
from typing import Optional, List

class Video(BaseModel):
    title: str = Field(..., description="Video title")
    embed_url: str = Field(..., description="Bilibili embed URL")

class Museum(BaseModel):
    id: str
    name: str
    location: str
    introduction: str
    images: List[str] = Field(default_factory=list)
    videos: List[Video] = Field(default_factory=list)
    sources: dict[str, str]

# Use Optional for nullable fields
class SearchResponse(BaseModel):
    results: Optional[List[Museum]] = None
    error: Optional[str] = None
```

#### Naming Conventions
- **Modules/filenames**: lowercase_with_underscores (`baidu_baike.py`, `cache_utils.py`)
- **Classes**: PascalCase (`CacheService`, `MuseumScraper`)
- **Functions/variables**: snake_case (`fetch_museum_intro`, `city_name`)
- **Constants**: UPPER_SNAKE_CASE (`MAX_REQUESTS_PER_MINUTE`, `CACHE_TTL_HOURS`)

#### Error Handling
```python
from fastapi import HTTPException, status
import logging

logger = logging.getLogger(__name__)

async def scrape_baidu_baike(museum_name: str) -> Optional[str]:
    """Scrape museum introduction from Baidu Baike."""
    try:
        url = f"https://baike.baidu.com/item/{museum_name}"
        async with aiohttp.ClientSession() as session:
            async with session.get(url, timeout=10) as response:
                if response.status != 200:
                    logger.warning(f"Failed to fetch {url}: {response.status}")
                    return None
                
                content = await response.text()
                # Parse and extract intro...
                return intro_text
    
    except aiohttp.ClientError as e:
        logger.error(f"AIOHTTP error scraping {museum_name}: {e}")
        return None
    except Exception as e:
        logger.exception(f"Unexpected error for {museum_name}: {e}")
        return None

# In API routes, raise HTTPException with appropriate status codes
@app.get("/api/museums/{city}")
async def get_museums(city: str):
    museums = await fetch_museums_for_city(city)
    if not museums:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"No museums found for city: {city}"
        )
    return {"city": city, "museums": museums}
```

#### Caching Strategy
```python
from functools import lru_cache
import redis

# Development: In-memory LRU cache (24hr TTL equivalent)
@lru_cache(maxsize=1000)
def cached_scrape_baidu(source: str, query: str) -> Optional[str]:
    """Scrape with automatic caching."""
    return scrape_data(source, query)

# Production: Redis for distributed caching
class CacheService:
    def __init__(self, redis_url: str):
        self.redis = redis.from_url(redis_url)
    
    async def get(self, key: str) -> Optional[dict]:
        data = await self.redis.get(key)
        return json.loads(data) if data else None
    
    async def set(self, key: str, value: dict, ttl: int = 86400):
        """Cache with 24-hour TTL by default."""
        await self.redis.setex(key, ttl, json.dumps(value))

# Use cache in scraper
@lru_cache(maxsize=500)
async def scrape_with_cache(museum_name: str) -> MuseumData:
    cached = await cache_service.get(f"museum:{museum_name}")
    if cached:
        return cached
    
    data = await scrape_museum(museum_name)
    await cache_service.set(f"museum:{museum_name}", data, ttl=86400)
    return data
```

#### Async Best Practices
- Use `async/await` for I/O operations (HTTP requests, database queries)
- Don't use blocking calls in async functions (`requests` → use `aiohttp`)
- Properly handle timeouts on all network requests
- Use connection pools for repeated HTTP calls

---

## General Guidelines

### Both Frontend & Backend
1. **No hardcoded secrets**: Always use environment variables
2. **Logging**: Add meaningful log messages at key points
3. **Documentation**: Include docstrings for functions and classes
4. **Testing**: Write tests for critical paths (unit + integration)
5. **Version control**: Commit frequently with descriptive commit messages

### Git Commit Style
```bash
# Format: type(scope): description

feat(frontend): add china map component
fix(backend): resolve caching issue in baidu scraper
docs: update API documentation
test(backend): add tests for museum endpoint
chore: update dependencies
```

### Performance Considerations
1. **Frontend**: Lazy load images, memoize expensive computations, debounce user input
2. **Backend**: Cache scraping results (24hr TTL), implement rate limiting, use async I/O
3. **Network**: Minimize API calls, batch requests where possible

### Security Practices
1. Never commit `.env` files or secrets
2. Validate all user inputs (frontend validation + backend sanitization)
3. Implement CORS properly between frontend and backend
4. Add rate limiting to prevent abuse of scraping endpoints

---

## File Structure Reference

```
world_museums/
├── frontend/
│   ├── src/
│   │   ├── components/        # Reusable UI components
│   │   │   ├── ChinaMap.tsx
│   │   │   ├── MuseumCard.tsx
│   │   │   └── ...
│   │   ├── pages/            # Next.js pages (App Router: app/)
│   │   │   └── index.tsx
│   │   ├── services/         # API clients, data fetching
│   │   │   ├── api.ts
│   │   │   └── types.ts
│   │   └── styles/           # Global styles
│   ├── public/               # Static assets (GeoJSON)
│   ├── tailwind.config.js
│   └── package.json
│
├── backend/
│   ├── app/
│   │   ├── routes/           # API route handlers
│   │   │   └── museums.py
│   │   ├── services/         # Business logic, scrapers
│   │   │   ├── baidu_baike.py
│   │   │   ├── bilibili.py
│   │   │   └── xiaohongshu.py
│   │   ├── models/           # Pydantic data models
│   │   │   └── museum.py
│   │   ├── utils/            # Utilities, helpers
│   │   │   ├── cache.py
│   │   │   └── config.py
│   │   └── main.py           # FastAPI app entry point
│   ├── data/                 # Mock data, fixtures
│   │   └── mock_museums.json
│   ├── tests/                # Unit and integration tests
│   ├── requirements.txt
│   └── README.md
│
├── AGENTS.md                 # This file
├── PLAN.md                   # Implementation plan
└── .gitignore
```

---

## Quick Reference

| Task | Frontend Command | Backend Command |
|------|-----------------|-----------------|
| Run dev server | `npm run dev` | `uvicorn app.main:app --reload` |
| Build for production | `npm run build` | N/A (Python) |
| Lint code | `npm run lint` | `flake8 app/` |
| Type check | `npx tsc --noEmit` | `mypy app/` |
| Run single test | `npm test -- file.tsx` | `pytest tests/file.py -v` |
| Run all tests | `npm test` | `pytest tests/ -v` |

---

*Last updated: February 26, 2026*
