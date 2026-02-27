import logging
from typing import List, Optional, Dict, Any
from fastapi import APIRouter, HTTPException, status

from app.models.museum import MuseumResponse, Museum, Video, Sources
from app.utils.cache import cache_service
from app.services.baidu_baike import baike_scraper
from app.services.bilibili import bilibili_scraper
from app.services.xiaohongshu import xiaohongshu_scraper

logger = logging.getLogger(__name__)

router = APIRouter()


# Mock data for development (fallback when scraping unavailable)
MOCK_MUSEUMS: Dict[str, List[Dict[str, Any]]] = {
    "beijing": [
        {
            "id": "national_museum_china",
            "name": "National Museum of China",
            "location": "Dongcheng District, Beijing",
            "introduction": "The National Museum of China is located on the eastern side of Tiananmen Square in Beijing. It is one of the largest museums in the world, housing over 1.4 million artifacts spanning Chinese history and culture.",
            "images": [
                "https://example.com/nmc1.jpg",
                "https://example.com/nmc2.jpg",
                "https://example.com/nmc3.jpg",
            ],
            "videos": [],
            "sources": {
                "baike": "https://baike.baidu.com/item/National Museum of China"
            },
        },
        {
            "id": "palace_museum",
            "name": "Palace Museum (Forbidden City)",
            "location": "Dongcheng District, Beijing",
            "introduction": "The Palace Museum is located in the center of Beijing within the former imperial palace complex. Built during the Ming Dynasty, it contains over 1.8 million artifacts and represents the pinnacle of Chinese palace architecture.",
            "images": [
                "https://example.com/palace1.jpg",
                "https://example.com/palace2.jpg",
            ],
            "videos": [],
            "sources": {"baike": "https://baike.baidu.com/item/Palace Museum"},
        },
    ],
    "shanghai": [
        {
            "id": "shanghai_museum",
            "name": "Shanghai Museum",
            "location": "People's Square, Huangpu District, Shanghai",
            "introduction": "The Shanghai Museum is a renowned museum of ancient Chinese art, located on People's Square. It houses over 120,000 artifacts including bronze ware, ceramics, calligraphy, painting, and jade.",
            "images": ["https://example.com/shm1.jpg", "https://example.com/shm2.jpg"],
            "videos": [],
            "sources": {"baike": "https://baike.baidu.com/item/Shanghai Museum"},
        }
    ],
    "xi_an": [
        {
            "id": "shaanxi_history_museum",
            "name": "Shaanxi History Museum",
            "location": "Xi'an, Shaanxi Province",
            "introduction": "The Shaanxi History Museum is one of the most important museums in China, located near the Giant Wild Goose Pagoda in Xi'an. It contains over 1.7 million artifacts from the Qin and Han dynasties.",
            "images": ["https://example.com/shm_xa1.jpg"],
            "videos": [],
            "sources": {"baike": "https://baike.baidu.com/item/Shaanxi History Museum"},
        }
    ],
}


def normalize_city_name(city: str) -> str:
    """Normalize city name for lookup."""
    return city.lower().strip().replace(" ", "_").replace("-", "_")


@router.get("/museums/{city}", response_model=MuseumResponse)
async def get_museums_for_city(city: str):
    """
    Get museums for a specific city.

    Attempts to fetch real data from Baidu Baike and Bilibili,
    falls back to mock data if scraping fails or is unavailable.
    """
    try:
        normalized_city = normalize_city_name(city)

        # Check cache first
        cache_key = f"museums:{normalized_city}"
        cached_data = cache_service.get(cache_key)
        if cached_data:
            logger.info(f"Cache hit for {city}")
            return MuseumResponse(**cached_data)

        # Try mock data first (development mode)
        museums = MOCK_MUSEUMS.get(normalized_city, [])

        # If no mock data, try to scrape real data
        if not museums:
            logger.warning(f"No mock data for {city}, would attempt scraping here")
            museums = await _scrape_museums(city)

        if not museums:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"No museums found for city: {city}",
            )

        # Cache the result (mock data is already dict format)
        cache_service.set(cache_key, {"city": city, "museums": museums})

        return MuseumResponse(city=city, museums=museums)

    except HTTPException:
        raise
    except Exception as e:
        logger.exception(f"Error fetching museums for {city}: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Failed to fetch museum data: {str(e)}",
        )


async def _scrape_museums(city_name: str) -> List[Dict[str, Any]]:
    """Scrape real museum data from external sources."""
    museums = []

    # This would iterate through known museums in the city
    # and scrape their data from Baidu Baike, Bilibili, etc.
    # For now, returns empty list - mock data is used instead

    return museums
