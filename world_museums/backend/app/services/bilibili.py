import logging
import requests
from typing import List, Optional, Dict
from urllib.parse import quote

logger = logging.getLogger(__name__)


class BilibiliScraper:
    """Scraper for museum videos from Bilibili."""

    SEARCH_API = "https://api.bilibili.com/x/web-interface/search/type"
    BASE_URL = "https://www.bilibili.com"
    USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

    def __init__(self):
        self.session = requests.Session()
        self.session.headers.update({
            "User-Agent": self.USER_AGENT,
            "Accept": "application/json, text/plain, */*",
            "Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
        })

    def search_videos(self, museum_name: str, limit: int = 3) -> List[Dict]:
        """Search for museum-related videos on Bilibili."""
        try:
            query = f"{museum_name} 博物馆"
            params = {
                "search_type": "video",
                "keyword": quote(query),
                "page": 1,
                "pagesize": limit,
            }

            logger.info(f"Searching Bilibili for: {query}")
            response = self.session.get(self.SEARCH_API, params=params, timeout=10)
            
            if response.status_code != 200:
                logger.warning(f"Bilibili API error: HTTP {response.status_code}")
                return []

            data = response.json()
            if data.get("code") != 0:
                logger.warning(f"Bilibili API returned error code: {data.get('code')}")
                return []

            results = data.get("data", {}).get("result", [])
            videos = []

            for result in results[:limit]:
                video_info = self._parse_video_result(result)
                if video_info and not any(v["title"] == video_info["title"] for v in videos):
                    videos.append(video_info)
                    if len(videos) >= limit:
                        break

            logger.info(f"Found {len(videos)} videos for {museum_name}")
            return videos

        except requests.RequestException as e:
            logger.error(f"Request error searching Bilibili for {museum_name}: {e}")
            return []
        except Exception as e:
            logger.exception(f"Unexpected error searching Bilibili for {museum_name}: {e}")
            return []

    def _parse_video_result(self, result: Dict) -> Optional[Dict]:
        """Parse Bilibili API search result into video dict."""
        try:
            title = result.get("title", "")
            cid = result.get("cid")
            play_url = result.get("play")

            if not cid or not title:
                return None

            embed_url = f"https://player.bilibili.com/player.html?cid={cid}&page=1&high_quality=1&danmaku=0"

            return {
                "title": title,
                "embedUrl": embed_url,
                "videoId": result.get("bvid", ""),
            }

        except Exception as e:
            logger.error(f"Error parsing Bilibili video result: {e}")
            return None


# Singleton instance
bilibili_scraper = BilibiliScraper()
