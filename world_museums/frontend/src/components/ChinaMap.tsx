'use client';

import React, { useEffect, useRef } from 'react';

interface ChinaMapProps {
  onCitySelect: (city: string) => void;
}

const ChinaMap: React.FC<ChinaMapProps> = ({ onCitySelect }) => {
  const mapRef = useRef<HTMLDivElement>(null);
  const mapInstance = useRef<any | null>(null);

  useEffect(() => {
    if (!mapRef.current || mapInstance.current) return;

    setTimeout(() => {
import('leaflet').then(async (L: any) => {
        await import('leaflet/dist/leaflet.css');
        
        // Handle both default and direct export
        const LMap = L.default?.map || L.map;
        const LTileLayer = L.default?.tileLayer || L.tileLayer;

        if (!LMap) {
          console.error('Leaflet map function not found');
          return;
        }

        // Create map with explicit dimensions and enable dragging
        const map = LMap(mapRef.current, {
          center: [35.8617, 104.1954],
          zoom: 4,
          minZoom: 2,
          maxZoom: 8,
          dragEnable: true, // Enable dragging/panning
        });

        LTileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
          attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
          maxZoom: 8,
          minZoom: 2,
        }).addTo(map);

        mapInstance.current = map;

        // Force map to resize after it's added to DOM
        setTimeout(() => {
          map.invalidateSize();
        }, 100);

        const handleMapClick = (e: any) => {
  // For now, use a simple city mapping based on coordinates
  const lat = e.latlng.lat;
  const lng = e.latlng.lng;
  
  let cityName = 'Beijing';
  
  // Simple coordinate-based city detection for demo purposes
  if (lat > 30 && lat < 32 && lng > 120 && lng < 122) {
    cityName = 'Shanghai';
  } else if (lat > 34 && lat < 35 && lng > 108 && lng < 109) {
    cityName = 'Xi_an';
  } else if (lat > 32 && lat < 33 && lng > 118 && lng < 119) {
    cityName = 'Nanjing';
  } else if (lat > 30 && lat < 31 && lng > 103 && lng < 104) {
    cityName = 'Chengdu';
  }
  
  console.log('City detected:', cityName, 'at', lat.toFixed(2), lng.toFixed(2));
  
  if (onCitySelect) {
    onCitySelect(cityName);
  }
};

        map.on('click', handleMapClick);

      }).catch((err: any) => {
        console.error('Failed to load Leaflet:', err);
      });
    }, 50);

    return () => {
      if (mapInstance.current) {
        mapInstance.current.remove();
        mapInstance.current = null;
      }
    };

  }, []);

  // Explicit height and width - 80vh for visible map area
  return (
    <div 
      ref={mapRef} 
      style={{ height: '80vh', width: '100%' }}
      className="bg-gray-200"
    >
      Map loading...
    </div>
  );
};

export default ChinaMap;
