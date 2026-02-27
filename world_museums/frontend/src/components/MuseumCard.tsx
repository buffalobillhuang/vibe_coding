import React from 'react';
import { Museum } from '@/services/types';

interface MuseumCardProps {
  museum: Museum;
}

const MuseumCard: React.FC<MuseumCardProps> = ({ museum }) => {
  return (
    <div className="border-b border-gray-200 pb-4 mb-4 last:border-b-0">
      {/* Header */}
      <h3 className="text-xl font-bold text-gray-900">{museum.name}</h3>
      <p className="text-sm text-gray-600 mt-1">{museum.location}</p>

      {/* Introduction */}
      <div className="mt-3">
        <h4 className="font-semibold text-gray-800 text-sm">简介</h4>
        <p className="text-sm text-gray-700 mt-1 leading-relaxed line-clamp-3">
          {museum.introduction}
        </p>
      </div>

      {/* Image Gallery (3-5 images) */}
      {museum.images.length > 0 && (
        <div className="mt-4">
          <h4 className="font-semibold text-gray-800 text-sm">图片 ({museum.images.length})</h4>
          <div className="mt-2 grid grid-cols-3 gap-2">
            {museum.images.slice(0, 5).map((url, idx) => (
              <img
                key={idx}
                src={url}
                alt={`${museum.name} ${idx + 1}`}
                className="w-full h-24 object-cover rounded-lg hover:scale-105 transition-transform duration-200 cursor-pointer"
                loading="lazy"
              />
            ))}
          </div>
        </div>
      )}

      {/* Videos (1-3 embeds) */}
      {museum.videos.length > 0 && (
        <div className="mt-4">
          <h4 className="font-semibold text-gray-800 text-sm">视频 ({museum.videos.length})</h4>
          <div className="mt-2 space-y-2">
            {museum.videos.slice(0, 3).map((video, idx) => (
              <iframe
                key={idx}
                src={video.embedUrl}
                title={video.title}
                className="w-full h-48 rounded-lg border-0"
                allowFullScreen
              />
            ))}
          </div>
        </div>
      )}

      {/* Sources */}
      <div className="mt-3 text-xs text-gray-500">
        {museum.sources.baike && (
          <span>来源：<a href={museum.sources.baike} target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">百度百科</a></span>
        )}
        {museum.sources.bilibili && (
          <span className="ml-2"><a href={museum.sources.bilibili} target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:underline">哔哩哔哩</a></span>
        )}
      </div>
    </div>
  );
};

export default MuseumCard;
