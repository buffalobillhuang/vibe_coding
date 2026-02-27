import os
from dotenv import load_dotenv

load_dotenv()


class Config:
    """Application configuration from environment variables."""

    # API settings
    API_HOST = os.getenv("API_HOST", "0.0.0.0")
    API_PORT = int(os.getenv("API_PORT", 8000))

    # Scraping settings
    MAX_REQUESTS_PER_MINUTE = int(os.getenv("MAX_REQUESTS_PER_MINUTE", 10))
    CACHE_TTL_HOURS = int(os.getenv("CACHE_TTL_HOURS", 24))

    # Caching configuration
    USE_REDIS = os.getenv("USE_REDIS", "false").lower() == "true"
    REDIS_URL = os.getenv("REDIS_URL")

    # Bilibili API settings (if using official API)
    BILIBILI_API_KEY = os.getenv("BILIBILI_API_KEY", "")

    # CORS origins for frontend
    CORS_ORIGINS = [
        origin.strip()
        for origin in os.getenv("CORS_ORIGINS", "http://localhost:3000").split(",")
    ]


config = Config()
