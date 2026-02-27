import json
import logging
from functools import lru_cache
from typing import Optional, Any, Dict

logger = logging.getLogger(__name__)


class CacheService:
    """Caching service for scraper results with TTL support."""

    def __init__(self, use_redis: bool = False, redis_url: str = None):
        self.use_redis = use_redis
        if use_redis and redis_url:
            try:
                import redis
                self.redis_client = redis.from_url(redis_url)
                logger.info("Redis cache client initialized")
            except Exception as e:
                logger.warning(f"Failed to initialize Redis: {e}, falling back to LRU cache")
                self.use_redis = False
        else:
            # Use in-memory LRU cache (24hr TTL equivalent)
            self._cache: Dict[str, Any] = {}

    def get(self, key: str) -> Optional[Dict]:
        """Get cached data by key."""
        try:
            if self.use_redis and hasattr(self, 'redis_client'):
                data = self.redis_client.get(key)
                return json.loads(data) if data else None
            
            # In-memory cache (LRU via dict ordering + manual TTL check)
            import time
            if key in self._cache:
                value, timestamp = self._cache[key]
                if time.time() - timestamp < 86400:  # 24 hour TTL
                    return value
                else:
                    del self._cache[key]
            return None
        except Exception as e:
            logger.error(f"Cache get error for key {key}: {e}")
            return None

    def set(self, key: str, value: Dict, ttl: int = 86400) -> bool:
        """Set cached data with TTL (default 24 hours)."""
        try:
            if self.use_redis and hasattr(self, 'redis_client'):
                self.redis_client.setex(key, ttl, json.dumps(value))
                return True
            
            # In-memory cache
            import time
            self._cache[key] = (value, time.time())
            return True
        except Exception as e:
            logger.error(f"Cache set error for key {key}: {e}")
            return False

    def delete(self, key: str) -> bool:
        """Delete cached data by key."""
        try:
            if self.use_redis and hasattr(self, 'redis_client'):
                self.redis_client.delete(key)
                return True
            
            # In-memory cache
            if key in self._cache:
                del self._cache[key]
                return True
            return False
        except Exception as e:
            logger.error(f"Cache delete error for key {key}: {e}")
            return False


# Global cache instance (can be replaced with Redis in production)
cache_service = CacheService(use_redis=False)


def cached_scrape(source: str, query: str):
    """Decorator for caching scraper results."""

    def decorator(func):
        @lru_cache(maxsize=1000)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper

    return decorator
