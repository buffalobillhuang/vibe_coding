from pydantic import BaseModel, Field
from typing import Optional, List


class Video(BaseModel):
    title: str = Field(..., description="Video title")
    embed_url: str = Field(..., description="Bilibili embed URL")


class Sources(BaseModel):
    baike: str
    bilibili: Optional[str] = None


class Museum(BaseModel):
    id: str
    name: str
    location: str
    introduction: str
    images: List[str] = Field(default_factory=list)
    videos: List[Video] = Field(default_factory=list)
    sources: Sources


class MuseumResponse(BaseModel):
    city: str
    museums: List[Museum]


class SearchResponse(BaseModel):
    results: Optional[List[Museum]] = None
    error: Optional[str] = None
