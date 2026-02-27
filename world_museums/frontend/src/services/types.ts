export interface Video {
  title: string;
  embedUrl: string;
}

export interface Sources {
  baike: string;
  bilibili?: string;
}

export interface Museum {
  id: string;
  name: string;
  location: string;
  introduction: string;
  images: string[];
  videos: Video[];
  sources: Sources;
}

export interface MuseumResponse {
  city: string;
  museums: Museum[];
}

export type MuseumList = Museum[];
