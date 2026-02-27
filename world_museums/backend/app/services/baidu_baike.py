import logging
import requests
from bs4 import BeautifulSoup
from typing import Optional, Dict

logger = logging.getLogger(__name__)


class BaiduBaikeScraper:
    """Scraper for museum introductions from Baidu Baike."""

    BASE_URL = "https://baike.baidu.com/item"
    USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

    def __init__(self):
        self.session = requests.Session()
        self.session.headers.update({
            "User-Agent": self.USER_AGENT,
            "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
            "Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
        })

    def scrape_intro(self, museum_name: str) -> Optional[str]:
        """Scrape museum introduction from Baidu Baike."""
        try:
            url = f"{self.BASE_URL}/{museum_name}"
            logger.info(f"Fetching {url}")

            response = self.session.get(url, timeout=10)
            if response.status_code != 200:
                logger.warning(f"Failed to fetch {url}: HTTP {response.status_code}")
                return None

            soup = BeautifulSoup(response.text, 'html.parser')
            
            # Try different selectors for intro text
            intro_selectors = [
                ".lemmaWrd-content",
                ".basic-info paras",
                "[data-ke-src$='intro']",
                ".para"
            ]

            for selector in intro_selectors:
                element = soup.select_one(selector)
                if element:
                    intro_text = self._clean_text(element.get_text())
                    if len(intro_text) > 50:  # Validate content length
                        logger.info(f"Successfully scraped intro for {museum_name}")
                        return intro_text[:2000]  # Limit to 2000 chars

            logger.warning(f"No introduction found for {museum_name}")
            return None

        except requests.RequestException as e:
            logger.error(f"Request error scraping {museum_name}: {e}")
            return None
        except Exception as e:
            logger.exception(f"Unexpected error scraping {museum_name}: {e}")
            return None

    def _clean_text(self, text: str) -> str:
        """Clean and normalize scraped text."""
        # Remove extra whitespace and newlines
        import re
        text = re.sub(r'\s+', ' ', text)
        
        # Remove common UI elements
        remove_patterns = [
            r'分享到：.*$',
            r'参考资料.*$',
            r'查看全部.*$',
            r'^编辑.*$',
        ]
        
        for pattern in remove_patterns:
            text = re.sub(pattern, '', text, flags=re.MULTILINE)
        
        return text.strip()


# Singleton instance
baike_scraper = BaiduBaikeScraper()
