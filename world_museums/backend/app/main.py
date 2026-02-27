from fastapi import FastAPI, HTTPException, status
from fastapi.middleware.cors import CORSMiddleware
import logging

from app.routes.museums import router as museums_router
from app.utils.config import config

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="World Museums API",
    description="API for fetching museum information from Baidu Baike, Bilibili, and Xiaohongshu",
    version="0.1.0"
)

# CORS middleware for frontend communication
app.add_middleware(
    CORSMiddleware,
    allow_origins=config.CORS_ORIGINS,
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(museums_router, prefix="/api", tags=["museums"])


@app.get("/", tags=["health"])
async def root():
    """Health check endpoint."""
    return {"status": "healthy", "service": "world-museums-api"}


@app.get("/health", tags=["health"])
async def health_check():
    """Detailed health check."""
    return {
        "status": "ok",
        "version": "0.1.0"
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "app.main:app",
        host=config.API_HOST,
        port=config.API_PORT,
        reload=True
    )
