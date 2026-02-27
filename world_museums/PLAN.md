# China Museums Web App - Implementation Plan

## Project Overview
A desktop-first web application displaying an interactive China map (left 80%) with clickable cities, showing museum information in the right panel (20%). Data sourced from Baidu Baike, Bilibili, and Xiaohongshu.

---

## Architecture

```
world_museums/
├── frontend/              # Next.js → Vercel deployment
│   ├── src/
│   │   ├── components/
│   │   │   ├── ChinaMap.tsx        # Leaflet map with GeoJSON layers
│   │   │   ├── MuseumPanel.tsx     # Right panel container
│   │   │   ├── MuseumCard.tsx      # Individual museum card
│   │   │   ├── ImageGallery.tsx    # 3-5 image grid/carousel
│   │   │   └── VideoEmbed.tsx      # Bilibili video player
│   │   ├── pages/
│   │   │   └── index.tsx           # Main split-layout page
│   │   ├── services/
│   │   │   ├── api.ts              # API client to backend
│   │   │   └── types.ts            # TypeScript interfaces
│   │   └── styles/globals.css      # Tailwind + custom
│   ├── public/china.geojson        # China province boundaries
│   ├── tailwind.config.js
│   └── next.config.js
│
├── backend/               # FastAPI → Vercel serverless or Railway
│   ├── app/
│   │   ├── main.py                 # FastAPI app + router registration
│   │   ├── routes/
│   │   │   └── museums.py          # /api/museums/{city} endpoint
│   │   ├── services/
│   │   │   ├── baidu_baike.py      # Scraper for museum intros
│   │   │   ├── bilibili.py         # Video search & embed URLs
│   │   │   └── xiaohongshu.py      # Image scraper (fallback)
│   │   ├── models/
│   │   │   └── museum.py           # Pydantic data schemas
│   │   ├── utils/
│   │   │   ├── cache.py            # LRU/Redis caching
│   │   │   └── config.py           # API keys, rate limits
│   │   └── data/mock_museums.json  # Fallback mock data
│   ├── requirements.txt
│   └── README.md
│
├── AGENTS.md              # Agentic coding guidelines (to create)
├── PLAN.md                # This document
└── .gitignore
```

---

## Tech Stack & Dependencies

### Frontend (Next.js 14 + TypeScript)
| Package | Purpose |
|---------|---------|
| next@14.0.0 | React framework with API routes option |
| react@18.2.0 | UI library |
| leaflet@1.9.4 | Map library |
| react-leaflet@4.2.1 | React wrappers for Leaflet |
| tailwindcss@3.3.5 | Utility-first CSS |
| axios@1.6.0 | HTTP client to backend API |

### Backend (FastAPI + Python 3.11)
| Package | Purpose |
|---------|---------|
| fastapi@0.109.0 | Async web framework |
| uvicorn[standard]@0.27.0 | ASGI server |
| requests@2.31.0 | HTTP client for scraping |
| beautifulsoup4@4.12.3 | HTML parsing |
| playwright@1.40.0 | JS-heavy site scraping |
| pydantic@2.5.3 | Data validation/models |
| aiohttp@3.9.1 | Async HTTP requests |

---

## API Endpoints

### `GET /api/museums/{city}`
**Purpose**: Fetch all museums for a selected city  
**Response**:
```json
{
  "city": "Beijing",
  "museums": [
    {
      "id": "national_museum_china",
      "name": "National Museum of China",
      "location": "Dongcheng District, Beijing",
      "introduction": "The National Museum of China...",
      "images": [
        "https://example.com/img1.jpg",
        "https://example.com/img2.jpg",
        "https://example.com/img3.jpg"
      ],
      "videos": [
        {"title": "Tour Guide Video", "embedUrl": "https://player.bilibili.com/player.html?cid=12345"}
      ],
      "sources": {
        "baike": "https://baike.baidu.com/item/National Museum of China",
        "bilibili": "https://www.bilibili.com/video/abc123"
      }
    }
  ]
}
```

---

## Map Implementation Details

### GeoJSON Source
- Use simplified China province-level boundaries from: https://github.com/deldersveld/topojson/tree/master/countries/china
- Convert TopoJSON to GeoJSON if needed
- File size target: <100KB for fast loading

### Leaflet Configuration
```typescript
// Map settings
- Center: [35.8617, 104.1954] (center of China)
- Zoom: 4 (nation-level view)
- Min zoom: 2, Max zoom: 8
- Attribution: OpenStreetMap contributors

// Interactivity
- Click on province → show city markers or fetch cities API
- Click on city marker → trigger museum panel update
- Hover tooltip: Show city name
```

---

## Python Scraping Strategy

### Phase 1: Mock Data First (Week 1)
Start with manually curated JSON for major cities to validate UI/UX without scraping complexities.

**Mock data structure** (`backend/data/mock_museums.json`):
```json
{
  "beijing": [
    {
      "name": "National Museum of China",
      "introduction": "...",
      "images": ["mock_url_1.jpg", ...],
      "videos": [...]
    }
  ],
  "shanghai": [...],
  "xi_an": [...],
  "nanjing": [...],
  "chengdu": [...]
}
```

### Phase 2: Baidu Baike Scraper (Week 2)
**File**: `backend/app/services/baidu_baike.py`

```python
# Strategy
1. Construct URL: f"https://baike.baidu.com/item/{museum_name}"
2. Use requests with User-Agent header to mimic browser
3. Parse HTML with BeautifulSoup4
4. Extract main intro paragraph (typically in .lemmaWrd-content or similar)
5. Return clean text (strip HTML tags)

# Caching
- Cache results for 24 hours using lru_cache or Redis
- Handle rate limiting: max 10 requests/minute per IP
```

### Phase 3: Bilibili Video Scraper (Week 2-3)
**File**: `backend/app/services/bilibili.py`

```python
# Strategy (prefer official API first)
1. Search query: f"{museum_name} 博物馆"
2. Use Bilibili API: https://api.bilibili.com/x/web-interface/search/type
   - Params: search_type=video, keyword=museum_name
3. Extract cid from top 3 results
4. Generate embed URLs: https://player.bilibili.com/player.html?cid={cid}

# Fallback if API unavailable
- Scrape search results page with BeautifulSoup
- Extract video IDs from HTML
```

### Phase 4: Xiaohongshu Image Scraper (Week 3) - **OPTIONAL**
**File**: `backend/app/services/xiaohongshu.py`

```python
# Challenges: Requires login, anti-bot measures
# Strategy options:
1. Check for public API documentation (unlikely)
2. Use official Xiaohongshu Open Platform (requires approval)
3. Fallback to alternative sources:
   - Unsplash (free images with "China museum" queries)
   - Official museum websites
   - Manual curation for key museums

# Recommendation: Start with mock/mock URLs, add real scraping if API becomes available
```

### Caching Layer (`backend/app/utils/cache.py`)
```python
from functools import lru_cache
import redis

# Development: In-memory LRU cache
@lru_cache(maxsize=1000)
def cached_scrape(source: str, query: str) -> Data:
    return scrape_data(source, query)

# Production: Redis for distributed caching
class CacheService:
    def __init__(self):
        self.redis = redis.Redis(host='localhost', port=6379, db=0)
    
    def get(self, key: str) -> Optional[Data]:
        data = self.redis.get(key)
        return json.loads(data) if data else None
    
    def set(self, key: str, value: Data, ttl: int = 86400):
        self.redis.setex(key, ttl, json.dumps(value))
```

---

## UI/UX Specifications (Desktop-First)

### Main Layout (`frontend/src/pages/index.tsx`)
```typescript
// Full viewport height, no scroll on main container
<div className="flex h-screen w-full bg-gray-100">
  
  {/* Left Panel: China Map - 80% width */}
  <div className="w-[80%] h-full relative shadow-lg">
    <div id="china-map" className="h-full w-full"></div>
    {loading && <LoadingSpinner />}
  </div>
  
  {/* Right Panel: Museum Details - 20% width */}
  <div className="w-[20%] h-full overflow-y-auto bg-white border-l shadow-xl">
    {!selectedCity ? (
      <EmptyState message="Click on a city in the map to view museums" />
    ) : (
      <MuseumPanel 
        city={selectedCity} 
        museums={museums} 
        loading={isLoading}
      />
    )}
  </div>
  
</div>
```

### MuseumCard Component (`frontend/src/components/MuseumCard.tsx`)
```typescript
// Structure per museum card
<div className="border-b pb-4 mb-4">
  {/* Header */}
  <h3 className="text-xl font-bold text-gray-900">{museum.name}</h3>
  <p className="text-sm text-gray-600 mt-1">{museum.location}</p>
  
  {/* Introduction */}
  <div className="mt-3">
    <h4 className="font-semibold text-gray-800">Introduction</h4>
    <p className="text-sm text-gray-700 mt-1 leading-relaxed">
      {museum.introduction.substring(0, 200)}...
    </p>
  </div>
  
  {/* Image Gallery (3-5 images) */}
  <div className="mt-4 grid grid-cols-3 gap-2">
    {museum.images.map((url, idx) => (
      <img key={idx} src={url} alt={`${museum.name} ${idx+1}`} 
           className="w-full h-24 object-cover rounded-lg hover:scale-105 transition" />
    ))}
  </div>
  
  {/* Videos (1-3 embeds) */}
  <div className="mt-4">
    <h4 className="font-semibold text-gray-800">Videos</h4>
    {museum.videos.map((video, idx) => (
      <iframe key={idx} src={video.embedUrl} 
              className="w-full h-48 mt-2 rounded-lg" />
    ))}
  </div>
  
  {/* Sources */}
  <div className="mt-3 text-xs text-gray-500">
    Source: <a href={museum.sources.baike}>Baidu Baike</a>, 
    <a href={museum.sources.bilibili}>Bilibili</a>
  </div>
</div>
```

### Tailwind Color Scheme
- Primary: `blue-600` for interactive elements
- Backgrounds: `gray-50` (map panel), `white` (details panel)
- Text: `gray-900` (headings), `gray-700` (body), `gray-500` (sources)
- Borders: `gray-200` for subtle separation

---

## Development Timeline

### Week 1: Foundation & Mock Data
**Days 1-2**: Project Setup
- [ ] Initialize Next.js frontend with TypeScript + Tailwind
- [ ] Initialize FastAPI backend structure
- [ ] Create AGENTS.md file
- [ ] Set up .gitignore and basic configs
- [ ] Test both run locally

**Days 3-5**: Map Foundation
- [ ] Download China GeoJSON (simplified province-level)
- [ ] Implement Leaflet map in Next.js
- [ ] Add click handlers to provinces/cities
- [ ] Create mock API endpoint returning dummy data

### Week 2: UI Components & Backend API
**Days 6-8**: Museum Panel UI
- [ ] Build MuseumCard component structure
- [ ] Create ImageGallery component (3-column grid)
- [ ] Implement VideoEmbed with Bilibili player
- [ ] Connect to mock backend data

**Days 9-10**: Backend API + Mock Data
- [ ] Set up FastAPI with CORS for frontend domain
- [ ] Create /api/museums/{city} endpoint
- [ ] Build comprehensive mock_museums.json (5 cities, 3 museums each)
- [ ] End-to-end testing: click city → see museum data

### Week 3: Python Scrapers Implementation
**Days 11-12**: Baidu Baike Scraper
- [ ] Implement basic scraper with requests + BeautifulSoup
- [ ] Test on 5 museums, extract clean intro text
- [ ] Add error handling and fallback to mock data
- [ ] Integrate LRU caching

**Days 13-14**: Bilibili Video Scraper
- [ ] Use official Bilibili search API
- [ ] Extract cid and generate embed URLs
- [ ] Handle cases with no videos (show placeholder)
- [ ] Add caching layer

### Week 4: Integration & Polish
**Days 15-16**: Full Integration
- [ ] Connect scrapers to API endpoints
- [ ] Implement loading states in UI
- [ ] Add error boundary components
- [ ] Optimize image lazy-loading

**Days 17-18**: Testing & Optimization
- [ ] Test all major cities (Beijing, Shanghai, Xi'an, Nanjing, Chengdu)
- [ ] Check performance: GeoJSON load time <2s
- [ ] Verify responsive design on different desktop resolutions
- [ ] Mobile view (basic, not priority)

**Days 19-20**: Deployment Preparation
- [ ] Deploy frontend to Vercel (connect GitHub repo)
- [ ] Deploy backend to Vercel serverless or Railway
- [ ] Configure environment variables (API keys if needed)
- [ ] Final production testing

---

## Success Criteria (MVP Definition)

### Must-Have Features:
- ✅ Desktop browser (Chrome/Firefox/Safari) shows interactive China map
- ✅ Clicking a province/city displays museums in right panel
- ✅ Each museum card contains:
  - Museum name and location
  - Introduction text (from Baidu Baike or mock data)
  - 3-5 images (scraped or fallback URLs)
  - 1-3 Bilibili video embeds
- ✅ Backend API serves data without errors
- ✅ Caching prevents excessive scraping on repeated visits
- ✅ Both frontend and backend deployable to Vercel

### Nice-to-Have (Post-MVP):
- [ ] Search bar to find specific museums/cities
- [ ] Province-level zoom interaction
- [ ] Mobile-responsive layout optimization
- [ ] Xiaohongshu image integration (if API becomes available)
- [ ] User favorites/bookmarks feature

---

## Technical Risks & Mitigations

| Risk | Impact | Probability | Mitigation Strategy |
|------|--------|-------------|---------------------|
| **Xiaohongshu scraping blocked** | Medium | High | Start with mock URLs; use Unsplash/official sites as fallback |
| **Baidu Baike anti-bot measures** | Medium | Medium | Aggressive caching (24h); rate limiting; consider manual curation for top museums |
| **Vercel serverless timeout** | Low | Low | Keep functions <1min; pre-generate popular city data; use background tasks |
| **CORS issues between frontend/backend** | Medium | Medium | Deploy backend to Vercel in same domain; or configure FastAPI CORS properly |
| **GeoJSON file too large** | Low | Low | Use simplified topology; load provinces on-demand only |

---

## Environment Variables (Required)

### Frontend (`frontend/.env.local`)
```bash
NEXT_PUBLIC_API_URL=https://your-backend.vercel.app/api
```

### Backend (`backend/.env`)
```bash
# Optional: API keys if using official APIs
BILIBILI_API_KEY=  # If required for search API
REDIS_URL=         # For production caching (optional)

# Scraping settings
MAX_REQUESTS_PER_MINUTE=10
CACHE_TTL_HOURS=24
```

---

## Resources & References

### Map Data Sources:
- China GeoJSON: https://github.com/deldersveld/topojson/tree/master/countries/china
- Leaflet docs: https://leafletjs.com/reference.html
- React-Leaflet: https://react-leaflet.js.org/

### API Documentation:
- Bilibili Open Platform: https://open.bilibili.com/
- FastAPI docs: https://fastapi.tiangolo.com/
- Next.js API routes: https://nextjs.org/docs/pages/building-your-application/routing/api-routes

### Scraping Best Practices:
- Respect robots.txt for all target sites
- Add delays between requests (1-2 seconds)
- Use proper User-Agent headers
- Implement exponential backoff on errors

---

## Next Steps

This plan covers the complete implementation from setup to deployment. The approach prioritizes:
1. **Mock data first** → Validate UI/UX quickly without scraping complexities
2. **Iterative scraper integration** → Add real data once foundation is solid
3. **Desktop-first** → Focus on optimal desktop experience as requested

**Implementation begins with**:
1. Create `AGENTS.md` file for agentic coding guidelines
2. Scaffold frontend (Next.js + Tailwind + Leaflet)
3. Scaffold backend (FastAPI + scraper templates)
4. Add mock data structure for testing

---

*Document created: February 26, 2026*
*Version: 1.0*
